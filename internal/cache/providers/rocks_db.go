// Package providers реализует провайдер кэша на базе RocksDB,
// где информация о сроке жизни (TTL) хранится в отдельной Column Family
// `ttl_cf`.  Такой подход даёт:
//   - Чёткое разделение пользовательских данных и метаданных TTL;
//   - Быстрый поиск / сканирование только по TTL‑меткам;
//   - Атомарные записи (ключ + TTL) внутри одного WriteBatch;
//   - Возможность фоновой чистки «холодных» просроченных ключей без
//     участия операций чтения.
//
// Структура CF:
//
//	default  — ключ → значение (полезная нагрузка);
//	ttl_cf   — ключ → время истечения UnixNano (int64 в []byte).
//
// Ключи считаются «устаревшими» (expired), если текущее время превышает
// сохранённый таймстемп. Удаление просроченных записей происходит двумя
// способами:
//  1. Ленивая очистка при чтении (BatchGet): если ключ устарел — он тут же
//     удаляется из обеих CF.
//  2. Фоновый коллектор (StartTTLCollector): периодически сканирует ttl_cf и
//     удаляет «мёртвые» пары, до которых ещё не добрались запросы.
package providers

import (
	"aur-cache-service/internal/cache/config"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
)

// Имена Column Family.
const (
	defaultCFName = "default"
	ttlCFName     = "ttl_cf"
)

// RocksDbCF — реализация CacheProvider c отдельной CF для TTL.
//
// Безопасность:
//   - Внутренний ttlCache защищён RW‑мьютексом (ttlMu);
//   - Методы Batch* допускают конкурентный вызов.
//
// Замечание: сам RocksDB потокобезопасен, если использовать независимые
// ReadOptions / WriteOptions (что мы и делаем).
type RocksDbCF struct {
	db        *grocksdb.DB
	defaultCF *grocksdb.ColumnFamilyHandle
	ttlCF     *grocksdb.ColumnFamilyHandle

	readOpts  *grocksdb.ReadOptions
	writeOpts *grocksdb.WriteOptions

	ttlMu    sync.RWMutex
	ttlCache map[string]int64 // key -> UnixNano
}

// -----------------------------------------------------------------------------
// Создание / закрытие базы
// -----------------------------------------------------------------------------

// NewRocksDbCF открывает базу с двумя CF (default, ttl_cf) и возвращает провайдер.

func NewRocksDbCF(cfg config.RocksDB) (*RocksDbCF, error) {
	// Shared options
	dbOpts := grocksdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(cfg.CreateIfMissing)
	dbOpts.SetCreateIfMissingColumnFamilies(true)
	if cfg.MaxOpenFiles > 0 {
		dbOpts.SetMaxOpenFiles(cfg.MaxOpenFiles)
	}

	// Block‑cache tuning (optional)
	blockCacheBytes, _ := cfg.BlockCacheBytes()
	blockSizeBytes, _ := cfg.BlockSizeBytes()
	if blockCacheBytes > 0 {
		bbto := grocksdb.NewDefaultBlockBasedTableOptions()
		bbto.SetBlockCache(grocksdb.NewLRUCache(blockCacheBytes))
		if blockSizeBytes > 0 {
			bbto.SetBlockSize(int(blockSizeBytes))
		}
		dbOpts.SetBlockBasedTableFactory(bbto)
	}

	// Column family list & per‑CF opts (reuse dbOpts)
	cfNames := []string{defaultCFName, ttlCFName}
	cfOpts := []*grocksdb.Options{dbOpts, dbOpts}

	db, cfHandles, err := grocksdb.OpenDbColumnFamilies(dbOpts, cfg.Path, cfNames, cfOpts)
	if err != nil {
		return nil, fmt.Errorf("open rocksdb with column families: %w", err)
	}

	return &RocksDbCF{
		db:        db,
		defaultCF: cfHandles[0],
		ttlCF:     cfHandles[1],
		readOpts:  grocksdb.NewDefaultReadOptions(),
		writeOpts: grocksdb.NewDefaultWriteOptions(),
		ttlCache:  make(map[string]int64),
	}, nil
}

func (c *RocksDbCF) Close() error {
	c.readOpts.Destroy()
	c.writeOpts.Destroy()
	// ColumnFamily handles must be destroyed before db Close.
	c.defaultCF.Destroy()
	c.ttlCF.Destroy()
	c.db.Close()
	return nil
}

// ---------------- helpers ----------------

// encodeInt64 converts int64 -> []byte (big endian) for lexicographical order.
func encodeInt64(v int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	return b[:]
}

