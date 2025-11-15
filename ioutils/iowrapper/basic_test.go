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
	"bytes"
	"io"
	"strings"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IOWrapper - Basic Operations", func() {
	Context("Creation", func() {
		It("should create wrapper from bytes.Buffer", func() {
			buf := bytes.NewBuffer([]byte("test"))
			wrapper := New(buf)

			Expect(wrapper).ToNot(BeNil())
		})

		It("should create wrapper from any type", func() {
			reader := strings.NewReader("data")
			wrapper := New(reader)

			Expect(wrapper).ToNot(BeNil())
		})

		It("should create wrapper from nil", func() {
			wrapper := New(nil)

			Expect(wrapper).ToNot(BeNil())
		})
	})

	Context("Default Read operations", func() {
		It("should read from underlying Reader", func() {
			reader := strings.NewReader("hello world")
			wrapper := New(reader)

			data := make([]byte, 5)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(data)).To(Equal("hello"))
		})

		It("should handle EOF", func() {
			reader := strings.NewReader("hi")
			wrapper := New(reader)

			data := make([]byte, 10)
			n, _ := wrapper.Read(data)

			// Should read 2 bytes
			Expect(n).To(Equal(2))
			Expect(string(data[:n])).To(Equal("hi"))
		})

		It("should handle empty reader", func() {
			reader := strings.NewReader("")
			wrapper := New(reader)

			data := make([]byte, 10)
			n, _ := wrapper.Read(data)

			Expect(n).To(Equal(0))
		})
	})

	Context("Default Write operations", func() {
		It("should write to underlying Writer", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			n, err := wrapper.Write([]byte("hello"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(buf.String()).To(Equal("hello"))
		})

		It("should handle multiple writes", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			wrapper.Write([]byte("hello"))
			wrapper.Write([]byte(" "))
			wrapper.Write([]byte("world"))

			Expect(buf.String()).To(Equal("hello world"))
		})

		It("should handle empty write", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			n, err := wrapper.Write([]byte{})

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("Default Seek operations", func() {
		It("should seek on underlying Seeker", func() {
			data := []byte("hello world")
			reader := bytes.NewReader(data)
			wrapper := New(reader)

			// Seek to position 6
			pos, err := wrapper.Seek(6, io.SeekStart)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(6)))

			// Read from new position
			buf := make([]byte, 5)
			n, _ := wrapper.Read(buf)
			Expect(string(buf[:n])).To(Equal("world"))
		})

		It("should seek relative to current position", func() {
			data := []byte("0123456789")
			reader := bytes.NewReader(data)
			wrapper := New(reader)

			// Read 3 bytes
			wrapper.Read(make([]byte, 3))

			// Seek 2 bytes forward from current
			pos, err := wrapper.Seek(2, io.SeekCurrent)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(5)))
		})

		It("should handle seek on non-seeker", func() {
			reader := strings.NewReader("test")
			wrapper := New(reader)

			_, err := wrapper.Seek(0, io.SeekStart)

			// strings.Reader implements Seeker, so this should work
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Default Close operations", func() {
		It("should close underlying Closer", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			err := wrapper.Close()

			// bytes.Buffer doesn't implement Closer, so should return nil
			Expect(err).ToNot(HaveOccurred())
		})

		It("should be safe to close multiple times", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			err1 := wrapper.Close()
			err2 := wrapper.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})
})
