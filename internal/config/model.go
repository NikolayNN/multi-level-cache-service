package config

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

type AppConfigIntermediary struct {
	Providers Providers `yaml:"providers"`
	Layers    []Layer   `yaml:"layers"`
	Caches    []Cache   `yaml:"caches"`
}

///////////////////////////////////////////////////////////
/// Providers structs
///////////////////////////////////////////////////////////

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

type Providers []Provider

func (p *Providers) UnmarshalYAML(value *yaml.Node) error {
	var raw []yaml.Node
	if err := value.Decode(&raw); err != nil {
		return err
	}

	for _, n := range raw {
		var meta ProviderMeta
		if err := n.Decode(&meta); err != nil {
			return err
		}

		var prov Provider
		switch meta.Type {
		case ProviderTypeRistretto:
			prov = &Ristretto{}
		case ProviderTypeRedis:
			prov = &Redis{}
		case ProviderTypeRocksDb:
			prov = &RocksDB{}
		default:
			return fmt.Errorf("unknown provider type: %q", meta.Type)
		}

		if err := n.Decode(prov); err != nil {
			return err
		}
		*p = append(*p, prov)
	}
	return nil
}

///////////////////////////////////////////////////////////
/// Layers structs
///////////////////////////////////////////////////////////

type LayerMode string

const (
	LayerModeDisabled LayerMode = "disabled"
	LayerModeEnabled  LayerMode = "enabled"
)

type Layer struct {
	Name string    `yaml:"name"`
	Mode LayerMode `yaml:"mode"`
}

func (m *LayerMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	switch s {
	case string(LayerModeDisabled), string(LayerModeEnabled):
		*m = LayerMode(s)
		return nil
	default:
		return fmt.Errorf("invalid cache layer mode: %q", s)
	}
}

///////////////////////////////////////////////////////////
/// Caches structs
///////////////////////////////////////////////////////////

type CacheLayerConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
}

type ApiBatchConfig struct {
	URL  string `yaml:"url"`
	Prop string `yaml:"prop"`
}

type ApiConfig struct {
	Enabled  bool           `yaml:"enabled"`
	GetBatch ApiBatchConfig `yaml:"getBatch"`
}

type Cache struct {
	Name   string             `yaml:"name"`
	Prefix string             `yaml:"prefix"`
	Layers []CacheLayerConfig `yaml:"layers"`
	Api    ApiConfig          `yaml:"api"`
}

///////////////////////////////////////////////////////////
/// UTILS
///////////////////////////////////////////////////////////

func ParseByteSize(s string) (uint64, error) {
	return humanize.ParseBytes(strings.TrimSpace(s))
}

func ParseBytesStr(bytesString string, errorPath string) uint64 {
	bytes, err := ParseByteSize(bytesString)
	if err != nil {
		panic(fmt.Sprintf("invalid config -> %v : %v has wrong value (%v)", errorPath, bytesString, err))
	}
	return bytes
}
