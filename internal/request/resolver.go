package request

import (
	"aur-cache-service/internal/prefix"
)

type Service struct {
	prefixService *prefix.Service
}

func NewResolverService(prefixService *prefix.Service) *Service {
	return &Service{prefixService: prefixService}
}

func (s *Service) Resolve(requests []GetCacheReq) []ResolvedGetCacheReq {
	resolved := make([]ResolvedGetCacheReq, 0, len(requests))
	for _, req := range requests {
		cacheReq, err := s.toResolved(&req)
		if err == nil {
			resolved = append(resolved, cacheReq)
		}
	}
	return resolved
}

func (s *Service) toResolved(req *GetCacheReq) (ResolvedGetCacheReq, error) {

	return ResolvedGetCacheReq{
		req:      req,
		cacheKey: s.prefixService.ToCacheKey(req.CacheName, req.Key),
	}, nil
}
