package providers

import (
	"aur-cache-service/internal/cache/config"
	"aur-cache-service/internal/metrics"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"go.uber.org/zap"
	"telegram-alerts-go/alert"
)

type Redis struct {
	rdb *redis.Client
}

const (
	minChunk = 400
	maxChunk = 500
)

func NewRedis(ctx context.Context, cfg config.Redis) (*Redis, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	// Connection check
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	zap.S().Infow("connected to Redis", "host", cfg.Host, "port", cfg.Port)

	return &Redis{
		rdb: rdb,
	}, nil
}

// BatchGet получает несколько значений за один запрос, разбивая их на chunk'и
func (c *Redis) BatchGet(ctx context.Context, keys []string) (result map[string]string, err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("redis", "get", time.Since(start).Seconds())
		metrics.RecordProviderOp("redis", "get", err)
	}()

	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	result = make(map[string]string, len(keys))
	chunks := splitKeysToChunks(keys, minChunk, maxChunk)

	for _, chunk := range chunks {
		vals, err := c.rdb.MGet(ctx, chunk...).Result()
		if err != nil {
			return nil, fmt.Errorf("ошибка пакетного получения из Redis: %w", err)
		}

		// Обрабатываем результаты текущего chunk'а
		for i, key := range chunk {
			if i < len(vals) && vals[i] != nil {
				if str, ok := vals[i].(string); ok {
					result[key] = str
				}
			}
		}
	}
	return result, nil
}

// BatchPut сохраняет несколько значений за один запрос, разбивая их на chunk'и
func (c *Redis) BatchPut(ctx context.Context, items map[string]string, ttls map[string]time.Duration) (err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("redis", "put", time.Since(start).Seconds())
		metrics.RecordProviderOp("redis", "put", err)
	}()

	if len(items) == 0 {
		return nil
	}

	chunks := splitKeyValueToChunks(items, minChunk, maxChunk)

	for chunkIndex, chunk := range chunks {
		pipe := c.rdb.Pipeline()

		for key, value := range chunk {
			var expiration time.Duration
			if ttl, exists := ttls[key]; exists && ttl > 0 {
				expiration = ttl
			}
			pipe.Set(ctx, key, value, expiration)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			zap.S().Errorw(alert.Prefix("redis pipeline exec error"), "chunk", chunkIndex, "error", err)
			return fmt.Errorf("ошибка пакетного сохранения в Redis (chunk %d): %w", chunkIndex, err)
		}
	}
	return nil
}

// BatchDelete удаляет несколько значений за один запрос, разбивая их на chunk'и
func (c *Redis) BatchDelete(ctx context.Context, keys []string) (err error) {
	start := time.Now()
	defer func() {
		metrics.RecordProviderLatency("redis", "delete", time.Since(start).Seconds())
		metrics.RecordProviderOp("redis", "delete", err)
	}()

	if len(keys) == 0 {
		return nil
	}

	chunks := splitKeysToChunks(keys, minChunk, maxChunk)

	// Обрабатываем каждый chunk отдельно
	for chunkIndex, chunk := range chunks {
		_, err = c.rdb.Unlink(ctx, chunk...).Result()
		if err != nil {
			zap.S().Errorw(alert.Prefix("redis unlink error"), "chunk", chunkIndex+1, "total", len(chunks), "error", err)
			return fmt.Errorf("ошибка пакетного удаления из Redis (chunk %d/%d, keys: %d): %w",
				chunkIndex+1, len(chunks), len(chunk), err)
		}
	}
	return nil
}

func (c *Redis) Close() error {
	return c.rdb.Close()
}
