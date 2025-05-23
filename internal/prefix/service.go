package prefix

import (
	"aur-cache-service/internal/config"
)

type Service struct {
	cachePrefixes map[string]string
}

const separator = ":"

func New(cacheConfigStorage *config.CacheStorage) *Service {
	if cacheConfigStorage == nil {
		panic("cacheConfigStorage is nil")
	}

	cachePrefixes := make(map[string]string)
	for key, value := range cacheConfigStorage.Configs {
		cachePrefixes[key] = value.Prefix
	}

	return &Service{
		cachePrefixes: cachePrefixes,
	}
}

func (s *Service) ToCacheKey(cacheName string, key string) string {
	prefix, exists := s.cachePrefixes[cacheName]
	if !exists {
		panic("unknown cache: " + cacheName)
	}
	return prefix + separator + key
}
