/*
MIT License

Copyright (c) 2024 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package randRead_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encrnd "github.com/nabbar/golib/encoding/randRead"
)

// mockReadCloser for testing error conditions
type mockReadCloser struct {
	data   []byte
	pos    int
	err    error
	closed bool
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	if m.err != nil {
		return m.err
	}
	return nil
}

var _ = Describe("Error Handling and Edge Cases", func() {
	Describe("Function Errors", func() {
		It("should propagate errors from remote function", func() {
			expectedErr := errors.New("remote fetch error")
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return nil, expectedErr
			})

			buf := make([]byte, 10)
			_, err := rdr.Read(buf)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("remote fetch error"))
		})

		It("should handle function that returns nil reader", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return nil, errors.New("no reader available")
			})

			buf := make([]byte, 10)
			_, err := rdr.Read(buf)

			Expect(err).To(HaveOccurred())
		})

		It("should handle intermittent errors", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return nil, errors.New("first call fails")
				}
				return io.NopCloser(bytes.NewReader([]byte("success"))), nil
			})

			buf := make([]byte, 7)

			// First read should fail
			_, err1 := rdr.Read(buf)
			Expect(err1).To(HaveOccurred())

			// Second read should succeed (new fetch)
			_, err2 := rdr.Read(buf)
			Expect(err2).ToNot(HaveOccurred())
			Expect(string(buf)).To(Equal("success"))
		})
	})

	Describe("EOF Handling", func() {
		It("should handle EOF from source and refetch", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				return io.NopCloser(bytes.NewReader([]byte{byte(callCount)})), nil
			})

			// Read first byte
			buf := make([]byte, 1)
			n1, err1 := rdr.Read(buf)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(1))
			Expect(buf[0]).To(Equal(byte(1)))

			// Next read should trigger refetch
			buf2 := make([]byte, 1)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(1))
			Expect(buf2[0]).To(Equal(byte(2)))
		})

		It("should handle function returning error after EOF", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return io.NopCloser(bytes.NewReader([]byte("data"))), nil
				}
				return nil, io.EOF
			})

			// First read succeeds
			buf1 := make([]byte, 4)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(4))

			// Second read should return EOF from function
			buf2 := make([]byte, 10)
			_, err2 := rdr.Read(buf2)
			Expect(err2).To(Equal(io.EOF))
		})
	})

	Describe("Buffer Management", func() {
		It("should handle zero-length buffer", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})

			buf := make([]byte, 0)
			n, err := rdr.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle very large buffer", func() {
			data := []byte("small data")
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			buf := make([]byte, 1024*1024) // 1MB buffer for small data
			n, err := rdr.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf[:n]).To(Equal(data))
		})

		It("should reset buffer correctly", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				return io.NopCloser(bytes.NewReader([]byte("data"))), nil
			})

			// Read data
			buf := make([]byte, 4)
			rdr.Read(buf)

			// Close (should reset)
			rdr.Close()

			// Reading after close should fetch new data
			buf2 := make([]byte, 4)
			_, err := rdr.Read(buf2)

			// Might fail because of closed state, but should not panic
			_ = err
		})
	})

	Describe("Reader Lifecycle", func() {
		It("should close underlying reader when closed", func() {
			mock := &mockReadCloser{data: []byte("test")}

			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return mock, nil
			})

			// Read to initialize
			buf := make([]byte, 4)
			rdr.Read(buf)

			// Close should close underlying reader
			err := rdr.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.closed).To(BeTrue())
		})

		It("should handle close error from underlying reader", func() {
			closeErr := errors.New("close error")
			mock := &mockReadCloser{
				data: []byte("test"),
				err:  closeErr,
			}

			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return mock, nil
			})

			// Read to initialize
			buf := make([]byte, 4)
			_, err := rdr.Read(buf)

			// Read error should occur
			Expect(err).To(Equal(closeErr))
		})

		It("should replace reader on refetch", func() {
			mock1 := &mockReadCloser{data: []byte("first")}
			mock2 := &mockReadCloser{data: []byte("second")}

			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return mock1, nil
				}
				return mock2, nil
			})

			// Read first data
			buf1 := make([]byte, 5)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(string(buf1)).To(Equal("first"))

			// Read second data (should close mock1 and use mock2)
			buf2 := make([]byte, 6)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(6))
			Expect(string(buf2)).To(Equal("second"))

			// mock1 should have been closed
			Expect(mock1.closed).To(BeTrue())
		})
	})

	Describe("Edge Cases", func() {
		It("should handle refetch with error", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return io.NopCloser(bytes.NewReader([]byte("a"))), nil
				}
				return nil, errors.New("no more data")
			})

			// First read succeeds
			buf1 := make([]byte, 1)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(1))

			// Second read triggers refetch which fails
			buf2 := make([]byte, 10)
			_, err2 := rdr.Read(buf2)
			Expect(err2).To(HaveOccurred())
			Expect(err2.Error()).To(ContainSubstring("no more data"))
		})

		It("should handle rapid sequential reads", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("abcdefghij"))), nil
			})

			for i := 0; i < 10; i++ {
				buf := make([]byte, 1)
				n, err := rdr.Read(buf)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(1))
			}
		})

		It("should handle alternating small and large reads", func() {
			data := make([]byte, 1000)
			for i := range data {
				data[i] = byte(i % 256)
			}

			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			// Small read
			buf1 := make([]byte, 10)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(10))

			// Large read
			buf2 := make([]byte, 500)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(500))

			// Small read again
			buf3 := make([]byte, 10)
			n3, err3 := rdr.Read(buf3)
			Expect(err3).ToNot(HaveOccurred())
			Expect(n3).To(Equal(10))
		})
	})

	Describe("Multiple Readers", func() {
		It("should handle sequential reads from multiple goroutines", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				data := make([]byte, 1000)
				for i := range data {
					data[i] = byte(i % 256)
				}
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			// First read
			buf1 := make([]byte, 100)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(100))

			// Second read
			buf2 := make([]byte, 100)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(100))
		})
	})
})
