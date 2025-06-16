package cache

import (
	"aur-cache-service/internal/cache/config"
	"aur-cache-service/internal/cache/providers"

	"go.uber.org/zap"
	"telegram-alerts-go/alert"
)

func CreateCacheService(configFilePath string) config.CacheService {
	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		zap.S().Errorw(alert.Prefix("error reading config file"), "error", err)
	}
	return config.NewCacheService(appConfig)
}

func CreateLayersCacheController(configFilePath string, configCacheService config.CacheService) Controller {

	appConfig, err := config.LoadAppConfig(configFilePath)
	if err != nil {
		zap.S().Errorw(alert.Prefix("error reading config file"), "error", err)
	}

	providerService := config.NewLayerProviderService(appConfig)

	clientServices, err := providers.CreateNewServiceList(providerService.LayerProviders, configCacheService)
	if err != nil {
		zap.S().Errorw(alert.Prefix("error creating service list"), "error", err)
	}

	return CreateControllerImpl(clientServices)
}
