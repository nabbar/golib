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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

var _ = Describe("TC-EC-001: Edge Cases and Boundary Tests", func() {
	Context("TC-EC-010: Empty data", func() {
		It("TC-EC-011: should compress empty reader", func() {
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(""))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			compressed, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(compressed)).To(BeNumerically(">", 0))
		})

		It("TC-EC-012: should compress empty writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("TC-EC-013: should handle zero-length write", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("TC-EC-020: Large data", func() {
		It("TC-EC-021: should compress very large data", func() {
			largeData := strings.Repeat("test data with some variety ", 100000)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(largeData))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			compressed, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(compressed)).To(BeNumerically(">", 0))
			Expect(len(compressed)).To(BeNumerically("<", len(largeData)))
		})

		It("TC-EC-022: should write very large data", func() {
			largeData := bytes.Repeat([]byte("data "), 100000)
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-023: should handle many small reads", func() {
			data := strings.Repeat("x", 10000)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			buf := make([]byte, 1)
			count := 0
			for {
				_, err := h.Read(buf)
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
				count++
				if count > 20000 {
					break
				}
			}
		})
	})

	Context("TC-EC-030: Special characters and binary data", func() {
		It("TC-EC-031: should handle binary data", func() {
			binaryData := make([]byte, 256)
			for i := range binaryData {
				binaryData[i] = byte(i)
			}

			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write(binaryData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(binaryData)))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-032: should handle null bytes", func() {
			data := []byte{0x00, 0x01, 0x00, 0x02, 0x00}
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-033: should handle unicode data", func() {
			unicodeData := "Hello 世界 🌍 Привет"
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(unicodeData))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			compressed, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(compressed)).To(BeNumerically(">", 0))
		})
	})

	Context("TC-EC-040: Error conditions", func() {
		It("TC-EC-041: should handle invalid compressed data", func() {
			invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, bytes.NewReader(invalidData))
			if err != nil {
				Expect(err).To(HaveOccurred())
				return
			}
			defer h.Close()

			_, err = io.ReadAll(h)
			Expect(err).To(HaveOccurred())
		})

		It("TC-EC-042: should handle truncated compressed data", func() {
			truncated := []byte{0x1f, 0x8b, 0x08}
			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, bytes.NewReader(truncated))
			if err != nil {
				Expect(err).To(HaveOccurred())
				return
			}
			defer h.Close()

			_, err = io.ReadAll(h)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("TC-EC-050: Buffer boundary conditions", func() {
		It("TC-EC-051: should handle reads at chunk boundaries", func() {
			data := strings.Repeat("a", 512)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			buf := make([]byte, 512)
			_, err = h.Read(buf)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-052: should handle writes at chunk boundaries", func() {
			data := bytes.Repeat([]byte("b"), 512)
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(512))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-053: should handle reads larger than internal buffer", func() {
			data := strings.Repeat("c", 2048)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			buf := make([]byte, 4096)
			n, err := h.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", 0))
		})
	})

	Context("TC-EC-060: Repeated operations", func() {
		It("TC-EC-061: should handle repeated small writes", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 1000; i++ {
				_, err = h.Write([]byte("x"))
				Expect(err).ToNot(HaveOccurred())
			}

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-EC-062: should handle alternating read sizes", func() {
			data := strings.Repeat("test", 1000)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			sizes := []int{1, 10, 100, 10, 1}
			for _, size := range sizes {
				buf := make([]byte, size)
				_, err := h.Read(buf)
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})
})
