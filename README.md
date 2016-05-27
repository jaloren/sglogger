# sglogger
golang library that provides a simple global logger which supports log levels. There are tons of logging libraries with highly sophististiced functionality that allows someone to implement elaborate logging facilities and output. That is *not* this library, so if that's what you are looking I'd recommend looking at logrus. This library does a lot less and because of that its much simpler to use. Someone can pick this library and start using it in seconds (or so I think) but its better then using the fmt.Printf or the standard log library. 

I wanted a logging library that:
- displayed a simple log message
- supported basic logic levels: FATAL, ERROR, WARNING, INFO, and DEBUG
- was thread and goroutine safe, without using locks!, and more importantly would be a single consistent logger used throughout an application. Once the logger is initialized, it can't be modified by another package. For example, you don't have a risk that package A changes the log level to INFO and then package B switches it to ERROR. Or package C decides to start logging to some other file off to the side. Once the log level and output is set for the logger, it can't be changed.
- would log to the console and to a file in a structured format that is human readable and machine parseable. This library does it in json.
- every log message should contain the function name, the file, and line number where the log statement was executed so a developer can easily locate where the log was generated. IMHO, that's critical for debugging.
- log timestamps are *always* in UTC to avoid timezone conversion hell when trying to aggregate logs.

## Log Message Example

```javascript
{
	"timestamp": "2016-05-27 22:42:48 UTC",
	"level": "ERROR",
	"msg": "This is an error message",
	"func": "github.com/jaloren/sglogger.genLogSuccessfully",
	"file": "github.com/jaloren/sglogger/loglevels_test.go",
	"lineno": 82
}
```

## Code Examples

### Initializing the logger

Since this is intended to be an immutable global logger used throughout an application, the expectation is that the logger would be set up in the main package's init method.

```golang

var globalLog = simplelog.GetGlobalLog()

func init(){
	logdir := "/tmp/httpbenchmark"
	os.MkdirAll(logdir, 0755)
	timestamp := time.Now().UTC().Format("2006-01-02T150405")
	filename := fmt.Sprintf("%s/client_%s.log", logdir, timestamp)
	err := globalLog.SetHandlers(filename,true)
	if err != nil {
		panic(err)
	}
	globalLog.SetLogLevel("INFO")
	globalLog.Freeze(true)
}
```

In above example, the init function does the following:
- create a directory to put the log file in.
- generates a UTC timestamp and converts that into a string.
- create a log filename with the utc timestamp burned into it.
- calls the SetHandler method to specify that logger will write data to this path that was just constructed.
- sets the log level to INFO
- the last method Freeze is *critical*. This prevents any other code from calling setter methods that will change the behavior of the global logger, such as the log level or the file that the logger writes to.

### Different Log Outputs

You also have the option of only writing to a file or only writing to a console. Instead of the SetHandlers method, which does both, simply invoke either SetConsoleHandler or SetFileHandler methods. If you don't want to write to a file or console, then pass an object that satistifies the io.Writer interface to the SetCustomHandler method.


### Log an Event

In a package where you want to log an event, first define a variable at the top of the package like so.
```
var globalLog = simplelog.GetGlobalLog()
globalLog.Error("Error message")
globalLog.Warning("Info message")

```

##Supported Log Levels
Note that like every other logger on the planet that supports levels (except golang's std lib, something i still can't quite believe), if the log level is set lower than the log level you have chosen to log at, then the event will not be logged. For example, if the log level is ERROR and the event is logged at WARNING, then the event will not in fact be logged.
 
- FATAL => 0
- ERROR => 1
- WARNING => 2
- INFO => 3
- DEBUG => 4