package providers

import (
	"aur-cache-service/internal/cache/config"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRistretto_BatchPutGetDelete(t *testing.T) {
	ctx := context.Background()
	client, err := NewRistretto(config.Ristretto{
		NumCounters: 1000,
		BufferItems: 64,
		MaxCost:     "1MB",
	})
	assert.NoError(t, err)
	defer client.Close()

	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	ttls := map[string]time.Duration{
		"key1": time.Minute,
		"key2": time.Minute,
	}

	// Put
	err = client.BatchPut(ctx, items, ttls)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Get
	result, err := client.BatchGet(ctx, []string{"key1", "key2", "missing"})
	assert.NoError(t, err)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	_, ok := result["missing"]
	assert.False(t, ok)

	// Delete
	err = client.BatchDelete(ctx, []string{"key1"})
	assert.NoError(t, err)

	// Ensure deletion
	result, err = client.BatchGet(ctx, []string{"key1", "key2"})
	assert.NoError(t, err)
	_, ok = result["key1"]
	assert.False(t, ok)
	assert.Equal(t, "value2", result["key2"])
}

func TestRistretto_ContextCancelDuringPut(t *testing.T) {
	client, _ := NewRistretto(config.Ristretto{
		NumCounters: 1000,
		BufferItems: 64,
		MaxCost:     "1MB",
	})
	defer client.Close()

	items := make(map[string]string)
	ttls := make(map[string]time.Duration)
	for i := 0; i < 500_000; i++ {
		key := fmt.Sprintf("key%d", i)
		items[key] = "value"
		ttls[key] = time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := client.BatchPut(ctx, items, ttls)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
