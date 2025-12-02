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
	"bufio"
	"bytes"
	"errors"
	"strings"

	. "github.com/nabbar/golib/ioutils/bufferReadCloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Writer tests verify the bufio.Writer wrapper with io.Closer support.
// Tests cover creation, write operations, flush behavior, close with reset,
// custom close functions, error propagation, and nil parameter handling.
var _ = Describe("Writer", func() {
	// Creation tests verify writer instantiation and nil handling (io.Discard).
	Context("Creation", func() {
		It("should create writer from bufio.Writer", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			Expect(writer).ToNot(BeNil())
		})

		It("should create writer with custom close function", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			closed := false

			writer := NewWriter(bw, func() error {
				closed = true
				return nil
			})

			Expect(writer).ToNot(BeNil())
			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closed).To(BeTrue())
		})

		It("should create writer to discard when writer is nil", func() {
			writer := NewWriter(nil, nil)
			Expect(writer).ToNot(BeNil())

			// Should be able to write without error (to io.Discard)
			n, err := writer.WriteString("test data")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(9))

			// Close should work
			err = writer.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// Write operations tests verify delegation and buffering behavior.
	Context("Write operations", func() {
		It("should write data", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			n, err := writer.Write([]byte("test data"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(9))

			// Need to close to flush
			writer.Close()
			Expect(dest.String()).To(Equal("test data"))
		})

		It("should write string", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			n, err := writer.WriteString("hello world")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(11))

			writer.Close()
			Expect(dest.String()).To(Equal("hello world"))
		})

		It("should write multiple times", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			n1, err1 := writer.WriteString("hello")
			n2, err2 := writer.WriteString(" ")
			n3, err3 := writer.WriteString("world")

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(err3).ToNot(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(n2).To(Equal(1))
			Expect(n3).To(Equal(5))

			writer.Close()
			Expect(dest.String()).To(Equal("hello world"))
		})

		It("should read from reader", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			source := strings.NewReader("data from reader")
			n, err := writer.ReadFrom(source)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(16)))

			writer.Close()
			Expect(dest.String()).To(Equal("data from reader"))
		})
	})

	// Close operations tests verify flush, reset, and custom function execution.
	Context("Close operations", func() {
		It("should flush and close writer", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			writer.WriteString("test")

			// Before close, data is buffered
			Expect(dest.Len()).To(Equal(0))

			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())

			// After close, data is flushed
			Expect(dest.String()).To(Equal("test"))
		})

		It("should call custom close function", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			closeCalled := false

			writer := NewWriter(bw, func() error {
				closeCalled = true
				return nil
			})

			err := writer.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closeCalled).To(BeTrue())
		})

		It("should return close function error", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			expectedErr := errors.New("close error")

			writer := NewWriter(bw, func() error {
				return expectedErr
			})

			err := writer.Close()
			Expect(err).To(Equal(expectedErr))
		})

		It("should be safe to close multiple times", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			err1 := writer.Close()
			err2 := writer.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	// Edge cases tests verify behavior with large writes and flush errors.
	Context("Edge cases", func() {
		It("should handle empty write", func() {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			n, err := writer.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))

			writer.Close()
			Expect(dest.Len()).To(Equal(0))
		})

		It("should handle large data", func() {
			largeData := make([]byte, 1024*1024) // 1 MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			writer := NewWriter(bw, nil)

			n, err := writer.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))

			writer.Close()
			Expect(dest.Len()).To(Equal(len(largeData)))
		})
	})
})
