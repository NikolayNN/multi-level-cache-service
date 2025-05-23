package config

import "fmt"

type LayerMode string

const (
	LayerModeDisabled  LayerMode = "disabled"
	LayerModeReadonly  LayerMode = "readonly"
	LayerModeReadwrite LayerMode = "readwrite"
)

type Layer struct {
	Name string    `yaml:"name"`
	Mode LayerMode `yaml:"mode"`
}

type Layers struct {
	Layers []Layer `yaml:"layers"`
}

func (m *LayerMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	switch s {
	case string(LayerModeDisabled), string(LayerModeReadonly), string(LayerModeReadwrite):
		*m = LayerMode(s)
		return nil
	default:
		return fmt.Errorf("invalid cache layer mode: %q", s)
	}
}
