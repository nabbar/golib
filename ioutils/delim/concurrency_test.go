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

package delim_test

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file validates concurrency safety and race conditions.
// IMPORTANT: These tests are designed to be run with the race detector:
//   CGO_ENABLED=1 go test -race
//
// The tests cover:
//   - Sequential access patterns (baseline for single goroutine)
//   - Multiple readers on separate BufferDelim instances (safe pattern)
//   - Synchronized access patterns with explicit locking
//   - Pipeline patterns with goroutine coordination
//   - Resource cleanup and proper Close() handling
//
// Note: BufferDelim is NOT safe for concurrent access by multiple
// goroutines on the same instance. Each instance should be used by
// a single goroutine. These tests validate this design and demonstrate
// safe concurrent patterns using separate instances.

var _ = Describe("BufferDelim Concurrency and Race Detection", func() {
	Describe("Sequential access patterns", func() {
		Context("with single goroutine", func() {
			It("should handle sequential reads safely", func() {
				data := strings.Repeat("line\n", 100)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0, false)

				for i := 0; i < 100; i++ {
					_, err := bd.ReadBytes()
					if err == io.EOF {
						break
					}
					Expect(err).To(BeNil())
				}
			})

			It("should handle sequential Read calls safely", func() {
				data := strings.Repeat("test\n", 50)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 100)
				for i := 0; i < 50; i++ {
					_, err := bd.Read(buf)
					if err == io.EOF {
						break
					}
					Expect(err).To(BeNil())
				}
			})
		})
	})

	Describe("Concurrent read operations", func() {
		Context("with multiple readers on separate instances", func() {
			It("should handle concurrent reads on different BufferDelim instances", func() {
				var wg sync.WaitGroup
				numReaders := 10
				wg.Add(numReaders)

				for i := 0; i < numReaders; i++ {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("line\n", 100)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						for j := 0; j < 100; j++ {
							_, err := bd.ReadBytes()
							if err == io.EOF {
								break
							}
							Expect(err).To(BeNil())
						}
					}(i)
				}

				wg.Wait()
			})

			It("should handle concurrent Read calls on different instances", func() {
				var wg sync.WaitGroup
				numReaders := 20
				wg.Add(numReaders)

				for i := 0; i < numReaders; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("test\n", 50)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						buf := make([]byte, 100)
						for j := 0; j < 50; j++ {
							_, err := bd.Read(buf)
							if err == io.EOF {
								break
							}
							Expect(err).To(BeNil())
						}
					}()
				}

				wg.Wait()
			})
		})

		Context("with readers and method calls on same instance", func() {
			It("should handle Delim() calls during reads", func() {
				data := strings.Repeat("line\n", 100)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0, false)

				var wg sync.WaitGroup
				wg.Add(2)

				// Reader goroutine
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for i := 0; i < 50; i++ {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
				}()

				// Delim reader goroutine
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for i := 0; i < 100; i++ {
						delim := bd.Delim()
						Expect(delim).To(Equal('\n'))
						time.Sleep(time.Microsecond)
					}
				}()

				wg.Wait()
			})

			It("should handle Reader() calls during reads", func() {
				data := strings.Repeat("line\n", 100)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0, false)

				var wg sync.WaitGroup
				wg.Add(2)

				// Reader goroutine
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for i := 0; i < 50; i++ {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
				}()

				// Reader accessor
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for i := 0; i < 100; i++ {
						reader := bd.Reader()
						Expect(reader).NotTo(BeNil())
						time.Sleep(time.Microsecond)
					}
				}()

				wg.Wait()
			})
		})
	})

	Describe("Concurrent write operations", func() {
		Context("with multiple writers on separate instances", func() {
			It("should handle concurrent WriteTo on different instances", func() {
				var wg sync.WaitGroup
				numWriters := 10
				wg.Add(numWriters)

				for i := 0; i < numWriters; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("line\n", 100)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						buf := &bytes.Buffer{}
						_, err := bd.WriteTo(buf)
						Expect(err).To(Equal(io.EOF))
					}()
				}

				wg.Wait()
			})

			It("should handle concurrent Copy on different instances", func() {
				var wg sync.WaitGroup
				numWriters := 15
				wg.Add(numWriters)

				for i := 0; i < numWriters; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("test\n", 50)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						buf := &bytes.Buffer{}
						_, err := bd.Copy(buf)
						Expect(err).To(Equal(io.EOF))
					}()
				}

				wg.Wait()
			})
		})
	})

	Describe("Mixed concurrent operations", func() {
		Context("with different operations on separate instances", func() {
			It("should handle mixed Read and Write operations", func() {
				var wg sync.WaitGroup
				numWorkers := 20
				wg.Add(numWorkers)

				for i := 0; i < numWorkers; i++ {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("line\n", 50)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						if id%2 == 0 {
							// Read operations
							buf := make([]byte, 100)
							for j := 0; j < 25; j++ {
								_, err := bd.Read(buf)
								if err == io.EOF {
									break
								}
							}
						} else {
							// Write operations
							buf := &bytes.Buffer{}
							_, _ = bd.WriteTo(buf)
						}
					}(i)
				}

				wg.Wait()
			})

			It("should handle all operations concurrently on separate instances", func() {
				var wg sync.WaitGroup
				numWorkers := 30
				wg.Add(numWorkers)

				for i := 0; i < numWorkers; i++ {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("test\n", 50)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', libsiz.Size(512), false)

						switch id % 4 {
						case 0:
							// Read
							buf := make([]byte, 100)
							for j := 0; j < 25; j++ {
								_, _ = bd.Read(buf)
							}
						case 1:
							// ReadBytes
							for j := 0; j < 25; j++ {
								_, err := bd.ReadBytes()
								if err == io.EOF {
									break
								}
							}
						case 2:
							// WriteTo
							buf := &bytes.Buffer{}
							_, _ = bd.WriteTo(buf)
						case 3:
							// Copy
							buf := &bytes.Buffer{}
							_, _ = bd.Copy(buf)
						}
					}(i)
				}

				wg.Wait()
			})
		})
	})

	Describe("Construction and close concurrency", func() {
		Context("with concurrent construction", func() {
			It("should handle multiple concurrent constructions", func() {
				var wg sync.WaitGroup
				numConstructors := 50
				wg.Add(numConstructors)

				for i := 0; i < numConstructors; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := "test\n"
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)
						Expect(bd).NotTo(BeNil())
						Expect(bd.Delim()).To(Equal('\n'))
					}()
				}

				wg.Wait()
			})

			It("should handle construction with various buffer sizes concurrently", func() {
				var wg sync.WaitGroup
				bufferSizes := []libsiz.Size{0, 64, 512, 1024, 4096, 8192}
				wg.Add(len(bufferSizes) * 10)

				for _, size := range bufferSizes {
					for i := 0; i < 10; i++ {
						go func(s libsiz.Size) {
							defer wg.Done()
							defer GinkgoRecover()

							data := "test\n"
							r := io.NopCloser(strings.NewReader(data))
							bd := iotdlm.New(r, '\n', s, false)
							Expect(bd).NotTo(BeNil())
						}(size)
					}
				}

				wg.Wait()
			})
		})

		Context("with concurrent close operations", func() {
			It("should handle close on separate instances", func() {
				var wg sync.WaitGroup
				numInstances := 20
				wg.Add(numInstances)

				for i := 0; i < numInstances; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := "test\n"
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						// Do some work
						_, _ = bd.ReadBytes()

						// Close
						err := bd.Close()
						Expect(err).NotTo(HaveOccurred())
					}()
				}

				wg.Wait()
			})
		})
	})

	Describe("Stress tests", func() {
		Context("with high load", func() {
			It("should handle many concurrent operations", func() {
				var wg sync.WaitGroup
				numWorkers := 100
				wg.Add(numWorkers)

				for i := 0; i < numWorkers; i++ {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("line\n", 1000)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						if id%3 == 0 {
							buf := make([]byte, 100)
							for {
								_, err := bd.Read(buf)
								if err == io.EOF {
									break
								}
							}
						} else if id%3 == 1 {
							for {
								_, err := bd.ReadBytes()
								if err == io.EOF {
									break
								}
							}
						} else {
							buf := &bytes.Buffer{}
							_, _ = bd.WriteTo(buf)
						}

						_ = bd.Close()
					}(i)
				}

				wg.Wait()
			})

			It("should handle rapid create-use-close cycles", func() {
				var wg sync.WaitGroup
				numCycles := 200
				wg.Add(numCycles)

				for i := 0; i < numCycles; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := "quick\ntest\n"
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0, false)

						_, _ = bd.ReadBytes()
						_ = bd.Close()
					}()
				}

				wg.Wait()
			})
		})

		Context("with varying data sizes", func() {
			It("should handle concurrent operations with different data sizes", func() {
				var wg sync.WaitGroup
				sizes := []int{10, 100, 1000, 10000}
				wg.Add(len(sizes) * 10)

				for _, size := range sizes {
					for i := 0; i < 10; i++ {
						go func(s int) {
							defer wg.Done()
							defer GinkgoRecover()

							data := strings.Repeat("x", s) + "\n"
							r := io.NopCloser(strings.NewReader(data))
							bd := iotdlm.New(r, '\n', 0, false)

							buf := &bytes.Buffer{}
							_, err := bd.WriteTo(buf)
							Expect(err).To(Equal(io.EOF))
							Expect(buf.Len()).To(Equal(s + 1))
						}(size)
					}
				}

				wg.Wait()
			})
		})
	})

	Describe("UnRead concurrency", func() {
		Context("with concurrent UnRead calls", func() {
			It("should handle UnRead on separate instances", func() {
				var wg sync.WaitGroup
				numWorkers := 20
				wg.Add(numWorkers)

				for i := 0; i < numWorkers; i++ {
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						data := strings.Repeat("test\n", 100)
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 100, false)

						// Read some data
						_, _ = bd.ReadBytes()

						// UnRead
						unread, err := bd.UnRead()
						Expect(err).To(BeNil())
						_ = unread
					}()
				}

				wg.Wait()
			})
		})
	})

	Describe("DiscardCloser concurrency", func() {
		Context("with concurrent DiscardCloser operations", func() {
			It("should handle concurrent operations safely", func() {
				var wg sync.WaitGroup
				dc := iotdlm.DiscardCloser{}
				numWorkers := 50
				wg.Add(numWorkers)

				for i := 0; i < numWorkers; i++ {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						switch id % 3 {
						case 0:
							buf := make([]byte, 100)
							for j := 0; j < 100; j++ {
								_, _ = dc.Read(buf)
							}
						case 1:
							data := []byte("test data")
							for j := 0; j < 100; j++ {
								_, _ = dc.Write(data)
							}
						case 2:
							for j := 0; j < 10; j++ {
								_ = dc.Close()
							}
						}
					}(i)
				}

				wg.Wait()
			})
		})
	})
})
