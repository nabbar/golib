/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package njs_logger

import (
	"log"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"bytes"
	"strconv"
)

const (
	tagStack  = "stack"
	tagTime   = "time"
	tagLevel  = "level"
	tagCaller = "func"
	tagFile   = "file"
	tagLine   = "line"
	tagMsg    = "message"
	tagErr    = "error"
	tagData   = "data"
)

var (
	currPkgs  = path.Base(reflect.TypeOf(IOWriter{}).PkgPath())
	modeColor = true
	timestamp = true
	filetrace = false
	enableGID = false
	enableVPR = true
)

// GetLogger return a golang log.logger instance linked with this main logger
//
// This function is useful to keep the format, mode, color, output... same as current config
/*
	msgPrefixPattern a pattern prefix to identify or comment all message passed throw this log.logger instance
	msgPrefixArgs a list of interface to apply on pattern with a fmt function
*/
func GetLogger(lvl Level, logFlags int, msgPrefixPattern string, msgPrefixArgs ...interface{}) *log.Logger {
	return log.New(GetIOWriter(lvl, msgPrefixPattern, msgPrefixArgs...), "", logFlags)
}

// GetLogger force the default golang log.logger instance linked with this main logger
//
// This function is useful to keep the format, mode, color, output... same as current config
/*
	msgPrefixPattern a pattern prefix to identify or comment all message passed throw this log.logger instance
	msgPrefixArgs a list of interface to apply on pattern with a fmt function
*/
func SetStdLogger(lvl Level, logFlags int, msgPrefixPattern string, msgPrefixArgs ...interface{}) {
	log.SetOutput(GetIOWriter(lvl, msgPrefixPattern, msgPrefixArgs...))
	log.SetPrefix("")
	log.SetFlags(logFlags)
}

// AddGID Reconfigure the current logger to add or not the thread GID before each message.
func AddGID(enable bool) {
	enableGID = enable
}

// Timestamp Reconfigure the current logger to add or not the timestamp before each message.
func Timestamp(enable bool) {
	timestamp = enable
}

// FileTrace Reconfigure the current logger to add or not the origin file/line of each message.
//
// This option is apply for all message except info message
func FileTrace(enable bool) {
	filetrace = enable
	setViperLogTrace()
}

// EnableColor Reconfigure the current logger to use color in messages format.
//
// This apply only for next message and only for TextFormat
func EnableColor() {
	modeColor = true
	updateFormatter(nilFormat)
}

// DisableColor Reconfigure the current logger to not use color in messages format.
//
// This apply only for next message and only for TextFormat
func DisableColor() {
	modeColor = false
	updateFormatter(nilFormat)
}

// EnableViperLog  or not the Gin Logger configuration
func EnableViperLog(enable bool) {
	enableVPR = enable
	setViperLogTrace()
}

func getFrame() runtime.Frame {
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, 0)
	n := runtime.Callers(0, programCounters)

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		more := true

		for more {
			var (
				frame runtime.Frame
			)

			frame, more = frames.Next()

			if strings.Contains(frame.Function, currPkgs) {
				continue
			}

			return frame
		}
	}

	return runtime.Frame{Function: "unknown", File: "unknown", Line: 0}
}

func getGID() uint64 {
	b := make([]byte, 64)

	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	n, _ := strconv.ParseUint(string(b), 10, 64) // #nosec

	return n
}

func ginTonicAddError(c *gin.Context, err error) {
	if c != nil && err != nil {
		_ = c.Error(err)
	}
}

func proceed(lvl Level) bool {
	return lvl != NilLevel && lvl <= curLevel
}
