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

package njs_ioutils

import errors "github.com/nabbar/golib/njs-errors"

const (
	EMPTY_PARAMS errors.CodeError = iota + errors.MIN_PKG_IOUtils
	SYSCALL_RLIMIT_GET
	SYSCALL_RLIMIT_SET
	IO_TEMP_FILE_NEW
	IO_TEMP_FILE_CLOSE
	IO_TEMP_FILE_REMOVE
)

func init() {
	errors.RegisterFctMessage(getMessage)
}

func getMessage(code errors.CodeError) (message string) {
	switch code {
	case EMPTY_PARAMS:
		return "given parameters is empty"
	case SYSCALL_RLIMIT_GET:
		return "error on retrieve value in syscall rlimit"
	case SYSCALL_RLIMIT_SET:
		return "error on changing value in syscall rlimit"
	case IO_TEMP_FILE_NEW:
		return "error occur while trying to create new temporary file"
	case IO_TEMP_FILE_CLOSE:
		return "closing temporary file occurs error"
	case IO_TEMP_FILE_REMOVE:
		return "error occurs on removing temporary file"
	}

	return ""
}
