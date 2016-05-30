package sglogger

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

var (
	chanClosed = false
)

func (l *SimpleLogger) gzipLog() error {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	logdata, err := ioutil.ReadFile(l.logfile)
	if err != nil {
		return fmt.Errorf("Failed to read log file. Error: %v", err.Error())
	}
	w.Write(logdata)
	w.Close()
	err = ioutil.WriteFile(l.logfile+".gz", b.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("Failed to write gzip data to disk. Error: %v", err.Error())
	}
	err = os.Remove(l.logfile)
	if err != nil {
		return fmt.Errorf("Failed to remove uncompress file %v. Error: %v", l.logfile, err.Error())
	}
	return nil
}

func (l *SimpleLogger) getFileSize(tmpLogFile *os.File) int64 {
	fileInfo, err := os.Stat(l.logfile)
	if err != nil {
		msg := fmt.Sprintf("Rotating logs failed through the rotation policy because stating log file %v failed. Error: %v\n", l.logfile, err)
		tmpLogFile.WriteString(msg)
		return 0
	}
	return fileInfo.Size()
}

func (l *SimpleLogger) StartRotationPolicy(threshold int64) error {
	if l.logfile == "" {
		return fmt.Errorf("Unable to start the rotation policy because a log file does not exist.")
	}
	if l.freeze {
		return frozeErrMsg
	} else {
		l.freeze = true
	}
	tmpLogFile, tempdir, err := getTempFile("rotation")
	if err != nil {
		return fmt.Errorf("Failed to create a temp log file for recording rotation policy failures. Error: %v", err.Error())
	}
	l.quit = make(chan bool)
	go func() {
		for {
			select {
			case <-l.quit:
				if fileInfo, _ := os.Stat(tmpLogFile.Name()); fileInfo.Size() < 2 {
					os.RemoveAll(tempdir)
				}
				chanClosed = true
				return
			default:
				fileSize := l.getFileSize(tmpLogFile)
				if fileSize == 0 {
					continue
				}
				if fileSize >= threshold {
					err := l.Rotate()
					if err != nil {
						msg := fmt.Sprintf("Rotation policy failed to rotate logs. Error: %v\n", err)
						tmpLogFile.WriteString(msg)
						fmt.Print(msg)
					}
				}
			}
		}
	}()
	return nil
}

func (l *SimpleLogger) StopRotationPolicy() error {
	if chanClosed {
		return fmt.Errorf("Rotation policy has already been stopped. It cannot be stopped more than once")
	}
	l.quit <- true
	err := l.Rotate()
	return err
}

func (l *SimpleLogger) Rotate() error {
	if l.logfile == "" {
		return errors.New("Unable to rotate logs because the log directory has not been set.")
	}
	l.rotateLock.Lock()
	defer l.rotateLock.Unlock()
	function, file, line, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(function).Name()
	logmsg := getLogMsg("FATAL", "Rotating log file", funcName, file, line, "]")
	_, err := l.logfd.WriteString(logmsg)
	if err != nil {
		return fmt.Errorf("When attempting to rotate logs, failed to write string to log %v. Error: %v", l.logfile, err.Error())
	}
	l.logfd.Sync()
	err = l.gzipLog()
	if err != nil {
		return fmt.Errorf("When attempting to rotate logs, the process of gzipping log file %v failed. Error: %v", l.logfile, err)
	}
	_, err = l.setInternalFileHandler(l.logdir, l.filePrefix)
	if err != nil {
		return fmt.Errorf("When attempting to rotate logs, failed to create a new log. Error: %v", err)
	}
	return nil
}
