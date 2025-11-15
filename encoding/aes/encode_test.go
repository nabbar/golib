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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
)

var _ = Describe("AES Encoding and Decoding", func() {
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

	Describe("New", func() {
		It("should create a new coder instance", func() {
			coder, err := encaes.New(key, nonce)
			Expect(err).ToNot(HaveOccurred())
			Expect(coder).ToNot(BeNil())
		})

		It("should accept zero key and nonce", func() {
			var zeroKey [32]byte
			var zeroNonce [12]byte

			coder, err := encaes.New(zeroKey, zeroNonce)
			Expect(err).ToNot(HaveOccurred())
			Expect(coder).ToNot(BeNil())
		})

		It("should create different instances with same key/nonce", func() {
			coder1, err1 := encaes.New(key, nonce)
			coder2, err2 := encaes.New(key, nonce)

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(coder1).ToNot(BeNil())
			Expect(coder2).ToNot(BeNil())

			// Both should produce same encrypted output for same input
			plaintext := []byte("test message")
			encrypted1 := coder1.Encode(plaintext)
			encrypted2 := coder2.Encode(plaintext)

			Expect(encrypted1).To(Equal(encrypted2))
		})
	})

	Describe("Encode", func() {
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

		It("should encode simple text", func() {
			plaintext := []byte("Hello, World!")
			encrypted := coder.Encode(plaintext)

			Expect(encrypted).ToNot(BeNil())
			Expect(len(encrypted)).To(BeNumerically(">", len(plaintext)))
			Expect(encrypted).ToNot(Equal(plaintext))
		})

		It("should handle empty byte slice", func() {
			encrypted := coder.Encode([]byte{})
			Expect(encrypted).ToNot(BeNil())
			Expect(len(encrypted)).To(Equal(0))
		})

		It("should handle nil byte slice", func() {
			encrypted := coder.Encode(nil)
			Expect(encrypted).ToNot(BeNil())
			Expect(len(encrypted)).To(Equal(0))
		})

		It("should produce different output for same input with different nonces", func() {
			plaintext := []byte("test message")
			encrypted1 := coder.Encode(plaintext)

			// Create new coder with different nonce
			newNonce, _ := encaes.GenNonce()
			coder2, _ := encaes.New(key, newNonce)
			encrypted2 := coder2.Encode(plaintext)

			Expect(encrypted1).ToNot(Equal(encrypted2))
		})

		It("should handle large data", func() {
			// Test with 1MB of data
			largeData := make([]byte, 1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			encrypted := coder.Encode(largeData)
			Expect(encrypted).ToNot(BeNil())
			Expect(len(encrypted)).To(BeNumerically(">", len(largeData)))
		})

		It("should handle binary data", func() {
			binaryData := []byte{0x00, 0xFF, 0x7F, 0x80, 0x01, 0xFE}
			encrypted := coder.Encode(binaryData)

			Expect(encrypted).ToNot(BeNil())
			Expect(encrypted).ToNot(Equal(binaryData))
		})

		It("should produce deterministic output for same input", func() {
			plaintext := []byte("consistent message")
			encrypted1 := coder.Encode(plaintext)
			encrypted2 := coder.Encode(plaintext)

			Expect(encrypted1).To(Equal(encrypted2))
		})
	})

	Describe("Decode", func() {
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

		It("should decode encrypted data", func() {
			plaintext := []byte("Hello, World!")
			encrypted := coder.Encode(plaintext)

			decrypted, err := coder.Decode(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(plaintext))
		})

		It("should handle empty byte slice", func() {
			decrypted, err := coder.Decode([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).ToNot(BeNil())
			Expect(len(decrypted)).To(Equal(0))
		})

		It("should handle nil byte slice", func() {
			decrypted, err := coder.Decode(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).ToNot(BeNil())
			Expect(len(decrypted)).To(Equal(0))
		})

		It("should return error for invalid encrypted data", func() {
			invalidData := []byte("not encrypted data")
			_, err := coder.Decode(invalidData)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for corrupted encrypted data", func() {
			plaintext := []byte("test message")
			encrypted := coder.Encode(plaintext)

			// Corrupt the encrypted data
			encrypted[0] ^= 0xFF

			_, err := coder.Decode(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for truncated encrypted data", func() {
			plaintext := []byte("test message")
			encrypted := coder.Encode(plaintext)

			// Truncate the encrypted data
			truncated := encrypted[:len(encrypted)/2]

			_, err := coder.Decode(truncated)
			Expect(err).To(HaveOccurred())
		})

		It("should fail with wrong key", func() {
			plaintext := []byte("secret message")
			encrypted := coder.Encode(plaintext)

			// Create new coder with different key
			wrongKey, _ := encaes.GenKey()
			wrongCoder, _ := encaes.New(wrongKey, nonce)

			_, err := wrongCoder.Decode(encrypted)
			Expect(err).To(HaveOccurred())
		})

		It("should fail with wrong nonce", func() {
			plaintext := []byte("secret message")
			encrypted := coder.Encode(plaintext)

			// Create new coder with different nonce
			wrongNonce, _ := encaes.GenNonce()
			wrongCoder, _ := encaes.New(key, wrongNonce)

			_, err := wrongCoder.Decode(encrypted)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Round-trip Encoding/Decoding", func() {
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

		It("should preserve data through encode/decode cycle", func() {
			testCases := [][]byte{
				[]byte("Simple text"),
				[]byte("Text with numbers: 12345"),
				[]byte("Special chars: !@#$%^&*()"),
				[]byte("UTF-8: こんにちは世界"),
				[]byte("Emoji: 🔒🔐🔑"),
				make([]byte, 1000), // Large zeroed data
			}

			for _, original := range testCases {
				encrypted := coder.Encode(original)
				decrypted, err := coder.Decode(encrypted)

				Expect(err).ToNot(HaveOccurred())
				Expect(decrypted).To(Equal(original))
			}
		})

		It("should handle multiple encode/decode operations", func() {
			plaintext := []byte("persistent data")

			for i := 0; i < 10; i++ {
				encrypted := coder.Encode(plaintext)
				decrypted, err := coder.Decode(encrypted)

				Expect(err).ToNot(HaveOccurred())
				Expect(decrypted).To(Equal(plaintext))
			}
		})
	})

	Describe("Reset", func() {
		It("should reset coder state", func() {
			coder, err := encaes.New(key, nonce)
			Expect(err).ToNot(HaveOccurred())

			// Encode something first
			plaintext := []byte("test")
			encrypted := coder.Encode(plaintext)
			Expect(encrypted).ToNot(BeNil())

			// Reset the coder
			coder.Reset()

			// After reset, encoding should return empty slice
			result := coder.Encode(plaintext)
			Expect(result).ToNot(BeNil())
			Expect(len(result)).To(Equal(0))
		})

		It("should handle decoding after reset", func() {
			coder, _ := encaes.New(key, nonce)
			plaintext := []byte("test")
			encrypted := coder.Encode(plaintext)

			coder.Reset()

			// After reset, decoding should return empty slice
			result, err := coder.Decode(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result)).To(Equal(0))
		})
	})
})
