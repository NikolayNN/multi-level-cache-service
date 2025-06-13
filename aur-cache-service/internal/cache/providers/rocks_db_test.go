package providers

import (
	"aur-cache-service/internal/cache/config"
	"context"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

func TestRocksDbCF_BatchPutGetDelete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "db")

	cfg := config.RocksDB{
		Path:            dbPath,
		CreateIfMissing: true,
		MaxOpenFiles:    100,
		BlockCache:      "512KB",
		BlockSize:       "4KB",
	}

	client, err := NewRocksDbCF(cfg)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	items := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	ttls := map[string]time.Duration{
		"foo": time.Minute,
		"baz": time.Minute,
	}

	// Put
	err = client.BatchPut(ctx, items, ttls)
	assert.NoError(t, err)

	// Get
	result, err := client.BatchGet(ctx, []string{"foo", "baz", "missing"})
	assert.NoError(t, err)
	assert.Equal(t, "bar", result["foo"])
	assert.Equal(t, "qux", result["baz"])
	_, found := result["missing"]
	assert.False(t, found)

	// Delete
	err = client.BatchDelete(ctx, []string{"foo"})
	assert.NoError(t, err)

	// Get after delete
	result, err = client.BatchGet(ctx, []string{"foo", "baz"})
	assert.NoError(t, err)
	_, found = result["foo"]
	assert.False(t, found)
	assert.Equal(t, "qux", result["baz"])
}

func TestRocksDbCF_TTLExpiration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "db")

	cfg := config.RocksDB{
		Path:            dbPath,
		CreateIfMissing: true,
		MaxOpenFiles:    100,
		BlockCache:      "512KB",
		BlockSize:       "4KB",
	}

	client, err := NewRocksDbCF(cfg)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	items := map[string]string{
		"expiring":   "soon",
		"persistent": "forever",
	}
	ttls := map[string]time.Duration{
		"expiring":   2 * time.Second,
		"persistent": time.Minute,
	}

	// Put values
	err = client.BatchPut(ctx, items, ttls)
	assert.NoError(t, err)

	// Immediate check
	result, err := client.BatchGet(ctx, []string{"expiring", "persistent"})
	assert.NoError(t, err)
	assert.Equal(t, "soon", result["expiring"])
	assert.Equal(t, "forever", result["persistent"])

	// Wait for TTL to expire
	time.Sleep(3 * time.Second)

	// Check again
	result, err = client.BatchGet(ctx, []string{"expiring", "persistent"})
	assert.NoError(t, err)

	_, found := result["expiring"]
	assert.False(t, found, "expiring key should be evicted")

	val, found := result["persistent"]
	assert.True(t, found)
	assert.Equal(t, "forever", val)
}
