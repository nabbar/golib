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

// This file contains helper functions and global test context for the iowrapper test suite.
//
// Purpose:
// This file centralizes reusable test utilities to:
//   - Reduce code duplication across test files
//   - Provide consistent test patterns (concurrency, counting, transformations)
//   - Manage global test context lifecycle
//
// It provides:
//   - Global test context with initialization (BeforeSuite) and cleanup (AfterSuite)
//   - Reader/Writer factory functions (newTestReader, newTestBuffer)
//   - Concurrency helpers (runConcurrent, runConcurrentIndexed)
//   - Custom function builders (makeCustomReadFunc, makeCustomWriteFunc)
//   - Common transformations (toUppercase, toLowercase)
//   - Counting wrappers for monitoring (newCountingReader, newCountingWriter)
//
// Usage:
// Import this file implicitly by using the test package. Helpers are available
// in all test files within the iowrapper_test package.
//
// Best Practices:
//   - Use runConcurrent for simple parallel execution without indices
//   - Use runConcurrentIndexed when goroutines need unique identifiers
//   - Use newAtomicCounter for thread-safe counting in custom functions
//   - Use makeCustomReadFunc/makeCustomWriteFunc for consistent function patterns
package iowrapper_test

import (
	"bytes"
	"context"
	"io"
	"sync"
	"sync/atomic"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
)

var (
	// testCtx is the global test context used across all test specs.
	// It is initialized in BeforeSuite and canceled in AfterSuite.
	testCtx context.Context

	// testCancel is the cancel function for the global test context.
	testCancel context.CancelFunc
)

// Initialize global test context before running any specs.
var _ = BeforeSuite(func() {
	// Create global test context with cancellation
	testCtx, testCancel = context.WithCancel(context.Background())
})

// Cleanup global test context after all specs have run.
var _ = AfterSuite(func() {
	// Cancel global test context to cleanup resources
	if testCancel != nil {
		testCancel()
	}
})

// newTestReader creates a new bytes.Reader with the given data for testing.
// This is a helper to avoid code duplication across test files.
func newTestReader(data string) io.Reader {
	return bytes.NewReader([]byte(data))
}

// newTestBuffer creates a new bytes.Buffer with the given data for testing.
// This is a helper to avoid code duplication across test files.
func newTestBuffer(data string) *bytes.Buffer {
	return bytes.NewBufferString(data)
}

// runConcurrent executes fn concurrently n times and waits for completion.
// This helper simplifies concurrent testing by handling WaitGroup management.
//
// Parameters:
//   - n: number of concurrent goroutines to spawn
//   - fn: function to execute concurrently
//
// Example usage:
//
//	runConcurrent(100, func() {
//	    wrapper.Read(buffer)
//	})
func runConcurrent(n int, fn func()) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	wg.Wait()
}

// runConcurrentIndexed executes fn concurrently n times with an index parameter.
// This helper simplifies concurrent testing when the goroutine needs an index.
//
// Parameters:
//   - n: number of concurrent goroutines to spawn
//   - fn: function to execute concurrently, receives the goroutine index
//
// Example usage:
//
//	runConcurrentIndexed(100, func(i int) {
//	    results[i] = wrapper.Read(buffers[i])
//	})
func runConcurrentIndexed(n int, fn func(int)) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			fn(idx)
		}()
	}
	wg.Wait()
}

// newAtomicCounter creates a new atomic counter for thread-safe counting in tests.
// Returns a pointer to atomic.Int64 initialized to 0.
func newAtomicCounter() *atomic.Int64 {
	var cnt atomic.Int64
	return &cnt
}

// makeCustomReadFunc creates a custom read function that reads from the given reader
// and applies an optional transformation function.
//
// Parameters:
//   - r: underlying reader to read from
//   - transform: optional transformation function (can be nil)
//
// Returns a FuncRead that can be used with SetRead.
func makeCustomReadFunc(r io.Reader, transform func([]byte) []byte) FuncRead {
	return func(p []byte) []byte {
		n, err := r.Read(p)
		if err != nil || n == 0 {
			return nil
		}
		data := p[:n]
		if transform != nil {
			data = transform(data)
		}
		return data
	}
}

// makeCustomWriteFunc creates a custom write function that writes to the given writer
// and applies an optional transformation function.
//
// Parameters:
//   - w: underlying writer to write to
//   - transform: optional transformation function (can be nil)
//
// Returns a FuncWrite that can be used with SetWrite.
func makeCustomWriteFunc(w io.Writer, transform func([]byte) []byte) FuncWrite {
	return func(p []byte) []byte {
		data := p
		if transform != nil {
			data = transform(p)
		}
		n, err := w.Write(data)
		if err != nil {
			return nil
		}
		return p[:n]
	}
}

// toUppercase transforms bytes to uppercase (ASCII only).
// This is a common transformation used in multiple tests.
func toUppercase(p []byte) []byte {
	result := make([]byte, len(p))
	for i, b := range p {
		if b >= 'a' && b <= 'z' {
			result[i] = b - 32
		} else {
			result[i] = b
		}
	}
	return result
}

// toLowercase transforms bytes to lowercase (ASCII only).
// This is a common transformation used in multiple tests.
func toLowercase(p []byte) []byte {
	result := make([]byte, len(p))
	for i, b := range p {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return result
}

// newCountingReader creates an IOWrapper that counts read operations.
// Returns the wrapper and a pointer to the atomic counter.
func newCountingReader(r io.Reader) (IOWrapper, *atomic.Int64) {
	wrapper := New(r)
	cnt := newAtomicCounter()

	wrapper.SetRead(func(p []byte) []byte {
		cnt.Add(1)
		n, err := r.Read(p)
		if err != nil || n == 0 {
			return nil
		}
		return p[:n]
	})

	return wrapper, cnt
}

// newCountingWriter creates an IOWrapper that counts write operations.
// Returns the wrapper and a pointer to the atomic counter.
func newCountingWriter(w io.Writer) (IOWrapper, *atomic.Int64) {
	wrapper := New(w)
	cnt := newAtomicCounter()

	wrapper.SetWrite(func(p []byte) []byte {
		cnt.Add(1)
		n, err := w.Write(p)
		if err != nil {
			return nil
		}
		return p[:n]
	})

	return wrapper, cnt
}
