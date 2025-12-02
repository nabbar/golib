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
	//
	// Windows applications start with 512 file handles available by default.
	// This is a Microsoft C runtime (CRT) limitation, not an OS limitation.
	//
	// Reference: https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio
	winDefaultMaxStdio = 512

	// winHardLimitMaxStdio is the maximum limit supported by Windows C runtime.
	//
	// This is an absolute hard limit imposed by the Windows C runtime library.
	// Values above 8192 will be automatically capped to this limit.
	// Unlike Unix systems, this limit cannot be increased with administrator privileges.
	//
	// Note: This applies to C stdio functions and handles. Native Windows API
	// calls may have different limits, but Go's standard library uses CRT functions.
	winHardLimitMaxStdio = 8192
)

// systemFileDescriptor is the Windows implementation of SystemFileDescriptor.
// It uses maxstdio.GetMaxStdio and maxstdio.SetMaxStdio to manage C runtime
// file descriptor limits (max 8192).
//
// Platform-Specific Behavior:
//   - Uses Windows C runtime _getmaxstdio and _setmaxstdio functions
//   - Default limit: 512 file descriptors
//   - Maximum limit: 8192 file descriptors (hard cap, cannot exceed)
//   - No privileges required: Any process can increase limit up to 8192
//   - Automatic capping: Values > 8192 are silently capped to 8192
//
// The function implements the following logic:
//  1. Query current limit via GetMaxStdio (returns -1 if unavailable)
//  2. Use default (512) if GetMaxStdio fails
//  3. Cap newValue at 8192 (Windows hard limit)
//  4. If newValue > current: attempt to increase via SetMaxStdio
//
// Thread Safety:
// The C runtime functions are thread-safe, making this function safe for concurrent use.
//
// Differences from Unix:
//   - Much lower maximum limit (8192 vs potentially unlimited on Unix)
//   - No separate soft/hard limit concept
//   - No privilege requirements for increases
//   - Cannot query actual file descriptor usage, only limit
func systemFileDescriptor(newValue int) (current int, max int, err error) {
	// Query current limit from Windows C runtime
	rLimit := maxstdio.GetMaxStdio()

	if rLimit < 0 {
		// GetMaxStdio returns -1 if the query fails (rare on modern Windows).
		// Fallback to the documented default Windows C runtime value.
		rLimit = winDefaultMaxStdio
	}

	// Cap requested value at Windows hard limit (8192).
	// Unlike Unix, Windows has a fixed maximum that cannot be exceeded.
	if newValue > winHardLimitMaxStdio {
		newValue = winHardLimitMaxStdio
	}

	// Only attempt to increase if requested value is higher than current.
	// This prevents unnecessary SetMaxStdio calls and potential errors.
	if newValue > rLimit {
		// SetMaxStdio returns the new limit (may be less than requested if capped).
		rLimit = int(maxstdio.SetMaxStdio(newValue))
		// Recursively query to get consistent state after modification.
		return SystemFileDescriptor(0)
	}

	// Return current limit and Windows hard maximum.
	// On Windows, max is always 8192 regardless of current setting.
	return rLimit, winHardLimitMaxStdio, nil
}
