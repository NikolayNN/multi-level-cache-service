package config

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"strings"
	"time"
)

type AppConfigIntermediary struct {
	Providers Providers `yaml:"providers"`
	Layers    []Layer   `yaml:"layers"`
	Caches    []Cache   `yaml:"caches"`
}

func (c *AppConfigIntermediary) Validate() error {
	if err := c.validateProviders(); err != nil {
		return err
	}

	if err := c.validateLayers(); err != nil {
		return err
	}

	if err := c.validateCaches(); err != nil {
		return err
	}
	return nil
}

func (c *AppConfigIntermediary) validateProviders() error {
	providerNames := make(map[string]bool)
	for i, p := range c.Providers {
		if p.GetType() == ProviderTypeUnknown {
			return fmt.Errorf("provider[%d]: unknown type '%s'", i, p.GetType())
		}
		if p.GetName() == "" {
			return fmt.Errorf("provider[%d]: name is required", i)
		}
		if providerNames[p.GetName()] {
			return fmt.Errorf("provider[%d]: duplicate name '%s'", i, p.GetName())
		}
		providerNames[p.GetName()] = true

		// -------- специфичные проверки -------------------------------------
		switch v := p.(type) {

		case *Ristretto:
			if err := c.validateRistretto(i, v); err != nil {
				return err
			}

		case *Redis:
			if err := c.validateRedis(i, v); err != nil {
				return err
			}

		case *RocksDB:
			if err := c.validateRocksDB(i, v); err != nil {
				return err
			}

		default:
			// Это случится, только если появится новый тип и забудут добавить case.
			return fmt.Errorf("provider[%d] (%s): validation not implemented for type %T",
				i, p.GetName(), p)
		}
	}
	return nil
}

func (c *AppConfigIntermediary) validateRistretto(idx int, r *Ristretto) error {
	if r.NumCounters <= 0 {
		return fmt.Errorf("provider[%d] (%s): numCounters must be > 0", idx, r.Name)
	}
	if r.BufferItems <= 0 {
		return fmt.Errorf("provider[%d] (%s): bufferItems must be > 0", idx, r.Name)
	}
	// maxCost должен конвертироваться и быть >0
	if bytes, err := ParseByteSize(r.MaxCost); err != nil || bytes == 0 {
		return fmt.Errorf("provider[%d] (%s): invalid maxCost '%s'", idx, r.Name, r.MaxCost)
	}
	if r.DefaultTTL < 0 {
		return fmt.Errorf("provider[%d] (%s): defaultTTL must be >= 0", idx, r.Name)
	}
	return nil
}

func (c *AppConfigIntermediary) validateRedis(idx int, r *Redis) error {
	if r.Host == "" {
		return fmt.Errorf("provider[%d] (%s): host is required", idx, r.Name)
	}
	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("provider[%d] (%s): port must be 1..65535", idx, r.Name)
	}
	if r.PoolSize <= 0 {
		return fmt.Errorf("provider[%d] (%s): poolSize must be > 0", idx, r.Name)
	}
	if r.Timeout <= 0 {
		return fmt.Errorf("provider[%d] (%s): timeout must be > 0", idx, r.Name)
	}
	return nil
}

func (c *AppConfigIntermediary) validateRocksDB(idx int, r *RocksDB) error {
	if r.Path == "" {
		return fmt.Errorf("provider[%d] (%s): path is required", idx, r.Name)
	}
	if r.MaxOpenFiles <= 0 {
		return fmt.Errorf("provider[%d] (%s): maxOpenFiles must be > 0", idx, r.Name)
	}

	// Проверка существования директории, если CreateIfMissing == false
	if !r.CreateIfMissing {
		info, err := os.Stat(r.Path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("provider[%d] (%s): path '%s' does not exist and createIfMissing=false", idx, r.Name, r.Path)
			}
			return fmt.Errorf("provider[%d] (%s): unable to access path '%s': %v", idx, r.Name, r.Path, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("provider[%d] (%s): path '%s' exists but is not a directory", idx, r.Name, r.Path)
		}
	}

	// Проверка на валидность размеров
	if r.BlockSize != "" {
		if _, err := r.BlockSizeBytes(); err != nil {
			return fmt.Errorf("provider[%d] (%s): %v", idx, r.Name, err)
		}
	} else if !r.CreateIfMissing {
		return fmt.Errorf("provider[%d] (%s): blockSize is empty but createIfMissing=false", idx, r.Name)
	}

	if r.BlockCache != "" {
		if _, err := r.BlockCacheBytes(); err != nil {
			return fmt.Errorf("provider[%d] (%s): %v", idx, r.Name, err)
		}
	} else if !r.CreateIfMissing {
		return fmt.Errorf("provider[%d] (%s): blockCache is empty but createIfMissing=false", idx, r.Name)
	}

	if r.WriteBufferSize != "" {
		if _, err := r.WriteBufferSizeBytes(); err != nil {
			return fmt.Errorf("provider[%d] (%s): %v", idx, r.Name, err)
		}
	} else if !r.CreateIfMissing {
		return fmt.Errorf("provider[%d] (%s): writeBufferSize is empty but createIfMissing=false", idx, r.Name)
	}

	return nil
}

