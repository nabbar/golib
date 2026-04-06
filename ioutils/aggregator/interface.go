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
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	librun "github.com/nabbar/golib/runner/startStop"
)

var (
	// ErrInvalidWriter is returned by New when Config.FctWriter is nil.
	// The aggregator requires a valid writer function to process and serialize data.
	// This function is the ultimate destination for all aggregated data and must
	// conform to the standard io.Writer signature.
	ErrInvalidWriter = errors.New("invalid writer")

	// ErrInvalidInstance is returned when the aggregator's internal state is corrupted,
	// when attempting to use an uninitialized instance, or when essential fields are missing.
	// This usually indicates a programmatic error or an incomplete initialization.
	ErrInvalidInstance = errors.New("invalid instance")

	// ErrStillRunning is returned by Start when the aggregator is already active.
	// This prevents multiple concurrent run loops from contending for the same writer
	// and ensures that the processing logic remains single-threaded for serialization.
	ErrStillRunning = errors.New("still running")

	// ErrClosedResources is returned by Write when attempting to queue data to an aggregator
	// that has been explicitly closed or whose context has been cancelled.
	// Once closed, an aggregator cannot be reused for writing; a new instance should be created
	// if further aggregation is required.
	ErrClosedResources = errors.New("closed resources")

	// ErrTimeout is returned by Start when the aggregator fails to reach a running state
	// within the internally defined timeout period (usually 1 second).
	// This can happen if the background goroutine is blocked during its startup sequence.
	ErrTimeout = errors.New("timeout")

	// closedChan is a pre-closed sentinel channel used to safely indicate that the
	// aggregator's write channel has been decommissioned without triggering panics.
	// Using a pointer to a byte slice (*[]byte) optimizes performance by reducing
	// heap allocations during interface conversions (convTslice).
	closedChan = make(chan *[]byte, 1)
)

// init performs package-level initialization.
// In this package, it ensures the closedChan sentinel is pre-closed to serve
// as an atomic marker for a decommissioned data pipeline.
func init() {
	close(closedChan)
}

// Aggregator provides a high-performance, thread-safe write aggregator.
// It serializes concurrent write operations from multiple goroutines to a single
// output function, implementing efficient buffering and batching mechanisms.
//
// The Aggregator interface embeds several standard and custom interfaces:
//   - context.Context: For propagating cancellation signals and deadlines throughout the pipeline.
//   - librun.StartStop: For lifecycle management (Start, Stop, Restart) using a standardized runner.
//   - io.Writer: The primary entry point for accepting data to be aggregated.
//   - io.Closer: For graceful shutdown and deterministic resource cleanup.
//
// Technical Performance Features:
//   - Zero-allocation data transfer: Uses pointers to byte slices (*[]byte) in channels
//     and pools to avoid costly 'runtime.convTslice' allocations.
//   - Lock-free counter paths: Implements atomic counters for real-time monitoring.
//   - Opportunistic Non-blocking: Attempts non-blocking channel sends to minimize
//     counter contention and latency under normal load.
//   - Batch Processing: Drains the internal channel in chunks of up to 100 items to
//     reduce context switching and system call overhead in the writer function.
type Aggregator interface {
	context.Context
	librun.StartStop

	io.Closer
	io.Writer

	// SetLoggerError registers a custom callback for internal error reporting.
	// This function is invoked for write failures, context errors, and recovered panics.
	// If the provided function is nil, a no-op function is internally assigned.
	// Thread-safety: Safely updatable during runtime using atomic storage.
	SetLoggerError(func(msg string, err ...error))

	// SetLoggerInfo registers a custom callback for lifecycle and operational events.
	// It is used to log startup progress, graceful shutdowns, and informational status.
	// If the provided function is nil, a no-op function is internally assigned.
	// Thread-safety: Safely updatable during runtime using atomic storage.
	SetLoggerInfo(func(msg string, arg ...any))

	// NbWaiting returns the instantaneous count of Write() calls currently blocked
	// waiting for space in the internal buffer (Config.BufWriter).
	//
	// Backpressure Analysis:
	//   - NbWaiting > 0: Indicates the consumer (FctWriter) is slower than the producers.
	//   - Sustained Growth: Suggests the need to optimize FctWriter or increase BufWriter.
	//   - Constant Zero: System is operating within its designed capacity.
	NbWaiting() int64

	// NbProcessing returns the number of discrete data items currently queued in the
	// internal channel, awaiting serialization by the processing goroutine.
	//
	// Buffer Utilization:
	//   - Value relative to BufWriter indicates the current load percentage.
	//   - Fluctuations are expected; a pegged value suggests a bottleneck in FctWriter.
	NbProcessing() int64

	// SizeWaiting returns the total volume in bytes of data held by Write() calls
	// currently blocked due to buffer saturation.
	//
	// Memory Impact Analysis:
	//   - Represents memory held in producer goroutines stack/heap that cannot be released.
	//   - Combined with SizeProcessing(), provides total memory footprint of the aggregator.
	SizeWaiting() int64

	// SizeProcessing returns the total volume in bytes of data buffered in the
	// internal channel, including the current batch being processed.
	//
	// Resource Monitoring:
	//   - Helps calculate average message size: SizeProcessing / NbProcessing.
	//   - Crucial for detecting memory spikes caused by unusually large messages.
	SizeProcessing() int64
}

