package providers

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/config"
	"encoding/json"
	"fmt"
)

// Service обеспечивает доступ к конкретному уровню кэширования.
// Реализации:
//
// ServiceImpl
// Каждый экземпляр ServiceImpl обслуживает только один слой (level) кэша —
// например, Ristretto (L0), Redis (L1) или RocksDB (L2) в многослойной архитектуре.
//
// Поведение методов:
//
//   - GetAll:
//
//   - Возвращает три категории ключей:
//
//   - hits    — ключи, для которых значение было найдено;
//
//   - misses  — ключи, у которых слой включён, но значение не найдено;
//
//   - skipped — ключи, у которых текущий слой отключён (disabled).
//
//   - PutAll:
//
//   - Сохраняет только те записи, у которых включён текущий слой;
//
//   - Остальные игнорируются (в skipped не возвращаются).
//
//   - DeleteAll:
//
//   - Удаляет только те записи, у которых включён текущий слой;
//
//   - Остальные игнорируются.
//
// Под капотом ServiceImpl использует клиента CacheProvider (BatchGet, BatchPut, BatchDelete).
// TTL для записи вычисляется на основе конфигурации слоя через configService.
//
// Использование:
//   - ServiceImpl создаётся через фабрику createService()
//   - В случае отключённого слоя вместо ServiceImpl создаётся ServiceDisabled.
//
// Это позволяет централизованно управлять включением/отключением слоёв без изменения клиентского кода.
type Service interface {
	GetAll(reqs []*dto.ResolvedCacheId) (*GetResult, error)
	PutAll(reqs []*dto.ResolvedCacheEntry) error
	DeleteAll(reqs []*dto.ResolvedCacheId) error
	Close() error
}

type GetResult struct {
	Hits    []*dto.ResolvedCacheHit
	Misses  []*dto.ResolvedCacheId
	Skipped []*dto.ResolvedCacheId
}

func CreateNewServiceList(providerConfigs []*config.LayerProvider, cacheServiceConfig config.CacheService) ([]Service, error) {
	services := make([]Service, 0, len(providerConfigs))

	for i, providerConfig := range providerConfigs {
		service, err := createService(providerConfig, cacheServiceConfig, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create service for provider index %d (name: %s): %w", i, providerConfig.Provider.GetName(), err)
		}
		services = append(services, service)
	}
	return services, nil
}

func createService(providerConfig *config.LayerProvider, cacheServiceConfig config.CacheService, level int) (Service, error) {
	if providerConfig.Mode == config.LayerModeDisabled {
		return &ServiceDisabled{}, nil
	}

	provider, err := initProvider(providerConfig.Provider)
	if err != nil {
		return nil, err
	}
	return &ServiceImpl{client: provider, configService: cacheServiceConfig, level: level}, nil
}

func initProvider(p interface{}) (CacheProvider, error) {
	switch c := p.(type) {
	case config.Ristretto:
		return NewRistretto(c)
	case config.Redis:
		return NewRedis(c)
	case config.RocksDB:
		return NewRocksDb(c)
	default:
		return nil, fmt.Errorf("unsupported provider type: %T", c)
	}
}

//////////////////////////
/// Concrete implementation
/////////////////////////

type ServiceImpl struct {
	client        CacheProvider
	configService config.CacheService
	level         int
}

