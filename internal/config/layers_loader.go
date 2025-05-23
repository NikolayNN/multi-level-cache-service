package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

var (
	layersOnce sync.Once
	layersCfg  *Layers
	layersErr  error
)

func loadLayers(path string) (*Layers, error) {
	LoadOnce(&layersOnce, &layersCfg, &layersErr, path, func(data []byte) (*Layers, error) {
		var cfg Layers
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("cannot parse layers config: %w", err)
		}
		return &cfg, nil
	})
	return layersCfg, layersErr
}
