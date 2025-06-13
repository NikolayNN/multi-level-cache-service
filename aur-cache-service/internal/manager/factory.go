package manager

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache"
	"aur-cache-service/internal/integration"
	"time"
)

func CreateAsyncManagerAdapter(mapper *dto.ResolverMapper, layerCacheController cache.Controller, httpCacheController integration.Controller, putAllTimeout time.Duration, evictAllTimeout time.Duration) AsyncManagerAdapter {

	manager := ManagerImpl{
		cacheController:    layerCacheController,
		externalController: httpCacheController,
		mapper:             mapper,
	}

	return AsyncManagerAdapter{
		manager:         &manager,
		putAllTimeout:   putAllTimeout,
		evictAllTimeout: evictAllTimeout,
	}
}
