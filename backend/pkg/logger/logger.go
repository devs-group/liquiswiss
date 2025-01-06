package logger

import (
	"go.uber.org/zap"
	"log"
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

type StdLogger struct{}

func (StdLogger) Fatalf(format string, v ...interface{}) { log.Fatalf(format, v...) }
func (StdLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }
