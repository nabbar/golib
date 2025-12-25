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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

var _ = Describe("TC-WR-001: Writer Operations", func() {
	Context("TC-WR-010: Compress writer", func() {
		It("TC-WR-011: should compress data to writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write([]byte("Hello, World!"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(13))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("TC-WR-012: should handle multiple writes", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			data := [][]byte{
				[]byte("First "),
				[]byte("Second "),
				[]byte("Third"),
			}

			for _, d := range data {
				n, err := h.Write(d)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(d)))
			}

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("TC-WR-013: should handle large writes", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			largeData := bytes.Repeat([]byte("data "), 10000)
			n, err := h.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("TC-WR-014: should not support Read operation", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			readBuf := make([]byte, 10)
			n, err := h.Read(readBuf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(helper.ErrInvalidSource))
		})

		It("TC-WR-015: should compress empty data", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))
		})
	})

	Context("TC-WR-020: Decompress writer", func() {
		It("TC-WR-021: should decompress data to writer", func() {
			compressed := []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd, 0xc9, 0xc9, 0x07,
				0x00, 0x82, 0x89, 0xd1, 0xf7, 0x05, 0x00, 0x00,
				0x00,
			}

			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write(compressed)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(compressed)))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(Equal("Hello"))
		})

		It("TC-WR-022: should handle multiple writes", func() {
			compressed := []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd, 0xc9, 0xc9, 0x07,
				0x00, 0x82, 0x89, 0xd1, 0xf7, 0x05, 0x00, 0x00,
				0x00,
			}

			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())

			mid := len(compressed) / 2
			n1, err := h.Write(compressed[:mid])
			Expect(err).ToNot(HaveOccurred())
			Expect(n1).To(Equal(mid))

			n2, err := h.Write(compressed[mid:])
			Expect(err).ToNot(HaveOccurred())
			Expect(n2).To(Equal(len(compressed) - mid))

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-023: should not support Read operation", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			readBuf := make([]byte, 10)
			n, err := h.Read(readBuf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(helper.ErrInvalidSource))
		})
	})

	Context("TC-WR-030: Close operations", func() {
		It("TC-WR-031: should close compress writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())

			h.Write([]byte("test"))
			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-032: should close decompress writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-033: should return error when writing to closed decompress writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())

			n, err := h.Write([]byte("test"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(helper.ErrClosedResource))
		})
	})

	Context("TC-WR-040: Round-trip operations", func() {
		It("TC-WR-041: should compress and decompress correctly", func() {
			original := "Test data for round trip"

			var compressed bytes.Buffer
			cw, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &compressed)
			Expect(err).ToNot(HaveOccurred())

			_, err = cw.Write([]byte(original))
			Expect(err).ToNot(HaveOccurred())
			err = cw.Close()
			Expect(err).ToNot(HaveOccurred())

			var decompressed bytes.Buffer
			dw, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &decompressed)
			Expect(err).ToNot(HaveOccurred())

			_, err = dw.Write(compressed.Bytes())
			Expect(err).ToNot(HaveOccurred())
			err = dw.Close()
			Expect(err).ToNot(HaveOccurred())

			Expect(decompressed.String()).To(Equal(original))
		})

		It("TC-WR-042: should handle io.Copy", func() {
			original := "Data for io.Copy test"

			var compressed bytes.Buffer
			cw, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &compressed)
			Expect(err).ToNot(HaveOccurred())

			_, err = io.Copy(cw, bytes.NewReader([]byte(original)))
			Expect(err).ToNot(HaveOccurred())
			err = cw.Close()
			Expect(err).ToNot(HaveOccurred())

			Expect(compressed.Len()).To(BeNumerically(">", 0))
		})
	})
})
