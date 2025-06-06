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

	cache, err := service.GetCache(id)
	assert.NoError(t, err)
	assert.Equal(t, "test", cache.Name)

	cache2, err := service.GetCacheByName("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", cache2.Name)

	prefix, err := service.GetPrefix(id)
	assert.NoError(t, err)
	assert.Equal(t, "t", prefix)

	ttl0, err := service.GetTtl(id, 0)
	assert.NoError(t, err)
	assert.Equal(t, time.Second, ttl0)

	ttl1, err := service.GetTtl(id, 1)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl1)

	enabled0, err := service.IsLevelEnabled(id, 0)
	assert.NoError(t, err)
	assert.True(t, enabled0)

	enabled1, err := service.IsLevelEnabled(id, 1)
	assert.NoError(t, err)
	assert.False(t, enabled1)
}

func TestCacheService_Getters_Invalid(t *testing.T) {
	cfg := &AppConfig{
		Caches: []Cache{},
	}

	service := NewCacheService(cfg)
	id := testCacheId("unknown")

	_, err := service.GetCache(id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = service.GetCacheByName("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = service.GetPrefix(id)
	assert.Error(t, err)

	_, err = service.GetTtl(id, 0)
	assert.Error(t, err)

	_, err = service.GetTtl(id, 100)
	assert.Error(t, err)

	_, err = service.IsLevelEnabled(id, 0)
	assert.Error(t, err)

	_, err = service.IsLevelEnabled(id, 100)
	assert.Error(t, err)
}
