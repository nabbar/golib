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
	"time"

	"github.com/nabbar/golib/runner"
)

// Close stops the aggregator and releases all resources.
//
// This method:
//  1. Stops the processing goroutine if running
//  2. Cancels the internal context
//  3. Closes the write channel
//
// After Close is called, any subsequent Write operations will return
// ErrClosedResources. Close is idempotent and can be called multiple times safely.
//
// Close implements io.Closer and should typically be called with defer:
//
//	agg, _ := aggregator.New(ctx, cfg, logger)
//	agg.Start(ctx)
//	defer agg.Close()  // Ensures cleanup on function exit
//
// The method blocks for up to 100ms waiting for the aggregator to stop gracefully.
func (o *agg) Close() error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/close", r)
		}
	}()

	var e error

	if o.IsRunning() {
		x, n := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer n()
		e = o.Stop(x)
	}

	o.cleanup()

	return e
}

// closeRun is the internal close function called by the runner.
// It stops the aggregator, closes the context, and closes the channel.
func (o *agg) closeRun(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/closeRun", r)
		}
	}()

	o.cleanup()

	return nil
}

// Write queues data to be written by the aggregator.
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
// The data is buffered in an internal channel and processed sequentially by the
// aggregator's processing goroutine.
//
// Parameters:
//   - p: Byte slice to write. Empty slices (len == 0) are ignored and return (0, nil).
//
// Returns:
//   - n: Number of bytes queued (always len(p) if no error)
//   - err: Error if write failed:
//   - ErrClosedResources: if aggregator is not running or has been closed
//   - ErrInvalidInstance: if aggregator's internal state is corrupted
//   - Context error: if the aggregator's context has been cancelled
//
// Write implements io.Writer. The write is non-blocking as long as the internal
// buffer (Config.BufWriter) is not full. If the buffer is full, Write blocks until
// space becomes available or the context is cancelled.
//
// Example:
//
//	n, err := agg.Write([]byte("data from goroutine 1"))
//	if err != nil {
//	    log.Printf("write failed: %v", err)
//	}
//
// Note: The aggregator must be started with Start() before calling Write.
func (o *agg) Write(p []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/write", r)
		}
	}()

	// Don't send empty data to channel
	n = len(p)
	if n == 0 {
		return 0, nil
	}

	// Track this write as waiting (will block if channel is full)
	o.cntWaitInc(n)
	defer o.cntWaitDec(n)

	// Check if channel is open
	if !o.op.Load() {
		return 0, ErrClosedResources
	} else if c := o.ch.Load(); c == nil {
		return 0, ErrInvalidInstance
	} else if c == closedChan {
		return 0, ErrClosedResources
	} else if o.Err() != nil {
		return 0, o.Err()
	} else {
		// Increment processing counter before sending to channel
		o.cntDataInc(n)

		// Send to channel (may block if buffer is full)
		// using new slice to prevent reset params slice p
		pCpy := make([]byte, n)
		copy(pCpy, p)

		c <- pCpy
		return len(p), nil
	}
}

// chanData returns the read-only channel for consuming write data.
// This is used internally by the processing goroutine in the run() loop.
// Returns closedChan sentinel if the channel is not initialized or has been closed.
func (o *agg) chanData() <-chan []byte {
	if c := o.ch.Load(); c == nil {
		return closedChan
	} else if c == closedChan {
		return closedChan
	} else {
		return c
	}
}

// chanOpen creates a new buffered channel for writes and marks it as open.
// This is called by run() when the aggregator starts, after verifying the
// aggregator is not already running. The channel capacity is determined by
// Config.BufWriter (stored in sh). The op flag is set to true atomically
// to signal that the channel is ready for writes.
func (o *agg) chanOpen() {
	// Mark channel as closing to prevent new writes
	o.op.Store(true)
	o.ch.Store(make(chan []byte, o.sh))
}

// chanClose marks the channel as closed and replaces it with closedChan sentinel.
// This prevents new writes and signals to readers that the channel is closed.
// The actual channel is not closed to avoid panics from concurrent writes;
// instead we use a pre-closed sentinel channel. The op flag is set to false
// atomically to signal that writes should be rejected.
func (o *agg) chanClose() {
	// Mark channel as closing to prevent new writes
	o.op.Store(false)
	o.ch.Store(closedChan)
}
