package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

const Timeformat = "2006-01-02 15:04:05"

func New(level string, timeformat string) *slog.Logger {
	logLevel := slog.LevelInfo

	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "local":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error", "err":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      logLevel,
		TimeFormat: timeformat,
		AddSource:  logLevel == slog.LevelDebug,
	}))
}
