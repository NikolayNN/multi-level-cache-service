package cache

import (
	"aur-cache-service/internal/cache/config"
	"aur-cache-service/internal/cache/providers"
	"log"
)

func CreateCacheService(configFilePath string) config.CacheService {
	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
	}
	return config.NewCacheService(appConfig)
}

func CreateLayersCacheController(configFilePath string, configCacheService config.CacheService) Controller {

	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	providerService := config.NewLayerProviderService(appConfig)

	clientServices, err := providers.CreateNewServiceList(providerService.LayerProviders, configCacheService)
	if err != nil {
		log.Printf("Error creating service list: %v", err)
	}

	return CreateControllerImpl(clientServices)
}
