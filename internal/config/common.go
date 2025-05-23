package config

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"strings"
)

func loadFile[T any](path string, unmarshalFn func([]byte) (*T, error)) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	obj, err := unmarshalFn(data)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func ParseByteSize(s string) (uint64, error) {
	return humanize.ParseBytes(strings.TrimSpace(s))
}

func ParseBytesStr(bytesString string, errorPath string) uint64 {
	bytes, err := ParseByteSize(bytesString)
	if err != nil {
		panic(fmt.Sprintf("invalid config -> %v : %v has wrong value (%v)", errorPath, bytesString, err))
	}
	return bytes
}
