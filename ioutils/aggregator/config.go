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
)

// Config defines the configuration for creating a new Aggregator.
//
// The configuration allows customization of buffering, periodic callbacks,
// and logging behavior.
type Config struct {
	// AsyncTimer specifies the interval for calling AsyncFct.
	// If zero or negative, async callbacks are disabled.
	// Must be > 0 and AsyncFct must be non-nil to enable async callbacks.
	AsyncTimer time.Duration

	// AsyncMax limits the maximum number of concurrent async function calls.
	// If negative or zero, async functions are called sequentially.
	// If positive, up to AsyncMax async calls can run concurrently.
	// Requires AsyncTimer > 0 and AsyncFct != nil.
	AsyncMax int

	// AsyncFct is the function called periodically at AsyncTimer intervals.
	// This function is called asynchronously (non-blocking) and can be used for:
	//   - Periodic maintenance tasks
	//   - Sending heartbeats
	//   - Flushing buffers
	// The function receives the aggregator's context and should respect cancellation.
	// Can be nil if async callbacks are not needed.
	AsyncFct func(ctx context.Context)

	// SyncTimer specifies the interval for calling SyncFct.
	// If zero or negative, sync callbacks are disabled.
	// Must be > 0 and SyncFct must be non-nil to enable sync callbacks.
	SyncTimer time.Duration

	// SyncFct is the function called periodically at SyncTimer intervals.
	// This function is called synchronously (blocking) and can be used for:
	//   - File rotation
	//   - Database checkpoints
	//   - Resource cleanup
	// The function receives the aggregator's context and should respect cancellation.
	// Can be nil if sync callbacks are not needed.
	SyncFct func(ctx context.Context)

	// BufWriter specifies the size of the internal write buffer (channel capacity).
	// A larger buffer reduces contention but uses more memory.
	// If zero, defaults to 1 (minimal buffering).
	// Recommended values:
	//   - Low frequency writes (< 10/sec): 10-50
	//   - Medium frequency (10-100/sec): 100-500
	//   - High frequency (> 100/sec): 1000+
	BufWriter int

	// FctWriter is the function that receives and processes each write.
	// This function is called sequentially (never concurrently) for each write operation.
	// It must:
	//   - Handle the provided byte slice
	//   - Return the number of bytes written and any error
	//   - Be thread-safe if it accesses shared resources
	// This field is required and cannot be nil.
	//
	// Example:
	//   FctWriter: func(p []byte) (int, error) {
	//       return file.Write(p)
	//   }
	FctWriter func(p []byte) (n int, err error)
}
