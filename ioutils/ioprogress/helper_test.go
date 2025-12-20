/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

// Package ioprogress_test provides common test helpers for the ioprogress package.
//
// This file contains shared test utilities to avoid code duplication across test files.
// All helpers are designed to be simple, reusable, and maintainable.
package ioprogress_test

import (
	"bytes"
	"io"
	"strings"
)

// closeableReader wraps strings.Reader with a Close() method for testing.
//
// This helper is necessary because strings.Reader doesn't implement io.ReadCloser,
// but the ioprogress package requires io.ReadCloser. The closed field allows tests
// to verify that Close() was called correctly.
//
// Thread Safety: Not thread-safe. Each test should create its own instance.
//
// Usage:
//
//	reader := newCloseableReader("test data")
//	defer reader.Close()
//	// ... perform operations
//	Expect(reader.closed).To(BeTrue())  // Verify Close() was called
type closeableReader struct {
	*strings.Reader
	closed bool // Tracks whether Close() has been called
}

// Close implements io.Closer for closeableReader.
//
// Sets the closed flag to true, allowing tests to verify Close() was invoked.
// Always returns nil (success) for simplicity in tests.
func (c *closeableReader) Close() error {
	c.closed = true
	return nil
}

// newCloseableReader creates a new closeableReader with the given string content.
//
// This is the primary factory function for creating test readers. The returned reader
// starts with closed=false and can be used immediately for Read() operations.
//
// Parameters:
//   - s: The string content to read from
//
// Returns:
//   - *closeableReader: A reader wrapping the string with Close() support
func newCloseableReader(s string) *closeableReader {
	return &closeableReader{
		Reader: strings.NewReader(s),
		closed: false,
	}
}

// closeableWriter wraps bytes.Buffer with a Close() method for testing.
//
// This helper is necessary because bytes.Buffer doesn't implement io.WriteCloser,
// but the ioprogress package requires io.WriteCloser. The closed field allows tests
// to verify that Close() was called correctly, and the embedded Buffer captures
// all written data for verification.
//
// Thread Safety: Not thread-safe. Each test should create its own instance.
//
// Usage:
//
//	writer := newCloseableWriter()
//	defer writer.Close()
//	writer.Write([]byte("data"))
//	Expect(writer.String()).To(Equal("data"))  // Verify written content
//	Expect(writer.closed).To(BeTrue())         // Verify Close() was called
type closeableWriter struct {
	*bytes.Buffer
	closed bool // Tracks whether Close() has been called
}

// Close implements io.Closer for closeableWriter.
//
// Sets the closed flag to true, allowing tests to verify Close() was invoked.
// The buffer contents remain accessible after closing for test verification.
// Always returns nil (success) for simplicity in tests.
func (c *closeableWriter) Close() error {
	c.closed = true
	return nil
}

// newCloseableWriter creates a new closeableWriter with an empty buffer.
//
// This is the primary factory function for creating test writers. The returned writer
// starts with an empty buffer and closed=false, ready for Write() operations.
//
// Returns:
//   - *closeableWriter: A writer wrapping an empty buffer with Close() support
func newCloseableWriter() *closeableWriter {
	return &closeableWriter{
		Buffer: &bytes.Buffer{},
		closed: false,
	}
}

// nopWriteCloser wraps an io.Writer to implement io.WriteCloser with a no-op Close().
//
// This helper is similar to io.NopCloser but for writers. It's useful when you need
// an io.WriteCloser but only have an io.Writer.
//
// Thread Safety: Depends on the underlying Writer.
//
// Usage:
//
//	var buf bytes.Buffer
//	writer := &nopWriteCloser{Writer: &buf}
//	defer writer.Close()  // No-op
type nopWriteCloser struct {
	io.Writer
}

// Close implements io.Closer for nopWriteCloser.
//
// This is a no-op implementation that always returns nil.
func (n *nopWriteCloser) Close() error {
	return nil
}
