package logger

import (
	"go.uber.org/zap"
)

// Init initializes a global zap logger and returns it.
func Init() (*zap.Logger, error) {
	l, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(l)
	return l, nil
}

// Sync flushes any buffered log entries.
func Sync() {
	_ = zap.L().Sync()
}
