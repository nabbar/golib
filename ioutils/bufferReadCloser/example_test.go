/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package bufferReadCloser_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nabbar/golib/ioutils/bufferReadCloser"
)

// Example_basic demonstrates the simplest use case: wrapping a bytes.Buffer
// with automatic cleanup on close.
func Example_basic() {
	// Create a buffer with some data
	buf := bytes.NewBufferString("Hello, World!")

	// Wrap it with Close support
	wrapped := bufferReadCloser.NewBuffer(buf, nil)
	defer wrapped.Close() // Automatically resets buffer

	// Read the data
	data := make([]byte, 5)
	n, _ := wrapped.Read(data)
	fmt.Printf("Read %d bytes: %s\n", n, string(data))

	// Output:
	// Read 5 bytes: Hello
}

// Example_nilHandling shows how the package handles nil parameters gracefully
// by providing sensible defaults instead of panicking.
func Example_nilHandling() {
	// Passing nil creates an empty buffer
	buf := bufferReadCloser.NewBuffer(nil, nil)
	defer buf.Close()

	// Can write to it normally
	buf.WriteString("test")

	// And read back
	data := make([]byte, 4)
	buf.Read(data)
	fmt.Println(string(data))

	// Output:
	// test
}

// Example_customClose demonstrates using a custom close function to perform
// additional cleanup beyond the default reset behavior.
func Example_customClose() {
	closed := false

	// Create buffer with custom close function
	buf := bytes.NewBuffer(nil)
	wrapped := bufferReadCloser.NewBuffer(buf, func() error {
		closed = true
		fmt.Println("Custom cleanup executed")
		return nil
	})

	wrapped.WriteString("data")
	wrapped.Close()

	fmt.Printf("Closed: %v\n", closed)

	// Output:
	// Custom cleanup executed
	// Closed: true
}

// Example_reader shows buffered reading with automatic resource cleanup.
// This is useful when reading from files or network connections.
func Example_reader() {
	// Simulate a data source
	source := strings.NewReader("Line 1\nLine 2\nLine 3")
	br := bufio.NewReader(source)

	// Wrap with close support
	reader := bufferReadCloser.NewReader(br, nil)
	defer reader.Close() // Automatically resets reader

	// Read some data
	data := make([]byte, 6)
	n, _ := reader.Read(data)
	fmt.Printf("Read: %s\n", string(data[:n]))

	// Output:
	// Read: Line 1
}

// Example_writer demonstrates buffered writing with automatic flush on close.
// This ensures all buffered data is written before cleanup.
func Example_writer() {
	// Create destination buffer
	dest := &bytes.Buffer{}
	bw := bufio.NewWriter(dest)

	// Wrap with close support
	writer := bufferReadCloser.NewWriter(bw, nil)

	// Write data (buffered, not yet visible in dest)
	writer.WriteString("Hello")
	writer.WriteString(" ")
	writer.WriteString("World")

	// Close flushes automatically
	writer.Close()

	fmt.Println(dest.String())

	// Output:
	// Hello World
}

// Example_writerFlushError shows how flush errors are properly returned
// from the Close() method, allowing proper error handling.
func Example_writerFlushError() {
	// Create a writer that will fail on flush
	failWriter := &failingWriter{}
	bw := bufio.NewWriter(failWriter)

	writer := bufferReadCloser.NewWriter(bw, nil)
	writer.WriteString("data")

	// Close returns the flush error
	err := writer.Close()
	if err != nil {
		fmt.Println("Flush error:", err)
	}

	// Output:
	// Flush error: write failed
}

// failingWriter is a helper type for demonstrating error handling
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write failed")
}

// Example_readWriter demonstrates bidirectional buffered I/O with automatic
// flush on close. Useful for network protocols or duplex communication.
func Example_readWriter() {
	// Create separate buffers for reading and writing to simulate duplex I/O
	readBuf := bytes.NewBufferString("Initial data")
	writeBuf := &bytes.Buffer{}

	// Create ReadWriter with separate read/write buffers
	rw := bufio.NewReadWriter(bufio.NewReader(readBuf), bufio.NewWriter(writeBuf))
	wrapped := bufferReadCloser.NewReadWriter(rw, nil)
	defer wrapped.Close() // Flushes writes

	// Read from input
	data := make([]byte, 7)
	wrapped.Read(data)
	fmt.Printf("Read: %s\n", string(data))

	// Write to output (buffered)
	wrapped.WriteString("Response data")
	wrapped.Close() // Flush to output buffer

	fmt.Printf("Written: %s\n", writeBuf.String())

	// Output:
	// Read: Initial
	// Written: Response data
}

