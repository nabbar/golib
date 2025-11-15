/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package iowrapper_test

import (
	"errors"
	"io"
	"strings"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// errorReader always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// errorWriter always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

// errorSeeker always returns an error
type errorSeeker struct{}

func (e *errorSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek error")
}

// errorCloser always returns an error
type errorCloser struct{}

func (e *errorCloser) Close() error {
	return errors.New("close error")
}

// partialErrorReader returns data but with error
type partialErrorReader struct {
	data []byte
	pos  int
}

func (p *partialErrorReader) Read(buf []byte) (n int, err error) {
	if p.pos >= len(p.data) {
		return 0, io.EOF
	}
	n = copy(buf, p.data[p.pos:])
	p.pos += n
	if p.pos >= len(p.data) {
		return n, io.EOF
	}
	return n, nil
}

var _ = Describe("IOWrapper - Error Handling", func() {
	Context("Read errors", func() {
		It("should propagate read errors from underlying reader", func() {
			reader := &errorReader{}
			wrapper := New(reader)

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			// The error is swallowed but we get 0 bytes
			Expect(n).To(Equal(0))
			Expect(err).ToNot(HaveOccurred()) // Returns data read, not error
		})

		It("should handle nil return from custom read function", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Read(make([]byte, 10))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle EOF from underlying reader", func() {
			reader := strings.NewReader("")
			wrapper := New(reader)

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			Expect(n).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle partial reads with errors", func() {
			reader := &partialErrorReader{data: []byte("test")}
			wrapper := New(reader)

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(string(data[:n])).To(Equal("test"))
		})
	})

	Context("Write errors", func() {
		It("should handle write errors from underlying writer", func() {
			writer := &errorWriter{}
			wrapper := New(writer)

			n, err := wrapper.Write([]byte("test"))

			// The error is swallowed but we get 0 bytes written
			Expect(n).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil return from custom write function", func() {
			wrapper := New(nil)

			wrapper.SetWrite(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Write([]byte("test"))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle custom write function that returns error indication", func() {
			wrapper := New(nil)

			wrapper.SetWrite(func(p []byte) []byte {
				// Return empty to indicate error
				return []byte{}
			})

			n, err := wrapper.Write([]byte("test"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("Seek errors", func() {
		It("should propagate seek errors from custom function", func() {
			wrapper := New(nil)

			expectedErr := errors.New("custom seek error")
			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				return 0, expectedErr
			})

			_, err := wrapper.Seek(0, io.SeekStart)

			Expect(err).To(Equal(expectedErr))
		})

		It("should handle seek on non-seekable underlying object", func() {
			nonSeeker := &errorReader{}
			wrapper := New(nonSeeker)

			_, err := wrapper.Seek(0, io.SeekStart)

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle invalid whence values gracefully", func() {
			wrapper := New(nil)

			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				if whence < 0 || whence > 2 {
					return 0, errors.New("invalid whence")
				}
				return offset, nil
			})

			_, err := wrapper.Seek(0, 999)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid whence"))
		})
	})

	Context("Close errors", func() {
		It("should propagate close errors from custom function", func() {
			wrapper := New(nil)

			expectedErr := errors.New("custom close error")
			wrapper.SetClose(func() error {
				return expectedErr
			})

			err := wrapper.Close()

			Expect(err).To(Equal(expectedErr))
		})

		It("should handle close on non-closeable underlying object", func() {
			nonCloser := &errorReader{}
			wrapper := New(nonCloser)

			err := wrapper.Close()

			// Non-closeable objects return nil
			Expect(err).ToNot(HaveOccurred())
		})

		It("should propagate close errors from underlying closer", func() {
			closer := &errorCloser{}
			wrapper := New(closer)

			err := wrapper.Close()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("close error"))
		})
	})

	Context("Multiple error scenarios", func() {
		It("should handle errors in sequence of operations", func() {
			wrapper := New(nil)

			// Read error
			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})
			_, err := wrapper.Read(make([]byte, 10))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Write error
			wrapper.SetWrite(func(p []byte) []byte {
				return nil
			})
			_, err = wrapper.Write([]byte("test"))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Seek error
			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				return 0, errors.New("seek fail")
			})
			_, err = wrapper.Seek(0, io.SeekStart)
			Expect(err).To(HaveOccurred())

			// Close error
			wrapper.SetClose(func() error {
				return errors.New("close fail")
			})
			err = wrapper.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should recover from errors with function replacement", func() {
			wrapper := New(nil)

			// Initially failing read
			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})
			_, err := wrapper.Read(make([]byte, 10))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Fix with new function
			wrapper.SetRead(func(p []byte) []byte {
				return []byte("fixed")
			})
			data := make([]byte, 10)
			n, err := wrapper.Read(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data[:n])).To(Equal("fixed"))
		})
	})

	Context("Error edge cases", func() {
		It("should handle panic recovery in custom functions", func() {
			wrapper := New(nil)

			// This should not panic the test
			wrapper.SetRead(func(p []byte) []byte {
				// Return valid data
				return []byte("ok")
			})

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(2))
		})

		It("should handle error with empty buffer", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Read([]byte{})

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle multiple consecutive errors", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})

			// Multiple reads should all fail consistently
			for i := 0; i < 5; i++ {
				_, err := wrapper.Read(make([]byte, 10))
				Expect(err).To(Equal(io.ErrUnexpectedEOF))
			}
		})
	})

	Context("Standard io errors", func() {
		It("should handle io.EOF correctly", func() {
			reader := strings.NewReader("ab")
			wrapper := New(reader)

			buf := make([]byte, 10)
			n, err := wrapper.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(2))

			// Second read should give EOF
			n, err = wrapper.Read(buf)
			Expect(n).To(Equal(0))
		})

		It("should handle io.ErrUnexpectedEOF", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Read(make([]byte, 10))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should work with io.ReadAll", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				// Simulate EOF by returning empty
				return []byte{}
			})

			data := make([]byte, 10)
			n, _ := wrapper.Read(data)
			Expect(n).To(Equal(0))
		})
	})
})
