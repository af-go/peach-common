package Logger

import (
	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func NewLogger(enableDebug bool) *logr.Logger {

	logrusLog := logrus.New()
	logger := logrusr.New(logrusLog)

	return &logger
}
