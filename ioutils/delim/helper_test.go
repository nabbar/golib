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

package delim_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync/atomic"

	libbuf "github.com/nabbar/golib/ioutils/bufferReadCloser"

	. "github.com/onsi/ginkgo/v2"
)

// Test context management
// Global test context with initialization and cancellation support
var (
	testCtx    context.Context
	testCancel context.CancelFunc
	closedBuf  = new(atomic.Bool)
)

// initTestContext initializes the global test context.
// This function should be called in BeforeSuite to create a parent context
// for all tests, allowing proper cleanup and cancellation.
func initTestContext() {
	testCtx, testCancel = context.WithCancel(context.Background())
}

// cleanupTestContext cleans up the global test context.
// This function should be called in AfterSuite to ensure proper
// cancellation and resource cleanup.
func cleanupTestContext() {
	if testCancel != nil {
		testCancel()
	}
}

// getTestContext returns a child context derived from the global test context.
// Each test should use this to get its own cancellable context.
func getTestContext() context.Context {
	if testCtx == nil {
		GinkgoT().Log("Warning: test context not initialized, using background context")
		return context.Background()
	}
	ctx, _ := context.WithCancel(testCtx)
	return ctx
}

// errorReader is a test helper that simulates read errors.
// It returns an error after a specified number of successful reads.
// This is useful for testing error handling and recovery mechanisms.
type errorReader struct {
	data      *strings.Reader
	errorOn   int
	readCount int
}

// newErrorReader creates a new errorReader that will fail after errorOn reads.
//
// Parameters:
//   - data: The string data to read from
//   - errorOn: Number of successful reads before returning an error (0 = immediate error)
//
// Example:
//
//	er := newErrorReader("test data", 2)  // Will succeed 2 times, then fail
//	r := io.NopCloser(er)
func newErrorReader(data string, errorOn int) *errorReader {
	return &errorReader{
		data:    strings.NewReader(data),
		errorOn: errorOn,
	}
}

// Read implements io.Reader, returning an error after errorOn reads.
func (r *errorReader) Read(p []byte) (n int, err error) {
	r.readCount++
	if r.readCount >= r.errorOn {
		return 0, errors.New("simulated read error")
	}
	return r.data.Read(p)
}

// Close implements io.Closer (no-op).
func (r *errorReader) Close() error {
	return nil
}

// errorWriter is a test helper that simulates write errors.
// It returns an error after a specified number of successful writes.
// This is useful for testing error propagation in write operations.
type errorWriter struct {
	buf        *bytes.Buffer
	errorAfter int
	writeCount int
}

// newErrorWriter creates a new errorWriter that will fail after errorAfter writes.
//
// Parameters:
//   - errorAfter: Number of successful writes before returning an error (0 = immediate error)
//
// Example:
//
//	ew := newErrorWriter(3)  // Will succeed 3 times, then fail
func newErrorWriter(errorAfter int) *errorWriter {
	return &errorWriter{
		buf:        &bytes.Buffer{},
		errorAfter: errorAfter,
	}
}

// Write implements io.Writer, returning an error after errorAfter writes.
func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount > w.errorAfter {
		return 0, errors.New("simulated write error")
	}
	return w.buf.Write(p)
}

// String returns the buffered data as a string.
func (w *errorWriter) String() string {
	return w.buf.String()
}

// Bytes returns the buffered data as a byte slice.
func (w *errorWriter) Bytes() []byte {
	return w.buf.Bytes()
}

// Len returns the number of bytes buffered.
func (w *errorWriter) Len() int {
	return w.buf.Len()
}

// testDataGenerator provides methods to generate test data with various characteristics.
type testDataGenerator struct{}

// newTestDataGenerator creates a new test data generator.
func newTestDataGenerator() *testDataGenerator {
	return &testDataGenerator{}
}

// simpleLines generates simple line-delimited data.
//
// Parameters:
//   - count: Number of lines to generate
//   - prefix: Prefix for each line
//
// Returns: String with newline-delimited data
func (g *testDataGenerator) simpleLines(count int, prefix string) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		b.WriteString(prefix)
		b.WriteString("\n")
	}
	return b.String()
}

