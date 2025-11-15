//go:build windows

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
	"github.com/nabbar/golib/ioutils/maxstdio"
)

const (
	// winDefaultMaxStdio is the default Windows C runtime limit for open files.
	// Reference: https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio
	winDefaultMaxStdio = 512

	// winHardLimitMaxStdio is the maximum limit supported by Windows C runtime.
	winHardLimitMaxStdio = 8192
)

// systemFileDescriptor is the Windows implementation of SystemFileDescriptor.
// It uses maxstdio.GetMaxStdio and maxstdio.SetMaxStdio to manage C runtime
// file descriptor limits (max 8192).
func systemFileDescriptor(newValue int) (current int, max int, err error) {
	rLimit := maxstdio.GetMaxStdio()

	if rLimit < 0 {
		// default windows value
		rLimit = winDefaultMaxStdio
	}

	if newValue > winHardLimitMaxStdio {
		newValue = winHardLimitMaxStdio
	}

	if newValue > rLimit {
		rLimit = int(maxstdio.SetMaxStdio(newValue))
		return SystemFileDescriptor(0)
	}

	return rLimit, winHardLimitMaxStdio, nil
	//	return 0, 0, fmt.Errorf("rLimit is nor implemented in current system")
}
