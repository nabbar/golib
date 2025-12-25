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

var _ = Describe("TC-PR-001: Parse Function", func() {
	Context("TC-PR-002: Valid algorithm strings", func() {
		It("TC-PR-003: should parse gzip", func() {
			Expect(compress.Parse("gzip")).To(Equal(compress.Gzip))
		})

		It("TC-PR-004: should parse bzip2", func() {
			Expect(compress.Parse("bzip2")).To(Equal(compress.Bzip2))
		})

		It("TC-PR-005: should parse lz4", func() {
			Expect(compress.Parse("lz4")).To(Equal(compress.LZ4))
		})

		It("TC-PR-006: should parse xz", func() {
			Expect(compress.Parse("xz")).To(Equal(compress.XZ))
		})

		It("TC-PR-007: should parse none", func() {
			Expect(compress.Parse("none")).To(Equal(compress.None))
		})
	})

	Context("TC-PR-008: Case insensitivity", func() {
		It("TC-PR-009: should parse uppercase GZIP", func() {
			Expect(compress.Parse("GZIP")).To(Equal(compress.Gzip))
		})

		It("TC-PR-010: should parse mixed case GzIp", func() {
			Expect(compress.Parse("GzIp")).To(Equal(compress.Gzip))
		})

		It("TC-PR-011: should parse uppercase BZIP2", func() {
			Expect(compress.Parse("BZIP2")).To(Equal(compress.Bzip2))
		})

		It("TC-PR-012: should parse uppercase LZ4", func() {
			Expect(compress.Parse("LZ4")).To(Equal(compress.LZ4))
		})

		It("TC-PR-013: should parse uppercase XZ", func() {
			Expect(compress.Parse("XZ")).To(Equal(compress.XZ))
		})
	})

	Context("TC-PR-014: Whitespace handling", func() {
		It("TC-PR-015: should trim leading whitespace", func() {
			Expect(compress.Parse("  gzip")).To(Equal(compress.Gzip))
		})

		It("TC-PR-016: should trim trailing whitespace", func() {
			Expect(compress.Parse("gzip  ")).To(Equal(compress.Gzip))
		})

		It("TC-PR-017: should trim both sides", func() {
			Expect(compress.Parse("  gzip  ")).To(Equal(compress.Gzip))
		})

		It("TC-PR-018: should trim tabs", func() {
			Expect(compress.Parse("\tgzip\t")).To(Equal(compress.Gzip))
		})

		It("TC-PR-019: should trim newlines", func() {
			Expect(compress.Parse("\ngzip\n")).To(Equal(compress.Gzip))
		})
	})

	Context("TC-PR-020: Quote handling", func() {
		It("TC-PR-021: should trim double quotes", func() {
			Expect(compress.Parse("\"gzip\"")).To(Equal(compress.Gzip))
		})

		It("TC-PR-022: should trim single quotes", func() {
			Expect(compress.Parse("'gzip'")).To(Equal(compress.Gzip))
		})

		It("TC-PR-023: should handle quotes with whitespace", func() {
			Expect(compress.Parse("  \"gzip\"  ")).To(Equal(compress.Gzip))
		})
	})

	Context("TC-PR-024: Invalid inputs", func() {
		It("TC-PR-025: should return None for unknown algorithm", func() {
			Expect(compress.Parse("unknown")).To(Equal(compress.None))
		})

		It("TC-PR-026: should return None for empty string", func() {
			Expect(compress.Parse("")).To(Equal(compress.None))
		})

		It("TC-PR-027: should return None for whitespace only", func() {
			Expect(compress.Parse("   ")).To(Equal(compress.None))
		})

		It("TC-PR-028: should return None for invalid format", func() {
			Expect(compress.Parse("gz ip")).To(Equal(compress.None))
		})

		It("TC-PR-029: should return None for numbers", func() {
			Expect(compress.Parse("123")).To(Equal(compress.None))
		})

		It("TC-PR-030: should return None for special characters", func() {
			Expect(compress.Parse("@#$%")).To(Equal(compress.None))
		})
	})

	Context("TC-PR-031: Edge cases", func() {
		It("TC-PR-032: should handle very long string", func() {
			longStr := "gzip" + string(make([]byte, 1000))
			Expect(compress.Parse(longStr)).To(Equal(compress.None))
		})

		It("TC-PR-033: should handle substring match", func() {
			Expect(compress.Parse("gzipextra")).To(Equal(compress.None))
		})

		It("TC-PR-034: should handle prefix only", func() {
			Expect(compress.Parse("gz")).To(Equal(compress.None))
		})
	})
})
