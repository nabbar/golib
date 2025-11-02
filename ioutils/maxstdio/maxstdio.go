//go:build windows && cgo
// +build windows,cgo

/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

// Package maxstdio provides Windows-specific functions for managing the maximum number
// of simultaneously open file handles (stdio streams) in a process.
//
// This package uses CGO to interface with the Windows C Runtime (CRT) functions
// _getmaxstdio() and _setmaxstdio(), allowing Go applications to query and modify
// the process-level limit for open file descriptors.
//
// Platform Requirements:
//   - Windows operating system only
//   - CGO enabled (CGO_ENABLED=1)
//   - C compiler (MinGW, MSVC, or TDM-GCC)
//
// The package will not compile on non-Windows platforms due to build constraints.
package maxstdio

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "maxstdio.h"
import "C"

// GetMaxStdio returns the current maximum number of simultaneously open file handles
// allowed for the process.
//
// This function wraps the Windows CRT _getmaxstdio() function, which retrieves
// the process-level limit for file descriptors. The default limit on Windows is
// typically 512, but this can be modified using SetMaxStdio().
//
// Returns:
//   - int: The current maximum number of open files
//
// The returned value represents the hard limit that the process cannot exceed.
// Attempting to open more files than this limit will result in errors.
//
// Example:
//
//	current := maxstdio.GetMaxStdio()
//	fmt.Printf("Current limit: %d files\n", current)
//
// See: https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/getmaxstdio
func GetMaxStdio() int {
	return int(C.CGetMaxSTDIO())
}

// SetMaxStdio sets the maximum number of simultaneously open file handles for the process.
//
// This function wraps the Windows CRT _setmaxstdio() function, which modifies the
// process-level limit for file descriptors. The actual limit may be constrained by
// system resources and Windows configuration.
//
// Parameters:
//   - newMax: The desired maximum number of open files (typically 512-8192)
//
// Returns:
//   - int: The previous maximum value before the change
//
// The function always returns the old limit, even if the new limit couldn't be
// fully applied due to system constraints. After calling SetMaxStdio(), use
// GetMaxStdio() to verify the actual limit that was set.
//
// Typical usage:
//   - Web servers handling many connections: 2048-4096
//   - Build systems processing many files: 2048-8192
//   - Default applications: 512 (Windows default)
//
// Example:
//
//	old := maxstdio.SetMaxStdio(2048)
//	fmt.Printf("Limit changed from %d to %d\n", old, maxstdio.GetMaxStdio())
//
// Note: Some systems may have hard limits that cannot be exceeded without
// administrator privileges or system reconfiguration.
//
// See: https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio
func SetMaxStdio(newMax int) int {
	return int(C.CSetMaxSTDIO(C.int(newMax)))
}
