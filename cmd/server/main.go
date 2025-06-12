package main

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache"
	"aur-cache-service/internal/httpserver"
	"aur-cache-service/internal/integration"
	"aur-cache-service/internal/manager"
	"time"
)

const (
	configFilePath  = "configs/config.yml"
	putAllTimeout   = 10 * time.Second
	evictAllTimeout = 10 * time.Second
)

func main() {

	configCacheService := cache.CreateCacheService(configFilePath)

	layersCacheController := cache.CreateLayersCacheController(configFilePath, configCacheService)

	httpCacheController := integration.CreateHttpCacheController(configCacheService)

	mapper := dto.NewResolverMapper(configCacheService)

	mainAdapter := manager.CreateAsyncManagerAdapter(mapper, layersCacheController, httpCacheController, putAllTimeout, evictAllTimeout)

	router := httpserver.NewRouter(&mainAdapter)
}
