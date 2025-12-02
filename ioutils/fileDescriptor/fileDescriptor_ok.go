//go:build !windows

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
	"math"
	"syscall"
)

// systemFileDescriptor is the Unix/Linux/macOS implementation of SystemFileDescriptor.
// It uses syscall.Getrlimit and syscall.Setrlimit with RLIMIT_NOFILE to manage
// file descriptor limits.
//
// Platform-Specific Behavior:
//   - Uses RLIMIT_NOFILE resource to query/set open file limits
//   - Soft limit (Cur): Current limit, can be increased up to hard limit without privileges
//   - Hard limit (Max): Maximum limit, requires root privileges to increase
//   - Returns error if attempting to exceed hard limit without sufficient privileges
//
// The function implements the following logic:
//  1. Query current limits via Getrlimit
//  2. If newValue <= 0 or <= current: return current limits unchanged
//  3. If newValue > current: attempt to increase via Setrlimit
//  4. Handle uint64 to int conversion safely (cap at math.MaxInt)
//
// Thread Safety:
// The syscalls are synchronized at kernel level, making this function naturally thread-safe.
func systemFileDescriptor(newValue int) (current int, max int, err error) {
	var rLimit syscall.Rlimit

	// Query current soft and hard limits from kernel
	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return 0, 0, err
	}

	// Query mode: return current state without modification
	if newValue <= 0 {
		return getCurMax(rLimit.Cur, rLimit.Max, nil)
	} else if uint64(newValue) < rLimit.Cur {
		// Already at or above requested value, no change needed
		return getCurMax(rLimit.Cur, rLimit.Max, nil)
	}

	var chg = false

	// If requesting above hard limit, attempt to increase hard limit first.
	// This requires root privileges on most systems.
	if uint64(newValue) > rLimit.Max {
		chg = true
		rLimit.Max = uint64(newValue)
	}
	// Increase soft limit to requested value (within hard limit range).
	// This typically doesn't require privileges if newValue <= original hard limit.
	if uint64(newValue) > rLimit.Cur {
		chg = true
		rLimit.Cur = uint64(newValue)
	}

	// Apply changes if any limit was modified
	if chg {
		// Setrlimit may fail if:
		// - newValue exceeds hard limit and we're not root
		// - newValue exceeds system-wide maximum (fs.nr_open on Linux)
		if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return 0, 0, err
		}

		// Recursively query to get consistent state after modification
		return SystemFileDescriptor(0)
	}

	// No change needed, return current values
	return getCurMax(rLimit.Cur, rLimit.Max, nil)
}

// getCurMax converts uint64 rlimit values to int, handling potential overflow
// by capping at math.MaxInt.
//
// The kernel returns limits as uint64, but Go applications typically use int
// for counts. On 32-bit systems, uint64 values may exceed int range.
//
// Safety Mechanism:
//   - If value <= math.MaxInt: safe conversion to int
//   - If value > math.MaxInt: cap at math.MaxInt to prevent overflow
//
// This ensures the function never returns invalid negative values due to
// integer overflow, which could cause security issues in calling code.
func getCurMax(rCur, rMax uint64, err error) (int, int, error) {
	var ic, im int
	if rCur <= uint64(math.MaxInt) {
		ic = int(rCur)
	} else {
		ic = math.MaxInt
	}
	if rMax <= uint64(math.MaxInt) {
		im = int(rMax)
	} else {
		im = math.MaxInt
	}

	return ic, im, err

}
