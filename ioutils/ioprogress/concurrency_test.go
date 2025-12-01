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

// Package ioprogress_test provides concurrency tests for the ioprogress package.
//
// These tests validate thread-safe operations including:
//   - Concurrent callback registration
//   - Concurrent I/O operations
//   - Atomic counter updates
//   - Lock-free concurrent access
//
// Running with Race Detector:
//   CGO_ENABLED=1 go test -race ./...
//
// All tests should pass with zero data races detected.
package ioprogress_test

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	. "github.com/nabbar/golib/ioutils/ioprogress"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrency", func() {
	Context("Reader concurrent operations", func() {
		It("should handle concurrent callback registration during reads", func() {
			data := strings.Repeat("x", 10000)
			reader := NewReadCloser(io.NopCloser(strings.NewReader(data)))
			defer reader.Close()

			var totalBytes int64
			var wg sync.WaitGroup

			// Start reading in one goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf := make([]byte, 1000)
				for {
					_, err := reader.Read(buf)
					if err == io.EOF {
						break
					}
				}
			}()

			// Register callbacks concurrently from multiple goroutines
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					reader.RegisterFctIncrement(func(size int64) {
						atomic.AddInt64(&totalBytes, size)
					})
				}()
			}

			wg.Wait()

			// Verify that some bytes were counted (at least one callback was active)
			Expect(atomic.LoadInt64(&totalBytes)).To(BeNumerically(">", 0))
		})

		It("should maintain correct count with concurrent reads", func() {
			// Note: Most readers don't support concurrent Read() operations
			// This test validates that the counter remains consistent even if
			// callbacks are invoked from multiple goroutines
			
			data := strings.Repeat("x", 5000)
			var totalFromCallbacks int64
			var readCount int32

			// Create multiple readers reading different data
			readers := make([]Reader, 5)
			for i := range readers {
				readers[i] = NewReadCloser(io.NopCloser(strings.NewReader(data)))
				readers[i].RegisterFctIncrement(func(size int64) {
					atomic.AddInt64(&totalFromCallbacks, size)
				})
			}

			var wg sync.WaitGroup

			// Read from all readers concurrently
			for i := range readers {
				wg.Add(1)
				go func(r Reader) {
					defer wg.Done()
					defer r.Close()
					
					buf := make([]byte, 500)
					for {
						n, err := r.Read(buf)
						if n > 0 {
							atomic.AddInt32(&readCount, 1)
						}
						if err == io.EOF {
							break
						}
					}
				}(readers[i])
			}

			wg.Wait()

			// Each reader should have counted exactly len(data) bytes
			expectedTotal := int64(len(data) * len(readers))
			Expect(atomic.LoadInt64(&totalFromCallbacks)).To(Equal(expectedTotal))
			Expect(atomic.LoadInt32(&readCount)).To(BeNumerically(">", 0))
		})

		It("should allow concurrent Reset() calls", func() {
			data := strings.Repeat("x", 1000)
			reader := NewReadCloser(io.NopCloser(strings.NewReader(data)))
			defer reader.Close()

			var resetCount atomic.Int32
			reader.RegisterFctReset(func(max, current int64) {
				resetCount.Add(1)
			})

			var wg sync.WaitGroup

			// Call Reset() concurrently from multiple goroutines
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(size int64) {
					defer wg.Done()
					reader.Reset(size)
				}(int64(i * 100))
			}

			wg.Wait()

			// All reset callbacks should have been invoked
			Expect(resetCount.Load()).To(Equal(int32(10)))
		})
	})

	Context("Writer concurrent operations", func() {
		It("should handle concurrent callback registration during writes", func() {
			writer := NewWriteCloser(newCloseableWriter())
			defer writer.Close()

			var totalBytes int64
			var wg sync.WaitGroup

			// Write data in one goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				data := []byte(strings.Repeat("x", 10000))
				for i := 0; i < len(data); i += 1000 {
					end := i + 1000
					if end > len(data) {
						end = len(data)
					}
					writer.Write(data[i:end])
				}
			}()

			// Register callbacks concurrently
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					writer.RegisterFctIncrement(func(size int64) {
						atomic.AddInt64(&totalBytes, size)
					})
				}()
			}

			wg.Wait()

			// Verify that some bytes were counted
			Expect(atomic.LoadInt64(&totalBytes)).To(BeNumerically(">", 0))
		})

		It("should maintain correct count with multiple concurrent writers", func() {
			data := []byte(strings.Repeat("x", 5000))
			var totalFromCallbacks int64
			var writeCount int32

			// Create multiple writers
			writers := make([]Writer, 5)
			for i := range writers {
				writers[i] = NewWriteCloser(newCloseableWriter())
				writers[i].RegisterFctIncrement(func(size int64) {
					atomic.AddInt64(&totalFromCallbacks, size)
				})
			}

			var wg sync.WaitGroup

			// Write to all writers concurrently
			for i := range writers {
				wg.Add(1)
				go func(w Writer) {
					defer wg.Done()
					defer w.Close()
					
					for j := 0; j < len(data); j += 500 {
						end := j + 500
						if end > len(data) {
							end = len(data)
						}
						n, _ := w.Write(data[j:end])
						if n > 0 {
							atomic.AddInt32(&writeCount, 1)
						}
					}
				}(writers[i])
			}

			wg.Wait()

			// Each writer should have counted exactly len(data) bytes
			expectedTotal := int64(len(data) * len(writers))
			Expect(atomic.LoadInt64(&totalFromCallbacks)).To(Equal(expectedTotal))
			Expect(atomic.LoadInt32(&writeCount)).To(BeNumerically(">", 0))
		})
	})

	Context("Callback replacement under load", func() {
		It("should safely replace callbacks during heavy I/O", func() {
			data := strings.Repeat("x", 50000)
			reader := NewReadCloser(io.NopCloser(strings.NewReader(data)))
			defer reader.Close()

			var counter1, counter2, counter3 atomic.Int64
			var wg sync.WaitGroup

			// Heavy I/O in background
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf := make([]byte, 100)
				for {
					_, err := reader.Read(buf)
					if err == io.EOF {
						break
					}
				}
			}()

			// Replace callbacks multiple times concurrently
			wg.Add(3)
			
			go func() {
				defer wg.Done()
				reader.RegisterFctIncrement(func(size int64) {
					counter1.Add(size)
				})
			}()
			
			go func() {
				defer wg.Done()
				reader.RegisterFctIncrement(func(size int64) {
					counter2.Add(size)
				})
			}()
			
			go func() {
				defer wg.Done()
				reader.RegisterFctIncrement(func(size int64) {
					counter3.Add(size)
				})
			}()

			wg.Wait()

			// At least one counter should have been incremented
			total := counter1.Load() + counter2.Load() + counter3.Load()
			Expect(total).To(BeNumerically(">", 0))
		})
	})

	Context("Memory consistency under concurrency", func() {
		It("should maintain memory consistency with concurrent operations", func() {
			// This test validates that atomic operations provide proper
			// memory ordering guarantees
			
			const iterations = 1000
			const goroutines = 10
			
			var wg sync.WaitGroup
			var globalCounter int64

			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					for i := 0; i < iterations; i++ {
						data := strings.Repeat("x", 100)
						reader := NewReadCloser(io.NopCloser(strings.NewReader(data)))
						
						reader.RegisterFctIncrement(func(size int64) {
							atomic.AddInt64(&globalCounter, size)
						})
						
						buf := make([]byte, 10)
						for {
							_, err := reader.Read(buf)
							if err == io.EOF {
								break
							}
						}
						
						reader.Close()
					}
				}()
			}

			wg.Wait()

			// Total should be exactly: goroutines * iterations * data_size
			expected := int64(goroutines * iterations * 100)
			Expect(atomic.LoadInt64(&globalCounter)).To(Equal(expected))
		})
	})

	Context("Stress test", func() {
		It("should handle sustained concurrent load without races", func() {
			// Stress test with multiple readers/writers operating concurrently
			const numReaders = 5
			const numWriters = 5
			const dataSize = 10000
			
			var wg sync.WaitGroup
			var totalRead int64
			var totalWritten int64

			// Launch concurrent readers
			for i := 0; i < numReaders; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					data := strings.Repeat("x", dataSize)
					reader := NewReadCloser(io.NopCloser(strings.NewReader(data)))
					defer reader.Close()
					
					reader.RegisterFctIncrement(func(size int64) {
						atomic.AddInt64(&totalRead, size)
					})
					
					io.Copy(io.Discard, reader)
				}()
			}

			// Launch concurrent writers
			for i := 0; i < numWriters; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					writer := NewWriteCloser(newCloseableWriter())
					defer writer.Close()
					
					writer.RegisterFctIncrement(func(size int64) {
						atomic.AddInt64(&totalWritten, size)
					})
					
					data := bytes.Repeat([]byte("x"), dataSize)
					writer.Write(data)
				}()
			}

			wg.Wait()

			// Verify totals
			Expect(atomic.LoadInt64(&totalRead)).To(Equal(int64(numReaders * dataSize)))
			Expect(atomic.LoadInt64(&totalWritten)).To(Equal(int64(numWriters * dataSize)))
		})
	})
})
