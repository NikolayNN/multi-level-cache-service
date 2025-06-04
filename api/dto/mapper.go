package dto

import (
	"aur-cache-service/internal/cache/config"
)

type ResolverMapper struct {
	cacheConfigService *config.CacheServiceImpl
}

const StorageKeySeparator = ":"

func NewResolverMapper(cacheConfigService *config.CacheServiceImpl) *ResolverMapper {
	return &ResolverMapper{cacheConfigService: cacheConfigService}
}

func (s *ResolverMapper) MapAllResolvedCacheEntry(cacheEntries []*CacheEntry) []*ResolvedCacheEntry {
	resolved := make([]*ResolvedCacheEntry, 0, len(cacheEntries))
	for _, req := range cacheEntries {
		cacheReq := s.mapResolvedCacheEntry(req)
		resolved = append(resolved, cacheReq)
	}
	return resolved
}

func (s *ResolverMapper) mapResolvedCacheEntry(cacheEntry *CacheEntry) *ResolvedCacheEntry {
	resolvedCacheId := s.ьapResolvedCacheId(cacheEntry.CacheId)
	return &ResolvedCacheEntry{
		ResolvedCacheId: resolvedCacheId,
		Value:           cacheEntry.Value,
	}
}

func (s *ResolverMapper) MapAllResolvedCacheId(cacheIds []*CacheId) []*ResolvedCacheId {
	resolved := make([]*ResolvedCacheId, 0, len(cacheIds))
	for _, req := range cacheIds {
		cacheReq := s.ьapResolvedCacheId(req)
		resolved = append(resolved, cacheReq)
	}
	return resolved
}

func (s *ResolverMapper) ьapResolvedCacheId(cacheId *CacheId) *ResolvedCacheId {
	return &ResolvedCacheId{
		CacheId:    cacheId,
		StorageKey: s.toStorageKey(cacheId),
	}
}

func (s *ResolverMapper) toStorageKey(cacheId CacheIdRef) string {
	prefix := s.cacheConfigService.GetPrefixByCacheId(cacheId)
	return prefix + StorageKeySeparator + cacheId.GetKey()
}

func (s *ResolverMapper) MapAllCacheEntryHit(resolvedHits []*ResolvedCacheHit) []*CacheEntryHit {
	hits := make([]*CacheEntryHit, 0, len(resolvedHits))
	for _, rh := range resolvedHits {
		hits = append(hits, s.mapCacheEntryHit(rh))
	}
	return hits
}

func (s *ResolverMapper) mapCacheEntryHit(resolved *ResolvedCacheHit) *CacheEntryHit {
	return &CacheEntryHit{
		CacheEntry: &CacheEntry{
			CacheId: resolved.ResolvedCacheEntry.ResolvedCacheId.CacheId,
			Value:   resolved.ResolvedCacheEntry.Value,
		},
		Found: resolved.Found,
	}
}
