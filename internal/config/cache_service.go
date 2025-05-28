package config

import (
	"aur-cache-service/internal/resolvers/cmn"
)

type CacheService interface {
	GetByName(cacheName string) Cache
	CacheKey(cacheNameKey cmn.CacheNameKey) string
	ToCacheKey(cacheName string, key string) string
}

type CacheServiceImpl struct {
	CacheStorage *CacheStorage
}

const prefixKeySeparator = ":"

func NewCacheService(cacheConfigStorage *CacheStorage) *CacheServiceImpl {
	if cacheConfigStorage == nil {
		panic("cacheConfigStorage is nil")
	}
	return &CacheServiceImpl{
		CacheStorage: cacheConfigStorage,
	}
}

func (s *CacheServiceImpl) GetByName(cacheName string) Cache {
	return s.CacheStorage.Configs[cacheName]
}

func (s *CacheServiceImpl) CacheKey(cacheNameKey cmn.CacheNameKey) string {
	return s.ToCacheKey(cacheNameKey.GetCacheName(), cacheNameKey.GetKey())
}

func (s *CacheServiceImpl) ToCacheKey(cacheName string, key string) string {
	prefix := s.CacheStorage.Configs[cacheName].Prefix
	return prefix + prefixKeySeparator + key
}
