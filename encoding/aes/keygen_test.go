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
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encaes "github.com/nabbar/golib/encoding/aes"
)

var _ = Describe("AES Key and Nonce Generation", func() {
	Describe("GenKey", func() {
		It("should generate a valid 32-byte key", func() {
			key, err := encaes.GenKey()
			Expect(err).ToNot(HaveOccurred())
			Expect(key).ToNot(BeZero())
			Expect(len(key)).To(Equal(32))
		})

		It("should generate unique keys", func() {
			key1, err1 := encaes.GenKey()
			key2, err2 := encaes.GenKey()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(key1).ToNot(Equal(key2))
		})

		It("should generate keys with proper entropy", func() {
			// Generate multiple keys and verify they're different
			keys := make(map[[32]byte]bool)
			for i := 0; i < 10; i++ {
				key, err := encaes.GenKey()
				Expect(err).ToNot(HaveOccurred())
				Expect(keys[key]).To(BeFalse(), "Duplicate key generated")
				keys[key] = true
			}
		})
	})

	Describe("GenNonce", func() {
		It("should generate a valid 12-byte nonce", func() {
			nonce, err := encaes.GenNonce()
			Expect(err).ToNot(HaveOccurred())
			Expect(nonce).ToNot(BeZero())
			Expect(len(nonce)).To(Equal(12))
		})

		It("should generate unique nonces", func() {
			nonce1, err1 := encaes.GenNonce()
			nonce2, err2 := encaes.GenNonce()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(nonce1).ToNot(Equal(nonce2))
		})

		It("should generate nonces with proper entropy", func() {
			// Generate multiple nonces and verify they're different
			nonces := make(map[[12]byte]bool)
			for i := 0; i < 10; i++ {
				nonce, err := encaes.GenNonce()
				Expect(err).ToNot(HaveOccurred())
				Expect(nonces[nonce]).To(BeFalse(), "Duplicate nonce generated")
				nonces[nonce] = true
			}
		})
	})

	Describe("GetHexKey", func() {
		It("should parse valid hex string to key", func() {
			// Generate a key and convert to hex
			originalKey, err := encaes.GenKey()
			Expect(err).ToNot(HaveOccurred())

			hexStr := hex.EncodeToString(originalKey[:])
			Expect(hexStr).To(HaveLen(64)) // 32 bytes = 64 hex characters

			// Parse it back
			parsedKey, err := encaes.GetHexKey(hexStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedKey).To(Equal(originalKey))
		})

		It("should handle uppercase hex strings", func() {
			originalKey, _ := encaes.GenKey()
			hexStr := hex.EncodeToString(originalKey[:])
			hexStrUpper := string([]byte(hexStr)) // Already lowercase, convert case

			parsedKey, err := encaes.GetHexKey(hexStrUpper)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedKey).To(Equal(originalKey))
		})

		It("should return error for invalid hex string", func() {
			_, err := encaes.GetHexKey("invalid-hex-string!!!")
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty string gracefully", func() {
			key, err := encaes.GetHexKey("")
			Expect(err).ToNot(HaveOccurred())
			// Empty input results in zero-filled key
			var zeroKey [32]byte
			Expect(key).To(Equal(zeroKey))
		})

		It("should handle short hex strings", func() {
			// Test with less than 32 bytes of hex data
			shortHex := "0123456789abcdef" // Only 8 bytes
			key, err := encaes.GetHexKey(shortHex)
			Expect(err).ToNot(HaveOccurred())
			// Should copy available bytes and zero-fill rest
			// First 8 bytes should match decoded hex, rest should be zero
			Expect(key[0]).To(Equal(byte(0x01)))
			Expect(key[1]).To(Equal(byte(0x23)))
			Expect(key[31]).To(Equal(byte(0x00))) // Zero-filled
		})

		It("should handle exact 64-character hex string", func() {
			hexStr := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			key, err := encaes.GetHexKey(hexStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(key)).To(Equal(32))
		})

		It("should handle longer hex strings by truncating", func() {
			// More than 64 hex chars
			longHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00112233"
			key, err := encaes.GetHexKey(longHex)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(key)).To(Equal(32))
		})

		It("should return error for odd-length hex string", func() {
			// Odd number of hex characters
			_, err := encaes.GetHexKey("0123456789abcde") // 15 chars
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetHexNonce", func() {
		It("should parse valid hex string to nonce", func() {
			// Generate a nonce and convert to hex
			originalNonce, err := encaes.GenNonce()
			Expect(err).ToNot(HaveOccurred())

			hexStr := hex.EncodeToString(originalNonce[:])
			Expect(hexStr).To(HaveLen(24)) // 12 bytes = 24 hex characters

			// Parse it back
			parsedNonce, err := encaes.GetHexNonce(hexStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedNonce).To(Equal(originalNonce))
		})

		It("should return error for invalid hex string", func() {
			_, err := encaes.GetHexNonce("invalid-hex-nonce!!!")
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty string gracefully", func() {
			nonce, err := encaes.GetHexNonce("")
			Expect(err).ToNot(HaveOccurred())
			// Empty input results in zero-filled nonce
			var zeroNonce [12]byte
			Expect(nonce).To(Equal(zeroNonce))
		})

		It("should handle short hex strings", func() {
			// Test with less than 12 bytes of hex data
			shortHex := "0123456789ab" // Only 6 bytes
			nonce, err := encaes.GetHexNonce(shortHex)
			Expect(err).ToNot(HaveOccurred())
			// Should copy available bytes and zero-fill rest
			// First 6 bytes should match decoded hex, rest should be zero
			Expect(nonce[0]).To(Equal(byte(0x01)))
			Expect(nonce[1]).To(Equal(byte(0x23)))
			Expect(nonce[11]).To(Equal(byte(0x00))) // Zero-filled
		})

		It("should handle exact 24-character hex string", func() {
			hexStr := "0123456789abcdef01234567"
			nonce, err := encaes.GetHexNonce(hexStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(nonce)).To(Equal(12))
		})

		It("should handle longer hex strings by truncating", func() {
			// More than 24 hex chars
			longHex := "0123456789abcdef0123456789abcdef"
			nonce, err := encaes.GetHexNonce(longHex)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(nonce)).To(Equal(12))
		})

		It("should return error for odd-length hex string", func() {
			// Odd number of hex characters
			_, err := encaes.GetHexNonce("0123456789abc") // 13 chars
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Round-trip conversions", func() {
		It("should preserve key through hex encoding/decoding", func() {
			originalKey, _ := encaes.GenKey()

			hexStr := hex.EncodeToString(originalKey[:])
			parsedKey, err := encaes.GetHexKey(hexStr)

			Expect(err).ToNot(HaveOccurred())
			Expect(parsedKey).To(Equal(originalKey))
		})

		It("should preserve nonce through hex encoding/decoding", func() {
			originalNonce, _ := encaes.GenNonce()

			hexStr := hex.EncodeToString(originalNonce[:])
			parsedNonce, err := encaes.GetHexNonce(hexStr)

			Expect(err).ToNot(HaveOccurred())
			Expect(parsedNonce).To(Equal(originalNonce))
		})
	})
})
