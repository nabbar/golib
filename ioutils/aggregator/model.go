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
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	"github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
	libsem "github.com/nabbar/golib/semaphore"
)

// agg is the primary internal implementation of the Aggregator interface.
//
// Technical Design and High-Level Architecture:
//   - Concurrency Management: The implementation utilizes high-performance atomic storage containers
//     (libatm.Value) for all operational state variables, including contexts, cancellation functions,
//     runners, and logging callbacks. This architectural choice enables lock-free read access during
//     the critical hot paths of data processing, effectively minimizing mutex contention when
//     multiple producer goroutines are active.
//   - Memory Efficiency: To achieve near-zero allocation throughput during steady-state operation,
//     a sync.Pool is strategically employed to recycle pointers to byte slices (*[]byte). This pattern
//     is crucial because it prevents the Go runtime from performing 'runtime.convTslice' operations.
//     Such operations occur when passing naked slices through channels or interface boundaries,
//     leading to unnecessary heap allocations, increased Garbage Collector (GC) pressure, and
//     higher CPU overhead.
//   - Monitoring and Backpressure: Real-time telemetry and system health monitoring are maintained
//     through a comprehensive suite of atomic Int64 counters. These counters provide instantaneous
//     visibility into the quantity and total cumulative byte size of:
//     a) 'Processing' items: Data blocks currently residing in the internal buffered channel.
//     b) 'Waiting' items: Producer goroutines (Write() calls) that are currently blocked due
//     to internal channel saturation (backpressure).
//   - Lifecycle Orchestration: The operational state of the aggregator is managed by a standardized
//     StartStop runner (librun). This ensures strictly ordered, thread-safe, and idempotent
//     transitions between the 'Stopped', 'Starting', 'Running', and 'Stopping' states.
type agg struct {
	// Root and Operational Contexts
	m libatm.Value[context.Context]    // m (master): Root context provided during instantiation, used as the base for all derived lifecycle contexts.
	x libatm.Value[context.Context]    // x (context): Operational context created specifically for the active run loop; cancelled upon stopping.
	n libatm.Value[context.CancelFunc] // n (now-cancel): Cancellation trigger used to signal a graceful shutdown of the background processing goroutine.

	// Lifecycle Management
	r libatm.Value[librun.StartStop] // r (runner): The dedicated orchestrator (StartStop) managing the background goroutine's execution.

	// Logging Callbacks
	le libatm.Value[func(msg string, err ...error)] // le (log-error): Atomic storage for the user-provided callback for internal error reporting.
	li libatm.Value[func(msg string, arg ...any)]   // li (log-info): Atomic storage for the user-provided callback for informational events.

	// Periodic Async Callback Configuration
	at time.Duration             // at (async-timer): The configured frequency at which the periodic asynchronous callback is executed.
	am int                       // am (async-max): Hard limit on the maximum number of concurrent asynchronous callback goroutines (semaphore based).
	af func(ctx context.Context) // af (async-fct): The user-defined function executed asynchronously at the 'at' interval.

	// Periodic Sync Callback Configuration
	st time.Duration             // st (sync-timer): The configured frequency at which the periodic synchronous callback is executed.
	sf func(ctx context.Context) // sf (sync-fct): The user-defined function executed synchronously within the main run loop.

	// Internal State and Data Pipeline
	lc sync.Mutex                        // lc (lifecycle-mutex): Guards high-level state transitions to prevent race conditions during Start/Stop phases.
	fw func(p []byte) (n int, err error) // fw (fct-writer): The core function (Config.FctWriter) where all aggregated data is ultimately serialized.
	sh int                               // sh (shelf-capacity): The specified buffer size defining how many data blocks can be queued before blocking.
	bs int                               // bs (buffer-size): The specified maximum size for each buffer allocated in the pool (default 8KB).
	ch libatm.Value[chan *[]byte]        // ch (channel): Atomic container for the buffered channel transferring data from Write() to the run loop.
	op *atomic.Bool                      // op (operational): High-speed atomic boolean indicating if the aggregator is currently active and accepting data.

	// Telemetry Counters (Items)
	cd *atomic.Int64 // cd (count-data): Processing counter tracking the number of discrete data blocks currently queued within the internal channel.
	cw *atomic.Int64 // cw (count-waiting): Waiting counter tracking the number of producer goroutines (Write() calls) currently blocked by buffer saturation.

	// Telemetry Counters (Bytes)
	sd *atomic.Int64 // sd (size-data): Processing size tracking the cumulative size in bytes of all data blocks currently residing in the pipeline.
	sw *atomic.Int64 // sw (size-waiting): Waiting size tracking the cumulative size in bytes of all data blocks currently blocked in pending Write() calls.

	// Resource Recycling
	bp sync.Pool // bp (buffer-pool): High-performance cache for *[]byte pointers, enabling buffer reuse and minimizing GC pressure.
}

