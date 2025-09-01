package httpserver

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/manager"
	"compress/gzip"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
	"telegram-alerts-go/alert"

	"github.com/go-chi/chi/v5"
)

const (
	maxBodySize           = 5 << 20                    // Максимальный размер тела запроса: 5 МБ (5 * 2^20 байт)
	gzipThreshold         = 500                        // Минимальный размер ответа для сжатия gzip: 500 байт
	baseAPIPath           = "/api/v1/cache"            // Базовый путь для всех API эндпоинтов
	getAllPath            = baseAPIPath + "/get_all"   // POST /api/v1/cache/get_all - массовое получение
	putAllPath            = baseAPIPath + "/put_all"   // POST /api/v1/cache/put_all - массовое сохранение
	evictAllPath          = baseAPIPath + "/evict_all" // POST /api/v1/cache/evict_all - массовое удаление
	contentTypeJSON       = "application/json"         // MIME-тип для JSON
	headerContentEncoding = "Content-Encoding"         // HTTP заголовок для указания кодировки
	headerAcceptEncoding  = "Accept-Encoding"          // HTTP заголовок с поддерживаемыми кодировками
	headerVary            = "Vary"                     // HTTP заголовок для указания зависимости от других заголовков
	encodingGzip          = "gzip"                     // Название gzip кодировки
	metricsPath           = "/metrics"                 // Путь для метрик Prometheus
	metricsHealthPath     = "/metrics/health"          // Путь для проверки состояния
)

func NewMetricRouter() http.Handler {
	metric_router := chi.NewRouter()

	// /metrics хендлер без middleware
	metric_router.Method(http.MethodGet, metricsPath, promhttp.Handler())

	// /metrics/health хендлер без middleware
	metric_router.Method(http.MethodGet, metricsHealthPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentTypeJSON)
		io.WriteString(w, `{"status":"UP"}`)
	}))
	return metric_router
}

// NewRouter возвращает http.Handler с зарегистрированными эндпоинтами.
func NewRouter(adapter manager.ManagerAdapter) http.Handler {
	api_router := chi.NewRouter()

	api_router.Use(limitBody(maxBodySize))
	api_router.Use(decompressGzip)
	api_router.Use(compressGzip(gzipThreshold))
	api_router.Use(MetricsMiddleware)

	api_router.Post(getAllPath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchGet(w, r, adapter)
	})
	api_router.Post(putAllPath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchPut(w, r, adapter)
	})
	api_router.Post(evictAllPath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchDelete(w, r, adapter)
	})

	// ранее здесь регистрировались одиночные операции GET, PUT и DELETE.
	// Они убраны, чтобы оставались только batch эндпоинты.

	return api_router
}

type batchRequest struct {
	Requests []dto.CacheEntry `json:"requests"`
}

type batchDeleteRequest struct {
	Requests []dto.CacheId `json:"requests"`
}

func handleBatchGet(w http.ResponseWriter, r *http.Request, adapter manager.ManagerAdapter) {
	defer r.Body.Close()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), contentTypeJSON) {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}
	var req struct {
		Requests []dto.CacheId `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Requests) == 0 {
		http.Error(w, "empty requests", http.StatusBadRequest)
		return
	}

	ids := make([]*dto.CacheId, len(req.Requests))
	for i := range req.Requests {
		ids[i] = &req.Requests[i]
	}
	hits := adapter.GetAll(r.Context(), ids)

	type cacheKey struct {
		cache string
		key   string
	}
	resMap := make(map[cacheKey]*dto.CacheEntryHit)
	for _, h := range hits {
		if h != nil && h.CacheEntry != nil && h.CacheEntry.CacheId != nil {
			k := cacheKey{h.CacheEntry.CacheId.CacheName, h.CacheEntry.CacheId.Key}
			resMap[k] = h
		}
	}

	results := make([]*dto.CacheEntryHit, len(ids))
	for i, id := range ids {
		k := cacheKey{id.CacheName, id.Key}
		if hit, ok := resMap[k]; ok {
			results[i] = hit
		} else {
			results[i] = &dto.CacheEntryHit{
				CacheEntry: &dto.CacheEntry{CacheId: id, Value: nil},
				Found:      false,
			}
		}
	}

	zap.S().Infow("processed batch get", "requests", req.Requests, "results", results)

	w.Header().Set("Content-Type", contentTypeJSON)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"results": results}); err != nil {
		zap.S().Errorw(alert.Prefix("encode error"), "error", err)
	}
}

func handleBatchPut(w http.ResponseWriter, r *http.Request, adapter manager.ManagerAdapter) {
	defer r.Body.Close()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), contentTypeJSON) {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}
	var req struct {
		Requests []dto.CacheEntry `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Requests) == 0 {
		http.Error(w, "empty requests", http.StatusBadRequest)
		return
	}

	entries := make([]*dto.CacheEntry, len(req.Requests))
	for i := range req.Requests {
		entries[i] = &req.Requests[i]
	}
	adapter.PutAll(r.Context(), entries)
	zap.S().Infow("processed batch put", "requests", req.Requests)
	w.WriteHeader(http.StatusOK)
}

func handleBatchDelete(w http.ResponseWriter, r *http.Request, adapter manager.ManagerAdapter) {
	defer r.Body.Close()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), contentTypeJSON) {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}
	var req struct {
		Requests []dto.CacheId `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Requests) == 0 {
		http.Error(w, "empty requests", http.StatusBadRequest)
		return
	}

	ids := make([]*dto.CacheId, len(req.Requests))
	for i := range req.Requests {
		ids[i] = &req.Requests[i]
	}
	adapter.EvictAll(r.Context(), ids)
	zap.S().Infow("processed batch delete", "requests", req.Requests)
	w.WriteHeader(http.StatusOK)
}

// ---- middleware ----

func limitBody(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

func decompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerContentEncoding) == encodingGzip {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = struct{ io.ReadCloser }{gz}
		}
		next.ServeHTTP(w, r)
	})
}

type bufferResponseWriter struct {
	http.ResponseWriter
	code int
	buf  strings.Builder
	once sync.Once
}

func (b *bufferResponseWriter) WriteHeader(statusCode int) {
	b.once.Do(func() { b.code = statusCode })
}

func (b *bufferResponseWriter) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

func compressGzip(threshold int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip) {
				next.ServeHTTP(w, r)
				return
			}
			brw := &bufferResponseWriter{ResponseWriter: w}
			next.ServeHTTP(brw, r)

			if brw.code == 0 {
				brw.code = http.StatusOK
			}

			data := brw.buf.String()
			if len(data) < threshold {
				w.WriteHeader(brw.code)
				io.WriteString(w, data)
				return
			}

			w.Header().Set(headerContentEncoding, encodingGzip)
			w.Header().Set(headerVary, headerAcceptEncoding)
			w.WriteHeader(brw.code)
			gz := gzip.NewWriter(w)
			if _, err := gz.Write([]byte(data)); err != nil {
				zap.S().Errorw(alert.Prefix("gzip write error"), "error", err)
			}
			gz.Close()
		})
	}
}
