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

package compress_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/compress"
)

var _ = Describe("TC-AL-001: Algorithm Type", func() {
	Context("TC-AL-002: List operations", func() {
		It("TC-AL-003: should return all algorithms", func() {
			lst := compress.List()
			Expect(lst).To(HaveLen(5))
			Expect(lst).To(ContainElements(
				compress.None,
				compress.Gzip,
				compress.Bzip2,
				compress.LZ4,
				compress.XZ,
			))
		})

		It("TC-AL-004: should return string list", func() {
			lst := compress.ListString()
			Expect(lst).To(HaveLen(5))
			Expect(lst).To(ContainElements("none", "gzip", "bzip2", "lz4", "xz"))
		})
	})

	Context("TC-AL-005: String method", func() {
		It("TC-AL-006: should return correct string for None", func() {
			Expect(compress.None.String()).To(Equal("none"))
		})

		It("TC-AL-007: should return correct string for Gzip", func() {
			Expect(compress.Gzip.String()).To(Equal("gzip"))
		})

		It("TC-AL-008: should return correct string for Bzip2", func() {
			Expect(compress.Bzip2.String()).To(Equal("bzip2"))
		})

		It("TC-AL-009: should return correct string for LZ4", func() {
			Expect(compress.LZ4.String()).To(Equal("lz4"))
		})

		It("TC-AL-010: should return correct string for XZ", func() {
			Expect(compress.XZ.String()).To(Equal("xz"))
		})

		It("TC-AL-011: should return none for invalid algorithm", func() {
			var invalid compress.Algorithm = 99
			Expect(invalid.String()).To(Equal("none"))
		})
	})

	Context("TC-AL-012: Extension method", func() {
		It("TC-AL-013: should return empty string for None", func() {
			Expect(compress.None.Extension()).To(BeEmpty())
		})

		It("TC-AL-014: should return .gz for Gzip", func() {
			Expect(compress.Gzip.Extension()).To(Equal(".gz"))
		})

		It("TC-AL-015: should return .bz2 for Bzip2", func() {
			Expect(compress.Bzip2.Extension()).To(Equal(".bz2"))
		})

		It("TC-AL-016: should return .lz4 for LZ4", func() {
			Expect(compress.LZ4.Extension()).To(Equal(".lz4"))
		})

		It("TC-AL-017: should return .xz for XZ", func() {
			Expect(compress.XZ.Extension()).To(Equal(".xz"))
		})

		It("TC-AL-018: should return empty for invalid algorithm", func() {
			var invalid compress.Algorithm = 99
			Expect(invalid.Extension()).To(BeEmpty())
		})
	})

	Context("TC-AL-019: IsNone method", func() {
		It("TC-AL-020: should return true for None", func() {
			Expect(compress.None.IsNone()).To(BeTrue())
		})

		It("TC-AL-021: should return false for Gzip", func() {
			Expect(compress.Gzip.IsNone()).To(BeFalse())
		})

		It("TC-AL-022: should return false for Bzip2", func() {
			Expect(compress.Bzip2.IsNone()).To(BeFalse())
		})

		It("TC-AL-023: should return false for LZ4", func() {
			Expect(compress.LZ4.IsNone()).To(BeFalse())
		})

		It("TC-AL-024: should return false for XZ", func() {
			Expect(compress.XZ.IsNone()).To(BeFalse())
		})
	})

	Context("TC-AL-025: DetectHeader method", func() {
		It("TC-AL-026: should detect Gzip header", func() {
			header := []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00}
			Expect(compress.Gzip.DetectHeader(header)).To(BeTrue())
			Expect(compress.Bzip2.DetectHeader(header)).To(BeFalse())
			Expect(compress.LZ4.DetectHeader(header)).To(BeFalse())
			Expect(compress.XZ.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-027: should detect Bzip2 header", func() {
			header := []byte{'B', 'Z', 'h', '9', 0x00, 0x00}
			Expect(compress.Bzip2.DetectHeader(header)).To(BeTrue())
			Expect(compress.Gzip.DetectHeader(header)).To(BeFalse())
			Expect(compress.LZ4.DetectHeader(header)).To(BeFalse())
			Expect(compress.XZ.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-028: should detect LZ4 header", func() {
			header := []byte{0x04, 0x22, 0x4D, 0x18, 0x00, 0x00}
			Expect(compress.LZ4.DetectHeader(header)).To(BeTrue())
			Expect(compress.Gzip.DetectHeader(header)).To(BeFalse())
			Expect(compress.Bzip2.DetectHeader(header)).To(BeFalse())
			Expect(compress.XZ.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-029: should detect XZ header", func() {
			header := []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}
			Expect(compress.XZ.DetectHeader(header)).To(BeTrue())
			Expect(compress.Gzip.DetectHeader(header)).To(BeFalse())
			Expect(compress.Bzip2.DetectHeader(header)).To(BeFalse())
			Expect(compress.LZ4.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-030: should return false for short header", func() {
			header := []byte{0x1F, 0x8B, 0x08}
			Expect(compress.Gzip.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-031: should return false for empty header", func() {
			header := []byte{}
			Expect(compress.Gzip.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-032: should return false for None algorithm", func() {
			header := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
			Expect(compress.None.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-033: should validate Bzip2 version digit", func() {
			validHeaders := [][]byte{
				{'B', 'Z', 'h', '0', 0x00, 0x00},
				{'B', 'Z', 'h', '5', 0x00, 0x00},
				{'B', 'Z', 'h', '9', 0x00, 0x00},
			}
			for _, h := range validHeaders {
				Expect(compress.Bzip2.DetectHeader(h)).To(BeTrue())
			}

			invalidHeaders := [][]byte{
				{'B', 'Z', 'h', 'a', 0x00, 0x00},
				{'B', 'Z', 'h', '/', 0x00, 0x00},
				{'B', 'Z', 'h', ':', 0x00, 0x00},
			}
			for _, h := range invalidHeaders {
				Expect(compress.Bzip2.DetectHeader(h)).To(BeFalse())
			}
		})
	})
})