// NbWaiting returns the instantaneous count of Write() calls currently blocked by buffer saturation.
//
// Technical Insight:
// This metric is vital for detecting producer-side congestion (backpressure). A non-zero
// value indicates that the ingestion rate exceeds the serialization rate of the FctWriter,
// causing the system to throttle producers to preserve stability.
//
// Example Use Case (Monitoring):
//
//	waitingCount := agg.NbWaiting()
//	if waitingCount > 0 {
//	    log.Printf("Performance Alert: Aggregator buffer is saturated, %d producers are waiting", waitingCount)
//	}
func (o *agg) NbWaiting() int64 {
	return o.cw.Load()
}

// SizeWaiting returns the total volume in bytes of data held by Write() calls currently blocked.
//
// Technical Insight:
// This represents the memory impact of backpressure on the system. High values indicate
// that a significant amount of memory is tied up in blocked producer goroutines' stacks
// or heap allocations that cannot be released until the pipeline drains.
//
// Example Use Case (Capacity Planning):
//
//	waitingBytes := agg.SizeWaiting()
//	if waitingBytes > 100 * 1024 * 1024 { // 100MB
//	    log.Fatal("Critical: Backpressure memory footprint exceeds safety threshold")
//	}
func (o *agg) SizeWaiting() int64 {
	return o.sw.Load()
}

// NbProcessing returns the current number of discrete data items residing in the internal buffer.
//
// Technical Insight:
// A value consistently close to the buffer capacity (Config.BufWriter) suggests that
// the FctWriter is the primary bottleneck. If this value stays near zero, the buffer
// might be oversized for the current workload.
//
// Example Use Case (Health Check):
//
//	processedItems := agg.NbProcessing()
//	usageRatio := float64(processedItems) / float64(maxBufferSize)
//	fmt.Printf("Buffer utilization: %.2f%%\n", usageRatio * 100)
func (o *agg) NbProcessing() int64 {
	return o.cd.Load()
}

// SizeProcessing returns the total volume in bytes of data currently buffered and awaiting serialization.
//
// Technical Insight:
// This provides a real-time measurement of the active memory footprint of the aggregator's
// data pipeline, excluding the pool of recycled buffers. Use this to monitor for memory
// spikes caused by unusually large messages.
//
// Example Use Case (Telemetry):
//
//	bufferedBytes := agg.SizeProcessing()
//	telemetry.Gauge("aggregator.buffer.bytes", bufferedBytes)
func (o *agg) SizeProcessing() int64 {
	return o.sd.Load()
}

