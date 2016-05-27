package sglogger

import (
	"encoding/json"
	"strings"
	"time"
)

type logMsg struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Function  string `json:"func"`
	File      string `json:"file"`
	Lineno    int    `json:"lineno"`
}

func getLogMsg(level string, msg string, funcName string, file string, lineno int) string {
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
		panic(err)
	}
	return string(jsontext)
}
