/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package mapCloser_test

import (
	"context"
	"errors"
	"sync"
)

// Global test context for the entire test suite.
// Initialized in BeforeSuite and cancelled in AfterSuite.
var (
	globalTestCtx    context.Context
	globalTestCancel context.CancelFunc
)

// mockCloser is a simple io.Closer implementation for testing.
// It tracks whether Close() has been called and can be configured to return errors.
//
// Thread-safe: Uses mutex to protect concurrent access to closed flag.
type mockCloser struct {
	closed   bool       // Flag indicating if Close() has been called
	closeErr error      // Error to return from Close(), nil for success
	mu       sync.Mutex // Protects concurrent access to closed flag
}

// Close implements io.Closer interface.
// Records that Close() was called and returns the configured error (if any).
func (m *mockCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return m.closeErr
}

// IsClosed returns whether Close() has been called on this closer.
// Thread-safe.
func (m *mockCloser) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// newMockCloser creates a new mock closer that succeeds when closed.
func newMockCloser() *mockCloser {
	return &mockCloser{closed: false, closeErr: nil}
}

// newErrorCloser creates a new mock closer that returns the specified error when closed.
func newErrorCloser(err error) *mockCloser {
	return &mockCloser{closed: false, closeErr: err}
}

// Standard errors for testing
var (
	errTest1 = errors.New("test error 1")
	errTest2 = errors.New("test error 2")
	errTest3 = errors.New("test error 3")
)

// testCloserCount returns the number of closers that are actually closed.
// Useful for verifying that Close() was called on all closers.
func testCloserCount(closers ...*mockCloser) int {
	count := 0
	for _, c := range closers {
		if c.IsClosed() {
			count++
		}
	}
	return count
}

// createMockClosers creates n mock closers.
func createMockClosers(n int) []*mockCloser {
	closers := make([]*mockCloser, n)
	for i := 0; i < n; i++ {
		closers[i] = newMockCloser()
	}
	return closers
}
