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
	"testing"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	"github.com/nabbar/golib/ioutils/aggregator"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// Global context for all tests
	testCtx    context.Context
	testCancel context.CancelFunc
	globalLog  liblog.Logger
)

// TestAggregator is the entry point for the Ginkgo test suite
func TestAggregator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOUtils/Aggregator Package Suite")
}

var _ = BeforeSuite(func() {
	testCtx, testCancel = context.WithCancel(context.Background())
	globalLog = liblog.New(context.Background())
	Expect(globalLog.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	if testCancel != nil {
		testCancel()
	}
})

// Helper functions

// testWriter is a thread-safe writer that captures all writes
type testWriter struct {
	mu      sync.Mutex
	data    [][]byte
	calls   atomic.Int32
	failAt  int32 // fail at this call number (0 = never fail)
	delayMs int   // delay each write by this many milliseconds
}

func newTestWriter() *testWriter {
	return &testWriter{
		data: make([][]byte, 0),
	}
}

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

func (w *testWriter) GetData() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([][]byte, len(w.data))
	copy(result, w.data)
	return result
}

func (w *testWriter) GetCallCount() int32 {
	return w.calls.Load()
}

func (w *testWriter) Reset() {
	w.mu.Lock()
	w.data = make([][]byte, 0)
	w.mu.Unlock()
	w.calls.Store(0)
}

func (w *testWriter) SetFailAt(callNum int32) {
	w.failAt = callNum
}

func (w *testWriter) SetDelay(ms int) {
	w.delayMs = ms
}

// testCounter tracks function calls
type testCounter struct {
	seq   *atomic.Uint64
	calls libatm.MapTyped[uint64, time.Time]
}

func newTestCounter() *testCounter {
	return &testCounter{
		seq:   new(atomic.Uint64),
		calls: libatm.NewMapTyped[uint64, time.Time](),
	}
}

func (c *testCounter) Inc() {
	c.seq.Add(1)
	c.calls.Store(c.seq.Load(), time.Now())
}

func (c *testCounter) Get() int {
	if i := c.seq.Load(); i > uint64(math.MaxInt) {
		return math.MaxInt
	} else {
		return int(i)
	}
}

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

func (c *testCounter) Reset() {
	c.seq.Store(0)
	c.calls.Range(func(k uint64, _ time.Time) bool {
		c.calls.Delete(k)
		return true
	})
}

// Errors for testing
var (
	ErrTestWriterFailed = errors.New("test writer failed")
)

// waitForCondition waits for a condition to be true or timeout
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

// startAndWait starts the aggregator and waits for it to be running
func startAndWait(agg aggregator.Aggregator, ctx context.Context) error {
	err := agg.Start(ctx)
	// ErrStillRunning means it's already starting/running, which is ok for concurrent calls
	if err != nil && err != aggregator.ErrStillRunning {
		return err
	}

	// Wait for aggregator to be fully running
	Eventually(func() bool {
		return agg.IsRunning()
	}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

	return nil
}
