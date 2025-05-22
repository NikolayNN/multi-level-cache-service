package config

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"strings"
	"sync"
)

func LoadOnce[T any](once *sync.Once, cache **T, loadErr *error, path string, unmarshalFn func([]byte) (*T, error)) {
	once.Do(func() {
		data, err := os.ReadFile(path)
		if err != nil {
			*loadErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		obj, err := unmarshalFn(data)
		if err != nil {
			*loadErr = err
			return
		}

		*cache = obj
	})
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
