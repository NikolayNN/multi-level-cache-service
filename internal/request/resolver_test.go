package request_test

import (
	"aur-cache-service/internal/config"
	"aur-cache-service/internal/prefix"
	"aur-cache-service/internal/request"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve_Success(t *testing.T) {
	prefixService := prefix.New(&config.CacheStorage{
		Configs: map[string]config.Cache{
			"user": {Prefix: "u"},
			"task": {Prefix: "t"},
		},
	})

	resolver := request.NewResolverService(prefixService)

	requests := []request.GetCacheReq{
		{CacheName: "user", Key: "42"},
		{CacheName: "task", Key: "xyz"},
	}

	resolved := resolver.Resolve(requests)

	require.Len(t, resolved, 2)

	require.Equal(t, "user", resolved[0].CacheName())
	require.Equal(t, "42", resolved[0].Key())
	require.Equal(t, "u:42", resolved[0].CacheKey)
}

func TestResolve_SkipUnknownCache(t *testing.T) {
	prefixService := prefix.New(&config.CacheStorage{
		Configs: map[string]config.Cache{
			"user": {Prefix: "u"},
		},
	})

	resolver := request.NewResolverService(prefixService)

	requests := []request.GetCacheReq{
		{CacheName: "user", Key: "1"},
		{CacheName: "unknown", Key: "2"}, // вызовет панику, но мы оборачиваем toResolved
	}

	// переопределим toResolved для безопасного теста — или создадим безопасную версию ToCacheKeySafe()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unknown cache")
		}
	}()

	_ = resolver.Resolve(requests)
}
