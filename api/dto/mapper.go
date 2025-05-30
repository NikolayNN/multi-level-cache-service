package dto

import (
	"aur-cache-service/internal/cache/config"
)

type ResolverMapper struct {
	cacheConfigService *config.CacheServiceImpl
}

const StorageKeySeparator = ":"

func NewResolverService(cacheConfigService *config.CacheServiceImpl) *ResolverMapper {
	return &ResolverMapper{cacheConfigService: cacheConfigService}
}

func (s *ResolverMapper) MapAllResolveCacheEntry(cacheEntries []*CacheEntry) []*ResolvedCacheEntry {
	resolved := make([]*ResolvedCacheEntry, 0, len(cacheEntries))
	for _, req := range cacheEntries {
		cacheReq := s.mapResolvedCacheEntry(req)
		resolved = append(resolved, cacheReq)
	}
	return resolved
}

func (s *ResolverMapper) mapResolvedCacheEntry(cacheEntry *CacheEntry) *ResolvedCacheEntry {
	resolvedCacheId := s.MapResolvedCacheId(cacheEntry.CacheId)
	return &ResolvedCacheEntry{
		ResolvedCacheId: resolvedCacheId,
		Value:           cacheEntry.Value,
	}
}

func (s *ResolverMapper) MapAllResolvedCacheId(cacheIds []*CacheId) []*ResolvedCacheId {
	resolved := make([]*ResolvedCacheId, 0, len(cacheIds))
	for _, req := range cacheIds {
		cacheReq := s.MapResolvedCacheId(req)
		resolved = append(resolved, cacheReq)
	}
	return resolved
}

func (s *ResolverMapper) MapResolvedCacheId(cacheId *CacheId) *ResolvedCacheId {
	return &ResolvedCacheId{
		CacheId:    cacheId,
		StorageKey: s.toStorageKey(cacheId),
	}
}

func (s *ResolverMapper) toStorageKey(cacheId CacheIdRef) string {
	prefix := s.cacheConfigService.GetPrefixByCacheId(cacheId)
	return prefix + StorageKeySeparator + cacheId.GetKey()
}
