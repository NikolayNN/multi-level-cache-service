package cache

import (
	"aur-cache-service/internal/cache/config"
	"aur-cache-service/internal/cache/providers"
	"fmt"
)

func CreateCacheService(configFilePath string) config.CacheService {
	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error read config file  %+v", err)
	}
	return config.NewCacheService(appConfig)
}

func CreateLayersCacheController(configFilePath string, configCacheService config.CacheService) Controller {

	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error read config file  %+v", err)
	}

	providerService := config.NewLayerProviderService(appConfig)

	clientServices, err := providers.CreateNewServiceList(providerService.LayerProviders, configCacheService)
	if err != nil {
		fmt.Printf("Error create service list  %+v", err)
	}

	return CreateControllerImpl(clientServices)
}
