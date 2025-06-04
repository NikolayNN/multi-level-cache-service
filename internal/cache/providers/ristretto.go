package providers

import (
	"aur-cache-service/internal/cache/config"
	"github.com/dgraph-io/ristretto"
	"time"
)

type Client struct {
	cache *ristretto.Cache
}

func NewRistretto(cfg config.Ristretto) (*Client, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     int64(cfg.MaxCostBytes()),
		BufferItems: cfg.BufferItems,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		cache: cache,
	}, nil
}

func (c *Client) BatchGet(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		val, ok := c.cache.Get(key)
		if ok {
			if strVal, castOk := val.(string); castOk {
				result[key] = strVal
			}
		}
	}
	return result, nil
}

func (c *Client) BatchPut(items map[string]string, ttls map[string]time.Duration) error {
	for key, val := range items {
		var expiration time.Duration
		if ttl, ok := ttls[key]; ok && ttl > 0 {
			expiration = ttl
		}
		c.cache.SetWithTTL(key, val, int64(len(val)), expiration)
	}
	return nil
}

func (c *Client) BatchDelete(keys []string) error {
	for _, key := range keys {
		c.cache.Del(key)
	}
	return nil
}

func (c *Client) Close() error {
	c.cache.Close()
	return nil
}
