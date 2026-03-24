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

package big_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fxamacker/cbor/v2"
	durbig "github.com/nabbar/golib/duration/big"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Big Duration Encoding Operations", func() {
	Describe("JSON Marshaling", func() {
		It("should marshal duration to JSON", func() {
			// TC-EN-001
			d := durbig.Hours(5) + durbig.Minutes(30)
			data, err := json.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"5h30m"`))
		})

		It("should marshal zero duration", func() {
			// TC-EN-002
			d := durbig.Seconds(0)
			data, err := json.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"0s"`))
		})

		It("should marshal negative duration", func() {
			// TC-EN-003
			d := durbig.Seconds(-30)
			data, err := json.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("-"))
		})

		It("should marshal duration with days", func() {
			// TC-EN-004
			d := durbig.Days(2) + durbig.Hours(3)
			data, err := json.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"2d3h"`))
		})

		It("should marshal large duration", func() {
			// TC-EN-005
			d := durbig.Days(1000)
			data, err := json.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("1000d"))
		})
	})

	Describe("JSON Unmarshaling", func() {
		It("should unmarshal valid JSON", func() {
			// TC-EN-006
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"5h30m"`), &d)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should unmarshal zero duration", func() {
			// TC-EN-007
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"0s"`), &d)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Seconds(0)))
		})

		It("should unmarshal duration with days", func() {
			// TC-EN-008
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"3d12h"`), &d)

			Expect(err).ToNot(HaveOccurred())
			expected := durbig.Days(3) + durbig.Hours(12)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})

		It("should return error for invalid JSON duration", func() {
			// TC-EN-009
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"invalid"`), &d)

			Expect(err).To(HaveOccurred())
		})

		It("should handle quoted strings with spaces", func() {
			// TC-EN-010
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"5h 30m"`), &d)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should handle negative durations", func() {
			// TC-EN-011
			var d durbig.Duration
			err := json.Unmarshal([]byte(`"-5h30m"`), &d)

			Expect(err).ToNot(HaveOccurred())
			expected := -(durbig.Hours(5) + durbig.Minutes(30))
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("YAML Marshaling", func() {
		It("should marshal duration to YAML", func() {
			// TC-EN-012
			d := durbig.Hours(5) + durbig.Minutes(30)
			data, err := yaml.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("5h30m\n"))
		})

		It("should marshal zero duration", func() {
			// TC-EN-013
			d := durbig.Seconds(0)
			data, err := yaml.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("0s\n"))
		})

		It("should marshal duration with days", func() {
			// TC-EN-014
			d := durbig.Days(1) + durbig.Hours(12)
			data, err := yaml.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("1d12h\n"))
		})

		It("should marshal negative duration", func() {
			// TC-EN-015
			d := durbig.Seconds(-45)
			data, err := yaml.Marshal(d)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("-"))
		})
	})

	Describe("YAML Unmarshaling", func() {
		It("should unmarshal valid YAML", func() {
			// TC-EN-016
			var d durbig.Duration
			err := yaml.Unmarshal([]byte("5h30m"), &d)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should return error for invalid YAML duration", func() {
			// TC-EN-017
			var d durbig.Duration
			err := yaml.Unmarshal([]byte("invalid"), &d)

			Expect(err).To(HaveOccurred())
		})

		It("should unmarshal duration with days", func() {
			// TC-EN-018
			var d durbig.Duration
			err := yaml.Unmarshal([]byte("7d12h"), &d)

			Expect(err).ToNot(HaveOccurred())
			expected := durbig.Days(7) + durbig.Hours(12)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})

		It("should handle zero duration", func() {
			// TC-EN-019
			var d durbig.Duration
			err := yaml.Unmarshal([]byte("0s"), &d)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Seconds(0)))
		})
	})

	Describe("TOML Marshaling", func() {
		It("should marshal duration to TOML", func() {
			// TC-EN-020
			d := durbig.Hours(5) + durbig.Minutes(30)
			data, err := d.MarshalTOML()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"5h30m"`))
		})

		It("should marshal duration with days", func() {
			// TC-EN-021
			d := durbig.Days(2) + durbig.Hours(6)
			data, err := d.MarshalTOML()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"2d6h"`))
		})

		It("should marshal zero duration", func() {
			// TC-EN-022
			d := durbig.Seconds(0)
			data, err := d.MarshalTOML()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"0s"`))
		})
	})

	Describe("TOML Unmarshaling", func() {
		It("should unmarshal TOML string", func() {
			// TC-EN-023
			var d durbig.Duration
			err := d.UnmarshalTOML("5h30m")

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should unmarshal TOML byte array", func() {
			// TC-EN-024
			var d durbig.Duration
			err := d.UnmarshalTOML([]byte(`"5h30m"`))

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should return error for invalid TOML format", func() {
			// TC-EN-025
			var d durbig.Duration
			err := d.UnmarshalTOML(12345)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not in valid format"))
		})

		It("should return error for invalid duration string", func() {
			// TC-EN-026
			var d durbig.Duration
			err := d.UnmarshalTOML("invalid")

			Expect(err).To(HaveOccurred())
		})

		It("should handle quoted strings", func() {
			// TC-EN-027
			var d durbig.Duration
			err := d.UnmarshalTOML([]byte(`"3d12h"`))

			Expect(err).ToNot(HaveOccurred())
			expected := durbig.Days(3) + durbig.Hours(12)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("Text Marshaling", func() {
		It("should marshal duration to text", func() {
			// TC-EN-028
			d := durbig.Hours(5) + durbig.Minutes(30)
			data, err := d.MarshalText()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("5h30m"))
		})

		It("should marshal zero duration", func() {
			// TC-EN-029
			d := durbig.Seconds(0)
			data, err := d.MarshalText()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("0s"))
		})

		It("should marshal negative duration", func() {
			// TC-EN-030
			d := durbig.Seconds(-45)
			data, err := d.MarshalText()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("-"))
		})

		It("should marshal duration with days", func() {
			// TC-EN-031
			d := durbig.Days(7)
			data, err := d.MarshalText()

			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("7d"))
		})
	})

	Describe("Text Unmarshaling", func() {
		It("should unmarshal valid text", func() {
			// TC-EN-032
			var d durbig.Duration
			err := d.UnmarshalText([]byte("5h30m"))

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should return error for invalid text", func() {
			// TC-EN-033
			var d durbig.Duration
			err := d.UnmarshalText([]byte("invalid"))

			Expect(err).To(HaveOccurred())
		})

		It("should handle duration with days", func() {
			// TC-EN-034
			var d durbig.Duration
			err := d.UnmarshalText([]byte("10d"))

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Days(10)))
		})

		It("should handle empty text", func() {
			// TC-EN-035
			var d durbig.Duration
			err := d.UnmarshalText([]byte(""))

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CBOR Marshaling", func() {
		It("should marshal duration to CBOR", func() {
			// TC-EN-036
			d := durbig.Hours(5) + durbig.Minutes(30)
			data, err := d.MarshalCBOR()

			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())

			// Verify it can be decoded
			var str string
			err = cbor.Unmarshal(data, &str)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal("5h30m"))
		})

		It("should marshal duration with days", func() {
			// TC-EN-037
			d := durbig.Days(3) + durbig.Hours(6)
			data, err := d.MarshalCBOR()

			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())
		})

		It("should marshal zero duration", func() {
			// TC-EN-038
			d := durbig.Seconds(0)
			data, err := d.MarshalCBOR()

			Expect(err).ToNot(HaveOccurred())

			var str string
			err = cbor.Unmarshal(data, &str)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal("0s"))
		})
	})

	Describe("CBOR Unmarshaling", func() {
		It("should unmarshal valid CBOR", func() {
			// TC-EN-039
			// Create CBOR encoded duration string
			cborData, err := cbor.Marshal("5h30m")
			Expect(err).ToNot(HaveOccurred())

			var d durbig.Duration
			err = d.UnmarshalCBOR(cborData)

			Expect(err).ToNot(HaveOccurred())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should return error for invalid CBOR data", func() {
			// TC-EN-040
			var d durbig.Duration
			err := d.UnmarshalCBOR([]byte{0xFF, 0xFF})

			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid duration in CBOR", func() {
			// TC-EN-041
			cborData, err := cbor.Marshal("invalid")
			Expect(err).ToNot(HaveOccurred())

			var d durbig.Duration
			err = d.UnmarshalCBOR(cborData)

			Expect(err).To(HaveOccurred())
		})

		It("should handle duration with days", func() {
			// TC-EN-042
			cborData, err := cbor.Marshal("7d12h")
			Expect(err).ToNot(HaveOccurred())

			var d durbig.Duration
			err = d.UnmarshalCBOR(cborData)

			Expect(err).ToNot(HaveOccurred())
			expected := durbig.Days(7) + durbig.Hours(12)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("Round-trip Encoding", func() {
		testDurations := []durbig.Duration{
			durbig.Seconds(0),
			durbig.Seconds(30),
			durbig.Minutes(5) + durbig.Seconds(30),
			durbig.Hours(2) + durbig.Minutes(30),
			durbig.Days(1) + durbig.Hours(12) + durbig.Minutes(30),
			durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13),
		}

		It("should handle JSON round-trip", func() {
			// TC-EN-043
			for _, original := range testDurations {
				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var decoded durbig.Duration
				err = json.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original), "for duration %s", original.String())
			}
		})

		It("should handle YAML round-trip", func() {
			// TC-EN-044
			for _, original := range testDurations {
				data, err := yaml.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var decoded durbig.Duration
				err = yaml.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original), "for duration %s", original.String())
			}
		})

		It("should handle TOML round-trip", func() {
			// TC-EN-045
			// TOML requires a struct, not a standalone value
			type TestStruct struct {
				Duration durbig.Duration `toml:"duration"`
			}

			for _, original := range testDurations {
				obj := TestStruct{Duration: original}
				data, err := toml.Marshal(obj)
				Expect(err).ToNot(HaveOccurred())

				var decoded TestStruct
				err = toml.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded.Duration).To(Equal(original), "for duration %s", original.String())
			}
		})

		It("should handle Text round-trip", func() {
			// TC-EN-046
			for _, original := range testDurations {
				data, err := original.MarshalText()
				Expect(err).ToNot(HaveOccurred())

				var decoded durbig.Duration
				err = decoded.UnmarshalText(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original), "for duration %s", original.String())
			}
		})

		It("should handle CBOR round-trip", func() {
			// TC-EN-047
			for _, original := range testDurations {
				data, err := original.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())

				var decoded durbig.Duration
				err = decoded.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original), "for duration %s", original.String())
			}
		})
	})
})
