package hLog

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"
)

/**
logConsole
*/

type logConsole struct {
	id            string
	logLevel      LEVEL
	format        string
	rolltype      ROLLTYPE
	isConsole     bool
	mutex         *sync.Mutex
	d, i, w, e, f string
	maxFileSize   int64
	maxFileCount  int32
}

type logFileList struct {
	fbs     map[string]*logFile
	rwMutex *sync.RWMutex
}

func (this *logFileList) add(dir, fileName string, _suffix int, maxFileSize int64, maxFileCount int32) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	id := md5str(fmt.Sprint(dir, fileName))
	if _, ok := this.fbs[id]; !ok {
		this.fbs[id] = newLogFile(dir, fileName, _suffix, maxFileSize, maxFileCount)
	}
}

func (this *logFileList) get(id string) *logFile {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	return this.fbs[id]
}

var lfl = &logFileList{fbs: make(map[string]*logFile, 0), rwMutex: new(sync.RWMutex)}

func (this *logConsole) setConsole(isConsole bool) {
	this.isConsole = isConsole
}

func (this *logConsole) setLevelFile(level LEVEL, dir string, fileName string) {
	key := md5str(fmt.Sprint(dir, fileName))
	switch level {
	case DEBUG:
		this.d = key
	case INFO:
		this.i = key
	case WARN:
		this.w = key
	case ERROR:
		this.e = key
	case FATAL:
		this.f = key
	default:
		return
	}
	var _suffix = 0
	if this.maxFileCount < 1<<31-1 {
		for i := 1; i < int(this.maxFileCount); i++ {
			if isExist(dir + "/" + fileName + "." + strconv.Itoa(i)) {
				_suffix = i
			} else {
				break
			}
		}
	}
	lfl.add(dir, fileName, _suffix, this.maxFileSize, this.maxFileCount)
}

func (this *logConsole) setLevel(level LEVEL) {
	this.logLevel = level
}

func (this *logConsole) setFormat(logFormat string) {
	this.format = logFormat
}

func (this *logConsole) setRollingFile(dir, fileName string, maxFileSize int64, maxFileCount int32) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if maxFileCount > 0 {
		this.maxFileCount = maxFileCount
	} else {
		this.maxFileCount = 1<<31 - 1
	}
	this.maxFileSize = maxFileSize
	this.rolltype = ROLLFILE
	mkdirLog(dir)
	var _suffix = 0
	for i := 1; i < int(maxFileSize); i++ {
		if isExist(dir + "/" + fileName + "." + strconv.Itoa(i)) {
			_suffix = i
		} else {
			break
		}
	}
	this.id = md5str(fmt.Sprint(dir, fileName))
	lfl.add(dir, fileName, _suffix, this.maxFileSize, this.maxFileCount)
}

func (this *logConsole) setRollingDaily(dir, fileName string) {
	this.rolltype = DAILY
	this.id = md5str(fmt.Sprint(dir, fileName))
	lfl.add(dir, fileName, 0, 0, 0)
}

func (this *logConsole) console(level string, val ...interface{}) {
	s := fmt.Sprint(val...)
	if this.isConsole {
		_, file, line, _ := runtime.Caller(SKIP)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		if this.format == "" {
			log.Println(file, strconv.Itoa(line), level, s)
		} else {
			vs := make([]interface{}, 0)
			vs = append(vs, file)
			vs = append(vs, strconv.Itoa(line))
			vs = append(vs, level)
			for _, vv := range val {
				vs = append(vs, vv)
			}
			log.Printf(fmt.Sprint("%s %s ", this.format, "\n"), vs...)
		}
	}
}

/**
日志输出的主要逻辑
*/
func (this *logConsole) log(level string, val ...interface{}) {
	defer catchError()
	s := fmt.Sprint(val...)
	length := len([]byte(s))
	var lg *logFile = lfl.get(this.id)
	var _level = ALL
	switch level {
	case "debug":
		if this.d != "" {
			lg = lfl.get(this.d)
		}
		_level = DEBUG
	case "info":
		if this.i != "" {
			lg = lfl.get(this.i)
		}
		_level = INFO
	case "warn":
		if this.w != "" {
			lg = lfl.get(this.w)
		}
		_level = WARN
	case "error":
		if this.e != "" {
			lg = lfl.get(this.e)
		}
		_level = ERROR
	case "fatal":
		if this.f != "" {
			lg = lfl.get(this.f)
		}
		_level = FATAL
	}
	log_level := fmt.Sprintf("[%s]", level)
	if lg != nil {
		this.fileCheck(lg)
		lg.addSize(int64(length))
		if this.logLevel <= _level {
			if lg != nil {
				if this.format == "" {
					lg.write(log_level, s)
				} else {
					lg.fwrite(this.format, val...)
				}
			}
			this.console(log_level, val...)
		}
	} else {
		this.console(log_level, val...)
	}
}

/**
文件检查 判断是否需要要更新新的log日志文件
*/
func (this *logConsole) fileCheck(fb *logFile) {
	defer catchError()
	if this.isMustRename(fb) {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if this.isMustRename(fb) {
			fb.rename(this.rolltype)
		}
	}
}

func (this *logConsole) debug(v ...interface{}) {
	this.log("debug", v...)
}
func (this *logConsole) info(v ...interface{}) {
	this.log("info", v...)
}
func (this *logConsole) warn(v ...interface{}) {
	this.log("warn", v...)
}
func (this *logConsole) error(v ...interface{}) {
	this.log("error", v...)
}
func (this *logConsole) fatal(v ...interface{}) {
	this.log("fatal", v...)
}

/**
判断文件是否需要重命名
*/
func (this *logConsole) isMustRename(fb *logFile) bool {
	switch this.rolltype {
	case DAILY:
		t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
		if t.After(*fb._date) {
			return true
		}
	case ROLLFILE:
		return fb.isOverSize()
	}
	return false
}

/**
new
*/
func newLogConsole() *logConsole {
	lc := &logConsole{}
	lc.mutex = new(sync.Mutex)
	lc.setConsole(true)
	return lc
}
