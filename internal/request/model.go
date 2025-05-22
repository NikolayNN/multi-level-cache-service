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
	Req      *GetCacheReq
	CacheKey string
}

func (r *ResolvedGetCacheReq) CacheName() string {
	return r.Req.CacheName
}

func (r *ResolvedGetCacheReq) Key() string {
	return r.Req.Key
}

func (r *ResolvedGetCacheReq) GetCacheKey() string {
	return r.CacheKey
}

type GetCacheResp struct {
	Req   *ResolvedGetCacheReq
	Value string
	Found bool
}

func (r *GetCacheResp) CacheName() string {
	return r.Req.CacheName()
}

func (r *GetCacheResp) Key() string {
	return r.Req.Key()
}

func (r *GetCacheResp) CacheKey() string {
	return r.Req.CacheName()
}
