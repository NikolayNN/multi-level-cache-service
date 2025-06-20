package cache

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache/providers"
	"aur-cache-service/internal/metrics"
	"context"

	"go.uber.org/zap"
)

// Controller определяет высокоуровневый интерфейс для работы с многослойным кэшем.
//
// Архитектура многослойного кэша подразумевает, что данные могут храниться
// на различных уровнях (L0, L1, L2 и т.д.), каждый из которых может быть
// активен или отключён, быстрым (например, Ristretto) или медленным (например, RocksDB).
//
// Controller абстрагирует обход всех уровней и обеспечивает:
//
//   - GetAll:
//     Итерирует по уровням сверху вниз, извлекая значения и собирая статистику по каждому уровню.
//     Возвращает срез результатов (hits/misses/skipped) для каждого слоя.
//
//   - PutAll:
//     Сохраняет значения во все уровни до заданного уровня включительно.
//     Это используется, чтобы "прокинуть" значения вниз (например, при кэшировании результата запроса).
//
//   - EvictAll:
//     Удаляет значения со всех уровней.
//
// Пример сценария:
//   1. Клиент запрашивает значения → GetAll обходит уровни и возвращает найденные значения.
//   2. После получения значений, недостающие ключи можно сохранить в нижние уровни через PutAll.
//   3. При необходимости — удалить данные из всех слоёв через EvictAll.
//
// Каждый слой реализован через интерфейс Service и управляется независимо.
// 	┌──────────────┐
// 	│  Client      │
// 	└─────┬────────┘
//  	  ↓
// 	┌──────────────┐
//	│ Level 0      │  hits: {1}   misses: {2,3,4}   skips: {5}
//	└─────┬────────┘
//	      ↓
//	┌──────────────┐
//	│ Level 1      │  hits: {2,3} misses: {4}       skips: {5}
//	└─────┬────────┘
//	      ↓
//	┌──────────────┐
//	│ Level 2      │  hits: {4,5} misses: {}        skips: {}
//	└──────────────┘

type Controller interface {
	GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) (results []*dto.GetResult)
	PutAll(ctx context.Context, entries []*dto.ResolvedCacheEntry, boundLevel int)
	PutAllToAllLevels(ctx context.Context, entries []*dto.ResolvedCacheEntry)
	DeleteAll(ctx context.Context, reqs []*dto.ResolvedCacheId)
}

type ControllerImpl struct {
	services []providers.Service
}

func CreateControllerImpl(services []providers.Service) Controller {
	return &ControllerImpl{services: services}
}

// GetAll обходит все уровни кэша сверху вниз, собирая значения и возвращая срез GetResult для каждого слоя.
func (c *ControllerImpl) GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) (results []*dto.GetResult) {

	results = make([]*dto.GetResult, len(c.services))
	for i, service := range c.services {
		r, err := service.GetAll(ctx, reqs)
		if err != nil {
			zap.S().Warnw("layer unavailable", "layer", i, "error", err)
			results[i] = &dto.GetResult{
				Hits:    []*dto.ResolvedCacheHit{},
				Misses:  []*dto.ResolvedCacheId{},
				Skipped: reqs,
			}
			metrics.RecordCacheLayer(i, 0, len(reqs))
			continue
		}
		results[i] = r
		metrics.RecordCacheLayer(i, len(r.Hits), len(r.Misses))
		nextReqs := make([]*dto.ResolvedCacheId, 0, len(r.Misses)+len(r.Skipped))
		nextReqs = append(nextReqs, r.Misses...)
		nextReqs = append(nextReqs, r.Skipped...)
		reqs = nextReqs
	}
	return
}

// PutAll вставляет значения во все уровни до boundLevel включительно
func (c *ControllerImpl) PutAll(ctx context.Context, entries []*dto.ResolvedCacheEntry, boundLevel int) {
	for i, service := range c.services {
		if i > boundLevel {
			break
		}
		err := service.PutAll(ctx, entries)
		if err != nil {
			zap.S().Warnw("layer unavailable", "layer", i, "error", err)
		}
	}
}

func (c *ControllerImpl) PutAllToAllLevels(ctx context.Context, entries []*dto.ResolvedCacheEntry) {
	c.PutAll(ctx, entries, len(c.services)-1)
}

// DeleteAll удаляет значения со всех уровней
func (c *ControllerImpl) DeleteAll(ctx context.Context, reqs []*dto.ResolvedCacheId) {

	for i, service := range c.services {
		err := service.DeleteAll(ctx, reqs)
		if err != nil {
			zap.S().Warnw("layer unavailable", "layer", i, "error", err)
		}
	}
}
