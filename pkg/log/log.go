package log

import (
	"os"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func NewLogger(enableDebug bool) *logr.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetReportCaller(true)
	if enableDebug {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.ErrorLevel)
	}

	logger := logrusr.New(l)

	return &logger
}
