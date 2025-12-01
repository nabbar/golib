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
	// The aggregator requires a valid writer function to process data.
	ErrInvalidWriter = errors.New("invalid writer")

	// ErrInvalidInstance is returned when the aggregator's internal state is corrupted
	// or when attempting to use an uninitialized instance.
	ErrInvalidInstance = errors.New("invalid instance")

	// ErrStillRunning is returned by Start when the aggregator is already running.
	// This prevents multiple concurrent run loops which could cause data corruption.
	ErrStillRunning = errors.New("still running")

	// ErrClosedResources is returned by Write when attempting to write to an aggregator
	// that has been closed or whose context has been cancelled.
	ErrClosedResources = errors.New("closed resources")

	// closedChan is a pre-closed channel used as a sentinel value to indicate
	// that the aggregator's write channel has been closed.
	closedChan = make(chan []byte, 1)
)

func init() {
	close(closedChan)
}

// Aggregator provides a thread-safe write aggregator that serializes concurrent
// write operations to a single output function.
//
// The Aggregator interface embeds:
//   - context.Context: for propagating cancellation and deadlines
//   - librun.StartStop: for lifecycle management (Start, Stop, Restart, IsRunning)
//   - io.Writer: for accepting write operations
//   - io.Closer: for cleanup and shutdown
//
// All methods are safe for concurrent use by multiple goroutines.
// Writes are buffered in a channel and processed sequentially by a single
// goroutine, ensuring that the output writer function is never called concurrently.
//
// The aggregator must be started with Start() before accepting writes.
// Attempting to write before Start() returns ErrClosedResources.
//
// Example:
//
//	agg, _ := aggregator.New(ctx, cfg, logger)
//	agg.Start(ctx)
//	defer agg.Close()
//	agg.Write([]byte("data"))
type Aggregator interface {
	context.Context
	librun.StartStop

	io.Closer
	io.Writer

	// SetLoggerError sets a custom error logging function.
	// If nil, a no-op function is used. Thread-safe.
	SetLoggerError(func(msg string, err ...error))

	// SetLoggerInfo sets a custom info logging function.
	// If nil, a no-op function is used. Thread-safe.
	SetLoggerInfo(func(msg string, arg ...any))

	// NbWaiting returns the number of Write() calls currently blocked waiting
	// to send data to the internal channel.
	//
	// This counter helps monitor backpressure:
	//   - If NbWaiting > 0: The buffer (BufWriter) is full and Write() calls are blocking
	//   - If NbWaiting grows: Consider increasing BufWriter or optimizing FctWriter
	//   - If NbWaiting == 0: All Write() calls are non-blocking (healthy state)
	//
	// Use this metric to detect when the aggregator cannot keep up with the write rate.
	NbWaiting() int64

	// NbProcessing returns the number of data items currently buffered in the
	// internal channel waiting to be processed by FctWriter.
	//
	// This counter helps monitor buffer utilization:
	//   - If NbProcessing ≈ BufWriter: The buffer is nearly full (high load)
	//   - If NbProcessing ≈ 0: The buffer is mostly empty (low load or fast processing)
	//   - If NbProcessing fluctuates: Normal behavior under variable load
	//
	// Combined with NbWaiting(), this provides complete visibility into the
	// aggregator's buffering state:
	//   - Total items in flight = NbWaiting + NbProcessing
	//   - Buffer usage % = (NbProcessing / BufWriter) × 100
	NbProcessing() int64

	// SizeWaiting returns the total size in bytes of all Write() calls currently
	// blocked waiting to send data to the internal channel.
	//
	// This metric helps estimate memory pressure from blocked writes:
	//   - If SizeWaiting > 0: Memory is allocated but blocked waiting for buffer space
	//   - If SizeWaiting is large: Consider increasing BufWriter to reduce blocking
	//   - If SizeWaiting == 0: All writes are non-blocking (healthy state)
	//
	// Combined with SizeProcessing(), this provides total memory usage:
	//   - Total memory = SizeWaiting + SizeProcessing bytes
	//   - Average message size = SizeProcessing / NbProcessing (if NbProcessing > 0)
	//
	// Use this to:
	//   - Detect memory buildup before it becomes a problem
	//   - Calculate actual vs theoretical buffer memory usage
	//   - Optimize BufWriter based on real data sizes
	SizeWaiting() int64

	// SizeProcessing returns the total size in bytes of all data items currently
	// buffered in the internal channel waiting to be processed by FctWriter.
	//
	// This metric helps monitor actual memory consumption:
	//   - Memory used by buffer ≈ SizeProcessing bytes
	//   - Average message size = SizeProcessing / NbProcessing
	//   - Theoretical max memory = BufWriter × Average message size
	//
	// Combined with NbProcessing(), this helps detect:
	//   - Large messages causing memory spikes
	//   - Uneven message size distribution
	//   - Actual memory usage vs configured buffer size
	//
	// Example monitoring:
	//
	//	avgSize := agg.SizeProcessing() / max(agg.NbProcessing(), 1)
	//	maxMemory := bufWriter × avgSize
	//	if maxMemory > memoryBudget {
	//	    // Need to reduce BufWriter or implement message size limits
	//	}
	SizeProcessing() int64
}

