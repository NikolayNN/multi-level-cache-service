package cache

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache/providers"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockService struct {
	getAllCalled    int
	putAllCalled    int
	deleteAllCalled int
	fail            bool
	layer           int
}

func (m *mockService) GetAll(_ context.Context, reqs []*dto.ResolvedCacheId) (*dto.GetResult, error) {
	m.getAllCalled++
	if m.fail {
		return nil, errors.New("service unavailable")
	}
	return &dto.GetResult{
		Hits:    []*dto.ResolvedCacheHit{{ResolvedCacheEntry: &dto.ResolvedCacheEntry{ResolvedCacheId: reqs[0]}}},
		Misses:  []*dto.ResolvedCacheId{},
		Skipped: []*dto.ResolvedCacheId{},
	}, nil
}

func (m *mockService) PutAll(_ context.Context, _ []*dto.ResolvedCacheEntry) error {
	m.putAllCalled++
	if m.fail {
		return errors.New("put failed")
	}
	return nil
}

func (m *mockService) DeleteAll(_ context.Context, _ []*dto.ResolvedCacheId) error {
	m.deleteAllCalled++
	if m.fail {
		return errors.New("delete failed")
	}
	return nil
}

func (m *mockService) Close() error {
	return nil
}

func TestController_GetAll(t *testing.T) {
	service := &mockService{}
	controller := CreateControllerImpl([]providers.Service{service})

	reqs := []*dto.ResolvedCacheId{
		{CacheId: &dto.CacheId{CacheName: "test", Key: "1"}, StorageKey: "test:1"},
	}

	results := controller.GetAll(context.Background(), reqs)
	assert.Len(t, results, 1)
	assert.Equal(t, 1, service.getAllCalled)
	assert.Len(t, results[0].Hits, 1)
	assert.Equal(t, "test:1", results[0].Hits[0].ResolvedCacheEntry.ResolvedCacheId.StorageKey)
}

func TestController_PutAll(t *testing.T) {
	s1 := &mockService{layer: 0}
	s2 := &mockService{layer: 1}
	controller := CreateControllerImpl([]providers.Service{s1, s2})

	entry := &dto.ResolvedCacheEntry{
		ResolvedCacheId: &dto.ResolvedCacheId{
			CacheId:    &dto.CacheId{CacheName: "test", Key: "1"},
			StorageKey: "test:1",
		},
	}
	controller.PutAll(context.Background(), []*dto.ResolvedCacheEntry{entry}, 0)

	assert.Equal(t, 1, s1.putAllCalled)
	assert.Equal(t, 0, s2.putAllCalled)
}

func TestController_PutAllToAllLevels(t *testing.T) {
	s1 := &mockService{}
	s2 := &mockService{}
	controller := CreateControllerImpl([]providers.Service{s1, s2})

	entry := &dto.ResolvedCacheEntry{
		ResolvedCacheId: &dto.ResolvedCacheId{
			CacheId:    &dto.CacheId{CacheName: "test", Key: "1"},
			StorageKey: "test:1",
		},
	}
	controller.PutAllToAllLevels(context.Background(), []*dto.ResolvedCacheEntry{entry})

	assert.Equal(t, 1, s1.putAllCalled)
	assert.Equal(t, 1, s2.putAllCalled)
}

func TestController_DeleteAll(t *testing.T) {
	s1 := &mockService{}
	s2 := &mockService{}
	controller := CreateControllerImpl([]providers.Service{s1, s2})

	reqs := []*dto.ResolvedCacheId{
		{CacheId: &dto.CacheId{CacheName: "test", Key: "1"}, StorageKey: "test:1"},
	}
	controller.DeleteAll(context.Background(), reqs)

	assert.Equal(t, 1, s1.deleteAllCalled)
	assert.Equal(t, 1, s2.deleteAllCalled)
}
