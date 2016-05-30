package sglogger

import (
	"errors"
	"fmt"
	"io"
	golog "log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type SimpleLogger struct {
	goLogger   *golog.Logger
	loglevel   int
	handler    []string
	logfile    string
	logfd      *os.File
	logdir     string
	filePrefix string
	freeze     bool
	quit       chan bool
	rotateLock sync.RWMutex
}

func (l *SimpleLogger) SyncLogFile() error {
	if l.logfile == "" {
		return fmt.Errorf("Unable flush log contents in mememory to a file on disk because the file does not exist.")
	}
	l.rotateLock.Lock()
	defer l.rotateLock.Unlock()
	l.logfd.Sync()
	return nil
}

func (l *SimpleLogger) lockAndLog(msg string) {
	l.rotateLock.RLock()
	defer l.rotateLock.RUnlock()
	l.goLogger.Println(msg)
}

func (l *SimpleLogger) GetLoggerAttrs() map[string]string {
	var lvlname string
	for k, v := range loglevels {
		if v == l.loglevel {
			lvlname = k
		}
	}
	attrs := map[string]string{
		"loglevel": lvlname,
		"handler":  fmt.Sprintf("%v", l.handler),
		"logfile":  l.logfile,
		"freeze":   strconv.FormatBool(l.freeze),
	}
	return attrs
}

func (l *SimpleLogger) Freeze(state bool) error {
	if l.freeze {
		return frozeErrMsg
	}
	l.freeze = true
	return nil
}

func (l *SimpleLogger) SetHandlers(logdir string, filePrefix string) error {
	if l.freeze {
		return frozeErrMsg
	}
	inode, err := l.SetFileHandler(logdir, filePrefix)
	if err != nil {
		return err
	}
	mw := io.MultiWriter(inode, os.Stderr)
	l.handler = []string{"file", "stderr"}
	l.goLogger.SetOutput(mw)
	return nil
}

func (l *SimpleLogger) SetConsoleHandler() error {
	if l.freeze {
		return frozeErrMsg
	}
	l.goLogger.SetOutput(os.Stderr)
	l.logfile = ""
	l.handler = []string{"stderr"}
	return nil
}

func (l *SimpleLogger) SetFileHandler(logdir string, filePrefix string) (*os.File, error) {
	var inode *os.File
	if l.freeze {
		return inode, frozeErrMsg
	}
	inode, err := l.setInternalFileHandler(logdir, filePrefix)
	l.handler = []string{"file"}
	return inode, err
}

func (l *SimpleLogger) setInternalFileHandler(logdir string, filePrefix string) (*os.File, error) {
	var inode *os.File
	re := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !re.MatchString(filePrefix) {
		return inode, fmt.Errorf("File prefix %v in log file path contains one or more invalid characters. Valid characters are alphanumeric.", filePrefix)
	}
	basename := fmt.Sprintf("%v_%v.log", filePrefix, time.Now().UTC().Format("2006-01-02T150405.000"))
	path := path.Join(logdir, basename)
	if !filepath.IsAbs(path) {
		return inode, fmt.Errorf("Log file path %v is not absolute.", path)
	}
	os.MkdirAll(logdir, 0700)
	err := os.Chmod(logdir, 0700)
	if err != nil {
		return inode, fmt.Errorf("Unable to change permissions on log directory %v to 700. Error: %v", logdir, err.Error())
	}
	inode, err = os.Create(path)
	if err != nil {
		return inode, fmt.Errorf("Failed to create log file. %v", err)
	}
	err = inode.Chmod(0600)
	if err != nil {
		return inode, fmt.Errorf("Failed to set file permissions on log file %v to 600. Error: %v", path, err.Error())
	}
	l.goLogger.SetOutput(inode)
	l.logfile = path
	l.logdir = logdir
	l.logfd = inode
	l.filePrefix = filePrefix
	inode.WriteString("[")
	return inode, err

}

func (l *SimpleLogger) SetLogLevel(level string) error {
	if l.freeze {
		return frozeErrMsg
	}
	if newLevel, ok := loglevels[level]; ok {
		l.loglevel = newLevel
		return nil
	}
	errmsg := fmt.Sprintf("Log level %s is invalid. Valid log levels: %v", level, GetLogLevels())
	return fmt.Errorf(errmsg)
}

func (l *SimpleLogger) Fatal(msg string, crash bool) {
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("FATAL", msg, funcName, file, line, ",")
	if crash {
		l.goLogger.Panic(logmsg)
	}
	l.lockAndLog(logmsg)

}

func (l *SimpleLogger) Error(msg string) {
	if l.loglevel < loglevels["ERROR"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("ERROR", msg, funcName, file, line, ",")
	l.lockAndLog(logmsg)
}

func (l *SimpleLogger) Exception(msg string, err error) error {
	if l.loglevel < loglevels["ERROR"] {
		return nil
	}
	if err == nil {
		return errors.New("Exception is not logged because error is nil.")
	}
	errMsg := fmt.Sprintf("%s, errorString: %v", msg, err)
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("ERROR", errMsg, funcName, file, line, ",")
	l.lockAndLog(logmsg)
	return nil
}

func (l *SimpleLogger) Warning(msg string) {
	if l.loglevel < loglevels["WARNING"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("WARNING", msg, funcName, file, line, ",")
	l.lockAndLog(logmsg)
}

func (l *SimpleLogger) Info(msg string) {
	if l.loglevel < loglevels["INFO"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("INFO", msg, funcName, file, line, ",")
	l.lockAndLog(logmsg)
}

func (l *SimpleLogger) Debug(msg string) {
	if l.loglevel < loglevels["DEBUG"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("DEBUG", msg, funcName, file, line, ",")
	l.lockAndLog(logmsg)
}
