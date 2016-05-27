package sglogger

import (
	"errors"
	golog "log"
	"os"
)

var (
	GlobalLog = &SimpleLog{
		simpleLogger: golog.New(os.Stderr, "", 0),
		loglevel:     1,
		handler:      []string{"stderr"},
		logfile:      "",
		freeze:       false,
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

func GetGlobalLog() *SimpleLog {
	return GlobalLog
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
