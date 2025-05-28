package cmn

import (
	"aur-cache-service/internal/config"
)

type ResolverService struct {
	cacheConfigService *config.CacheServiceImpl
}

func NewResolverService(cacheConfigService *config.CacheServiceImpl) *ResolverService {
	return &ResolverService{cacheConfigService: cacheConfigService}
}

func (s *ResolverService) Resolve(requests []*CacheReq) []*CacheReqResolved {
	resolved := make([]*CacheReqResolved, 0, len(requests))
	for _, req := range requests {
		cacheReq, err := s.toResolved(req)
		if err == nil {
			resolved = append(resolved, cacheReq)
		}
	}
	return resolved
}

func (s *ResolverService) toResolved(req *CacheReq) (*CacheReqResolved, error) {

	return &CacheReqResolved{
		Req:      req,
		CacheKey: s.cacheConfigService.CacheKey(req),
	}, nil
}
