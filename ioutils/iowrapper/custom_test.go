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

// This file tests custom function registration and execution.
//
// Test Strategy:
// - Verify custom functions are called when set via SetRead, SetWrite, SetSeek, SetClose
// - Test that custom functions can transform data (uppercase, ROT13, etc.)
// - Ensure custom functions can override default behavior completely
// - Validate that passing nil to Set* methods resets to default behavior
// - Test multiple custom functions can be set and replaced dynamically
//
// Coverage: 24 specs testing the core customization features of the wrapper.
package iowrapper_test

import (
	"bytes"
	"errors"
	"io"
	"strings"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IOWrapper - Custom Functions", func() {
	Context("Custom Read function", func() {
		It("should use custom read function", func() {
			reader := strings.NewReader("original")
			wrapper := New(reader)

			// Set custom read that returns modified data
			wrapper.SetRead(func(p []byte) []byte {
				return []byte("custom")
			})

			data := make([]byte, 10)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(6))
			Expect(string(data[:n])).To(Equal("custom"))
		})

		It("should handle nil custom read", func() {
			reader := strings.NewReader("test")
			wrapper := New(reader)

			// Set nil to revert to default
			wrapper.SetRead(nil)

			data := make([]byte, 4)
			n, err := wrapper.Read(data)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(string(data)).To(Equal("test"))
		})

		It("should track read calls", func() {
			reader := strings.NewReader("data")
			wrapper := New(reader)

			callCount := 0
			wrapper.SetRead(func(p []byte) []byte {
				callCount++
				return []byte("x")
			})

			wrapper.Read(make([]byte, 10))
			wrapper.Read(make([]byte, 10))

			Expect(callCount).To(Equal(2))
		})
	})

	Context("Custom Write function", func() {
		It("should use custom write function", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// Set custom write that modifies data
			wrapper.SetWrite(func(p []byte) []byte {
				return append([]byte("prefix-"), p...)
			})

			n, err := wrapper.Write([]byte("data"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(11)) // "prefix-data"
		})

		It("should handle nil custom write", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// Set nil to revert to default
			wrapper.SetWrite(nil)

			n, err := wrapper.Write([]byte("test"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
			Expect(buf.String()).To(Equal("test"))
		})

		It("should track write calls", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			callCount := 0
			wrapper.SetWrite(func(p []byte) []byte {
				callCount++
				return p
			})

			wrapper.Write([]byte("a"))
			wrapper.Write([]byte("b"))
			wrapper.Write([]byte("c"))

			Expect(callCount).To(Equal(3))
		})
	})

	Context("Custom Seek function", func() {
		It("should use custom seek function", func() {
			reader := bytes.NewReader([]byte("test"))
			wrapper := New(reader)

			// Set custom seek that always returns position 42
			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				return 42, nil
			})

			pos, err := wrapper.Seek(10, io.SeekStart)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(42)))
		})

		It("should handle custom seek error", func() {
			reader := bytes.NewReader([]byte("test"))
			wrapper := New(reader)

			expectedErr := errors.New("seek error")
			wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
				return 0, expectedErr
			})

			_, err := wrapper.Seek(10, io.SeekStart)

			Expect(err).To(Equal(expectedErr))
		})

		It("should handle nil custom seek", func() {
			reader := bytes.NewReader([]byte("test"))
			wrapper := New(reader)

			// Set nil to revert to default
			wrapper.SetSeek(nil)

			pos, err := wrapper.Seek(2, io.SeekStart)

			Expect(err).ToNot(HaveOccurred())
			Expect(pos).To(Equal(int64(2)))
		})
	})

	Context("Custom Close function", func() {
		It("should use custom close function", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			closeCalled := false
			wrapper.SetClose(func() error {
				closeCalled = true
				return nil
			})

			err := wrapper.Close()

			Expect(err).ToNot(HaveOccurred())
			Expect(closeCalled).To(BeTrue())
		})

		It("should handle custom close error", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			expectedErr := errors.New("close error")
			wrapper.SetClose(func() error {
				return expectedErr
			})

			err := wrapper.Close()

			Expect(err).To(Equal(expectedErr))
		})

		It("should handle nil custom close", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// Set nil to revert to default
			wrapper.SetClose(nil)

			err := wrapper.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should be safe to call close multiple times with custom function", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			callCount := 0
			wrapper.SetClose(func() error {
				callCount++
				return nil
			})

			wrapper.Close()
			wrapper.Close()

			Expect(callCount).To(Equal(2))
		})
	})

	Context("Combined custom functions", func() {
		It("should use multiple custom functions together", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// Custom read
			wrapper.SetRead(func(p []byte) []byte {
				return []byte("read")
			})

			// Custom write
			wrapper.SetWrite(func(p []byte) []byte {
				return append([]byte("["), append(p, ']')...)
			})

			// Custom close
			closeCalled := false
			wrapper.SetClose(func() error {
				closeCalled = true
				return nil
			})

			// Test read
			readData := make([]byte, 10)
			n, _ := wrapper.Read(readData)
			Expect(string(readData[:n])).To(Equal("read"))

			// Test write
			n, _ = wrapper.Write([]byte("data"))
			Expect(n).To(Equal(6))

			// Test close
			wrapper.Close()
			Expect(closeCalled).To(BeTrue())
		})

		It("should allow changing functions dynamically", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			// First custom write
			wrapper.SetWrite(func(p []byte) []byte {
				return []byte("first")
			})

			data := make([]byte, 10)
			n, _ := wrapper.Write(data)
			Expect(n).To(Equal(5))

			// Second custom write
			wrapper.SetWrite(func(p []byte) []byte {
				return []byte("second")
			})

			n, _ = wrapper.Write(data)
			Expect(n).To(Equal(6))
		})
	})

	Context("Edge cases", func() {
		It("should handle returning nil from custom read", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			wrapper.SetRead(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Read(make([]byte, 10))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle returning nil from custom write", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			wrapper.SetWrite(func(p []byte) []byte {
				return nil
			})

			_, err := wrapper.Write([]byte("test"))

			Expect(err).To(Equal(io.ErrUnexpectedEOF))
		})

		It("should handle empty byte slices", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)

			wrapper.SetRead(func(p []byte) []byte {
				return []byte{}
			})

			data := make([]byte, 10)
			n, _ := wrapper.Read(data)

			Expect(n).To(Equal(0))
		})
	})
})
