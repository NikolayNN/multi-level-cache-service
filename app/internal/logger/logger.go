package logger

import (
	"go.uber.org/zap"

	"telegram-alerts-go/config"
	"telegram-alerts-go/loghook"
	"telegram-alerts-go/telegram"
)

// Init initializes a global zap logger and returns it.
func Init() (*zap.Logger, error) {
	l, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	cfg := config.LoadFromEnv()
	if cfg.BotToken != "" && cfg.ChannelID != "" {
		client := telegram.NewClient(cfg.BotToken, cfg.ChannelID)
		l = l.WithOptions(loghook.NewTelegramHook(client, cfg.ServiceName))
	}
	zap.ReplaceGlobals(l)
	return l, nil
}

// Sync flushes any buffered log entries.
func Sync() {
	_ = zap.L().Sync()
}
