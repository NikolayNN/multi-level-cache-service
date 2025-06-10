package manager

import (
	"aur-cache-service/api/dto"
	"context"
	"log"
)

// AsyncManagerAdapter provides a simplified interface over Manager.
// PutAll and EvictAll are executed asynchronously to reduce
// blocking of the calling goroutine.
type AsyncManagerAdapter interface {
	Get(ctx context.Context, id *dto.CacheId) *dto.CacheEntryHit
	Put(ctx context.Context, entry *dto.CacheEntry)
	Evict(ctx context.Context, id *dto.CacheId)

	GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit
	PutAll(ctx context.Context, entries []*dto.CacheEntry)
	EvictAll(ctx context.Context, ids []*dto.CacheId)
}

type asyncManagerAdapter struct {
	manager Manager
}

// NewAsyncAdapter creates a new AsyncManagerAdapter for the provided Manager.
func NewAsyncAdapter(m Manager) AsyncManagerAdapter {
	return &asyncManagerAdapter{manager: m}
}

func (f *asyncManagerAdapter) Get(ctx context.Context, id *dto.CacheId) *dto.CacheEntryHit {
	if id == nil {
		return nil
	}
	res := f.GetAll(ctx, []*dto.CacheId{id})
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

func (f *asyncManagerAdapter) Put(ctx context.Context, entry *dto.CacheEntry) {
	if entry == nil {
		return
	}
	f.PutAll(ctx, []*dto.CacheEntry{entry})
}

func (f *asyncManagerAdapter) Evict(ctx context.Context, id *dto.CacheId) {
	if id == nil {
		return
	}
	f.EvictAll(ctx, []*dto.CacheId{id})
}

func (f *asyncManagerAdapter) GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit {
	return f.manager.GetAll(ctx, ids)
}

func (f *asyncManagerAdapter) PutAll(ctx context.Context, entries []*dto.CacheEntry) {
	if len(entries) == 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in PutAll goroutine: %v", r)
			}
		}()
		f.manager.PutAll(ctx, entries)
	}()
}

func (f *asyncManagerAdapter) EvictAll(ctx context.Context, ids []*dto.CacheId) {
	if len(ids) == 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in EvictAll goroutine: %v", r)
			}
		}()
		f.manager.EvictAll(ctx, ids)
	}()
}
