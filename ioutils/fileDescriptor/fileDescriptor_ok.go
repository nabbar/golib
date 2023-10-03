//go:build !windows
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

package fileDescriptor

import (
	"syscall"
)

func systemFileDescriptor(newValue int) (current int, max int, err error) {
	var rLimit syscall.Rlimit

	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return 0, 0, err
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
		if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return 0, 0, err
		}

		return SystemFileDescriptor(0)
	}

	return int(rLimit.Cur), int(rLimit.Max), nil
}
