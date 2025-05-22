package ristrettoClient_test

import (
	"aur-cache-service/internal/clients/ristrettoClient"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newTestClient(t *testing.T) *ristrettoClient.Client {
	client, err := ristrettoClient.New(ristrettoClient.Config{
		NumCounters: 1000,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	require.NoError(t, err)
	return client
}

func TestPutAndGet(t *testing.T) {
	client := newTestClient(t)

	key := "testKey"
	val := "testValue"

	err := client.Put(key, val, 0)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Подождать, чтобы кэш успел обработать

	got, found, err := client.Get(key)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, val, got)
}

func TestDelete(t *testing.T) {
	client := newTestClient(t)

	_ = client.Put("toDelete", "value", 0)
	time.Sleep(10 * time.Millisecond)

	deleted, err := client.Delete("toDelete")
	require.NoError(t, err)
	require.True(t, deleted)

	got, found, _ := client.Get("toDelete")
	require.False(t, found)
	require.Equal(t, "", got)
}

func TestBatchPutAndBatchGet(t *testing.T) {
	client := newTestClient(t)

	items := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	ttls := map[string]int{
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

func TestBatchDelete(t *testing.T) {
	client := newTestClient(t)

	_ = client.Put("keyA", "valA", 0)
	_ = client.Put("keyB", "valB", 0)
	time.Sleep(10 * time.Millisecond)

	err := client.BatchDelete([]string{"keyA", "keyB"})
	require.NoError(t, err)

	_, foundA, _ := client.Get("keyA")
	_, foundB, _ := client.Get("keyB")
	require.False(t, foundA)
	require.False(t, foundB)
}
