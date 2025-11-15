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

// Package sem provides semaphore implementations for controlling concurrent goroutine execution.
// It offers two strategies: weighted semaphores (with limits) and WaitGroup-based (unlimited).
package sem

import (
	"context"
	"runtime"
	"sync"

	semtps "github.com/nabbar/golib/semaphore/types"
	goxsem "golang.org/x/sync/semaphore"
)

// MaxSimultaneous returns the maximum number of concurrent goroutines
// based on GOMAXPROCS.
func MaxSimultaneous() int {
	return runtime.GOMAXPROCS(0)
}

// SetSimultaneous calculates the actual simultaneous limit based on the requested value.
// Returns:
//   - MaxSimultaneous() if n < 1
//   - MaxSimultaneous() if n > MaxSimultaneous()
//   - n otherwise
func SetSimultaneous(n int) int64 {
	m := MaxSimultaneous()
	if n < 1 {
		return int64(m)
	} else if m < n {
		return int64(m)
	} else {
		return int64(n)
	}
}

// New creates a new semaphore instance.
//
// The behavior depends on nbrSimultaneous:
//   - == 0: Uses MaxSimultaneous() as the limit (weighted semaphore)
//   - > 0: Uses the provided value as the limit (weighted semaphore)
//   - < 0: No limit, uses WaitGroup instead of semaphore
//
// Parameters:
//   - ctx: Parent context for cancellation
//   - nbrSimultaneous: Number of concurrent workers allowed
//
// Returns:
//   - Sem: A semaphore instance implementing the Sem interface
func New(ctx context.Context, nbrSimultaneous int) semtps.Sem {
	var (
		x context.Context
		n context.CancelFunc
	)

	x, n = context.WithCancel(ctx)

	if nbrSimultaneous == 0 {
		b := SetSimultaneous(nbrSimultaneous)
		return &sem{
			c: n,
			x: x,
			s: goxsem.NewWeighted(b),
			n: b,
		}
	} else if nbrSimultaneous > 0 {
		return &sem{
			c: n,
			x: x,
			s: goxsem.NewWeighted(int64(nbrSimultaneous)),
			n: int64(nbrSimultaneous),
		}
	} else {
		return &wg{
			c: n,
			x: x,
			w: sync.WaitGroup{},
		}
	}
}
