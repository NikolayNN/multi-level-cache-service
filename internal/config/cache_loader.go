package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

var (
	cacheOnce sync.Once
	cacheCfg  *CacheStorage
	cacheErr  error
)

func LoadCacheStorage(path string) (*CacheStorage, error) {
	LoadOnce(&cacheOnce, &cacheCfg, &cacheErr, path, func(data []byte) (*CacheStorage, error) {
		var cfg Configs
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
		}

		m := make(map[string]Cache, len(cfg.Caches))
		for _, c := range cfg.Caches {
			m[c.Name] = c
		}

		return &CacheStorage{Configs: m}, nil
	})
	return cacheCfg, cacheErr
}
