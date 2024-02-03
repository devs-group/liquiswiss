package logger

import (
	"liquiswiss/pkg/utils"

	"go.uber.org/zap"
)

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

type zapLogger struct {
	*zap.SugaredLogger
}

func (zl *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	zl.SugaredLogger.Infow(msg, keysAndValues...)
}

func (zl *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	zl.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (zl *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	zl.SugaredLogger.Debugw(msg, keysAndValues...)
}

func NewZapLogger() Logger {
	var logger *zap.Logger
	if utils.IsProduction() {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	return &zapLogger{logger.Sugar()}
}
