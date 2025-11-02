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
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
)

// errorReader always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (e *errorReader) Close() error {
	return fmt.Errorf("simulated close error")
}

// errorWriter always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
}

func (e *errorWriter) Close() error {
	return fmt.Errorf("simulated close error")
}

var _ = Describe("AES Edge Cases and Error Handling", func() {
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

	Describe("Error Handling", func() {
		It("should export ErrInvalidBufferSize error", func() {
			Expect(encaes.ErrInvalidBufferSize).ToNot(BeNil())
			Expect(encaes.ErrInvalidBufferSize.Error()).To(ContainSubstring("buffer"))
		})
	})

	Describe("Boundary Conditions", func() {
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

		It("should handle single byte", func() {
			data := []byte{0x42}
			encrypted := coder.Encode(data)
			decrypted, err := coder.Decode(encrypted)

			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(data))
		})

		It("should handle all zero bytes", func() {
			data := make([]byte, 100)
			encrypted := coder.Encode(data)
			decrypted, err := coder.Decode(encrypted)

			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(data))
		})

		It("should handle all 0xFF bytes", func() {
			data := make([]byte, 100)
			for i := range data {
				data[i] = 0xFF
			}
			encrypted := coder.Encode(data)
			decrypted, err := coder.Decode(encrypted)

			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(data))
		})

		It("should handle alternating pattern", func() {
			data := make([]byte, 1000)
			for i := range data {
				data[i] = byte(i % 2 * 255)
			}
			encrypted := coder.Encode(data)
			decrypted, err := coder.Decode(encrypted)

			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(data))
		})

		It("should handle very large data", func() {
			// 10MB of data
			largeData := make([]byte, 10*1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			encrypted := coder.Encode(largeData)
			decrypted, err := coder.Decode(encrypted)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(decrypted)).To(Equal(len(largeData)))
			Expect(decrypted).To(Equal(largeData))
		})
	})

	Describe("Reader Edge Cases", func() {
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

		It("should handle reader with immediate EOF", func() {
			reader := bytes.NewReader([]byte{})
			encReader := coder.EncodeReader(reader)

			buffer := make([]byte, 100)
			_, err := encReader.Read(buffer)
			Expect(err).To(Equal(io.EOF))
		})

		It("should handle reader errors in EncodeReader", func() {
			errReader := &errorReader{}
			encReader := coder.EncodeReader(errReader)

			buffer := make([]byte, 100)
			_, err := encReader.Read(buffer)
			Expect(err).To(HaveOccurred())
		})

		It("should handle reader errors in DecodeReader", func() {
			errReader := &errorReader{}
			decReader := coder.DecodeReader(errReader)

			buffer := make([]byte, 100)
			_, err := decReader.Read(buffer)
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in EncodeReader", func() {
			errReader := &errorReader{}
			encReader := coder.EncodeReader(errReader)

			err := encReader.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in DecodeReader", func() {
			errReader := &errorReader{}
			decReader := coder.DecodeReader(errReader)

			err := decReader.Close()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Writer Edge Cases", func() {
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

		It("should handle writer errors in EncodeWriter", func() {
			errWriter := &errorWriter{}
			encWriter := coder.EncodeWriter(errWriter)

			_, err := encWriter.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle writer errors in DecodeWriter", func() {
			encrypted := coder.Encode([]byte("test"))

			errWriter := &errorWriter{}
			decWriter := coder.DecodeWriter(errWriter)

			_, err := decWriter.Write(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in EncodeWriter", func() {
			errWriter := &errorWriter{}
			encWriter := coder.EncodeWriter(errWriter)

			err := encWriter.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle close errors in DecodeWriter", func() {
			errWriter := &errorWriter{}
			decWriter := coder.DecodeWriter(errWriter)

			err := decWriter.Close()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Security Edge Cases", func() {
		It("should produce different outputs with different keys", func() {
			key1, _ := encaes.GenKey()
			key2, _ := encaes.GenKey()

			coder1, _ := encaes.New(key1, nonce)
			coder2, _ := encaes.New(key2, nonce)

			plaintext := []byte("secret data")
			encrypted1 := coder1.Encode(plaintext)
			encrypted2 := coder2.Encode(plaintext)

			Expect(encrypted1).ToNot(Equal(encrypted2))
		})

		It("should produce different outputs with different nonces", func() {
			nonce1, _ := encaes.GenNonce()
			nonce2, _ := encaes.GenNonce()

			coder1, _ := encaes.New(key, nonce1)
			coder2, _ := encaes.New(key, nonce2)

			plaintext := []byte("secret data")
			encrypted1 := coder1.Encode(plaintext)
			encrypted2 := coder2.Encode(plaintext)

			Expect(encrypted1).ToNot(Equal(encrypted2))
		})

		It("should not allow decryption with wrong key", func() {
			key1, _ := encaes.GenKey()
			key2, _ := encaes.GenKey()

			coder1, _ := encaes.New(key1, nonce)
			coder2, _ := encaes.New(key2, nonce)

			plaintext := []byte("secret data")
			encrypted := coder1.Encode(plaintext)

			_, err := coder2.Decode(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should detect tampered data", func() {
			coder, _ := encaes.New(key, nonce)

			plaintext := []byte("important data")
			encrypted := coder.Encode(plaintext)

			// Tamper with the encrypted data
			if len(encrypted) > 0 {
				encrypted[len(encrypted)/2] ^= 0xFF
			}

			_, err := coder.Decode(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should detect truncated data", func() {
			coder, _ := encaes.New(key, nonce)

			plaintext := []byte("data to be truncated")
			encrypted := coder.Encode(plaintext)

			// Truncate the encrypted data
			if len(encrypted) > 5 {
				truncated := encrypted[:len(encrypted)-5]
				_, err := coder.Decode(truncated)
				Expect(err).To(HaveOccurred())
			}
		})

		It("should detect extended data", func() {
			coder, _ := encaes.New(key, nonce)

			plaintext := []byte("data to be extended")
			encrypted := coder.Encode(plaintext)

			// Extend the encrypted data
			extended := append(encrypted, []byte("extra data")...)
			_, err := coder.Decode(extended)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Concurrency Safety", func() {
		It("should handle concurrent encoding", func() {
			coder, _ := encaes.New(key, nonce)

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(id int) {
					defer GinkgoRecover()
					data := []byte(fmt.Sprintf("message %d", id))
					encrypted := coder.Encode(data)
					Expect(encrypted).ToNot(BeNil())
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent decoding", func() {
			coder, _ := encaes.New(key, nonce)

			// Pre-encrypt messages
			var encrypted [][]byte
			for i := 0; i < 10; i++ {
				data := []byte(fmt.Sprintf("message %d", i))
				encrypted = append(encrypted, coder.Encode(data))
			}

			done := make(chan bool, 10)
			for i, enc := range encrypted {
				go func(id int, data []byte) {
					defer GinkgoRecover()
					decrypted, err := coder.Decode(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(decrypted).ToNot(BeNil())
					done <- true
				}(i, enc)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Reset Edge Cases", func() {
		It("should handle multiple resets", func() {
			coder, _ := encaes.New(key, nonce)

			coder.Reset()
			coder.Reset()
			coder.Reset()

			// After multiple resets, encode should return empty
			result := coder.Encode([]byte("test"))
			Expect(len(result)).To(Equal(0))
		})

		It("should handle operations after reset", func() {
			coder, _ := encaes.New(key, nonce)

			plaintext := []byte("test before reset")
			encrypted := coder.Encode(plaintext)
			Expect(len(encrypted)).To(BeNumerically(">", 0))

			coder.Reset()

			// After reset
			result := coder.Encode([]byte("test after reset"))
			Expect(len(result)).To(Equal(0))

			result2, err := coder.Decode(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(result2)).To(Equal(0))
		})
	})
})
