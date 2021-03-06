# sglogger
golang library that provides a simple global logger which supports log levels. There are tons of logging libraries with highly sophististiced functionality that allows someone to implement elaborate logging facilities and output. That is *not* this library, so if that's what you are looking I'd recommend looking at logrus. This library does a lot less and because of that its much simpler to use. Someone can pick this library and start using it in seconds (or so I think) but its better then using the fmt.Printf or the standard log library. 

I wanted a highly opinionated logging library that:
- displayed a simple log message
- supported manual and policy based log rotation that is goroutine safe. The rotation policy is based solely on log size
- supported basic logic levels: FATAL, ERROR, WARNING, INFO, and DEBUG
- a single consistent logger used throughout an application. Once the logger is initialized, it can't be modified by another package. For example, you don't have a risk that package A changes the log level to INFO and then package B switches it to ERROR. Or package C decides to start logging to some other file off to the side. Once the log level and output is set for the logger, it can't be changed.
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

var globalLog = sglogger.GlobalLogger

func init(){
	logdir := "/tmp/httpbenchmark"
	filePrefix := "test"
	err := globalLog.SetHandlers(logdir,filePrefix)
	if err != nil {
		panic(err)
	}
	globalLog.SetLogLevel("INFO")
	globalLog.Freeze(true)
}
```

In above example, the init function does the following:
- create a directory to put the log file in.
- the log file name is composed of two parts. The first part is the file prefix that a user defined. The second part is the UTC timestamp with millisecond precision (the timestamp is generated when the file is initially created) and the file extension .llog. Note the file prefix only allows alphanumeric characters.
- generates a UTC timestamp and converts that into a string.
- create a log filename with the utc timestamp burned into it.
- calls the SetHandler method to specify that logger will write data to this path that was just constructed.
- sets the log level to INFO
- the last method Freeze is *critical*. This prevents any other code from calling setter methods that will change the behavior of the global logger, such as the log level or the file that the logger writes to.

### Different Log Outputs

You also have the option of only writing to a file or only writing to a console. Instead of the SetHandlers method, which does both, simply invoke either SetConsoleHandler or SetFileHandler methods. 

### Log an Event

In a package where you want to log an event, first define a variable at the top of the package like so.
```golang
var globalLog = sglogger.GlobalLogger
globalLog.Error("Error message")
globalLog.Warning("Info message")
```

### Automatically Rotate Logs Based on File Size
The most common reason for log rotation is to keep the log file from growing to large. This library supports a log rotation policy that automatically rotates logs based on a user-defined file size reached in bytes. The following example demonstrates this policy.
```golang

package main

var globalLog = sglogger.GlobalLogger


func init(){
	logdir := "/tmp/lgodir"
	_, err := globalLog.SetFileHandler(logdir, "test")
	if err != nil {
		panic(err)
	}
	err = globalLog.StartRotationPolicy(10485760)
	if err != nil {
		panic(err)
	}
	globalLog.SetLogLevel("INFO")
	globalLog.Freeze(true)
}

func main(){
     defer globalLog.StopRotationPolicy()
}
```
Please note the following important points:
- The int64 passed to the StartRotationPolicy method is the file size threshold in bytes that must be reached before a log rotation is triggered. This does *not* mean the log will be rotated at precisely that size. The window of time between when the rotation is trigged and actually completes can allow some amount of additional logs to be added to the file before the rotation completes.
- StartRotationPolicy starts a goroutine to monitor log file sizes and initiate log rotations. To avoid, deadlocks or infinite loops, use the StopRotationPolicy to gracefully terminate the goroutine and ensure that the last file rotated is a valid json document. Note that the StopRotationPolicy may only be executed at most one time and will return an error if executed multiple times.


## Parsing Log Files

The simple global logger's format is json. When the logger writes an event into a file, its actually appending this to a json array. Since a json array is an *ordered* collection of items, the log events appear in the order in which they were generated. For example, if you have three items in the array, then the second log event occurred after the first log event but before the third log event. By default, this will *not* be a valid JSON document. Because objects are being appended to the array, the logger does not add the closing brace to make the array a valid JSON object. 

To make the log file a valid JSON document, you must execute the Rotate method after all logging has completed. To ensure you only terminate a log file after all log events have completed, make this a defer statement in the main method. Here's an example.

```golang
package main

var globalLog = GetGlobalLogger

func init(){
	logdir := "/tmp/httpbenchmark"
	filePrefix := "test"
	err := globalLog.SetHandlers(logdir,filePrefix)
	if err != nil {
		panic(err)
	}
	globalLog.SetLogLevel("INFO")
	globalLog.Freeze(true)
}

func main(){
     defer globalLog.Rotate()
}

```

Note you do not need and should avoid using the Rotate method if you are using the rotation policy.

##Supported Log Levels
Note that like every other logger on the planet that supports levels (except golang's std lib, something i still can't quite believe), if the log level is set lower than the log level you have chosen to log at, then the event will not be logged. For example, if the log level is ERROR and the event is logged at WARNING, then the event is not logged.
 
- FATAL => 0
- ERROR => 1
- WARNING => 2
- INFO => 3
- DEBUG => 4