package manager

import (
	"aur-cache-service/api/dto"
	"aur-cache-service/internal/cache/config"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockCacheService struct {
	prefixMap map[string]string
}

func (m *mockCacheService) GetPrefix(id config.CacheNameable) (string, error) {
	return m.prefixMap[id.GetCacheName()], nil
}
func (m *mockCacheService) GetCache(config.CacheNameable) (config.Cache, error) {
	return config.Cache{}, nil
}
func (m *mockCacheService) GetCacheByName(string) (config.Cache, error) { return config.Cache{}, nil }
func (m *mockCacheService) GetTtl(config.CacheNameable, int) (time.Duration, error) {
	return 0, nil
}
func (m *mockCacheService) IsLevelEnabled(config.CacheNameable, int) (bool, error) {
	return true, nil
}

// mocks for cache.Controller and integration.Controller

type mockCacheController struct {
	getReqs      []*dto.ResolvedCacheId
	getReturn    []*dto.GetResult
	putEntries   []*dto.ResolvedCacheEntry
	putBound     []int
	putAllCalled int
	getCalled    int
	deleteReqs   []*dto.ResolvedCacheId
	deleteCalled int
	putAllToAll  int
	putAllWG     sync.WaitGroup
}

func (m *mockCacheController) GetAll(_ context.Context, reqs []*dto.ResolvedCacheId) []*dto.GetResult {
	m.getCalled++
	m.getReqs = reqs
	return m.getReturn
}

func (m *mockCacheController) PutAll(_ context.Context, entries []*dto.ResolvedCacheEntry, bound int) {
	m.putAllCalled++
	m.putEntries = append(m.putEntries, entries...)
	m.putBound = append(m.putBound, bound)
	m.putAllWG.Done()
}

func (m *mockCacheController) PutAllToAllLevels(_ context.Context, entries []*dto.ResolvedCacheEntry) {
	m.putAllToAll++
	m.putEntries = entries
}

func (m *mockCacheController) DeleteAll(_ context.Context, reqs []*dto.ResolvedCacheId) {
	m.deleteCalled++
	m.deleteReqs = reqs
}

type mockExternalController struct {
	reqs   []*dto.ResolvedCacheId
	result *dto.GetResult
	called int
}

func (m *mockExternalController) GetAll(reqs []*dto.ResolvedCacheId) *dto.GetResult {
	m.called++
	m.reqs = reqs
	if m.result == nil {
		return &dto.GetResult{}
	}
	return m.result
}

func TestManager_GetAll_FillMissing(t *testing.T) {
	mapper := dto.NewResolverMapper(&mockCacheService{prefixMap: map[string]string{"c": "p"}})
	id := &dto.CacheId{CacheName: "c", Key: "1"}
	rid := &dto.ResolvedCacheId{CacheId: id, StorageKey: "p:1"}
	entry := &dto.ResolvedCacheEntry{ResolvedCacheId: rid, Value: nil}
	hit := &dto.ResolvedCacheHit{ResolvedCacheEntry: entry, Found: true}

	ctrl := &mockCacheController{getReturn: []*dto.GetResult{
		{Hits: []*dto.ResolvedCacheHit{}, Misses: []*dto.ResolvedCacheId{rid}},
		{Hits: []*dto.ResolvedCacheHit{}, Misses: []*dto.ResolvedCacheId{rid}},
		{Hits: []*dto.ResolvedCacheHit{hit}, Misses: []*dto.ResolvedCacheId{}},
	}}
	ctrl.putAllWG.Add(1)

	ext := &mockExternalController{}

	mgr := &ManagerImpl{cacheController: ctrl, externalController: ext, mapper: *mapper}

	res := mgr.GetAll(context.Background(), []*dto.CacheId{id})
	ctrl.putAllWG.Wait()

	assert.Equal(t, 1, ctrl.getCalled)
	assert.Equal(t, 1, ext.called)
	assert.Len(t, res, 1)
	assert.Equal(t, "1", res[0].CacheEntry.CacheId.Key)
	assert.Equal(t, 1, ctrl.putAllCalled)
	assert.Equal(t, 0, ctrl.putBound[0])
	assert.Equal(t, rid.StorageKey, ctrl.putEntries[0].ResolvedCacheId.StorageKey)
}

func TestManager_GetAll_Empty(t *testing.T) {
	mapper := dto.NewResolverMapper(&mockCacheService{prefixMap: map[string]string{"c": "p"}})
	ctrl := &mockCacheController{getReturn: []*dto.GetResult{}}
	ext := &mockExternalController{}
	mgr := &ManagerImpl{cacheController: ctrl, externalController: ext, mapper: *mapper}

	res := mgr.GetAll(context.Background(), []*dto.CacheId{{CacheName: "c", Key: "k"}})

	assert.Len(t, res, 0)
	assert.Equal(t, 0, ext.called)
	assert.Equal(t, 0, ctrl.putAllCalled)
}

func TestManager_PutAndEvict(t *testing.T) {
	mapper := dto.NewResolverMapper(&mockCacheService{prefixMap: map[string]string{"c": "p"}})
	ctrl := &mockCacheController{}
	mgr := &ManagerImpl{cacheController: ctrl, externalController: &mockExternalController{}, mapper: *mapper}

	raw := json.RawMessage(`"v"`)
	entry := &dto.CacheEntry{CacheId: &dto.CacheId{CacheName: "c", Key: "1"}, Value: &raw}
	mgr.PutAll(context.Background(), []*dto.CacheEntry{entry})
	assert.Equal(t, 1, ctrl.putAllToAll)
	assert.Equal(t, "p:1", ctrl.putEntries[0].ResolvedCacheId.StorageKey)

	id := &dto.CacheId{CacheName: "c", Key: "2"}
	mgr.EvictAll(context.Background(), []*dto.CacheId{id})
	assert.Equal(t, 1, ctrl.deleteCalled)
	assert.Equal(t, "p:2", ctrl.deleteReqs[0].StorageKey)
}
