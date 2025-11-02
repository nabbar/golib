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

package aes_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
)

// mockWriter for testing error conditions
type mockWriter struct {
	buffer bytes.Buffer
	err    error
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.buffer.Write(p)
}

func (m *mockWriter) Close() error {
	return m.err
}

var _ = Describe("AES Writer Operations", func() {
	var (
		key   [32]byte
		nonce [12]byte
	)

	BeforeEach(func() {
		var err error
		key, err = encaes.GenKey()
		Expect(err).ToNot(HaveOccurred())

		nonce, err = encaes.GenNonce()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("EncodeWriter", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			var err error
			coder, err = encaes.New(key, nonce)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should create encode writer", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)
			Expect(encWriter).ToNot(BeNil())
		})

		It("should encode data through writer", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			plaintext := []byte("Hello, World!")
			n, err := encWriter.Write(plaintext)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(plaintext)))
			Expect(buffer.Len()).To(BeNumerically(">", len(plaintext)))

			// Verify we can decrypt what was written
			encrypted := buffer.Bytes()
			decrypted, err := coder.Decode(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle empty write", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			n, err := encWriter.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle nil write", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			n, err := encWriter.Write(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should close underlying writer if closeable", func() {
			mockW := &mockWriter{}
			encWriter := coder.EncodeWriter(mockW)

			err := encWriter.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should propagate write errors", func() {
			expectedErr := errors.New("write error")
			mockW := &mockWriter{err: expectedErr}
			encWriter := coder.EncodeWriter(mockW)

			_, err := encWriter.Write([]byte("test"))
			Expect(err).To(Equal(expectedErr))
		})

		It("should handle multiple writes", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			messages := [][]byte{
				[]byte("First message"),
				[]byte("Second message"),
				[]byte("Third message"),
			}

			for _, msg := range messages {
				n, err := encWriter.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))
			}

			Expect(buffer.Len()).To(BeNumerically(">", 0))
		})

		It("should handle large data", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			// Write 1MB of data
			largeData := make([]byte, 1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			n, err := encWriter.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(buffer.Len()).To(BeNumerically(">", len(largeData)))
		})

		It("should return reported bytes written as input size", func() {
			buffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(buffer)

			plaintext := []byte("test data")
			n, err := encWriter.Write(plaintext)

			// Should report input size, not encrypted size
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(plaintext)))
		})
	})

	Describe("DecodeWriter", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			var err error
			coder, err = encaes.New(key, nonce)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should create decode writer", func() {
			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)
			Expect(decWriter).ToNot(BeNil())
		})

		It("should decode data through writer", func() {
			plaintext := []byte("Hello, World!")
			encrypted := coder.Encode(plaintext)

			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			n, err := decWriter.Write(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(encrypted)))
			Expect(buffer.Bytes()).To(Equal(plaintext))
		})

		It("should handle empty write", func() {
			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			n, err := decWriter.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle nil write", func() {
			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			n, err := decWriter.Write(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should return error for invalid encrypted data", func() {
			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			invalidData := []byte("not encrypted")
			_, err := decWriter.Write(invalidData)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for corrupted data", func() {
			plaintext := []byte("test")
			encrypted := coder.Encode(plaintext)
			encrypted[0] ^= 0xFF // Corrupt it

			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			_, err := decWriter.Write(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should close underlying writer if closeable", func() {
			mockW := &mockWriter{}
			decWriter := coder.DecodeWriter(mockW)

			err := decWriter.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should propagate write errors", func() {
			plaintext := []byte("test")
			encrypted := coder.Encode(plaintext)

			expectedErr := errors.New("write error")
			mockW := &mockWriter{err: expectedErr}
			decWriter := coder.DecodeWriter(mockW)

			_, err := decWriter.Write(encrypted)
			Expect(err).To(Equal(expectedErr))
		})

		It("should return reported bytes written as input size", func() {
			plaintext := []byte("test data")
			encrypted := coder.Encode(plaintext)

			buffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(buffer)

			n, err := decWriter.Write(encrypted)

			// Should report input size, not decrypted size
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(encrypted)))
		})
	})

	Describe("Writer Round-trip", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			var err error
			coder, err = encaes.New(key, nonce)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should preserve data through encode/decode writers", func() {
			plaintext := []byte("Test message for round-trip")

			// Encode through writer
			encBuffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(encBuffer)
			n, err := encWriter.Write(plaintext)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(plaintext)))

			encrypted := encBuffer.Bytes()

			// Decode through writer
			decBuffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(decBuffer)
			n, err = decWriter.Write(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(encrypted)))

			decrypted := decBuffer.Bytes()
			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle large data through writers", func() {
			// 100KB of data
			largeData := make([]byte, 100*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			// Encode
			encBuffer := &bytes.Buffer{}
			encWriter := coder.EncodeWriter(encBuffer)
			_, err := encWriter.Write(largeData)
			Expect(err).ToNot(HaveOccurred())

			encrypted := encBuffer.Bytes()

			// Decode
			decBuffer := &bytes.Buffer{}
			decWriter := coder.DecodeWriter(decBuffer)
			_, err = decWriter.Write(encrypted)
			Expect(err).ToNot(HaveOccurred())

			decrypted := decBuffer.Bytes()
			Expect(decrypted).To(Equal(largeData))
		})

		It("should handle multiple write operations", func() {
			messages := [][]byte{
				[]byte("First"),
				[]byte("Second"),
				[]byte("Third"),
			}

			// Encode all messages
			var encrypted [][]byte
			for _, msg := range messages {
				buf := &bytes.Buffer{}
				w := coder.EncodeWriter(buf)
				_, err := w.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				encrypted = append(encrypted, buf.Bytes())
			}

			// Decode all messages
			for i, enc := range encrypted {
				buf := &bytes.Buffer{}
				w := coder.DecodeWriter(buf)
				_, err := w.Write(enc)
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.Bytes()).To(Equal(messages[i]))
			}
		})
	})
})
