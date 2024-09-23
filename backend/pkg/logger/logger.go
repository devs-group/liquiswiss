package logger

import (
	"liquiswiss/pkg/utils"

	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func NewZapLogger() *zap.SugaredLogger {
	var l *zap.Logger
	if utils.IsProduction() {
		l, _ = zap.NewProduction()
	} else {
		l, _ = zap.NewDevelopment()
	}
	Logger = l.Sugar()
	return Logger
}
