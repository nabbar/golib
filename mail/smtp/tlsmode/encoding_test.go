/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package tlsmode_test

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/mail/smtp/tlsmode"
	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS Mode Encoding", func() {

	Describe("JSON Marshaling", func() {
		Context("Marshal to JSON", func() {
			It("should marshal TLSNone to empty string", func() {
				data, err := json.Marshal(TLSNone)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`""`))
			})

			It("should marshal TLSStartTLS to 'starttls'", func() {
				data, err := json.Marshal(TLSStartTLS)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"starttls"`))
			})

			It("should marshal TLSStrictTLS to 'tls'", func() {
				data, err := json.Marshal(TLSStrictTLS)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"tls"`))
			})
		})

		Context("Unmarshal from JSON", func() {
			It("should unmarshal string values", func() {
				var mode TLSMode
				err := json.Unmarshal([]byte(`"starttls"`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = json.Unmarshal([]byte(`"tls"`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))

				err = json.Unmarshal([]byte(`""`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSNone))
			})

			It("should unmarshal numeric values", func() {
				var mode TLSMode
				err := json.Unmarshal([]byte(`0`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSNone))

				err = json.Unmarshal([]byte(`1`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = json.Unmarshal([]byte(`2`), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))
			})

			It("should handle invalid JSON", func() {
				var mode TLSMode
				err := json.Unmarshal([]byte(`true`), &mode)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("JSON roundtrip", func() {
			It("should preserve values through marshal/unmarshal", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, original := range modes {
					data, err := json.Marshal(original)
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = json.Unmarshal(data, &decoded)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(original))
				}
			})
		})
	})

	Describe("YAML Marshaling", func() {
		Context("Marshal to YAML", func() {
			It("should marshal TLSNone to empty string", func() {
				data, err := yaml.Marshal(TLSNone)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring(`""`))
			})

			It("should marshal TLSStartTLS to 'starttls'", func() {
				data, err := yaml.Marshal(TLSStartTLS)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("starttls"))
			})

			It("should marshal TLSStrictTLS to 'tls'", func() {
				data, err := yaml.Marshal(TLSStrictTLS)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(ContainSubstring("tls"))
			})
		})

		Context("Unmarshal from YAML", func() {
			It("should unmarshal string values", func() {
				var mode TLSMode
				err := yaml.Unmarshal([]byte("starttls"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = yaml.Unmarshal([]byte("tls"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))
			})

			It("should unmarshal numeric values", func() {
				var mode TLSMode
				err := yaml.Unmarshal([]byte("0"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSNone))

				err = yaml.Unmarshal([]byte("1"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = yaml.Unmarshal([]byte("2"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))
			})
		})

		Context("YAML roundtrip", func() {
			It("should preserve values through marshal/unmarshal", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, original := range modes {
					data, err := yaml.Marshal(original)
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = yaml.Unmarshal(data, &decoded)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(original))
				}
			})
		})
	})

	Describe("TOML Marshaling", func() {
		Context("Marshal to TOML", func() {
			It("should marshal all modes", func() {
				data, err := TLSNone.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStartTLS.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStrictTLS.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())
			})
		})

		Context("Unmarshal from TOML", func() {
			It("should unmarshal string values", func() {
				var mode TLSMode
				err := mode.UnmarshalTOML("starttls")
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = mode.UnmarshalTOML("tls")
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))
			})

			It("should unmarshal byte values", func() {
				var mode TLSMode
				err := mode.UnmarshalTOML([]byte("starttls"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))
			})

			It("should unmarshal numeric values", func() {
				var mode TLSMode
				err := mode.UnmarshalTOML(uint64(1))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = mode.UnmarshalTOML(int64(2))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))

				err = mode.UnmarshalTOML(float64(1.0))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))
			})

			It("should return error for invalid types", func() {
				var mode TLSMode
				err := mode.UnmarshalTOML(true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not in valid format"))
			})
		})
	})

	Describe("Text Marshaling", func() {
		Context("Marshal to text", func() {
			It("should marshal all modes to text", func() {
				data, err := TLSNone.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(""))

				data, err = TLSStartTLS.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("starttls"))

				data, err = TLSStrictTLS.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("tls"))
			})
		})

		Context("Unmarshal from text", func() {
			It("should unmarshal text values", func() {
				var mode TLSMode
				err := mode.UnmarshalText([]byte("starttls"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))

				err = mode.UnmarshalText([]byte("tls"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStrictTLS))

				err = mode.UnmarshalText([]byte(""))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSNone))
			})
		})

		Context("Text roundtrip", func() {
			It("should preserve values through marshal/unmarshal", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, original := range modes {
					data, err := original.MarshalText()
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = decoded.UnmarshalText(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(original))
				}
			})
		})
	})

	Describe("CBOR Marshaling", func() {
		Context("Marshal to CBOR", func() {
			It("should marshal all modes", func() {
				data, err := TLSNone.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStartTLS.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStrictTLS.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())
			})
		})

		Context("Unmarshal from CBOR", func() {
			It("should unmarshal string CBOR values", func() {
				// Manually create CBOR-encoded strings
				data, _ := cbor.Marshal("starttls")
				var mode TLSMode
				err := mode.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))
			})

			It("should unmarshal numeric CBOR values", func() {
				// Manually create CBOR-encoded integers
				data, _ := cbor.Marshal(int64(1))
				var mode TLSMode
				err := mode.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))
			})

			It("should return error for invalid CBOR", func() {
				var mode TLSMode
				err := mode.UnmarshalCBOR([]byte{0xFF, 0xFF})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("CBOR roundtrip", func() {
			It("should preserve values through marshal/unmarshal", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, original := range modes {
					data, err := original.MarshalCBOR()
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = decoded.UnmarshalCBOR(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(original))
				}
			})
		})
	})

	Describe("Binary Marshaling", func() {
		Context("Marshal to binary", func() {
			It("should marshal all modes", func() {
				data, err := TLSNone.MarshalBinary()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStartTLS.MarshalBinary()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())

				data, err = TLSStrictTLS.MarshalBinary()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeNil())
			})
		})

		Context("Unmarshal from binary", func() {
			It("should unmarshal binary values", func() {
				// MarshalBinary uses CBOR internally
				data, _ := cbor.Marshal("starttls")
				var mode TLSMode
				err := mode.UnmarshalBinary(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(TLSStartTLS))
			})
		})

		Context("Binary roundtrip", func() {
			It("should preserve values through marshal/unmarshal", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, original := range modes {
					data, err := original.MarshalBinary()
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = decoded.UnmarshalBinary(data)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(original))
				}
			})
		})
	})

	Describe("Complex Struct Encoding", func() {
		type Config struct {
			Mode TLSMode `json:"mode" yaml:"mode"`
			Name string  `json:"name" yaml:"name"`
		}

		It("should handle JSON encoding in structs", func() {
			cfg := Config{Mode: TLSStartTLS, Name: "test"}
			data, err := json.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())

			var decoded Config
			err = json.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Mode).To(Equal(TLSStartTLS))
			Expect(decoded.Name).To(Equal("test"))
		})

		It("should handle YAML encoding in structs", func() {
			cfg := Config{Mode: TLSStrictTLS, Name: "test"}
			data, err := yaml.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())

			var decoded Config
			err = yaml.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Mode).To(Equal(TLSStrictTLS))
			Expect(decoded.Name).To(Equal("test"))
		})
	})
})
