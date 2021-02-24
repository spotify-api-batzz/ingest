package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type VerbosityLevel int

type LogCallback func(message interface{}, verbosityLevel VerbosityLevel, data ...interface{})

var callback LogCallback
var verbosity VerbosityLevel

const (
	Debug VerbosityLevel = iota
	Notice
	Warning
	Error
)

type Logger struct {
	CurrentVerbosity string
	Callback         LogCallback
}

type LogMessage struct {
	Verbosity VerbosityLevel
	Message   string
}

func Log(message interface{}, level VerbosityLevel, data ...interface{}) {
	callback(message, level, data...)

	var datetime = time.Now()

	if int(level) >= int(verbosity) { // Only print appropriate verbosity messages
		fmt.Println(fmt.Sprintf("[%s] [%s]: %s (%s)", datetime.Format(time.RFC3339), level.String(), message, getFrame(2).Function))
	}
}

func (d VerbosityLevel) String() string {
	return [...]string{"Debug", "Notice", "Warning", "Error"}[d]
}

func SetCallback(cb LogCallback) {
	callback = cb
}

func SetVerbosity(level string) {
	verbosityLevels := make(map[string]VerbosityLevel)
	verbosityLevels["debug"] = Debug
	verbosityLevels["notice"] = Notice
	verbosityLevels["warning"] = Warning
	verbosityLevels["error"] = Error

	verbosity = verbosityLevels[strings.ToLower(level)]
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
