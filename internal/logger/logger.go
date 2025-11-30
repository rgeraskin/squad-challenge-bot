package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

var Log *log.Logger

// Init initializes the logger with the specified level
func Init(level string) {
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	switch strings.ToLower(level) {
	case "debug":
		Log.SetLevel(log.DebugLevel)
	case "info":
		Log.SetLevel(log.InfoLevel)
	case "warn":
		Log.SetLevel(log.WarnLevel)
	case "error":
		Log.SetLevel(log.ErrorLevel)
	default:
		Log.SetLevel(log.InfoLevel)
	}
}

// Helper functions for convenience
func Debug(msg string, keyvals ...interface{}) {
	Log.Debug(msg, keyvals...)
}

func Info(msg string, keyvals ...interface{}) {
	Log.Info(msg, keyvals...)
}

func Warn(msg string, keyvals ...interface{}) {
	Log.Warn(msg, keyvals...)
}

func Error(msg string, keyvals ...interface{}) {
	Log.Error(msg, keyvals...)
}

func Fatal(msg string, keyvals ...interface{}) {
	Log.Fatal(msg, keyvals...)
}

// WithPrefix returns a logger with a prefix
func WithPrefix(prefix string) *log.Logger {
	return Log.WithPrefix(prefix)
}
