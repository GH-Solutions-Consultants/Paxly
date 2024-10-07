// core/logger.go
package core

import (
	"os"

	"github.com/sirupsen/logrus"
)

// InitializeLogger sets up the global logger.
func InitializeLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.WarnLevel) // Default level
}

// LogFatal logs the error and exits the application.
func LogFatal(err error) {
	logrus.Fatal(err)
}