// New creates and initializes a new Aggregator instance based on the provided configuration.
//
// Initialization Process:
//  1. Context Normalization: If the provided ctx is nil, context.Background() is used.
//  2. Struct Allocation: Initializes the internal 'agg' struct with high-performance defaults.
//  3. Atomic Containers: Sets up libatm.Value wrappers for state that requires thread-safe updates.
//  4. Timer Defaults: Sets callback intervals to 1 hour by default to avoid accidental high-frequency triggers.
//  5. Buffer Configuration: Sets the internal channel capacity ('sh') and maximum buffer size ('bs').
//  6. Object Pool Warm-up: Pre-allocates a pool of byte slice pointers (*[]byte) to minimize
//     heap allocations during the initial burst of Write() calls.
//  7. Writer Validation: Ensures a valid FctWriter is provided, otherwise returns ErrInvalidWriter.
//
// Parameters:
//   - ctx: Parent context. If nil, context.Background() is used. Inherits values and deadlines.
//   - cfg: Configuration parameters defining buffer depth, the writer function, and callback intervals.
//
// Returns:
//   - Aggregator: A fully initialized instance in a stopped state (requires Start()).
//   - error: ErrInvalidWriter if cfg.FctWriter is missing.
func New(ctx context.Context, cfg Config) (Aggregator, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	a := &agg{
		// Atomic storage for contexts and runners to ensure lock-free access in the hot path.
		m:  libatm.NewValue[context.Context](),
		x:  libatm.NewValue[context.Context](),
		n:  libatm.NewValue[context.CancelFunc](),
		r:  libatm.NewValue[librun.StartStop](),
		le: libatm.NewValue[func(msg string, err ...error)](),
		li: libatm.NewValue[func(msg string, arg ...any)](),

		// Default timer intervals (1 hour) to prevent runaway execution if misconfigured.
		at: time.Hour,
		am: -1,
		af: nil,
		st: time.Hour,
		sf: nil,

		// Thread-safety and writer configuration.
		lc: sync.Mutex{},
		fw: nil,
		sh: 1,    // Minimal default buffering.
		bs: 8192, // Default buffer size set to 8KB (standard page/block size).

		// Data pipeline components using atomic storage for safe swaps.
		ch: libatm.NewValue[chan *[]byte](),
		op: new(atomic.Bool),

		// Telemetry counters using atomic primitives for low-overhead monitoring.
		cd: new(atomic.Int64),
		cw: new(atomic.Int64),
		sd: new(atomic.Int64),
		sw: new(atomic.Int64),

		// Resource recycling.
		bp: sync.Pool{},
	}

	// Persist the root context. Note: the operational channel is opened only during Start().
	a.m.Store(ctx)
	a.ctxNew()
	a.op.Store(false)

	// Runner initialization is mandatory to handle concurrent Start/Stop requests safely.
	a.setRunner(nil)

	// Async callback configuration.
	if cfg.AsyncMax > -1 {
		a.am = cfg.AsyncMax
	}
	if cfg.AsyncTimer > 0 && cfg.AsyncFct != nil {
		a.at = cfg.AsyncTimer
		a.af = cfg.AsyncFct
	}

	// Sync callback configuration.
	if cfg.SyncTimer > 0 && cfg.SyncFct != nil {
		a.st = cfg.SyncTimer
		a.sf = cfg.SyncFct
	}

	// Buffer depth configuration.
	if cfg.BufWriter != 0 {
		a.sh = cfg.BufWriter
	}

	// Buffer max size configuration.
	if cfg.BufMaxSize > 0 {
		a.bs = cfg.BufMaxSize
	}

	// Object Pool pre-allocation and New function setup.
	a.bp.New = func() any {
		// Allocates a fresh buffer pointer when the pool is exhausted.
		b := make([]byte, a.bs)
		return &b
	}

	// Warm up the pool with a set of buffers equal to the channel capacity plus one for processing.
	// This ensures that under normal steady-state load, the aggregator performs zero heap allocations.
	for i := 0; i < a.sh+1; i++ {
		b := make([]byte, a.bs)
		a.bp.Put(&b)
	}

	// Writer function validation.
	if cfg.FctWriter != nil {
		a.fw = cfg.FctWriter
	} else {
		return nil, ErrInvalidWriter
	}

	return a, nil
}
