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

// Package multi provides a thread-safe multiplexer for io.Reader and io.Writer operations.
//
// This package implements a concurrent-safe I/O multiplexer that allows:
//   - Broadcasting writes to multiple io.Writer destinations simultaneously
//   - Managing a single input io.ReadCloser source
//   - Thread-safe operations using sync.Map and atomic.Value
//   - Seamless integration with Go's io package interfaces
//
// # Key Features
//
//   - Thread-safe concurrent write operations to multiple destinations
//   - Dynamic addition and removal of write destinations
//   - Single input source management with thread-safe replacement
//   - Built-in support for io.Copy operations from input to all outputs
//   - Implements io.ReadWriteCloser and io.StringWriter interfaces
//   - Zero-allocation read/write operations in steady state
//
// # Basic Usage
//
// Creating a new multiplexer and broadcasting writes:
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
// # Implementation Details
//
// The Multi type uses sync/atomic.Value to ensure thread-safe access to
// the current reader and writer without locks on the read/write path.
// Writers are stored in a sync.Map for concurrent-safe addition and removal.
//
// The implementation guarantees that:
//   - atomic.Value always stores consistent types (via wrappers)
//   - io.MultiWriter is used for all write operations (even single writers)
//   - Default initialization uses DiscardCloser to prevent nil panics
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
// # Limitations and Best Practices
//
// While the Multi type itself is thread-safe for all operations, the underlying
// io.ReadCloser set via SetInput may not support concurrent reads. For safe
// concurrent usage:
//   - Use external synchronization when reading from the same Multi instance
//     across multiple goroutines
//   - Writers registered via AddWriter should be safe for concurrent writes,
//     or alternatively use the safeBuffer pattern for synchronization
//   - Close() only closes the input reader, not the registered writers -
//     caller must manage writer lifecycles independently
//
// # Performance Considerations
//
// The implementation is designed for minimal allocation overhead:
//   - Zero-allocation read/write operations in steady state
//   - Atomic operations avoid mutex contention on hot paths
//   - Writers are collected and combined only during AddWriter/Clean operations
//
// For optimal performance:
//   - Add all known writers at initialization rather than incrementally
//   - Reuse Multi instances when possible
//   - Consider the overhead of io.MultiWriter for single-writer scenarios
//
// # Integration
//
// This package is part of github.com/nabbar/golib/ioutils and integrates
// with other I/O utilities in the golib ecosystem.
//
// See also:
//   - io.MultiWriter for the underlying write broadcasting mechanism
//   - io.Copy for the data copying semantics
//   - sync/atomic for the concurrency guarantees
package multi
