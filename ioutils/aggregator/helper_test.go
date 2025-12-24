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
 */

package aggregator_test

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/gomega"
)

// Helper functions for testing

// testWriter is a thread-safe writer implementation that captures all writes.
// It provides configurable failure and delay behavior for testing edge cases.
type testWriter struct {
	mu      sync.Mutex
	data    [][]byte
	calls   atomic.Int32
	failAt  int32 // fail at this call number (0 = never fail)
	delayMs int   // delay each write by this many milliseconds
}

// newTestWriter creates a new testWriter instance.
func newTestWriter() *testWriter {
	return &testWriter{
		data: make([][]byte, 0),
	}
}

// Write implements io.Writer interface with optional failure and delay.
func (w *testWriter) Write(p []byte) (n int, err error) {
	callNum := w.calls.Add(1)

	// Simulate delay if configured
	if w.delayMs > 0 {
		time.Sleep(time.Duration(w.delayMs) * time.Millisecond)
	}

	// Check if we should fail
	if w.failAt > 0 && callNum == w.failAt {
		return 0, ErrTestWriterFailed
	}

	// Make a copy to avoid data races
	copied := make([]byte, len(p))
	copy(copied, p)

	w.mu.Lock()
	w.data = append(w.data, copied)
	w.mu.Unlock()

	return len(p), nil
}

// GetData returns a copy of all data written so far.
func (w *testWriter) GetData() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([][]byte, len(w.data))
	copy(result, w.data)
	return result
}

// GetCallCount returns the number of times Write was called.
func (w *testWriter) GetCallCount() int32 {
	return w.calls.Load()
}

// Reset clears all captured data and resets the call counter.
func (w *testWriter) Reset() {
	w.mu.Lock()
	w.data = make([][]byte, 0)
	w.mu.Unlock()
	w.calls.Store(0)
}

// SetFailAt configures the writer to fail at a specific call number.
func (w *testWriter) SetFailAt(callNum int32) {
	w.failAt = callNum
}

// SetDelay configures a delay in milliseconds for each write operation.
func (w *testWriter) SetDelay(ms int) {
	w.delayMs = ms
}

// testCounter is a thread-safe counter that tracks function calls with timestamps.
type testCounter struct {
	seq   *atomic.Uint64
	calls libatm.MapTyped[uint64, time.Time]
}

// newTestCounter creates a new testCounter instance.
func newTestCounter() *testCounter {
	return &testCounter{
		seq:   new(atomic.Uint64),
		calls: libatm.NewMapTyped[uint64, time.Time](),
	}
}

// Inc increments the counter and records the current timestamp.
func (c *testCounter) Inc() {
	c.seq.Add(1)
	c.calls.Store(c.seq.Load(), time.Now())
}

// Get returns the current counter value as an int.
func (c *testCounter) Get() int {
	if i := c.seq.Load(); i > uint64(math.MaxInt) {
		return math.MaxInt
	} else {
		return int(i)
	}
}

// GetCalls returns all timestamps of recorded calls in order.
func (c *testCounter) GetCalls() []time.Time {
	var l int
	if i := c.seq.Load(); i > uint64(math.MaxInt) {
		l = math.MaxInt
	} else {
		l = int(i)
	}

	var result = make([]time.Time, l)
	c.calls.Range(func(k uint64, v time.Time) bool {
		if k > uint64(l) {
			return false
		}

		result[k] = v
		return true
	})

	return result
}

// Reset clears the counter and all recorded timestamps.
func (c *testCounter) Reset() {
	c.seq.Store(0)
	c.calls.Range(func(k uint64, _ time.Time) bool {
		c.calls.Delete(k)
		return true
	})
}

// Test-specific errors
var (
	// ErrTestWriterFailed is returned by testWriter when configured to fail.
	ErrTestWriterFailed = errors.New("test writer failed")
)

// waitForCondition polls a condition function until it returns true or timeout occurs.
// Returns true if condition became true, false if timeout occurred.
func waitForCondition(timeout time.Duration, checkInterval time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		if condition() {
			return true
		}

		if time.Now().After(deadline) {
			return false
		}

		<-ticker.C
	}
}

// startAndWait starts the aggregator and waits for it to be fully running.
// It handles ErrStillRunning gracefully for concurrent start attempts.
func startAndWait(agg iotagg.Aggregator, ctx context.Context) error {
	err := agg.Start(ctx)
	// ErrStillRunning means it's already starting/running, which is ok for concurrent calls
	if err != nil && err != iotagg.ErrStillRunning {
		return err
	}

	// Wait for aggregator to be fully running
	Eventually(func() bool {
		return agg.IsRunning()
	}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

	return nil
}
