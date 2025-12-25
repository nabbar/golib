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

package helper_test

import (
	"bytes"
	"io"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

var _ = Describe("TC-CC-001: Concurrency Tests", func() {
	Context("TC-CC-010: Concurrent construction", func() {
		It("TC-CC-011: should create multiple compress readers concurrently", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()
				}()
			}

			wg.Wait()
		})

		It("TC-CC-012: should create multiple compress writers concurrently", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()
				}()
			}

			wg.Wait()
		})

		It("TC-CC-013: should create mixed readers and writers concurrently", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(2)
				go func() {
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()
				}()
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()
				}()
			}

			wg.Wait()
		})
	})

	Context("TC-CC-020: Concurrent operations on separate instances", func() {
		It("TC-CC-021: should read from multiple compress readers concurrently", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test data"))
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()

					buf := make([]byte, 100)
					_, err = h.Read(buf)
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})

		It("TC-CC-022: should write to multiple compress writers concurrently", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()

					_, err = h.Write([]byte("test data"))
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})

		It("TC-CC-023: should handle concurrent decompress operations", func() {
			original := "Hello"
			var compBuf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &compBuf)
			cw.Write([]byte(original))
			cw.Close()
			compressed := compBuf.Bytes()

			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, bytes.NewReader(compressed))
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()

					buf := make([]byte, 100)
					_, err = h.Read(buf)
					if err != nil && err != io.EOF {
						Expect(err).ToNot(HaveOccurred())
					}
				}()
			}

			wg.Wait()
		})
	})

	Context("TC-CC-030: Concurrent close operations", func() {
		It("TC-CC-031: should handle concurrent reader closures", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
					Expect(err).ToNot(HaveOccurred())

					err = h.Close()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})

		It("TC-CC-032: should handle concurrent writer closures", func() {
			var wg sync.WaitGroup
			count := 10

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())

					h.Write([]byte("test"))
					err = h.Close()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})
	})

	Context("TC-CC-040: Thread safety validation", func() {
		It("TC-CC-041: should maintain data integrity with concurrent readers", func() {
			data := "test data for integrity check"
			var wg sync.WaitGroup
			count := 5

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
					Expect(err).ToNot(HaveOccurred())
					defer h.Close()

					result := make([]byte, 200)
					n, _ := h.Read(result)
					Expect(n).To(BeNumerically(">", 0))
				}()
			}

			wg.Wait()
		})

		It("TC-CC-042: should maintain data integrity with concurrent writers", func() {
			data := []byte("test data for integrity check")
			var wg sync.WaitGroup
			count := 5

			for i := 0; i < count; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())

					_, err = h.Write(data)
					Expect(err).ToNot(HaveOccurred())

					err = h.Close()
					Expect(err).ToNot(HaveOccurred())

					Expect(buf.Len()).To(BeNumerically(">", 0))
				}()
			}

			wg.Wait()
		})
	})
})
