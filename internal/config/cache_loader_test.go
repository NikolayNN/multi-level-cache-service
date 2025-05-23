package config_test

import (
	"aur-cache-service/internal/config"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var pathCacheConfigFile = "testData/cache_test.yml"

func TestLoadCache_Success(t *testing.T) {

	cfg, err := config.LoadCacheStorage(pathCacheConfigFile)

	fmt.Printf("%+v\n", cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Configs, 2)

	expected := &config.CacheStorage{
		Configs: map[string]config.Cache{
			"user": {
				Name:   "user",
				Prefix: "u",
				Layers: []config.CacheLayerConfig{
					{Enabled: true, TTL: 30 * time.Second},
					{Enabled: true, TTL: 10 * time.Minute},
					{Enabled: true, TTL: 6 * time.Hour},
				},
				Api: config.ApiConfig{
					Enabled: true,
					GetBatch: config.ApiBatchConfig{
						URL:  "localhost:8080/user",
						Prop: "id",
					},
				},
			},
			"order": {
				Name:   "order",
				Prefix: "o",
				Layers: []config.CacheLayerConfig{
					{Enabled: true, TTL: 60 * time.Second},
					{Enabled: true, TTL: 20 * time.Minute},
					{Enabled: true, TTL: 12 * time.Hour},
				},
				Api: config.ApiConfig{
					Enabled: true,
					GetBatch: config.ApiBatchConfig{
						URL:  "localhost:8080/order",
						Prop: "id",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, cfg)
}
