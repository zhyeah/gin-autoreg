package log

import "github.com/sirupsen/logrus"

var Logger = logrus.New()

func init() {
	// Logger.SetLevel(logrus.DebugLevel)
	Logger.SetLevel(logrus.ErrorLevel)
	Logger.SetFormatter(&logrus.TextFormatter{})
}
