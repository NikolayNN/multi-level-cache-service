package config

import (
	"time"
)

type cacheNameable interface {
	GetCacheName() string
}

type CacheService interface {
	GetCache(cacheId cacheNameable) Cache
	GetPrefix(cacheId cacheNameable) string
	GetTtl(cacheId cacheNameable, level int) time.Duration
	IsLevelEnabled(cacheId cacheNameable, level int) bool
}

type CacheServiceImpl struct {
	Caches map[string]Cache
}

func NewCacheService(cfg *AppConfig) CacheService {
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

func (s *CacheServiceImpl) GetCache(cacheId cacheNameable) Cache {
	return s.Caches[cacheId.GetCacheName()]
}

func (s *CacheServiceImpl) GetPrefix(cacheId cacheNameable) string {
	return s.Caches[cacheId.GetCacheName()].Prefix
}

func (s *CacheServiceImpl) GetTtl(cacheId cacheNameable, level int) time.Duration {
	return s.Caches[cacheId.GetCacheName()].Layers[level].TTL
}

func (s *CacheServiceImpl) IsLevelEnabled(cacheId cacheNameable, level int) bool {
	return s.Caches[cacheId.GetCacheName()].Layers[level].Enabled
}
