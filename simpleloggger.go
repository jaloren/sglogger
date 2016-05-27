package sglogger

import (
	"errors"
	"fmt"
	"io"
	golog "log"
	"os"
	"runtime"
	"strconv"
)

type SimpleLog struct {
	simpleLogger *golog.Logger
	loglevel     int
	handler      []string
	logfile      string
	freeze       bool
}

func (l *SimpleLog) GetGlobalLogAttr() map[string]string {
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

func (l *SimpleLog) Freeze(state bool) error {
	if l.freeze {
		return frozeErrMsg
	}
	l.freeze = true
	return nil
}

func (l *SimpleLog) SetHandlers(path string, overwrite bool) error {
	if l.freeze {
		return frozeErrMsg
	}
	inode, err := l.SetFileHandler(path, overwrite)
	if err != nil {
		return err
	}
	mw := io.MultiWriter(inode, os.Stderr)
	l.handler = []string{"file", "stderr"}
	l.simpleLogger.SetOutput(mw)
	return nil
}

func (l *SimpleLog) SetConsolehandler() error {
	if l.freeze {
		return frozeErrMsg
	}
	l.simpleLogger.SetOutput(os.Stderr)
	l.logfile = ""
	l.handler = []string{"stderr"}
	return nil
}

func (l *SimpleLog) SetCustomHandler(writer io.Writer) error {
	if l.freeze {
		return frozeErrMsg
	}
	l.simpleLogger.SetOutput(writer)
	return nil
}

func (l *SimpleLog) SetFileHandler(path string, overwrite bool) (*os.File, error) {
	var inode *os.File
	if l.freeze {
		return inode, frozeErrMsg
	}
	if _, err := os.Stat(path); err == nil && !overwrite {
		return inode, fmt.Errorf("Cannot log to file %s because it already exists\n", path)
	}
	inode, err := os.Create(path)
	if err != nil {
		return inode, fmt.Errorf("Failed to create log file. %v", err)
	}
	l.simpleLogger.SetOutput(inode)
	l.logfile = path
	l.handler = []string{"file"}
	return inode, err

}

func (l *SimpleLog) SetLogLevel(level string) error {
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

func (l *SimpleLog) Fatal(msg string, crash bool) {
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("FATAL", msg, funcName, file, line)
	if crash {
		l.simpleLogger.Panic(logmsg)
	}
	l.simpleLogger.Println(logmsg)

}

func (l *SimpleLog) Error(msg string) {
	if l.loglevel < loglevels["ERROR"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("ERROR", msg, funcName, file, line)
	l.simpleLogger.Println(logmsg)
}

func (l *SimpleLog) Exception(msg string, err error) error {
	if l.loglevel < loglevels["ERROR"] {
		return nil
	}
	if err == nil {
		return errors.New("Exception is not logged because error is nil.")
	}
	errMsg := fmt.Sprintf("%s, errorString: %v", msg, err)
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("ERROR", errMsg, funcName, file, line)
	l.simpleLogger.Println(logmsg)
	return nil
}

func (l *SimpleLog) Warning(msg string) {
	if l.loglevel < loglevels["WARNING"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("WARNING", msg, funcName, file, line)
	l.simpleLogger.Println(logmsg)
}

func (l *SimpleLog) Info(msg string) {
	if l.loglevel < loglevels["INFO"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("INFO", msg, funcName, file, line)
	l.simpleLogger.Println(logmsg)
}

func (l *SimpleLog) Debug(msg string) {
	if l.loglevel < loglevels["DEBUG"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("DEBUG", msg, funcName, file, line)
	l.simpleLogger.Println(logmsg)
}
