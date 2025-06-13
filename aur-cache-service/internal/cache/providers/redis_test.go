package providers

import (
	"aur-cache-service/internal/cache/config"

	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func setupTestRedis(t *testing.T) (*Redis, func()) {
	srv, err := miniredis.Run()
	assert.NoError(t, err)

	port, _ := strconv.Atoi(srv.Port())

	cfg := config.Redis{
		Host:     srv.Host(),
		Port:     port,
		DB:       0,
		Password: "",
		PoolSize: 5,
		Timeout:  time.Second,
	}

	ctx := context.Background()
	redis, err := NewRedis(ctx, cfg)
	assert.NoError(t, err)

	return redis, func() {
		_ = redis.Close()
		srv.Close()
	}
}

func TestRedis_BatchPut_And_BatchGet(t *testing.T) {
	r, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	// Сохраняем два ключа
	items := map[string]string{
		"user:1": "Alice",
		"user:2": "Bob",
	}
	ttls := map[string]time.Duration{
		"user:1": 10 * time.Second,
		"user:2": 0, // без TTL
	}

	err := r.BatchPut(ctx, items, ttls)
	assert.NoError(t, err)

	// Читаем оба и несуществующий
	result, err := r.BatchGet(ctx, []string{"user:1", "user:2", "user:404"})
	assert.NoError(t, err)

	assert.Equal(t, "Alice", result["user:1"])
	assert.Equal(t, "Bob", result["user:2"])
	_, exists := result["user:404"]
	assert.False(t, exists)
}

func TestRedis_BatchDelete(t *testing.T) {
	r, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	_ = r.BatchPut(ctx, map[string]string{
		"session:1": "token",
		"session:2": "secret",
	}, nil)

	// Удаляем один
	err := r.BatchDelete(ctx, []string{"session:1"})
	assert.NoError(t, err)

	result, err := r.BatchGet(ctx, []string{"session:1", "session:2"})
	assert.NoError(t, err)
	_, exists := result["session:1"]
	assert.False(t, exists)
	assert.Equal(t, "secret", result["session:2"])
}
