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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

var _ = Describe("TC-CN-001: Helper Constructor Tests", func() {
	Context("TC-CN-010: New function", func() {
		It("TC-CN-011: should create compress reader from io.Reader", func() {
			r := strings.NewReader("test data")
			h, err := helper.New(arccmp.Gzip, helper.Compress, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-012: should create compress writer from io.Writer", func() {
			var buf bytes.Buffer
			h, err := helper.New(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-013: should create decompress reader from io.Reader", func() {
			original := "test"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.New(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-014: decompress writer creation without data - known limitation", func() {
			var buf bytes.Buffer
			h, err := helper.New(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).To(HaveOccurred())
			Expect(h).To(BeNil())
		})

		It("TC-CN-015: should return error for invalid source", func() {
			h, err := helper.New(arccmp.Gzip, helper.Compress, "invalid")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(helper.ErrInvalidSource))
			Expect(h).To(BeNil())
		})

		It("TC-CN-016: should return error for nil source", func() {
			h, err := helper.New(arccmp.Gzip, helper.Compress, nil)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(helper.ErrInvalidSource))
			Expect(h).To(BeNil())
		})
	})

	Context("TC-CN-020: NewReader function", func() {
		It("TC-CN-021: should create compress reader", func() {
			r := strings.NewReader("test data")
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, r)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-022: should create decompress reader", func() {
			original := "test"
			var buf bytes.Buffer
			cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			cw.Write([]byte(original))
			cw.Close()

			h, err := helper.NewReader(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-023: should return error for invalid operation", func() {
			r := strings.NewReader("test")
			h, err := helper.NewReader(arccmp.Gzip, 99, r)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(helper.ErrInvalidOperation))
			Expect(h).To(BeNil())
		})

		It("TC-CN-024: should work with io.ReadCloser", func() {
			rc := newLimitReader(strings.NewReader("test data"), 100)
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, rc)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})
	})

	Context("TC-CN-030: NewWriter function", func() {
		It("TC-CN-031: should create compress writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-032: should create decompress writer", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, helper.Decompress, &buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})

		It("TC-CN-033: should return error for invalid operation", func() {
			var buf bytes.Buffer
			h, err := helper.NewWriter(arccmp.Gzip, 99, &buf)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(helper.ErrInvalidOperation))
			Expect(h).To(BeNil())
		})

		It("TC-CN-034: should work with io.WriteCloser", func() {
			var buf bytes.Buffer
			wc := newCountWriter(&buf)
			h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, wc)
			Expect(err).ToNot(HaveOccurred())
			Expect(h).ToNot(BeNil())
			defer h.Close()
		})
	})

	Context("TC-CN-040: Interface compliance", func() {
		It("TC-CN-041: should implement io.ReadWriteCloser", func() {
			r := strings.NewReader("test")
			h, err := helper.NewReader(arccmp.Gzip, helper.Compress, r)
			Expect(err).ToNot(HaveOccurred())
			defer h.Close()

			var _ helper.Helper = h
		})
	})
})
