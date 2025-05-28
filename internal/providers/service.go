package providers

import (
	"aur-cache-service/internal/resolvers/get"
	"log"
)

type Service struct {
	client CacheProvider
}

func NewService(provider CacheProvider) *Service {
	return &Service{
		client: provider,
	}
}

func (s *Service) Get(req *get.CacheReqResolved) *get.CacheResp {
	v, isFound, _ := s.client.Get(req.GetCacheKey())
	return &get.CacheResp{
		Req:   req,
		Value: v,
		Found: isFound,
	}
}

func (s *Service) BatchGet(reqs []get.CacheReqResolved) []get.CacheResp {
	reqMap, reqKeys := s.resolvedReqsToMap(reqs)

	values, err := s.client.BatchGet(reqKeys)
	if err != nil {
		log.Printf("BatchGet error: %v", err)
		return make([]get.CacheResp, 0)
	}

	return s.collectGetCacheResponse(values, reqMap, reqKeys)
}

func (s *Service) resolvedReqsToMap(reqs []get.CacheReqResolved) (
	reqMap map[string]*get.CacheReqResolved,
	reqKeys []string,
) {
	reqMap = make(map[string]*get.CacheReqResolved, len(reqs))
	reqKeys = make([]string, 0, len(reqs))
	for i := range reqs {
		req := &reqs[i]
		key := req.GetCacheKey()
		reqMap[key] = req
		reqKeys = append(reqKeys, key)
	}
	return
}

func (s *Service) collectGetCacheResponse(values map[string]string, reqMap map[string]*get.CacheReqResolved, reqKeys []string) (responses []get.CacheResp) {
	responses = make([]get.CacheResp, 0, len(reqMap))
	for _, key := range reqKeys {
		value, found := values[key]
		responses = append(responses, get.CacheResp{
			Req:   reqMap[key],
			Value: value,
			Found: found,
		})
	}
	return
}

func (s *Service) Close() error {
	return s.client.Close()
}
