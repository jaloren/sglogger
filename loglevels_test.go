package sglogger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

const errorLogMsg = "This is an error message"
const warningLogMsg = "This is a warning message"
const infoLogMsg = "This is a info message"
const debugLogMsg = "This is a debug message"

var (
	tmpdir = "/tmp/gotests"
)

func createTempFile(t *testing.T) (filename string) {
	os.MkdirAll(tmpdir, 755)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create temporary file. Error: %v", err)
		t.Fatalf(errMsg)
	}
	_, err = GlobalLog.SetFileHandler(tmpfile.Name(), true)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to set file handler. Error: %v", err)
		t.Fatalf(errMsg)
	}
	return tmpfile.Name()
}

func getTempFile(t *testing.T, filename string) []byte {
	tempfile, err := ioutil.ReadFile(filename)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read file. Error: %v", err)
		t.Fatalf(errMsg)
	}
	return tempfile
}

func convertFromJson(t *testing.T, raw []byte, filename string) (logMsg, error) {
	var logExample logMsg
	if err := json.Unmarshal(raw, &logExample); err != nil {
		errMsg := fmt.Errorf("Failed to unmarshal json data from file %v into a struct. Error: %v", filename, err)
		return logExample, errMsg
	}
	return logExample, nil
}

func genLogFails(t *testing.T, msg string, level string) {
	tempfilename := createTempFile(t)
	switch {
	case level == "ERROR":
		GlobalLog.Error(msg)
	case level == "WARNING":
		GlobalLog.Warning(msg)
	case level == "INFO":
		GlobalLog.Info(msg)
	case level == "DEBUG":
		GlobalLog.Debug(msg)
	default:
		t.Errorf("Log level %v is not valid.", level)
	}
	tempfile := getTempFile(t, tempfilename)
	logdata, err := convertFromJson(t, tempfile, tempfilename)
	if err == nil {
		t.Errorf("A log was generated successfully at log level %v. The log level is %v, which means the level should not have been high enough to have generated this log. Log data is: %v", logdata, GlobalLog.loglevel, logdata)
	}
}

func genLogSuccessfully(t *testing.T, msg string, level string) {
	tempfilename := createTempFile(t)
	switch {
	case level == "FATAL":
		GlobalLog.Fatal(msg, false)
	case level == "ERROR":
		GlobalLog.Error(msg)
	case level == "WARNING":
		GlobalLog.Warning(msg)
	case level == "INFO":
		GlobalLog.Info(msg)
	case level == "DEBUG":
		GlobalLog.Debug(msg)
	default:
		t.Errorf("Log level %v is not valid.", level)
	}
	tempfile := getTempFile(t, tempfilename)
	logdata, err := convertFromJson(t, tempfile, tempfilename)
	if err != nil {
		t.Errorf("%v", err)
	}
	if logdata.Msg != msg {
		t.Errorf("Log message was %v, but it should have been %v", logdata.Msg, msg)
	}
	function, file, _, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(function).Name()
	if funcName != logdata.Function {
		t.Errorf("In log, function name is %v but it should have been %v", logdata.Function, funcName)
	}
	filePart := strings.Split(file, "src/")[1]
	if filePart != logdata.File {
		t.Errorf("In log, file is %v but it should have been %v", logdata.File, filePart)
	}

}

func successfullySetLogLevel(t *testing.T, loglevel string) {
	err := GlobalLog.SetLogLevel(loglevel)
	if err != nil {
		t.Errorf("Failed to set log level. Error: %v", err)
	}
}

func TestLevels(t *testing.T) {
	//	defer os.RemoveAll(tmpdir)
	loglevel_names := GetLogLevels()
	for _, levelname := range loglevel_names {
		successfullySetLogLevel(t, levelname)
		switch {
		case levelname == "FATAL":
			genLogSuccessfully(t, errorLogMsg, levelname)
			genLogFails(t, warningLogMsg, "ERROR")
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
			t.Errorf("Log level %v is not valid.", levelname)
		}

	}
}
