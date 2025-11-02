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
	"encoding/json"
	"os"

	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permission Edge Cases", func() {
	Describe("Zero Permission", func() {
		It("should handle zero permission", func() {
			perm, err := Parse("0")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0)))
		})

		It("should format zero permission", func() {
			perm := Perm(0)
			Expect(perm.String()).To(Equal("0"))
		})

		It("should handle zero in JSON", func() {
			perm := Perm(0)
			data, err := json.Marshal(perm)
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = json.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(perm))
		})
	})

	Describe("Special Permissions", func() {
		It("should handle setuid bit (04000)", func() {
			perm, err := Parse("04755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(04755)))
			// Note: FileMode conversion may not preserve setuid bit as a mode flag
			// The value is preserved in the numeric representation
		})

		It("should handle setgid bit (02000)", func() {
			perm, err := Parse("02755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(02755)))
			// Note: FileMode conversion may not preserve setgid bit as a mode flag
			// The value is preserved in the numeric representation
		})

		It("should handle sticky bit (01000)", func() {
			perm, err := Parse("01777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(01777)))
			// Note: FileMode conversion may not preserve sticky bit as a mode flag
			// The value is preserved in the numeric representation
		})

		It("should handle all special bits together", func() {
			perm, err := Parse("07777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(07777)))
		})
	})

	Describe("Common Permissions", func() {
		DescribeTable("Standard Unix Permissions",
			func(octal string, expected uint64) {
				perm, err := Parse(octal)
				Expect(err).ToNot(HaveOccurred())
				Expect(perm.Uint64()).To(Equal(expected))
			},
			Entry("readonly file", "0400", uint64(0400)),
			Entry("writable by owner", "0600", uint64(0600)),
			Entry("standard file", "0644", uint64(0644)),
			Entry("executable by owner", "0700", uint64(0700)),
			Entry("standard executable", "0755", uint64(0755)),
			Entry("group writable", "0775", uint64(0775)),
			Entry("world writable", "0777", uint64(0777)),
		)
	})

	Describe("Quote Handling", func() {
		It("should strip double quotes", func() {
			perm, err := Parse("\"0644\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should strip single quotes", func() {
			perm, err := Parse("'0755'")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should strip multiple quotes", func() {
			perm, err := Parse("\"'0644'\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should handle quotes in ParseByte", func() {
			perm, err := ParseByte([]byte("\"0777\""))
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})
	})

	Describe("Invalid Inputs", func() {
		It("should reject invalid octal digits", func() {
			_, err := Parse("0888")
			Expect(err).To(HaveOccurred())
		})

		It("should reject non-numeric input", func() {
			_, err := Parse("rwxr-xr-x")
			Expect(err).To(HaveOccurred())
		})

		It("should reject empty string", func() {
			_, err := Parse("")
			Expect(err).To(HaveOccurred())
		})

		It("should reject whitespace", func() {
			_, err := Parse("   ")
			Expect(err).To(HaveOccurred())
		})

		It("should reject mixed valid/invalid", func() {
			_, err := Parse("064x")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Boundary Values", func() {
		It("should handle maximum 3-digit permission", func() {
			perm, err := Parse("0777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0777)))
		})

		It("should handle maximum 4-digit permission", func() {
			perm, err := Parse("07777")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(07777)))
		})

		It("should handle single digit", func() {
			perm, err := Parse("7")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(07)))
		})

		It("should handle two digits", func() {
			perm, err := Parse("77")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(077)))
		})
	})

	Describe("Type Conversions Edge Cases", func() {
		It("should handle Int64 with very large Perm value", func() {
			perm := Perm(07777)
			int64Val := perm.Int64()
			Expect(int64Val).To(Equal(int64(07777)))
		})

		It("should handle Int32 with valid Perm value", func() {
			perm := Perm(0644)
			int32Val := perm.Int32()
			Expect(int32Val).To(Equal(int32(0644)))
		})

		It("should handle Uint conversions", func() {
			perm := Perm(0755)
			Expect(perm.Uint()).To(Equal(uint(0755)))
			Expect(perm.Uint32()).To(Equal(uint32(0755)))
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})
	})

	Describe("FileMode Integration", func() {
		It("should create valid os.FileMode", func() {
			perm := Perm(0644)
			fileMode := perm.FileMode()
			Expect(fileMode.Perm()).To(Equal(os.FileMode(0644)))
		})

		It("should preserve permission bits in FileMode", func() {
			perm := Perm(0755)
			fileMode := perm.FileMode()
			Expect(fileMode.Perm()).To(Equal(os.FileMode(0755)))
		})

		It("should preserve permission value in FileMode", func() {
			perm := Perm(04755)
			fileMode := perm.FileMode()
			// Verify the FileMode contains the permission bits
			Expect(fileMode).To(Equal(os.FileMode(04755)))
		})
	})

	Describe("Concurrent Safety", func() {
		It("should handle concurrent parsing", func() {
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					perm, err := Parse("0644")
					Expect(err).ToNot(HaveOccurred())
					Expect(perm.Uint64()).To(Equal(uint64(0644)))
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent formatting", func() {
			perm := Perm(0755)
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					str := perm.String()
					Expect(str).To(Equal("0755"))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Real-world Scenarios", func() {
		It("should handle typical file permissions", func() {
			// Common file permissions
			testCases := []struct {
				name  string
				octal string
			}{
				{"private file", "0600"},
				{"public readable file", "0644"},
				{"executable script", "0755"},
				{"group shared file", "0664"},
			}

			for _, tc := range testCases {
				perm, err := Parse(tc.octal)
				Expect(err).ToNot(HaveOccurred(), "Failed for "+tc.name)
				Expect(perm.String()).To(Equal(tc.octal), "Mismatch for "+tc.name)
			}
		})

		It("should handle directory permissions", func() {
			// Common directory permissions
			dirPerms := []string{"0755", "0775", "0700", "01777"}

			for _, octal := range dirPerms {
				perm, err := Parse(octal)
				Expect(err).ToNot(HaveOccurred())
				fileMode := perm.FileMode()
				Expect(fileMode).ToNot(BeNil())
			}
		})
	})
})
