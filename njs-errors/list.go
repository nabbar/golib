/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package njs_errors

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

var (
	currPkgs        = path.Base(reflect.TypeOf(ERR_UNKNOWN).PkgPath())
	_listCodeErrors = make(map[string]ErrorType, 0)
)

// SetErrorCode Register a new error with code and an error string as ErrorType
func SetErrorCode(code string, err ErrorType) {
	if _listCodeErrors == nil || len(_listCodeErrors) < 1 {
		_listCodeErrors = make(map[string]ErrorType, 0)
	}

	_listCodeErrors[code] = err
}

// SetErrorCodeString Register a new error with code and an error string as string
func SetErrorCodeString(code, err string) {
	SetErrorCode(code, ErrorType(err))
}

// DelErrorCode Remove an error with code from the register list
func DelErrorCode(code string) {
	var _lst = _listCodeErrors

	DelAllErrorCode()

	for k, v := range _lst {
		if k != code {
			_listCodeErrors[k] = v
		}
	}
}

// DelAllErrorCode Clean the complete list of couple code - error
func DelAllErrorCode() {
	_listCodeErrors = make(map[string]ErrorType, 0)
}

// GetErrorCode return an ErrorCode interface mapped to code given in parameters.
// If the code is not found an 'ERR_UNKNOWN' will be used instead of the awaiting error
// If an origin error is given in params, this origin error will be used in the reference of generated error or string
func GetErrorCode(code string, origin error, argPattern ...interface{}) ErrorCode {
	return getErrorCode(code, origin, getNilFrame(), argPattern...)
}

// GetTraceErrorCode return an ErrorCode interface mapped to given params code.
// Deprecated: this function is replaced by GetErrorCodeTrace
func GetTraceErrorCode(code string, origin error) ErrorCode {
	return getErrorCode(code, origin, getFrame())
}

// GetErrorCodeTrace return an ErrorCode interface mapped to given params code.
// If the code is not found an 'ERR_UNKNOWN' will be used instead of the awaiting error
// If an origin error is given in params, this origin error will be used in the reference of generated error or string
// This function add a trace of error generated
func GetErrorCodeTrace(code string, origin error, argPattern ...interface{}) ErrorCode {
	return getErrorCode(code, origin, getFrame(), argPattern...)
}

func getErrorCode(code string, origin error, trace runtime.Frame, argList ...interface{}) ErrorCode {
	var (
		e  ErrorType
		ok bool
	)

	if e, ok = _listCodeErrors[code]; !ok {
		e = ERR_UNKNOWN
	} else {
		e = ErrorType(fmt.Sprintf(e.String(), argList...))
	}

	return &errorCode{
		code:  code,
		err:   e,
		ori:   origin,
		trace: trace,
	}
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

	return getNilFrame()
}

func getNilFrame() runtime.Frame {
	return runtime.Frame{Function: "", File: "", Line: 0}
}
