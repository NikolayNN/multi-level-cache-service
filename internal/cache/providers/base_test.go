package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcChunkSize(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		min      int
		max      int
		expected int
	}{
		{"Below max size", 5, 2, 10, 5},
		{"Equal to max size", 10, 2, 10, 10},
		{"Above max size", 25, 2, 10, 9},
		{"Huge size", 1000, 50, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := calcChunkSize(tt.total, tt.min, tt.max)
			assert.GreaterOrEqual(t, size, tt.min)
			assert.LessOrEqual(t, size, tt.max)
		})
	}
}

func TestSplitKeysToChunks(t *testing.T) {
	keys := make([]string, 25)
	for i := range keys {
		keys[i] = "key" + string(rune('A'+i))
	}

	chunks := splitKeysToChunks(keys, 3, 10)
	total := 0
	for _, chunk := range chunks {
		assert.GreaterOrEqual(t, len(chunk), 1)
		assert.LessOrEqual(t, len(chunk), 10)
		total += len(chunk)
	}
	assert.Equal(t, len(keys), total)
}

func TestSplitKeyValueToChunks(t *testing.T) {
	data := map[string]string{}
	for i := 0; i < 25; i++ {
		data[string(rune('A'+i))] = string(rune('a' + i))
	}

	chunks := splitKeyValueToChunks(data, 3, 10)
	total := 0
	for _, chunk := range chunks {
		assert.GreaterOrEqual(t, len(chunk), 1)
		assert.LessOrEqual(t, len(chunk), 10)
		total += len(chunk)
	}
	assert.Equal(t, len(data), total)
}

func TestEmptyInputs(t *testing.T) {
	assert.Empty(t, splitKeysToChunks(nil, 1, 5))
	assert.Empty(t, splitKeysToChunks([]string{}, 1, 5))
	assert.Empty(t, splitKeyValueToChunks(nil, 1, 5))
	assert.Empty(t, splitKeyValueToChunks(map[string]string{}, 1, 5))
}
