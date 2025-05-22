package prefix_test

import (
	"aur-cache-service/internal/config/caches"
	"aur-cache-service/internal/prefix"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService_OK(t *testing.T) {
	cfg := &caches.CacheConfigsStorage{
		Configs: map[string]caches.Config{
			"user":  {Prefix: "u"},
			"order": {Prefix: "o"},
		},
	}

	s := prefix.New(cfg)

	require.NotNil(t, s)
	require.Equal(t, "u:test", s.ToCacheKey("user", "test"))
	require.Equal(t, "o:123", s.ToCacheKey("order", "123"))
}

func TestNewService_PanicIfNil(t *testing.T) {
	require.PanicsWithValue(t, "cacheConfigStorage is nil", func() {
		prefix.New(nil)
	})
}

func TestToCacheKey_PanicIfUnknownCache(t *testing.T) {
	cfg := &caches.CacheConfigsStorage{
		Configs: map[string]caches.Config{
			"user": {Prefix: "u"},
		},
	}
	s := prefix.New(cfg)

	require.PanicsWithValue(t, "unknown cache: not_found", func() {
		s.ToCacheKey("not_found", "key")
	})
}
