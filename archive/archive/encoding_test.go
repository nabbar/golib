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
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/archive"
)

var _ = Describe("TC-EN-001: Encoding Operations", func() {
	Describe("TC-EN-002: MarshalText", func() {
		It("TC-EN-003: should marshal Tar algorithm to text", func() {
			text, err := archive.Tar.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(text)).To(Equal("tar"))
		})

		It("TC-EN-004: should marshal Zip algorithm to text", func() {
			text, err := archive.Zip.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(text)).To(Equal("zip"))
		})

		It("TC-EN-005: should marshal None algorithm to text", func() {
			text, err := archive.None.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(text)).To(Equal("none"))
		})
	})

	Describe("TC-EN-006: UnmarshalText", func() {
		It("TC-EN-007: should unmarshal 'tar' correctly", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("tar"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-EN-008: should unmarshal 'TAR' correctly (case insensitive)", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("TAR"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-EN-009: should unmarshal 'zip' correctly", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("zip"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-EN-010: should unmarshal 'ZIP' correctly (case insensitive)", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("ZIP"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-EN-011: should unmarshal unknown format to None", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("unknown"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.None))
		})

		It("TC-EN-012: should handle whitespace trimming", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("  tar  "))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-EN-013: should handle quoted strings", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte(`"zip"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-EN-014: should handle single-quoted strings", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalText([]byte("'tar'"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
		})
	})

	Describe("TC-EN-015: MarshalJSON", func() {
		It("TC-EN-016: should marshal Tar algorithm to JSON", func() {
			data, err := archive.Tar.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"tar"`))
		})

		It("TC-EN-017: should marshal Zip algorithm to JSON", func() {
			data, err := archive.Zip.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"zip"`))
		})

		It("TC-EN-018: should marshal None algorithm to JSON null", func() {
			data, err := archive.None.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("null"))
		})

		It("TC-EN-019: should marshal in struct correctly", func() {
			type Config struct {
				Format archive.Algorithm `json:"format"`
			}
			cfg := Config{Format: archive.Tar}
			data, err := json.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`{"format":"tar"}`))
		})
	})

	Describe("TC-EN-020: UnmarshalJSON", func() {
		It("TC-EN-021: should unmarshal 'tar' from JSON", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalJSON([]byte(`"tar"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-EN-022: should unmarshal 'zip' from JSON", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalJSON([]byte(`"zip"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-EN-023: should unmarshal null to None", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalJSON([]byte("null"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.None))
		})

		It("TC-EN-024: should unmarshal unknown format to None", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalJSON([]byte(`"unknown"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(archive.None))
		})

		It("TC-EN-025: should return error for invalid JSON", func() {
			var alg archive.Algorithm
			err := alg.UnmarshalJSON([]byte(`invalid`))
			Expect(err).To(HaveOccurred())
		})

		It("TC-EN-026: should unmarshal in struct correctly", func() {
			type Config struct {
				Format archive.Algorithm `json:"format"`
			}
			var cfg Config
			err := json.Unmarshal([]byte(`{"format":"zip"}`), &cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Format).To(Equal(archive.Zip))
		})
	})

	Describe("TC-EN-027: Round-trip Encoding", func() {
		It("TC-EN-028: should round-trip Tar through text encoding", func() {
			text, _ := archive.Tar.MarshalText()
			var alg archive.Algorithm
			_ = alg.UnmarshalText(text)
			Expect(alg).To(Equal(archive.Tar))
		})

		It("TC-EN-029: should round-trip Zip through JSON encoding", func() {
			data, _ := archive.Zip.MarshalJSON()
			var alg archive.Algorithm
			_ = alg.UnmarshalJSON(data)
			Expect(alg).To(Equal(archive.Zip))
		})

		It("TC-EN-030: should round-trip None through JSON encoding", func() {
			data, _ := archive.None.MarshalJSON()
			var alg archive.Algorithm
			_ = alg.UnmarshalJSON(data)
			Expect(alg).To(Equal(archive.None))
		})
	})
})
