package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testCacheId string

func (t testCacheId) GetCacheName() string {
	return string(t)
}

func TestCacheService_Getters_Valid(t *testing.T) {
	cfg := &AppConfig{
		Caches: []Cache{
			{
				Name:   "test",
				Prefix: "t",
				Layers: []CacheLayerConfig{
					{Enabled: true, TTL: time.Second},
					{Enabled: false, TTL: 0},
				},
			},
		},
	}

	service := NewCacheService(cfg)
	id := testCacheId("test")

	cache, ok := service.GetCache(id)
	assert.True(t, ok)
	assert.Equal(t, "test", cache.Name)

	cache2, ok := service.GetCacheByName("test")
	assert.True(t, ok)
	assert.Equal(t, "test", cache2.Name)

	prefix, ok := service.GetPrefix(id)
	assert.True(t, ok)
	assert.Equal(t, "t", prefix)

	ttl0, ok := service.GetTtl(id, 0)
	assert.True(t, ok)
	assert.Equal(t, time.Second, ttl0)

	ttl1, ok := service.GetTtl(id, 1)
	assert.True(t, ok)
	assert.Equal(t, time.Duration(0), ttl1)

	enabled0, ok := service.IsLevelEnabled(id, 0)
	assert.True(t, ok)
	assert.True(t, enabled0)

	enabled1, ok := service.IsLevelEnabled(id, 1)
	assert.True(t, ok)
	assert.False(t, enabled1)
}

func TestCacheService_Getters_Invalid(t *testing.T) {
	cfg := &AppConfig{
		Caches: []Cache{},
	}

	service := NewCacheService(cfg)
	id := testCacheId("unknown")

	_, ok := service.GetCache(id)
	assert.False(t, ok)

	_, ok = service.GetCacheByName("unknown")
	assert.False(t, ok)

	_, ok = service.GetPrefix(id)
	assert.False(t, ok)

	_, ok = service.GetTtl(id, 0)
	assert.False(t, ok)

	_, ok = service.GetTtl(id, 100)
	assert.False(t, ok)

	_, ok = service.IsLevelEnabled(id, 0)
	assert.False(t, ok)

	_, ok = service.IsLevelEnabled(id, 100)
	assert.False(t, ok)
}
