package providers

type CacheProvider interface {
	BatchGet(keys []string) (map[string]string, error)
	BatchPut(items map[string]string, ttls map[string]int64) error
	BatchDelete(keys []string) error

	Close() error
}
