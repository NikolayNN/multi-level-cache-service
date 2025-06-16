package httpserver

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/manager"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"go.uber.org/zap"
	"telegram-alerts-go/alert"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	maxBodySize           = 5 << 20                       // Максимальный размер тела запроса: 5 МБ (5 * 2^20 байт)
	gzipThreshold         = 500                           // Минимальный размер ответа для сжатия gzip: 500 байт
	baseAPIPath           = "/api/cache"                  // Базовый путь для всех API эндпоинтов
	batchGetPath          = baseAPIPath + "/batch/get"    // POST /api/cache/batch/get - массовое получение
	batchPutPath          = baseAPIPath + "/batch/put"    // POST /api/cache/batch/put - массовое сохранение
	batchDeletePath       = baseAPIPath + "/batch/delete" // POST /api/cache/batch/delete - массовое удаление
	contentTypeJSON       = "application/json"            // MIME-тип для JSON
	headerContentEncoding = "Content-Encoding"            // HTTP заголовок для указания кодировки
	headerAcceptEncoding  = "Accept-Encoding"             // HTTP заголовок с поддерживаемыми кодировками
	headerVary            = "Vary"                        // HTTP заголовок для указания зависимости от других заголовков
	encodingGzip          = "gzip"                        // Название gzip кодировки
)

// NewRouter возвращает http.Handler с зарегистрированными эндпоинтами.
func NewRouter(adapter manager.ManagerAdapter) http.Handler {
	r := chi.NewRouter()

	r.Use(limitBody(maxBodySize))
	r.Use(decompressGzip)
	r.Use(compressGzip(gzipThreshold))
	r.Use(MetricsMiddleware)

	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	r.Post(batchGetPath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchGet(w, r, adapter)
	})
	r.Post(batchPutPath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchPut(w, r, adapter)
	})
	r.Post(batchDeletePath, func(w http.ResponseWriter, r *http.Request) {
		handleBatchDelete(w, r, adapter)
	})

	r.Route(baseAPIPath, func(r chi.Router) {
		r.MethodFunc(http.MethodGet, "/*", func(w http.ResponseWriter, r *http.Request) {
			handleSingle(w, r, adapter)
		})
		r.MethodFunc(http.MethodPut, "/*", func(w http.ResponseWriter, r *http.Request) {
			handleSingle(w, r, adapter)
		})
		r.MethodFunc(http.MethodDelete, "/*", func(w http.ResponseWriter, r *http.Request) {
			handleSingle(w, r, adapter)
		})
	})

	return r
}

func parsePath(path string) (cacheName, key string, ok bool) {
	// expected path: /api/cache/{cache}/{key...}
	trimmed := strings.TrimPrefix(path, baseAPIPath+"/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	cacheName, err1 := url.PathUnescape(parts[0])
	key, err2 := url.PathUnescape(parts[1])
	if err1 != nil || err2 != nil {
		return "", "", false
	}
	return cacheName, key, true
}

func handleSingle(w http.ResponseWriter, r *http.Request, adapter manager.ManagerAdapter) {
	cacheName, key, ok := parsePath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	id := &dto.CacheId{CacheName: cacheName, Key: key}

	switch r.Method {
	case http.MethodGet:
		hit := adapter.Get(r.Context(), id)
		if hit == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentTypeJSON)
		if err := json.NewEncoder(w).Encode(hit); err != nil {
			zap.S().Errorw(alert.Prefix("encode error"), "error", err)
		}

	case http.MethodPut:
		defer r.Body.Close()
		if !strings.HasPrefix(r.Header.Get("Content-Type"), contentTypeJSON) {
			http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
			return
		}
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		adapter.Put(r.Context(), &dto.CacheEntry{CacheId: id, Value: &raw})
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		adapter.Evict(r.Context(), id)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
