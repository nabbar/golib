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
	"context"
	"sync"
	"sync/atomic"
)

// testContext creates a global test context with cancellation.
// This is used across all test suites to ensure proper cleanup.
var (
	testCtx    context.Context
	testCancel context.CancelFunc
	testOnce   sync.Once
)

// getTestContext returns the global test context, initializing it if needed.
func getTestContext() context.Context {
	testOnce.Do(func() {
		testCtx, testCancel = context.WithCancel(context.Background())
	})
	return testCtx
}

// cancelTestContext cancels the global test context.
// This should be called at the end of the test suite.
func cancelTestContext() {
	if testCancel != nil {
		testCancel()
	}
}

// concurrentCounter is a helper for tracking concurrent operations safely.
type concurrentCounter struct {
	count int64
}

// inc increments the counter atomically.
func (c *concurrentCounter) inc() {
	atomic.AddInt64(&c.count, 1)
}

// dec decrements the counter atomically.
func (c *concurrentCounter) dec() {
	atomic.AddInt64(&c.count, -1)
}

// get returns the current counter value atomically.
func (c *concurrentCounter) get() int64 {
	return atomic.LoadInt64(&c.count)
}

// reset resets the counter to zero atomically.
func (c *concurrentCounter) reset() {
	atomic.StoreInt64(&c.count, 0)
}

// concurrentRunner executes a function concurrently n times and waits for completion.
func concurrentRunner(n int, fn func(id int)) {
	var wg sync.WaitGroup
	wg.Add(n)
	
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			fn(id)
		}(i)
	}
	
	wg.Wait()
}

// generateTestData generates test data of specified size.
func generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}
