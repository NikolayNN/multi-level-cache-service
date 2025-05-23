package config

import (
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	providersOnce sync.Once
	providersCfg  *ProvidersConfig
	providersErr  error
)

func LoadProviders(path string) (*ProvidersConfig, error) {
	LoadOnce(&providersOnce, &providersCfg, &providersErr, path, func(data []byte) (*ProvidersConfig, error) {
		var cfg ProvidersConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("cannot parse providers config: %w", err)
		}
		return &cfg, nil
	})
	return providersCfg, providersErr
}
