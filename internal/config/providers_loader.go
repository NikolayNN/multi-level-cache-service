package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

func LoadProviders(path string) (providers *ProvidersConfig, err error) {
	providers, err = loadFile(
		path,
		func(data []byte) (*ProvidersConfig, error) {
			var cfg ProvidersConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("cannot parse providers config: %w", err)
			}
			return &cfg, nil
		})
	return
}
