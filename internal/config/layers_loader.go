package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

var (
	layersOnce sync.Once
	layersCfg  *CacheLayers
	layersErr  error
)

func LoadLayers(path string) (*CacheLayers, error) {
	LoadOnce(&layersOnce, &layersCfg, &layersErr, path, func(data []byte) (*CacheLayers, error) {
		var cfg CacheLayers
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("cannot parse layers config: %w", err)
		}
		return &cfg, nil
	})
	return layersCfg, layersErr
}
