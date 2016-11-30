package main

import (
	"github.com/Sirupsen/logrus"
)

type logFields logrus.Fields

var logger *logrus.Logger

func init() {
	logger = logrus.New()
}
