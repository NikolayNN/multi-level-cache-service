package cmn

type CacheNameKey interface {
	GetCacheName() string
	GetKey() string
}

type CacheReqResolvedI interface {
	CacheNameKey
	CacheKey() string
}
