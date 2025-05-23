package config

import "fmt"

type CacheLayerMode string

const (
	CacheLayerModeDisabled  CacheLayerMode = "disabled"
	CacheLayerModeReadonly  CacheLayerMode = "readonly"
	CacheLayerModeReadwrite CacheLayerMode = "readwrite"
)

type CacheLayer struct {
	Name string         `yaml:"name"`
	Mode CacheLayerMode `yaml:"mode"`
}

type CacheLayers struct {
	Layers []CacheLayer `yaml:"layers"`
}

func (m *CacheLayerMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	switch s {
	case string(CacheLayerModeDisabled), string(CacheLayerModeReadonly), string(CacheLayerModeReadwrite):
		*m = CacheLayerMode(s)
		return nil
	default:
		return fmt.Errorf("invalid cache layer mode: %q", s)
	}
}
