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
	"errors"
	"io"

	. "github.com/nabbar/golib/ioutils/nopwritecloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NopWriteCloser - Edge Cases", func() {
	Context("Large data", func() {
		It("should handle large writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			largeData := make([]byte, 10*1024*1024) // 10 MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			n, err := wc.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(buf.Len()).To(Equal(len(largeData)))
		})

		It("should handle many small writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			for i := 0; i < 100000; i++ {
				_, err := wc.Write([]byte("x"))
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(buf.Len()).To(Equal(100000))
		})

		It("should handle variable-sized writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			totalBytes := 0
			for i := 1; i <= 1000; i++ {
				data := make([]byte, i)
				n, err := wc.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(i))
				totalBytes += i
			}

			Expect(buf.Len()).To(Equal(totalBytes))
		})
	})

	Context("Error propagation", func() {
		It("should propagate write errors from underlying writer", func() {
			expectedErr := errors.New("write failed")
			errWriter := &errorWriter{err: expectedErr}
			wc := New(errWriter)

			_, err := wc.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		})

		It("should handle errors after successful writes", func() {
			limitedWriter := &limitedErrorWriter{
				remaining: 3,
				err:       errors.New("quota exceeded"),
			}
			wc := New(limitedWriter)

			// First 3 writes should succeed
			for i := 0; i < 3; i++ {
				_, err := wc.Write([]byte("ok"))
				Expect(err).ToNot(HaveOccurred())
			}

			// 4th write should fail
			_, err := wc.Write([]byte("fail"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("quota exceeded"))
		})

		It("should allow close even if writes failed", func() {
			errWriter := &errorWriter{err: errors.New("write error")}
			wc := New(errWriter)

			_, writeErr := wc.Write([]byte("test"))
			Expect(writeErr).To(HaveOccurred())

			// Close should still succeed
			closeErr := wc.Close()
			Expect(closeErr).ToNot(HaveOccurred())
		})
	})

	Context("Special writers", func() {
		It("should work with io.Discard", func() {
			wc := New(io.Discard)

			data := []byte("this will be discarded")
			n, err := wc.Write(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			err = wc.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should work with counting writer", func() {
			counter := &countingWriter{}
			wc := New(counter)

			for i := 0; i < 10; i++ {
				wc.Write([]byte("test"))
			}

			Expect(counter.count).To(Equal(10))
			Expect(counter.buf.Len()).To(Equal(40))
		})

		It("should work with multiple nopwritecloser layers", func() {
			buf := &bytes.Buffer{}
			wc1 := New(buf)
			wc2 := New(wc1)
			wc3 := New(wc2)

			_, err := wc3.Write([]byte("nested"))
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(Equal("nested"))

			err = wc3.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Boundary conditions", func() {
		It("should handle zero-length buffer", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			n, err := wc.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(buf.Len()).To(Equal(0))
		})

		It("should handle single byte writes", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			for i := byte(0); i < 255; i++ {
				n, err := wc.Write([]byte{i})
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(1))
			}

			Expect(buf.Len()).To(Equal(255))
		})

		It("should handle maximum int-sized write", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			// Large but reasonable size
			size := 100 * 1024 * 1024 // 100 MB
			data := make([]byte, size)

			n, err := wc.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(size))
		})
	})

	Context("State transitions", func() {
		It("should handle write-close-write-close pattern", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			wc.Write([]byte("1"))
			wc.Close()
			wc.Write([]byte("2"))
			wc.Close()
			wc.Write([]byte("3"))
			wc.Close()

			Expect(buf.String()).To(Equal("123"))
		})

		It("should handle close-write-close-write pattern", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			wc.Close()
			wc.Write([]byte("1"))
			wc.Close()
			wc.Write([]byte("2"))

			Expect(buf.String()).To(Equal("12"))
		})

		It("should handle repeated close calls", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			for i := 0; i < 100; i++ {
				err := wc.Close()
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})

	Context("Type compatibility", func() {
		It("should be assignable to io.Writer", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			var w io.Writer = wc
			n, err := w.Write([]byte("test"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should be assignable to io.Closer", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			var c io.Closer = wc
			err := c.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should be assignable to io.WriteCloser", func() {
			buf := &bytes.Buffer{}
			wc := New(buf)

			var wcInterface io.WriteCloser = wc

			n, err := wcInterface.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))

			err = wcInterface.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
