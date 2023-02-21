/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// @deprecated: only for retro compatibility
var defaultLogger Logger

func init() {
	defaultLogger = New(context.Background)
	defaultLogger.SetLevel(InfoLevel)
}

// GetDefault return the default logger
// @deprecated: create a logger and call GetLevel() like New().GetLevel()
func GetDefault() Logger {
	return defaultLogger
}

// GetCurrentLevel return the current loglevel setting in the logger. All log entry matching this level or below will be logged.
// @deprecated: create a logger and call GetLevel() like New().GetLevel()
func GetCurrentLevel() Level {
	return defaultLogger.GetLevel()
}

// SetLevel Change the Level of all log entry with the Level type given in parameter. The change is apply for next log entry only.
// If the given Level type is not matching a correct Level type, no change will be apply.
// @deprecated: create a logger and call GetLevel() like New().GetLevel...
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// AddGID Reconfigure the current logger to add or not the thread GID before each message.
// @deprecated: create a logger and update the options like New().SetOptions...
func AddGID(enable bool) {
	opt := defaultLogger.GetOptions()
	opt.DisableStack = !enable
	_ = defaultLogger.SetOptions(opt)
}

// Timestamp Reconfigure the current logger to add or not the timestamp before each message.
// @deprecated: create a logger and update the options like New().SetOptions...
func Timestamp(enable bool) {
	opt := defaultLogger.GetOptions()
	opt.DisableTimestamp = !enable
	_ = defaultLogger.SetOptions(opt)
}

// IsTimeStamp will return true if timestamp is added or not  on log message
// @deprecated: create a logger and get the options like New().GetOptions...
func IsTimeStamp() bool {
	return !defaultLogger.GetOptions().DisableTimestamp
}

// FileTrace Reconfigure the current logger to add or not the origin file/line of each message.
// This option is apply for all message except info message.
// @deprecated: create a logger and update the options like New().SetOptions...
func FileTrace(enable bool) {
	opt := defaultLogger.GetOptions()
	opt.EnableTrace = enable
	_ = defaultLogger.SetOptions(opt)
}

// IsFileTrace will return true if trace is added or not on log message
// @deprecated: create a logger and get the options like New().GetOptions...
func IsFileTrace() bool {
	return defaultLogger.GetOptions().EnableTrace
}

// ModeColor will reconfigure the current logger to use or not color in messages format.
// This apply only for next message and only for TextFormat.
// @deprecated: create a logger and update the options like New().SetOptions...
func ModeColor(enable bool) {
	opt := defaultLogger.GetOptions()
	opt.DisableColor = !enable
	_ = defaultLogger.SetOptions(opt)
}

// IsModeColor will return true if color is configured on log message
// @deprecated: create a logger and get the options like New().GetOptions...
func IsModeColor() bool {
	return !defaultLogger.GetOptions().DisableColor
}

// EnableColor Reconfigure the current logger to use color in messages format.
// This apply only for next message and only for TextFormat.
// @deprecated: create a logger and update the options like New().SetOptions...
func EnableColor() {
	ModeColor(true)
}

// DisableColor Reconfigure the current logger to not use color in messages format.
// This apply only for next message and only for TextFormat.
// @deprecated: create a logger and update the options like New().SetOptions...
func DisableColor() {
	ModeColor(false)
}

// EnableViperLog enable or not the Gin Logger configuration.
// @deprecated: create a logger and call function SetSPF13Level like New().SetSPF13Level...
func EnableViperLog(enable bool) {
	defaultLogger.SetSPF13Level(defaultLogger.GetLevel(), nil)
}

// SetTracePathFilter customize the filter apply to filepath on trace.
// @deprecated: create a logger and update the options like New().SetOptions...
func SetTracePathFilter(path string) {
	opt := defaultLogger.GetOptions()
	opt.TraceFilter = path
	_ = defaultLogger.SetOptions(opt)
}

// Log Simple function to log directly the given message with the attached log Level.
/*
	message a string message to be logged with the attached log Level
*/
//@deprecated: create a logger and call one of this function : New().Debug, New().Info, New().Warning, New().Error, New().Fatal, New().Panic, New().LogDetails or New().Entry
func (l Level) Log(message string) {
	defaultLogger.LogDetails(l, message, nil, nil, nil)
}

// Logf Simple function to log (to the attached log Level) with a fmt function a given pattern and arguments in parameters.
/*
	format a string pattern for fmt function
	args a list of interface to match the references in the pattern
*/
//@deprecated: create a logger and call one of this function : New().Debug, New().Info, New().Warning, New().Error, New().Fatal, New().Panic, New().LogDetails or New().Entry
func (l Level) Logf(format string, args ...interface{}) {
	defaultLogger.LogDetails(l, fmt.Sprintf(format, args...), nil, nil, nil)
}

// LogData Simple function to log directly the given message with given data with the attached log Level.
/*
	message a string message to be logged with the attached log Level
	data an interface of data to be logged with the message. (In Text format, the data will be json marshaled)
*/
//@deprecated: create a logger and call one of this function : New().Debug, New().Info, New().Warning, New().Error, New().Fatal, New().Panic, New().LogDetails or New().Entry
func (l Level) LogData(message string, data interface{}) {
	defaultLogger.LogDetails(l, message, data, nil, nil)
}

// WithFields Simple function to log directly the given message with given fields with the attached log Level.
/*
	message a string message to be logged with the attached log Level
	fields a map of string key and interfaces value for a complete list of field ("field name" => value interface)
*/
//@deprecated: create a logger and call one of this function : New().Debug, New().Info, New().Warning, New().Error, New().Fatal, New().Panic, New().LogDetails or New().Entry
func (l Level) WithFields(message string, fields Fields) {
	defaultLogger.LogDetails(l, message, nil, nil, fields)
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
//@deprecated: create a logger and call one of this function : New().CheckError or New().Entry.Check
func (l Level) LogError(err error) bool {
	return defaultLogger.CheckError(l, NilLevel, "", err)
}

// LogErrorCtx Function to test, log and inform about the given error object.
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
//@deprecated: create a logger and call one of this function : New().CheckError or New().Entry.Check
func (l Level) LogErrorCtx(levelElse Level, context string, err error) bool {
	return defaultLogger.Entry(l, context).ErrorAdd(true, err).Check(levelElse)
}

// LogErrorCtxf Function to test, log and inform about the given error object, but with a context based on a pattern and matching args.
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
//@deprecated: create a logger and call one of this function : New().CheckError or New().Entry.Check
func (l Level) LogErrorCtxf(levelElse Level, contextPattern string, err error, args ...interface{}) bool {
	return defaultLogger.Entry(l, contextPattern, args...).ErrorAdd(true, err).Check(levelElse)
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
//@deprecated: create a logger and call one of this function : New().CheckError or New().Entry.SetGinContext.Check
func (l Level) LogGinErrorCtxf(levelElse Level, contextPattern string, err error, c *gin.Context, args ...interface{}) bool {
	return defaultLogger.Entry(l, contextPattern, args...).SetGinContext(c).ErrorAdd(true, err).Check(levelElse)
}

// LogGinErrorCtx Function to test, log and inform about the given error object.
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
c a valid Go GinTonic Context reference to add current error to the Gin Tonic Error Context.
*/
//@deprecated: create a logger and call one of this function : New().CheckError or New().Entry.SetGinContext.Check
func (l Level) LogGinErrorCtx(levelElse Level, context string, err error, c *gin.Context) bool {
	return defaultLogger.Entry(l, context).SetGinContext(c).ErrorAdd(true, err).Check(levelElse)
}
