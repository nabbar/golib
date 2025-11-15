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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libenc "github.com/nabbar/golib/encoding"
	enchex "github.com/nabbar/golib/encoding/hexa"
)

var _ = Describe("Hexadecimal Encoding and Decoding", func() {
	Describe("New", func() {
		It("should create a new hexadecimal coder instance", func() {
			coder := enchex.New()
			Expect(coder).ToNot(BeNil())
		})

		It("should create independent instances", func() {
			coder1 := enchex.New()
			coder2 := enchex.New()

			Expect(coder1).ToNot(BeNil())
			Expect(coder2).ToNot(BeNil())

			// Both should produce same output for same input
			plaintext := []byte("test")
			encoded1 := coder1.Encode(plaintext)
			encoded2 := coder2.Encode(plaintext)

			Expect(encoded1).To(Equal(encoded2))
		})
	})

	Describe("Encode", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should encode simple text", func() {
			plaintext := []byte("Hello, World!")
			encoded := coder.Encode(plaintext)

			Expect(encoded).ToNot(BeNil())
			Expect(len(encoded)).To(Equal(len(plaintext) * 2)) // Hex doubles the size
			Expect(string(encoded)).To(Equal("48656c6c6f2c20576f726c6421"))
		})

		It("should handle empty byte slice", func() {
			encoded := coder.Encode([]byte{})
			Expect(encoded).ToNot(BeNil())
			Expect(len(encoded)).To(Equal(0))
		})

		It("should handle nil byte slice", func() {
			encoded := coder.Encode(nil)
			Expect(encoded).ToNot(BeNil())
			Expect(len(encoded)).To(Equal(0))
		})

		It("should encode single byte", func() {
			encoded := coder.Encode([]byte{0x42})
			Expect(string(encoded)).To(Equal("42"))
		})

		It("should encode all zero bytes", func() {
			data := []byte{0x00, 0x00, 0x00}
			encoded := coder.Encode(data)
			Expect(string(encoded)).To(Equal("000000"))
		})

		It("should encode all 0xFF bytes", func() {
			data := []byte{0xFF, 0xFF, 0xFF}
			encoded := coder.Encode(data)
			Expect(string(encoded)).To(Equal("ffffff"))
		})

		It("should encode binary data", func() {
			data := []byte{0x00, 0xFF, 0x7F, 0x80, 0x01, 0xFE}
			encoded := coder.Encode(data)
			Expect(string(encoded)).To(Equal("00ff7f8001fe"))
		})

		It("should handle large data", func() {
			// Test with 1MB of data
			largeData := make([]byte, 1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			encoded := coder.Encode(largeData)
			Expect(encoded).ToNot(BeNil())
			Expect(len(encoded)).To(Equal(len(largeData) * 2))
		})

		It("should produce deterministic output", func() {
			plaintext := []byte("consistent message")
			encoded1 := coder.Encode(plaintext)
			encoded2 := coder.Encode(plaintext)

			Expect(encoded1).To(Equal(encoded2))
		})

		It("should encode UTF-8 text correctly", func() {
			plaintext := []byte("Hello 世界")
			encoded := coder.Encode(plaintext)

			Expect(encoded).ToNot(BeNil())
			Expect(len(encoded)).To(Equal(len(plaintext) * 2))
		})
	})

	Describe("Decode", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
		})

		AfterEach(func() {
			if coder != nil {
				coder.Reset()
			}
		})

		It("should decode hex-encoded data", func() {
			hexData := []byte("48656c6c6f")
			decoded, err := coder.Decode(hexData)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should handle empty byte slice", func() {
			decoded, err := coder.Decode([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).ToNot(BeNil())
			Expect(len(decoded)).To(Equal(0))
		})

		It("should handle nil byte slice", func() {
			decoded, err := coder.Decode(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).ToNot(BeNil())
			Expect(len(decoded)).To(Equal(0))
		})

		It("should return error for invalid hex characters", func() {
			invalidHex := []byte("48656c6c6g") // 'g' is invalid
			_, err := coder.Decode(invalidHex)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for odd-length hex string", func() {
			oddHex := []byte("48656c6c6") // Odd length
			_, err := coder.Decode(oddHex)
			Expect(err).To(HaveOccurred())
		})

		It("should handle uppercase hex", func() {
			hexData := []byte("48656C6C6F")
			decoded, err := coder.Decode(hexData)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should handle mixed case hex", func() {
			hexData := []byte("48656C6c6F")
			decoded, err := coder.Decode(hexData)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(decoded)).To(Equal("Hello"))
		})

		It("should decode all zeros", func() {
			hexData := []byte("000000")
			decoded, err := coder.Decode(hexData)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal([]byte{0x00, 0x00, 0x00}))
		})

		It("should decode all FFs", func() {
			hexData := []byte("ffffff")
			decoded, err := coder.Decode(hexData)

			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal([]byte{0xFF, 0xFF, 0xFF}))
		})
	})

	Describe("Round-trip Encoding/Decoding", func() {
		var coder libenc.Coder

		BeforeEach(func() {
			coder = enchex.New()
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
				[]byte{0x00, 0xFF, 0x7F, 0x80}, // Binary data
				make([]byte, 1000),             // Large zeroed data
			}

			for _, original := range testCases {
				encoded := coder.Encode(original)
				decoded, err := coder.Decode(encoded)

				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			}
		})

		It("should handle multiple encode/decode operations", func() {
			plaintext := []byte("persistent data")

			for i := 0; i < 10; i++ {
				encoded := coder.Encode(plaintext)
				decoded, err := coder.Decode(encoded)

				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(plaintext))
			}
		})

		It("should handle very large data", func() {
			// 10MB of data
			largeData := make([]byte, 10*1024*1024)
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			encoded := coder.Encode(largeData)
			decoded, err := coder.Decode(encoded)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(decoded)).To(Equal(len(largeData)))
			Expect(decoded).To(Equal(largeData))
		})
	})

	Describe("Reset", func() {
		It("should reset coder state", func() {
			coder := enchex.New()

			// Encode something first
			plaintext := []byte("test")
			encoded := coder.Encode(plaintext)
			Expect(encoded).ToNot(BeNil())

			// Reset the coder (no-op for hexa, but should not panic)
			coder.Reset()

			// After reset, should still work
			encoded2 := coder.Encode(plaintext)
			Expect(encoded2).To(Equal(encoded))
		})
	})
})
