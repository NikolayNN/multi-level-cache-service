package redisClient_test

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"aur-cache-service/internal/clients/redisClient"
)

func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, *redisClient.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	port, err := strconv.Atoi(mr.Port())

	client, err := redisClient.New(redisClient.Config{
		Host:     mr.Host(),
		Port:     port,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  time.Second * 5,
	})
	require.NoError(t, err)

	return mr, client
}

func TestNew(t *testing.T) {
	// Проверка успешного создания клиента
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	port, err := strconv.Atoi(mr.Port())
	client, err := redisClient.New(redisClient.Config{
		Host:     mr.Host(),
		Port:     port,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  time.Second * 5,
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Проверка с неправильным адресом
	_, err = redisClient.New(redisClient.Config{
		Host:     "non_existent_host",
		Port:     12345,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  time.Second * 1, // Маленький таймаут для быстрого теста
	})
	assert.Error(t, err)
}

func TestClient_Get(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Тест на отсутствующий ключ
	val, exists, err := client.Get("non_existent_key")
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Empty(t, val)

	// Тест на успешное получение
	mr.Set("test_key", "test_value")
	val, exists, err = client.Get("test_key")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "test_value", val)
}

func TestClient_Put(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Тест на успешное сохранение без TTL
	err := client.Put("test_key", "test_value", 0)
	assert.NoError(t, err)

	val, err := mr.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", val)

	// Проверка, что TTL не установлен
	ttl := mr.TTL("test_key")
	assert.Equal(t, time.Duration(0), ttl) // 0 означает, что TTL не установлен

	// Тест на успешное сохранение с TTL
	err = client.Put("test_key_ttl", "test_value_ttl", 60)
	assert.NoError(t, err)

	val, err = mr.Get("test_key_ttl")
	assert.NoError(t, err)
	assert.Equal(t, "test_value_ttl", val)

	// Проверка, что TTL установлен
	ttl = mr.TTL("test_key_ttl")
	assert.True(t, ttl > 0) // TTL должен быть положительным
}

func TestClient_Delete(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Подготовка данных
	mr.Set("test_key", "test_value")

	// Тест на успешное удаление существующего ключа
	deleted, err := client.Delete("test_key")
	assert.NoError(t, err)
	assert.True(t, deleted)
	assert.False(t, mr.Exists("test_key"))

	// Тест на удаление несуществующего ключа
	deleted, err = client.Delete("non_existent_key")
	assert.NoError(t, err)
	assert.False(t, deleted)
}

func TestClient_BatchGet(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Подготовка данных
	mr.Set("key1", "value1")
	mr.Set("key2", "value2")
	mr.Set("key3", "value3")

	// Тест на успешное получение нескольких значений
	result, err := client.BatchGet([]string{"key1", "key2", "key3", "non_existent_key"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}, result)

	// Проверка на пустой список ключей
	result, err = client.BatchGet([]string{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestClient_BatchPut(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Тест на успешное сохранение нескольких значений с разными TTL
	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	ttls := map[string]int{
		"key1": 60,
		"key2": 0, // Без TTL
		"key3": 120,
	}

	err := client.BatchPut(items, ttls)
	assert.NoError(t, err)

	// Проверка сохраненных значений
	for key, expectedValue := range items {
		val, err := mr.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, val)

		// Проверка TTL
		if ttl, exists := ttls[key]; exists && ttl > 0 {
			// TTL должен быть установлен
			assert.True(t, mr.TTL(key) > 0)
		} else {
			// TTL не должен быть установлен (0)
			assert.Equal(t, time.Duration(0), mr.TTL(key))
		}
	}

	// Проверка на пустой список элементов
	err = client.BatchPut(map[string]string{}, map[string]int{})
	assert.NoError(t, err)
}

func TestClient_BatchDelete(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Подготовка данных
	mr.Set("key1", "value1")
	mr.Set("key2", "value2")
	mr.Set("key3", "value3")

	// Тест на успешное удаление нескольких ключей
	err := client.BatchDelete([]string{"key1", "key2", "non_existent_key"})
	assert.NoError(t, err)

	// Проверка результатов удаления
	assert.False(t, mr.Exists("key1"))
	assert.False(t, mr.Exists("key2"))
	assert.True(t, mr.Exists("key3"))

	// Проверка на пустой список ключей
	err = client.BatchDelete([]string{})
	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()

	// Тест на закрытие соединения
	err := client.Close()
	assert.NoError(t, err)
}
