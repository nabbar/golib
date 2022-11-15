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
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const (
	PathSeparator = "/"
	pathVendor    = "vendor"
	pathMod       = "mod"
	pathPkg       = "pkg"
	pkgRuntime    = "runtime"
)

var (
	filterPkg = path.Clean(ConvPathFromLocal(reflect.TypeOf(UNK_ERROR).PkgPath()))
	currPkgs  = path.Base(ConvPathFromLocal(filterPkg))
)

func ConvPathFromLocal(str string) string {
	return strings.Replace(str, string(filepath.Separator), PathSeparator, -1)
}

func init() {
	if i := strings.LastIndex(filterPkg, PathSeparator+pathVendor+PathSeparator); i != -1 {
		filterPkg = filterPkg[:i+1]
	}
}

func getFrame() runtime.Frame {
	// Set size to targetFrameIndex+20 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, 20, 255)
	n := runtime.Callers(2, programCounters)

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

			return runtime.Frame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			}
		}
	}

	return getNilFrame()
}

func getFrameVendor() []runtime.Frame {
	// Set size to targetFrameIndex+20 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, 20, 255)
	n := runtime.Callers(2, programCounters)

	res := make([]runtime.Frame, 0)

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		more := true

		for more {
			var (
				frame runtime.Frame
			)

			frame, more = frames.Next()

			item := runtime.Frame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			}

			if strings.Contains(item.Function, currPkgs) {
				continue
			} else if strings.Contains(ConvPathFromLocal(frame.File), PathSeparator+pathVendor+PathSeparator) {
				continue
			} else if strings.HasPrefix(frame.Function, pkgRuntime) {
				continue
			} else if frameInSlice(res, item) {
				continue
			}

			res = append(res, item)

			if len(res) > 4 {
				return res
			}
		}
	}

	return res
}

func frameInSlice(s []runtime.Frame, f runtime.Frame) bool {
	if len(s) < 1 {
		return false
	}

	for _, i := range s {
		if i.Function != f.Function {
			continue
		}

		if i.File != f.File {
			continue
		}

		if i.Line != i.Line {
			continue
		}

		return true
	}

	return false
}

func getNilFrame() runtime.Frame {
	return runtime.Frame{Function: "", File: "", Line: 0}
}

func filterPath(pathname string) string {
	var (
		filterMod    = PathSeparator + pathPkg + PathSeparator + pathMod + PathSeparator
		filterVendor = PathSeparator + pathVendor + PathSeparator
	)

	pathname = ConvPathFromLocal(pathname)

	if i := strings.LastIndex(pathname, filterMod); i != -1 {
		i = i + len(filterMod)
		pathname = pathname[i:]
	}

	if i := strings.LastIndex(pathname, filterPkg); i != -1 {
		i = i + len(filterPkg)
		pathname = pathname[i:]
	}

	if i := strings.LastIndex(pathname, filterVendor); i != -1 {
		i = i + len(filterVendor)
		pathname = pathname[i:]
	}

	pathname = path.Clean(pathname)

	return strings.Trim(pathname, PathSeparator)
}
