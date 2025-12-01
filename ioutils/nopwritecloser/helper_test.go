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

package nopwritecloser_test

import (
	"bytes"
	"sync"
)

// safeBuffer wraps bytes.Buffer with a mutex for thread-safe operations.
// This helper is used in concurrency tests to safely share a buffer
// between multiple goroutines without data races.
type safeBuffer struct {
	mu  sync.Mutex   // Protects concurrent access to buf
	buf bytes.Buffer // Underlying buffer for data storage
}

// Write implements io.Writer in a thread-safe manner.
// It acquires the mutex before delegating to the underlying buffer.
func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

// Len returns the number of bytes in the buffer in a thread-safe manner.
func (s *safeBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}

// String returns the buffer contents as a string in a thread-safe manner.
func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// errorWriter is a test helper that always returns an error on write.
// It is used to verify error propagation from the underlying writer.
type errorWriter struct {
	err error // The error to return on every write attempt
}

// Write implements io.Writer and always returns the configured error.
func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, e.err
}

// limitedErrorWriter is a test helper that succeeds for N writes, then errors.
// It is used to test error handling after a series of successful writes.
type limitedErrorWriter struct {
	remaining int          // Number of successful writes remaining
	err       error        // Error to return once remaining reaches 0
	buf       bytes.Buffer // Buffer for successful writes
}

// Write implements io.Writer and returns an error after the configured number of writes.
func (l *limitedErrorWriter) Write(p []byte) (n int, err error) {
	if l.remaining <= 0 {
		return 0, l.err
	}
	l.remaining--
	return l.buf.Write(p)
}

// countingWriter is a test helper that counts the number of write calls.
// It is used to verify that writes are properly delegated to the underlying writer.
type countingWriter struct {
	count int          // Number of Write() calls received
	buf   bytes.Buffer // Buffer for storing written data
}

// Write implements io.Writer and increments the call counter before writing.
func (c *countingWriter) Write(p []byte) (n int, err error) {
	c.count++
	return c.buf.Write(p)
}
