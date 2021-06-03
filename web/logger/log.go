// Package logger is a wrapper around a Logrus logging instance.
// It implements a custom formatter for logging messages.
package logger

import (
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"os"
)

// Logger - an exported Logrus instance.
var Logger *log.Logger

func init() {
	Logger = &log.Logger{
		Out:   os.Stderr,
		Level: log.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05.55",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}
}
