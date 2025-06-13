package dto

import (
	"aur-cache-service/internal/cache/config"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockCacheService struct {
	prefixMap map[string]string
	failMap   map[string]bool
}

func (m *mockCacheService) GetPrefix(cacheId config.CacheNameable) (string, error) {
	if m.failMap[cacheId.GetCacheName()] {
		return "", errors.New("prefix not found")
	}
	return m.prefixMap[cacheId.GetCacheName()], nil
}

func (m *mockCacheService) GetCache(config.CacheNameable) (config.Cache, error) {
	panic("not used in the test")
}
func (m *mockCacheService) GetCacheByName(string) (config.Cache, error) {
	panic("not used in the test")
}
func (m *mockCacheService) GetTtl(config.CacheNameable, int) (ttl time.Duration, err error) {
	panic("not used in the test")
}
func (m *mockCacheService) IsLevelEnabled(config.CacheNameable, int) (enabled bool, err error) {
	panic("not used in the test")
}

func TestMapAllResolvedCacheId(t *testing.T) {
	mapper := NewResolverMapper(&mockCacheService{
		prefixMap: map[string]string{"test": "pfx"},
		failMap:   map[string]bool{"fail": true},
	})
	ids := []*CacheId{
		{CacheName: "test", Key: "key1"},
		{CacheName: "fail", Key: "key2"},
	}

	result := mapper.MapAllResolvedCacheId(ids)
	assert.Len(t, result, 1)
	assert.Equal(t, "pfx:key1", result[0].StorageKey)
}

func TestMapAllResolvedCacheEntry_FilterFailing(t *testing.T) {
	mapper := NewResolverMapper(&mockCacheService{
		prefixMap: map[string]string{"ok": "x"},
		failMap:   map[string]bool{"fail": true},
	})

	raw := json.RawMessage(`"value"`)

	entries := []*CacheEntry{
		{CacheId: &CacheId{CacheName: "ok", Key: "1"}, Value: &raw},
		{CacheId: &CacheId{CacheName: "fail", Key: "2"}, Value: &raw},
	}

	result := mapper.MapAllResolvedCacheEntry(entries)

	assert.Len(t, result, 1)
	assert.Equal(t, "x:1", result[0].ResolvedCacheId.StorageKey)
	assert.Equal(t, &raw, result[0].Value)
}

func TestMapAllCacheEntryHit_Mixed(t *testing.T) {
	v := json.RawMessage(`"v"`)

	hits := []*ResolvedCacheHit{
		{
			ResolvedCacheEntry: &ResolvedCacheEntry{
				ResolvedCacheId: &ResolvedCacheId{
					CacheId:    &CacheId{CacheName: "ok", Key: "1"},
					StorageKey: "x:1",
				},
				Value: &v,
			},
			Found: true,
		},
		{
			ResolvedCacheEntry: &ResolvedCacheEntry{
				ResolvedCacheId: &ResolvedCacheId{
					CacheId:    &CacheId{CacheName: "fail", Key: "2"},
					StorageKey: "f:2",
				},
				Value: &v,
			},
			Found: false,
		},
	}

	result := NewResolverMapper(nil).MapAllCacheEntryHit(hits)

	assert.Len(t, result, 2)
	assert.Equal(t, true, result[0].Found)
	assert.Equal(t, false, result[1].Found)
	assert.Equal(t, "ok", result[0].CacheEntry.CacheId.CacheName)
	assert.Equal(t, "fail", result[1].CacheEntry.CacheId.CacheName)
}
