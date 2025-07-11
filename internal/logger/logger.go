package logger

import (
	"log/slog"
	"os"

	"github.com/MatusOllah/slogcolor"
)

// Global logger instance.
var Logger *slog.Logger

// Setup initializes the logger with colored output using slogcolor.
func Setup() {
	// Initialize logger with colored output from slogcolor
	opts := &slogcolor.Options{
		Level:      slog.LevelInfo,
		TimeFormat: "2006-01-02 15:04:05", // Standard Go time format
	}

	// Use slogcolor's handler for colored output
	handler := slogcolor.NewHandler(os.Stdout, opts)
	Logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(Logger)
}

// This function ensures the Logger is properly initialized before use.
func GetLogger() *slog.Logger {
	// If Logger is nil, initialize it
	if Logger == nil {
		Setup()
	}

	return Logger
}
