package caches

import (
	"time"
)

type CacheLevelConfig struct {
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

type Config struct {
	Name   string       `yaml:"name"`
	Prefix string       `yaml:"prefix"`
	Levels LevelsConfig `yaml:"levels"`
	Api    ApiConfig    `yaml:"Api"`
}

type LevelsConfig struct {
	L0 CacheLevelConfig `yaml:"l0"`
	L1 CacheLevelConfig `yaml:"l1"`
	L2 CacheLevelConfig `yaml:"l2"`
}

type Configs struct {
	Caches []Config `yaml:"caches"`
}

type CacheConfigsStorage struct {
	Configs map[string]Config
}

func (s *CacheConfigsStorage) Get(name string) (Config, bool) {
	cfg, ok := s.Configs[name]
	return cfg, ok
}
