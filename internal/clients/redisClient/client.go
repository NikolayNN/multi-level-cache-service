package redisClient

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
	Timeout  time.Duration
}

type Client struct {
	rdb *redis.Client
	ctx context.Context
}

func New(cfg Config) (*Client, error) {
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

	return &Client{
		rdb: rdb,
		ctx: ctx,
	}, nil
}

// Get получает значение по ключу из Redis (string версия)
func (c *Client) Get(key string) (string, bool, error) {
	val, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Ключ не найден
			return "", false, nil
		}
		return "", false, fmt.Errorf("ошибка получения значения из Redis: %w", err)
	}
	return val, true, nil
}

// Put сохраняет значение по ключу в Redis с опциональным TTL (string версия)
func (c *Client) Put(key string, value string, ttl int) error {
	var expiration time.Duration
	if ttl > 0 {
		expiration = time.Duration(ttl) * time.Second
	}

	err := c.rdb.Set(c.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения в Redis: %w", err)
	}
	return nil
}

// Delete удаляет значение по ключу из Redis
func (c *Client) Delete(key string) (bool, error) {
	res, err := c.rdb.Del(c.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("error delete value from Redis: %w", err)
	}
	// Возвращаем true, если ключ был найден и удален
	return res > 0, nil
}

// BatchGet получает несколько значений за один запрос
func (c *Client) BatchGet(keys []string) (map[string]string, error) {
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
func (c *Client) BatchPut(items map[string]string, ttls map[string]int) error {
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
func (c *Client) BatchDelete(keys []string) error {
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

func (c *Client) Close() error {
	return c.rdb.Close()
}
