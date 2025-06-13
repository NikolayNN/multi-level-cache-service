package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidate_FullValidConfig(t *testing.T) {
	tmp := t.TempDir()
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "l0", Type: ProviderTypeRistretto},
				NumCounters:  10000,
				BufferItems:  64,
				MaxCost:      "128MB",
				DefaultTTL:   time.Second,
			},
			&Redis{
				ProviderMeta: ProviderMeta{Name: "l1", Type: ProviderTypeRedis},
				Host:         "localhost",
				Port:         6379,
				PoolSize:     10,
				Timeout:      time.Second,
			},
			&RocksDB{
				ProviderMeta:    ProviderMeta{Name: "l2", Type: ProviderTypeRocksDb},
				Path:            tmp,
				MaxOpenFiles:    100,
				BlockSize:       "4KB",
				BlockCache:      "64MB",
				WriteBufferSize: "8MB",
				CreateIfMissing: true,
			},
		},
		Layers: []Layer{
			{Name: "l0", Mode: LayerModeEnabled},
			{Name: "l1", Mode: LayerModeEnabled},
			{Name: "l2", Mode: LayerModeEnabled},
		},
		Caches: []Cache{
			{
				Name:   "sample-cache",
				Prefix: "sc",
				Layers: []CacheLayerConfig{
					{Enabled: true, TTL: time.Second},
					{Enabled: true, TTL: 2 * time.Second},
					{Enabled: false, TTL: 0},
				},
				Api: ApiConfig{
					Enabled: true,
					GetBatch: ApiBatchConfig{
						URL:     "http://localhost/api/sc",
						Prop:    "id",
						KeyType: KeyTypeNumber,
						Timeout: 3 * time.Second,
					},
				},
			},
		},
	}

	assert.NoError(t, appCfg.Validate())
}

func TestValidate_SuccessRistretto(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  1000,
				BufferItems:  64,
				MaxCost:      "128MB",
				DefaultTTL:   time.Second,
			},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
		},
		Caches: []Cache{
			{
				Name:   "users",
				Prefix: "u",
				Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}},
				Api: ApiConfig{
					Enabled: true,
					GetBatch: ApiBatchConfig{
						URL:     "http://localhost/api/users",
						Prop:    "id",
						KeyType: KeyTypeString,
						Timeout: time.Second,
					},
				},
			},
		},
	}

	assert.NoError(t, appCfg.Validate())
}

func TestValidate_DuplicateProviderName(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Redis{
				ProviderMeta: ProviderMeta{Name: "dup", Type: ProviderTypeRedis},
				Host:         "localhost",
				Port:         6379,
				PoolSize:     10,
				Timeout:      time.Second,
			},
			&Redis{
				ProviderMeta: ProviderMeta{Name: "dup", Type: ProviderTypeRedis},
				Host:         "localhost",
				Port:         6379,
				PoolSize:     10,
				Timeout:      time.Second,
			},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "duplicate name")
}

func TestValidate_InvalidRocksDBPath(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&RocksDB{
				ProviderMeta:    ProviderMeta{Name: "db", Type: ProviderTypeRocksDb},
				Path:            "/invalid-path",
				MaxOpenFiles:    100,
				BlockSize:       "4KB",
				BlockCache:      "64MB",
				WriteBufferSize: "8MB",
				CreateIfMissing: false,
			},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "does not exist and createIfMissing=false")
}

func TestValidate_InvalidLayerProviderRef(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{},
		Layers: []Layer{
			{Name: "unknown", Mode: LayerModeEnabled},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "no matching provider found")
}

func TestValidate_MismatchedCacheLayers(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  100, BufferItems: 10, MaxCost: "1MB",
			},
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem2", Type: ProviderTypeRistretto},
				NumCounters:  100, BufferItems: 10, MaxCost: "1MB",
			},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
			{Name: "mem2", Mode: LayerModeDisabled},
		},
		Caches: []Cache{
			{
				Name:   "short",
				Prefix: "s",
				Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}},
			},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "number of cache layers")
}

func TestValidate_InvalidApiURL(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  10, BufferItems: 10, MaxCost: "1MB",
			},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
		},
		Caches: []Cache{
			{
				Name:   "fail-api",
				Prefix: "f",
				Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}},
				Api: ApiConfig{
					Enabled: true,
					GetBatch: ApiBatchConfig{
						URL:     "://bad-url",
						Prop:    "id",
						KeyType: KeyTypeString,
						Timeout: time.Second,
					},
				},
			},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "invalid api.getBatch.url")
}

