package clients

import (
	"aur-cache-service/internal/config"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	once          sync.Once
	configOnce    *Config
	configOnceErr error
)

func LoadConfig(path string) (*Config, error) {
	config.LoadOnce(&once, &configOnce, &configOnceErr, path, func(data []byte) (*Config, error) {
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("cannot parse client config: %w", err)
		}
		return &cfg, nil
	})
	return configOnce, configOnceErr
}
