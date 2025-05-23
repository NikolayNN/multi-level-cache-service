package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var pathLayerProvider = "testData/layer_provider_test.yml"
var pathLayerProviderFail = "testData/layer_provider_fail_test.yml"

func TestLoadLayerProviders_Success(t *testing.T) {

	actual, err := LoadLayersProviders(pathLayerProvider)

	if err != nil {
		t.Fatalf("loadLayers failed: %v", err)
	}

	for _, p := range actual {
		fmt.Printf("%+v\n", p)
		fmt.Printf("%+v\n", p.Provider)
	}

	assert.Equal(t, 3, len(actual))

	assert.Equal(t, LayerModeDisabled, actual[0].Mode)
	assert.Equal(t, LayerModeEnabled, actual[1].Mode)
	assert.Equal(t, LayerModeEnabled, actual[2].Mode)

	assert.Equal(t, ProviderTypeRistretto, actual[0].Provider.GetType())
	assert.Equal(t, ProviderTypeRedis, actual[1].Provider.GetType())
	assert.Equal(t, ProviderTypeRocksDb, actual[2].Provider.GetType())
}

func TestLoadLayerProviders_UnknownProviderName(t *testing.T) {

	_, err := LoadLayersProviders(pathLayerProviderFail)

	require.Error(t, err)
	require.Contains(t, err.Error(), "can't find provider with name: \"unknown_provider_name\"")
}
