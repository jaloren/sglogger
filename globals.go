package sglogger

import (
	"encoding/json"
	"errors"
	"fmt"
	golog "log"
	"os"
	"strings"
	"time"
)

var (
	GlobalLogger = &SimpleLogger{
		goLogger: golog.New(os.Stderr, "", 0),
		loglevel: 1,
		handler:  []string{"stderr"},
		logfile:  "",
		freeze:   false,
	}
	loglevels = map[string]int{
		"FATAL":   0,
		"ERROR":   1,
		"WARNING": 2,
		"INFO":    3,
		"DEBUG":   4,
	}
	frozeErrMsg = errors.New("The global log struct has been frozen. Once the struct has been frozen, none of its fields may be changed.")
)

type logMsg struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Function  string `json:"func"`
	File      string `json:"file"`
	Lineno    int    `json:"lineno"`
}

func getLogMsg(level string, msg string, funcName string, file string, lineno int, closeChar string) string {
	fileParts := strings.Split(file, "src/")
	logmsg := logMsg{
		Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		Level:     level,
		Msg:       msg,
		Function:  funcName,
		File:      fileParts[1],
		Lineno:    lineno,
	}
	jsontext, err := json.MarshalIndent(logmsg, "", "\t")
	if err != nil {
		return fmt.Sprintf("Failed to marshal struct %+v into a JSON object. It's highly likely that there's a bug in the logging library. Error: %v", logmsg, err)
	}
	return string(jsontext) + closeChar
}

func GetGlobalLogger() *SimpleLogger {
	return GlobalLogger
}

func GetLogLevels() []string {
	keys := make([]string, len(loglevels))
	i := 0
	for k := range loglevels {
		keys[i] = k
		i++
	}
	return keys
}
