/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package ioprogress

import (
	"errors"
	"io"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libfpg "github.com/nabbar/golib/file/progress"
)

// rdr implements the Reader interface by wrapping an io.ReadCloser with progress
// tracking capabilities using atomic operations for lock-free thread safety.
//
// Architecture: This structure uses composition to extend io.ReadCloser functionality
// while maintaining full compatibility with the standard io interfaces. All state
// mutations use atomic primitives to ensure thread-safe concurrent access without mutexes.
//
// Memory Layout:
//   - r: 16 bytes (interface: pointer + type)
//   - cr: 8 bytes (pointer to atomic.Int64)
//   - fi, fe, fr: ~24 bytes each (atomic.Value with function pointer)
//     Total: ~120 bytes per instance
//
// Thread Safety: All fields except 'r' use atomic operations. The underlying reader 'r'
// is accessed only through the Read() and Close() methods, and concurrent reads on the
// same reader should be managed by the caller according to io.Reader semantics.
type rdr struct {
	r  io.ReadCloser                     // underlying reader (caller manages concurrency)
	cr *atomic.Int64                     // cumulative byte counter (atomic updates)
	fi libatm.Value[libfpg.FctIncrement] // increment callback (atomic load/store)
	fe libatm.Value[libfpg.FctEOF]       // EOF callback (atomic load/store)
	fr libatm.Value[libfpg.FctReset]     // reset callback (atomic load/store)
}

// Read implements io.Reader by delegating to the underlying reader and tracking progress.
//
// This method performs the following operations in order:
//  1. Delegate the read operation to the underlying reader
//  2. Update the cumulative byte counter atomically (even if err != nil)
//  3. Invoke the increment callback with bytes read
//  4. If EOF is detected, invoke the EOF callback
//  5. Return the original (n, err) from the underlying reader
//
// Behavior:
//   - The increment callback is invoked for every Read() call, even when n=0 or err!=nil
//   - EOF detection uses errors.Is() to handle wrapped errors correctly
//   - The EOF callback is invoked only once per EOF encounter
//   - All state updates use atomic operations for thread-safe concurrent reads
//
// Thread Safety: Multiple goroutines can call Read() concurrently if the underlying
// reader supports it. However, most io.Reader implementations are not safe for
// concurrent reads, so callers should serialize Read() calls unless documented otherwise.
//
// Performance: Adds ~50-100ns overhead per call due to atomic operations and callback
// invocation. Zero allocations in normal operation.
func (r *rdr) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.inc(n) // Always track progress, even on error

	if errors.Is(err, io.EOF) {
		r.finish() // Invoke EOF callback
	}

	return n, err
}

// Close implements io.Closer by closing the underlying reader.
//
// This method propagates the close operation to the wrapped reader without
// modifying any progress tracking state. The cumulative counter and registered
// callbacks remain accessible after closing, though no further I/O should occur.
//
// Thread Safety: Safe to call concurrently, though behavior depends on the
// underlying reader's Close() implementation. Most readers are safe to close
// multiple times (idempotent).
//
// Typical Usage: Always defer Close() immediately after creating the wrapper:
//
//	reader := ioprogress.NewReadCloser(file)
//	defer reader.Close()  // Closes both wrapper and underlying file
func (r *rdr) Close() error {
	return r.r.Close()
}

// RegisterFctIncrement implements Progress by storing the increment callback.
//
// This method registers a callback that will be invoked after every Read() operation
// with the number of bytes transferred in that operation. The callback executes
// synchronously in the same goroutine as Read().
//
// Nil Handling: If fct is nil, it is converted to a no-op function before storage.
// This prevents atomic.Value.Store(nil) panic, as atomic.Value cannot store nil.
// This approach ensures the atomic field always contains a valid function pointer,
// eliminating the need for nil checks during Load() operations.
//
// Thread Safety: This method uses atomic.Value.Store() for lock-free concurrent access.
// It can be called safely from multiple goroutines, including during ongoing Read()
// operations. The new callback replaces any previously registered callback atomically.
//
// Performance: The callback must complete quickly (<1ms recommended) to avoid degrading
// I/O throughput. Use atomic operations for counter updates to maintain thread safety.
func (r *rdr) RegisterFctIncrement(fct libfpg.FctIncrement) {
	if fct == nil {
		// Convert nil to no-op to prevent atomic.Value.Store(nil) panic.
		// This is more robust than checking for nil on every Load().
		fct = func(size int64) {}
	}

	r.fi.Store(fct)
}

