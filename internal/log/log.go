package log

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	Info  func(msg string, args ...any)
	Debug func(msg string, args ...any)
	Error func(msg string, args ...any)
)

func Setup(level slog.Level) {
	slog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	Info = slog.Info
	Debug = slog.Debug
	Error = slog.Error
}

func Level(l string) slog.Level {
	switch l {
	case "info":
		return slog.LevelInfo.Level()
	case "error":
		return slog.LevelError.Level()
	case "debug":
		return slog.LevelDebug.Level()
	default:
		panic(fmt.Sprintf("unkown log level %s", l))
	}
}
