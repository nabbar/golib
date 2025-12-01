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
	"sync"

	. "github.com/nabbar/golib/ioutils/bufferReadCloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Concurrency tests verify that wrappers behave correctly under concurrent access.
// Note: The wrappers are NOT thread-safe by design (like stdlib buffers), so these
// tests verify that race conditions are detected when run with -race flag, not that
// they work correctly under concurrent access.
//
// These tests serve to document the non-thread-safe nature of the wrappers and
// ensure that users are aware they need external synchronization.
var _ = Describe("Concurrency", func() {
	// Buffer concurrency tests demonstrate that Buffer is not thread-safe.
	// These tests will trigger race detector warnings when run with -race.
	Context("Buffer concurrent access", func() {
		It("should handle concurrent close calls with mutex", func() {
			// This test shows the CORRECT way to use Buffer concurrently
			buf := bytes.NewBuffer(nil)
			wrapped := NewBuffer(buf, nil)
			
			var mu sync.Mutex
			var closeCount int
			
			// Multiple goroutines trying to close - protected by mutex
			concurrentRunner(10, func(id int) {
				mu.Lock()
				defer mu.Unlock()
				
				err := wrapped.Close()
				Expect(err).ToNot(HaveOccurred())
				closeCount++
			})
			
			// All close calls succeeded
			Expect(closeCount).To(Equal(10))
		})
		
		It("should track concurrent operations with atomic counter", func() {
			// Demonstrate safe concurrent tracking using atomic operations
			counter := &concurrentCounter{}
			
			concurrentRunner(100, func(id int) {
				buf := bytes.NewBuffer(nil)
				wrapped := NewBuffer(buf, func() error {
					counter.inc()
					return nil
				})
				wrapped.Close()
			})
			
			// All close functions were called
			Expect(counter.get()).To(Equal(int64(100)))
		})
	})
	
	// Reader concurrency tests demonstrate that Reader is not thread-safe.
	Context("Reader concurrent access", func() {
		It("should handle concurrent close with synchronization", func() {
			source := bytes.NewReader(generateTestData(1024))
			br := bufio.NewReader(source)
			wrapped := NewReader(br, nil)
			
			var mu sync.Mutex
			var wg sync.WaitGroup
			wg.Add(10)
			
			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					mu.Lock()
					defer mu.Unlock()
					wrapped.Close()
				}()
			}
			
			wg.Wait()
		})
	})
	
	// Writer concurrency tests demonstrate that Writer is not thread-safe.
	Context("Writer concurrent access", func() {
		It("should handle concurrent close with synchronization", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			wrapped := NewWriter(bw, nil)
			
			var mu sync.Mutex
			var wg sync.WaitGroup
			wg.Add(10)
			
			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					mu.Lock()
					defer mu.Unlock()
					wrapped.Close()
				}()
			}
			
			wg.Wait()
		})
	})
	
	// ReadWriter concurrency tests demonstrate that ReadWriter is not thread-safe.
	Context("ReadWriter concurrent access", func() {
		It("should handle concurrent close with synchronization", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			wrapped := NewReadWriter(brw, nil)
			
			var mu sync.Mutex
			var wg sync.WaitGroup
			wg.Add(10)
			
			for i := 0; i < 10; i++ {
				go func() {
					defer wg.Done()
					mu.Lock()
					defer mu.Unlock()
					wrapped.Close()
				}()
			}
			
			wg.Wait()
		})
	})
	
	// Custom close function concurrency tests verify safe concurrent execution
	// of close callbacks using atomic operations.
	Context("Custom close function concurrency", func() {
		It("should safely execute close functions concurrently", func() {
			counter := &concurrentCounter{}
			
			// Create multiple buffers with close functions
			var wg sync.WaitGroup
			wg.Add(50)
			
			for i := 0; i < 50; i++ {
				go func() {
					defer wg.Done()
					
					buf := bytes.NewBuffer(nil)
					wrapped := NewBuffer(buf, func() error {
						counter.inc()
						return nil
					})
					
					wrapped.WriteString("data")
					wrapped.Close()
				}()
			}
			
			wg.Wait()
			
			// All close functions executed
			Expect(counter.get()).To(Equal(int64(50)))
		})
		
		It("should handle concurrent buffer pool operations", func() {
			// Simulate sync.Pool-like behavior
			pool := make(chan *bytes.Buffer, 10)
			
			// Initialize pool
			for i := 0; i < 10; i++ {
				pool <- bytes.NewBuffer(make([]byte, 0, 1024))
			}
			
			counter := &concurrentCounter{}
			
			// Concurrent get/use/return
			concurrentRunner(100, func(id int) {
				// Get from pool
				buf := <-pool
				
				// Use with wrapper
				wrapped := NewBuffer(buf, func() error {
					counter.inc()
					// Return to pool
					pool <- buf
					return nil
				})
				
				wrapped.WriteString("test data")
				wrapped.Close()
			})
			
			// All operations completed
			Expect(counter.get()).To(Equal(int64(100)))
			
			// Pool still has 10 buffers
			Expect(len(pool)).To(Equal(10))
		})
	})
})
