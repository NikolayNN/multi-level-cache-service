package providers

import (
	"aur-cache-service/internal/cache/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type Redis struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedis(cfg config.Redis) (*Redis, error) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	// Проверка соединения
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	return &Redis{
		rdb: rdb,
		ctx: ctx,
	}, nil
}

// BatchGet получает несколько значений за один запрос
func (c *Redis) BatchGet(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	result := make(map[string]string)

	vals, err := c.rdb.MGet(c.ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка пакетного получения из Redis: %w", err)
	}

	for i, key := range keys {
		if vals[i] != nil {
			if str, ok := vals[i].(string); ok {
				result[key] = str
			}
		}
	}

	return result, nil
}

// BatchPut сохраняет несколько значений за один запрос
func (c *Redis) BatchPut(items map[string]string, ttls map[string]int64) error {
	if len(items) == 0 {
		return nil
	}

	pipe := c.rdb.Pipeline()

	for key, value := range items {
		var expiration time.Duration
		if ttl, exists := ttls[key]; exists && ttl > 0 {
			expiration = time.Duration(ttl) * time.Second
		}
		pipe.Set(c.ctx, key, value, expiration)
	}

	_, err := pipe.Exec(c.ctx)
	if err != nil {
		return fmt.Errorf("ошибка пакетного сохранения в Redis: %w", err)
	}

	return nil
}

// BatchDelete удаляет несколько значений за один запрос
func (c *Redis) BatchDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Удаление всех ключей одной командой
	_, err := c.rdb.Del(c.ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("error batch delete event: %w", err)
	}

	return nil
}

func (c *Redis) Close() error {
	return c.rdb.Close()
}
