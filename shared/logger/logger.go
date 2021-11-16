package logger

import (
	"fmt"
	"log"
	"os"

	logrus "github.com/sirupsen/logrus"
)

// Logger is a new instance of the logger
var Logger = logrus.New()

// Setup configures the logger.
func Setup() {
	log.SetOutput(os.Stdout)

	Logger.Out = os.Stdout
	Logger.Level = logrus.DebugLevel
}

// WithFunc creates a logger with pkgpath and function
func WithFunc(pkgPath string, function string) *logrus.Entry {
	return Logger.WithFields(
		logrus.Fields{
			"func": fmt.Sprintf("(%s).%s", pkgPath, function),
		})
}

// WithClassFunc creates a logger with pkgpath and function
func WithClassFunc(pkgPath, class, function string) *logrus.Entry {
	return Logger.WithFields(
		logrus.Fields{
			"func": fmt.Sprintf("(%s.%s).%s", pkgPath, class, function),
		})
}
