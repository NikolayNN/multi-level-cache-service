package clients

import (
	"aur-cache-service/internal/config"
	"time"
)

type Config struct {
	Ristretto RistrettoConfig `yaml:"ristretto"`
	Redis     RedisConfig     `yaml:"redis"`
	RocksDB   RocksDBConfig   `yaml:"rocksdb"`
}

type RistrettoConfig struct {
	Enabled     bool          `yaml:"enabled"`
	NumCounters int64         `yaml:"numCounters"`
	BufferItems int           `yaml:"bufferItems"`
	MaxCost     string        `yaml:"maxCost"`
	DefaultTTL  time.Duration `yaml:"defaultTTL"`
}

func (r *RistrettoConfig) MaxCostBytes() uint64 {
	return config.ParseBytesStr(r.MaxCost, "ristretto -> maxCost")
}

type RedisConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	PoolSize int           `yaml:"poolSize"`
	Timeout  time.Duration `yaml:"timeout"`
}

type RocksDBConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Path            string `yaml:"path"`
	CreateIfMissing bool   `yaml:"createIfMissing"`
	MaxOpenFiles    int    `yaml:"maxOpenFiles"`
	BlockSize       string `yaml:"blockSize"`
	BlockCache      string `yaml:"blockCache"`
	WriteBufferSize string `yaml:"writeBufferSize"`
}

func (r *RocksDBConfig) BlockSizeBytes() uint64 {
	return config.ParseBytesStr(r.BlockSize, "rocksDb -> blockSize")
}

func (r *RocksDBConfig) BlockCacheBytes() uint64 {
	return config.ParseBytesStr(r.BlockCache, "rocksDb -> blockCache")
}

func (r *RocksDBConfig) WriteBufferSizeBytes() uint64 {
	return config.ParseBytesStr(r.WriteBufferSize, "rocksDb -> writeBufferSize")
}
