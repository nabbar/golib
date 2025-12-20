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
 *
 */

// Package multi provides a thread-safe, adaptive multi-writer that extends Go's standard
// io.MultiWriter with advanced features including adaptive sequential/parallel execution,
// latency monitoring, and comprehensive concurrency support.
//
// # Overview
//
// This package complements the standard library's io.MultiWriter by adding:
//   - Adaptive strategy: automatically switches between sequential and parallel writes
//   - Thread-safe operations using atomic primitives and concurrent maps
//   - Dynamic writer management (add/remove writers on the fly)
//   - Input source management with a single io.Reader
//   - Real-time latency monitoring and statistics
//   - Safe defaults (DiscardCloser) to prevent nil pointer panics
//
// Unlike io.MultiWriter which only handles writes, this package provides a complete
// ReadWriteCloser implementation with adaptive performance optimization based on
// observed write latency.
//
// # Relationship with io.MultiWriter
//
// The multi package uses io.MultiWriter internally for sequential write operations.
// When in sequential mode, writes are delegated to io.MultiWriter for efficient
// fan-out to all registered writers. However, multi extends this with:
//   - Parallel execution mode for high-latency scenarios
//   - Adaptive switching between modes based on performance metrics
//   - Thread-safe writer addition/removal during operation
//   - Input source management and Copy() operations
//
// # Key Features
//
//   - Adaptive write strategy (sequential ↔ parallel) based on latency monitoring
//   - Thread-safe concurrent operations (AddWriter, SetInput, Write, Read)
//   - Dynamic writer management with atomic state updates
//   - Built-in io.Copy support from input to all outputs
//   - Implements io.ReadWriteCloser, io.StringWriter interfaces
//   - Zero-allocation read/write in steady state (sequential mode)
//   - Comprehensive statistics (latency, mode, writer count)
//
// # Basic Usage
//
// Creating a new multi-writer and broadcasting writes:
//
//	m := multi.New()
//
//	// Add multiple write destinations
//	var buf1, buf2, buf3 bytes.Buffer
//	m.AddWriter(&buf1, &buf2, &buf3)
//
//	// Write data - it will be sent to all writers
//	m.Write([]byte("broadcast data"))
//	// buf1, buf2, and buf3 now all contain "broadcast data"
//
// # Input Source Management
//
// Setting an input reader and copying to all outputs:
//
//	// Set the input source
//	input := io.NopCloser(strings.NewReader("source data"))
//	m.SetInput(input)
//
//	// Copy from input to all registered writers
//	n, err := m.Copy()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Copied %d bytes\n", n)
//
// # Dynamic Writer Management
//
// Add writers dynamically and clean them:
//
//	// Add more writers on the fly
//	var buf4 bytes.Buffer
//	m.AddWriter(&buf4)
//
//	// Remove all writers
//	m.Clean()
//
//	// Add new set of writers
//	var newBuf bytes.Buffer
//	m.AddWriter(&newBuf)
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently:
//
//	var wg sync.WaitGroup
//
//	// Concurrent writes
//	for i := 0; i < 100; i++ {
//	    wg.Add(1)
//	    go func(i int) {
//	        defer wg.Done()
//	        m.Write([]byte(fmt.Sprintf("message %d", i)))
//	    }(i)
//	}
//
//	// Concurrent writer additions
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        var buf bytes.Buffer
//	        m.AddWriter(&buf)
//	    }()
//	}
//
//	wg.Wait()
//
// # Architecture
//
// The package follows a layered architecture with adaptive strategy selection:
//
//	┌────────────────────────────────────────────────────────────┐
//	│                           Multi                            │
//	├────────────────────────────────────────────────────────────┤
//	│                                                            │
//	│  ┌──────────────┐           ┌─────────────────────┐        │
//	│  │ Input Source │           │ Output Destinations │        │
//	│  │ (io.Reader)  │           │ (io.Writer Map)     │        │
//	│  └──────┬───────┘           └──────────┬──────────┘        │
//	│         │                              │                   │
//	│         ▼                              ▼                   │
//	│  ┌──────────────┐           ┌─────────────────────┐        │
//	│  │ ReaderWrap   │           │    WriteWrapper     │        │
//	│  └──────┬───────┘           └──────────┬──────────┘        │
//	│         │                              │                   │
//	│         │                   ┌──────────┴──────────┐        │
//	│         │                   │                     │        │
//	│         │           ┌───────▼──────┐      ┌───────▼──────┐ │
//	│         │           │ Sequential   │      │   Parallel   │ │
//	│         │           │ (io.Multi)   │      │ (Goroutines) │ │
//	│         │           └──────────────┘      └──────────────┘ │
//	│         │                   │                     │        │
//	│         ▼                   └─────────────────────┘        │
//	│   Client Read/Write/Copy Operations                        │
//	│                                                            │
//	└────────────────────────────────────────────────────────────┘
//
// # Data Flow
//
// Write Operation Flow:
//  1. Data arrives at Write(p []byte)
//  2. Latency measurement starts (time.Now())
//  3. Current write strategy is retrieved atomically
//  4. Sequential mode: delegates to io.MultiWriter for fan-out
//  5. Parallel mode: spawns goroutines if len(p) >= MinimalSize
//  6. Latency is recorded and accumulated atomically
//  7. After SampleWrite operations, mean latency triggers mode check
//  8. Adaptive mode may switch strategy if thresholds are crossed
//
// Read Operation Flow:
//  1. Read(p []byte) delegates to atomic readerWrapper
//  2. ReaderWrapper delegates to underlying io.Reader
//  3. If no reader set, DiscardCloser returns (0, nil)
//
// # Adaptive Strategy
//
// The multi-writer monitors write latency to optimize execution strategy:
//
// Sampling Window:
//   - Every SampleWrite operations (default: 100 writes)
//   - Calculates mean latency from accumulated measurements
//   - Resets counters after evaluation
//
// Mode Switching Logic:
//   - Sequential → Parallel: when mean latency > ThresholdLatency (5µs)
//     AND writer count >= MinimalWriter (3 writers)
//   - Parallel → Sequential: when mean latency < ThresholdLatency
//
// Parallel Execution Threshold:
//   - Parallel mode only executes concurrently if len(data) >= MinimalSize (512 bytes)
//   - Small writes use sequential execution to avoid goroutine overhead
//
// This adaptive approach ensures optimal performance by:
//   - Using sequential writes (io.MultiWriter) for low-latency scenarios
//   - Switching to parallel goroutines when write latency indicates blocking
//   - Avoiding goroutine overhead for small data sizes
//
// # Implementation Details
//
// The Multi type uses atomic primitives for lock-free concurrency:
//   - atomic.Value stores readerWrapper and writeWrapper with consistent types
//   - sync.Map (via libatm.MapTyped) stores registered writers with unique keys
//   - atomic.Int64 tracks writer count, sample count, and cumulative latency
//   - atomic.Bool flags for adaptive mode and current write strategy
//
// Thread-Safety Guarantees:
//   - All public methods are safe for concurrent use
//   - Writer addition/removal uses atomic map operations
//   - Mode switching is serialized through atomic compare-and-swap patterns
//   - No mutexes on hot path (read/write operations)
//
// Memory Management:
//   - Sequential mode: zero allocations per write (uses io.MultiWriter)
//   - Parallel mode: allocates goroutines and error channel per large write
//   - Default initialization with DiscardCloser prevents nil panics
//   - Close() properly cleans up resources and closes input reader
//
// # Error Handling
//
// The package defines ErrInstance which is returned when operations
// are attempted on invalid or uninitialized internal state. This typically
// should not occur during normal usage as New() initializes all required state.
//
// Write and read errors from underlying io.Writer and io.Reader implementations
// are propagated unchanged to the caller.
//
// # Performance Benchmarks
//
// Benchmark results on AMD64 architecture with Go 1.25:
//
// Operation Performance:
//   - Multi Creation: median 4.6µs, mean 5.3µs, max 8.8µs
//   - SetInput: <1µs (atomic pointer swap)
//   - AddWriter: <1µs (concurrent map insert)
//   - Sequential Write (3 writers, 1KB): median 400µs, mean 400µs
//   - Parallel Write (3 writers, 1KB): median 200µs, mean 233µs
//   - Adaptive Write: automatically selects optimal mode, mean 266µs
//   - Copy Operation: mean 266µs (input to all writers)
//   - Read Operations: mean 366µs (delegate to input source)
//
// Key Findings:
//   - Parallel mode shows ~50% latency reduction under high writer latency
//   - Sequential mode has zero allocation overhead (io.MultiWriter)
//   - Adaptive mode adds minimal overhead (<5%) for dynamic optimization
//   - Thread-safe operations have negligible contention (lock-free design)
//
// Memory Usage:
//   - Base overhead: minimal (struct fields + atomic pointers)
//   - Sequential writes: zero allocations per operation
//   - Parallel writes: goroutine stack + error channel per write
//   - Optimization: parallel mode activates only when beneficial
//
// Scalability:
//   - Linear scaling with CPU count in parallel mode
//   - No lock contention on hot path (read/write operations)
//   - Dynamic writer management without global locks
//
// # Use Cases
//
// 1. Log Broadcasting
//
// Send application logs to multiple destinations (stdout, file, network):
//
//	m := multi.New(false, false, multi.DefaultConfig())
//	m.AddWriter(os.Stdout, logFile, networkSocket)
//	log.SetOutput(m)
//	// All log.Println() calls now broadcast to all three destinations
//
// 2. Stream Replication with Adaptive Performance
//
// Copy data streams to multiple storage backends with automatic optimization:
//
//	m := multi.New(true, false, multi.DefaultConfig()) // Adaptive mode
//	m.SetInput(sourceStream)
//	m.AddWriter(s3Uploader, localDisk, backupServer)
//	n, err := m.Copy()
//	// Automatically switches to parallel if writers are slow
//
// 3. High-Throughput Writing with Mixed Latencies
//
// Handle scenarios with variable writer speeds (fast local + slow network):
//
//	cfg := multi.Config{
//	    SampleWrite:      50,   // Check every 50 writes
//	    ThresholdLatency: 1000, // 1µs threshold
//	    MinimalWriter:    2,
//	    MinimalSize:      1024,
//	}
//	m := multi.New(true, false, cfg)
//	m.AddWriter(fastLocalWriter, slowNetworkWriter)
//	// Prevents slow writer from blocking fast writer in parallel mode
//
// 4. Dynamic Writer Management
//
// Add/remove writers during operation (e.g., hot-swappable log destinations):
//
//	m := multi.New(false, false, multi.DefaultConfig())
//	m.AddWriter(initialWriter)
//	// Runtime: add monitoring without interrupting service
//	m.AddWriter(metricsWriter)
//	// Later: remove writers
//	m.Clean()
//	m.AddWriter(newWriter)
//
// # Limitations and Best Practices
//
// Limitations:
//
//   - Input Reader Concurrency: While Multi is thread-safe, the underlying
//     io.Reader set via SetInput may not support concurrent reads. Use external
//     synchronization for concurrent Read() calls on the same Multi instance.
//
//   - Writer Lifecycle: Close() only closes the input reader. Writers registered
//     via AddWriter are NOT automatically closed. Caller must manage writer
//     lifecycles independently to prevent resource leaks.
//
//   - Blocking Writers: A single slow/blocking writer will block all writes in
//     sequential mode. Use adaptive or parallel mode for mixed-latency scenarios.
//
//   - Error Propagation: First write error is returned immediately. Subsequent
//     writers may not receive the data if an earlier writer fails. Consider
//     error handling strategies for critical replication scenarios.
//
//   - Goroutine Overhead: Parallel mode spawns goroutines per write operation.
//     For high-frequency small writes, the overhead may exceed benefits. Use
//     MinimalSize threshold to limit parallel execution to large writes only.
//
// Best Practices:
//
// DO:
//   - Start with DefaultConfig() for sensible adaptive thresholds
//   - Use defer m.Close() to ensure input reader cleanup
//   - Monitor Stats() in production for latency/mode visibility
//   - Use adaptive mode for unknown or variable writer latencies
//   - Add all known writers at initialization when possible
//   - Use parallel mode for high-latency writers (network, slow disks)
//
// DON'T:
//   - Don't force parallel mode for tiny writes (goroutine overhead)
//   - Don't assume writers are closed by Multi.Close()
//   - Don't use blocking writers without considering parallel mode
//   - Don't check for nil after New() (always returns valid instance)
//   - Don't add writers incrementally in hot loops (impacts performance)
//
// # Performance Optimization Tips
//
// Sequential Mode Optimization:
//   - Zero allocations per write (uses io.MultiWriter)
//   - Best for low-latency writers (memory buffers, fast files)
//   - Ideal for small, frequent writes
//
// Parallel Mode Optimization:
//   - Allocates goroutines: best for large writes or slow writers
//   - Set MinimalSize to avoid overhead for small writes (default 512 bytes)
//   - Use when writer count >= 3 and latency is variable
//
// Adaptive Mode Optimization:
//   - Minimal overhead: <5% compared to sticky modes
//   - Automatically finds optimal strategy based on real measurements
//   - Adjust SampleWrite for faster/slower adaptation (default 100)
//   - Tune ThresholdLatency based on target latency requirements
//
// # Integration and Related Packages
//
// This package is part of github.com/nabbar/golib/ioutils and integrates
// with other I/O utilities in the golib ecosystem.
//
// Standard Library Integration:
//   - io.MultiWriter: Used internally for sequential write fan-out
//   - io.Copy: Provides Copy() semantics from reader to all writers
//   - sync/atomic: Lock-free concurrency primitives for performance
//   - time: Latency measurement for adaptive strategy decisions
//
// Related golib Packages:
//   - github.com/nabbar/golib/atomic: Typed atomic primitives (MapTyped, Value)
//   - github.com/nabbar/golib/ioutils/aggregator: Write aggregator for serialization
//
// External References:
//   - Go Concurrency Patterns: Pipeline and fan-out/fan-in techniques
//   - Effective Go: Interface compliance and error handling patterns
//
// # Testing and Quality
//
// The package includes comprehensive testing with:
//   - 80.8% code coverage (target: >80%)
//   - 120 test specifications using BDD methodology (Ginkgo v2 + Gomega)
//   - Race detector validation (zero races detected with -race flag)
//   - Performance benchmarks (8 aggregated experiments)
//   - Concurrent operation tests (AddWriter, Write, SetInput, Mixed)
//   - Edge case and error handling tests
//
// Test Categories:
//   - Constructor tests: Instance creation and interface compliance
//   - Reader tests: Input source management and Read operations
//   - Writer tests: Output management, broadcasting, and WriteString
//   - Copy tests: Stream replication and integration scenarios
//   - Concurrent tests: Thread-safety and race condition validation
//   - Mode tests: Sequential, parallel, and adaptive switching
//   - Edge case tests: Nil handling, large data, error propagation
//   - Benchmark tests: Performance validation across strategies
//
// For detailed test documentation, see TESTING.md in the package directory.
package multi
