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

var _ = Describe("TC-RD-001: Reader Operations", func() {
	Context("TC-RD-010: Compress reader", func() {
		It("TC-RD-011: should compress data from reader", func() {
			input := "Hello, World!"
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			compressed, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(compressed)).To(BeNumerically(">", 0))
			Expect(len(compressed)).To(BeNumerically("<", len(input)*5))
		})

		It("TC-RD-012: should handle multiple reads", func() {
			input := strings.Repeat("test data ", 100)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			buf := make([]byte, 64)
			totalRead := 0
			for {
				n, err := h.Read(buf)
				totalRead += n
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(totalRead).To(BeNumerically(">", 0))
		})

		It("TC-RD-013: should handle small buffer reads", func() {
			input := "Test"
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			buf := make([]byte, 1)
			n, err := h.Read(buf)
			Expect(n).To(Equal(1))
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-RD-014: should handle large data", func() {
			input := strings.Repeat("large data chunk ", 10000)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			compressed, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(compressed)).To(BeNumerically(">", 0))
		})

		It("TC-RD-015: should return EOF when source exhausted", func() {
			input := "short"
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			_, _ = io.ReadAll(h)
			buf := make([]byte, 10)
			n, err := h.Read(buf)
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})

		It("TC-RD-016: should not support Write operation", func() {
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			n, err := h.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(helper.ErrInvalidSource))
		})
	})

	Context("TC-RD-020: Decompress reader", func() {
		It("TC-RD-021: should decompress gzip data", func() {
			original := "Hello, World!"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			data, err := io.ReadAll(h)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(original))
		})

		It("TC-RD-022: should handle multiple reads", func() {
			original := "Test"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			readBuf := make([]byte, 2)
			n1, err := h.Read(readBuf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n1).To(BeNumerically(">", 0))

			_, err = h.Read(readBuf)
			if err != io.EOF {
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("TC-RD-023: should not support Write operation", func() {
			original := "test"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			n, err := h.Write([]byte("data"))
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(helper.ErrInvalidSource))
		})
	})

	Context("TC-RD-030: Close operations", func() {
		It("TC-RD-031: should close compress reader", func() {
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-RD-032: should close decompress reader", func() {
			original := "test"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-RD-033: should handle double close", func() {
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())

			err = h.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