// GetAll получает значения для ключей, у которых включён текущий слой.
// На выходе — разделение на hits/misses/skipped + возможная ошибка клиента
func (s *ServiceImpl) GetAll(reqs []*dto.ResolvedCacheId) (*GetResult, error) {
	keyToRequest, enabledKeys, skipped := s.categorizeRequests(reqs)
	if len(enabledKeys) == 0 {
		return &GetResult{Hits: []*dto.ResolvedCacheHit{}, Misses: []*dto.ResolvedCacheId{}, Skipped: skipped}, nil
	}

	values, err := s.client.BatchGet(enabledKeys)
	if err != nil {
		return nil, fmt.Errorf("BatchGet error: %w", err)
	}

	hits := make([]*dto.ResolvedCacheHit, 0, len(enabledKeys))
	misses := make([]*dto.ResolvedCacheId, 0, len(enabledKeys))
	for _, key := range enabledKeys {
		if val, ok := values[key]; ok {
			hits = append(hits, &dto.ResolvedCacheHit{
				ResolvedCacheEntry: &dto.ResolvedCacheEntry{
					ResolvedCacheId: keyToRequest[key],
					Value:           json.RawMessage(val),
				},
				Found: true,
			})
		} else {
			misses = append(misses, keyToRequest[key])
		}
	}
	return &GetResult{
		Hits:    hits,
		Misses:  misses,
		Skipped: skipped,
	}, nil
}

// PutAll сохраняет все значения в слой, если он включён для соответствующего CacheId.
// Пропускает записи с отключённым слоем. Возвращает ошибку, если BatchPut не удался.
func (s *ServiceImpl) PutAll(reqs []*dto.ResolvedCacheEntry) (err error) {
	entries := make(map[string]string, len(reqs))
	ttls := make(map[string]int64, len(reqs))
	for _, req := range reqs {
		if !s.isEnabled(req) {
			continue
		}
		key := req.GetStorageKey()
		entries[key] = string(req.Value)
		ttls[key] = s.getTtl(req)
	}
	if len(entries) == 0 {
		return
	}
	return s.client.BatchPut(entries, ttls)
}

// DeleteAll удаляет все значения, у которых включён текущий слой.
// Пропускает отключённые. Возвращает ошибку, если удаление не удалось.
func (s *ServiceImpl) DeleteAll(reqs []*dto.ResolvedCacheId) (err error) {
	_, keys, _ := s.categorizeRequests(reqs)
	if len(keys) == 0 {
		return
	}

	return s.client.BatchDelete(keys)
}

func (s *ServiceImpl) Close() error {
	return s.client.Close()
}

// categorizeRequests отбирает только те запросы, у которых включён слой.
// Возвращает мапу key→req и список ключей.
func (s *ServiceImpl) categorizeRequests(reqs []*dto.ResolvedCacheId) (keyToRequest map[string]*dto.ResolvedCacheId, enabledKeys []string, skipped []*dto.ResolvedCacheId) {
	keyToRequest = make(map[string]*dto.ResolvedCacheId, len(reqs))
	enabledKeys = make([]string, 0, len(reqs))
	skipped = make([]*dto.ResolvedCacheId, 0, len(reqs))
	for _, req := range reqs {
		if s.isEnabled(req) {
			key := req.GetStorageKey()
			keyToRequest[key] = req
			enabledKeys = append(enabledKeys, key)
		} else {
			skipped = append(skipped, req)
		}
	}
	return
}

func (s *ServiceImpl) getTtl(cacheId dto.CacheIdRef) int64 {
	cache := s.configService.GetCacheByCacheId(cacheId)
	return cache.Layers[s.level].TTL.Milliseconds()
}

func (s *ServiceImpl) isEnabled(cacheId dto.CacheIdRef) bool {
	cache := s.configService.GetCacheByCacheId(cacheId)
	return cache.Layers[s.level].Enabled
}

//////////////////////////
/// DisabledService
/////////////////////////

type ServiceDisabled struct {
}

func (s *ServiceDisabled) GetAll(reqs []*dto.ResolvedCacheId) (*GetResult, error) {
	return &GetResult{
			Hits:    []*dto.ResolvedCacheHit{},
			Misses:  []*dto.ResolvedCacheId{},
			Skipped: reqs,
		},
		nil
}

func (s *ServiceDisabled) PutAll(reqs []*dto.ResolvedCacheEntry) error {
	return nil
}
func (s *ServiceDisabled) DeleteAll(reqs []*dto.ResolvedCacheId) error {
	return nil
}

func (s *ServiceDisabled) Close() error {
	return nil
}
