package manager

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache"
	"aur-cache-service/internal/integration"
	"context"
	"log"
	"time"
)

// Manager определяет высокоуровневый интерфейс управления данными в многослойном кэше.
//
// Интерфейс инкапсулирует:
//   - преобразование внешних dto.CacheId / dto.CacheEntry во внутренние ResolvedCacheId / ResolvedCacheEntry;
//   - получение данных из кэша с fallback на внешний источник;
//   - автоматическое заполнение пропущенных уровней при hit на нижних слоях;
//   - массовую вставку и удаление значений во всех слоях кэша.
//
// Это упрощает использование кэша, позволяя клиентскому коду работать с простыми типами и не заботиться о внутренних деталях.
type Manager interface {

	// GetAll получает значения по заданным ключам.
	// Выполняет поиск во всех слоях кэша сверху вниз, при промахе — запрашивает внешний источник.
	// Затем актуализирует недостающие уровни кэша.
	GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit

	// PutAll вставляет записи во все уровни кэша.
	PutAll(ctx context.Context, entries []*dto.CacheEntry)

	// EvictAll удаляет записи со всех уровней кэша.
	EvictAll(ctx context.Context, ids []*dto.CacheId)
}

type ManagerImpl struct {
	cacheController    cache.Controller
	externalController integration.Controller
	mapper             *dto.ResolverMapper
}

func (m *ManagerImpl) GetAll(ctx context.Context, cacheIds []*dto.CacheId) []*dto.CacheEntryHit {

        log.Printf("manager GetAll %d ids", len(cacheIds))

        resolvedIds := m.mapper.MapAllResolvedCacheId(cacheIds)
        getResults := m.cacheController.GetAll(ctx, resolvedIds)

	// collect
	finalHits := make([]*dto.ResolvedCacheHit, 0, len(cacheIds))
	for _, r := range getResults {
		finalHits = append(finalHits, r.Hits...)
	}

	if len(getResults) == 0 {
		return []*dto.CacheEntryHit{}
	}

        if len(getResults[len(getResults)-1].Misses) > 0 {
                log.Printf("fetching %d ids from external source", len(getResults[len(getResults)-1].Misses))
        }
        fromExternal := m.externalController.GetAll(ctx, getResults[len(getResults)-1].Misses)
        finalHits = append(finalHits, fromExternal.Hits...)

        derivedCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
        defer cancel()

        log.Printf("start fillMissingLevels goroutine")
        go m.fillMissingLevels(derivedCtx, finalHits, getResults)

	return m.mapper.MapAllCacheEntryHit(finalHits)
}

func (m *ManagerImpl) fillMissingLevels(ctx context.Context, finalHits []*dto.ResolvedCacheHit, getResults []*dto.GetResult) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in goroutine fillMissingLevels: %v", r)
		}
	}()

	hitMap := make(map[string]*dto.ResolvedCacheHit, len(finalHits))
	for _, hit := range finalHits {
		hitMap[hit.GetStorageKey()] = hit
	}

	for level, result := range getResults {
		if level == 0 {
			continue
		}

		toPut := make([]*dto.ResolvedCacheEntry, 0)
		for _, missed := range result.Misses {
			if hit, ok := hitMap[missed.GetStorageKey()]; ok {
				toPut = append(toPut, hit.ResolvedCacheEntry)
			}
		}

		if len(toPut) > 0 {
			m.cacheController.PutAll(ctx, toPut, level-1)
		}
	}
}

func (m *ManagerImpl) PutAll(ctx context.Context, entries []*dto.CacheEntry) {
	resolvedEntries := m.mapper.MapAllResolvedCacheEntry(entries)
	m.cacheController.PutAllToAllLevels(ctx, resolvedEntries)
}

func (m *ManagerImpl) EvictAll(ctx context.Context, ids []*dto.CacheId) {
	resolvedIds := m.mapper.MapAllResolvedCacheId(ids)
	m.cacheController.DeleteAll(ctx, resolvedIds)
}
