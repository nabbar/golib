/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
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

package size_test

import (
	"encoding/json"

	. "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Encoding and Marshalling", func() {
	type TestWrapper struct {
		Size Size `json:"size" yaml:"size"`
	}

	Describe("JSON Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to JSON string", func() {
				w := TestWrapper{Size: 5 * SizeMega}
				b, err := json.Marshal(w)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("size"))
				Expect(string(b)).To(ContainSubstring("MB"))
			})

			It("should marshal zero value", func() {
				w := TestWrapper{Size: SizeNul}
				b, err := json.Marshal(w)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).NotTo(BeEmpty())
			})

			It("should marshal various sizes", func() {
				sizes := []Size{
					SizeUnit,
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					w := TestWrapper{Size: size}
					b, err := json.Marshal(w)
					Expect(err).ToNot(HaveOccurred())
					Expect(b).NotTo(BeEmpty())
				}
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal JSON string to Size", func() {
				w := TestWrapper{Size: 5 * SizeMega}
				b, err := json.Marshal(w)
				Expect(err).ToNot(HaveOccurred())

				var w2 TestWrapper
				err = json.Unmarshal(b, &w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(w2.Size).To(BeNumerically("~", w.Size, float64(w.Size)*0.01))
			})

			It("should unmarshal zero value", func() {
				w := TestWrapper{Size: SizeNul}
				b, err := json.Marshal(w)
				Expect(err).ToNot(HaveOccurred())

				var w2 TestWrapper
				err = json.Unmarshal(b, &w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(w2.Size).To(Equal(SizeNul))
			})

			It("should handle invalid JSON", func() {
				var w TestWrapper
				err := json.Unmarshal([]byte(`{"size":"invalid"}`), &w)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Roundtrip", func() {
			It("should roundtrip various sizes", func() {
				sizes := []Size{
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
					SizeTera,
				}

				for _, size := range sizes {
					w := TestWrapper{Size: size}
					b, err := json.Marshal(w)
					Expect(err).ToNot(HaveOccurred())

					var w2 TestWrapper
					err = json.Unmarshal(b, &w2)
					Expect(err).ToNot(HaveOccurred())
					Expect(w2.Size).To(BeNumerically("~", size, float64(size)*0.05))
				}
			})
		})
	})

	Describe("YAML Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to YAML string", func() {
				w := TestWrapper{Size: 10 * SizeGiga}
				b, err := yaml.Marshal(w)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("size"))
			})

			It("should marshal zero value", func() {
				w := TestWrapper{Size: SizeNul}
				b, err := yaml.Marshal(w)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).NotTo(BeEmpty())
			})

			It("should marshal various sizes", func() {
				sizes := []Size{
					SizeUnit,
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					w := TestWrapper{Size: size}
					b, err := yaml.Marshal(w)
					Expect(err).ToNot(HaveOccurred())
					Expect(b).NotTo(BeEmpty())
				}
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal YAML string to Size", func() {
				w := TestWrapper{Size: 10 * SizeGiga}
				b, err := yaml.Marshal(w)
				Expect(err).ToNot(HaveOccurred())

				var w2 TestWrapper
				err = yaml.Unmarshal(b, &w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(w2.Size).To(BeNumerically("~", w.Size, float64(w.Size)*0.01))
			})

			It("should unmarshal zero value", func() {
				w := TestWrapper{Size: SizeNul}
				b, err := yaml.Marshal(w)
				Expect(err).ToNot(HaveOccurred())

				var w2 TestWrapper
				err = yaml.Unmarshal(b, &w2)
				Expect(err).ToNot(HaveOccurred())
				Expect(w2.Size).To(Equal(SizeNul))
			})

			It("should handle invalid YAML", func() {
				var w TestWrapper
				err := yaml.Unmarshal([]byte("size: invalid"), &w)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Roundtrip", func() {
			It("should roundtrip various sizes", func() {
				sizes := []Size{
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
					SizeTera,
				}

				for _, size := range sizes {
					w := TestWrapper{Size: size}
					b, err := yaml.Marshal(w)
					Expect(err).ToNot(HaveOccurred())

					var w2 TestWrapper
					err = yaml.Unmarshal(b, &w2)
					Expect(err).ToNot(HaveOccurred())
					Expect(w2.Size).To(BeNumerically("~", size, float64(size)*0.05))
				}
			})
		})
	})

	Describe("Text Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to text", func() {
				s := 3 * SizeKilo
				b, err := s.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).NotTo(BeEmpty())
				Expect(string(b)).To(ContainSubstring("KB"))
			})

			It("should marshal zero value", func() {
				s := SizeNul
				b, err := s.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).NotTo(BeEmpty())
			})

			It("should marshal various sizes", func() {
				sizes := []Size{
					SizeUnit,
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					b, err := size.MarshalText()
					Expect(err).ToNot(HaveOccurred())
					Expect(b).NotTo(BeEmpty())
				}
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal text to Size", func() {
				s := 3 * SizeKilo
				b, err := s.MarshalText()
				Expect(err).ToNot(HaveOccurred())

				var s2 Size
				err = s2.UnmarshalText(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(s2).To(BeNumerically("~", s, float64(s)*0.01))
			})

			It("should unmarshal direct text", func() {
				var s Size
				err := s.UnmarshalText([]byte("5MB"))
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
			})

			It("should handle invalid text", func() {
				var s Size
				err := s.UnmarshalText([]byte("invalid"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Roundtrip", func() {
			It("should roundtrip various sizes", func() {
				sizes := []Size{
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					b, err := size.MarshalText()
					Expect(err).ToNot(HaveOccurred())

					var s2 Size
					err = s2.UnmarshalText(b)
					Expect(err).ToNot(HaveOccurred())
					Expect(s2).To(BeNumerically("~", size, float64(size)*0.05))
				}
			})
		})
	})

	Describe("TOML Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to TOML", func() {
				s := 5 * SizeMega
				b, err := s.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).NotTo(BeEmpty())
			})

			It("should marshal zero value", func() {
				s := SizeNul
				b, err := s.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).NotTo(BeEmpty())
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal string", func() {
				var s Size
				err := s.UnmarshalTOML("5MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
			})

			It("should unmarshal bytes", func() {
				var s Size
				err := s.UnmarshalTOML([]byte("10KB"))
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", 10*SizeKilo, float64(10*SizeKilo)*0.01))
			})

			It("should reject invalid types", func() {
				var s Size
				err := s.UnmarshalTOML(123)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not in valid format"))
			})

			It("should reject invalid sizes", func() {
				var s Size
				err := s.UnmarshalTOML("invalid")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CBOR Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to CBOR", func() {
				s := 5 * SizeMega
				b, err := s.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).NotTo(BeEmpty())
			})

			It("should marshal zero value", func() {
				s := SizeNul
				b, err := s.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).NotTo(BeEmpty())
			})

			It("should marshal various sizes", func() {
				sizes := []Size{
					SizeUnit,
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					b, err := size.MarshalCBOR()
					Expect(err).ToNot(HaveOccurred())
					Expect(b).NotTo(BeEmpty())
				}
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal CBOR to Size", func() {
				s := 5 * SizeMega
				b, err := s.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())

				var s2 Size
				err = s2.UnmarshalCBOR(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(s2).To(BeNumerically("~", s, float64(s)*0.01))
			})

			It("should handle invalid CBOR", func() {
				var s Size
				err := s.UnmarshalCBOR([]byte{0xFF, 0xFF})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Roundtrip", func() {
			It("should roundtrip various sizes", func() {
				sizes := []Size{
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					b, err := size.MarshalCBOR()
					Expect(err).ToNot(HaveOccurred())

					var s2 Size
					err = s2.UnmarshalCBOR(b)
					Expect(err).ToNot(HaveOccurred())
					Expect(s2).To(BeNumerically("~", size, float64(size)*0.05))
				}
			})
		})
	})

	Describe("Binary Marshalling", func() {
		Context("Marshal", func() {
			It("should marshal Size to binary", func() {
				s := 5 * SizeMega
				b, err := s.MarshalBinary()
				Expect(err).ToNot(HaveOccurred())
				Expect(b).NotTo(BeEmpty())
			})
		})

		Context("Unmarshal", func() {
			It("should unmarshal binary to Size", func() {
				s := 5 * SizeMega
				b, err := s.MarshalBinary()
				Expect(err).ToNot(HaveOccurred())

				var s2 Size
				err = s2.UnmarshalBinary(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(s2).To(BeNumerically("~", s, float64(s)*0.01))
			})
		})

		Context("Roundtrip", func() {
			It("should roundtrip various sizes", func() {
				sizes := []Size{
					SizeKilo,
					5 * SizeMega,
					10 * SizeGiga,
				}

				for _, size := range sizes {
					b, err := size.MarshalBinary()
					Expect(err).ToNot(HaveOccurred())

					var s2 Size
					err = s2.UnmarshalBinary(b)
					Expect(err).ToNot(HaveOccurred())
					Expect(s2).To(BeNumerically("~", size, float64(size)*0.05))
				}
			})
		})
	})
})
