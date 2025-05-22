package caches_test

import (
	"aur-cache-service/internal/config/caches"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var pathConfigFile = "testData/cache_test.yml"

func TestLoadConfig_Success(t *testing.T) {

	cfg, err := caches.LoadConfig(pathConfigFile)

	fmt.Printf("%+v\n", cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Configs, 2)

	actual, _ := cfg.Get("test")

	expected := caches.Config{
		Name:   "test",
		Prefix: "t",
		Levels: caches.LevelsConfig{
			L0: caches.CacheLevelConfig{
				Enabled: true,
				TTL:     60 * time.Second,
			},
			L1: caches.CacheLevelConfig{
				Enabled: true,
				TTL:     10 * time.Minute,
			},
			L2: caches.CacheLevelConfig{
				Enabled: true,
				TTL:     6 * time.Hour,
			},
		},
		Api: caches.ApiConfig{
			Enabled: true,
			GetBatch: caches.ApiBatchConfig{
				URL:  "localhost:8080/test",
				Prop: "key-id",
			},
		},
	}

	require.Equal(t, expected, actual)
}
