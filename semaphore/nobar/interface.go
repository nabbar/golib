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

// Package bar provides a semaphore wrapper with integrated progress bar support.
// It combines semaphore functionality with visual progress tracking using the MPB library.
package nobar

import (
	sdkmpb "github.com/vbauerster/mpb/v8"

	semtps "github.com/nabbar/golib/semaphore/types"
)

// New creates a new progress bar-enabled semaphore.
//
// Parameters:
//   - sem: The parent semaphore with progress support
//   - tot: Total number of items/tasks to track
//   - drop: If true, removes the bar from display when complete
//   - opts: Additional MPB bar options
//
// Returns:
//   - SemBar: A semaphore with integrated progress bar
//
// The returned SemBar implements both semaphore and progress bar interfaces,
// allowing concurrent worker management with visual progress tracking.
func New(sem semtps.SemPgb, _ int64, _ bool, _ ...sdkmpb.BarOption) semtps.SemBar {
	return &bar{
		s: sem.New(),
	}
}
