package hLog

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

/**
logFile
*/

type logFile struct {
	id           string
	_suffix      int
	_date        *time.Time
	rwMutex      *sync.RWMutex
	dir          string
	fileName     string
	logFile      *os.File
	lg           *log.Logger
	filesSize    int64
	maxFileSize  int64
	maxFileCount int32
}

func (this *logFile) nextSuffix() int {
	return int(this._suffix%int(this.maxFileCount) + 1)
}

func newLogFile(dir, fileName string, _suffix int, maxFileSize int64, maxFileCount int32) (lf *logFile) {
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	lf = &logFile{dir: dir, fileName: fileName, rwMutex: new(sync.RWMutex)}
	lf.logFile, _ = os.OpenFile(dir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	lf.lg = log.New(lf.logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	lf._suffix = _suffix
	lf.maxFileCount = maxFileCount
	lf.maxFileSize = maxFileSize
	lf.filesSize = getFileSize(dir + "/" + fileName)
	lf._date = &t
	return
}

func (this *logFile) rename(rolltype ROLLTYPE) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	this.close()
	nextFileName := ""
	switch rolltype {
	case DAILY:
		nextFileName = fmt.Sprint(this.dir, "/", this.fileName, ".", this._date.Format(_DATEFORMAT))
	case ROLLFILE:
		nextFileName = fmt.Sprint(this.dir, "/", this.fileName, ".", this.nextSuffix())
		this._suffix = this.nextSuffix()
	}
	if isExist(nextFileName) {
		os.Remove(nextFileName)
	}
	os.Rename(this.dir+"/"+this.fileName, nextFileName)
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	this._date = &t
	this.logFile, _ = os.OpenFile(this.dir+"/"+this.fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	this.lg = log.New(this.logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	this.filesSize = getFileSize(this.dir + "/" + this.fileName)
}

func (this *logFile) write(level string, val ...interface{}) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	this.lg.Output(SKIP+1, fmt.Sprintln(level, fmt.Sprint(val...)))
}

func (this *logFile) fwrite(format string, val ...interface{}) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	this.lg.Output(SKIP+1, fmt.Sprintf(format, val...))
}

func (this *logFile) addSize(size int64) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	atomic.AddInt64(&this.filesSize, size)
}

func (this *logFile) isOverSize() bool {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	return this.filesSize >= this.maxFileSize
}

func (this *logFile) close() {
	this.logFile.Close()
}
