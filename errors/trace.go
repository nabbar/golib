/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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
 *
 */

package errors

import (
	"path"
	"reflect"
	"runtime"
	"strings"
)

var (
	currPkgs  = path.Base(reflect.TypeOf(UNK_ERROR).PkgPath())
	filterPkg = path.Clean(reflect.TypeOf(UNK_ERROR).PkgPath())
)

func init() {
	if i := strings.LastIndex(filterPkg, "/vendor/"); i != -1 {
		filterPkg = filterPkg[:i+1]
	}
}

func getFrame() runtime.Frame {
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, 10, 255)
	n := runtime.Callers(1, programCounters)

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
