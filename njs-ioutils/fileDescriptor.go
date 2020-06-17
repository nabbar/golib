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

package njs_ioutils

import (
	"syscall"
)

/**
 * SystemFileDescriptor is returning current Limit & max system limit for file descriptor (open file or I/O resource) currently set in the system
 * This function return the current setting (current number of file descriptor and the max value) if the newValue given is zero
 * Otherwise if the newValue is more than the current system limit, try to change the current limit in the system for this process only
 */
func SystemFileDescriptor(newValue int) (current int, max int, err error) {
	var rLimit syscall.Rlimit

	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
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
		if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return
		}

		return SystemFileDescriptor(0)
	}

	return int(rLimit.Cur), int(rLimit.Max), nil
}
