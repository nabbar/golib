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

var _ = Describe("SHA-256 Writer Operations", func() {
	Describe("EncodeWriter", func() {
		It("should create a writer wrapper", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)
			Expect(writer).ToNot(BeNil())
		})

		It("should pass through data while hashing", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			input := []byte("Hello, World!")

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write(input)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(input)))
			Expect(dest.Bytes()).To(Equal(input))

			// Verify hash was computed
			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle empty write", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write([]byte{})

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(dest.Len()).To(Equal(0))
		})

		It("should handle nil write", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write(nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle large data", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			largeData := make([]byte, 1024*1024) // 1MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write(largeData)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(dest.Bytes()).To(Equal(largeData))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(largeData)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle multiple writes", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)

			// First write
			data1 := []byte("First ")
			n1, err1 := writer.Write(data1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(n1).To(Equal(len(data1)))

			// Second write
			data2 := []byte("Second ")
			n2, err2 := writer.Write(data2)
			Expect(err2).ToNot(HaveOccurred())
			Expect(n2).To(Equal(len(data2)))

			// Third write
			data3 := []byte("Third")
			n3, err3 := writer.Write(data3)
			Expect(err3).ToNot(HaveOccurred())
			Expect(n3).To(Equal(len(data3)))

			// Verify all data written
			expected := append(data1, data2...)
			expected = append(expected, data3...)
			Expect(dest.Bytes()).To(Equal(expected))

			// Verify hash
			hash := hasher.Encode(nil)
			expectedHash := sha256.Sum256(expected)
			Expect(hash).To(Equal(expectedHash[:]))
		})

		It("should close underlying writer if closeable", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			closeable := &closeableBuffer{Buffer: dest}

			writer := hasher.EncodeWriter(closeable)
			err := writer.Close()

			Expect(err).ToNot(HaveOccurred())
			Expect(closeable.closed).To(BeTrue())
		})

		It("should handle non-closeable writer", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)
			err := writer.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle binary data", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			binary := []byte{0x00, 0xFF, 0x7F, 0x80, 0xDE, 0xAD, 0xBE, 0xEF}

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write(binary)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(binary)))
			Expect(dest.Bytes()).To(Equal(binary))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(binary)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle UTF-8 text", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			utf8 := []byte("Hello 世界 🔒")

			writer := hasher.EncodeWriter(dest)
			n, err := writer.Write(utf8)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(utf8)))
			Expect(dest.Bytes()).To(Equal(utf8))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(utf8)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should work with io.Copy", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			input := []byte("Data to copy")
			source := bytes.NewReader(input)

			writer := hasher.EncodeWriter(dest)
			n, err := io.Copy(writer, source)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(int64(len(input))))
			Expect(dest.Bytes()).To(Equal(input))

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})
	})

	Describe("DecodeWriter", func() {
		It("should return nil (no decode for hash)", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.DecodeWriter(dest)
			Expect(writer).To(BeNil())
		})
	})

	Describe("Writer Edge Cases", func() {
		It("should handle small incremental writes", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			input := []byte("test data")

			writer := hasher.EncodeWriter(dest)

			// Write one byte at a time
			for i, b := range input {
				n, err := writer.Write([]byte{b})
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(1))
				Expect(dest.Bytes()[i]).To(Equal(b))
			}

			hash := hasher.Encode(nil)
			expected := sha256.Sum256(input)
			Expect(hash).To(Equal(expected[:]))
		})

		It("should handle alternating write sizes", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)

			// Small write
			small := []byte("ab")
			writer.Write(small)

			// Large write
			large := make([]byte, 1000)
			for i := range large {
				large[i] = byte(i % 256)
			}
			writer.Write(large)

			// Medium write
			medium := []byte("medium data here")
			writer.Write(medium)

			// Verify all data
			expected := append(small, large...)
			expected = append(expected, medium...)
			Expect(dest.Bytes()).To(Equal(expected))

			hash := hasher.Encode(nil)
			expectedHash := sha256.Sum256(expected)
			Expect(hash).To(Equal(expectedHash[:]))
		})

		It("should handle write after close", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}

			writer := hasher.EncodeWriter(dest)
			writer.Close()

			// Write after close should still work (data written, just closed)
			n, err := writer.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should handle multiple close calls", func() {
			hasher := encsha.New()
			dest := &bytes.Buffer{}
			closeable := &closeableBuffer{Buffer: dest}

			writer := hasher.EncodeWriter(closeable)

			err1 := writer.Close()
			Expect(err1).ToNot(HaveOccurred())

			err2 := writer.Close()
			Expect(err2).ToNot(HaveOccurred())
		})
	})
})

// closeableBuffer is a buffer that implements io.WriteCloser for testing
type closeableBuffer struct {
	*bytes.Buffer
	closed bool
}

func (c *closeableBuffer) Close() error {
	c.closed = true
	return nil
}
