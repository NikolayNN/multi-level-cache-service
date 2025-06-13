package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestLoadAppConfig_Success(t *testing.T) {
	yaml := `
providers:
  - name: l0
    type: ristretto
    numCounters: 10000
    bufferItems: 64
    maxCost: 128MB
    defaultTTL: 1s

  - name: l1
    type: redis
    host: localhost
    port: 6379
    poolSize: 10
    timeout: 1s

  - name: l2
    type: rocksdb
    path: ./tmp
    createIfMissing: true
    maxOpenFiles: 100
    blockSize: 4KB
    blockCache: 64MB
    writeBufferSize: 8MB

layers:
  - name: l0
    mode: enabled
  - name: l1
    mode: enabled
  - name: l2
    mode: enabled

caches:
  - name: test-cache
    prefix: tc
    layers:
      - enabled: true
        ttl: 1s
      - enabled: true
        ttl: 2s
      - enabled: false
        ttl: 0s
    api:
      enabled: true
      getBatch:
        url: http://localhost/api/test
        prop: id
        keyType: number
        timeout: 3s
`

	tmpFile := t.TempDir() + "/yml"
	err := os.WriteFile(tmpFile, []byte(yaml), 0644)
	assert.NoError(t, err)

	got, err := LoadAppConfig(tmpFile)
	assert.NoError(t, err)

	expected := &AppConfig{
		Provider: []Provider{
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
				Path:            "./tmp",
				CreateIfMissing: true,
				MaxOpenFiles:    100,
				BlockSize:       "4KB",
				BlockCache:      "64MB",
				WriteBufferSize: "8MB",
			},
		},
		Layers: []Layer{
			{Name: "l0", Mode: LayerModeEnabled},
			{Name: "l1", Mode: LayerModeEnabled},
			{Name: "l2", Mode: LayerModeEnabled},
		},
		Caches: []Cache{
			{
				Name:   "test-cache",
				Prefix: "tc",
				Layers: []CacheLayerConfig{
					{Enabled: true, TTL: time.Second},
					{Enabled: true, TTL: 2 * time.Second},
					{Enabled: false, TTL: 0},
				},
				Api: ApiConfig{
					Enabled: true,
					GetBatch: ApiBatchConfig{
						URL:     "http://localhost/api/test",
						Prop:    "id",
						KeyType: KeyTypeNumber,
						Timeout: 3 * time.Second,
					},
				},
			},
		},
	}

	assert.Equal(t, expected, got)
}
