package config

import (
	"fmt"
	"time"
)

type cacheNameable interface {
	GetCacheName() string
}

type CacheService interface {
	GetCache(cacheId cacheNameable) (cache Cache, err error)
	GetCacheByName(cacheName string) (cache Cache, err error)
	GetPrefix(cacheId cacheNameable) (prefix string, err error)
	GetTtl(cacheId cacheNameable, level int) (ttl time.Duration, err error)
	IsLevelEnabled(cacheId cacheNameable, level int) (enabled bool, err error)
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

func (s *CacheServiceImpl) GetCache(cacheId cacheNameable) (Cache, error) {
	return s.GetCacheByName(cacheId.GetCacheName())
}

func (s *CacheServiceImpl) GetCacheByName(cacheName string) (Cache, error) {
	cache, ok := s.Caches[cacheName]
	if !ok {
		return Cache{}, fmt.Errorf("cache with name %q not found", cacheName)
	}
	return cache, nil
}

func (s *CacheServiceImpl) GetPrefix(cacheId cacheNameable) (string, error) {
	cache, err := s.GetCache(cacheId)
	if err != nil {
		return "", err
	}
	return cache.Prefix, nil
}

func (s *CacheServiceImpl) GetTtl(cacheId cacheNameable, level int) (time.Duration, error) {
	cache, err := s.GetCache(cacheId)
	if err != nil {
		return 0, err
	}

	layer, err := s.getLevel(cache, level)
	if err != nil {
		return 0, err
	}

	return layer.TTL, nil
}

func (s *CacheServiceImpl) IsLevelEnabled(cacheId cacheNameable, level int) (bool, error) {
	cache, err := s.GetCache(cacheId)
	if err != nil {
		return false, err
	}

	layer, err := s.getLevel(cache, level)
	if err != nil {
		return false, err
	}
	return layer.Enabled, nil
}

func (s *CacheServiceImpl) getLevel(cache Cache, level int) (CacheLayerConfig, error) {
	if level < 0 || level >= len(cache.Layers) {
		return CacheLayerConfig{}, fmt.Errorf("request wrong level %q for cacheName %d", cache.Name, level)
	}
	return cache.Layers[level], nil
}
