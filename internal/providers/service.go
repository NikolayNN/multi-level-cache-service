package providers

import (
	"aur-cache-service/internal/config"
	"aur-cache-service/internal/resolvers/cmn"
	"aur-cache-service/internal/resolvers/get"
	"aur-cache-service/internal/resolvers/put"
	"fmt"
	"log"
)

func СreateNewServiceList(providerConfigs []*config.LayerProvider) ([]Service, error) {
	services := make([]Service, 0, len(providerConfigs))

	for i, providerConfig := range providerConfigs {
		service, err := createNewService(providerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create service for provider index %d (name: %s): %w", i, providerConfig.Provider.GetName(), err)
		}
		services = append(services, service)
	}
	return services, nil
}

func createNewService(providerConfig *config.LayerProvider) (Service, error) {
	if providerConfig.Mode == config.LayerModeDisabled {
		return &ServiceDisabled{}, nil
	}

	var cacheProvider CacheProvider
	var err error

	switch cfg := providerConfig.Provider.(type) {
	case config.Ristretto:
		cacheProvider, err = NewRistretto(cfg)
	case config.Redis:
		cacheProvider, err = NewRedis(cfg)
	case config.RocksDB:
		cacheProvider, err = NewRocksDb(cfg)
	default:
		return nil, fmt.Errorf("unsupported cache provider type: %T", cfg.GetType())
	}

	if err != nil {
		return nil, err
	}
	if cacheProvider == nil {
		return nil, fmt.Errorf("cache provider is nil for: %v", providerConfig.Provider.GetName())
	}

	return NewService(cacheProvider), nil
}

type Service interface {
	GetAll(reqs []*cmn.CacheReqResolved) (success []*get.CacheResp, notFound []*cmn.CacheReqResolved)
	PutAll(reqs []*put.CacheReqResolved)
	DeleteAll(reqs []*cmn.CacheReqResolved)
	Close() error
}

type ServiceImpl struct {
	client CacheProvider
}

func NewService(provider CacheProvider) *ServiceImpl {
	return &ServiceImpl{
		client: provider,
	}
}

func (s *ServiceImpl) GetAll(reqs []*cmn.CacheReqResolved) (success []*get.CacheResp, notFound []*cmn.CacheReqResolved) {
	reqMap, reqKeys := s.resolvedReqsToMap(reqs)

	values, err := s.client.BatchGet(reqKeys)
	if err != nil {
		log.Printf("BatchGet error: %v", err)
		return nil, reqs
	}

	return s.collectGetCacheResponse(values, reqMap, reqKeys)
}

func (s *ServiceImpl) resolvedReqsToMap(reqs []*cmn.CacheReqResolved) (
	reqMap map[string]*cmn.CacheReqResolved,
	reqKeys []string,
) {
	reqMap = make(map[string]*cmn.CacheReqResolved, len(reqs))
	reqKeys = make([]string, 0, len(reqs))
	for i := range reqs {
		req := reqs[i]
		key := req.GetCacheKey()
		reqMap[key] = req
		reqKeys = append(reqKeys, key)
	}
	return
}

func (s *ServiceImpl) collectGetCacheResponse(values map[string]string, reqMap map[string]*cmn.CacheReqResolved, reqKeys []string) (success []*get.CacheResp, notFound []*cmn.CacheReqResolved) {
	for _, key := range reqKeys {
		value, found := values[key]
		if found {
			success = append(success, &get.CacheResp{
				Req:   reqMap[key],
				Value: value,
				Found: found,
			})
		} else {
			notFound = append(notFound, reqMap[key])
		}
	}
	return
}

func (s *ServiceImpl) PutAll(reqs []*put.CacheReqResolved) {
	items := make(map[string]string, len(reqs))
	ttls := make(map[string]uint, len(reqs))
	for _, req := range reqs {
		items[req.CacheKey] = req.GetValue()
		ttls[req.CacheKey] = uint(req.Ttl.Milliseconds())
	}
	err := s.client.BatchPut(items, ttls)
	if err != nil {
		log.Printf("BatchPut error: %v", err)
	}
}

func (s *ServiceImpl) DeleteAll(reqs []*cmn.CacheReqResolved) {
	_, reqKeys := s.resolvedReqsToMap(reqs)
	err := s.client.BatchDelete(reqKeys)
	if err != nil {
		log.Printf("BatchDelete error: %v", err)
	}
}

func (s *ServiceImpl) Close() error {
	return s.client.Close()
}

// ==== disabled ====

type ServiceDisabled struct {
}

func (s *ServiceDisabled) GetAll(reqs []*cmn.CacheReqResolved) (success []*get.CacheResp, notFound []*cmn.CacheReqResolved) {
	return []*get.CacheResp{}, reqs
}

func (s *ServiceDisabled) PutAll(reqs []*put.CacheReqResolved) {
}
func (s *ServiceDisabled) DeleteAll(reqs []*cmn.CacheReqResolved) {
}

func (s *ServiceDisabled) Close() error {
	return nil
}
