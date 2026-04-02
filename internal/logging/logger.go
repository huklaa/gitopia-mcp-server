package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Stdout is reserved for the MCP protocol, so all logs go to stderr.
	log.SetOutput(os.Stderr)

	// Use JSON formatter for structured logs
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false, // Set to true for development
	})

	// Default to Info level, can be configured later by the application
	log.SetLevel(logrus.InfoLevel)
}

// Config represents the logging configuration.
// This is a subset of the main config to avoid circular dependencies.
type Config struct {
	Level string `json:"level" yaml:"level"`
}

// ConfigureLogger sets up the logger based on the provided configuration.
func ConfigureLogger(cfg Config) {
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		log.WithError(err).Warnf("Invalid log level '%s', defaulting to 'info'", cfg.Level)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	if level == logrus.DebugLevel {
		log.SetReportCaller(true) // Include file and line number in debug mode
	}
}

// WithError creates a new log entry with the given error.
func WithError(err error) *logrus.Entry {
	return log.WithError(err)
}

// WithField creates a new log entry with the given key and value.
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields creates a new log entry with the given fields.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}
