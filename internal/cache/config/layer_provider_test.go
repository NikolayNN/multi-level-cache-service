package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLayerProviderService_Success(t *testing.T) {
	cfg := &AppConfig{
		Provider: []Provider{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  10000,
				BufferItems:  64,
				MaxCost:      "128MB",
				DefaultTTL:   time.Second,
			},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
		},
	}

	svc := NewLayerProviderService(cfg)
	assert.Len(t, svc.layerProviders, 1)
	assert.Equal(t, LayerModeEnabled, svc.layerProviders[0].Mode)
	assert.Equal(t, "mem", svc.layerProviders[0].Provider.GetName())
}

func TestNewLayerProviderService_Success_WithExpected(t *testing.T) {
	cfg := &AppConfig{
		Provider: []Provider{
			&Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  10000,
				BufferItems:  64,
				MaxCost:      "128MB",
				DefaultTTL:   time.Second,
			},
		},
		Layers: []Layer{
			{Name: "mem", Mode: LayerModeEnabled},
		},
	}

	svc := NewLayerProviderService(cfg)

	expected := []*LayerProvider{
		{
			Mode: LayerModeEnabled,
			Provider: &Ristretto{
				ProviderMeta: ProviderMeta{Name: "mem", Type: ProviderTypeRistretto},
				NumCounters:  10000,
				BufferItems:  64,
				MaxCost:      "128MB",
				DefaultTTL:   time.Second,
			},
		},
	}

	assert.Equal(t, expected, svc.layerProviders)
}

func TestNewLayerProviderService_PanicWhenProviderMissing(t *testing.T) {
	cfg := &AppConfig{
		Provider: []Provider{}, // пустой список провайдеров
		Layers: []Layer{
			{Name: "missing", Mode: LayerModeEnabled},
		},
	}

	assert.PanicsWithError(t, "can't create layer providers. can't find provider with name: \"missing\"", func() {
		_ = NewLayerProviderService(cfg)
	})
}
