package rocksDbClient_test

import (
	"aur-cache-service/internal/clients/rocksDbClient"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

// 1. docker build -f Dockerfile.test -t rocks-tests .
// 2. docker run --rm rocks-tests

// setupClient создает временную БД и возвращает клиента и функцию очистки
func setupClient(t *testing.T) (*rocksDbClient.Client, func()) {
	dir, err := ioutil.TempDir("", "rocksdb-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	cfg := rocksDbClient.Config{
		Path:            dir,
		CreateIfMissing: true,
	}
	client, err := rocksDbClient.New(cfg)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Не удалось создать клиент RocksDB: %v", err)
	}

	return client, func() {
		client.Close()
		os.RemoveAll(dir)
	}
}

func TestPutGet(t *testing.T) {
	fmt.Println("RUN TestPutGet")
	client, cleanup := setupClient(t)
	defer cleanup()

	err := client.Put("test-key", "test-value", 0)
	if err != nil {
		t.Fatalf("Put вернул ошибку: %v", err)
	}

	value, exists, err := client.Get("test-key")
	if err != nil {
		t.Fatalf("Get вернул ошибку: %v", err)
	}
	if !exists || value != "test-value" {
		t.Errorf("Ожидалось 'test-value', получено '%v', существует: %v", value, exists)
	}
}

func TestDelete(t *testing.T) {
	fmt.Println("RUN TestDelete")
	client, cleanup := setupClient(t)
	defer cleanup()

	err := client.Put("del-key", "to-delete", 0)
	if err != nil {
		t.Fatalf("Put вернул ошибку: %v", err)
	}

	// Удаляем существующий ключ
	deleted, err := client.Delete("del-key")
	if err != nil {
		t.Fatalf("Delete вернул ошибку: %v", err)
	}
	if !deleted {
		t.Errorf("Ожидалось true для удаления существующего ключа, получено %v", deleted)
	}

	// Проверяем, что ключ действительно удалён
	_, exists, err := client.Get("del-key")
	if err != nil {
		t.Fatalf("Get вернул ошибку после удаления: %v", err)
	}
	if exists {
		t.Error("Ожидалось, что ключ не существует после удаления")
	}

	// Удаляем несуществующий ключ
	deleted, err = client.Delete("missing-key")
	if err != nil {
		t.Fatalf("Delete вернул ошибку для несуществующего ключа: %v", err)
	}
	if deleted {
		t.Error("Ожидалось false при попытке удаления несуществующего ключа")
	}
}

func TestTTLKeyExpiration(t *testing.T) {
	fmt.Println("RUN TestTTLKeyExpiration")
	client, cleanup := setupClient(t)
	defer cleanup()

	// Устанавливаем TTL в 1 секунду
	err := client.Put("ttl-key", "temp-value", 1)
	if err != nil {
		t.Fatalf("Put вернул ошибку: %v", err)
	}

	// Непосредственно после сохранения значение должно существовать
	value, exists, err := client.Get("ttl-key")
	if err != nil {
		t.Fatalf("Get вернул ошибку: %v", err)
	}
	if !exists || value != "temp-value" {
		t.Errorf("Ожидалось 'temp-value', получено '%v', существует: %v", value, exists)
	}

	// Ждём истечения TTL

	time.Sleep(2 * time.Second)

	_, exists, err = client.Get("ttl-key")
	if err != nil {
		t.Fatalf("Get вернул ошибку после истечения TTL: %v", err)
	}
	if exists {
		t.Error("Ожидалось, что значение не существует после истечения TTL")
	}
}

func TestBatchPutGetAndDelete(t *testing.T) {
	fmt.Println("RUN TestBatchPutGetAndDelete")
	client, cleanup := setupClient(t)
	defer cleanup()

	items := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	err := client.BatchPut(items, map[string]int{})
	if err != nil {
		t.Fatalf("BatchPut вернул ошибку: %v", err)
	}

	// Проверяем BatchGet
	result, err := client.BatchGet([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("BatchGet вернул ошибку: %v", err)
	}
	if !reflect.DeepEqual(result, items) {
		t.Errorf("Ожидалось %v, получено %v", items, result)
	}

	// Удаляем часть ключей
	err = client.BatchDelete([]string{"a", "c"})
	if err != nil {
		t.Fatalf("BatchDelete вернул ошибку: %v", err)
	}

	// Проверяем остаток
	remaining, err := client.BatchGet([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("BatchGet вернул ошибку после BatchDelete: %v", err)
	}
	expected := map[string]string{"b": "2"}
	if !reflect.DeepEqual(remaining, expected) {
		t.Errorf("Ожидалось %v, получено %v", expected, remaining)
	}

	// Проверяем BatchGet с пустым списком
	empty, err := client.BatchGet([]string{})
	if err != nil {
		t.Fatalf("BatchGet вернул ошибку для пустого списка: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("Ожидалось пустую карту для пустого списка, получено %v", empty)
	}
}
