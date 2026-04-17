/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package aggregator

import (
	"context"

	"github.com/nabbar/golib/runner"
)

// Close shuts down the aggregator and ensures all internal resources are released.
//
// Operational Lifecycle:
//  1. Graceful Shutdown: Invokes Stop() with a background context to terminate the
//     processing goroutine, timers, and any active callbacks.
//  2. Resource Releasing: The underlying closeRun hook is triggered by the runner,
//     which calls cleanup() to deactivate the data channel and contexts.
//
// Thread-Safety and Idempotency:
// Multiple calls to Close() are safe. It handles re-entry through the runner's
// lifecycle state machine, ensuring only one shutdown sequence is executed.
//
// Implements: io.Closer
func (o *agg) Close() error {
	defer func() {
		// Recovery from panics during close.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/close", r)
		}
	}()

	// Delegate the shutdown process to the runner with a fresh background context.
	return o.Stop(context.Background())
}

// closeRun serves as the standardized internal shutdown entry point for the runner.
// It performs resource cleanup (context and channel) ensuring no deadlocks occur
// during the finalization phase.
//
// Parameters:
//   - ctx: Context provided by the runner for shutdown coordination.
func (o *agg) closeRun(ctx context.Context) error {
	defer func() {
		// Recovery from panics during internal runner shutdown.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/closeRun", r)
		}
	}()

	// Decommission operational contexts and data channels.
	o.cleanup()

	return nil
}

// Write queues data to the internal aggregator buffer for asynchronous serialization.
//
// Technical Implementation Details:
//  1. Memory Recycling: Retrieves a pre-allocated byte slice pointer (*[]byte) from
//     the internal sync.Pool. If the buffer is undersized, it reallocates;
//     otherwise, it reuses the capacity to avoid heap thrashing and GC pressure.
//  2. High-Performance Counters: It increments the "waiting" counters (cw, sw) before
//     attempting the write, providing visibility into producer-side congestion
//     (backpressure) during buffer saturation.
//  3. Atomic Pipeline Validation: Resolves the current data channel through an atomic
//     load. If the channel is decommissioned or the aggregator is stopped, the
//     write is rejected immediately with ErrClosedResources.
//  4. Safe Termination: The blocking send respects the aggregator's context
//     cancellation, ensuring no producer goroutine remains hung during shutdown.
//
// Efficiency Note:
// Using pointers to byte slices (*[]byte) prevents 'convTslice' allocations when
// sending data through channels, which is critical for achieving high throughput.
//
// Parameters:
//   - p: Byte slice to be aggregated. Empty slices (len == 0) are ignored with 0/nil.
//
// Returns:
//   - n: Number of bytes successfully queued (always len(p) on success).
//   - err: ErrClosedResources if the aggregator is not active, context error
//     if the operation is interrupted, or ErrInvalidInstance on corruption.
//
// Implements: io.Writer
func (o *agg) Write(p []byte) (n int, err error) {
	defer func() {
		// Recovery from panics during write operations to prevent crashing producers.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/write", r)
		}
	}()

	// Ignore empty writes to optimize the pipeline and avoid useless allocations or counter churn.
	n = len(p)
	if n == 0 {
		return 0, nil
	}

	// Backpressure Telemetry: Signal that this producer is attempting a write.
	// We decrement the waiting counters once the write either succeeds or is cancelled.
	defer o.cntWaitDec(n)
	o.cntWaitInc(n)

	// Validate current operational state using a fast atomic load of the 'op' flag.
	if !o.op.Load() {
		return 0, ErrClosedResources
	}

	// Resolve the current data channel through the atomic value container.
	c := o.ch.Load()
	if c == nil || c == closedChan {
		// The data pipeline is decommissioned; reject all writes.
		return 0, ErrClosedResources
	}

	// Check for any ongoing cancellation errors in the aggregator's context before allocating memory.
	if err = o.Err(); err != nil {
		return 0, err
	}

	// 1. Efficient Buffer Allocation Strategy:
	// Obtain a recycled buffer from the pool and copy the input data into it.
	// This ensures that the caller can safely modify their 'p' slice after Write returns.
	pCpy := o.getBuffer(n)
	copy(*pCpy, p)

	// 2. Data Ingestion:
	select {
	case c <- pCpy:
		// Data successfully accepted into the internal channel.
		// Update processing telemetry counters.
		o.cntDataInc(n)
		return n, nil
	case <-o.Done():
		// The aggregator is being stopped or its context was cancelled.
		// Return the allocated buffer to the pool to prevent a memory leak.
		o.bp.Put(pCpy)
		return 0, o.Err()
	}
}

// getBuffer manages the high-performance sync.Pool for pointers to byte slices (*[]byte).
//
// Efficiency Strategy:
//  1. Pooling: Attempts to reuse an existing pointer and its underlying byte slice array.
//  2. Resizing: If the reused slice's capacity is smaller than 'n', it discards it
//     and allocates a new appropriately sized buffer. This balance minimizes heap
//     allocations while avoiding memory fragmentation or huge buffers for small data.
//  3. Pointer Usage: Using pointers (*[]byte) is critical to prevent the Go runtime
//     from performing 'convTslice' conversions when data crosses interface or
//     channel boundaries, which would cause heap allocations and GC pressure.
//
// Parameters:
//   - n: The required minimum capacity (length) for the data buffer.
//
// Returns:
//   - *[]byte: A pointer to a byte slice with at least length 'n'.
func (o *agg) getBuffer(n int) *[]byte {
	var pCpy *[]byte

	// Attempt to retrieve a recycled pointer from the pool.
	if v := o.bp.Get(); v != nil {
		pCpy = v.(*[]byte)
	}

	// Fallback to fresh allocation if the pool is empty or the retrieved buffer is nil.
	if pCpy == nil {
		buf := make([]byte, n)
		return &buf
	}

	// Resizing Logic:
	// If the underlying array's capacity is insufficient, we must allocate a new one.
	if cap(*pCpy) < n {
		buf := make([]byte, n)
		return &buf
	}

	// Re-slice the pooled buffer to the exact requested length; the underlying array is reused.
	*pCpy = (*pCpy)[:n]
	return pCpy
}

// chanData retrieves the current operational data channel from atomic storage.
// It returns the 'closedChan' sentinel if the aggregator's pipeline is not yet initialized.
func (o *agg) chanData() <-chan *[]byte {
	if c := o.ch.Load(); c == nil {
		return closedChan
	} else {
		return c
	}
}

// chanOpen initializes the data channel with the configured capacity and activates the 'op' flag.
// This is executed during the Start phase to ensure a fresh, empty buffer for processing.
func (o *agg) chanOpen() {
	o.op.Store(true)
	o.ch.Store(make(chan *[]byte, o.sh))
}

// chanClose deactivates the 'op' flag and replaces the data channel with the pre-closed sentinel.
// This ensures that all subsequent Write() calls detect the closure atomically and safely
// without encountering panics from writing to a nil or closed channel.
func (o *agg) chanClose() {
	o.op.Store(false)
	o.ch.Store(closedChan)
}
