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

// Package fileDescriptor provides cross-platform utilities for managing file descriptor limits.
// It offers a unified API for querying and modifying the maximum number of open files or I/O
// resources allowed for a process on Unix/Linux, macOS, and Windows systems.
package fileDescriptor

// SystemFileDescriptor returns the current and maximum file descriptor limits for the process.
// It can optionally attempt to increase the limit if newValue is greater than the current limit.
//
// Parameters:
//   - newValue: Desired new limit. Use 0 or negative values to query current limits without modification.
//
// Returns:
//   - current: Current (soft) file descriptor limit
//   - max: Maximum (hard) file descriptor limit
//   - err: Error if the operation fails
//
// Behavior:
//   - Query Mode (newValue <= 0 or newValue <= current): Returns current limits without modification
//   - Increase Mode (newValue > current): Attempts to increase the current limit to newValue
//   - Will not decrease existing limits
//   - May require elevated privileges (root/admin) to increase beyond current soft limit
//   - Respects system hard limits
//
// Platform-Specific Implementation:
//   - Unix/Linux/macOS: Uses syscall.Getrlimit/Setrlimit with RLIMIT_NOFILE
//   - Windows: Uses maxstdio.GetMaxStdio/SetMaxStdio (max 8192)
//
// Example:
//
//	// Query current limits
//	current, max, err := SystemFileDescriptor(0)
//	fmt.Printf("Current: %d, Max: %d\n", current, max)
//
//	// Increase limit to 4096
//	newCurrent, newMax, err := SystemFileDescriptor(4096)
//	if err != nil {
//	    log.Printf("Cannot increase limit: %v", err)
//	}
func SystemFileDescriptor(newValue int) (current int, max int, err error) {
	return systemFileDescriptor(newValue)
}
