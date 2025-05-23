package providers

import (
	"aur-cache-service/internal/request"
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

func (s *Service) Get(req *request.ResolvedGetCacheReq) *request.GetCacheResp {
	v, isFound, _ := s.client.Get(req.GetCacheKey())
	return &request.GetCacheResp{
		Req:   req,
		Value: v,
		Found: isFound,
	}
}

func (s *Service) BatchGet(reqs []request.ResolvedGetCacheReq) []request.GetCacheResp {
	reqMap, reqKeys := s.resolvedReqsToMap(reqs)

	values, err := s.client.BatchGet(reqKeys)
	if err != nil {
		log.Printf("BatchGet error: %v", err)
		return make([]request.GetCacheResp, 0)
	}

	return s.collectGetCacheResponse(values, reqMap, reqKeys)
}

func (s *Service) resolvedReqsToMap(reqs []request.ResolvedGetCacheReq) (
	reqMap map[string]*request.ResolvedGetCacheReq,
	reqKeys []string,
) {
	reqMap = make(map[string]*request.ResolvedGetCacheReq, len(reqs))
	reqKeys = make([]string, 0, len(reqs))
	for i := range reqs {
		req := &reqs[i]
		key := req.GetCacheKey()
		reqMap[key] = req
		reqKeys = append(reqKeys, key)
	}
	return
}

func (s *Service) collectGetCacheResponse(values map[string]string, reqMap map[string]*request.ResolvedGetCacheReq, reqKeys []string) (responses []request.GetCacheResp) {
	responses = make([]request.GetCacheResp, 0, len(reqMap))
	for _, key := range reqKeys {
		value, found := values[key]
		responses = append(responses, request.GetCacheResp{
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
