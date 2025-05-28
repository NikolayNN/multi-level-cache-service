package cmn

type CacheNameKey interface {
	GetCacheName() string
	GetKey() string
}

type CacheReqResolvedI interface {
	CacheNameKey
	CacheKey() string
}

type CacheReq struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
}

func (r *CacheReq) GetCacheName() string {
	return r.CacheName
}

func (r *CacheReq) GetKey() string {
	return r.Key
}

type CacheReqResolved struct {
	Req      *CacheReq
	CacheKey string
}

func (r *CacheReqResolved) GetCacheName() string {
	return r.Req.CacheName
}

func (r *CacheReqResolved) GetKey() string {
	return r.Req.Key
}

func (r *CacheReqResolved) GetCacheKey() string {
	return r.CacheKey
}
