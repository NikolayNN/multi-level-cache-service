package providers_test

import (
	"aur-cache-service/internal/config"
	"aur-cache-service/internal/providers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newTestClient(t *testing.T) *providers.Client {
	client, err := providers.NewRistretto(config.Ristretto{
		NumCounters: 1000,
		MaxCost:     "20MiB",
		BufferItems: 64,
	})
	require.NoError(t, err)
	return client
}

func TestBatchPutAndBatchGet(t *testing.T) {
	client := newTestClient(t)

	items := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	ttls := map[string]uint{
		"key1": 1, // секунда
		"key2": 1,
	}

	err := client.BatchPut(items, ttls)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	got, err := client.BatchGet([]string{"key1", "key2", "missing"})
	require.NoError(t, err)
	require.Equal(t, "val1", got["key1"])
	require.Equal(t, "val2", got["key2"])
	_, ok := got["missing"]
	require.False(t, ok)
}
