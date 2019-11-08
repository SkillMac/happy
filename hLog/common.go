package hLog

import (
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
	"os"
)

/*
type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}
*/

// 0 - 5  C-D

var log = logging.MustGetLogger("happyyLog")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfile} %{longfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func InitLogger(fileName string, fileSize int, fileMax int, logLv int, LogConsolePrint bool) {

	backendFile := logging.NewLogBackend(&lumberjack.Logger{
		Filename: "./logs/" + fileName + ".log",
		MaxSize:  fileSize, // megabytes
		Compress: true,     // disabled by default

		MaxAge:     7,
		MaxBackups: fileMax,
		LocalTime:  true,
	}, "", 0)

	backendFileErr := logging.NewLogBackend(&lumberjack.Logger{
		Filename: "./logs/" + fileName + "-error.log",
		MaxSize:  fileSize, // megabytes
		Compress: false,    // disabled by default

		MaxAge:     7,
		MaxBackups: fileMax,
		LocalTime:  true,
	}, "", 0)

	backendFileFormatter := logging.NewBackendFormatter(backendFile, format)
	backendFileErrFormatter := logging.NewBackendFormatter(backendFileErr, format)

	backendFileLeveled := logging.AddModuleLevel(backendFileFormatter)
	backendFileLeveled.SetLevel(logging.Level(logLv), "")
	backendFileErrLeveled := logging.AddModuleLevel(backendFileErrFormatter)
	backendFileErrLeveled.SetLevel(logging.ERROR, "")

	var backend2 *logging.LogBackend = nil
	var backend2Formatter logging.Backend

	if LogConsolePrint {
		backend2 = logging.NewLogBackend(os.Stdout, "", 0)
		backend2Formatter = logging.NewBackendFormatter(backend2, format)
	}

	if backend2 == nil {
		logging.SetBackend(backendFileLeveled, backendFileErrLeveled)
	} else {
		logging.SetBackend(backendFileLeveled, backendFileErrLeveled, backend2Formatter)
	}
}

var Debug = log.Debug
var Debugf = log.Debugf
var Info = log.Info
var Infof = log.Infof
var Notice = log.Notice
var Noticef = log.Noticef
var Warn = log.Warning
var Warnf = log.Warningf
var Error = log.Error
var Errorf = log.Errorf

var Critical = log.Error
var Criticalf = log.Criticalf

var Fatal = log.Fatal
var Fatalf = log.Fatalf
var Panic = log.Panic
var Panicf = log.Panicf
