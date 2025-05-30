package controller

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache/providers"
	"log"
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
//   - DeleteAll:
//     Удаляет значения со всех уровней.
//
// Пример сценария:
//   1. Клиент запрашивает значения → GetAll обходит уровни и возвращает найденные значения.
//   2. После получения значений, недостающие ключи можно сохранить в нижние уровни через PutAll.
//   3. При необходимости — удалить данные из всех слоёв через DeleteAll.
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
	GetAll(reqs []*dto.ResolvedCacheId) (results []*providers.GetResult)
	PutAll(entries []*dto.ResolvedCacheEntry, boundLevel int)
	DeleteAll(reqs []*dto.ResolvedCacheId)
}

type ControllerImpl struct {
	services []providers.Service
}

func createControllerImpl(services []providers.Service) Controller {
	return &ControllerImpl{services: services}
}

// GetAll обходит все уровни кэша сверху вниз, собирая значения и возвращая срез GetResult для каждого слоя.
func (c *ControllerImpl) GetAll(reqs []*dto.ResolvedCacheId) (results []*providers.GetResult) {

	results = make([]*providers.GetResult, len(c.services))
	for i, service := range c.services {
		r, err := service.GetAll(reqs)
		if err != nil {
			log.Printf("Layer %d unavailable: %v", i, err)
			results[i] = &providers.GetResult{
				Hits:    []*dto.ResolvedCacheHit{},
				Misses:  []*dto.ResolvedCacheId{},
				Skipped: reqs,
			}
			continue
		}
		results[i] = r
		reqs = append(r.Misses, r.Skipped...)
	}
	return
}

// PutAll вставляет значения во все уровни до boundLevel включительно
func (c *ControllerImpl) PutAll(entries []*dto.ResolvedCacheEntry, boundLevel int) {
	for i, service := range c.services {
		if i > boundLevel {
			break
		}
		err := service.PutAll(entries)
		if err != nil {
			log.Printf("Layer %d unavailable: %v", i, err)
		}
	}
}

// DeleteAll удаляет значения со всех уровней
func (c *ControllerImpl) DeleteAll(reqs []*dto.ResolvedCacheId) {
	for i, service := range c.services {
		err := service.DeleteAll(reqs)
		if err != nil {
			log.Printf("Layer %d unavailable: %v", i, err)
		}
	}
}
