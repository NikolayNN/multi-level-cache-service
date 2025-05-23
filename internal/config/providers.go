package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type ProviderType string

const (
	ProviderTypeRistretto ProviderType = "ristretto"
	ProviderTypeRedis     ProviderType = "redis"
	ProviderTypeRocksDb   ProviderType = "rocksdb"
)

/* ---------- общее ядро ---------- */

type ProviderMeta struct {
	Name string       `yaml:"name"`
	Type ProviderType `yaml:"type"`
}

func (m ProviderMeta) GetName() string       { return m.Name }
func (m ProviderMeta) GetType() ProviderType { return m.Type }

/* ---------- интерфейс ---------- */

type Provider interface {
	GetName() string
	GetType() ProviderType
}

/* ---------- конкретные типы ---------- */

type Ristretto struct {
	ProviderMeta `yaml:",inline"`

	NumCounters int64         `yaml:"numCounters"`
	BufferItems int64         `yaml:"bufferItems"`
	MaxCost     string        `yaml:"maxCost"`
	DefaultTTL  time.Duration `yaml:"defaultTTL"`
}

func (r *Ristretto) MaxCostBytes() uint64 {
	return ParseBytesStr(r.MaxCost, r.Name+" -> maxCost")
}

type Redis struct {
	ProviderMeta `yaml:",inline"`

	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	PoolSize int           `yaml:"poolSize"`
	Timeout  time.Duration `yaml:"timeout"`
}

type RocksDB struct {
	ProviderMeta `yaml:",inline"`

	Path            string `yaml:"path"`
	CreateIfMissing bool   `yaml:"createIfMissing"`
	MaxOpenFiles    int    `yaml:"maxOpenFiles"`
	BlockSize       string `yaml:"blockSize"`
	BlockCache      string `yaml:"blockCache"`
	WriteBufferSize string `yaml:"writeBufferSize"`
}

func (r *RocksDB) BlockSizeBytes() uint64 {
	return ParseBytesStr(r.BlockSize, r.Name+" -> blockSize")
}
func (r *RocksDB) BlockCacheBytes() uint64 {
	return ParseBytesStr(r.BlockCache, r.Name+" -> blockCache")
}

func (r *RocksDB) WriteBufferSizeBytes() uint64 {
	return ParseBytesStr(r.WriteBufferSize, r.Name+" -> writeBufferSize")
}

/* ---------- оболочка для YAML ---------- */

type ProvidersConfig struct {
	Providers []Provider `yaml:"providers"`
}

/* ---------- кастомный Unmarshal ---------- */

func (pt *ProviderType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	switch s {
	case string(ProviderTypeRistretto), string(ProviderTypeRedis), string(ProviderTypeRocksDb):
		*pt = ProviderType(s)
		return nil
	default:
		return fmt.Errorf("unknown provider type: %q", s)
	}
}

func (pc *ProvidersConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Providers []yaml.Node `yaml:"providers"`
	}
	if err := value.Decode(&raw); err != nil {
		return err
	}

	for _, n := range raw.Providers {
		var meta ProviderMeta
		if err := n.Decode(&meta); err != nil {
			return err
		}

		var p Provider
		switch meta.Type {
		case ProviderTypeRistretto:
			p = &Ristretto{}
		case ProviderTypeRedis:
			p = &Redis{}
		case ProviderTypeRocksDb:
			p = &RocksDB{}
		}

		if err := n.Decode(p); err != nil {
			return err
		}
		pc.Providers = append(pc.Providers, p)
	}
	return nil
}
