/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

var _ = Describe("Multi Reader Operations", func() {
	var m multi.Multi

	BeforeEach(func() {
		m = multi.New()
	})

	Describe("SetInput", func() {
		Context("setting input reader", func() {
			It("should set input successfully", func() {
				input := io.NopCloser(strings.NewReader("test data"))
				m.SetInput(input)

				buf := make([]byte, 9)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(9))
				Expect(string(buf)).To(Equal("test data"))
			})

			It("should replace previous input", func() {
				input1 := io.NopCloser(strings.NewReader("first"))
				m.SetInput(input1)

				input2 := io.NopCloser(strings.NewReader("second"))
				m.SetInput(input2)

				buf := make([]byte, 6)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(6))
				Expect(string(buf)).To(Equal("second"))
			})
		})

		Context("setting nil input", func() {
			It("should use DiscardCloser for nil input", func() {
				m.SetInput(nil)

				buf := make([]byte, 10)
				n, err := m.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})
		})

		Context("on nil receiver", func() {
			It("should not panic on nil receiver", func() {
				// The code has a check for o == nil in SetInput
				// This is difficult to test directly through the interface
				Skip("Nil receiver check is internal implementation detail")
			})
		})
	})

	Describe("Reader", func() {
		Context("getting reader instance", func() {
			It("should return reader after SetInput", func() {
				input := io.NopCloser(strings.NewReader("data"))
				m.SetInput(input)

				reader := m.Reader()
				Expect(reader).NotTo(BeNil())

				buf := make([]byte, 4)
				n, err := reader.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(4))
				Expect(string(buf)).To(Equal("data"))
			})
		})
	})

	Describe("Read", func() {
		Context("reading from input", func() {
			It("should read data successfully", func() {
				input := io.NopCloser(strings.NewReader("hello world"))
				m.SetInput(input)

				buf := make([]byte, 11)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(11))
				Expect(string(buf)).To(Equal("hello world"))
			})

			It("should handle partial reads", func() {
				input := io.NopCloser(strings.NewReader("hello world"))
				m.SetInput(input)

				buf := make([]byte, 5)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(5))
				Expect(string(buf)).To(Equal("hello"))

				// Read remaining
				buf2 := make([]byte, 6)
				n, err = m.Read(buf2)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(6))
				Expect(string(buf2)).To(Equal(" world"))
			})

			It("should return EOF when input is exhausted", func() {
				input := io.NopCloser(strings.NewReader("short"))
				m.SetInput(input)

				buf := make([]byte, 10)
				n, err := m.Read(buf)
				Expect(n).To(Equal(5))

				// Next read should return EOF
				n, err = m.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(0))
			})
		})

		Context("reading with default input", func() {
			It("should handle read with default DiscardCloser", func() {
				// With new initialization, a default DiscardCloser is set
				buf := make([]byte, 10)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0)) // DiscardCloser returns 0
			})
		})

		Context("reading large data", func() {
			It("should handle large reads", func() {
				largeData := strings.Repeat("x", 1024*1024) // 1MB
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				buf := make([]byte, 1024*1024)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(1024 * 1024))
			})
		})
	})

	Describe("Close", func() {
		Context("closing input", func() {
			It("should close input successfully", func() {
				input := io.NopCloser(strings.NewReader("test"))
				m.SetInput(input)

				err := m.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should propagate close error", func() {
				// Create a closer that returns an error
				errorCloser := &closeErrorReader{
					Reader:   strings.NewReader("test"),
					closeErr: io.ErrClosedPipe,
				}
				m.SetInput(errorCloser)

				err := m.Close()
				Expect(err).To(Equal(io.ErrClosedPipe))
			})
		})

		Context("closing with default input", func() {
			It("should close successfully with default DiscardCloser", func() {
				// With new initialization, a default DiscardCloser is set
				err := m.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("closing with DiscardCloser", func() {
			It("should close successfully with nil input", func() {
				m.SetInput(nil)

				err := m.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

// closeErrorReader is a test helper that returns an error on Close
type closeErrorReader struct {
	io.Reader
	closeErr error
}

func (e *closeErrorReader) Close() error {
	return e.closeErr
}
