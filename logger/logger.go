// logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger(outputToTerminal bool) {
	file, err := os.OpenFile("matterfeed.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file: %v", err)
	}

	if outputToTerminal {
		Logger.SetOutput(io.MultiWriter(file, os.Stdout))
	} else {
		Logger.SetOutput(file)
	}

	Logger.SetLevel(logrus.InfoLevel)

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func LogError(err error, context string) {
	message := fmt.Sprintf("An error occurred: %s, context: %s", err, context)
	Logger.WithFields(logrus.Fields{
		"context": context,
		"error":   err,
	}).Error(message)
}

func LogInfo(message string) {
	Logger.Info(message)
}

func LogAndReturnError(err error, context string) error {
	loggerError := fmt.Errorf("%s: %w", context, err)
	LogError(loggerError, context)
	return loggerError
}
