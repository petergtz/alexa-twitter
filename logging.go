package main

import (
	"bytes"
	"log"
	"strings"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func createLoggerWith(logLevel string) *zap.Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zapLogLevelFrom(logLevel)
	loggerConfig.DisableStacktrace = true
	logger, e := loggerConfig.Build()
	if e != nil {
		log.Panic(e)
	}
	return logger
}

func zapLogLevelFrom(configLogLevel string) zap.AtomicLevel {
	switch strings.ToLower(configLogLevel) {
	case "", "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		log.Fatal("Invalid log level in config", "log-level", configLogLevel)
		return zap.NewAtomicLevelAt(-1)
	}
}

const (
	_stdLogDefaultDepth = 2
	_loggerWriterDepth  = 1
)

// Copied from go.uber.org/zap/global.go and changed to use Error instead of Info:
func NewStdLog(l *zap.Logger) *log.Logger {
	return log.New(&loggerWriter{l.WithOptions(
		zap.AddCallerSkip(_stdLogDefaultDepth + _loggerWriterDepth),
	)}, "" /* prefix */, 0 /* flags */)
}

type loggerWriter struct{ logger *zap.Logger }

func (l *loggerWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	l.logger.Error(string(p))
	return len(p), nil
}