// New creates a new Aggregator instance with the given configuration.
//
// Parameters:
//   - ctx: Parent context for cancellation propagation. If nil, context.Background() is used.
//   - cfg: Configuration specifying buffer size, writer function, and optional callbacks.
//   - lg: Logger for internal operations. If nil, a default logger is created.
//
// Returns:
//   - Aggregator: A new aggregator instance ready to be started.
//   - error: ErrInvalidWriter if cfg.FctWriter is nil, or logger configuration error.
//
// The returned aggregator is in a stopped state and must be started with Start()
// before accepting writes. The aggregator will inherit deadlines and values from
// the parent context.
//
// Example:
//
//	cfg := aggregator.Config{
//	    BufWriter: 100,
//	    FctWriter: func(p []byte) (int, error) {
//	        return file.Write(p)
//	    },
//	}
//	agg, err := aggregator.New(ctx, cfg, logger)
//	if err != nil {
//	    return err
//	}
func New(ctx context.Context, cfg Config) (Aggregator, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	a := &agg{
		x:  libatm.NewValue[context.Context](),
		n:  libatm.NewValue[context.CancelFunc](),
		r:  libatm.NewValue[librun.StartStop](),
		le: libatm.NewValue[func(msg string, err ...error)](),
		li: libatm.NewValue[func(msg string, arg ...any)](),
		at: time.Minute,
		am: -1,
		af: nil,
		st: time.Minute,
		sf: nil,
		mw: sync.Mutex{},
		fw: nil,
		sh: 1,
		ch: libatm.NewValue[chan []byte](),
		op: new(atomic.Bool),
		cd: new(atomic.Int64),
		cw: new(atomic.Int64),
		sd: new(atomic.Int64),
		sw: new(atomic.Int64),
	}

	// Store initial context (but don't open channel yet - done in run())
	a.ctxNew(ctx)
	a.op.Store(false)

	if cfg.AsyncMax > -1 {
		a.am = cfg.AsyncMax
	}

	if cfg.AsyncTimer > 0 && cfg.AsyncFct != nil {
		a.at = cfg.AsyncTimer
		a.af = cfg.AsyncFct
	}

	if cfg.SyncTimer > 0 && cfg.SyncFct != nil {
		a.st = cfg.SyncTimer
		a.sf = cfg.SyncFct
	}

	if cfg.BufWriter != 0 {
		a.sh = cfg.BufWriter
	}

	if cfg.FctWriter != nil {
		a.fw = cfg.FctWriter
	} else {
		return nil, ErrInvalidWriter
	}

	return a, nil
}
