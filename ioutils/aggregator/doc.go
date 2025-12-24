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

// Package aggregator provides a thread-safe write aggregator that buffers and serializes
// concurrent write operations to a single writer function.
//
// # Overview
//
// The aggregator package is designed to safely aggregate multiple concurrent write operations
// into a single sequential output stream. This is particularly useful when dealing with
// resources that don't support concurrent writes, such as files, network sockets, or
// databases.
//
// Key features:
//   - Thread-safe buffering of concurrent writes
//   - Configurable buffer size
//   - Automatic context propagation and cancellation
//   - Optional periodic async and sync callbacks
//   - Built-in logging support
//   - Lifecycle management (Start/Stop/Restart)
//
// # Use Cases
//
// Common scenarios where aggregator is useful:
//
//  1. Socket to File: Aggregating data received from multiple socket connections
//     (e.g., github.com/nabbar/golib/socket/server) and writing to a single file
//     that doesn't support concurrent writes.
//
//  2. Log Aggregation: Collecting logs from multiple goroutines and writing them
//     sequentially to a log file or remote service.
//
//  3. Database Batch Writes: Buffering multiple write operations before batching
//     them into a single database transaction.
//
//  4. Network Protocol: Serializing messages from multiple producers before sending
//     them over a network connection that requires sequential writes.
//
// # Architecture
//
// The aggregator uses a channel-based architecture:
//
//	┌─────────────┐
//	│  Writer 1   │──┐
//	└─────────────┘  │
//	                 │    ┌──────────┐      ┌──────────┐      ┌─────────────┐
//	┌─────────────┐  ├───→│ Channel  │─────→│ Goroutine│─────→│ FctWriter() │
//	│  Writer 2   │──┤    │ (Buffer) │      │  (run)   │      └─────────────┘
//	└─────────────┘  │    └──────────┘      └──────────┘
//	                 │                            │
//	┌─────────────┐  │                            ├──→ AsyncFct (periodic)
//	│  Writer N   │──┘                            └──→ SyncFct  (periodic)
//	└─────────────┘
//
// All writes go through a buffered channel and are processed sequentially by a single
// goroutine, ensuring thread-safety and preventing race conditions.
//
// # Basic Usage
//
// Create and start an aggregator:
//
//	ctx := context.Background()
//
//	// Define your writer function
//	writer := func(p []byte) (int, error) {
//	    // Write to file, socket, database, etc.
//	    return file.Write(p)
//	}
//
//	// Configure the aggregator
//	cfg := aggregator.Config{
//	    BufWriter: 100,        // Buffer up to 100 writes
//	    FctWriter: writer,     // Writer function
//	}
//
//	// Create aggregator
//	agg, err := aggregator.New(ctx, cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start processing
//	if err := agg.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer agg.Close()
//
//	// Write data (thread-safe)
//	agg.Write([]byte("data from goroutine 1"))
//	agg.Write([]byte("data from goroutine 2"))
//
// # Advanced Features
//
// Periodic Callbacks:
//
//	cfg := aggregator.Config{
//	    BufWriter:  100,
//	    FctWriter:  writer,
//	    AsyncTimer: 10 * time.Second,
//	    AsyncMax:   5,  // Max 5 concurrent async calls
//	    AsyncFct: func(ctx context.Context) {
//	        // Called every 10 seconds (async)
//	        // e.g., flush buffers, send heartbeat
//	    },
//	    SyncTimer: 1 * time.Minute,
//	    SyncFct: func(ctx context.Context) {
//	        // Called every minute (sync)
//	        // e.g., rotate files, checkpoint
//	    },
//	}
//
// # Buffer Sizing (BufWriter)
//
// The BufWriter parameter is critical for the aggregator's performance and behavior.
// It defines the capacity of the internal channel that buffers write operations before
// they are processed by FctWriter.
//
// Sizing Formula:
//
// The buffer should be sized to accommodate all write operations that occur during
// the longest of these time periods:
//
//  1. The interval between SyncFct calls (if configured)
//  2. The maximum execution time of a single FctWriter call
//  3. Plus a safety margin for handling bursts and system variability
//
// Mathematical approach:
//
//	BufferSize = (WriteRate * MaxProcessingTime) + SafetyMargin
//
// Where:
//   - WriteRate: Expected number of Write() calls per second
//   - MaxProcessingTime: max(SyncTimer, FctWriter execution time)
//   - SafetyMargin: 20-50% of calculated size for burst handling
//
// Trade-offs:
//
// Too Small Buffer:
//   - Write() operations will block when the buffer is full
//   - This creates backpressure on calling goroutines
//   - Can cause cascading delays in concurrent systems
//   - Symptoms: High Write() latency, goroutine blocking
//
// Too Large Buffer:
//   - Excessive memory consumption (each buffered write holds a []byte)
//   - Memory usage ≈ BufWriter × Average write size
//   - May hide performance issues by masking slow FctWriter
//   - Increased latency between Write() call and actual write to destination
//
// Optimal Buffer:
//   - Write() operations never block under normal load
//   - Memory usage is reasonable for the workload
//   - Provides headroom for temporary bursts
//   - Quick feedback when system is overloaded
//
// Example Calculations:
//
//	Scenario 1: Low frequency, fast writes
//	  - WriteRate: 10 writes/second
//	  - SyncTimer: 5 seconds (SyncFct takes 100ms)
//	  - FctWriter: 10ms per write
//	  - Calculation: (10 writes/s × 5s) + 50% = 75
//	  → Recommended: BufWriter = 100
//
//	Scenario 2: High frequency, slow writes
//	  - WriteRate: 100 writes/second
//	  - SyncTimer: not used
//	  - FctWriter: 50ms per write (network I/O)
//	  - Calculation: (100 × 0.05s) + 50% = 7.5
//	  → Recommended: BufWriter = 10-20 (writes queue during processing)
//
//	Scenario 3: Burst traffic (socket server)
//	  - WriteRate: Varies (0-1000 writes/second in bursts)
//	  - SyncTimer: 10 seconds (file sync)
//	  - FctWriter: 5ms per write
//	  - Peak during burst: 1000 × 0.1s = 100 writes
//	  - During sync: 1000 × 10s = 10,000 writes (unrealistic)
//	  - Practical: Handle 1s of peak traffic = 1000
//	  → Recommended: BufWriter = 1000-2000
//
// Monitoring:
//
// The aggregator provides four metrics for real-time buffer monitoring:
//
//	// Count-based metrics (number of items)
//	waiting := agg.NbWaiting()      // Number of Write() calls blocked
//	processing := agg.NbProcessing() // Number of items in buffer
//	bufferUsage := (processing * 100) / BufWriter  // Buffer utilization %
//
//	// Size-based metrics (memory usage in bytes)
//	sizeWaiting := agg.SizeWaiting()      // Total bytes waiting to enter buffer
//	sizeProcessing := agg.SizeProcessing() // Total bytes in buffer
//	totalMemory := sizeWaiting + sizeProcessing // Total memory in flight
//	avgMsgSize := sizeProcessing / max(processing, 1) // Average message size
//
// Signs that BufWriter is too small:
//   - NbWaiting() > 0: Write() calls are blocking (backpressure)
//   - NbWaiting() growing: Buffer cannot keep up with write rate
//   - NbProcessing() ≈ BufWriter: Buffer is consistently full
//   - SizeWaiting() increasing: Memory building up in blocked writes
//   - Write() calls take longer than expected
//   - Use pprof to check goroutine blocking profiles
//
// Signs that BufWriter is too large:
//   - NbProcessing() always near 0: Buffer is oversized
//   - SizeProcessing() low despite large BufWriter: Wasted capacity
//   - Memory usage grows proportionally with BufWriter even under low load
//   - Large delay between Write() call and actual disk/network write
//
// Signs of memory pressure:
//   - SizeProcessing() + SizeWaiting() exceeds memory budget
//   - Large average message size: SizeProcessing() / NbProcessing()
//   - Need to reduce BufWriter or implement message size limits
//   - Consider splitting large messages or using streaming
//
// Healthy State:
//   - NbWaiting() == 0: No blocking writes
//   - NbProcessing() varies with load but < BufWriter
//   - Buffer usage peaks at 60-80% during bursts
//   - SizeProcessing() stays within expected memory budget
//   - SizeWaiting() == 0: No memory blocked
//
// Best Practice:
//
// Start with a conservative estimate and monitor:
//  1. Measure actual WriteRate under typical load
//  2. Profile FctWriter and SyncFct execution times
//  3. Measure average message size in production
//  4. Calculate: BufWriter = (WriteRate × MaxTime) × 1.5
//  5. Estimate memory: MaxMemory = BufWriter × AvgMessageSize
//  6. Deploy and monitor all four metrics in production:
//     - NbWaiting() and NbProcessing() for buffer utilization
//     - SizeWaiting() and SizeProcessing() for memory usage
//  7. Alert if:
//     - NbWaiting() > 0 for extended periods (backpressure)
//     - SizeProcessing() + SizeWaiting() > memory budget (memory pressure)
//     - Average message size spikes unexpectedly
//  8. Adjust BufWriter based on observed metrics and memory constraints
//
// # Error Handling
//
// The package defines several error types:
//
//   - ErrInvalidWriter: Returned when Config.FctWriter is nil
//   - ErrInvalidInstance: Returned when internal state is corrupted
//   - ErrStillRunning: Returned when trying to Start() while already running
//   - ErrClosedResources: Returned when writing to a closed aggregator
//
// # Thread Safety
//
// All public methods of the Aggregator interface are thread-safe and can be called
// concurrently from multiple goroutines. The aggregator uses atomic operations and
// a single processing goroutine to ensure data consistency.
//
// # Context Integration
//
// The Aggregator implements context.Context, allowing it to be used in any function
// that accepts a context. When the parent context is cancelled, the aggregator
// automatically stops processing and closes.
//
// # Performance
//
// Based on real benchmarks (see benchmark_test.go):
//   - Write throughput: 1000+ writes/second
//   - Write latency: < 1ms median (including metrics overhead)
//   - Metrics read latency: < 5µs median for all 4 metrics
//   - Start/Stop time: ~130ms median (includes synchronization)
//   - Memory overhead: ~100 bytes per buffered write
//
// Buffer size should be tuned based on write frequency and latency requirements.
// A larger buffer reduces contention but increases memory usage. See the
// Buffer Sizing section above for detailed guidance.
//
// # Dependencies
//
// This package depends on:
//   - github.com/nabbar/golib/atomic: Thread-safe atomic value storage
//   - github.com/nabbar/golib/logger: Structured logging
//   - github.com/nabbar/golib/runner/startStop: Lifecycle management
//   - github.com/nabbar/golib/semaphore: Concurrency control for async functions
//
// # Example with Socket Server
//
// Using aggregator with github.com/nabbar/golib/socket/server:
//
//	// Create a temporary file for aggregated socket data
//	tmpFile, _ := os.CreateTemp("", "socket-data-*.tmp")
//	defer tmpFile.Close()
//
//	// Create aggregator to serialize writes to the file
//	agg, _ := aggregator.New(ctx, aggregator.Config{
//	    BufWriter: 1000,
//	    FctWriter: tmpFile.Write,
//	}, logger)
//	agg.Start(ctx)
//	defer agg.Close()
//
//	// Socket server handler
//	socketHandler := func(conn net.Conn) {
//	    defer conn.Close()
//	    buf := make([]byte, 4096)
//	    for {
//	        n, err := conn.Read(buf)
//	        if err != nil {
//	            break
//	        }
//	        // Write to aggregator (thread-safe)
//	        agg.Write(buf[:n])
//	    }
//	}
//
// See the example_test.go file for more detailed usage examples.
package aggregator
