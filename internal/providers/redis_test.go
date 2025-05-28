package providers_test

import (
	"aur-cache-service/internal/config"
	"aur-cache-service/internal/providers"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, *providers.Redis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	port, err := strconv.Atoi(mr.Port())

	client, err := providers.NewRedis(config.Redis{
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
	client, err := providers.NewRedis(config.Redis{
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
	_, err = providers.NewRedis(config.Redis{
		Host:     "non_existent_host",
		Port:     12345,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  time.Second * 1, // Маленький таймаут для быстрого теста
	})
	assert.Error(t, err)
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

	ttls := map[string]uint{
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
	err = client.BatchPut(map[string]string{}, map[string]uint{})
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
