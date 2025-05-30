package dto

import (
	"encoding/json"
)

// /////////////////////
//// Внешний API
///////////////////////

// Внешний API содержит минимальные данные для того чтобы найти значение  для GET
type CacheId struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
}

// Внешний API содержит минимальные данные для идентификации и значение используется для PUT
type CacheEntry struct {
	CacheId *CacheId        `json:",inline"`
	Value   json.RawMessage `json:"v"`
}

// Внешний API содержит минимальные данные для идентифиации, значение и др информацию это результат GET
type CacheEntryHit struct {
	CacheEntry *CacheEntry `json:",inline"`
	Found      bool        `json:"f"`
}

// /////////////////////
//// Внутренний API
///////////////////////

// Жизненный циклы
// GET: CacheId -> ResolvedCacheId -> ResolvedCacheHit -> CacheEntryHit
// PUT: CacheEntry -> ResolvedCacheEntry -> ResolvedCacheEntryLevel -> CacheId
// DELETE: CacheId -> ResolvedCacheId -> CacheId

type CacheIdRef interface {
	GetCacheName() string
	GetKey() string
}

func (c *CacheEntry) GetCacheName() string { return c.CacheId.CacheName }
func (c *CacheEntry) GetKey() string       { return c.CacheId.Key }

func (c *CacheId) GetCacheName() string { return c.CacheName }
func (c *CacheId) GetKey() string       { return c.Key }

type StorageKeyRef interface {
	GetStorageKey() string
}

// содержит все координаты кэша и сам StorageKey по которому можно найти данные в кэше
type ResolvedCacheId struct {
	CacheId    *CacheId
	StorageKey string
}

func (r *ResolvedCacheId) GetCacheName() string  { return r.CacheId.CacheName }
func (r *ResolvedCacheId) GetKey() string        { return r.CacheId.Key }
func (r *ResolvedCacheId) GetStorageKey() string { return r.StorageKey }

// содержит все координаты кэша и его значение
type ResolvedCacheEntry struct {
	ResolvedCacheId *ResolvedCacheId
	Value           json.RawMessage
}

func (r *ResolvedCacheEntry) GetStorageKey() string { return r.ResolvedCacheId.GetStorageKey() }
func (r *ResolvedCacheEntry) GetCacheName() string  { return r.ResolvedCacheId.GetCacheName() }
func (r *ResolvedCacheEntry) GetKey() string        { return r.ResolvedCacheId.GetKey() }

// содержит все координаты кэша его значение и результат выполнения операции
type ResolvedCacheHit struct {
	ResolvedCacheEntry *ResolvedCacheEntry
	Found              bool
}

func (r *ResolvedCacheHit) GetStorageKey() string { return r.ResolvedCacheEntry.GetStorageKey() }

func (r *ResolvedCacheHit) GetCacheName() string { return r.ResolvedCacheEntry.GetCacheName() }
func (r *ResolvedCacheHit) GetKey() string       { return r.ResolvedCacheEntry.GetKey() }
