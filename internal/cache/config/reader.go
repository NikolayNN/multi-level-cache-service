package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type AppConfig struct {
	Provider []Provider
	Layers   []Layer
	Caches   []Cache
}

func loadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	var interm AppConfigIntermediary
	if err := yaml.Unmarshal(data, &interm); err != nil {
		return nil, fmt.Errorf("yaml unmarshal error: %w", err)
	}

	if err := interm.Validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	return &AppConfig{
		Provider: interm.Providers,
		Layers:   interm.Layers,
		Caches:   interm.Caches,
	}, nil
}
