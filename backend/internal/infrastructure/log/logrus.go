// Package log wraps logrus to provide a project-wide structured logger
// with consistent formatting and level configuration.
package log

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Log is the singleton logger used through the L() accessor.
var Log = logrus.New()

// Init configures the logger level and output format. Format may be "json"
// (default for production) or anything else (logrus text formatter).
func Init(level, format string) {
	Log.SetOutput(os.Stdout)
	switch strings.ToLower(format) {
	case "json":
		Log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000Z07:00"})
	default:
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	Log.SetLevel(parseLevel(level))
}

// L returns the global logger instance.
func L() *logrus.Logger {
	return Log
}

func parseLevel(s string) logrus.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "info":
		fallthrough
	default:
		return logrus.InfoLevel
	}
}
