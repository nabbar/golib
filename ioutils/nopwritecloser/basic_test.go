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

package nopwritecloser_test

import (
	"bytes"
	"io"

	. "github.com/nabbar/golib/ioutils/nopwritecloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NopWriteCloser - Basic Operations", func() {
	Context("Creation", func() {
		It("should create a NopWriteCloser from Writer", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			Expect(wc).ToNot(BeNil())
		})

		It("should implement io.WriteCloser interface", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Type assertion to ensure interface compliance
			var _ io.WriteCloser = wc

			// Verify it can be used as io.Writer
			var _ io.Writer = wc

			// Verify it can be used as io.Closer
			var _ io.Closer = wc
		})

		It("should work with nil writer", func() {
			// While not recommended, it shouldn't panic on creation
			wc := New(nil)
			Expect(wc).ToNot(BeNil())
		})
	})

	Context("Write operations", func() {
		It("should write data correctly", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			data := []byte("test data")
			n, err := wc.Write(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf.String()).To(Equal("test data"))
		})

		It("should support multiple writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			n1, err1 := wc.Write([]byte("hello"))
			n2, err2 := wc.Write([]byte(" "))
			n3, err3 := wc.Write([]byte("world"))

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(err3).ToNot(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(n2).To(Equal(1))
			Expect(n3).To(Equal(5))
			Expect(buf.String()).To(Equal("hello world"))
		})

		It("should handle empty writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			n, err := wc.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(buf.Len()).To(Equal(0))
		})

		It("should handle nil byte slice", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			n, err := wc.Write(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should preserve write order", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			for i := 0; i < 10; i++ {
				_, err := wc.Write([]byte{byte('0' + i)})
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(buf.String()).To(Equal("0123456789"))
		})

		It("should write unicode correctly", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			data := []byte("Hello, 世界! 🌍")
			n, err := wc.Write(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf.String()).To(Equal("Hello, 世界! 🌍"))
		})

		It("should write binary data correctly", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			data := []byte{0x00, 0xFF, 0x7F, 0x80, 0xAA, 0x55}
			n, err := wc.Write(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf.Bytes()).To(Equal(data))
		})
	})

	Context("Close operations", func() {
		It("should close without error", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			err := wc.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should be safe to close multiple times", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			err1 := wc.Close()
			err2 := wc.Close()
			err3 := wc.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(err3).ToNot(HaveOccurred())
		})

		It("should allow writes after close", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			err := wc.Close()
			Expect(err).ToNot(HaveOccurred())

			// NopWriteCloser doesn't actually close the underlying writer
			n, err := wc.Write([]byte("after close"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(11))
			Expect(buf.String()).To(Equal("after close"))
		})

		It("should allow close before any writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			err := wc.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(Equal(0))
		})

		It("should allow interleaved writes and closes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			wc.Write([]byte("first"))
			wc.Close()
			wc.Write([]byte(" second"))
			wc.Close()
			wc.Write([]byte(" third"))

			Expect(buf.String()).To(Equal("first second third"))
		})
	})
})
