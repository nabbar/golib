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

// ReadWriter tests verify the bufio.ReadWriter wrapper with io.Closer support.
// Tests cover creation, bidirectional I/O, flush on close (no reset due to
// ambiguous methods), custom close functions, and nil parameter handling.
var _ = Describe("ReadWriter", func() {
	// Creation tests verify readwriter instantiation and nil handling.
	Context("Creation", func() {
		It("should create readwriter from bufio.ReadWriter", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			Expect(readwriter).ToNot(BeNil())
		})

		It("should create readwriter with custom close function", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			closed := false

			readwriter := NewReadWriter(brw, func() error {
				closed = true
				return nil
			})

			Expect(readwriter).ToNot(BeNil())
			err := readwriter.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closed).To(BeTrue())
		})

		It("should create readwriter with defaults when readwriter is nil", func() {
			rw := NewReadWriter(nil, nil)
			Expect(rw).ToNot(BeNil())

			// Should be able to write (to io.Discard)
			n, err := rw.WriteString("test")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))

			// Should return EOF on read (empty source)
			data := make([]byte, 10)
			n, err = rw.Read(data)
			Expect(err).To(HaveOccurred()) // EOF
			Expect(n).To(Equal(0))

			// Close should work
			err = rw.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// Read operations tests verify read delegation in bidirectional context.
	Context("Read operations", func() {
		It("should read data", func() {
			buf := bytes.NewBufferString("hello world")
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			data := make([]byte, 5)
			n, err := readwriter.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data)).To(Equal("hello"))
		})

		It("should write to writer", func() {
			buf := bytes.NewBufferString("source data")
			dest := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(dest))
			readwriter := NewReadWriter(brw, nil)

			n, err := readwriter.WriteTo(dest)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", 0))
		})

		It("should read from reader", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			source := strings.NewReader("data from reader")
			n, err := readwriter.ReadFrom(source)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(16)))
		})
	})

	// Write operations tests verify write delegation and buffering.
	Context("Write operations", func() {
		It("should write data", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			n, err := readwriter.Write([]byte("test data"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(9))

			// Close to flush
			readwriter.Close()
			Expect(buf.String()).To(ContainSubstring("test data"))
		})

		It("should write string", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			n, err := readwriter.WriteString("hello")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))

			readwriter.Close()
			Expect(buf.String()).To(ContainSubstring("hello"))
		})
	})

	Context("Combined read/write operations", func() {
		It("should support both read and write", func() {
			// Create a buffer with initial data
			buf := bytes.NewBufferString("initial data\n")
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			// Read some data
			data := make([]byte, 7)
			n, err := readwriter.Read(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data[:n])).To(Equal("initial"))

			// Write some data
			n, err = readwriter.WriteString(" more")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))

			readwriter.Close()
		})

		It("should handle alternating read and write", func() {
			buf := bytes.NewBufferString("read1\nread2\n")
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			// Read
			data1 := make([]byte, 6)
			n1, _ := readwriter.Read(data1)
			Expect(string(data1[:n1])).To(Equal("read1\n"))

			// Write
			readwriter.WriteString("write1\n")

			// Read
			data2 := make([]byte, 6)
			n2, _ := readwriter.Read(data2)
			Expect(string(data2[:n2])).To(Equal("read2\n"))

			readwriter.Close()
		})
	})

	// Close operations tests verify flush (no reset) and custom function execution.
	Context("Close operations", func() {
		It("should flush and close readwriter", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			readwriter.WriteString("test")

			err := readwriter.Close()
			Expect(err).ToNot(HaveOccurred())

			// Data should be flushed after close
			Expect(buf.String()).To(Equal("test"))
		})

		It("should call custom close function", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			closeCalled := false

			readwriter := NewReadWriter(brw, func() error {
				closeCalled = true
				return nil
			})

			err := readwriter.Close()
			Expect(err).ToNot(HaveOccurred())
			Expect(closeCalled).To(BeTrue())
		})

		It("should return close function error", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			expectedErr := errors.New("close error")

			readwriter := NewReadWriter(brw, func() error {
				return expectedErr
			})

			err := readwriter.Close()
			Expect(err).To(Equal(expectedErr))
		})

		It("should be safe to close multiple times", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			err1 := readwriter.Close()
			err2 := readwriter.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	// Edge cases tests verify combined read/write and error handling.
	Context("Edge cases", func() {
		It("should handle empty buffer", func() {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			data := make([]byte, 10)
			n, err := readwriter.Read(data)

			Expect(err).To(HaveOccurred()) // EOF
			Expect(n).To(Equal(0))
		})

		It("should handle large data", func() {
			largeData := make([]byte, 1024*1024) // 1 MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			readwriter := NewReadWriter(brw, nil)

			n, err := readwriter.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))

			readwriter.Close()
		})
	})
})
