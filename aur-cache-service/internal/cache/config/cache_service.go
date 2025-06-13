package config

import (
	"fmt"
	"time"
)

type CacheNameable interface {
	GetCacheName() string
}

type CacheService interface {
	GetCache(cacheId CacheNameable) (cache Cache, err error)
	GetCacheByName(cacheName string) (cache Cache, err error)
	GetPrefix(cacheId CacheNameable) (prefix string, err error)
	GetTtl(cacheId CacheNameable, level int) (ttl time.Duration, err error)
	IsLevelEnabled(cacheId CacheNameable, level int) (enabled bool, err error)
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

func (s *CacheServiceImpl) GetCache(cacheId CacheNameable) (Cache, error) {
	return s.GetCacheByName(cacheId.GetCacheName())
}

func (s *CacheServiceImpl) GetCacheByName(cacheName string) (Cache, error) {
	cache, ok := s.Caches[cacheName]
	if !ok {
		return Cache{}, fmt.Errorf("cache with name %q not found", cacheName)
	}
	return cache, nil
}

func (s *CacheServiceImpl) GetPrefix(cacheId CacheNameable) (string, error) {
	cache, err := s.GetCache(cacheId)
	if err != nil {
		return "", err
	}
	return cache.Prefix, nil
}

func (s *CacheServiceImpl) GetTtl(cacheId CacheNameable, level int) (time.Duration, error) {
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

func (s *CacheServiceImpl) IsLevelEnabled(cacheId CacheNameable, level int) (bool, error) {
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
		return CacheLayerConfig{}, fmt.Errorf("request wrong level %d for cacheName %q", level, cache.Name)
	}
	return cache.Layers[level], nil
}
