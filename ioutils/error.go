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

package ioutils

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgIOUtils
	ErrorSyscallRLimitGet
	ErrorSyscallRLimitSet
	ErrorIOFileStat
	ErrorIOFileSeek
	ErrorIOFileOpen
	ErrorIOFileTempNew
	ErrorIOFileTempClose
	ErrorIOFileTempRemove
	ErrorNilPointer
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/ioutils"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorSyscallRLimitGet:
		return "error on retrieve value in syscall rlimit"
	case ErrorSyscallRLimitSet:
		return "error on changing value in syscall rlimit"
	case ErrorIOFileStat:
		return "error occur while trying to get stat of file"
	case ErrorIOFileSeek:
		return "error occur while trying seek into file"
	case ErrorIOFileOpen:
		return "error occur while trying to open file"
	case ErrorIOFileTempNew:
		return "error occur while trying to create new temporary file"
	case ErrorIOFileTempClose:
		return "closing temporary file occurs error"
	case ErrorIOFileTempRemove:
		return "error occurs on removing temporary file"
	case ErrorNilPointer:
		return "cannot call function for a nil pointer"
	}

	return liberr.NullMessage
}
