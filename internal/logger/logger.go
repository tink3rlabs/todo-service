package logger

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

// mapLogLevel maps a string log level from config to slog.Level
func mapLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to Info level
	}
}

func GetLogger() *slog.Logger {

	var handler slog.Handler

	// Fetch the log level and format from the config file
	levelStr := viper.GetString("logger.log_level")
	formatJSON := viper.GetBool("logger.format_json")

	logLevel := mapLogLevel(levelStr)

	// Choose the handler based on the format and log level from the config
	if formatJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	// Initialize the logger with the selected handler
	logger := slog.New(handler)

	return logger
}
