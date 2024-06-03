package logutil

import (
	lgs "github.com/sirupsen/logrus"
)


func init() {
	lgs.SetReportCaller(true)
	Formatter := new(lgs.JSONFormatter)
	Formatter.TimestampFormat = "2006-01-02 15:04:05"
	lgs.SetFormatter(Formatter)
}