// decodeInt64 does the reverse.
func decodeInt64(b []byte) int64 {
	if len(b) < 8 {
		// fallback: parse decimal string (compat)
		n, _ := strconv.ParseInt(string(b), 10, 64)
		return n
	}
	return int64(binary.BigEndian.Uint64(b))
}

// getTTL returns expiry ts (UnixNano) and a boolean.
func (c *RocksDbCF) getTTL(key string) (int64, bool) {
	c.ttlMu.RLock()
	if ts, ok := c.ttlCache[key]; ok {
		c.ttlMu.RUnlock()
		return ts, true
	}
	c.ttlMu.RUnlock()

	slice, err := c.db.GetCF(c.readOpts, c.ttlCF, []byte(key))
	if err != nil || !slice.Exists() {
		if slice != nil {
			slice.Free()
		}
		return 0, false
	}
	ts := decodeInt64(slice.Data())
	slice.Free()

	c.ttlMu.Lock()
	c.ttlCache[key] = ts
	c.ttlMu.Unlock()
	return ts, true
}

// setTTL writes ttl to CF and cache.
func (c *RocksDbCF) setTTL(batch *grocksdb.WriteBatch, key string, exp time.Time) {
	ts := exp.UnixNano()
	batch.PutCF(c.ttlCF, []byte(key), encodeInt64(ts))
	c.ttlMu.Lock()
	c.ttlCache[key] = ts
	c.ttlMu.Unlock()
}

// deleteTTL removes TTL from CF and cache.
func (c *RocksDbCF) deleteTTL(batch *grocksdb.WriteBatch, key string) {
	batch.DeleteCF(c.ttlCF, []byte(key))
	c.ttlMu.Lock()
	delete(c.ttlCache, key)
	c.ttlMu.Unlock()
}

// expired checks if key is expired w.r.t now(). Does not delete.
func (c *RocksDbCF) expired(key string, now time.Time) bool {
	if ts, ok := c.getTTL(key); ok {
		return now.UnixNano() > ts
	}
	return false
}

// ---------------- CacheProvider interface ----------------

func (c *RocksDbCF) BatchGet(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	result := make(map[string]string, len(keys))
	now := time.Now()

	expiredKeys := make([]string, 0)

	for i, key := range keys {
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
		if c.expired(key, now) {
			expiredKeys = append(expiredKeys, key)
			continue
		}
		slice, err := c.db.GetCF(c.readOpts, c.defaultCF, []byte(key))
		if err != nil {
			return nil, fmt.Errorf("rocksdb get: %w", err)
		}
		if slice.Exists() {
			result[key] = string(slice.Data())
		}
		slice.Free()
	}

	if len(expiredKeys) > 0 {
		_ = c.BatchDelete(ctx, expiredKeys) // best‑effort cleanup
	}

	return result, nil
}

func (c *RocksDbCF) BatchPut(ctx context.Context, items map[string]string, ttls map[string]time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	now := time.Now()
	count := 0
	for key, val := range items {
		if count%100 == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		count++
		batch.PutCF(c.defaultCF, []byte(key), []byte(val))
		if ttl, ok := ttls[key]; ok && ttl > 0 {
			c.setTTL(batch, key, now.Add(ttl))
		}
	}
	if err := c.db.Write(c.writeOpts, batch); err != nil {
		return fmt.Errorf("rocksdb batch put: %w", err)
	}
	return nil
}

func (c *RocksDbCF) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	for i, key := range keys {
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		batch.DeleteCF(c.defaultCF, []byte(key))
		c.deleteTTL(batch, key)
	}
	return c.db.Write(c.writeOpts, batch)
}

// ---------------- Background TTL collector ----------------

// StartTTLCollector launches a goroutine that every `interval` scans the ttl_cf
// and hard‑deletes expired keys. Cancel the ctx to stop the cleaner.
func (c *RocksDbCF) StartTTLCollector(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.collectOnce()
			}
		}
	}()
}

// collectOnce scans ttl_cf and removes expired pairs.
func (c *RocksDbCF) collectOnce() {
	now := time.Now().UnixNano()
	it := c.db.NewIteratorCF(c.readOpts, c.ttlCF)
	defer it.Close()

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key().Data()
		ts := decodeInt64(it.Value().Data())
		if now > ts {
			batch.DeleteCF(c.defaultCF, key)
			batch.DeleteCF(c.ttlCF, key)
			c.ttlMu.Lock()
			delete(c.ttlCache, string(key))
			c.ttlMu.Unlock()
		}
		it.Key().Free()
		it.Value().Free()
	}

	if batch.Count() > 0 {
		_ = c.db.Write(c.writeOpts, batch) // ignore error for collector
	}
}

// ---------------- compile‑time check ----------------
var _ CacheProvider = (*RocksDbCF)(nil)
