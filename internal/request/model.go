package request

type GetCacheReqI interface {
	CacheName() string
	Key() string
}

type ResolvedCacheReqI interface {
	GetCacheReqI
	CacheKey() string
}

type GetCacheReq struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
}

type ResolvedGetCacheReq struct {
	req      *GetCacheReq
	cacheKey string
}

func (r *ResolvedGetCacheReq) CacheName() string {
	return r.req.CacheName
}

func (r *ResolvedGetCacheReq) Key() string {
	return r.req.Key
}

func (r *ResolvedGetCacheReq) CacheKey() string {
	return r.cacheKey
}

type GetCacheResp struct {
	req   *ResolvedGetCacheReq
	value string
	found bool
}

func (r *GetCacheResp) CacheName() string {
	return r.req.CacheName()
}

func (r *GetCacheResp) Key() string {
	return r.req.Key()
}

func (r *GetCacheResp) CacheKey() string {
	return r.req.CacheName()
}