// csvData generates CSV-like data with comma delimiter.
//
// Parameters:
//   - rows: Number of rows
//   - cols: Number of columns per row
//
// Returns: String with CSV data
func (g *testDataGenerator) csvData(rows, cols int) string {
	var b strings.Builder
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			b.WriteString("field")
			if j < cols-1 {
				b.WriteString(",")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// binaryData generates binary data with specified delimiter.
//
// Parameters:
//   - blocks: Number of data blocks
//   - delimiter: Delimiter byte
//
// Returns: Byte slice with binary data
func (g *testDataGenerator) binaryData(blocks int, delimiter byte) []byte {
	var b bytes.Buffer
	for i := 0; i < blocks; i++ {
		b.Write([]byte{0x00, 0xFF, 0xDE, 0xAD, 0xBE, 0xEF})
		b.WriteByte(delimiter)
	}
	return b.Bytes()
}

// largeData generates large data for stress testing.
//
// Parameters:
//   - sizeKB: Approximate size in kilobytes
//   - lineLength: Average line length
//
// Returns: String with large data set
func (g *testDataGenerator) largeData(sizeKB int, lineLength int) string {
	line := strings.Repeat("x", lineLength) + "\n"
	linesNeeded := (sizeKB * 1024) / len(line)
	return strings.Repeat(line, linesNeeded)
}

// unicodeData generates data with Unicode characters.
//
// Parameters:
//   - count: Number of lines
//
// Returns: String with Unicode data
func (g *testDataGenerator) unicodeData(count int) string {
	var b strings.Builder
	unicodeChars := []string{"α", "β", "γ", "δ", "ε", "ζ", "η", "θ", "€", "£", "¥", "©"}
	for i := 0; i < count; i++ {
		b.WriteString(unicodeChars[i%len(unicodeChars)])
		b.WriteString("\n")
	}
	return b.String()
}

// mixedDelimiters generates data with mixed delimiter types for testing.
//
// Parameters:
//   - count: Number of entries
//   - delimiters: Slice of delimiter characters to use
//
// Returns: String with mixed delimiter data
func (g *testDataGenerator) mixedDelimiters(count int, delimiters []rune) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		b.WriteString("data")
		delim := delimiters[i%len(delimiters)]
		b.WriteRune(delim)
	}
	return b.String()
}

// readerCloserWrapper wraps an io.Reader as io.ReadCloser for testing.
type readerCloserWrapper struct {
	io.Reader
	closed bool
}

// newReaderCloser creates a new io.ReadCloser from an io.Reader.
// The Close method is a no-op but tracks if it was called.
func newReaderCloser(r io.Reader) *readerCloserWrapper {
	return &readerCloserWrapper{Reader: r}
}

// Close implements io.Closer and tracks close calls.
func (r *readerCloserWrapper) Close() error {
	r.closed = true
	return nil
}

// IsClosed returns whether Close was called.
func (r *readerCloserWrapper) IsClosed() bool {
	return r.closed
}

// assertNoError is a helper to fail the test if an unexpected error occurs.
// This provides cleaner test code by avoiding repetitive error checking.
func assertNoError(err error, msg string) {
	if err != nil {
		GinkgoT().Fatalf("%s: %v", msg, err)
	}
}

// assertError is a helper to fail the test if an expected error doesn't occur.
func assertError(err error, msg string) {
	if err == nil {
		GinkgoT().Fatalf("%s: expected error but got nil", msg)
	}
}

// newClosableBuffer creates a new io.ReadCloser that returns error after Close().
// After Close() is called, any subsequent Read() will return io.ErrClosedPipe.
func newClosableBuffer(data string) io.ReadCloser {
	closedBuf.Store(false)
	return libbuf.NewBuffer(bytes.NewBuffer([]byte(data)), func() error {
		o := closedBuf.Swap(true)
		if o {
			return os.ErrClosed
		} else {
			return nil
		}
	})
}

// mockReader0Nil simulates a Reader that returns (0, nil) on first call, then EOF.
type mockReader0Nil struct {
	called bool
}

func (m *mockReader0Nil) Read(p []byte) (n int, err error) {
	if !m.called {
		m.called = true
		return 0, nil
	}
	return 0, io.EOF
}

// mockReaderError simulates a Reader that returns a specific error.
type mockReaderError struct {
	err error
}

func (m *mockReaderError) Read(p []byte) (n int, err error) {
	return 0, m.err
}

// mockReaderEOFData returns data and EOF in the same Read call
type mockReaderEOFData struct {
	data string
	done bool
}

func (m *mockReaderEOFData) Read(p []byte) (n int, err error) {
	if m.done {
		return 0, io.EOF
	}
	m.done = true
	if len(p) < len(m.data) {
		copy(p, []byte(m.data))
		return len(p), io.EOF // Partial read with EOF
	}
	copy(p, []byte(m.data))
	return len(m.data), io.EOF
}

func (m *mockReaderEOFData) Close() error {
	return nil
}

// transientReader simulates a temporary error followed by successful read.
type transientReader struct {
	data []byte
	err  error
	call int
}

func (r *transientReader) Read(p []byte) (n int, err error) {
	r.call++
	if r.call == 1 {
		// First call: return partial data and error
		if len(p) < len(r.data) {
			copy(p, r.data)
			return len(p), r.err
		}
		copy(p, r.data)
		return len(r.data), r.err
	}
	// Second call: return delimiter
	if len(p) > 0 {
		p[0] = '\n'
		return 1, nil
	}
	return 0, nil
}

func (r *transientReader) Close() error {
	return nil
}
