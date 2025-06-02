package integration

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache/config"
	"context"
	"encoding/json"
	"log"
	"sync"
)

// Service представляет интерфейс получения значений из внешнего API по списку идентификаторов кеша.
//
// Метод GetAll предназначен для пакетной загрузки значений по ключам, сгруппированным по CacheName.
// Каждая уникальная группа (по CacheName) отправляется отдельным HTTP-запросом через httpBatchFetcher.
// Реализация может использовать параллелизм, при этом общее количество одновременных запросов ограничено.
//
// Поведение:
//   - Если данные по ключу успешно получены — они попадают в Hits.
//   - Если данные отсутствуют — ключ считается Miss.
//   - Если возникла ошибка при запросе всей группы — все ключи из группы считаются Skipped.
//
// Метод гарантирует:
//   - Потокобезопасное слияние результатов.
//   - Не более maxParallel параллельных HTTP-запросов.
type Service interface {
	GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) *dto.GetResult
}

type ServiceImpl struct {
	fetcher       httpBatchFetcher
	configService config.CacheService
}

const maxParallel = 8

// GetAll запрашивает данные для всех ResolvedCacheId, параллельно обрабатывая группы
// по CacheName. Одновременно выполняется не более 8 HTTP-запросов.
func (s *ServiceImpl) GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) *dto.GetResult {
	if reqs == nil {
		return &dto.GetResult{}
	}
	grouped := s.group(reqs)
	final := &dto.GetResult{}

	var wg sync.WaitGroup
	var mu sync.Mutex
	limiter := make(chan struct{}, maxParallel)

	for cacheName, group := range grouped {
		wg.Add(1)

		limiter <- struct{}{} // занять слот

		go func(name string, grp []*dto.ResolvedCacheId) {
			defer wg.Done()
			defer func() { <-limiter }() // освободить слот

			result := s.handleGroup(ctx, name, grp)

			mu.Lock()
			final.Merge(result)
			mu.Unlock()
		}(cacheName, group)
	}

	wg.Wait()
	return final
}

// обрабатывает одну группу ключей одного кэша
func (s *ServiceImpl) handleGroup(ctx context.Context, cacheName string, group []*dto.ResolvedCacheId) *dto.GetResult {
	cfg := s.configService.GetByName(cacheName).Api.GetBatch

	keys := s.extractKeys(group)

	respMap, err := s.fetcher.GetAll(ctx, keys, &cfg)
	if err != nil {
		log.Printf("fetch error for %s: %v", cacheName, err)
		return &dto.GetResult{Skipped: group}
	}

	return s.classify(group, respMap)
}

// превращает []*ResolvedCacheId → []string ключей
func (s *ServiceImpl) extractKeys(group []*dto.ResolvedCacheId) []string {
	keys := make([]string, len(group))
	for i, r := range group {
		keys[i] = r.CacheId.Key
	}
	return keys
}

// формирует Hits / Misses
func (s *ServiceImpl) classify(group []*dto.ResolvedCacheId, respMap map[string]*json.RawMessage) *dto.GetResult {
	res := &dto.GetResult{}
	for _, r := range group {
		if val, ok := respMap[r.CacheId.Key]; ok {
			res.Hits = append(res.Hits, &dto.ResolvedCacheHit{
				ResolvedCacheEntry: &dto.ResolvedCacheEntry{
					ResolvedCacheId: r,
					Value:           val,
				},
				Found: true,
			})
		} else {
			res.Misses = append(res.Misses, r)
		}
	}
	return res
}

// группирует входные запросы по CacheName
func (s *ServiceImpl) group(reqs []*dto.ResolvedCacheId) map[string][]*dto.ResolvedCacheId {
	grouped := make(map[string][]*dto.ResolvedCacheId)
	for _, r := range reqs {
		name := r.CacheId.CacheName
		grouped[name] = append(grouped[name], r)
	}
	return grouped
}
