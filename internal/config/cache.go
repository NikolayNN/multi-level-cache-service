package config

import (
	"time"
)

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

type Configs struct {
	Caches []Cache `yaml:"caches"`
}

type CacheStorage struct {
	Configs map[string]Cache
}

func (s *CacheStorage) Get(name string) (Cache, bool) {
	cfg, ok := s.Configs[name]
	return cfg, ok
}
