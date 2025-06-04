package providers

import "time"

type CacheProvider interface {
	BatchGet(keys []string) (map[string]string, error)
	BatchPut(items map[string]string, ttls map[string]time.Duration) error
	BatchDelete(keys []string) error

	Close() error
}
