package config

import (
	"time"
)

type cacheNameable interface {
	GetCacheName() string
}

type CacheService interface {
	GetCache(cacheId cacheNameable) (cache Cache, ok bool)
	GetCacheByName(cacheName string) (cache Cache, ok bool)
	GetPrefix(cacheId cacheNameable) (prefix string, ok bool)
	GetTtl(cacheId cacheNameable, level int) (ttl time.Duration, ok bool)
	IsLevelEnabled(cacheId cacheNameable, level int) (enabled bool, ok bool)
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

func (s *CacheServiceImpl) GetCache(cacheId cacheNameable) (Cache, bool) {
	return s.GetCacheByName(cacheId.GetCacheName())
}

func (s *CacheServiceImpl) GetCacheByName(cacheName string) (Cache, bool) {
	cache, ok := s.Caches[cacheName]
	return cache, ok
}

func (s *CacheServiceImpl) GetPrefix(cacheId cacheNameable) (string, bool) {
	cache, ok := s.GetCache(cacheId)
	if !ok {
		return "", ok
	}
	return cache.Prefix, ok
}

func (s *CacheServiceImpl) GetTtl(cacheId cacheNameable, level int) (time.Duration, bool) {
	cache, ok := s.GetCache(cacheId)
	if !ok || level < 0 || level >= len(cache.Layers) {
		return 0, false
	}

	return cache.Layers[level].TTL, ok
}

func (s *CacheServiceImpl) IsLevelEnabled(cacheId cacheNameable, level int) (bool, bool) {
	cache, ok := s.GetCache(cacheId)
	if !ok || level < 0 || level >= len(cache.Layers) {
		return false, false
	}
	return cache.Layers[level].Enabled, ok
}
