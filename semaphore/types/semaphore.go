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

// Package types defines the core interfaces for semaphore implementations with progress tracking.
// These interfaces provide the foundation for concurrent worker management with optional visual progress bars.
//
// See the implementation packages:
//   - github.com/nabbar/golib/semaphore/sem - Base semaphore implementations
//   - github.com/nabbar/golib/semaphore/bar - Progress bar-enabled semaphores
//   - github.com/nabbar/golib/semaphore - Combined semaphore with progress tracking
package types

import "context"

// SemPgb combines semaphore functionality with MPB (Multi-Progress Bar) progress tracking.
// It extends the base Sem interface with access to the underlying MPB progress container.
//
// Use this interface when you need both worker management and progress visualization.
// See: github.com/nabbar/golib/semaphore
type SemPgb interface {
	Sem         // Core semaphore functionality
	ProgressMPB // Access to MPB progress container
}

// SemBar combines semaphore functionality with progress bar operations.
// It extends the base Sem interface with Bar methods for tracking progress.
//
// Use this interface when you need worker management with progress bar updates.
// See: github.com/nabbar/golib/semaphore/bar
type SemBar interface {
	Sem // Core semaphore functionality
	Bar // Progress bar operations
}

// Sem defines the core semaphore interface for managing concurrent goroutine execution.
// It implements context.Context for lifecycle management and provides worker slot acquisition/release.
//
// The semaphore can operate in two modes:
//   - Limited: Fixed number of concurrent workers (weighted semaphore)
//   - Unlimited: No concurrency limit (WaitGroup-based)
//
// See: github.com/nabbar/golib/semaphore/sem
type Sem interface {
	context.Context // Inherits context functionality for cancellation and deadlines

	// NewWorker acquires a worker slot, blocking until one is available or context is cancelled.
	// Returns an error if the context is cancelled before acquiring a slot.
	//
	// Usage:
	//   if err := sem.NewWorker(); err != nil {
	//       return err
	//   }
	//   defer sem.DeferWorker()
	NewWorker() error

	// NewWorkerTry attempts to acquire a worker slot without blocking.
	// Returns true if successful, false if no slots are available.
	//
	// Usage:
	//   if sem.NewWorkerTry() {
	//       defer sem.DeferWorker()
	//       // process work
	//   }
	NewWorkerTry() bool

	// DeferWorker releases a previously acquired worker slot.
	// Should be called with defer immediately after NewWorker() or NewWorkerTry() succeeds.
	DeferWorker()

	// DeferMain cancels the semaphore's context and releases all resources.
	// Should be called with defer in the main goroutine that created the semaphore.
	DeferMain()

	// WaitAll blocks until all worker slots are available (all workers have completed).
	// Returns an error if the context is cancelled while waiting.
	//
	// Use this to wait for all spawned workers to complete before proceeding.
	WaitAll() error

	// Weighted returns the maximum number of concurrent workers allowed.
	// Returns -1 for unlimited concurrency (WaitGroup mode).
	// Returns the configured limit for weighted semaphore mode.
	Weighted() int64

	// New creates a new independent semaphore with the same concurrency limit.
	// The new semaphore inherits the parent's context but operates independently.
	New() Sem
}