// run serves as the primary processing engine for the aggregator, executing in a background goroutine.
//
// Operational Logic and High-Throughput Optimizations:
//  1. Atomic Initialization: Prepares the operational context, opens the data transmission channel,
//     and resets all monitoring counters to ensure a deterministic and clean startup state.
//  2. Local Caching Strategy: To minimize atomic overhead during the high-frequency 'select' loop,
//     essential references (data channel, 'Done' channels) are cached in local stack variables.
//     This significantly reduces the number of 'atomic.Load' calls required per iteration.
//  3. Signal Multiplexing: The loop performs non-blocking orchestration of multiple signals:
//     - Context-driven graceful shutdown (both from the runner and parent context).
//     - Periodic synchronous and asynchronous callback triggers.
//     - High-speed data ingestion from the internal buffered channel.
//  4. Opportunistic Batch Processing: When a message is received, the loop invokes 'processBatch'.
//     This greedily drains up to 100 additional messages from the channel without returning to
//     the main 'select'. This optimization reduces context-switching costs and improves the
//     throughput of the target writer function (FctWriter).
//
// Parameters:
//   - ctx: The operational context provided by the runner. Cancellation of this context
//     initiates the graceful shutdown sequence of the processing loop.
//
// Returns:
//   - error: Returns the reason for termination (typically context cancellation or internal state error).
func (o *agg) run(ctx context.Context) error {
	defer func() {
		// Panic recovery ensures the background goroutine doesn't crash the entire application.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/run", r)
		}
	}()

	// Critical Validation: Operation is impossible without a target writer function.
	if o.fw == nil {
		return ErrInvalidInstance
	}

	// Prevent accidental multi-start by verifying the high-speed operational flag.
	if o.op.Load() {
		return ErrStillRunning
	}

	// Lifecycle State Setup: Initialize operational context and the primary data channel.
	o.ctxNew()
	o.chanOpen()
	o.cntReset() // Ensure all telemetry counters start from zero for accurate monitoring of the new run.

	var (
		// Semaphore for managing maximum concurrency of async callbacks.
		sem = libsem.New(o, o.am, false)
		// Tickers for periodic callback execution.
		tckAsc = time.NewTicker(o.at)
		tckSnc = time.NewTicker(o.st)
		// Executor closures for callbacks.
		fctAsyn = o.callASyn()
		fctSyn  = o.callSyn()

		// Performance Optimization: Cache atomic values locally before entering the select loop.
		// These variables reside on the stack, providing faster access than atomic loads.
		chData = o.chanData()
		ctxAgg = o.x.Load()
		ctxDon = ctx.Done()
		aggDon = ctxAgg.Done()
	)

	defer func() {
		// Ordered Cleanup: ensure timers are stopped and semaphore workers are finalized.
		if sem != nil {
			sem.DeferMain()
		}
		tckSnc.Stop()
		tckAsc.Stop()
		o.logInfo("stopping aggregator processing loop")
	}()

	o.logInfo("starting aggregator processing loop")

	// Primary Orchestration Loop: Manages I/O, timers, and lifecycle signals.
	for {
		select {
		case <-aggDon:
			// Signal from the internal operational context (e.g., via manual Stop() call).
			// We trigger an explicit Stop to ensure the runner state is updated correctly.
			_ = o.Stop(context.Background())
			return o.Err()

		case <-ctxDon:
			// Signal from the parent context (e.g., application shutdown).
			_ = o.Stop(context.Background())
			return ctx.Err()

		case <-tckAsc.C:
			// Periodic trigger for the asynchronous callback.
			// Execution is decoupled from the main loop via a semaphore and goroutines.
			fctAsyn(sem)

		case <-tckSnc.C:
			// Periodic trigger for the synchronous callback.
			// Executed sequentially in the main run loop; blocks ingestion while running.
			fctSyn()

		case p, ok := <-chData:
			// Primary data ingestion point.
			if !ok {
				// The channel was closed (typically by chanClose during shutdown).
				return nil
			}
			// Execute high-throughput batching logic to process multiple items in one go.
			o.processBatch(p, chData)
		}
	}
}

// processBatch implements a high-performance draining strategy for the internal buffered channel.
//
// Performance Rationale:
// It processes the initial message 'p' and then attempts to ingest up to 100 subsequent
// available messages using a non-blocking 'select' pattern. By aggregating work, it
// minimizes the overhead of the Go runtime's scheduler and the 'select' mechanism,
// while increasing the payload size provided to the target writer function (o.fw),
// which often results in fewer system calls or network packets.
//
// Parameters:
//   - p: Pointer to the first byte slice retrieved from the channel.
//   - chData: The data channel from which to attempt additional non-blocking receives.
func (o *agg) processBatch(p *[]byte, chData <-chan *[]byte) {
	var (
		ok    bool
		count int64 = 1
		size        = int64(len(*p))
	)

	// Pass the first data block to the target writer.
	o.logError("error writing data", o.fctWrt(p))
	// Return the buffer pointer to the sync.Pool for future reuse.
	o.bp.Put(p)

	// Opportunistic Greedy Drain: loop up to 100 times to process pending channel items.
	// This helps empty the buffer faster during traffic bursts.
	for i := 0; i < 100; i++ {
		select {
		case p, ok = <-chData:
			if !ok {
				// Channel closed mid-batch; update counters and return.
				o.cntDataDec(count, size)
				return
			}
			count++
			size += int64(len(*p))
			// Write the additional block and recycle the buffer.
			o.logError("error writing data", o.fctWrt(p))
			o.bp.Put(p)
		default:
			// Buffer is empty; update global counters and return control to the main orchestration loop.
			o.cntDataDec(count, size)
			return
		}
	}

	// Batch limit reached; update counters before potentially starting another batch or returning to select.
	o.cntDataDec(count, size)
}

