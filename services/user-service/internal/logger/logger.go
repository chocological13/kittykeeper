package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// NewLogger returns a new logrus with a custom field for service/module name
func NewLogger(module string) *log.Entry {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	return log.WithField("module", module)
}
