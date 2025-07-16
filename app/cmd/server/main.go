package main

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache"
	"aur-cache-service/internal/httpserver"
	"aur-cache-service/internal/integration"
	"aur-cache-service/internal/logger"
	"aur-cache-service/internal/manager"
	"aur-cache-service/internal/metrics"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"sync"
	"time"

	"telegram-alerts-go/alert"
)

const (
	configFilePath  = "/configs/config.yml"
	putAllTimeout   = 10 * time.Second
	evictAllTimeout = 10 * time.Second
	portApi         = 8080
	portMetrics     = 9080
)

func main() {

	if _, err := logger.Init(); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	defer logger.Sync()

	metrics.Register()

	configCacheService := cache.CreateCacheService(configFilePath)

	layersCacheController := cache.CreateLayersCacheController(configFilePath, configCacheService)

	httpCacheController := integration.CreateHttpCacheController(configCacheService)

	mapper := dto.NewResolverMapper(configCacheService)

	mainAdapter := manager.CreateAsyncManagerAdapter(mapper, layersCacheController, httpCacheController, putAllTimeout, evictAllTimeout)

	routerApi := httpserver.NewRouter(&mainAdapter)
	routerMetrics := httpserver.NewMetricRouter()

	// запуск двух HTTP-серверов параллельно (в отдельных горутинах),
	// и ожидание их завершения через sync.WaitGroup
	//
	// Это нужно, чтобы:
	// - main-функция не завершилась раньше времени
	// - оба сервера работали одновременно
	//
	// Без этого программа бы завершилась сразу после запуска горутин.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		listenServer(routerApi, portApi)
	}()

	go func() {
		defer wg.Done()
		listenServer(routerMetrics, portMetrics)
	}()

	wg.Wait()
	//// end
}

func listenServer(router http.Handler, port int) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	zap.S().Infow("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.S().Fatalw(alert.Prefix("server error"), "error", err)
	}
}
