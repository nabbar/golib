/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

// Package bufferReadCloser provides lightweight wrappers around Go's standard
// buffered I/O types (bytes.Buffer, bufio.Reader, bufio.Writer, bufio.ReadWriter)
// that add io.Closer support with automatic resource cleanup and custom close callbacks.
//
// # Design Philosophy
//
// The package follows these core principles:
//
//  1. Minimal Overhead: Thin wrappers with zero-copy passthrough to underlying buffers
//  2. Lifecycle Management: Automatic reset and cleanup on close
//  3. Flexibility: Optional custom close functions for additional cleanup logic
//  4. Standard Compatibility: Implements all relevant io.* interfaces
//  5. Defensive Programming: Provides sensible defaults when nil parameters are passed
//
// # Architecture
//
// The package provides four main wrapper types:
//
//	┌─────────────────────────────────────────────────┐
//	│          bufferReadCloser Package               │
//	└─────────────────┬───────────────────────────────┘
//	                  │
//	     ┌────────────┼────────────┬─────────────┐
//	     │            │            │             │
//	┌────▼─────┐ ┌───▼────┐ ┌────▼─────┐ ┌─────▼────────┐
//	│  Buffer  │ │ Reader │ │  Writer  │ │ ReadWriter   │
//	├──────────┤ ├────────┤ ├──────────┤ ├──────────────┤
//	│bytes.    │ │bufio.  │ │bufio.    │ │bufio.        │
//	│Buffer    │ │Reader  │ │Writer    │ │ReadWriter    │
//	│    +     │ │   +    │ │    +     │ │      +       │
//	│io.Closer │ │io.     │ │io.Closer │ │io.Closer     │
//	│          │ │Closer  │ │          │ │              │
//	└──────────┘ └────────┘ └──────────┘ └──────────────┘
//
// Each wrapper delegates all I/O operations directly to the underlying buffer type,
// ensuring zero performance overhead. The Close() method performs cleanup specific
// to each type and optionally calls a custom close function.
//
// # Wrapper Behavior
//
// Buffer (bytes.Buffer wrapper):
//   - On Close: Resets buffer (clears all data) + calls custom close
//   - Nil handling: Creates empty buffer
//   - Use case: In-memory read/write with lifecycle management
//
// Reader (bufio.Reader wrapper):
//   - On Close: Resets reader (releases resources) + calls custom close
//   - Nil handling: Creates reader from empty source (returns EOF)
//   - Use case: Buffered reading with automatic cleanup
//
// Writer (bufio.Writer wrapper):
//   - On Close: Flushes buffered data + resets writer + calls custom close
//   - Nil handling: Creates writer to io.Discard
//   - Use case: Buffered writing with guaranteed flush
//
// ReadWriter (bufio.ReadWriter wrapper):
//   - On Close: Flushes buffered data + calls custom close (no reset due to API limitation)
//   - Nil handling: Creates readwriter with empty source and io.Discard destination
//   - Use case: Bidirectional buffered I/O
//   - Limitation: Cannot call Reset() due to ambiguous methods in bufio.ReadWriter
//
// # Advantages
//
//   - Single defer statement handles both buffer cleanup and resource closing
//   - Prevents resource leaks by ensuring cleanup always occurs
//   - Composable: Custom close functions enable chaining of cleanup operations
//   - Type-safe: Preserves all standard io.* interfaces
//   - Zero dependencies: Only uses standard library
//   - Defensive: Handles nil parameters gracefully with sensible defaults
//
// # Disadvantages and Limitations
//
//   - Not thread-safe: Like stdlib buffers, requires external synchronization
//   - ReadWriter limitation: Cannot reset on close due to ambiguous Reset methods
//   - Memory overhead: 24 bytes per wrapper (pointer + function pointer)
//   - Nil parameter handling: Creates default instances which may not be desired behavior
//
// # Performance Characteristics
//
//   - Zero-copy operations: All I/O delegates directly to underlying buffers
//   - Minimal allocation: Single wrapper struct per buffer
//   - No additional buffering: Uses existing bufio buffers
//   - Constant memory: O(1) overhead regardless of data size
//   - Inline-friendly: Method calls are often inlined by compiler
//
// # Typical Use Cases
//
// File Processing with Automatic Cleanup:
//
//	file, _ := os.Open("data.txt")
//	reader := bufferReadCloser.NewReader(bufio.NewReader(file), file.Close)
//	defer reader.Close() // Closes both reader and file
//
// Network Connection Management:
//
//	conn, _ := net.Dial("tcp", "example.com:80")
//	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
//	wrapper := bufferReadCloser.NewReadWriter(rw, conn.Close)
//	defer wrapper.Close() // Flushes and closes connection
//
// Buffer Pool Integration:
//
//	buf := bufferPool.Get().(*bytes.Buffer)
//	wrapped := bufferReadCloser.NewBuffer(buf, func() error {
//	    bufferPool.Put(buf)
//	    return nil
//	})
//	defer wrapped.Close() // Resets and returns to pool
//
// Testing with Lifecycle Tracking:
//
//	tracker := &TestTracker{}
//	buf := bufferReadCloser.NewBuffer(bytes.NewBuffer(nil), tracker.OnClose)
//	defer buf.Close()
//	// Test code...
//	// tracker.Closed will be true after Close()
//
// # Error Handling
//
// Close operations may return errors from:
//   - Flush operations (Writer, ReadWriter): If buffered data cannot be written
//   - Custom close functions: Any error returned by the FuncClose callback
//
// The package follows Go conventions: errors are returned, never panicked.
// When nil parameters are provided, sensible defaults are created instead of panicking.
//
// # Thread Safety
//
// Like the underlying stdlib types, these wrappers are NOT thread-safe.
// Concurrent access requires external synchronization (e.g., sync.Mutex).
//
// # Minimum Go Version
//
// This package requires Go 1.18 or later. All functions used are from the
// standard library and have been stable since Go 1.0-1.2.
package bufferReadCloser
