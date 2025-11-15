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
	"crypto/sha256"
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	encsha "github.com/nabbar/golib/encoding/sha256"
)

var _ = Describe("SHA-256 Hash Operations", func() {
	Describe("New", func() {
		It("should create a new hasher instance", func() {
			hasher := encsha.New()
			Expect(hasher).ToNot(BeNil())
		})

		It("should create independent instances", func() {
			h1 := encsha.New()
			h2 := encsha.New()

			hash1 := h1.Encode([]byte("test"))
			hash2 := h2.Encode([]byte("test"))

			Expect(hash1).To(Equal(hash2))
		})
	})

	Describe("Encode", func() {
		It("should hash simple text", func() {
			hasher := encsha.New()
			input := []byte("Hello, World!")

			result := hasher.Encode(input)
			expected := sha256.Sum256(input)

			Expect(result).To(Equal(expected[:]))
		})

		It("should hash empty input", func() {
			hasher := encsha.New()

			result := hasher.Encode([]byte{})
			expected := sha256.Sum256([]byte{})

			Expect(result).To(Equal(expected[:]))
		})

		It("should hash nil input", func() {
			hasher := encsha.New()

			result := hasher.Encode(nil)
			expected := sha256.Sum256(nil)

			Expect(result).To(Equal(expected[:]))
		})

		It("should produce deterministic output", func() {
			input := []byte("deterministic test")

			h1 := encsha.New()
			hash1 := h1.Encode(input)

			h2 := encsha.New()
			hash2 := h2.Encode(input)

			Expect(hash1).To(Equal(hash2))
		})

		It("should produce different hashes for different inputs", func() {
			hasher := encsha.New()

			hash1 := hasher.Encode([]byte("input1"))
			hasher.Reset()
			hash2 := hasher.Encode([]byte("input2"))

			Expect(hash1).ToNot(Equal(hash2))
		})

		It("should hash binary data", func() {
			hasher := encsha.New()
			binary := []byte{0x00, 0xFF, 0x7F, 0x80, 0xDE, 0xAD, 0xBE, 0xEF}

			result := hasher.Encode(binary)
			expected := sha256.Sum256(binary)

			Expect(result).To(Equal(expected[:]))
		})

		It("should hash large data", func() {
			hasher := encsha.New()
			largeData := make([]byte, 1024*1024) // 1MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			result := hasher.Encode(largeData)
			expected := sha256.Sum256(largeData)

			Expect(result).To(Equal(expected[:]))
		})

		It("should hash UTF-8 text", func() {
			hasher := encsha.New()
			utf8Text := []byte("Hello 世界 🔒")

			result := hasher.Encode(utf8Text)
			expected := sha256.Sum256(utf8Text)

			Expect(result).To(Equal(expected[:]))
		})
	})

	Describe("Known Test Vectors", func() {
		// NIST test vectors for SHA-256
		It("should match NIST test vector 1 (empty string)", func() {
			hasher := encsha.New()
			input := []byte("")

			result := hasher.Encode(input)
			expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

			Expect(hex.EncodeToString(result)).To(Equal(expected))
		})

		It("should match NIST test vector 2 (abc)", func() {
			hasher := encsha.New()
			input := []byte("abc")

			result := hasher.Encode(input)
			expected := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

			Expect(hex.EncodeToString(result)).To(Equal(expected))
		})

		It("should match NIST test vector 3", func() {
			hasher := encsha.New()
			input := []byte("abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq")

			result := hasher.Encode(input)
			expected := "248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1"

			Expect(hex.EncodeToString(result)).To(Equal(expected))
		})
	})

	Describe("Reset", func() {
		It("should reset the hasher state", func() {
			hasher := encsha.New()
			input := []byte("test data")

			// First hash
			hash1 := hasher.Encode(input)

			// Reset and hash again
			hasher.Reset()
			hash2 := hasher.Encode(input)

			Expect(hash1).To(Equal(hash2))
		})

		It("should allow hashing different data after reset", func() {
			hasher := encsha.New()

			hash1 := hasher.Encode([]byte("first"))
			hasher.Reset()
			hash2 := hasher.Encode([]byte("second"))

			Expect(hash1).ToNot(Equal(hash2))
		})

		It("should reset to empty state", func() {
			hasher := encsha.New()

			// Hash some data
			hasher.Encode([]byte("some data"))

			// Reset
			hasher.Reset()

			// Hash should be same as new instance
			result := hasher.Encode([]byte("test"))
			expected := encsha.New().Encode([]byte("test"))

			Expect(result).To(Equal(expected))
		})
	})

	Describe("Decode", func() {
		It("should return error (hashes are one-way)", func() {
			hasher := encsha.New()
			input := []byte("test")

			_, err := hasher.Decode(input)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected call"))
		})
	})

	Describe("Edge Cases", func() {
		It("should handle single byte", func() {
			hasher := encsha.New()
			input := []byte{0x42}

			result := hasher.Encode(input)
			expected := sha256.Sum256(input)

			Expect(result).To(Equal(expected[:]))
		})

		It("should handle all zero bytes", func() {
			hasher := encsha.New()
			input := make([]byte, 100)

			result := hasher.Encode(input)
			expected := sha256.Sum256(input)

			Expect(result).To(Equal(expected[:]))
		})

		It("should handle all 0xFF bytes", func() {
			hasher := encsha.New()
			input := make([]byte, 100)
			for i := range input {
				input[i] = 0xFF
			}

			result := hasher.Encode(input)
			expected := sha256.Sum256(input)

			Expect(result).To(Equal(expected[:]))
		})

		It("should handle sequential bytes", func() {
			hasher := encsha.New()
			input := make([]byte, 256)
			for i := range input {
				input[i] = byte(i)
			}

			result := hasher.Encode(input)
			expected := sha256.Sum256(input)

			Expect(result).To(Equal(expected[:]))
		})

		It("should return 32-byte hash", func() {
			hasher := encsha.New()
			result := hasher.Encode([]byte("test"))

			Expect(len(result)).To(Equal(32))
		})

		It("should handle multiple encodes without reset", func() {
			hasher := encsha.New()

			// First encode
			hasher.Encode([]byte("part1"))

			// Second encode (accumulates)
			result := hasher.Encode([]byte("part2"))

			// Should be hash of "part1part2"
			expected := sha256.Sum256([]byte("part1part2"))
			Expect(result).To(Equal(expected[:]))
		})
	})
})