func (c *AppConfigIntermediary) validateLayers() error {
	layerNames := make(map[string]bool)

	providerNames := make(map[string]bool)
	for _, p := range c.Providers {
		providerNames[p.GetName()] = true
	}

	for i, l := range c.Layers {
		if l.Name == "" {
			return fmt.Errorf("layer[%d]: name is required", i)
		}

		if l.Mode == LayerModeUnknown {
			return fmt.Errorf("layer[%d]: invalid mode '%s'", i, l.Mode)
		}

		if layerNames[l.Name] {
			return fmt.Errorf("layer[%d]: duplicate name '%s'", i, l.Name)
		}
		if !providerNames[l.Name] {
			return fmt.Errorf("layer[%d]: no matching provider found for name '%s'", i, l.Name)
		}
		layerNames[l.Name] = true
	}
	return nil
}

func (c *AppConfigIntermediary) validateCaches() error {
	cacheNames := make(map[string]bool)
	prefixes := make(map[string]bool)
	for i, cache := range c.Caches {
		if cache.Name == "" {
			return fmt.Errorf("cache[%d]: name is required", i)
		}
		if cacheNames[cache.Name] {
			return fmt.Errorf("cache[%d]: duplicate cache name '%s'", i, cache.Name)
		}
		cacheNames[cache.Name] = true

		if prefixes[cache.Prefix] {
			return fmt.Errorf("cache[%d]: duplicate prefix '%s'", i, cache.Prefix)
		}
		prefixes[cache.Prefix] = true

		if cache.Prefix == "" {
			return fmt.Errorf("cache[%d]: prefix is required", i)
		}

		if len(cache.Layers) != len(c.Layers) {
			return fmt.Errorf("cache[%d]: number of cache layers (%d) must match global layers (%d)", i, len(cache.Layers), len(c.Layers))
		}

		for j, layer := range cache.Layers {
			if layer.Enabled && layer.TTL < 0 {
				return fmt.Errorf("cache[%d].layer[%d]: TTL must be >= 0 when enabled", i, j)
			}
		}

		if err := c.validateIntegrationApi(i, cache.Api); err != nil {
			return err
		}
	}
	return nil
}

func (c *AppConfigIntermediary) validateIntegrationApi(i int, apiConfig ApiConfig) error {
	if !apiConfig.Enabled {
		return nil
	}

	if apiConfig.GetBatch.Prop == "" {
		return fmt.Errorf("cache[%d]: api.getBatch.prop is required", i)
	}
	if apiConfig.GetBatch.KeyType != KeyTypeString && apiConfig.GetBatch.KeyType != KeyTypeNumber {
		return fmt.Errorf("cache[%d]: invalid keyType '%s'", i, apiConfig.GetBatch.KeyType)
	}
	if apiConfig.GetBatch.URL == "" {
		return fmt.Errorf("cache[%d]: api.getBatch.url is required", i)
	}

	u, err := url.Parse(apiConfig.GetBatch.URL)
	if err != nil {
		return fmt.Errorf("cache[%d]: invalid api.getBatch.url '%s': %v", i, apiConfig.GetBatch.URL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("cache[%d]: unsupported scheme '%s' in api.getBatch.url", i, u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("cache[%d]: missing host in api.getBatch.url '%s'", i, apiConfig.GetBatch.URL)
	}

	if apiConfig.GetBatch.Timeout <= 0 {
		return fmt.Errorf("cache[%d]: api.getBatch.timeout must be > 0", i)
	}
	return nil
}

///////////////////////////////////////////////////////////
/// Providers structs
///////////////////////////////////////////////////////////

type ProviderType string

const (
	ProviderTypeRistretto ProviderType = "ristretto"
	ProviderTypeRedis     ProviderType = "redis"
	ProviderTypeRocksDb   ProviderType = "rocksdb"
	ProviderTypeUnknown   ProviderType = "unknown"
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

func (r *Ristretto) MaxCostBytes() (uint64, error) {
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

type Unknown struct {
	ProviderMeta `yaml:",inline"`
}

func (r *RocksDB) BlockSizeBytes() (uint64, error) {
	return ParseBytesStr(r.BlockSize, r.Name+" -> blockSize")
}
func (r *RocksDB) BlockCacheBytes() (uint64, error) {
	return ParseBytesStr(r.BlockCache, r.Name+" -> blockCache")
}

func (r *RocksDB) WriteBufferSizeBytes() (uint64, error) {
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
			prov = &Unknown{}
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
	LayerModeUnknown  LayerMode = "unknown"
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
	default:
		*m = LayerModeUnknown
	}
	return nil
}

///////////////////////////////////////////////////////////
/// Caches structs
///////////////////////////////////////////////////////////

type CacheLayerConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"` // 0 = no expiration
}

type KeyType string

const (
	KeyTypeString KeyType = "string"
	KeyTypeNumber KeyType = "number"
)

type ApiBatchConfig struct {
	URL     string            `yaml:"url"`
	Prop    string            `yaml:"prop"`
	KeyType KeyType           `yaml:"keyType"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
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

func ParseBytesStr(bytesString string, errorPath string) (uint64, error) {
	bytes, err := ParseByteSize(bytesString)
	if err != nil {
		return 0, fmt.Errorf("invalid config -> %v: %v has wrong value (%v)", errorPath, bytesString, err)
	}
	return bytes, nil
}
