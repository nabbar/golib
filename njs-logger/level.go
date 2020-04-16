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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/jwalterweatherman"
)

//Level a uint8 type customized with function to log message with the current log level
type Level uint8

const (
	// PanicLevel Panic level for entry log, will result on a Panic() call (trace + fatal)
	PanicLevel Level = iota
	// FatalLevel Fatal level for entry log, will result on os.Exit with error
	FatalLevel
	// ErrorLevel Error level for entry log who's meaning the caller stop his process and return to the pre caller
	ErrorLevel
	// WarnLevel Warning level for entry log who's meaning the caller don't stop his process and try to continue it
	WarnLevel
	// InfoLevel Info level for entry log who's meaning it is just an information who's have no impact on caller's process but can be useful to inform human of a state, event, success, ...
	InfoLevel
	// DebugLevel Debug level for entry log who's meaning the caller has no problem and the information is only useful to identify a potential problem who's can arrive later
	DebugLevel
	// NilLevel Nil level will never log anything and is used to completely disable current log entry. It cannot be used in the SetLogLevel function
	NilLevel
)

var (
	curLevel = InfoLevel
)

//GetCurrentLevel return the current loglevel setting in the logger. All log entry matching this level or below will be logged
func GetCurrentLevel() Level {
	return curLevel
}

// GetLevelListString return a list ([]string) of all string loglevel available
func GetLevelListString() []string {
	return []string{
		strings.ToLower(PanicLevel.String()),
		strings.ToLower(FatalLevel.String()),
		strings.ToLower(ErrorLevel.String()),
		strings.ToLower(WarnLevel.String()),
		strings.ToLower(InfoLevel.String()),
		strings.ToLower(DebugLevel.String()),
	}
}

// SetLevel Change the Level of all log entry with the Level type given in parameter. The change is apply for next log entry only
//
// If the given Level type is not matching a correct Level type, no change will be apply.
/*
	level a Level type to use to specify the new level of logger message
*/
func SetLevel(level Level) {
	switch level {
	case PanicLevel:
		curLevel = PanicLevel
		logrus.SetLevel(logrus.PanicLevel)

	case FatalLevel:
		curLevel = FatalLevel
		logrus.SetLevel(logrus.FatalLevel)

	case ErrorLevel:
		curLevel = ErrorLevel
		logrus.SetLevel(logrus.ErrorLevel)

	case WarnLevel:
		curLevel = WarnLevel
		logrus.SetLevel(logrus.WarnLevel)

	case InfoLevel:
		curLevel = InfoLevel
		logrus.SetLevel(logrus.InfoLevel)

	case DebugLevel:
		curLevel = DebugLevel
		logrus.SetLevel(logrus.DebugLevel)
	}

	DebugLevel.Logf("Change Log Level to %s", logrus.GetLevel().String())
}

func setGinLogTrace() {
	if !enableGIN {
		return
	}

	if filetrace {
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelTrace)
		return
	}

	switch curLevel {
	case PanicLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelCritical)

	case FatalLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelFatal)

	case ErrorLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelError)

	case WarnLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelWarn)

	case InfoLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelInfo)

	case DebugLevel:
		jwalterweatherman.SetStdoutThreshold(jwalterweatherman.LevelDebug)
	}
}

// GetLevelString return a valid Level Type matching the given string parameter. If the given parameter don't represent a valid level, the InfoLevel will be return
/*
	level the string representation of a Level type
*/
func GetLevelString(level string) Level {
	switch strings.ToLower(level) {
	case strings.ToLower(PanicLevel.String()):
		return PanicLevel

	case strings.ToLower(FatalLevel.String()):
		return FatalLevel

	case strings.ToLower(ErrorLevel.String()):
		return ErrorLevel

	case strings.ToLower(WarnLevel.String()):
		return WarnLevel

	case strings.ToLower(DebugLevel.String()):
		return DebugLevel

	default:
		return InfoLevel
	}
}

// Uint8 Convert the current Level type to a uint8 value. E.g. FatalLevel becomes 1.
func (level Level) Uint8() uint8 {
	return uint8(level)
}

