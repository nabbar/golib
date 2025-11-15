/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

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

package sha256_test

import (
	"bytes"
	"crypto/sha256"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encsha "github.com/nabbar/golib/encoding/sha256"
)

var _ = Describe("SHA-256 Reader Operations", func() {
	Describe("EncodeReader", func() {
		It("should create a reader wrapper", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte("test"))

			reader := hasher.EncodeReader(source)
			Expect(reader).ToNot(BeNil())
		})

		It("should pass through data while hashing", func() {
			hasher := encsha.New()
			input := []byte("Hello, World!")
			source := bytes.NewReader(input)

			reader := hasher.EncodeReader(source)
			output, err := io.ReadAll(reader)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(input))

			// Verify hash was computed
			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle empty reader", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte{})

			reader := hasher.EncodeReader(source)
			output, err := io.ReadAll(reader)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(BeEmpty())

			hash := hasher.Encode(nil)
			expected := sha256.Sum256([]byte{})
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle large data", func() {
			hasher := encsha.New()
			largeData := make([]byte, 1024*1024) // 1MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}
			source := bytes.NewReader(largeData)

			reader := hasher.EncodeReader(source)
			output, err := io.ReadAll(reader)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(largeData))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(largeData)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle chunked reads", func() {
			hasher := encsha.New()
			input := []byte("This is a test of chunked reading")
			source := bytes.NewReader(input)

			reader := hasher.EncodeReader(source)

			// Read in small chunks
			var output []byte
			buf := make([]byte, 5)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					output = append(output, buf[:n]...)
				}
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(output).To(Equal(input))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should close underlying reader if closeable", func() {
			hasher := encsha.New()
			source := io.NopCloser(bytes.NewReader([]byte("test")))

			reader := hasher.EncodeReader(source)
			err := reader.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle non-closeable reader", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte("test"))

			reader := hasher.EncodeReader(source)
			err := reader.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle binary data", func() {
			hasher := encsha.New()
			binary := []byte{0x00, 0xFF, 0x7F, 0x80, 0xDE, 0xAD, 0xBE, 0xEF}
			source := bytes.NewReader(binary)

			reader := hasher.EncodeReader(source)
			output, err := io.ReadAll(reader)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(binary))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(binary)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle UTF-8 text", func() {
			hasher := encsha.New()
			utf8 := []byte("Hello 世界 🔒")
			source := bytes.NewReader(utf8)

			reader := hasher.EncodeReader(source)
			output, err := io.ReadAll(reader)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(utf8))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(utf8)
			Expect(hash).To(Equal(expected[:]))
		})
	})

	Describe("DecodeReader", func() {
		It("should return nil (no decode for hash)", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte("test"))

			reader := hasher.DecodeReader(source)
			Expect(reader).To(BeNil())
		})
	})

	Describe("Reader Edge Cases", func() {
		It("should handle multiple reads", func() {
			hasher := encsha.New()
			input := []byte("test data for multiple reads")
			source := bytes.NewReader(input)

			reader := hasher.EncodeReader(source)

			// First read
			buf1 := make([]byte, 10)
			n1, err1 := reader.Read(buf1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(10))

			// Second read
			buf2 := make([]byte, 10)
			n2, err2 := reader.Read(buf2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(10))

			// Read rest
			rest, err := io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred())

			// Verify complete data
			complete := append(buf1[:n1], buf2[:n2]...)
			complete = append(complete, rest...)
			Expect(complete).To(Equal(input))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle single byte reads", func() {
			hasher := encsha.New()
			input := []byte("test")
			source := bytes.NewReader(input)

			reader := hasher.EncodeReader(source)

			var output []byte
			buf := make([]byte, 1)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					output = append(output, buf[0])
				}
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(output).To(Equal(input))
		})

		It("should handle zero-length reads", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte("test"))

			reader := hasher.EncodeReader(source)

			buf := make([]byte, 0)
			n, err := reader.Read(buf)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle EOF correctly", func() {
			hasher := encsha.New()
			source := bytes.NewReader([]byte("test"))

			reader := hasher.EncodeReader(source)

			// Read all data
			io.ReadAll(reader)

			// Next read should return EOF
			buf := make([]byte, 10)
			n, err := reader.Read(buf)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})
	})
})