// Example_fileProcessing shows a realistic use case: reading a file with
// automatic cleanup of both the buffer and file handle.
func Example_fileProcessing() {
	// Simulate file content
	fileContent := strings.NewReader("File content here")

	// In real code, this would be: file, _ := os.Open("data.txt")
	// We simulate with a ReadCloser
	file := io.NopCloser(fileContent)

	// Create buffered reader with file close chained
	br := bufio.NewReader(file)
	reader := bufferReadCloser.NewReader(br, file.Close)
	defer reader.Close() // Closes both reader and file

	// Read data
	data := make([]byte, 12)
	n, _ := reader.Read(data)
	fmt.Printf("Read: %s\n", string(data[:n]))

	// Output:
	// Read: File content
}

// Example_bufferPool demonstrates integration with sync.Pool for efficient
// buffer reuse with automatic reset and return to pool.
func Example_bufferPool() {
	// Simulate a buffer pool (in real code, use sync.Pool)
	pool := &simplePool{buf: bytes.NewBuffer(make([]byte, 0, 1024))}

	// Get buffer from pool
	buf := pool.Get()

	// Wrap with custom close that returns to pool
	wrapped := bufferReadCloser.NewBuffer(buf, func() error {
		pool.Put(buf)
		return nil
	})

	// Use buffer
	wrapped.WriteString("temporary data")

	// Close resets and returns to pool
	wrapped.Close()

	// Verify buffer was reset
	fmt.Printf("Buffer length after close: %d\n", buf.Len())

	// Output:
	// Buffer length after close: 0
}

// simplePool is a simplified pool for demonstration
type simplePool struct {
	buf *bytes.Buffer
}

func (p *simplePool) Get() *bytes.Buffer {
	return p.buf
}

func (p *simplePool) Put(b *bytes.Buffer) {
	// In real sync.Pool, this would return to pool
}

// Example_chainingCleanup shows how to chain multiple cleanup operations
// using custom close functions.
func Example_chainingCleanup() {
	var cleanupOrder []string

	// Create nested cleanup chain
	buf := bytes.NewBuffer(nil)
	wrapped := bufferReadCloser.NewBuffer(buf, func() error {
		cleanupOrder = append(cleanupOrder, "buffer cleanup")
		return nil
	})

	// Simulate additional resource
	resource := &mockResource{
		onClose: func() error {
			cleanupOrder = append(cleanupOrder, "resource cleanup")
			return nil
		},
	}

	// Use both
	wrapped.WriteString("data")

	// Close in order
	wrapped.Close()
	resource.Close()

	fmt.Printf("Cleanup order: %v\n", cleanupOrder)

	// Output:
	// Cleanup order: [buffer cleanup resource cleanup]
}

// mockResource simulates an external resource with cleanup
type mockResource struct {
	onClose func() error
}

func (r *mockResource) Close() error {
	if r.onClose != nil {
		return r.onClose()
	}
	return nil
}

// Example_errorPropagation demonstrates how errors from custom close functions
// are properly propagated to the caller.
func Example_errorPropagation() {
	// Create buffer with failing close function
	buf := bytes.NewBuffer(nil)
	wrapped := bufferReadCloser.NewBuffer(buf, func() error {
		return fmt.Errorf("cleanup failed")
	})

	wrapped.WriteString("data")

	// Error is returned from Close()
	err := wrapped.Close()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Output:
	// Error: cleanup failed
}

// Example_writerToDiscard shows how a nil writer creates a writer to io.Discard,
// useful for testing or when output needs to be silently discarded.
func Example_writerToDiscard() {
	// Passing nil creates writer to io.Discard
	writer := bufferReadCloser.NewWriter(nil, nil)
	defer writer.Close()

	// Writes succeed but data is discarded
	n, err := writer.WriteString("discarded data")
	fmt.Printf("Wrote %d bytes, error: %v\n", n, err)

	// Output:
	// Wrote 14 bytes, error: <nil>
}

// Example_readerEOF shows how a nil reader creates a reader that immediately
// returns EOF, useful for testing or empty data scenarios.
func Example_readerEOF() {
	// Passing nil creates reader from empty source
	reader := bufferReadCloser.NewReader(nil, nil)
	defer reader.Close()

	// Read immediately returns EOF
	data := make([]byte, 10)
	n, err := reader.Read(data)
	fmt.Printf("Read %d bytes, EOF: %v\n", n, err == io.EOF)

	// Output:
	// Read 0 bytes, EOF: true
}