// String Convert the current Level type to a string. E.g. PanicLevel becomes "Critical Error".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "Debug"
	case InfoLevel:
		return "Info"
	case WarnLevel:
		return "Warning"
	case ErrorLevel:
		return "Error"
	case FatalLevel:
		return "Fatal Error"
	case PanicLevel:
		return "Critical Error"
	}

	return "unknown"
}

// Log Simple function to log directly the given message with the attached log Level
/*
	message a string message to be logged with the attached log Level
*/
func (level Level) Log(message string) {
	level.logDetails(message, nil, nil, nil)
}

// Logf Simple function to log (to the attached log Level) with a fmt function a given pattern and arguments in parameters
/*
	format a string pattern for fmt function
	args a list of interface to match the references in the pattern
*/
func (level Level) Logf(format string, args ...interface{}) {
	level.logDetails(fmt.Sprintf(format, args...), nil, nil, nil)
}

// LogData Simple function to log directly the given message with given data with the attached log Level
/*
	message a string message to be logged with the attached log Level
	data an interface of data to be logged with the message. (In Text format, the data will be json marshalled)
*/
func (level Level) LogData(message string, data interface{}) {
	level.logDetails(message, data, nil, nil)
}

// WithFields Simple function to log directly the given message with given fields with the attached log Level
/*
	message a string message to be logged with the attached log Level
	fields a map of string key and interfaces value for a complete list of field ("field name" => value interface)
*/
func (level Level) WithFields(message string, fields map[string]interface{}) {
	level.logDetails(message, nil, nil, fields)
}

// LogError Simple function to log directly the given error with the attached log Level.
//
// How iot works :
//  + when the err is a valid error, this function will :
//  +--- log the Error with the attached log Level
//  +--- return true
//  + when the err is nil, this function will :
//  +--- return false
/*
	err an error object message to be logged with the attached log Level
*/
func (level Level) LogError(err error) bool {
	return level.LogGinErrorCtx(NilLevel, "", err, nil)
}

// LogErrorCtx Function to test, log and inform about the given error object
//
// How iot works :
//  + when the err is a valid error, this function will :
//  +--- log the Error with the attached log Level
//  +--- return true
//  + when the err is nil, this function will :
//  +--- use the levelElse if valid to inform with context there is no error found
//  +--- return false
/*
	levelElse level used if the err is nil before returning a False result
	context a string for the context of the current test of the error
	err a error object to be log with the attached log level before return true, if the err is nil, the levelElse is used to log there are no error and return false
*/
func (level Level) LogErrorCtx(levelElse Level, context string, err error) bool {
	return level.LogGinErrorCtx(levelElse, context, err, nil)
}

// LogErrorCtxf Function to test, log and inform about the given error object, but with a context based on a pattern and matching args
//
// How iot works :
//  + when the err is a valid error, this function will :
//  +--- log the Error with the attached log Level
//  +--- return true
//  + when the err is nil, this function will :
//  +--- use the levelElse if valid to inform with context there is no error found
//  +--- return false
/*
	levelElse level used if the err is nil before returning a False result
	contextPattern a pattern string for the context of the current test of the error. This string will be used in a fmt function as pattern string
	err a error object to be log with the attached log level before return true, if the err is nil, the levelElse is used to log there are no error and return false
	args a list of interface for the context of the current test of the error. This list of interface will be used in a fmt function as the matching args for the pattern string
*/
func (level Level) LogErrorCtxf(levelElse Level, contextPattern string, err error, args ...interface{}) bool {
	return level.LogGinErrorCtx(levelElse, fmt.Sprintf(contextPattern, args...), err, nil)
}

// LogGinErrorCtxf Function to test, log and inform about the given error object, but with a context based on a couple of pattern and matching args.
// This function will also add an Gin Tonic Error if the c parameters is a valid GinTonic Context reference.
//
// How iot works :
//  + when the err is a valid error, this function will :
//  +--- log the Error with the attached log Level
//  +--- if the Context Gin Tonic is valid, add the Error into this context
//  +--- return true
//  + when the err is nil, this function will :
//  +--- use the levelElse if valid to inform with context there is no error found
//  +--- return false
/*
	levelElse level used if the err is nil before returning a False result
	contextPattern a pattern string for the context of the current test of the error. This string will be used in a fmt function as pattern string
	err a error object to be log with the attached log level before return true, if the err is nil, the levelElse is used to log there are no error and return false
	c a valid Go GinTonic Context reference to add current error to the Gin Tonic Error Context
	args a list of interface for the context of the current test of the error. This list of interface will be used in a fmt function as the matching args for the pattern string
*/
func (level Level) LogGinErrorCtxf(levelElse Level, contextPattern string, err error, c *gin.Context, args ...interface{}) bool {
	return level.LogGinErrorCtx(levelElse, fmt.Sprintf(contextPattern, args...), err, c)
}

