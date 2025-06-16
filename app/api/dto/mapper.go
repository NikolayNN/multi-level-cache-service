package dto

import (
	"aur-cache-service/internal/cache/config"
	"go.uber.org/zap"
)

type ResolverMapper struct {
	cacheConfigService config.CacheService
}

const StorageKeySeparator = ":"

func NewResolverMapper(cacheConfigService config.CacheService) *ResolverMapper {
	return &ResolverMapper{cacheConfigService: cacheConfigService}
}

func (s *ResolverMapper) MapAllResolvedCacheEntry(cacheEntries []*CacheEntry) []*ResolvedCacheEntry {
	resolvedList := make([]*ResolvedCacheEntry, 0, len(cacheEntries))
	for _, cacheEntry := range cacheEntries {
		resolved, err := s.mapResolvedCacheEntry(cacheEntry)
		if err != nil {
			zap.S().Errorw("while resolve cacheEntry", "cacheEntry", cacheEntry, "error", err)
		} else {
			resolvedList = append(resolvedList, resolved)
		}
	}
	return resolvedList
}

func (s *ResolverMapper) mapResolvedCacheEntry(cacheEntry *CacheEntry) (*ResolvedCacheEntry, error) {
	resolvedCacheId, err := s.mapResolvedCacheId(cacheEntry.CacheId)
	if err != nil {
		return nil, err
	}
	return &ResolvedCacheEntry{
		ResolvedCacheId: resolvedCacheId,
		Value:           cacheEntry.Value,
	}, nil
}

func (s *ResolverMapper) MapAllResolvedCacheId(cacheIds []*CacheId) []*ResolvedCacheId {
	resolvedList := make([]*ResolvedCacheId, 0, len(cacheIds))
	for _, cacheId := range cacheIds {
		resolved, err := s.mapResolvedCacheId(cacheId)
		if err != nil {
			zap.S().Errorw("while resolve cacheId", "cacheId", cacheId, "error", err)
		} else {
			resolvedList = append(resolvedList, resolved)
		}
	}
	return resolvedList
}

func (s *ResolverMapper) mapResolvedCacheId(cacheId *CacheId) (*ResolvedCacheId, error) {
	storageKey, err := s.toStorageKey(cacheId)
	if err != nil {
		return nil, err
	}
	return &ResolvedCacheId{
		CacheId:    cacheId,
		StorageKey: storageKey,
	}, nil
}

func (s *ResolverMapper) toStorageKey(cacheId CacheIdRef) (string, error) {
	prefix, err := s.cacheConfigService.GetPrefix(cacheId)
	if err != nil {
		return "", err
	}
	return prefix + StorageKeySeparator + cacheId.GetKey(), nil
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