func TestValidate_RistrettoFailures(t *testing.T) {
	tests := []struct {
		name string
		r    *Ristretto
		err  string
	}{
		{"zero numCounters", &Ristretto{
			ProviderMeta: ProviderMeta{Name: "r1", Type: ProviderTypeRistretto},
			NumCounters:  0, BufferItems: 10, MaxCost: "1MB"}, "numCounters must be > 0"},
		{"zero bufferItems", &Ristretto{
			ProviderMeta: ProviderMeta{Name: "r2", Type: ProviderTypeRistretto},
			NumCounters:  10, BufferItems: 0, MaxCost: "1MB"}, "bufferItems must be > 0"},
		{"bad maxCost", &Ristretto{
			ProviderMeta: ProviderMeta{Name: "r3", Type: ProviderTypeRistretto},
			NumCounters:  10, BufferItems: 10, MaxCost: "abc"}, "invalid maxCost"},
		{"negative defaultTTL", &Ristretto{
			ProviderMeta: ProviderMeta{Name: "r4", Type: ProviderTypeRistretto},
			NumCounters:  10, BufferItems: 10, MaxCost: "1MB", DefaultTTL: -time.Second}, "defaultTTL must be >="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCfg := AppConfigIntermediary{
				Providers: Providers{tt.r},
			}
			err := appCfg.Validate()
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestValidate_RedisFailures(t *testing.T) {
	tests := []struct {
		name string
		r    *Redis
		err  string
	}{
		{"empty host", &Redis{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRedis},
			Host:         "", Port: 6379, PoolSize: 10, Timeout: time.Second}, "host is required"},
		{"bad port", &Redis{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRedis},
			Host:         "localhost", Port: 70000, PoolSize: 10, Timeout: time.Second}, "port must be 1..65535"},
		{"zero poolSize", &Redis{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRedis},
			Host:         "localhost", Port: 6379, PoolSize: 0, Timeout: time.Second}, "poolSize must be > 0"},
		{"zero timeout", &Redis{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRedis},
			Host:         "localhost", Port: 6379, PoolSize: 10, Timeout: 0}, "timeout must be > 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCfg := AppConfigIntermediary{
				Providers: Providers{tt.r},
			}
			err := appCfg.Validate()
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestValidate_RocksDB_EmptyFields(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name string
		r    *RocksDB
		err  string
	}{
		{"missing path", &RocksDB{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRocksDb},
			Path:         "", MaxOpenFiles: 10, CreateIfMissing: false}, "path is required"},
		{"missing dir", &RocksDB{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRocksDb},
			Path:         "/non/existing", MaxOpenFiles: 10, CreateIfMissing: false}, "does not exist"},
		{"non-dir path", &RocksDB{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRocksDb},
			Path:         tmp + "/file", MaxOpenFiles: 10, CreateIfMissing: false}, "not a directory"},
		{"blockSize empty + !create", &RocksDB{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRocksDb},
			Path:         tmp, MaxOpenFiles: 10, BlockSize: "", CreateIfMissing: false}, "blockSize is empty"},
		{"blockSize invalid", &RocksDB{
			ProviderMeta: ProviderMeta{Name: "r", Type: ProviderTypeRocksDb},
			Path:         tmp, MaxOpenFiles: 10, BlockSize: "abc", CreateIfMissing: true}, "blockSize"},
	}

	_ = os.WriteFile(tmp+"/file", []byte("data"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCfg := AppConfigIntermediary{
				Providers: Providers{tt.r},
			}
			err := appCfg.Validate()
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestValidate_LayerFailures(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto}, NumCounters: 10, BufferItems: 10, MaxCost: "1MB"},
		},
		Layers: []Layer{
			{Name: "", Mode: LayerModeEnabled},
			{Name: "mem", Mode: LayerModeUnknown},
			{Name: "mem", Mode: LayerModeEnabled},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "name is required")
}

func TestValidate_CacheFailures(t *testing.T) {
	appCfg := AppConfigIntermediary{
		Providers: Providers{
			&Ristretto{ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto}, NumCounters: 10, BufferItems: 10, MaxCost: "1MB"},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
		},
		Caches: []Cache{
			{Name: "", Prefix: "p", Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}}},
			{Name: "c", Prefix: "", Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}}},
			{Name: "c", Prefix: "p", Layers: []CacheLayerConfig{{Enabled: true, TTL: time.Second}}},
			{Name: "c2", Prefix: "p", Layers: []CacheLayerConfig{{Enabled: true, TTL: -1}}},
		},
	}

	err := appCfg.Validate()
	assert.ErrorContains(t, err, "name is required")
}
