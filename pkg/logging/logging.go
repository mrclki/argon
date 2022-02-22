package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Sys defines the field denoting the system when creating a new logger
	Sys = "sys"

	// Subsys defines the field denoting the subsystem when creating a new logger
	Subsys = "subsys"
)

// Logger is the default logger
var Logger = defaultLogger()

func defaultLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	}
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

// SetLogLevel sets the log loggel on Logger.
func SetLogLevel(level logrus.Level) {
	Logger.SetLevel(level)
}
