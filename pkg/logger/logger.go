package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	// Global logger instance
	Logger *slog.Logger
)

// Config holds logger configuration
type Config struct {
	LogDir      string
	Level       slog.Level
	Format      string // "json" or "text"
	ServiceName string
}

// Init initializes the file-based logger
func Init(cfg Config) error {
	var writer io.Writer = os.Stdout
	if cfg.LogDir != "" {
		// Create logs directory if it doesn't exist
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return err
		}

		// Create log file with timestamp
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFile := filepath.Join(cfg.LogDir, cfg.ServiceName+"_"+timestamp+".log")

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		// Multi-writer: both file and stdout
		writer = io.MultiWriter(file, os.Stdout)
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	// Add service name to all logs
	Logger = slog.New(handler)
	Logger = Logger.With("service", cfg.ServiceName)

	return nil
}

// InitDefault initializes logger with default settings
func InitDefault(serviceName string) error {
	return Init(Config{
		LogDir:      "logs",
		Level:       slog.LevelInfo,
		Format:      "text",
		ServiceName: serviceName,
	})
}

// LogStartup logs application startup
func LogStartup(version string, port string) {
	Logger.Info("application starting",
		"version", version,
		"port", port,
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// LogShutdown logs application shutdown
func LogShutdown(reason string) {
	Logger.Info("application shutting down",
		"reason", reason,
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// LogComponentStartup logs when a component starts
func LogComponentStartup(component string, details map[string]string) {
	args := []any{"component", component, "status", "starting", "timestamp", time.Now().Format(time.RFC3339)}
	for k, v := range details {
		args = append(args, k, v)
	}
	Logger.Info("component startup", args...)
}

// LogComponentShutdown logs when a component stops
func LogComponentShutdown(component string, reason string) {
	Logger.Info("component shutting down",
		"component", component,
		"reason", reason,
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// LogError logs an error with context
func LogError(operation string, err error, details map[string]string) {
	args := []any{"operation", operation, "error", err.Error(), "timestamp", time.Now().Format(time.RFC3339)}
	for k, v := range details {
		args = append(args, k, v)
	}
	Logger.Error("operation error", args...)
}

// LogInfo logs informational message
func LogInfo(message string, details map[string]string) {
	args := []any{"message", message, "timestamp", time.Now().Format(time.RFC3339)}
	for k, v := range details {
		args = append(args, k, v)
	}
	Logger.Info("info message", args...)
}
