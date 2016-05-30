package sglogger_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jaloren/sglogger"
)

var GlobalLogger = sglogger.GlobalLogger

func TestRotation(t *testing.T) {
	threshold := int64(10485760)
	skew := threshold + int64(1000)
	logdir, err := ioutil.TempDir("/tmp", "testsglogger")
	if err != nil {
		t.Fatalf("Failed to create directory %v. Error: %v", logdir, err.Error())
	}
	defer os.RemoveAll(logdir)
	_, err = GlobalLogger.SetFileHandler(logdir, "test")
	if err != nil {
		t.Fatalf("Failed create log file. Error: %v", err.Error())
	}
	err = GlobalLogger.StartRotationPolicy(threshold)
	if err != nil {
		t.Fatalf("Failed to start rotation policy. Error: %v", err.Error())
	}
	for inc, _ := range [100000]int{} {
		logmsg := fmt.Sprintf("Loop %v - rotation test", inc)
		GlobalLogger.Error(logmsg)
	}
	err = GlobalLogger.StopRotationPolicy()
	if err != nil {
		t.Fatalf("Failed to stop the rotation policy. Error: %v", err.Error())
	}
	files, err := ioutil.ReadDir(logdir)
	if err != nil {
		t.Fatalf("Failed to get files from %v. Error: %v", logdir, err.Error())
	}
	if len(files) != 4 {
		t.Fatalf("There is more or less than 4 files in directory %v", logdir)
	}
	for _, f := range files {
		size := f.Size()
		if size > skew {
			t.Fatalf("File %v size %d exceeds threshold skew %d", f.Name(), size, skew)
		}
	}
}
