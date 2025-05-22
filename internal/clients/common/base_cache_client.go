package common

type CacheClient interface {
	Get(key string) (string, bool, error)
	Put(key string, value string, ttl int) error
	Delete(key string) (bool, error)

	BatchGet(keys []string) (map[string]string, error)
	BatchPut(items map[string]string, ttls map[string]int) error
	BatchDelete(keys []string) error

	Close() error
}
