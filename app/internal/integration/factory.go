package integration

import (
	"aur-cache-service/internal/cache/config"
	"net/http"
	"time"
)

func CreateHttpCacheController(configCacheService config.CacheService) Controller {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	fetcher := CreateHttpBatchFetcher(client)

	service := NewIntegrationService(configCacheService, fetcher)

	return CreateController(service)
}
