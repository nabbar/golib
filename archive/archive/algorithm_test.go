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

package archive_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive"
)

var _ = Describe("TC-AL-001: Algorithm Operations", func() {
	Describe("TC-AL-002: String Representation", func() {
		It("TC-AL-003: should return correct string for Tar", func() {
			Expect(archive.Tar.String()).To(Equal("tar"))
		})

		It("TC-AL-004: should return correct string for Zip", func() {
			Expect(archive.Zip.String()).To(Equal("zip"))
		})

		It("TC-AL-005: should return correct string for None", func() {
			Expect(archive.None.String()).To(Equal("none"))
		})
	})

	Describe("TC-AL-006: Extension", func() {
		It("TC-AL-007: should return .tar for Tar algorithm", func() {
			Expect(archive.Tar.Extension()).To(Equal(".tar"))
		})

		It("TC-AL-008: should return .zip for Zip algorithm", func() {
			Expect(archive.Zip.Extension()).To(Equal(".zip"))
		})

		It("TC-AL-009: should return empty string for None algorithm", func() {
			Expect(archive.None.Extension()).To(Equal(""))
		})
	})

	Describe("TC-AL-010: IsNone", func() {
		It("TC-AL-011: should return true for None algorithm", func() {
			Expect(archive.None.IsNone()).To(BeTrue())
		})

		It("TC-AL-012: should return false for Tar algorithm", func() {
			Expect(archive.Tar.IsNone()).To(BeFalse())
		})

		It("TC-AL-013: should return false for Zip algorithm", func() {
			Expect(archive.Zip.IsNone()).To(BeFalse())
		})
	})

	Describe("TC-AL-014: Parse", func() {
		It("TC-AL-015: should parse 'tar' correctly", func() {
			alg := archive.Parse("tar")
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-AL-016: should parse 'TAR' correctly (case insensitive)", func() {
			alg := archive.Parse("TAR")
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-AL-017: should parse 'zip' correctly", func() {
			alg := archive.Parse("zip")
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-AL-018: should parse 'ZIP' correctly (case insensitive)", func() {
			alg := archive.Parse("ZIP")
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-AL-019: should return None for unknown format", func() {
			alg := archive.Parse("unknown")
			Expect(alg).To(Equal(archive.None))
		})

		It("TC-AL-020: should return None for empty string", func() {
			alg := archive.Parse("")
			Expect(alg).To(Equal(archive.None))
		})
	})

	Describe("TC-AL-021: DetectHeader", func() {
		It("TC-AL-022: should detect TAR header correctly", func() {
			header := make([]byte, 265)
			copy(header[257:], []byte("ustar\x00"))
			Expect(archive.Tar.DetectHeader(header)).To(BeTrue())
		})

		It("TC-AL-023: should detect ZIP header correctly", func() {
			header := make([]byte, 265)
			copy(header[0:], []byte{0x50, 0x4b, 0x03, 0x04})
			Expect(archive.Zip.DetectHeader(header)).To(BeTrue())
		})

		It("TC-AL-024: should return false for None algorithm", func() {
			header := make([]byte, 265)
			Expect(archive.None.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-025: should return false for invalid TAR header", func() {
			header := make([]byte, 265)
			copy(header[257:], []byte("wrong\x00"))
			Expect(archive.Tar.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-026: should return false for invalid ZIP header", func() {
			header := make([]byte, 265)
			copy(header[0:], []byte{0x00, 0x00, 0x00, 0x00})
			Expect(archive.Zip.DetectHeader(header)).To(BeFalse())
		})

		It("TC-AL-027: should return false for truncated header", func() {
			header := make([]byte, 10)
			Expect(archive.Tar.DetectHeader(header)).To(BeFalse())
			Expect(archive.Zip.DetectHeader(header)).To(BeFalse())
		})
	})
})