// RegisterFctReset implements Progress by storing the reset callback.
//
// This method registers a callback that will be invoked when Reset() is called,
// receiving the maximum expected size and current cumulative progress. Useful for
// multi-stage operations where progress needs to be reported relative to different
// total sizes or processing phases.
//
// Nil Handling: If fct is nil, it is converted to a no-op function before storage.
// This prevents atomic.Value.Store(nil) panic, ensuring robust operation without
// runtime nil pointer checks.
//
// Thread Safety: This method uses atomic.Value.Store() for lock-free concurrent access.
// Safe to call from multiple goroutines concurrently.
func (r *rdr) RegisterFctReset(fct libfpg.FctReset) {
	if fct == nil {
		// Convert nil to no-op to prevent atomic.Value.Store(nil) panic.
		fct = func(size, current int64) {}
	}

	r.fr.Store(fct)
}

// RegisterFctEOF implements Progress by storing the EOF callback.
//
// This method registers a callback that will be invoked when io.EOF is detected
// during a Read() operation, indicating that all data has been consumed from the
// underlying reader. The callback executes immediately after EOF detection.
//
// Nil Handling: If fct is nil, it is converted to a no-op function before storage.
// This prevents atomic.Value.Store(nil) panic, maintaining robust operation.
//
// Thread Safety: This method uses atomic.Value.Store() for lock-free concurrent access.
// Safe to call from multiple goroutines concurrently.
func (r *rdr) RegisterFctEOF(fct libfpg.FctEOF) {
	if fct == nil {
		// Convert nil to no-op to prevent atomic.Value.Store(nil) panic.
		fct = func() {}
	}

	r.fe.Store(fct)
}

// inc atomically increments the cumulative byte counter and invokes the increment callback.
//
// This internal method is called after every Read() operation to update progress tracking.
// It performs two atomic operations:
//  1. Add n to the cumulative counter using atomic.Int64.Add()
//  2. Load and invoke the increment callback using atomic.Value.Load()
//
// Behavior:
//   - Always invoked by Read(), even when n=0 or an error occurred
//   - The callback receives the incremental bytes (n), not the cumulative total
//   - Nil check on r is defensive programming for safety, though should never be nil
//   - No nil check on f because RegisterFct* methods ensure it's never nil
//
// Thread Safety: Uses atomic operations exclusively, safe for concurrent access.
//
// Performance: ~30-50ns overhead (atomic add + load + function call).
func (r *rdr) inc(n int) {
	if r == nil {
		// Defensive nil check, should never happen in normal operation
		return
	}

	r.cr.Add(int64(n)) // Atomic increment of cumulative counter

	f := r.fi.Load() // Atomic load of callback function
	if f != nil {
		f(int64(n)) // Invoke with incremental size, not cumulative
	}
}

// finish invokes the EOF callback when EOF is detected.
//
// This internal method is called by Read() when errors.Is(err, io.EOF) returns true,
// indicating that the underlying reader has reached end-of-file. It provides a
// notification mechanism for completion detection.
//
// Behavior:
//   - Called automatically by Read() upon EOF detection
//   - May be invoked multiple times if Read() is called repeatedly after EOF
//   - Uses atomic.Value.Load() to retrieve the callback safely
//   - No nil check on f because RegisterFctEOF ensures it's never nil
//
// Thread Safety: Safe for concurrent access through atomic operations.
func (r *rdr) finish() {
	if r == nil {
		// Defensive nil check
		return
	}

	f := r.fe.Load() // Atomic load of EOF callback
	if f != nil {
		f() // Invoke EOF notification
	}
}

// Reset implements Progress by invoking the reset callback with the maximum size
// and current progress.
//
// This method provides a way to report progress relative to a known total size or
// to signal stage transitions in multi-phase operations. It does NOT reset the
// cumulative counter; it only invokes the callback with current values.
//
// Parameters:
//   - max: The maximum expected size in bytes, or 0 if unknown
//
// Behavior:
//   - Loads the current cumulative count atomically
//   - Invokes the reset callback with (max, current)
//   - The cumulative counter remains unchanged
//   - Useful for updating progress bars or reporting stage completion
//
// Thread Safety: Safe for concurrent access. Uses atomic operations for both
// callback retrieval (atomic.Value.Load) and counter reading (atomic.Int64.Load).
//
// Example Use Case:
//
//	reader.Reset(fileSize)  // Update progress bar with total size
//	validateData(reader)    // First pass
//	reader.Reset(fileSize)  // Update for second pass
//	processData(reader)     // Second pass
func (r *rdr) Reset(max int64) {
	if r == nil {
		// Defensive nil check
		return
	}

	f := r.fr.Load() // Atomic load of reset callback
	if f != nil {
		f(max, r.cr.Load()) // Invoke with max and current cumulative count
	}
}
