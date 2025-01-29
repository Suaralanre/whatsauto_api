package utils

import (
	"log/slog"
	"os"
)

func CustomLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}
