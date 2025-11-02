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

package nopwritecloser_test

import (
	"bytes"
	"sync"
	"sync/atomic"

	. "github.com/nabbar/golib/ioutils/nopwritecloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// safeBuffer wraps bytes.Buffer with a mutex for thread-safe operations
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

var _ = Describe("NopWriteCloser - Concurrency", func() {
	Context("Concurrent writes", func() {
		It("should handle concurrent writes safely", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			numGoroutines := 50
			writesPerGoroutine := 100

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < writesPerGoroutine; j++ {
						_, err := wc.Write([]byte("x"))
						Expect(err).ToNot(HaveOccurred())
					}
				}(i)
			}

			wg.Wait()

			// Verify all writes succeeded
			expectedLen := numGoroutines * writesPerGoroutine
			Expect(buf.Len()).To(Equal(expectedLen))
		})

		It("should handle concurrent writes with different data", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			var writeCount atomic.Int64

			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					n, err := wc.Write([]byte("test"))
					Expect(err).ToNot(HaveOccurred())
					Expect(n).To(Equal(4))
					writeCount.Add(1)
				}()
			}

			wg.Wait()

			Expect(writeCount.Load()).To(Equal(int64(20)))
			Expect(buf.Len()).To(Equal(80)) // 20 * 4
		})
	})

	Context("Concurrent closes", func() {
		It("should handle concurrent closes safely", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			numGoroutines := 50

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					err := wc.Close()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})
	})

	Context("Mixed concurrent operations", func() {
		It("should handle concurrent writes and closes", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			var writeCount atomic.Int64
			var closeCount atomic.Int64

			// Writers
			for i := 0; i < 25; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 10; j++ {
						_, err := wc.Write([]byte("a"))
						Expect(err).ToNot(HaveOccurred())
						writeCount.Add(1)
					}
				}()
			}

			// Closers
			for i := 0; i < 25; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 10; j++ {
						err := wc.Close()
						Expect(err).ToNot(HaveOccurred())
						closeCount.Add(1)
					}
				}()
			}

			wg.Wait()

			Expect(writeCount.Load()).To(Equal(int64(250)))
			Expect(closeCount.Load()).To(Equal(int64(250)))
			Expect(buf.Len()).To(Equal(250))
		})

		It("should handle rapid creation and usage", func() {
			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					buf := &safeBuffer{}
					wc := New(buf)

					wc.Write([]byte("test"))
					wc.Close()

					Expect(buf.Len()).To(Equal(4))
				}()
			}

			wg.Wait()
		})
	})

	Context("High load scenarios", func() {
		It("should handle high-frequency writes", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			totalWrites := 10000

			for i := 0; i < totalWrites; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, err := wc.Write([]byte("x"))
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
			Expect(buf.Len()).To(Equal(totalWrites))
		})

		It("should handle large concurrent writes", func() {
			buf := &safeBuffer{}
			wc := New(buf)

			var wg sync.WaitGroup
			dataSize := 1024 // 1KB per write
			numGoroutines := 100

			data := make([]byte, dataSize)
			for i := range data {
				data[i] = byte(i % 256)
			}

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					n, err := wc.Write(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(n).To(Equal(dataSize))
				}()
			}

			wg.Wait()

			expectedSize := dataSize * numGoroutines
			Expect(buf.Len()).To(Equal(expectedSize))
		})
	})
})
