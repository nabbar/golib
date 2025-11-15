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

package hexa_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	enchex "github.com/nabbar/golib/encoding/hexa"
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

var _ = Describe("Hexadecimal Reader Operations", func() {
	Describe("EncodeReader", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
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
			plaintext := []byte("Hello!")
			reader := bytes.NewReader(plaintext)

			encReader := coder.EncodeReader(reader)
			buffer := make([]byte, 100)

			n, err := encReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(plaintext) * 2))

			// Verify hex encoding
			hexEncoded := buffer[:n]
			decoded, err := coder.Decode(hexEncoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(plaintext))
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

			// Buffer too small (less than 2 bytes)
			buffer := make([]byte, 1)
			_, err := encReader.Read(buffer)

			Expect(err).To(Equal(enchex.ErrInvalidBufferSize))
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

			var encoded []byte
			buffer := make([]byte, 20) // Small buffer to force multiple reads

			for {
				n, err := encReader.Read(buffer)
				if err == io.EOF {
					break
				}
				if err == enchex.ErrInvalidBufferSize {
					// Buffer too small, increase it
					buffer = make([]byte, len(buffer)*2)
					continue
				}
				Expect(err).ToNot(HaveOccurred())
				encoded = append(encoded, buffer[:n]...)
			}

			// Verify we got something
			Expect(len(encoded)).To(BeNumerically(">", 0))
		})
	})

	Describe("DecodeReader", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should create decode reader", func() {
			hexData := []byte("48656c6c6f")
			reader := bytes.NewReader(hexData)

			decReader := coder.DecodeReader(reader)
			Expect(decReader).ToNot(BeNil())
		})

		It("should decode data through reader", func() {
			plaintext := []byte("Hello!")
			hexData := coder.Encode(plaintext)
			reader := bytes.NewReader(hexData)

			decReader := coder.DecodeReader(reader)
			buffer := make([]byte, 100)

			n, err := decReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer[:n]).To(Equal(plaintext))
		})

		It("should handle empty hex data", func() {
			reader := bytes.NewReader([]byte{})
			decReader := coder.DecodeReader(reader)

			buffer := make([]byte, 100)
			n, err := decReader.Read(buffer)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should return error for invalid hex data", func() {
			// Not valid hex
			invalidData := []byte("not valid hex!!!")
			reader := bytes.NewReader(invalidData)
			decReader := coder.DecodeReader(reader)

			buffer := make([]byte, 100)
			_, err := decReader.Read(buffer)

			Expect(err).To(HaveOccurred())
		})

		It("should close underlying reader if closeable", func() {
			hexData := coder.Encode([]byte("test"))
			mockR := &mockReader{data: hexData}

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
	})

	Describe("Reader Round-trip", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
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
			encBuffer := make([]byte, len(plaintext)*3) // Large enough
			encN, err := encReader.Read(encBuffer)
			Expect(err).ToNot(HaveOccurred())
			hexEncoded := encBuffer[:encN]

			// Decode through reader
			decReader := coder.DecodeReader(bytes.NewReader(hexEncoded))
			decBuffer := make([]byte, len(plaintext)*2)
			decN, err := decReader.Read(decBuffer)
			Expect(err).ToNot(HaveOccurred())
			decrypted := decBuffer[:decN]

			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle large data through readers", func() {
			// 1KB of data (more reasonable for single Read operations)
			largeData := make([]byte, 1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			// Encode in single operation
			encReader := coder.EncodeReader(bytes.NewReader(largeData))
			buffer := make([]byte, len(largeData)*3) // Large enough for hex

			encN, err := encReader.Read(buffer)
			Expect(err).ToNot(HaveOccurred())
			hexEncoded := buffer[:encN]

			// Decode through reader with io.ReadAll for complete data
			decReader := coder.DecodeReader(bytes.NewReader(hexEncoded))
			decrypted, err := io.ReadAll(decReader)
			Expect(err).ToNot(HaveOccurred())

			Expect(decrypted).To(Equal(largeData))
		})
	})
})
