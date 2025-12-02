/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package perm_test

import (
	"math"
	"os"

	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Coverage Improvements", func() {
	Describe("ParseFileMode", func() {
		It("should convert os.FileMode to Perm", func() {
			mode := os.FileMode(0644)
			perm := ParseFileMode(mode)
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should handle executable permission", func() {
			mode := os.FileMode(0755)
			perm := ParseFileMode(mode)
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should handle directory mode", func() {
			mode := os.ModeDir | os.FileMode(0755)
			perm := ParseFileMode(mode)
			// Should include the directory bit
			Expect(perm.FileMode()).To(Equal(mode))
		})
	})

	Describe("Type Conversions", func() {
		Context("Int64", func() {
			It("should handle normal values", func() {
				p := Perm(0644)
				Expect(p.Int64()).To(Equal(int64(0644)))
			})

			It("should handle maximum permission value", func() {
				// Perm is uint32, so max value is MaxUint32
				p := Perm(math.MaxUint32)
				result := p.Int64()
				Expect(result).To(Equal(int64(math.MaxUint32)))
			})
		})

		Context("Int", func() {
			It("should handle normal values", func() {
				p := Perm(0755)
				Expect(p.Int()).To(Equal(int(0755)))
			})

			It("should handle large permission values", func() {
				p := Perm(0777777) // Large but valid octal
				Expect(p.Int()).To(Equal(int(0777777)))
			})
		})

		Context("Uint32", func() {
			It("should handle normal values", func() {
				p := Perm(0777)
				Expect(p.Uint32()).To(Equal(uint32(0777)))
			})

			It("should handle maximum uint32 value", func() {
				p := Perm(math.MaxUint32)
				Expect(p.Uint32()).To(Equal(uint32(math.MaxUint32)))
			})
		})

		Context("Uint", func() {
			It("should handle normal values", func() {
				p := Perm(0600)
				Expect(p.Uint()).To(Equal(uint(0600)))
			})

			It("should handle large permission values", func() {
				p := Perm(0177777) // All special bits + all permissions
				Expect(p.Uint()).To(Equal(uint(0177777)))
			})
		})
	})

	Describe("Symbolic Parsing Edge Cases", func() {
		It("should parse with file type prefix -", func() {
			perm, err := Parse("-rw-r--r--")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should parse directory with d prefix", func() {
			perm, err := Parse("drwxr-xr-x")
			Expect(err).ToNot(HaveOccurred())
			// Directory bit should be set
			Expect(perm.FileMode() & os.ModeDir).To(Equal(os.ModeDir))
		})

		It("should parse symbolic link with l prefix", func() {
			perm, err := Parse("lrwxrwxrwx")
			Expect(err).ToNot(HaveOccurred())
			// Symlink bit should be set
			Expect(perm.FileMode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
		})

		It("should parse character device with c prefix", func() {
			perm, err := Parse("crw-rw-rw-")
			Expect(err).ToNot(HaveOccurred())
			// Character device bits should be set
			Expect(perm.FileMode() & os.ModeCharDevice).To(Equal(os.ModeCharDevice))
		})

		It("should parse block device with b prefix", func() {
			perm, err := Parse("brw-rw-rw-")
			Expect(err).ToNot(HaveOccurred())
			// Device bit should be set
			Expect(perm.FileMode() & os.ModeDevice).To(Equal(os.ModeDevice))
		})

		It("should parse FIFO with p prefix", func() {
			perm, err := Parse("prw-rw-rw-")
			Expect(err).ToNot(HaveOccurred())
			// Named pipe bit should be set
			Expect(perm.FileMode() & os.ModeNamedPipe).To(Equal(os.ModeNamedPipe))
		})

		It("should parse socket with s prefix", func() {
			perm, err := Parse("srwxrwxrwx")
			Expect(err).ToNot(HaveOccurred())
			// Socket bit should be set
			Expect(perm.FileMode() & os.ModeSocket).To(Equal(os.ModeSocket))
		})

		It("should parse irregular file with D prefix", func() {
			perm, err := Parse("Drw-rw-rw-")
			Expect(err).ToNot(HaveOccurred())
			// Irregular bit should be set
			Expect(perm.FileMode() & os.ModeIrregular).To(Equal(os.ModeIrregular))
		})

		It("should reject invalid file type character", func() {
			_, err := Parse("Xrwxrwxrwx")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid file type character"))
		})

		It("should reject invalid read permission character", func() {
			_, err := Parse("Xwxrwxrwx")
			Expect(err).To(HaveOccurred())
		})

		It("should reject invalid write permission character", func() {
			_, err := Parse("rXxrwxrwx")
			Expect(err).To(HaveOccurred())
		})

		It("should reject invalid execute permission character", func() {
			_, err := Parse("rwXrwxrwx")
			Expect(err).To(HaveOccurred())
		})

		It("should reject string that is too short", func() {
			_, err := Parse("rwx")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid permission"))
		})

		It("should reject string that is too long", func() {
			_, err := Parse("rwxrwxrwxrwx")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid permission"))
		})
	})

	Describe("Additional Int32 Coverage", func() {
		It("should handle maximum uint32 value for Int32", func() {
			// When Perm value exceeds MaxInt32, Int32() should return MaxInt32
			p := Perm(math.MaxUint32) // This is > MaxInt32
			Expect(p.Int32()).To(Equal(int32(math.MaxInt32)))
		})

		It("should handle normal Int32 values", func() {
			p := Perm(0644)
			Expect(p.Int32()).To(Equal(int32(0644)))
		})
	})
})
