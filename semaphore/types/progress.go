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

import (
	sdkmpb "github.com/vbauerster/mpb/v8"
)

// Progress defines the interface for creating progress bars with different display formats.
// It provides factory methods for common progress bar types (bytes, time, numbers)
// with pre-configured decorators and formatting.
//
// See: github.com/nabbar/golib/semaphore
type Progress interface {
	// BarBytes creates a progress bar for tracking byte quantities (files, downloads, etc.).
	// The bar displays sizes in human-readable format (KB, MB, GB, etc.).
	//
	// Parameters:
	//   - name: Display name for the bar
	//   - job: Job description to display
	//   - tot: Total number of bytes to process
	//   - drop: If true, removes the bar from display when complete
	//   - bar: Optional existing bar to queue after (for sequential display)
	//
	// Returns:
	//   - SemBar: A semaphore with progress bar configured for byte tracking
	BarBytes(name, job string, tot int64, drop bool, bar SemBar) SemBar

	// BarTime creates a progress bar for time-based operations.
	// Displays progress with time estimates (ETA, elapsed time).
	//
	// Parameters:
	//   - name: Display name for the bar
	//   - job: Job description to display
	//   - tot: Total number of time units
	//   - drop: If true, removes the bar from display when complete
	//   - bar: Optional existing bar to queue after
	//
	// Returns:
	//   - SemBar: A semaphore with progress bar configured for time tracking
	BarTime(name, job string, tot int64, drop bool, bar SemBar) SemBar

	// BarNumber creates a progress bar for tracking numeric quantities.
	// Displays progress as a simple counter (e.g., "45/100").
	//
	// Parameters:
	//   - name: Display name for the bar
	//   - job: Job description to display
	//   - tot: Total number of items to process
	//   - drop: If true, removes the bar from display when complete
	//   - bar: Optional existing bar to queue after
	//
	// Returns:
	//   - SemBar: A semaphore with progress bar configured for number tracking
	BarNumber(name, job string, tot int64, drop bool, bar SemBar) SemBar

	// BarOpts creates a progress bar with custom MPB options.
	// Use this for full control over bar appearance and behavior.
	//
	// Parameters:
	//   - tot: Total number of items to process
	//   - drop: If true, removes the bar from display when complete
	//   - opts: MPB bar options for customization
	//
	// Returns:
	//   - SemBar: A semaphore with custom progress bar
	//
	// See: github.com/vbauerster/mpb/v8 for available options
	BarOpts(tot int64, drop bool, opts ...sdkmpb.BarOption) SemBar
}

// ProgressMPB provides access to the underlying MPB (Multi-Progress Bar) container.
// This interface allows access to the progress bar container for advanced use cases.
//
// See: github.com/vbauerster/mpb/v8
type ProgressMPB interface {
	// GetMPB returns the underlying MPB progress container.
	// Returns nil if progress tracking is disabled.
	//
	// Use this for:
	//   - Adding custom progress bars
	//   - Accessing MPB container-level operations
	//   - Coordinating multiple progress bars
	GetMPB() *sdkmpb.Progress
}
