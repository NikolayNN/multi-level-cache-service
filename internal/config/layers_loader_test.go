package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var pathConfigFile = "testData/layers_config_test.yml"
var pathConfigFileWrongMode = "testData/layers_config_wrong_mode_test.yml"

func TestLoadLayers_Success(t *testing.T) {

	actual, err := loadLayers(pathConfigFile)

	if err != nil {
		t.Fatalf("loadLayers failed: %v", err)
	}

	for _, p := range actual.Layers {
		fmt.Printf("%+v\n", p)
	}

	expectedl0 := Layer{
		Name: "ristretto-l0",
		Mode: LayerModeEnabled,
	}

	assert.Equal(t, expectedl0, actual.Layers[0])

	expectedl1 := Layer{
		Name: "redis-l1",
		Mode: LayerModeEnabled,
	}

	assert.Equal(t, expectedl1, actual.Layers[1])

	expectedl2 := Layer{
		Name: "rocksdb-l2",
		Mode: LayerModeEnabled,
	}

	assert.Equal(t, expectedl2, actual.Layers[2])
}

func TestLoadLayers_WrongMode(t *testing.T) {

	_, err := loadLayers(pathConfigFileWrongMode)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid cache layer mode: \"wrongmode\"")
}
