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

// Reader tests verify the bufio.Reader wrapper with io.Closer support.
// Tests cover creation, read operations, WriteTo, close behavior with reset,
// custom close functions, and nil parameter handling.
var _ = Describe("Reader", func() {
	// Creation tests verify reader instantiation and nil handling.
	Context("Creation", func() {
		It("should create reader from bufio.Reader", func() {
			source := strings.NewReader("test data")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			Expect(reader).ToNot(BeNil())
		})

		It("should create reader with custom close function", func() {
			source := strings.NewReader("test")
			br := bufio.NewReader(source)
			closed := false

			reader := NewReader(br, func() error {
				closed = true
				return nil
			})

			Expect(reader).ToNot(BeNil())
			err := reader.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closed).To(BeTrue())
		})

		It("should create reader from empty source when reader is nil", func() {
			reader := NewReader(nil, nil)
			Expect(reader).ToNot(BeNil())

			// Should return EOF immediately
			data := make([]byte, 10)
			n, err := reader.Read(data)
			Expect(err).To(HaveOccurred()) // EOF
			Expect(n).To(Equal(0))
		})
	})

	// Read operations tests verify delegation to underlying bufio.Reader.
	Context("Read operations", func() {
		It("should read data", func() {
			source := strings.NewReader("hello world")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			data := make([]byte, 5)
			n, err := reader.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data)).To(Equal("hello"))
		})

		It("should read multiple times", func() {
			source := strings.NewReader("abcdefghij")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			// First read
			data1 := make([]byte, 3)
			n1, err1 := reader.Read(data1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(3))
			Expect(string(data1)).To(Equal("abc"))

			// Second read
			data2 := make([]byte, 3)
			n2, err2 := reader.Read(data2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(3))
			Expect(string(data2)).To(Equal("def"))
		})

		It("should write to writer", func() {
			source := strings.NewReader("source data")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			dest := &bytes.Buffer{}
			n, err := reader.WriteTo(dest)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(11)))
			Expect(dest.String()).To(Equal("source data"))
		})

		It("should handle EOF", func() {
			source := strings.NewReader("short")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			// Read all data
			data := make([]byte, 10)
			n, err := reader.Read(data)

			Expect(n).To(Equal(5))
			// bufio.Reader may or may not return EOF on first read
			if err != nil {
				Expect(err).To(MatchError("EOF"))
			}
		})
	})

	// Close operations tests verify reset and custom function execution.
	Context("Close operations", func() {
		It("should close and reset reader", func() {
			source := strings.NewReader("data")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			err := reader.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call custom close function", func() {
			source := strings.NewReader("data")
			br := bufio.NewReader(source)
			closeCalled := false

			reader := NewReader(br, func() error {
				closeCalled = true
				return nil
			})

			err := reader.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closeCalled).To(BeTrue())
		})

		It("should return close function error", func() {
			source := strings.NewReader("data")
			br := bufio.NewReader(source)
			expectedErr := errors.New("close error")

			reader := NewReader(br, func() error {
				return expectedErr
			})

			err := reader.Close()
			Expect(err).To(Equal(expectedErr))
		})

		It("should be safe to close multiple times", func() {
			source := strings.NewReader("data")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			err1 := reader.Close()
			err2 := reader.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	// Edge cases tests verify behavior with empty sources and EOF.
	Context("Edge cases", func() {
		It("should handle empty reader", func() {
			source := strings.NewReader("")
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			data := make([]byte, 10)
			n, err := reader.Read(data)

			Expect(err).To(HaveOccurred()) // EOF
			Expect(n).To(Equal(0))
		})

		It("should handle large data", func() {
			largeData := make([]byte, 1024*1024) // 1 MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			source := bytes.NewReader(largeData)
			br := bufio.NewReader(source)
			reader := NewReader(br, nil)

			dest := &bytes.Buffer{}
			n, err := reader.WriteTo(dest)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(largeData))))
		})
	})
})
