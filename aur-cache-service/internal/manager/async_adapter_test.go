package manager

import (
	"aur-cache-service/api/dto"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockManager struct {
	mu             sync.Mutex
	getAllCalled   int
	putAllCalled   int
	evictAllCalled int
	wait           time.Duration
	putWG          sync.WaitGroup
	evictWG        sync.WaitGroup
}

func (m *mockManager) GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit {
	m.mu.Lock()
	m.getAllCalled++
	m.mu.Unlock()
	hits := make([]*dto.CacheEntryHit, len(ids))
	for i, id := range ids {
		hits[i] = &dto.CacheEntryHit{CacheEntry: &dto.CacheEntry{CacheId: id}, Found: true}
	}
	return hits
}

func (m *mockManager) PutAll(ctx context.Context, entries []*dto.CacheEntry) {
	defer m.putWG.Done()
	time.Sleep(m.wait)
	m.mu.Lock()
	m.putAllCalled++
	m.mu.Unlock()
}

func (m *mockManager) EvictAll(ctx context.Context, ids []*dto.CacheId) {
	defer m.evictWG.Done()
	time.Sleep(m.wait)
	m.mu.Lock()
	m.evictAllCalled++
	m.mu.Unlock()
}

func TestAsyncAdapter_GetWrapsManager(t *testing.T) {
	mgr := &mockManager{}
	f := NewAsyncManagerAdapter(mgr, 1*time.Second, 1*time.Second)
	id := &dto.CacheId{CacheName: "c", Key: "k"}
	res := f.Get(context.Background(), id)
	assert.NotNil(t, res)
	assert.Equal(t, 1, mgr.getAllCalled)
	assert.Equal(t, id, res.CacheEntry.CacheId)
}

func TestAsyncAdapter_PutAllAsync(t *testing.T) {
	mgr := &mockManager{wait: 50 * time.Millisecond}
	mgr.putWG.Add(1)
	f := NewAsyncManagerAdapter(mgr, 1*time.Second, 1*time.Second)
	entry := &dto.CacheEntry{CacheId: &dto.CacheId{CacheName: "c", Key: "k"}}

	start := time.Now()
	f.PutAll(context.Background(), []*dto.CacheEntry{entry})
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 30*time.Millisecond)

	mgr.putWG.Wait()
	assert.Equal(t, 1, mgr.putAllCalled)
}

func TestAsyncAdapter_EvictAllAsync(t *testing.T) {
	mgr := &mockManager{wait: 50 * time.Millisecond}
	mgr.evictWG.Add(1)
	f := NewAsyncManagerAdapter(mgr, 1*time.Second, 1*time.Second)
	id := &dto.CacheId{CacheName: "c", Key: "k"}

	start := time.Now()
	f.EvictAll(context.Background(), []*dto.CacheId{id})
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 30*time.Millisecond)

	mgr.evictWG.Wait()
	assert.Equal(t, 1, mgr.evictAllCalled)
}

func TestAsyncAdapter_SingleHelpers(t *testing.T) {
	mgr := &mockManager{wait: 10 * time.Millisecond}
	mgr.putWG.Add(1)
	mgr.evictWG.Add(1)
	f := NewAsyncManagerAdapter(mgr, 1*time.Second, 1*time.Second)

	id := &dto.CacheId{CacheName: "c", Key: "k"}
	entry := &dto.CacheEntry{CacheId: id}

	f.Put(context.Background(), entry)
	f.Evict(context.Background(), id)
	res := f.Get(context.Background(), id)

	mgr.putWG.Wait()
	mgr.evictWG.Wait()

	assert.Equal(t, 1, mgr.putAllCalled)
	assert.Equal(t, 1, mgr.evictAllCalled)
	assert.Equal(t, 1, mgr.getAllCalled)
	assert.NotNil(t, res)
}