// fctWrt serves as a centralized, safe wrapper for invoking the user-provided writer function.
// It performs basic validation (e.g., length checks) and ensures consistent error propagation.
func (o *agg) fctWrt(p *[]byte) error {
	// Guard against nil pointers or empty buffers to avoid panic or useless function calls.
	if p == nil || len(*p) == 0 {
		return nil
	}
	// Execute the writer function defined in Config.FctWriter.
	_, err := o.fw(*p)
	return err
}

// callASyn constructs the periodic asynchronous callback executor.
//
// Technical Implementation:
// It utilizes a semaphore to strictly enforce maximum concurrency limits (Config.AsyncMax).
// Each interval execution is spawned in its own goroutine, wrapped in a panic recovery
// mechanism to prevent callback failures from crashing the aggregator's main loop.
func (o *agg) callASyn() func(sem libsem.Semaphore) {
	// Pre-validation to return a no-op function if the aggregator is inactive or misconfigured.
	if !o.op.Load() || o.af == nil || o.x.Load() == nil {
		return func(sem libsem.Semaphore) {}
	}

	return func(sem libsem.Semaphore) {
		// Non-blocking acquisition attempt. If the semaphore is full, the call is skipped for this cycle
		// to avoid queuing up stale maintenance tasks.
		if !sem.NewWorkerTry() {
			return
		}

		// Parallel execution in a managed goroutine to ensure it doesn't block data aggregation.
		go func() {
			defer func() {
				// Individual goroutine recovery for the async callback.
				if r := recover(); r != nil {
					runner.RecoveryCaller("golib/ioutils/aggregator/callasyn", r)
				}
			}()

			// Ensure the worker slot is released back to the semaphore upon completion or panic.
			defer sem.DeferWorker()

			// Invoke the callback with the operational context for cancellation propagation.
			o.af(o.x.Load())
		}()
	}
}

// callSyn constructs the periodic synchronous callback executor.
//
// Technical Insight:
// These callbacks are executed sequentially within the main processing goroutine.
// They must be highly optimized and non-blocking to avoid stalling the data pipeline,
// as the main loop will not ingest new data until this function returns.
func (o *agg) callSyn() func() {
	// Pre-validation for configuration and operational state.
	if !o.op.Load() || o.sf == nil || o.x.Load() == nil {
		return func() {}
	}

	return func() {
		defer func() {
			// Synchronous callback recovery to protect the main run loop.
			if r := recover(); r != nil {
				runner.RecoveryCaller("golib/ioutils/aggregator/callsyn", r)
			}
		}()

		// Execute the callback synchronously within the run loop.
		o.sf(o.x.Load())
	}
}

// cntDataInc performs an atomic increment of the telemetry counters tracking data blocks
// currently queued in the internal pipeline (buffer channel).
func (o *agg) cntDataInc(i int) {
	o.cd.Add(1)
	o.sd.Add(int64(i))
}

// cntDataDec performs an atomic decrement of the processing telemetry counters for a batch of items.
// This is used by processBatch to update counters once for multiple items.
func (o *agg) cntDataDec(count, size int64) {
	o.cd.Add(-count)
	if o.cd.Load() < 0 {
		o.cd.Store(0)
	}
	o.sd.Add(-size)
	if o.sd.Load() < 0 {
		o.sd.Store(0)
	}
}

// cntWaitInc performs an atomic increment of the telemetry counters tracking producer
// goroutines currently blocked by backpressure (internal channel saturation).
func (o *agg) cntWaitInc(i int) {
	o.cw.Add(1)
	o.sw.Add(int64(i))
}

// cntWaitDec performs an atomic decrement of the waiting telemetry counters.
// It includes safety checks to ensure counters maintain a non-negative state.
func (o *agg) cntWaitDec(i int) {
	o.cw.Add(-1)
	if o.cw.Load() < 0 {
		o.cw.Store(0)
	}
	o.sw.Add(int64(-i))
	if o.sw.Load() < 0 {
		o.sw.Store(0)
	}
}

// cntReset restores all monitoring and telemetry counters to their initial zero state.
// This is typically called during the Start lifecycle phase to ensure accurate data for each run.
func (o *agg) cntReset() {
	o.cd.Store(0)
	o.sd.Store(0)
	o.cw.Store(0)
	o.sw.Store(0)
}

// cleanup provides a deterministic and thread-safe method to release internal resources.
// It decommission the operational context and marks the data channel as closed.
func (o *agg) cleanup() {
	o.ctxClose()
	o.chanClose()
}
