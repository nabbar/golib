/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package static

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the static package.
// These errors use the github.com/nabbar/golib/errors package for error management.
const (
	// ErrorParamEmpty indicates that required parameters are empty or missing
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgStatic

	// ErrorFileInfo indicates failure to retrieve file information
	ErrorFileInfo

	// ErrorFileOpen indicates failure to open a file from the embedded filesystem
	ErrorFileOpen

	// ErrorFiletemp indicates failure to create a temporary file
	ErrorFiletemp

	// ErrorFileNotFound indicates the requested file does not exist
	ErrorFileNotFound

	// ErrorPathInvalid indicates an invalid or malformed path
	ErrorPathInvalid

	// ErrorPathTraversal indicates a path traversal attempt was detected
	ErrorPathTraversal

	// ErrorPathDotFile indicates an attempt to access a dot file (hidden file)
	ErrorPathDotFile

	// ErrorPathDepth indicates the path depth exceeds the configured maximum
	ErrorPathDepth

	// ErrorPathBlocked indicates the path matches a blocked pattern
	ErrorPathBlocked

	// ErrorMimeTypeDenied indicates the file's MIME type is not allowed
	ErrorMimeTypeDenied
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/static"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorFileInfo:
		return "cannot get file info"
	case ErrorFileOpen:
		return "cannot open file"
	case ErrorFiletemp:
		return "cannot create temporary file"
	case ErrorFileNotFound:
		return "file not found"
	case ErrorPathInvalid:
		return "invalid path"
	case ErrorPathTraversal:
		return "path traversal attempt detected"
	case ErrorPathDotFile:
		return "dot file access not allowed"
	case ErrorPathDepth:
		return "path depth exceeds maximum"
	case ErrorPathBlocked:
		return "path contains blocked pattern"
	case ErrorMimeTypeDenied:
		return "mime type not allowed"
	}

	return liberr.NullMessage
}
