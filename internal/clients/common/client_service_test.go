package common

import (
	"aur-cache-service/internal/request"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCacheClient struct {
	getFunc      func(key string) (string, bool, error)
	batchGetFunc func(keys []string) (map[string]string, error)
	closeCalled  bool
}

func (m *mockCacheClient) Get(key string) (string, bool, error) {
	return m.getFunc(key)
}

func (m *mockCacheClient) Put(key string, value string, ttl int) error {
	return nil
}

func (m *mockCacheClient) Delete(key string) (bool, error) {
	return true, nil
}

func (m *mockCacheClient) BatchGet(keys []string) (map[string]string, error) {
	return m.batchGetFunc(keys)
}

func (m *mockCacheClient) BatchPut(items map[string]string, ttls map[string]int) error {
	return nil
}

func (m *mockCacheClient) BatchDelete(keys []string) error {
	return nil
}

func (m *mockCacheClient) Close() error {
	m.closeCalled = true
	return nil
}

func TestCacheService_Get_KeyExists(t *testing.T) {
	mockClient := &mockCacheClient{
		getFunc: func(key string) (string, bool, error) {
			if key == "u:1" {
				return "value1", true, nil
			}
			return "", false, nil
		},
	}

	service := NewCacheService(mockClient)
	req := &request.ResolvedGetCacheReq{
		Req: &request.GetCacheReq{
			CacheName: "user",
			Key:       "1",
		},
		CacheKey: "u:1",
	}

	resp := service.Get(req)

	assert.Equal(t, req, resp.Req)
	assert.Equal(t, "value1", resp.Value)
	assert.True(t, resp.Found)
}

func TestCacheService_Get_KeyNotExists(t *testing.T) {
	mockClient := &mockCacheClient{
		getFunc: func(key string) (string, bool, error) {
			if key == "u:1" {
				return "value1", true, nil
			}
			return "", false, nil
		},
	}

	service := NewCacheService(mockClient)
	req := &request.ResolvedGetCacheReq{
		Req: &request.GetCacheReq{
			CacheName: "user",
			Key:       "999",
		},
		CacheKey: "u:999",
	}

	resp := service.Get(req)

	assert.Equal(t, req, resp.Req)
	assert.Equal(t, "", resp.Value)
	assert.False(t, resp.Found)
}

func TestCacheService_BatchGet_allExists(t *testing.T) {
	mockClient := &mockCacheClient{
		batchGetFunc: func(keys []string) (map[string]string, error) {
			return map[string]string{
				"u:1": "user1",
				"u:2": "user2",
			}, nil
		},
	}

	service := NewCacheService(mockClient)

	reqs := []request.ResolvedGetCacheReq{
		{
			Req: &request.GetCacheReq{
				CacheName: "user",
				Key:       "1",
			},
			CacheKey: "u:1",
		},
		{
			Req: &request.GetCacheReq{
				CacheName: "user",
				Key:       "2",
			},
			CacheKey: "u:2",
		},
	}

	responses := service.BatchGet(reqs)

	assert.Len(t, responses, 2)
	assert.Equal(t, "user1", responses[0].Value)
	assert.True(t, responses[0].Found)
	assert.Equal(t, "user2", responses[1].Value)
	assert.True(t, responses[1].Found)
}

func TestCacheService_BatchGet_oneNotExists(t *testing.T) {
	mockClient := &mockCacheClient{
		batchGetFunc: func(keys []string) (map[string]string, error) {
			return map[string]string{
				"u:2": "user2",
			}, nil
		},
	}

	service := NewCacheService(mockClient)

	reqs := []request.ResolvedGetCacheReq{
		{
			Req: &request.GetCacheReq{
				CacheName: "user",
				Key:       "1",
			},
			CacheKey: "u:1",
		},
		{
			Req: &request.GetCacheReq{
				CacheName: "user",
				Key:       "2",
			},
			CacheKey: "u:2",
		},
	}

	responses := service.BatchGet(reqs)

	assert.Len(t, responses, 2)
	assert.Equal(t, "", responses[0].Value)
	assert.False(t, responses[0].Found)
	assert.Equal(t, "user2", responses[1].Value)
	assert.True(t, responses[1].Found)
}