// LogGinErrorCtx Function to test, log and inform about the given error object
// This function will also add an Gin Tonic Error if the c parameters is a valid GinTonic Context reference.
//
// How iot works :
//  + when the err is a valid error, this function will :
//  +--- log the Error with the attached log Level
//  +--- if the Context Gin Tonic is valid, add the Error into this context
//  +--- return true
//  + when the err is nil, this function will :
//  +--- use the levelElse if valid to inform with context there is no error found
//  +--- return false
/*
levelElse level used if the err is nil before returning a False result
context a string for the context of the current test of the error
err a error object to be log with the attached log level before return true, if the err is nil, the levelElse is used to log there are no error and return false
c a valid Go GinTonic Context reference to add current error to the Gin Tonic Error Context
*/
func (level Level) LogGinErrorCtx(levelElse Level, context string, err error, c *gin.Context) bool {
	if err != nil {
		level.logDetails(fmt.Sprintf("KO : %s", context), nil, err, nil)
		ginTonicAddError(c, err)
		return true
	} else if proceed(levelElse) {
		levelElse.logDetails(fmt.Sprintf("OK : %s", context), nil, err, nil)
	}

	return false
}

func (level Level) logDetails(message string, data interface{}, err error, fields logrus.Fields) {
	if !proceed(level) {
		return
	}

	var tags = make(map[string]interface{}, 0)

	if enableGID {
		tags[tagStack] = getGID()
	}

	if timestamp {
		tags[tagTime] = time.Now().Format(time.RFC3339Nano)
	}

	tags[tagTime] = level.String()

	if filetrace && level != InfoLevel {
		frame := getFrame()
		tags[tagCaller] = frame.Function
		tags[tagFile] = frame.File
		tags[tagLine] = frame.Line
	}

	tags[tagMsg] = message
	tags[tagErr] = err
	tags[tagData] = data

	var (
		ent = logrus.NewEntry(logrus.StandardLogger())
		msg string
	)

	if fields != nil && len(fields) > 0 {
		ent.WithFields(fields)
	}

	switch curFormat {
	case TextFormat:
		if _, ok := tags[tagStack]; ok {
			msg += fmt.Sprintf("[%d] ", tags[tagStack])
		}

		if _, ok := tags[tagCaller]; ok {
			msg += fmt.Sprintf("[%s] ", tags[tagCaller])
		}

		var line string
		if _, ok := tags[tagLine]; ok {
			line = fmt.Sprintf("(%d) ", tags[tagLine])
		}

		if _, ok := tags[tagFile]; ok || len(line) > 0 {
			msg += fmt.Sprintf("[%s%s] ", line, tags[tagFile])
		}

		msg += fmt.Sprintf("%s", tags[tagMsg])

		if tags[tagErr] != nil {
			msg += fmt.Sprintf(" -- err : %v", err)
		}

		if tags[tagData] != nil {
			if str, err := json.MarshalIndent(data, "", "  "); err == nil {
				msg += fmt.Sprintf(" -- data : \n%s", string(str))
			} else {
				msg += fmt.Sprintf(" -- data : %v", err)
			}
		}

	case JsonFormat:
		ent.WithFields(tags)
		msg = tags[tagMsg].(string)
	}

	switch level {
	case DebugLevel:
		ent.Debugln(msg)

	case InfoLevel:
		ent.Infoln(msg)

	case WarnLevel:
		ent.Warnln(msg)

	case ErrorLevel:
		ent.Errorln(msg)

	case FatalLevel:
		ent.Fatalln(msg)

	case PanicLevel:
		ent.Panicln(msg)
	}
}
