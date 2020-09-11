// +build !windows

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

package ioutils

import (
	"syscall"

	. "github.com/nabbar/golib/errors"
)

func systemFileDescriptor(newValue int) (current int, max int, err Error) {
	var (
		rLimit syscall.Rlimit
		e      error
	)

	if e = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); e != nil {
		err = ErrorSyscallRLimitGet.ErrorParent(e)
		return
	}

	if newValue < 1 {
		return int(rLimit.Cur), int(rLimit.Max), nil
	}

	if newValue < int(rLimit.Cur) {
		return int(rLimit.Cur), int(rLimit.Max), nil
	}

	var chg = false

	if newValue > int(rLimit.Max) {
		chg = true
		rLimit.Max = uint64(newValue)
	}
	if newValue > int(rLimit.Cur) {
		chg = true
		rLimit.Cur = uint64(newValue)
	}

	if chg {
		if e = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); e != nil {
			err = ErrorSyscallRLimitSet.ErrorParent(e)
			return
		}

		return SystemFileDescriptor(0)
	}

	return int(rLimit.Cur), int(rLimit.Max), nil
}
