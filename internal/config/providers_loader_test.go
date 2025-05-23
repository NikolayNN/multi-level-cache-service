package config_test

import (
	"aur-cache-service/internal/config"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var pathConfigFile = "testData/providers_config_test.yml"

func TestLoadConfig_Success(t *testing.T) {

	actual, _ := config.LoadProviders(pathConfigFile)

	for _, p := range actual.Providers {
		fmt.Printf("%+v\n", p)
	}

	expectedl0 := &config.Ristretto{
		ProviderMeta: config.ProviderMeta{
			Name: "ristretto-l0",
			Type: "ristretto",
		},
		NumCounters: 1_000_000,
		BufferItems: 64,
		MaxCost:     "64MiB",
		DefaultTTL:  15 * time.Second,
	}

	assert.Equal(t, expectedl0, actual.Providers[0])

	expectedl1 := &config.Redis{
		ProviderMeta: config.ProviderMeta{
			Name: "redis-l1",
			Type: "redis",
		},
		Host:     "localhost",
		Port:     6370,
		Password: "12345",
		DB:       0,
		PoolSize: 10,
		Timeout:  5 * time.Second,
	}

	assert.Equal(t, expectedl1, actual.Providers[1])

	expectedl2 := &config.RocksDB{
		ProviderMeta: config.ProviderMeta{
			Name: "rocksdb-l2",
			Type: "rocksdb",
		},
		Path:            "/path",
		CreateIfMissing: true,
		MaxOpenFiles:    100,
		BlockSize:       "60MiB",
		BlockCache:      "61MiB",
		WriteBufferSize: "62MiB",
	}

	assert.Equal(t, expectedl2, actual.Providers[2])
}
