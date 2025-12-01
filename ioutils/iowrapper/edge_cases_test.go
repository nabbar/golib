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

// This file tests edge cases and boundary conditions.
//
// Test Strategy:
//   - Test wrapping objects that don't implement any I/O interface
//   - Verify behavior with nil objects and nil custom functions
//   - Test boundary conditions (empty buffers, zero-length reads/writes)
//   - Ensure graceful degradation when operations are not supported
//   - Validate rapid function replacement and updates
//
// Coverage: 18 specs testing edge cases to ensure robustness.
package iowrapper_test

import (
	"io"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// mockNonReader is a struct that doesn't implement any io interfaces
type mockNonReader struct{}

// mockPartialReadWriter implements Reader but not Writer
type mockPartialReadWriter struct {
	data []byte
	pos  int
}

func (m *mockPartialReadWriter) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

var _ = Describe("IOWrapper - Edge Cases", func() {
	Context("Non-interface supporting objects", func() {
		It("should handle object without Reader interface", func() {
			nonReader := &mockNonReader{}
			wrapper := New(nonReader)

			data := make([]byte, 10)
			_, err := wrapper.Read(data)

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle object without Writer interface", func() {
			nonWriter := &mockNonReader{}
			wrapper := New(nonWriter)

			_, err := wrapper.Write([]byte("test"))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle object without Seeker interface", func() {
			nonSeeker := &mockNonReader{}
			wrapper := New(nonSeeker)

			_, err := wrapper.Seek(0, io.SeekStart)

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle object without Closer interface", func() {
			nonCloser := &mockNonReader{}
			wrapper := New(nonCloser)

			err := wrapper.Close()

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Partial interface implementation", func() {
		It("should work with partial Reader implementation", func() {
			partial := &mockPartialReadWriter{data: []byte("hello")}
			wrapper := New(partial)

			data := make([]byte, 5)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data)).To(Equal("hello"))
		})

		It("should fail writing to read-only object", func() {
			partial := &mockPartialReadWriter{data: []byte("hello")}
			wrapper := New(partial)

			_, err := wrapper.Write([]byte("test"))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})
	})

	Context("Buffer size edge cases", func() {
		It("should handle custom read returning more data than buffer size", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				// Return more data than buffer can hold
				return []byte("this is a very long string that exceeds buffer size")
			})

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(10)) // Should only read what fits
			Expect(string(data)).To(Equal("this is a "))
		})

		It("should handle custom read returning less data than buffer size", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return []byte("small")
			})

			data := make([]byte, 100)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data[:n])).To(Equal("small"))
		})

		It("should handle zero-length buffer", func() {
			wrapper := New(nil)

			wrapper.SetRead(func(p []byte) []byte {
				return []byte("data")
			})

			data := make([]byte, 0)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("Nil and empty values", func() {
		It("should handle nil wrapper input", func() {
			wrapper := New(nil)

			Expect(wrapper).ToNot(BeNil())
		})

		It("should handle multiple operations on nil input", func() {
			wrapper := New(nil)

			// Read should fail
			_, err := wrapper.Read(make([]byte, 10))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Write should fail
			_, err = wrapper.Write([]byte("test"))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Seek should fail
			_, err = wrapper.Seek(0, io.SeekStart)
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Close should succeed
			err = wrapper.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle empty slice writes", func() {
			wrapper := New(nil)

			wrapper.SetWrite(func(p []byte) []byte {
				return []byte{}
			})

			n, err := wrapper.Write([]byte("data"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("Seek edge cases", func() {
		It("should handle negative seek offset", func() {
			wrapper := New(nil)

			called := false
			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				called = true
				Expect(offset).To(Equal(int64(-5)))
				return -5, nil
			})

			pos, err := wrapper.Seek(-5, io.SeekCurrent)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(-5)))
			Expect(called).To(BeTrue())
		})

		It("should handle seek beyond end", func() {
			wrapper := New(nil)

			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				return offset, nil
			})

			pos, err := wrapper.Seek(999999, io.SeekStart)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(999999)))
		})

		It("should handle all seek whence values", func() {
			wrapper := New(nil)

			whenceValues := []int{io.SeekStart, io.SeekCurrent, io.SeekEnd}

			for _, whence := range whenceValues {
				called := false
				wrapper.SetSeek(func(offset int64, w int) (int64, error) {
					called = true
					Expect(w).To(Equal(whence))
					return offset, nil
				})

				_, err := wrapper.Seek(10, whence)
				Expect(err).ToNot(HaveOccurred())
				Expect(called).To(BeTrue())
			}
		})
	})

	Context("Function replacement edge cases", func() {
		It("should handle rapid function replacement", func() {
			wrapper := New(nil)

			for i := 0; i < 100; i++ {
				value := i
				wrapper.SetRead(func(p []byte) []byte {
					return []byte{byte(value)}
				})

				data := make([]byte, 1)
				n, err := wrapper.Read(data)

				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(1))
			}
		})

		It("should handle nil function followed by custom function", func() {
			wrapper := New(nil)

			// Set nil first
			wrapper.SetRead(nil)

			// Should use default (and fail)
			_, err := wrapper.Read(make([]byte, 10))
			Expect(err).To(Equal(io.ErrUnexpectedEOF))

			// Now set custom
			wrapper.SetRead(func(p []byte) []byte {
				return []byte("works")
			})

			data := make([]byte, 10)
			n, _ := wrapper.Read(data)
			Expect(string(data[:n])).To(Equal("works"))
		})
	})

	Context("Interface type assertions", func() {
		It("should satisfy io.Reader interface", func() {
			wrapper := New(nil)
			var _ io.Reader = wrapper
		})

		It("should satisfy io.Writer interface", func() {
			wrapper := New(nil)
			var _ io.Writer = wrapper
		})

		It("should satisfy io.Seeker interface", func() {
			wrapper := New(nil)
			var _ io.Seeker = wrapper
		})

		It("should satisfy io.Closer interface", func() {
			wrapper := New(nil)
			var _ io.Closer = wrapper
		})

		It("should satisfy io.ReadWriteSeeker interface", func() {
			wrapper := New(nil)
			var _ io.ReadWriteSeeker = wrapper
		})
	})
})
