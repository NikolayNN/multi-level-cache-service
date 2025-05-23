package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

func loadLayers(path string) (layers *Layers, err error) {
	layers, err = loadFile(
		path,
		func(data []byte) (*Layers, error) {
			var cfg Layers
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("cannot parse layers config: %w", err)
			}
			return &cfg, nil
		},
	)
	return
}
