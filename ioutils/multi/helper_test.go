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

package multi_test

import (
	"bytes"
	"io"
	"sync"
)

// closeErrorReader is a test helper that returns an error on Close.
// It wraps an io.Reader and allows testing of error propagation
// during close operations.
type closeErrorReader struct {
	io.Reader
	closeErr error
}

// Close implements io.Closer and returns the configured error.
func (e *closeErrorReader) Close() error {
	return e.closeErr
}

// errorReader is a test helper that always returns an error on Read.
// Useful for testing read error propagation.
type errorReader struct {
	err error
}

// Read implements io.Reader and always returns the configured error.
func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// errorWriter is a test helper that always returns an error on Write.
// Useful for testing write error propagation.
type errorWriter struct {
	err error
}

// Write implements io.Writer and always returns the configured error.
func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, e.err
}

// partialWriter is a test helper that writes only up to maxBytes bytes.
// Useful for testing partial write scenarios and error handling.
type partialWriter struct {
	maxBytes int
	written  int
}

// Write implements io.Writer with a maximum byte limit.
// Returns io.ErrShortWrite when the limit is reached or exceeded.
func (p *partialWriter) Write(data []byte) (n int, err error) {
	rem := p.maxBytes - p.written
	if rem <= 0 {
		return 0, io.ErrShortWrite
	}
	if len(data) > rem {
		p.written += rem
		return rem, io.ErrShortWrite
	}
	p.written += len(data)
	return len(data), nil
}

// errorReadCloser is a test helper that can return errors on both Read and Close.
// Allows comprehensive testing of error handling in read/close operations.
type errorReadCloser struct {
	io.Reader
	readErr  error
	closeErr error
}

// Read implements io.Reader with optional error injection.
func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	if e.readErr != nil {
		return 0, e.readErr
	}
	return e.Reader.Read(p)
}

// Close implements io.Closer with optional error injection.
func (e *errorReadCloser) Close() error {
	return e.closeErr
}

// safeBuffer is a thread-safe bytes.Buffer wrapper for concurrent testing.
// All methods are protected by a mutex to ensure safe concurrent access.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write implements io.Writer with thread-safe access.
func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

// Len returns the number of bytes buffered, thread-safe.
func (s *safeBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}

// String returns the buffer contents as a string, thread-safe.
func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}
