package put

import (
	"aur-cache-service/internal/config"
)

type Resolver interface {
	resolve(reqs []*CacheReq) []*CacheReqResolved
}

type ResolverImpl struct {
	cacheConfigService config.CacheService
}

func NewResolverImpl(cacheConfigService config.CacheService) *ResolverImpl {
	return &ResolverImpl{
		cacheConfigService: cacheConfigService,
	}
}

func (r *ResolverImpl) resolve(reqs []*CacheReq, level int) (resolvedRequests []*CacheReqResolved) {
	for _, req := range reqs {
		resolvedReq := r.resolveSingle(req, level)
		if resolvedReq != nil {
			resolvedRequests = append(resolvedRequests, resolvedReq)
		}
	}
	return
}

func (r *ResolverImpl) resolveSingle(req *CacheReq, level int) (resolvedReq *CacheReqResolved) {
	levelOpts := r.cacheConfigService.GetByName(req.GetCacheName()).Layers[level]
	if !levelOpts.Enabled {
		return nil
	}
	resolvedReq = &CacheReqResolved{
		Req:      req,
		CacheKey: r.cacheConfigService.CacheKey(req),
		Ttl:      levelOpts.TTL,
	}
	return
}
