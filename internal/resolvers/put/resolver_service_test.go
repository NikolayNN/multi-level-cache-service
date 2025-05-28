package put

import (
	"aur-cache-service/internal/resolvers/cmn"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"aur-cache-service/internal/config"
)

////////////////////////////////////////////////////////////////////////////////
// Test fakes
////////////////////////////////////////////////////////////////////////////////

type fakeCacheService struct{}

var _ config.CacheService = (*fakeCacheService)(nil)

func (fakeCacheService) CacheKey(req cmn.CacheNameKey) string {
	return "u:" + req.GetKey()
}

func (fakeCacheService) GetByName(name string) config.Cache {
	return config.Cache{
		Name:   name,
		Prefix: "u", // doesn’t matter for the tests
		Layers: []config.CacheLayerConfig{
			{Enabled: true, TTL: 15 * time.Second}, // level 0
			{Enabled: true, TTL: 30 * time.Second}, // level 1
			{Enabled: true, TTL: 60 * time.Second}, // level 2
		},
		Api: config.ApiConfig{},
	}
}

func (fakeCacheService) ToCacheKey(cacheName string, key string) string {
	return "u:" + key
}

////////////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////////////

func TestConcreteResolverDisabled(t *testing.T) {
	r := NewConcreteResolverDisabled()
	got := r.resolve([]CacheReq{
		{
			CacheName: "test",
			Key:       "111",
			Value:     "value",
		},
	})
	if len(got) != 0 {
		t.Fatalf("disabled resolver should return empty slice, got %d elements", len(got))
	}
}

func TestCreateConcreteResolvers(t *testing.T) {
	cacheSvc := &fakeCacheService{}
	layerProviders := []config.LayerProvider{
		{Mode: config.LayerModeDisabled}, // level 0 should be disabled
		{Mode: config.LayerModeEnabled},  // level 1 should be active
	}

	resolvers := createConcreteResolvers(cacheSvc, layerProviders)

	if len(resolvers) != 2 {
		t.Fatalf("expected 2 resolvers, got %d", len(resolvers))
	}

	if _, ok := resolvers[0].(*ConcreteResolverDisabled); !ok {
		t.Errorf("expected first resolver to be ConcreteResolverDisabled, got %T", resolvers[0])
	}
	if _, ok := resolvers[1].(*ConcreteResolverLevel); !ok {
		t.Errorf("expected second resolver to be ConcreteResolverLevel, got %T", resolvers[1])
	}
}

func TestConcreteResolverLevel0(t *testing.T) {
	cacheSvc := &fakeCacheService{}
	r := NewConcreteResolverLevel(cacheSvc, 0) // first (enabled) level – TTL 15s

	req := CacheReq{
		CacheName: "test",
		Key:       "1111",
		Value:     "value",
	}
	actual := r.resolve([]CacheReq{req})

	if len(actual) != 1 {
		t.Fatalf("expected one resolved request, actual %d", len(actual))
	}

	expected := CacheReqResolved{
		Req:      &req,
		CacheKey: "u:1111",
		Ttl:      15 * time.Second,
	}

	require.Equal(t, []CacheReqResolved{expected}, actual)
}

func TestConcreteResolverLevel1(t *testing.T) {
	cacheSvc := &fakeCacheService{}
	r := NewConcreteResolverLevel(cacheSvc, 1) // second (enabled) level – TTL 15s

	req := CacheReq{
		CacheName: "test",
		Key:       "1111",
		Value:     "value",
	}
	actual := r.resolve([]CacheReq{req})

	if len(actual) != 1 {
		t.Fatalf("expected one resolved request, actual %d", len(actual))
	}

	expected := CacheReqResolved{
		Req:      &req,
		CacheKey: "u:1111",
		Ttl:      30 * time.Second,
	}

	require.Equal(t, []CacheReqResolved{expected}, actual)
}

func TestConcreteResolverLevel2(t *testing.T) {
	cacheSvc := &fakeCacheService{}
	r := NewConcreteResolverLevel(cacheSvc, 2) // third (enabled) level – TTL 15s

	req := CacheReq{
		CacheName: "test",
		Key:       "1111",
		Value:     "value",
	}
	actual := r.resolve([]CacheReq{req})

	if len(actual) != 1 {
		t.Fatalf("expected one resolved request, actual %d", len(actual))
	}

	expected := CacheReqResolved{
		Req:      &req,
		CacheKey: "u:1111",
		Ttl:      60 * time.Second,
	}

	require.Equal(t, []CacheReqResolved{expected}, actual)
}

func TestMainResolverResolve(t *testing.T) {
	cacheSvc := &fakeCacheService{}
	layerProviders := []config.LayerProvider{
		{Mode: config.LayerModeEnabled}, // level 0
		{Mode: config.LayerModeEnabled}, // level 1
	}

	mainResolver := createMainResolver(cacheSvc, layerProviders)
	reqs := []CacheReq{
		{
			CacheName: "foo",
			Key:       "111",
			Value:     "value1",
		},
		{
			CacheName: "foo",
			Key:       "222",
			Value:     "value1",
		},
	}
	matrix := mainResolver.Resolve(reqs)

	if len(matrix) != 2 {
		t.Fatalf("expected 2 matrix rows, got %d", len(matrix))
	}
	for i, row := range matrix {
		if len(row) == 0 {
			t.Errorf("row %d should not be empty", i)
		}
	}
}
