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

// Config defines the configuration for creating a new Aggregator instance.
//
// The configuration allows precise customization of buffering strategies, periodic
// maintenance callbacks, and internal resource limits. Tuning these parameters
// is essential for balancing throughput, latency, and memory consumption.
type Config struct {
	// AsyncTimer specifies the execution interval for the AsyncFct callback.
	//
	// Operational Details:
	//   - If zero or negative, the asynchronous callback mechanism is disabled.
	//   - Requires AsyncFct to be non-nil for activation.
	//   - The timer is precise, but execution depends on goroutine scheduling.
	AsyncTimer time.Duration

	// AsyncMax limits the maximum number of concurrent executions of AsyncFct.
	//
	// Concurrency Control:
	//   - If negative, the aggregator uses a sequential execution model (one at a time).
	//   - If zero, it defaults to sequential execution.
	//   - If positive, it uses an internal semaphore to bound the number of concurrent
	//     goroutines executing the callback. If the limit is reached, a scheduled
	//     execution cycle is skipped (non-blocking).
	AsyncMax int

	// AsyncFct is the user-defined function executed periodically at the AsyncTimer interval.
	//
	// Parallel Execution:
	// This function is invoked in its own goroutine (managed by the AsyncMax limit) and
	// does not stall the main data aggregation pipeline.
	//
	// Use Cases:
	//   - Periodic cache invalidation or flushes.
	//   - Sending telemetry heartbeats or health signals.
	//   - Asynchronous data cleanup tasks.
	//
	// The function receives the aggregator's operational context and must respect
	// ctx.Done() signals to ensure graceful shutdown.
	AsyncFct func(ctx context.Context)

	// SyncTimer specifies the execution interval for the SyncFct callback.
	//
	// Operational Details:
	//   - If zero or negative, the synchronous callback mechanism is disabled.
	//   - Requires SyncFct to be non-nil for activation.
	SyncTimer time.Duration

	// SyncFct is the user-defined function executed periodically at the SyncTimer interval.
	//
	// Sequential Execution:
	// This function is executed synchronously within the main processing goroutine.
	// While SyncFct is running, NO data ingestion from the buffer channel occurs.
	// Therefore, it must be highly optimized and return quickly to avoid causing
	// backpressure in the pipeline.
	//
	// Use Cases:
	//   - Deterministic file rotation (where no writes should occur during rotation).
	//   - Database transaction checkpoints.
	//   - Critical state updates that require the pipeline to be paused.
	//
	// The function receives the aggregator's operational context and should monitor
	// for cancellation.
	SyncFct func(ctx context.Context)

	// BufWriter specifies the capacity of the internal buffered channel.
	//
	// Performance Impact:
	// This value defines the depth of the pipeline before producer goroutines (Write() calls)
	// begin to block.
	//
	// Recommended Sizing:
	//   - Low frequency writes (< 10/sec): 10-50 (minimal memory footprint).
	//   - Medium frequency (10-100/sec): 100-500.
	//   - High frequency (> 100/sec): 1000+ (to absorb bursts).
	//
	// See the 'Buffer Sizing' section in doc.go for a detailed sizing formula.
	BufWriter int

	// BufMaxSize specifies the internal maximum size for the pre-allocated buffers
	// managed by the sync.Pool.
	//
	// Memory Strategy:
	// This defines the initial length of the byte slices created during the pool's
	// warm-up phase (see New()). If a Write() call provides data exceeding this size,
	// a fresh buffer is allocated for that specific call to avoid truncation.
	// Standard value: 8192 (8KB).
	BufMaxSize int

	// FctWriter is the core serialization function that receives all aggregated data.
	//
	// Execution Guarantee:
	// This function is called sequentially by a single goroutine. It will never
	// be invoked concurrently by the aggregator, ensuring safe writes to non-thread-safe
	// destinations like standard files or certain network protocols.
	//
	// Responsibilities:
	//   - It must handle the provided byte slice 'p' entirely.
	//   - It must return the number of bytes processed and any encounter error.
	//   - It should avoid internal buffering if possible, as the aggregator already
	//     provides a high-performance buffer.
	//
	// This field is REQUIRED. New() will return ErrInvalidWriter if it is nil.
	FctWriter func(p []byte) (n int, err error)
}
