package services

import (
	"context"
	log "github.com/sirupsen/logrus"
)

// LoggerContextKey is the context key for storing logger instances
const LoggerContextKey = "logger"

// GetLogger returns the logger from context or creates a new one
func GetLogger(ctx context.Context) *log.Entry {
	logger, ok := ctx.Value(LoggerContextKey).(*log.Entry)
	if !ok {
		return log.NewEntry(log.StandardLogger())
	}
	return logger
}
