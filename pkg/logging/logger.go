package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new logger with the specified log level
func NewLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}
