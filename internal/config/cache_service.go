package config

import (
	"aur-cache-service/api/dto"
	"time"
)

type CacheService interface {
	GetByName(cacheName string) Cache
	GetCacheByCacheId(cacheId dto.CacheIdRef) Cache
	toStorageKey(cacheNameKey dto.CacheIdRef) string
}

type CacheServiceImpl struct {
	Caches map[string]Cache
}

func NewCacheService(cfg *AppConfig) *CacheServiceImpl {
	if cfg == nil {
		panic("cacheConfigStorage is nil")
	}
	caches := make(map[string]Cache)
	for _, cache := range cfg.Caches {
		caches[cache.Name] = cache
	}
	return &CacheServiceImpl{
		Caches: caches,
	}
}

func (s *CacheServiceImpl) GetCacheByCacheId(cacheId dto.CacheIdRef) Cache {
	return s.Caches[cacheId.GetCacheName()]
}

func (s *CacheServiceImpl) GetPrefixByCacheId(cacheId dto.CacheIdRef) string {
	return s.Caches[cacheId.GetCacheName()].Prefix
}

func (s *CacheServiceImpl) GetTtlByCacheId(cacheId dto.CacheIdRef, level int) time.Duration {
	return s.Caches[cacheId.GetCacheName()].Layers[level].TTL
}

func (s *CacheServiceImpl) IsLevelEnabledByCacheId(cacheId dto.CacheIdRef, level int) bool {
	return s.Caches[cacheId.GetCacheName()].Layers[level].Enabled
}
