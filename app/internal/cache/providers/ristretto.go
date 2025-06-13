package providers

import (
	"aur-cache-service/internal/cache/config"
	"aur-cache-service/internal/metrics"
	"context"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"time"
)

type Client struct {
	cache *ristretto.Cache
}

const contextCheckInterval = 100

func NewRistretto(cfg config.Ristretto) (*Client, error) {
	maxCostBytes, _ := cfg.MaxCostBytes()
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     int64(maxCostBytes),
		BufferItems: cfg.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Ristretto кэш: %w", err)
	}

	return &Client{
		cache: cache,
	}, nil
}

func (c *Client) BatchGet(ctx context.Context, keys []string) (result map[string]string, err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("ristretto", "get", time.Since(start).Seconds())
		metrics.RecordProviderOp("ristretto", "get", err)
	}()

	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	result = make(map[string]string, len(keys))
	for i, key := range keys {

		// Проверяем контекст каждые 100 итераций
		if i%contextCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		val, ok := c.cache.Get(key)
		if ok {
			if strVal, castOk := val.(string); castOk {
				result[key] = strVal
			}
		}
	}
	return result, nil
}

func (c *Client) BatchPut(ctx context.Context, items map[string]string, ttls map[string]time.Duration) (err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("ristretto", "put", time.Since(start).Seconds())
		metrics.RecordProviderOp("ristretto", "put", err)
	}()

	if len(items) == 0 {
		return nil
	}

	count := 0

	for key, val := range items {

		// Проверяем контекст каждые 100 итераций
		if count%contextCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		count++

		var expiration time.Duration
		if ttl, ok := ttls[key]; ok && ttl > 0 {
			expiration = ttl
		}
		c.cache.SetWithTTL(key, val, int64(len(val)), expiration)
	}
	return nil
}

func (c *Client) BatchDelete(ctx context.Context, keys []string) (err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("ristretto", "delete", time.Since(start).Seconds())
		metrics.RecordProviderOp("ristretto", "delete", err)
	}()

	if len(keys) == 0 {
		return nil
	}

	for i, key := range keys {

		// Проверяем контекст каждые 100 итераций
		if i%contextCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		c.cache.Del(key)
	}
	return nil
}

func (c *Client) Close() error {
	if c.cache != nil {
		c.cache.Close()
		c.cache = nil
	}
	return nil
}
