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

package iowrapper_test

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IOWrapper - Concurrency", func() {
	Context("Concurrent reads", func() {
		It("should handle concurrent read operations safely", func() {
			wrapper := New(nil)

			var counter atomic.Int64
			wrapper.SetRead(func(p []byte) []byte {
				counter.Add(1)
				return []byte("data")
			})

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					data := make([]byte, 10)
					wrapper.Read(data)
				}()
			}

			wg.Wait()
			Expect(counter.Load()).To(Equal(int64(concurrency)))
		})

		It("should handle concurrent reads with custom read function", func() {
			wrapper := New(nil)

			var mu sync.Mutex
			var counter int
			wrapper.SetRead(func(p []byte) []byte {
				mu.Lock()
				defer mu.Unlock()
				counter++
				return []byte("data")
			})

			var wg sync.WaitGroup
			concurrency := 50
			results := make([]int, concurrency)

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				idx := i
				go func() {
					defer wg.Done()
					data := make([]byte, 5)
					n, _ := wrapper.Read(data)
					results[idx] = n
				}()
			}

			wg.Wait()

			// Verify no panics occurred and all reads completed
			Expect(counter).To(Equal(concurrency))
			for _, n := range results {
				Expect(n).To(Equal(4)) // "data" has 4 bytes
			}
		})
	})

	Context("Concurrent writes", func() {
		It("should handle concurrent write operations safely", func() {
			wrapper := New(nil)

			var counter atomic.Int64
			wrapper.SetWrite(func(p []byte) []byte {
				counter.Add(1)
				return p
			})

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.Write([]byte("data"))
				}()
			}

			wg.Wait()
			Expect(counter.Load()).To(Equal(int64(concurrency)))
		})

		It("should handle concurrent writes to underlying writer", func() {
			buf := &bytes.Buffer{}
			var mu sync.Mutex
			wrapper := New(buf)

			// Custom write to ensure thread-safe writes
			wrapper.SetWrite(func(p []byte) []byte {
				mu.Lock()
				defer mu.Unlock()
				buf.Write(p)
				return p
			})

			var wg sync.WaitGroup
			concurrency := 50

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.Write([]byte("x"))
				}()
			}

			wg.Wait()
			Expect(buf.Len()).To(Equal(concurrency))
		})
	})

	Context("Concurrent function updates", func() {
		It("should handle concurrent SetRead calls", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				value := i
				go func() {
					defer wg.Done()
					wrapper.SetRead(func(p []byte) []byte {
						return []byte{byte(value)}
					})
				}()
			}

			wg.Wait()

			// Should complete without panics
			data := make([]byte, 1)
			_, err := wrapper.Read(data)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent SetWrite calls", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				value := i
				go func() {
					defer wg.Done()
					wrapper.SetWrite(func(p []byte) []byte {
						return []byte{byte(value)}
					})
				}()
			}

			wg.Wait()

			// Should complete without panics
			_, err := wrapper.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent SetSeek calls", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				value := i
				go func() {
					defer wg.Done()
					wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
						return int64(value), nil
					})
				}()
			}

			wg.Wait()

			// Should complete without panics
			_, err := wrapper.Seek(0, io.SeekStart)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent SetClose calls", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup
			concurrency := 100

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.SetClose(func() error {
						return nil
					})
				}()
			}

			wg.Wait()

			// Should complete without panics
			err := wrapper.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Mixed concurrent operations", func() {
		It("should handle concurrent reads, writes, and seeks", func() {
			wrapper := New(nil)

			var readCounter atomic.Int64
			var writeCounter atomic.Int64
			var seekCounter atomic.Int64

			wrapper.SetRead(func(p []byte) []byte {
				readCounter.Add(1)
				return []byte("data")
			})

			wrapper.SetWrite(func(p []byte) []byte {
				writeCounter.Add(1)
				return p
			})

			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				seekCounter.Add(1)
				return 0, nil
			})

			var wg sync.WaitGroup
			operations := 300

			// Concurrent reads
			for i := 0; i < operations/3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					data := make([]byte, 10)
					wrapper.Read(data)
				}()
			}

			// Concurrent writes
			for i := 0; i < operations/3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.Write([]byte("test"))
				}()
			}

			// Concurrent seeks
			for i := 0; i < operations/3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.Seek(0, io.SeekStart)
				}()
			}

			wg.Wait()
			// Verify all operations completed
			Expect(readCounter.Load()).To(Equal(int64(operations / 3)))
			Expect(writeCounter.Load()).To(Equal(int64(operations / 3)))
			Expect(seekCounter.Load()).To(Equal(int64(operations / 3)))
		})

		It("should handle concurrent operations with function updates", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup
			operations := 200

			// Concurrent reads
			for i := 0; i < operations/4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.SetRead(func(p []byte) []byte {
						return []byte("r")
					})
					data := make([]byte, 1)
					wrapper.Read(data)
				}()
			}

			// Concurrent writes
			for i := 0; i < operations/4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.SetWrite(func(p []byte) []byte {
						return p
					})
					wrapper.Write([]byte("w"))
				}()
			}

			// Concurrent seeks
			for i := 0; i < operations/4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
						return 0, nil
					})
					wrapper.Seek(0, io.SeekStart)
				}()
			}

			// Concurrent closes
			for i := 0; i < operations/4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wrapper.SetClose(func() error {
						return nil
					})
					wrapper.Close()
				}()
			}

			wg.Wait()
			// Should complete without panics
		})
	})

	Context("Race condition detection", func() {
		It("should not have races when reading and updating function", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup

			// Reader goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					data := make([]byte, 1)
					wrapper.Read(data)
				}
			}()

			// Updater goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					wrapper.SetRead(func(p []byte) []byte {
						return []byte("x")
					})
				}
			}()

			wg.Wait()
			// Should complete without data races (run with -race flag)
		})

		It("should not have races when writing and updating function", func() {
			wrapper := New(nil)

			var wg sync.WaitGroup

			// Writer goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					wrapper.Write([]byte("x"))
				}
			}()

			// Updater goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					wrapper.SetWrite(func(p []byte) []byte {
						return p
					})
				}
			}()

			wg.Wait()
			// Should complete without data races (run with -race flag)
		})
	})
})
