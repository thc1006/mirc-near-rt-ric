package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// LogrusLogger wraps logrus.Logger to implement our Logger interface
type LogrusLogger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	
	return &LogrusLogger{Logger: logger}
}

// NewLoggerWithLevel creates a new logger with specified level
func NewLoggerWithLevel(level string) Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	
	return &LogrusLogger{Logger: logger}
}

// WithField creates a new logger with the specified field
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{Logger: l.Logger.WithField(key, value).Logger}
}

// WithFields creates a new logger with the specified fields
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{Logger: l.Logger.WithFields(fields).Logger}
}