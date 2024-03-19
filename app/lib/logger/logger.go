package logger

import (
	"log/slog"
	"os"

	"github.com/eugene-static/Level0/app/lib/config"
)

func New(cfg *config.Logger) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	}
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	var l *slog.Logger
	switch cfg.Handler {
	case "text":
		l = slog.New(textHandler)
	case "json":
		l = slog.New(jsonHandler)
	}
	return l
}

func NewEmpty() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdin, nil))
}
