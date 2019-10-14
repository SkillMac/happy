package hLog

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"runtime/debug"
)

type LEVEL int32
type ROLLTYPE int

const (
	_        = iota
	KB int64 = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

const _DATEFORMAT = "2006-01-02"
const _NEWFILEFORMAT = "%s-%s.log"

var SKIP = 4

const (
	DAILY ROLLTYPE = iota
	ROLLFILE
)

var CURRENT_LOG_MODE = DAILY

func md5str(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func getFileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func mkdirLog(dir string) (e error) {
	a := isExist(dir)
	if !a {
		if err := os.MkdirAll(dir, 0666); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func catchError() {
	if err := recover(); err != nil {
		fmt.Println(string((debug.Stack())))
	}
}

/**
public
*/
var defaultLog *logConsole = newLogConsole()

func SetConsole(isConsole bool) {
	defaultLog.setConsole(isConsole)
}

func SetLevel(_level LEVEL) {
	defaultLog.setLevel(_level)
}

func SetFormat(logFormat string) {
	defaultLog.setFormat(logFormat)
}

func SetRollingFile(dir, fileName string, maxFileSize int64, maxFileCount int32) {
	defaultLog.setRollingFile(dir, fileName, maxFileSize, maxFileCount)
}

func SetRollingDaily(fileDir, fileName string) {
	defaultLog.setRollingDaily(fileDir, fileName)
}

func Debug(v ...interface{}) {
	defaultLog.debug(v...)
}
func Info(v ...interface{}) {
	defaultLog.info(v...)
}
func Warn(v ...interface{}) {
	defaultLog.warn(v...)
}
func Error(v ...interface{}) {
	defaultLog.error(v...)
}
func Fatal(v ...interface{}) {
	defaultLog.fatal(v...)
}

func SetLevelFile(level LEVEL, dir, fileName string) {
	defaultLog.setLevelFile(level, dir, fileName)
}
