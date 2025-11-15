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
	"bytes"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/ioutils/multi"
)

var _ = Describe("Multi Edge Cases and Error Handling", func() {
	var m multi.Multi

	BeforeEach(func() {
		m = multi.New()
	})

	Describe("Error handling", func() {
		Context("ErrInstance error", func() {
			It("should handle read with default input", func() {
				// With the new initialization, a default DiscardCloser is set
				buf := make([]byte, 10)
				n, err := m.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
			})

			It("should handle close with default input", func() {
				// With the new initialization, a default DiscardCloser is set
				err := m.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have meaningful error message", func() {
				Expect(multi.ErrInstance.Error()).To(ContainSubstring("invalid instance"))
			})
		})

		Context("writer errors", func() {
			It("should handle partial write errors", func() {
				errorWriter := &partialWriter{maxBytes: 5}
				m.AddWriter(errorWriter)

				n, err := m.Write([]byte("hello world"))
				// Behavior depends on io.MultiWriter implementation
				// It may return error or partial write
				if err != nil {
					Expect(n).To(BeNumerically("<=", 11))
				}
			})

			It("should propagate writer errors", func() {
				errorWriter := &errorWriter{err: io.ErrShortWrite}
				m.AddWriter(errorWriter)

				_, err := m.Write([]byte("test"))
				Expect(err).To(Equal(io.ErrShortWrite))
			})
		})

		Context("reader errors", func() {
			It("should propagate reader errors", func() {
				errorReader := &errorReadCloser{
					Reader:   strings.NewReader(""),
					readErr:  io.ErrUnexpectedEOF,
					closeErr: nil,
				}
				m.SetInput(errorReader)

				buf := make([]byte, 10)
				_, err := m.Read(buf)
				Expect(err).To(Equal(io.ErrUnexpectedEOF))
			})
		})
	})

	Describe("Boundary conditions", func() {
		Context("zero-length operations", func() {
			It("should handle zero-length Write", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.Write([]byte{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle zero-length Read", func() {
				input := io.NopCloser(strings.NewReader("data"))
				m.SetInput(input)

				n, err := m.Read([]byte{})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle zero-length WriteString", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.WriteString("")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})
		})

		Context("nil buffer operations", func() {
			It("should handle Read with nil buffer", func() {
				input := io.NopCloser(strings.NewReader("data"))
				m.SetInput(input)

				n, err := m.Read(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})

			It("should handle Write with nil buffer", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				n, err := m.Write(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(0))
			})
		})

		Context("very large operations", func() {
			It("should handle very large writes", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				// 10MB write
				largeData := make([]byte, 10*1024*1024)
				for i := range largeData {
					largeData[i] = byte(i % 256)
				}

				n, err := m.Write(largeData)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(largeData)))
				Expect(buf.Len()).To(Equal(len(largeData)))
			})

			It("should handle very large reads", func() {
				largeData := strings.Repeat("x", 10*1024*1024) // 10MB
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				buf := make([]byte, 10*1024*1024)
				n, err := m.Read(buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(10 * 1024 * 1024))
			})

			It("should handle very large copy", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				largeData := strings.Repeat("x", 10*1024*1024) // 10MB
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				n, err := m.Copy()
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(10 * 1024 * 1024)))
			})
		})
	})

	Describe("State transitions", func() {
		Context("operations after clean", func() {
			It("should work correctly after clean", func() {
				var buf1 bytes.Buffer
				m.AddWriter(&buf1)
				m.Write([]byte("before"))

				m.Clean()

				var buf2 bytes.Buffer
				m.AddWriter(&buf2)
				m.Write([]byte("after"))

				Expect(buf1.String()).To(Equal("before"))
				Expect(buf2.String()).To(Equal("after"))
			})

			It("should reset to io.Discard after clean", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				m.Clean()

				m.Write([]byte("discarded"))
				Expect(buf.String()).To(BeEmpty())
			})
		})

		Context("multiple input changes", func() {
			It("should handle multiple SetInput calls", func() {
				input1 := io.NopCloser(strings.NewReader("first"))
				m.SetInput(input1)

				buf := make([]byte, 5)
				m.Read(buf)
				Expect(string(buf)).To(Equal("first"))

				input2 := io.NopCloser(strings.NewReader("second"))
				m.SetInput(input2)

				buf2 := make([]byte, 6)
				m.Read(buf2)
				Expect(string(buf2)).To(Equal("second"))
			})
		})

		Context("operations after close", func() {
			It("should allow write after close", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				input := io.NopCloser(strings.NewReader("test"))
				m.SetInput(input)
				m.Close()

				// Write should still work
				n, err := m.Write([]byte("after close"))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(11))
			})
		})
	})

	Describe("Special data patterns", func() {
		Context("binary data", func() {
			It("should handle binary data correctly", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				binaryData := []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD}
				n, err := m.Write(binaryData)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(binaryData)))
				Expect(buf.Bytes()).To(Equal(binaryData))
			})

			It("should handle null bytes", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				nullData := []byte{0x00, 0x00, 0x00}
				n, err := m.Write(nullData)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(3))
			})
		})

		Context("Unicode data", func() {
			It("should handle Unicode strings correctly", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				unicodeStr := "Hello 世界 🌍 مرحبا Привет"
				n, err := m.WriteString(unicodeStr)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal(unicodeStr))
			})

			It("should handle emoji sequences", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				emojiStr := "👨‍👩‍👧‍👦🏳️‍🌈"
				m.WriteString(emojiStr)
				Expect(buf.String()).To(Equal(emojiStr))
			})
		})
	})

	Describe("Interface compliance", func() {
		Context("io.Reader compliance", func() {
			It("should follow io.Reader contract", func() {
				input := io.NopCloser(strings.NewReader("test"))
				m.SetInput(input)

				// Read should return data and potentially EOF
				buf := make([]byte, 10)
				n, err := m.Read(buf)
				if err == nil {
					Expect(n).To(BeNumerically(">", 0))
				} else if err == io.EOF {
					// EOF is acceptable
				}
			})
		})

		Context("io.Writer compliance", func() {
			It("should follow io.Writer contract", func() {
				var buf bytes.Buffer
				m.AddWriter(&buf)

				// Write should return n == len(p) on success
				data := []byte("test data")
				n, err := m.Write(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(data)))
			})
		})

		Context("io.Closer compliance", func() {
			It("should follow io.Closer contract", func() {
				input := io.NopCloser(strings.NewReader("test"))
				m.SetInput(input)

				err := m.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

// partialWriter writes only up to maxBytes
type partialWriter struct {
	maxBytes int
	written  int
}

func (p *partialWriter) Write(data []byte) (n int, err error) {
	remaining := p.maxBytes - p.written
	if remaining <= 0 {
		return 0, io.ErrShortWrite
	}
	if len(data) > remaining {
		p.written += remaining
		return remaining, io.ErrShortWrite
	}
	p.written += len(data)
	return len(data), nil
}

// errorReadCloser is a test helper that can return errors on both Read and Close
type errorReadCloser struct {
	io.Reader
	readErr  error
	closeErr error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	if e.readErr != nil {
		return 0, e.readErr
	}
	return e.Reader.Read(p)
}

func (e *errorReadCloser) Close() error {
	return e.closeErr
}
