package ristrettoClient

import (
	"github.com/dgraph-io/ristretto"
	"time"
)

type Config struct {
	MaxCost     int64
	NumCounters int64
	BufferItems int64
}

type Client struct {
	cache *ristretto.Cache
}

func New(cfg Config) (*Client, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		cache: cache,
	}, nil
}

func (c *Client) Get(key string) (string, bool, error) {
	val, found := c.cache.Get(key)
	if !found {
		return "", false, nil
	}

	strVal, ok := val.(string)
	if !ok {
		return "", false, nil
	}

	return strVal, true, nil
}

func (c *Client) Put(key string, value string, ttl int) error {
	cost := int64(len(value))

	var success bool
	if ttl > 0 {
		expiration := time.Duration(ttl) * time.Millisecond
		success = c.cache.SetWithTTL(key, value, cost, expiration)
	} else {
		success = c.cache.Set(key, value, cost)
	}

	if !success {
		// Не попал в кэш (например, не прошёл через буфер или вытеснен)
	}
	return nil
}

func (c *Client) Delete(key string) (bool, error) {
	c.cache.Del(key)
	return true, nil
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

func (c *Client) BatchPut(items map[string]string, ttls map[string]int) error {
	for key, val := range items {
		var expiration time.Duration
		if ttl, ok := ttls[key]; ok && ttl > 0 {
			expiration = time.Duration(ttl) * time.Second
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
