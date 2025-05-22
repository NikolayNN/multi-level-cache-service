package clients_test

import (
	"aur-cache-service/internal/config/clients"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var pathConfigFile = "testData/cache_test.yml"

func TestLoadConfig_Success(t *testing.T) {

	actual, _ := clients.LoadConfig(pathConfigFile)

	fmt.Printf("%+v\n", actual)

	expected := clients.Config{
		Ristretto: clients.RistrettoConfig{
			Enabled:     true,
			NumCounters: 1_000_000,
			BufferItems: 64,
			MaxCost:     "64MiB",
			DefaultTTL:  15 * time.Second,
		},
		Redis: clients.RedisConfig{
			Enabled:  true,
			Host:     "localhost",
			Port:     6370,
			Password: "12345",
			DB:       0,
			PoolSize: 10,
			Timeout:  5 * time.Second,
		},
		RocksDB: clients.RocksDBConfig{
			Enabled:         true,
			Path:            "/path",
			CreateIfMissing: true,
			MaxOpenFiles:    100,
			BlockSize:       "60MiB",
			BlockCache:      "61MiB",
			WriteBufferSize: "62MiB",
		},
	}

	require.Equal(t, &expected, actual)
}
