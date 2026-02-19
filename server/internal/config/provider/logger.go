package provider

import (
	"log/slog"
	"os"
)

func NewLoggerProvider(env *EnvProvider) *slog.Logger {
	level := slog.LevelDebug
	if env.appEnv == "production" {
		level = slog.LevelInfo
	}

	loggerOpts := slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, &loggerOpts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
