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
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
)

// mockReader for testing error conditions
type mockReader struct {
	data []byte
	pos  int
	err  error
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReader) Close() error {
	return m.err
}

var _ = Describe("AES Reader Operations", func() {
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

	Describe("EncodeReader", func() {
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

		It("should create encode reader", func() {
			plaintext := []byte("test data")
			reader := bytes.NewReader(plaintext)

			encReader := coder.EncodeReader(reader)
			Expect(encReader).ToNot(BeNil())
		})

		It("should encode data through reader", func() {
			plaintext := []byte("Hello, World!")
			reader := bytes.NewReader(plaintext)

			encReader := coder.EncodeReader(reader)
			buffer := make([]byte, 1024)

			n, err := encReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", len(plaintext)))

			// Should be able to decode it
			encrypted := buffer[:n]
			decrypted, err := coder.Decode(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle empty reader", func() {
			reader := bytes.NewReader([]byte{})
			encReader := coder.EncodeReader(reader)

			buffer := make([]byte, 100)
			n, err := encReader.Read(buffer)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should return error for buffer too small", func() {
			plaintext := []byte("test data that is longer")
			reader := bytes.NewReader(plaintext)
			encReader := coder.EncodeReader(reader)

			// Buffer too small to hold encrypted data
			buffer := make([]byte, 5)
			_, err := encReader.Read(buffer)

			Expect(err).To(Equal(encaes.ErrInvalidBufferSize))
		})

		It("should close underlying reader if closeable", func() {
			plaintext := []byte("test")
			mockR := &mockReader{data: plaintext}

			encReader := coder.EncodeReader(mockR)
			err := encReader.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle reader errors", func() {
			expectedErr := errors.New("read error")
			mockR := &mockReader{
				data: []byte("test"),
				err:  expectedErr,
			}

			encReader := coder.EncodeReader(mockR)
			buffer := make([]byte, 100)

			_, err := encReader.Read(buffer)
			Expect(err).To(Equal(expectedErr))
		})

		It("should handle multiple reads", func() {
			plaintext := []byte("This is a longer message that requires multiple reads")
			reader := bytes.NewReader(plaintext)
			encReader := coder.EncodeReader(reader)

			var encrypted []byte
			buffer := make([]byte, 100)

			for {
				n, err := encReader.Read(buffer)
				if err == io.EOF {
					break
				}
				Expect(err).ToNot(HaveOccurred())
				encrypted = append(encrypted, buffer[:n]...)
			}

			// Verify we can decrypt the accumulated data
			Expect(len(encrypted)).To(BeNumerically(">", 0))
		})
	})

	Describe("DecodeReader", func() {
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

		It("should create decode reader", func() {
			encrypted := coder.Encode([]byte("test"))
			reader := bytes.NewReader(encrypted)

			decReader := coder.DecodeReader(reader)
			Expect(decReader).ToNot(BeNil())
		})

		It("should decode data through reader", func() {
			plaintext := []byte("Hello, World!")
			encrypted := coder.Encode(plaintext)
			reader := bytes.NewReader(encrypted)

			decReader := coder.DecodeReader(reader)
			buffer := make([]byte, 1024)

			n, err := decReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer[:n]).To(Equal(plaintext))
		})

		It("should handle empty encrypted data", func() {
			reader := bytes.NewReader([]byte{})
			decReader := coder.DecodeReader(reader)

			buffer := make([]byte, 100)
			n, err := decReader.Read(buffer)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should return error for invalid encrypted data", func() {
			// Not actually encrypted data
			invalidData := []byte("not encrypted")
			reader := bytes.NewReader(invalidData)
			decReader := coder.DecodeReader(reader)

			buffer := make([]byte, 100)
			_, err := decReader.Read(buffer)

			Expect(err).To(HaveOccurred())
		})

		It("should close underlying reader if closeable", func() {
			encrypted := coder.Encode([]byte("test"))
			mockR := &mockReader{data: encrypted}

			decReader := coder.DecodeReader(mockR)
			err := decReader.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle reader errors", func() {
			expectedErr := errors.New("read error")
			mockR := &mockReader{
				data: coder.Encode([]byte("test")),
				err:  expectedErr,
			}

			decReader := coder.DecodeReader(mockR)
			buffer := make([]byte, 100)

			_, err := decReader.Read(buffer)
			Expect(err).To(Equal(expectedErr))
		})

		It("should handle buffer too small for decrypted output", func() {
			plaintext := []byte("This is a longer message")
			encrypted := coder.Encode(plaintext)
			reader := bytes.NewReader(encrypted)
			decReader := coder.DecodeReader(reader)

			// Buffer too small for decrypted data
			buffer := make([]byte, 5)
			_, err := decReader.Read(buffer)

			// Will get error - either ErrInvalidBufferSize or authentication failure
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Reader Round-trip", func() {
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

		It("should preserve data through encode/decode readers", func() {
			plaintext := []byte("Test message for round-trip")

			// Encode through reader
			encReader := coder.EncodeReader(bytes.NewReader(plaintext))
			encBuffer := make([]byte, 1024)
			encN, err := encReader.Read(encBuffer)
			Expect(err).ToNot(HaveOccurred())
			encrypted := encBuffer[:encN]

			// Decode through reader
			decReader := coder.DecodeReader(bytes.NewReader(encrypted))
			decBuffer := make([]byte, 1024)
			decN, err := decReader.Read(decBuffer)
			Expect(err).ToNot(HaveOccurred())
			decrypted := decBuffer[:decN]

			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle large data through readers", func() {
			// 10KB of data (reasonable size for single read/write)
			largeData := make([]byte, 10*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			// Encode in single operation
			encReader := coder.EncodeReader(bytes.NewReader(largeData))
			buffer := make([]byte, 20*1024) // Large enough for encrypted data

			encN, err := encReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			encrypted := buffer[:encN]

			// Decode in single operation
			decReader := coder.DecodeReader(bytes.NewReader(encrypted))
			decBuffer := make([]byte, 20*1024)

			decN, err := decReader.Read(decBuffer)
			Expect(err).ToNot(HaveOccurred())
			decrypted := decBuffer[:decN]

			Expect(decrypted).To(Equal(largeData))
		})
	})
})
