package caches

import (
	"aur-cache-service/internal/config"
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

var (
	once          sync.Once
	configOnce    *CacheConfigsStorage
	configOnceErr error
)

func LoadConfig(path string) (*CacheConfigsStorage, error) {
	config.LoadOnce(&once, &configOnce, &configOnceErr, path, func(data []byte) (*CacheConfigsStorage, error) {
		var cfg Configs
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
		}

		m := make(map[string]Config, len(cfg.Caches))
		for _, c := range cfg.Caches {
			m[c.Name] = c
		}

		return &CacheConfigsStorage{Configs: m}, nil
	})
	return configOnce, configOnceErr
}
