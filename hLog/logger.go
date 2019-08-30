package hLog

/**
logger
*/

type logger struct {
	lc *logConsole
}

func (this *logger) SetConsole(isConsole bool) {
	this.lc.setConsole(isConsole)
}

func (this *logger) SetLevel(_level LEVEL) {
	this.lc.setLevel(_level)
}

func (this *logger) SetFormat(logFormat string) {
	this.lc.setFormat(logFormat)
}

func (this *logger) SetRollingFile(dir, fileName string, maxFileSize int64, maxFileCount int32) {
	this.lc.setRollingFile(dir, fileName, maxFileSize, maxFileCount)
}

func (this *logger) SetRollingDaily(dir, fileName string) {
	this.lc.setRollingDaily(dir, fileName)
}

func (this *logger) Debug(v ...interface{}) {
	this.lc.debug(v...)
}
func (this *logger) Info(v ...interface{}) {
	this.lc.info(v...)
}
func (this *logger) Warn(v ...interface{}) {
	this.lc.warn(v...)
}
func (this *logger) Error(v ...interface{}) {
	this.lc.error(v...)
}
func (this *logger) Fatal(v ...interface{}) {
	this.lc.fatal(v...)
}

func (this *logger) SetLevelFile(level LEVEL, dir, fileName string) {
	this.lc.setLevelFile(level, dir, fileName)
}

func newLogger() (l *logger) {
	l = new(logger)
	l.lc = newLogConsole()
	return
}
