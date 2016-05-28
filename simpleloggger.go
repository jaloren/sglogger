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

type SimpleLogger struct {
	goLogger *golog.Logger
	loglevel int
	handler  []string
	logfile  string
	freeze   bool
}

func (l *SimpleLogger) TerminateLogFile() {
	if l.logfile == "" {
		return
	}
	function, file, line, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("FATAL", "Terminates log file", funcName, file, line, "]")
	l.goLogger.Println(logmsg)

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

func (l *SimpleLogger) SetHandlers(path string, overwrite bool) error {
	if l.freeze {
		return frozeErrMsg
	}
	inode, err := l.SetFileHandler(path, overwrite)
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

func (l *SimpleLogger) SetFileHandler(path string, overwrite bool) (*os.File, error) {
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
	l.goLogger.SetOutput(inode)
	l.logfile = path
	l.handler = []string{"file"}
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
	l.goLogger.Println(logmsg)

}

func (l *SimpleLogger) Error(msg string) {
	if l.loglevel < loglevels["ERROR"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("ERROR", msg, funcName, file, line, ",")
	l.goLogger.Println(logmsg)
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
	l.goLogger.Println(logmsg)
	return nil
}

func (l *SimpleLogger) Warning(msg string) {
	if l.loglevel < loglevels["WARNING"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("WARNING", msg, funcName, file, line, ",")
	l.goLogger.Println(logmsg)
}

func (l *SimpleLogger) Info(msg string) {
	if l.loglevel < loglevels["INFO"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("INFO", msg, funcName, file, line, ",")
	l.goLogger.Println(logmsg)
}

func (l *SimpleLogger) Debug(msg string) {
	if l.loglevel < loglevels["DEBUG"] {
		return
	}
	function, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("DEBUG", msg, funcName, file, line, ",")
	l.goLogger.Println(logmsg)
}
