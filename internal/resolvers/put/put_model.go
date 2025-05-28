package put

import "time"

type CacheReq struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
	Value     string `json:"v"`
}

func (r CacheReq) GetCacheName() string {
	return r.CacheName
}

func (r CacheReq) GetKey() string {
	return r.Key
}

type CacheReqResolved struct {
	Req      *CacheReq
	CacheKey string
	Ttl      time.Duration
}

func (c *CacheReqResolved) GetValue() string {
	return c.Req.Value
}
