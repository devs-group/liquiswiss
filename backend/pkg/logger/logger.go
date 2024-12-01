package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func NewZapLogger(isProduction bool) *zap.SugaredLogger {
	var l *zap.Logger
	if isProduction {
		l, _ = zap.NewProduction()
	} else {
		l, _ = zap.NewDevelopment()
	}
	Logger = l.Sugar()
	return Logger
}
