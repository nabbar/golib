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
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/archive/compress"
)

var _ = Describe("TC-EN-001: Encoding/Marshaling", func() {
	Context("TC-EN-002: MarshalText", func() {
		It("TC-EN-003: should marshal Gzip", func() {
			data, err := compress.Gzip.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("gzip"))
		})

		It("TC-EN-004: should marshal Bzip2", func() {
			data, err := compress.Bzip2.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("bzip2"))
		})

		It("TC-EN-005: should marshal LZ4", func() {
			data, err := compress.LZ4.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("lz4"))
		})

		It("TC-EN-006: should marshal XZ", func() {
			data, err := compress.XZ.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("xz"))
		})

		It("TC-EN-007: should marshal None", func() {
			data, err := compress.None.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("none"))
		})
	})

	Context("TC-EN-008: UnmarshalText", func() {
		It("TC-EN-009: should unmarshal gzip", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("gzip"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-010: should unmarshal bzip2", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("bzip2"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Bzip2))
		})

		It("TC-EN-011: should unmarshal lz4", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("lz4"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.LZ4))
		})

		It("TC-EN-012: should unmarshal xz", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("xz"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.XZ))
		})

		It("TC-EN-013: should unmarshal none", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("none"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
		})

		It("TC-EN-014: should handle uppercase", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("GZIP"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-015: should trim whitespace", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("  gzip  "))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-016: should trim quotes", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("\"gzip\""))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-017: should trim apostrophes", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("'gzip'"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-018: should default to None for unknown", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte("unknown"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
		})

		It("TC-EN-019: should default to None for empty", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalText([]byte(""))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
		})
	})

	Context("TC-EN-020: MarshalJSON", func() {
		It("TC-EN-021: should marshal Gzip to JSON", func() {
			data, err := compress.Gzip.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"gzip"`))
		})

		It("TC-EN-022: should marshal Bzip2 to JSON", func() {
			data, err := compress.Bzip2.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"bzip2"`))
		})

		It("TC-EN-023: should marshal LZ4 to JSON", func() {
			data, err := compress.LZ4.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"lz4"`))
		})

		It("TC-EN-024: should marshal XZ to JSON", func() {
			data, err := compress.XZ.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"xz"`))
		})

		It("TC-EN-025: should marshal None to null", func() {
			data, err := compress.None.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("null"))
		})

		It("TC-EN-026: should marshal in struct", func() {
			type cfg struct {
				Compression compress.Algorithm `json:"compression"`
			}
			c := cfg{Compression: compress.Gzip}
			data, err := json.Marshal(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`{"compression":"gzip"}`))
		})

		It("TC-EN-027: should marshal None in struct as null", func() {
			type cfg struct {
				Compression compress.Algorithm `json:"compression"`
			}
			c := cfg{Compression: compress.None}
			data, err := json.Marshal(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`{"compression":null}`))
		})
	})

	Context("TC-EN-028: UnmarshalJSON", func() {
		It("TC-EN-029: should unmarshal gzip from JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"gzip"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-030: should unmarshal bzip2 from JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"bzip2"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Bzip2))
		})

		It("TC-EN-031: should unmarshal lz4 from JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"lz4"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.LZ4))
		})

		It("TC-EN-032: should unmarshal xz from JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"xz"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.XZ))
		})

		It("TC-EN-033: should unmarshal null to None", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte("null"))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
		})

		It("TC-EN-034: should unmarshal from struct", func() {
			type cfg struct {
				Compression compress.Algorithm `json:"compression"`
			}
			var c cfg
			err := json.Unmarshal([]byte(`{"compression":"lz4"}`), &c)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Compression).To(Equal(compress.LZ4))
		})

		It("TC-EN-035: should unmarshal null in struct", func() {
			type cfg struct {
				Compression compress.Algorithm `json:"compression"`
			}
			var c cfg
			err := json.Unmarshal([]byte(`{"compression":null}`), &c)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Compression).To(Equal(compress.None))
		})

		It("TC-EN-036: should handle uppercase in JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"GZIP"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.Gzip))
		})

		It("TC-EN-037: should return error for invalid JSON", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`{invalid`))
			Expect(err).To(HaveOccurred())
		})

		It("TC-EN-038: should default to None for unknown value", func() {
			var alg compress.Algorithm
			err := alg.UnmarshalJSON([]byte(`"unknown"`))
			Expect(err).ToNot(HaveOccurred())
			Expect(alg).To(Equal(compress.None))
		})
	})

	Context("TC-EN-039: Round-trip encoding", func() {
		It("TC-EN-040: should round-trip text encoding", func() {
			original := compress.Gzip
			data, err := original.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			var decoded compress.Algorithm
			err = decoded.UnmarshalText(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("TC-EN-041: should round-trip JSON encoding", func() {
			original := compress.Bzip2
			data, err := original.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			var decoded compress.Algorithm
			err = decoded.UnmarshalJSON(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("TC-EN-042: should round-trip all algorithms via text", func() {
			for _, alg := range compress.List() {
				data, err := alg.MarshalText()
				Expect(err).ToNot(HaveOccurred())

				var decoded compress.Algorithm
				err = decoded.UnmarshalText(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(alg))
			}
		})

		It("TC-EN-043: should round-trip all algorithms via JSON", func() {
			for _, alg := range compress.List() {
				data, err := alg.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				var decoded compress.Algorithm
				err = decoded.UnmarshalJSON(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(alg))
			}
		})
	})
})
