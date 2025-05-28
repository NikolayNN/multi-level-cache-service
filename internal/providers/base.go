package providers

type CacheProvider interface {
	BatchGet(keys []string) (map[string]string, error)
	BatchPut(items map[string]string, ttls map[string]uint) error
	BatchDelete(keys []string) error

	Close() error
}

type ApiClient interface {
	Get(key string) (string, bool, error)
	BatchGet(keys []string) (map[string]string, error)
}
