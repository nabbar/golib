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

// Package semaphore provides a high-level semaphore implementation with integrated progress bar support.
// It combines the functionality of github.com/nabbar/golib/semaphore/sem (base semaphore)
// with github.com/nabbar/golib/semaphore/bar (progress bars) for convenient concurrent worker management
// with optional visual progress tracking.
//
// See also:
//   - github.com/nabbar/golib/semaphore/sem - Base semaphore implementations
//   - github.com/nabbar/golib/semaphore/bar - Progress bar functionality
//   - github.com/nabbar/golib/semaphore/types - Core interfaces
package semaphore

import (
	"context"

	semsem "github.com/nabbar/golib/semaphore/sem"
	semtps "github.com/nabbar/golib/semaphore/types"
	sdkmpb "github.com/vbauerster/mpb/v8"
)

// Semaphore is the main interface combining semaphore functionality with progress bar support.
// It extends context.Context, Sem, and Progress interfaces to provide:
//   - Context lifecycle management (cancellation, deadlines)
//   - Worker concurrency control (acquire/release)
//   - Progress bar creation and management
//
// See: github.com/nabbar/golib/semaphore/types for interface details
type Semaphore interface {
	context.Context // Lifecycle management
	semtps.Sem      // Worker management
	semtps.Progress // Progress bar creation

	// Clone creates a copy of this semaphore with independent worker management
	// but shared MPB progress container (if enabled).
	Clone() Semaphore
}

// MaxSimultaneous returns the maximum number of concurrent goroutines
// based on GOMAXPROCS.
//
// See: github.com/nabbar/golib/semaphore/sem.MaxSimultaneous
func MaxSimultaneous() int {
	return semsem.MaxSimultaneous()
}

// SetSimultaneous calculates the actual simultaneous limit based on the requested value.
//
// Returns:
//   - MaxSimultaneous() if n < 1
//   - MaxSimultaneous() if n > MaxSimultaneous()
//   - n otherwise
//
// See: github.com/nabbar/golib/semaphore/sem.SetSimultaneous
func SetSimultaneous(n int) int64 {
	return semsem.SetSimultaneous(n)
}

// New creates a new Semaphore instance.
//
// Parameters:
//   - ctx: Parent context for lifecycle management
//   - nbrSimultaneous: Number of concurrent workers allowed
//   - == 0: Uses MaxSimultaneous() as the limit
//   - > 0: Uses the provided value as the limit
//   - < 0: No limit (WaitGroup mode)
//   - progress: Enable visual progress bars (MPB)
//   - opt: Optional MPB container options (only used if progress is true)
//
// Returns:
//   - Semaphore: A new semaphore instance
//
// Example:
//
//	sem := semaphore.New(ctx, 10, true)
//	defer sem.DeferMain()
//
//	bar := sem.BarNumber("Tasks", "processing", 100, false, nil)
//	// Use bar for worker management with progress tracking
func New(ctx context.Context, nbrSimultaneous int, progress bool, opt ...sdkmpb.ContainerOption) Semaphore {
	var (
		m *sdkmpb.Progress
	)

	if progress {
		m = sdkmpb.New(opt...)
	}

	return &sem{
		s: semsem.New(ctx, nbrSimultaneous),
		m: m,
	}
}
