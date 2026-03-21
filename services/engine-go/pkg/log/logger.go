package log

import (
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
)

const (
	defaultTimeFormat string = "2006-01-02 15:04:05"
	defaultLevel             = slog.LevelInfo
)

const (
	levelDebug string = "debug"
	levelInfo  string = "info"
	levelWarn  string = "warn"
	levelErr   string = "error"
)

var (
	once sync.Once
	log  *slog.Logger
)

func NewDefaultLogger() *slog.Logger {
	if log == nil {
		return New[slog.Level](defaultLevel, defaultTimeFormat)
	}

	panic("logger already initialized")
}

func Logger() *slog.Logger {
	if log == nil {
		return NewDefaultLogger()
	}
	return log
}

func New[L ~string | slog.Level](level L, timeFormat string) *slog.Logger {
	once.Do(func() {
		var logLevel slog.Level

		if val, ok := any(level).(string); ok {
			logLevel = getLevelByString(val)
		} else {
			logLevel = any(level).(slog.Level)
		}

		log = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  logLevel == slog.LevelDebug,
			Level:      logLevel,
			TimeFormat: timeFormat,
		}))

		slog.SetDefault(log)
	})

	return log
}

func getLevelByString(s string) slog.Level {
	switch s {
	case levelDebug:
		return slog.LevelDebug
	case levelInfo:
		return slog.LevelInfo
	case levelWarn:
		return slog.LevelWarn
	case levelErr:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
