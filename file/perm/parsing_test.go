/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2024 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package perm_test

import (
	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permission Parsing", func() {
	Describe("Parse", func() {
		It("should parse valid octal string 0644", func() {
			perm, err := Parse("0644")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse valid octal string 0755", func() {
			perm, err := Parse("0755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should parse valid octal string 0777", func() {
			perm, err := Parse("0777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})

		It("should parse valid octal string 0400", func() {
			perm, err := Parse("0400")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0400)))
		})

		It("should parse octal string without leading zero", func() {
			perm, err := Parse("644")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse string with quotes", func() {
			perm, err := Parse("\"0644\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse string with single quotes", func() {
			perm, err := Parse("'0755'")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should return error for invalid octal", func() {
			_, err := Parse("0999")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-numeric string", func() {
			_, err := Parse("rwxr-xr-x")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for empty string", func() {
			_, err := Parse("")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ParseInt", func() {
		It("should parse valid int 420 (0644)", func() {
			perm, err := ParseInt(420)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse valid int 493 (0755)", func() {
			perm, err := ParseInt(493)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should parse valid int 511 (0777)", func() {
			perm, err := ParseInt(511)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})

		It("should parse valid int 0", func() {
			perm, err := ParseInt(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0)))
		})

		It("should handle negative int by converting to octal", func() {
			_, err := ParseInt(-1)
			// Negative values will be converted to octal string, which should fail
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ParseInt64", func() {
		It("should parse valid int64 420 (0644)", func() {
			perm, err := ParseInt64(420)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse valid int64 493 (0755)", func() {
			perm, err := ParseInt64(493)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should parse large valid int64", func() {
			perm, err := ParseInt64(511)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})
	})

	Describe("ParseByte", func() {
		It("should parse valid byte slice", func() {
			perm, err := ParseByte([]byte("0644"))
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse byte slice without leading zero", func() {
			perm, err := ParseByte([]byte("755"))
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should parse byte slice with quotes", func() {
			perm, err := ParseByte([]byte("\"0777\""))
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})

		It("should return error for invalid byte slice", func() {
			_, err := ParseByte([]byte("invalid"))
			Expect(err).To(HaveOccurred())
		})

		It("should return error for empty byte slice", func() {
			_, err := ParseByte([]byte(""))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Edge Cases", func() {
		It("should handle maximum valid permission 07777", func() {
			perm, err := Parse("07777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(07777)))
		})

		It("should handle permission with special bits", func() {
			// Setuid bit (04000)
			perm, err := Parse("04755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(04755)))
		})

		It("should handle permission with sticky bit", func() {
			// Sticky bit (01000)
			perm, err := Parse("01777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(01777)))
		})

		It("should handle setgid bit", func() {
			// Setgid bit (02000)
			perm, err := Parse("02755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(02755)))
		})
	})
})
