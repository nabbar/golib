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
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encrnd "github.com/nabbar/golib/encoding/randRead"
)

var _ = Describe("Random Reader Operations", func() {
	Describe("New", func() {
		It("should return nil when function is nil", func() {
			rdr := encrnd.New(nil)
			Expect(rdr).To(BeNil())
		})

		It("should return valid reader with valid function", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})
			Expect(rdr).ToNot(BeNil())
		})

		It("should implement io.ReadCloser interface", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})

			var _ io.Reader = rdr
			var _ io.Closer = rdr
			Expect(rdr).ToNot(BeNil())
		})
	})

	Describe("Read Operations", func() {
		It("should read data from source", func() {
			data := []byte("Hello, World!")
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			buf := make([]byte, len(data))
			n, err := rdr.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf).To(Equal(data))
		})

		It("should read data in multiple chunks", func() {
			data := []byte("1234567890")
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			// First read
			buf1 := make([]byte, 5)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(buf1).To(Equal([]byte("12345")))

			// Second read
			buf2 := make([]byte, 5)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(5))
			Expect(buf2).To(Equal([]byte("67890")))
		})

		It("should handle small buffer sizes", func() {
			data := []byte("Test")
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(data)), nil
			})

			buf := make([]byte, 1)
			for i := 0; i < len(data); i++ {
				n, err := rdr.Read(buf)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(1))
				Expect(buf[0]).To(Equal(data[i]))
			}
		})

		It("should read varying data", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				p := make([]byte, 32)
				for i := range p {
					p[i] = byte(callCount * i)
				}
				return io.NopCloser(bytes.NewReader(p)), nil
			})

			buf := make([]byte, 32)
			n, err := rdr.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(32))
			// Data should be unique
			allZero := true
			for _, b := range buf {
				if b != 0 {
					allZero = false
					break
				}
			}
			Expect(allZero).To(BeFalse())
		})

		It("should handle small initial source and refetch", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return io.NopCloser(bytes.NewReader([]byte("ab"))), nil
				}
				return io.NopCloser(bytes.NewReader([]byte("cd"))), nil
			})

			// First read gets first source
			buf1 := make([]byte, 2)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(2))
			Expect(buf1).To(Equal([]byte("ab")))

			// Second read triggers refetch
			buf2 := make([]byte, 2)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(2))
			Expect(buf2).To(Equal([]byte("cd")))
		})

		It("should fetch new data when current source is exhausted", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				return io.NopCloser(bytes.NewReader([]byte{byte(callCount)})), nil
			})

			// First read - should get '1'
			buf1 := make([]byte, 1)
			n1, err1 := rdr.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(1))
			Expect(buf1[0]).To(Equal(byte(1)))

			// Second read - source exhausted, should fetch and get '2'
			buf2 := make([]byte, 1)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(1))
			Expect(buf2[0]).To(Equal(byte(2)))
		})

		It("should handle large data reads", func() {
			largeData := make([]byte, 1024*1024) // 1MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(largeData)), nil
			})

			buf := make([]byte, len(largeData))
			n, err := io.ReadFull(rdr, buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(buf).To(Equal(largeData))
		})
	})

	Describe("Close Operations", func() {
		It("should close without error", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})

			err := rdr.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should close after reading", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})

			buf := make([]byte, 4)
			rdr.Read(buf)

			err := rdr.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should close multiple times without error", func() {
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte("test"))), nil
			})

			err1 := rdr.Close()
			Expect(err1).ToNot(HaveOccurred())

			err2 := rdr.Close()
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	Describe("Sequential Fetches", func() {
		It("should handle multiple fetches in sequence", func() {
			callCount := 0
			rdr := encrnd.New(func() (io.ReadCloser, error) {
				callCount++
				if callCount == 1 {
					return io.NopCloser(bytes.NewReader([]byte("First "))), nil
				}
				return io.NopCloser(bytes.NewReader([]byte("Second"))), nil
			})

			// Read first chunk
			buf := make([]byte, 6)
			n, err := rdr.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(6))
			Expect(string(buf)).To(Equal("First "))

			// Read second chunk (should trigger new fetch)
			buf2 := make([]byte, 6)
			n2, err2 := rdr.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(6))
			Expect(string(buf2)).To(Equal("Second"))
		})
	})
})
