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
 *
 */

package types

import sdkmpb "github.com/vbauerster/mpb/v8"

// Bar defines the interface for progress bar operations.
// It provides methods for updating, querying, and controlling a progress bar's state.
//
// The Bar interface is typically combined with Sem to create SemBar,
// allowing progress tracking alongside worker management.
//
// See: github.com/nabbar/golib/semaphore/bar
type Bar interface {
	// Inc increments the progress bar by n.
	// This is a convenience method that calls Inc64 with int64(n).
	Inc(n int)

	// Dec decrements the progress bar by n.
	// This is a convenience method that calls Dec64 with int64(n).
	Dec(n int)

	// Inc64 increments the progress bar by n (64-bit version).
	// Use this for large progress values or when n is already int64.
	Inc64(n int64)

	// Dec64 decrements the progress bar by n (64-bit version).
	// Implemented internally as Inc64(-n).
	Dec64(n int64)

	// Reset resets the progress bar to new total and current values.
	// This allows reusing the same bar for different operations.
	//
	// Parameters:
	//   - tot: New total/maximum value
	//   - current: New current progress value
	Reset(tot, current int64)

	// Complete marks the progress bar as complete.
	// If MPB is enabled, this triggers the completion animation.
	Complete()

	// Completed returns true if the progress bar has completed or was aborted.
	// Without MPB, this always returns true.
	Completed() bool

	// Current returns the current progress value.
	// Without MPB, returns the total value.
	Current() int64

	// Total returns the total/maximum value of the progress bar.
	Total() int64
}

// BarMPB provides access to the underlying MPB (Multi-Progress Bar) instance.
// This interface allows low-level manipulation of the progress bar when needed.
//
// See: github.com/vbauerster/mpb/v8
type BarMPB interface {
	// GetMPB returns the underlying MPB bar instance.
	// Returns nil if progress bar is disabled.
	//
	// Use this for advanced MPB operations not covered by the Bar interface.
	GetMPB() *sdkmpb.Bar
}
