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

package multi_test

import (
	"bytes"
	"io"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

// Tests for Multi concurrent operations and thread safety.
// These tests verify that all operations (Read, Write, AddWriter, SetInput,
// Clean) are safe for concurrent use and do not cause data races or panics.
// Uses the --race flag to detect race conditions.
var _ = Describe("Multi Concurrent Operations", func() {
	var m multi.Multi

	BeforeEach(func() {
		m = multi.New()
	})

	Describe("Concurrent writes", func() {
		Context("multiple goroutines writing", func() {
			It("should handle concurrent writes safely", func() {
				var buf safeBuffer
				m.AddWriter(&buf)

				var wg sync.WaitGroup
				iterations := 100

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						m.Write([]byte("x"))
					}(i)
				}

				wg.Wait()
				Expect(buf.Len()).To(Equal(iterations))
			})

			It("should handle concurrent WriteString calls", func() {
				var buf safeBuffer
				m.AddWriter(&buf)

				var wg sync.WaitGroup
				iterations := 100

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.WriteString("y")
					}()
				}

				wg.Wait()
				Expect(buf.Len()).To(Equal(iterations))
			})
		})

		Context("concurrent writes to multiple writers", func() {
			It("should broadcast to all writers concurrently", func() {
				var buf1, buf2, buf3 safeBuffer
				m.AddWriter(&buf1, &buf2, &buf3)

				var wg sync.WaitGroup
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.Write([]byte("a"))
					}()
				}

				wg.Wait()
				Expect(buf1.Len()).To(Equal(iterations))
				Expect(buf2.Len()).To(Equal(iterations))
				Expect(buf3.Len()).To(Equal(iterations))
			})
		})
	})

	Describe("Concurrent reads", func() {
		Context("sequential reads with proper synchronization", func() {
			It("should handle reads when properly synchronized", func() {
				// Note: concurrent reads from the same Reader is not safe
				// as the underlying Reader (like strings.Reader) is not thread-safe.
				// This test verifies that reads work correctly with synchronization.
				input := io.NopCloser(strings.NewReader(strings.Repeat("x", 1000)))
				m.SetInput(input)

				var mu sync.Mutex
				var wg sync.WaitGroup
				iterations := 10

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						mu.Lock()
						defer mu.Unlock()
						buf := make([]byte, 10)
						m.Read(buf)
					}()
				}

				wg.Wait()
				// Should complete without panic
			})
		})
	})

	Describe("Concurrent AddWriter calls", func() {
		Context("adding writers concurrently", func() {
			It("should handle concurrent AddWriter calls", func() {
				var wg sync.WaitGroup
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						var buf bytes.Buffer
						m.AddWriter(&buf)
					}()
				}

				wg.Wait()

				// Should not panic and writer should work
				var testBuf bytes.Buffer
				m.AddWriter(&testBuf)
				m.Write([]byte("test"))
				Expect(testBuf.String()).To(Equal("test"))
			})

			It("should handle concurrent AddWriter with nil values", func() {
				var wg sync.WaitGroup
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						if index%2 == 0 {
							var buf bytes.Buffer
							m.AddWriter(&buf)
						} else {
							m.AddWriter(nil)
						}
					}(i)
				}

				wg.Wait()
				// Should not panic
			})
		})
	})

	Describe("Concurrent Clean calls", func() {
		Context("cleaning while writing", func() {
			It("should handle concurrent Clean and Write", func() {
				var buf safeBuffer
				m.AddWriter(&buf)

				var wg sync.WaitGroup
				iterations := 50

				// Concurrent writes
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.Write([]byte("x"))
					}()
				}

				// Concurrent cleans
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.Clean()
					}()
				}

				wg.Wait()
				// Should complete without panic
			})
		})

		Context("cleaning while adding writers", func() {
			It("should handle concurrent Clean and AddWriter", func() {
				var wg sync.WaitGroup
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(2)
					go func() {
						defer wg.Done()
						var buf bytes.Buffer
						m.AddWriter(&buf)
					}()
					go func() {
						defer wg.Done()
						m.Clean()
					}()
				}

				wg.Wait()
				// Should complete without panic
			})
		})
	})

	Describe("Concurrent SetInput calls", func() {
		Context("setting input concurrently", func() {
			It("should handle concurrent SetInput calls", func() {
				var wg sync.WaitGroup
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						input := io.NopCloser(strings.NewReader("data"))
						m.SetInput(input)
					}()
				}

				wg.Wait()

				// Reader should be set
				reader := m.Reader()
				Expect(reader).NotTo(BeNil())
			})
		})
	})

	Describe("Mixed concurrent operations", func() {
		Context("all operations running concurrently", func() {
			It("should handle mixed concurrent operations", func() {
				var wg sync.WaitGroup
				iterations := 20

				// Concurrent AddWriter
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						var buf safeBuffer
						m.AddWriter(&buf)
					}()
				}

				// Concurrent Write
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.Write([]byte("x"))
					}()
				}

				// Concurrent SetInput
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						input := io.NopCloser(strings.NewReader("data"))
						m.SetInput(input)
					}()
				}

				// Note: Concurrent Read removed to avoid data race on underlying Reader
				// as strings.Reader is not thread-safe for concurrent reads

				// Concurrent Clean
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.Clean()
					}()
				}

				wg.Wait()
				// Should complete without panic
			})
		})

		Context("sequential copy operations", func() {
			It("should handle sequential Copy calls", func() {
				var buf safeBuffer
				m.AddWriter(&buf)

				// Sequential copy operations to avoid data race on underlying Reader
				// Note: io.Copy uses Reader.WriteTo if available, and strings.Reader
				// is not thread-safe for concurrent operations
				for i := 0; i < 10; i++ {
					input := io.NopCloser(strings.NewReader("data"))
					m.SetInput(input)
					n, err := m.Copy()
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(int64(4)))
				}

				Expect(buf.Len()).To(Equal(40)) // 10 * 4
			})
		})
	})
})
