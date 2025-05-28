package get

import "aur-cache-service/internal/resolvers/cmn"

type CacheResp struct {
	Req   *cmn.CacheReqResolved
	Value string
	Found bool
}

func (r *CacheResp) GetCacheName() string {
	return r.Req.GetCacheName()
}

func (r *CacheResp) GetKey() string {
	return r.Req.GetKey()
}

func (r *CacheResp) CacheKey() string {
	return r.Req.GetCacheName()
}
