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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

// Tests for Multi write operations and output management.
// These tests verify proper handling of multiple writers, write broadcasting,
// and dynamic writer management through AddWriter() and Clean().
var _ = Describe("[TC-WR] Multi Writer Operations", func() {
	var m multi.Multi

	BeforeEach(func() {
		m = multi.New(false, false, multi.DefaultConfig())
	})

	Describe("AddWriter", func() {
		Context("adding single writer", func() {
			It("[TC-WR-001] should add one writer successfully", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.Write([]byte("test"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
				Expect(buf.String()).To(Equal("test"))
			})

			It("[TC-WR-001] should replace io.Discard after adding writer", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				writer := m.Writer()
				Expect(writer).NotTo(BeNil())
			})
		})

		Context("adding multiple writers", func() {
			It("[TC-WR-001] should add multiple writers at once", func() {
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)

				n, err := m.Write([]byte("broadcast"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(9))
				Expect(buf1.String()).To(Equal("broadcast"))
				Expect(buf2.String()).To(Equal("broadcast"))
				Expect(buf3.String()).To(Equal("broadcast"))
			})

			It("[TC-WR-001] should add writers incrementally", func() {
				var buf1, buf2 bytes.Buffer
				m.AddWriter(&buf1)
				m.AddWriter(&buf2)

				n, err := m.Write([]byte("data"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
				Expect(buf1.String()).To(Equal("data"))
				Expect(buf2.String()).To(Equal("data"))
			})
		})

		Context("handling nil writers", func() {
			It("[TC-WR-001] should skip nil writers", func() {
				var buf bytes.Buffer
				m.AddWriter(nil, &buf, nil)

				n, err := m.Write([]byte("test"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
				Expect(buf.String()).To(Equal("test"))
			})

			It("[TC-WR-001] should handle all nil writers gracefully", func() {
				m.AddWriter(nil, nil, nil)

				// Should use io.Discard
				n, err := m.Write([]byte("discarded"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(9))
			})
		})

		Context("no writers added", func() {
			It("[TC-WR-001] should use io.Discard by default", func() {
				// No writers added
				n, err := m.Write([]byte("test"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
			})
		})
	})

	Describe("Writer", func() {
		Context("getting writer instance", func() {
			It("[TC-WR-005] should return non-nil writer", func() {
				writer := m.Writer()
				Expect(writer).NotTo(BeNil())
			})

			It("[TC-WR-005] should return writer after adding writers", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				writer := m.Writer()
				Expect(writer).NotTo(BeNil())

				n, err := writer.Write([]byte("direct"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(6))
				Expect(buf.String()).To(Equal("direct"))
			})
		})
	})

	Describe("Write", func() {
		Context("writing to single writer", func() {
			It("[TC-WR-002] should write data successfully", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.Write([]byte("hello world"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(11))
				Expect(buf.String()).To(Equal("hello world"))
			})

			It("[TC-WR-002] should handle empty write", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.Write([]byte{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("[TC-WR-002] should handle multiple writes", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				m.Write([]byte("first "))
				m.Write([]byte("second "))
				m.Write([]byte("third"))

				Expect(buf.String()).To(Equal("first second third"))
			})
		})

		Context("writing to multiple writers", func() {
			It("[TC-WR-002] should write to all writers", func() {
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)

				data := []byte("replicated data")
				n, err := m.Write(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(data)))

				Expect(buf1.String()).To(Equal("replicated data"))
				Expect(buf2.String()).To(Equal("replicated data"))
				Expect(buf3.String()).To(Equal("replicated data"))
			})
		})

		Context("writing large data", func() {
			It("[TC-WR-002] should handle large writes", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				largeData := make([]byte, 1024*1024) // 1MB
				for i := range largeData {
					largeData[i] = byte(i % 256)
				}

				n, err := m.Write(largeData)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(largeData)))
				Expect(buf.Len()).To(Equal(len(largeData)))
			})
		})
	})

	Describe("WriteString", func() {
		Context("writing strings", func() {
			It("[TC-WR-003] should write string successfully", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.WriteString("string data")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(11))
				Expect(buf.String()).To(Equal("string data"))
			})

			It("[TC-WR-003] should handle empty string", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.WriteString("")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("[TC-WR-003] should write to multiple writers", func() {
				var buf1, buf2 bytes.Buffer
				m.AddWriter(&buf1, &buf2)

				n, err := m.WriteString("broadcast string")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(16))
				Expect(buf1.String()).To(Equal("broadcast string"))
				Expect(buf2.String()).To(Equal("broadcast string"))
			})

			It("[TC-WR-003] should handle Unicode strings", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				unicodeStr := "Hello 世界 🌍"
				n, err := m.WriteString(unicodeStr)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal(unicodeStr))
			})
		})
	})

	Describe("Clean", func() {
		Context("cleaning writers", func() {
			It("[TC-WR-004] should remove all writers", func() {
				var buf1, buf2 bytes.Buffer
				m.AddWriter(&buf1, &buf2)

				// Write before clean
				m.Write([]byte("before"))
				Expect(buf1.String()).To(Equal("before"))
				Expect(buf2.String()).To(Equal("before"))

				// Clean writers
				m.Clean()

				// Write after clean - should go to io.Discard
				m.Write([]byte("after"))
				Expect(buf1.String()).To(Equal("before")) // unchanged
				Expect(buf2.String()).To(Equal("before")) // unchanged
			})

			It("[TC-WR-004] should reset writer count", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.Clean()

				// Add writer again after clean
				var newBuf bytes.Buffer
				m.AddWriter(&newBuf)

				m.Write([]byte("test"))
				Expect(newBuf.String()).To(Equal("test"))
				Expect(buf.String()).To(BeEmpty())
			})

			It("[TC-WR-004] should handle multiple clean calls", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				m.Clean()
				m.Clean() // Second clean should not panic

				m.Write([]byte("test"))
				Expect(buf.String()).To(BeEmpty())
			})
		})

		Context("clean on empty multi", func() {
			It("[TC-WR-004] should handle clean without writers", func() {
				Expect(func() {
					m.Clean()
				}).NotTo(Panic())
			})
		})
	})
})
