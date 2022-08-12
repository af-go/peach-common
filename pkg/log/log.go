package log

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

/*
func NewLogger(enableDebug bool) *logr.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetReportCaller(false)
	if enableDebug {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.ErrorLevel)
	}
	logger := logrusr.New(l)

	return &logger
}*/

func NewLogger(enableDebug bool) *logr.Logger {
	var zc zap.Config
	if enableDebug {
		zc = zap.NewDevelopmentConfig()
	} else {
		zc = zap.NewProductionConfig()
	}
	z, err := zc.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger (%v)?", err))
	}
	logger := zapr.NewLogger(z)
	return &logger
}
