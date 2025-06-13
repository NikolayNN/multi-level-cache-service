package main

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache"
	"aur-cache-service/internal/httpserver"
	"aur-cache-service/internal/integration"
	"aur-cache-service/internal/manager"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	configFilePath  = "configs/config.yml"
	putAllTimeout   = 10 * time.Second
	evictAllTimeout = 10 * time.Second
	port            = 8080
)

func main() {

	configCacheService := cache.CreateCacheService(configFilePath)

	layersCacheController := cache.CreateLayersCacheController(configFilePath, configCacheService)

	httpCacheController := integration.CreateHttpCacheController(configCacheService)

	mapper := dto.NewResolverMapper(configCacheService)

	mainAdapter := manager.CreateAsyncManagerAdapter(mapper, layersCacheController, httpCacheController, putAllTimeout, evictAllTimeout)

	router := httpserver.NewRouter(&mainAdapter)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Printf("starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
