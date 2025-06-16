package manager

import (
	"aur-cache-service/api/dto"
	"context"
	"time"

	"go.uber.org/zap"
	"telegram-alerts-go/alert"
)

// Максимум 64 одновременных async-операций (PutAll/EvictAll).
var tokens = make(chan struct{}, 64)

// ManagerAdapter provides a simplified interface over Manager.
// PutAll и EvictAll выполняются асинхронно, поэтому доступ лимитируем семафором.
type ManagerAdapter interface {
	Get(ctx context.Context, id *dto.CacheId) *dto.CacheEntryHit
	Put(ctx context.Context, entry *dto.CacheEntry)
	Evict(ctx context.Context, id *dto.CacheId)

	GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit
	PutAll(ctx context.Context, entries []*dto.CacheEntry)
	EvictAll(ctx context.Context, ids []*dto.CacheId)
}

type AsyncManagerAdapter struct {
	manager         Manager
	putAllTimeout   time.Duration
	evictAllTimeout time.Duration
}

const defaultTimeout = 5 * time.Second

// NewAsyncManagerAdapter создаёт адаптер с явными или дефолтными тайм-аутами.
func NewAsyncManagerAdapter(m Manager, putTO, evictTO time.Duration) *AsyncManagerAdapter {
	if putTO <= 0 {
		zap.S().Warnf("putAllTimeout ≤ 0 set %s", defaultTimeout)
		putTO = defaultTimeout
	}
	if evictTO <= 0 {
		zap.S().Warnf("evictAllTimeout ≤ 0 set %s", defaultTimeout)
		evictTO = defaultTimeout
	}
	return &AsyncManagerAdapter{
		manager:         m,
		putAllTimeout:   putTO,
		evictAllTimeout: evictTO,
	}
}

/* ---------- синхронные обёртки ---------- */

func (a *AsyncManagerAdapter) Get(ctx context.Context, id *dto.CacheId) *dto.CacheEntryHit {
	if id == nil {
		return nil
	}
	res := a.GetAll(ctx, []*dto.CacheId{id})
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

func (a *AsyncManagerAdapter) Put(ctx context.Context, e *dto.CacheEntry) {
	if e != nil {
		a.PutAll(ctx, []*dto.CacheEntry{e})
	}
}

func (a *AsyncManagerAdapter) Evict(ctx context.Context, id *dto.CacheId) {
	if id != nil {
		a.EvictAll(ctx, []*dto.CacheId{id})
	}
}

func (a *AsyncManagerAdapter) GetAll(ctx context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit {
	return a.manager.GetAll(ctx, ids)
}

/* ---------- ограниченный async-раннер ---------- */

func (a *AsyncManagerAdapter) runAsync(name string, f func(ctx context.Context), d time.Duration) {
	/* забираем токен — если буфер полон, ждём */
	tokens <- struct{}{}

	go func() {
		/* освобождаем токен при выходе */
		defer func() { <-tokens }()

		ctx, cancel := makeCtx(d)
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				zap.S().Errorf(alert.Prefix("async panic: %v"), r)
			}
		}()

		zap.S().Infow("async started", "name", name)
		f(ctx)
		zap.S().Infow("async finished", "name", name)
	}()
}

func makeCtx(d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), d)
}

/* ---------- асинхронные методы ---------- */

func (a *AsyncManagerAdapter) PutAll(_ context.Context, entries []*dto.CacheEntry) {
	if len(entries) == 0 {
		return
	}
	a.runAsync("putAll", func(ctx context.Context) {
		a.manager.PutAll(ctx, entries)
	}, a.putAllTimeout)
}

func (a *AsyncManagerAdapter) EvictAll(_ context.Context, ids []*dto.CacheId) {
	if len(ids) == 0 {
		return
	}
	a.runAsync("evictAll", func(ctx context.Context) {
		a.manager.EvictAll(ctx, ids)
	}, a.evictAllTimeout)
}
