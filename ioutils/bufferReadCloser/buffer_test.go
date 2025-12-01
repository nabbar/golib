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

package bufferReadCloser_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/nabbar/golib/ioutils/bufferReadCloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Buffer tests verify the bytes.Buffer wrapper with io.Closer support.
// Tests cover:
//   - Creation with and without custom close functions
//   - All read operations (Read, ReadByte, ReadRune, ReadFrom)
//   - All write operations (Write, WriteString, WriteByte, WriteTo)
//   - Close behavior (reset + custom function execution)
//   - Nil parameter handling (creates empty buffer)
//   - Edge cases (empty buffers, large data, multiple closes)
var _ = Describe("Buffer", func() {
	// Creation tests verify that buffers can be created with various configurations
	// and that nil parameters are handled gracefully.
	Context("Creation", func() {
		It("should create buffer from bytes.Buffer", func() {
			b := bytes.NewBufferString("test data")
			buf := NewBuffer(b, nil)

			Expect(buf).ToNot(BeNil())
		})

		It("should create buffer with custom close function", func() {
			b := bytes.NewBufferString("test")
			closed := false

			buf := NewBuffer(b, func() error {
				closed = true
				return nil
			})

			Expect(buf).ToNot(BeNil())
			err := buf.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closed).To(BeTrue())
		})

		It("should create buffer using deprecated New function", func() {
			b := bytes.NewBufferString("test")
			buf := New(b)

			Expect(buf).ToNot(BeNil())
		})

		It("should create empty buffer when buffer is nil", func() {
			buf := NewBuffer(nil, nil)
			Expect(buf).ToNot(BeNil())

			// Should be able to write and read
			n, err := buf.WriteString("test")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))

			data := make([]byte, 4)
			n, err = buf.Read(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(string(data)).To(Equal("test"))
		})

		It("should create empty buffer when buffer is nil using deprecated New", func() {
			buf := New(nil)
			Expect(buf).ToNot(BeNil())

			// Should be able to write
			n, err := buf.WriteString("test")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})
	})

	// Read operations tests verify that all read methods delegate correctly
	// to the underlying bytes.Buffer.
	Context("Read operations", func() {
		It("should read data", func() {
			b := bytes.NewBufferString("hello world")
			buf := NewBuffer(b, nil)

			data := make([]byte, 5)
			n, err := buf.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data)).To(Equal("hello"))
		})

		It("should read byte", func() {
			b := bytes.NewBufferString("abc")
			buf := NewBuffer(b, nil)

			bt, err := buf.ReadByte()
			Expect(err).ToNot(HaveOccurred())
			Expect(bt).To(Equal(byte('a')))
		})

		It("should read rune", func() {
			b := bytes.NewBufferString("café")
			buf := NewBuffer(b, nil)

			r, size, err := buf.ReadRune()
			Expect(err).ToNot(HaveOccurred())
			Expect(r).To(Equal('c'))
			Expect(size).To(Equal(1))
		})

		It("should read from reader", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			source := strings.NewReader("data from reader")
			n, err := buf.ReadFrom(source)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(16)))
			Expect(b.String()).To(Equal("data from reader"))
		})
	})

	// Write operations tests verify that all write methods delegate correctly
	// to the underlying bytes.Buffer.
	Context("Write operations", func() {
		It("should write data", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			n, err := buf.Write([]byte("test write"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(10))
			Expect(b.String()).To(Equal("test write"))
		})

		It("should write string", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			n, err := buf.WriteString("hello")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(b.String()).To(Equal("hello"))
		})

		It("should write byte", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			err := buf.WriteByte('X')
			Expect(err).ToNot(HaveOccurred())
			Expect(b.String()).To(Equal("X"))
		})

		It("should write to writer", func() {
			b := bytes.NewBufferString("source data")
			buf := NewBuffer(b, nil)

			dest := &bytes.Buffer{}
			n, err := buf.WriteTo(dest)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(11)))
			Expect(dest.String()).To(Equal("source data"))
		})
	})

	// Combined operations tests verify that read and write can be used together
	// on the same buffer.
	Context("Combined operations", func() {
		It("should support read and write", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			// Write
			_, err := buf.WriteString("test")
			Expect(err).ToNot(HaveOccurred())

			// Read
			data := make([]byte, 4)
			n, err := buf.Read(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(string(data)).To(Equal("test"))
		})
	})

	// Close operations tests verify that Close() properly resets the buffer
	// and executes custom close functions, including error propagation.
	Context("Close operations", func() {
		It("should close and reset buffer", func() {
			b := bytes.NewBufferString("data")
			buf := NewBuffer(b, nil)

			err := buf.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(b.Len()).To(Equal(0))
		})

		It("should call custom close function", func() {
			b := bytes.NewBuffer(nil)
			closeCalled := false

			buf := NewBuffer(b, func() error {
				closeCalled = true
				return nil
			})

			err := buf.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closeCalled).To(BeTrue())
		})

		It("should return close function error", func() {
			b := bytes.NewBuffer(nil)
			expectedErr := errors.New("close error")

			buf := NewBuffer(b, func() error {
				return expectedErr
			})

			err := buf.Close()
			Expect(err).To(Equal(expectedErr))
		})

		It("should be safe to close multiple times", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			err1 := buf.Close()
			err2 := buf.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	// Edge cases tests verify behavior with unusual inputs like empty buffers
	// and very large data.
	Context("Edge cases", func() {
		It("should handle empty buffer", func() {
			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			data := make([]byte, 10)
			n, err := buf.Read(data)

			Expect(err).To(HaveOccurred()) // EOF
			Expect(n).To(Equal(0))
		})

		It("should handle large data", func() {
			largeData := make([]byte, 1024*1024) // 1 MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			b := bytes.NewBuffer(nil)
			buf := NewBuffer(b, nil)

			n, err := buf.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
		})
	})
})
