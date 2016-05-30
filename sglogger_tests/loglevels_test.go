package sglogger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/jaloren/sglogger"
)

const fatalLogMsg = "This is a fatal error message"
const errorLogMsg = "This is an error message"
const warningLogMsg = "This is a warning message"
const infoLogMsg = "This is a info message"
const debugLogMsg = "This is a debug message"

var GlobalLogger = sglogger.GlobalLogger

var (
	tmpdir = "/tmp/gotests"
)

func getLogFile(t *testing.T, rotate bool) []byte {
	logfn := getLogFileName()
	if rotate {
		err := GlobalLogger.Rotate()
		if err != nil {
			t.Fatalf("Failed to rotate logs. Error: %v", err.Error())
		}
	}
	GlobalLogger.SyncLogFile()
	tempfile, err := ioutil.ReadFile(logfn)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read file. Error: %v", err)
		t.Fatalf(errMsg)
	}
	return tempfile
}

func convertFromJson(t *testing.T, raw []byte) (m []sglogger.LogMsg, e error) {
	var logExample []sglogger.LogMsg
	if err := json.Unmarshal(raw, &logExample); err != nil {
		errMsg := fmt.Errorf("Failed to unmarshal json data from file %v into a struct. Error: %v", getLogFileName(), err)
		return logExample, errMsg
	}
	return logExample, nil
}

func genLogFails(t *testing.T, msg string, level string) {
	getLogFile(t, true)
	switch {
	case level == "FATAL":
		GlobalLogger.Fatal(msg, false)
	case level == "ERROR":
		GlobalLogger.Error(msg)
	case level == "WARNING":
		GlobalLogger.Warning(msg)
	case level == "INFO":
		GlobalLogger.Info(msg)
	case level == "DEBUG":
		GlobalLogger.Debug(msg)
	default:
		t.Errorf("Log level %v is not valid.", level)
	}
	tempfile := getLogFile(t, false)
	logdata, err := convertFromJson(t, tempfile)
	if err == nil {
		t.Errorf("A log was generated successfully at log level %v.\nThe log level is %v, which means the level should not have been high enough to have generated this log.\nLog data is: %v", logdata, getLogLevel(), logdata)
	}

}

func genLogSuccessfully(t *testing.T, msg string, level string) {
	switch {
	case level == "FATAL":
		GlobalLogger.Fatal(msg, false)
	case level == "ERROR":
		GlobalLogger.Error(msg)
	case level == "WARNING":
		GlobalLogger.Warning(msg)
	case level == "INFO":
		GlobalLogger.Info(msg)
	case level == "DEBUG":
		GlobalLogger.Debug(msg)
	default:
		t.Fatalf("Log level %v is not valid.", level)
	}
	tempfile := getLogFile(t, true)
	logdata, err := convertFromJson(t, tempfile)
	if err != nil {
		t.Fatalf("Failed to successfully generate log at level %v, Error: %v", level, err)
	}
	logevent := logdata[0]
	if logevent.Msg != msg {
		t.Fatalf("Log message was %v, but it should have been %v", logevent.Msg, msg)
	}
	function, file, _, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(function).Name()
	if funcName != logevent.Function {
		t.Fatalf("In log, function name is %v but it should have been %v", logevent.Function, funcName)
	}
	filePart := strings.Split(file, "src/")[1]
	if filePart != logevent.File {
		t.Fatalf("In log, file is %v but it should have been %v", logevent.File, filePart)
	}

}

func getLogFileName() string {
	attrs := GlobalLogger.GetLoggerAttrs()
	return attrs["logfile"]
}

func getLogLevel() string {
	attrs := GlobalLogger.GetLoggerAttrs()
	return attrs["loglevel"]
}

func successfullySetLogLevel(t *testing.T, loglevel string) {
	err := GlobalLogger.SetLogLevel(loglevel)
	if err != nil {
		t.Fatalf("Failed to set log level. Error: %v", err)
	}
}

func TestLevels(t *testing.T) {
	defer os.RemoveAll(tmpdir)
	_, err := GlobalLogger.SetFileHandler(tmpdir, "loglevels")
	if err != nil {
		t.Fatalf("Failed to create log file %v. Error: %v", getLogFileName(), err.Error())
	}
	loglevel_names := sglogger.GetLogLevels()
	for _, levelname := range loglevel_names {
		successfullySetLogLevel(t, levelname)
		switch {
		case levelname == "FATAL":
			genLogSuccessfully(t, fatalLogMsg, "FATAL")
			genLogFails(t, errorLogMsg, "ERROR")
			genLogFails(t, warningLogMsg, "WARNING")
			genLogFails(t, infoLogMsg, "INFO")
			genLogFails(t, debugLogMsg, "DEBUG")
		case levelname == "ERROR":
			genLogSuccessfully(t, errorLogMsg, "FATAL")
			genLogSuccessfully(t, errorLogMsg, levelname)
			genLogFails(t, warningLogMsg, "WARNING")
			genLogFails(t, infoLogMsg, "INFO")
			genLogFails(t, debugLogMsg, "DEBUG")
		case levelname == "WARNING":
			genLogSuccessfully(t, errorLogMsg, "FATAL")
			genLogSuccessfully(t, errorLogMsg, "ERROR")
			genLogSuccessfully(t, warningLogMsg, levelname)
			genLogFails(t, infoLogMsg, "INFO")
			genLogFails(t, debugLogMsg, "DEBUG")
		case levelname == "INFO":
			genLogSuccessfully(t, errorLogMsg, "FATAL")
			genLogSuccessfully(t, errorLogMsg, "ERROR")
			genLogSuccessfully(t, warningLogMsg, "WARNING")
			genLogSuccessfully(t, infoLogMsg, levelname)
			genLogFails(t, debugLogMsg, "DEBUG")
		case levelname == "DEBUG":
			genLogSuccessfully(t, errorLogMsg, "FATAL")
			genLogSuccessfully(t, errorLogMsg, "ERROR")
			genLogSuccessfully(t, warningLogMsg, "WARNING")
			genLogSuccessfully(t, infoLogMsg, "INFO")
			genLogSuccessfully(t, debugLogMsg, "DEBUG")

		default:
			t.Fatalf("Log level %v is not valid.", levelname)

		}

	}
}
