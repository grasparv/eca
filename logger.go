package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

func makeLogger() *slog.Logger {
	logger := slog.New(
		console.NewHandler(os.Stdout,
			&console.HandlerOptions{
				Level:      slog.LevelInfo,
				TimeFormat: time.TimeOnly,
			},
		),
	)

	return logger
}
