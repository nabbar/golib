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
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Permission Encoding", func() {
	Describe("JSON", func() {
		It("should marshal to JSON", func() {
			perm := Perm(0644)
			data, err := json.Marshal(perm)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("\"0644\""))
		})

		It("should unmarshal from JSON", func() {
			data := []byte("\"0644\"")
			var perm Perm
			err := json.Unmarshal(data, &perm)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should round-trip through JSON", func() {
			original := Perm(0755)
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = json.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should handle permission with special bits in JSON", func() {
			original := Perm(04755)
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = json.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should return error for invalid JSON", func() {
			data := []byte("\"invalid\"")
			var perm Perm
			err := json.Unmarshal(data, &perm)
			Expect(err).To(HaveOccurred())
		})

		It("should marshal struct with Perm field to JSON", func() {
			type Config struct {
				Mode Perm `json:"mode"`
			}
			cfg := Config{Mode: Perm(0644)}
			data, err := json.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("\"mode\":\"0644\""))
		})

		It("should unmarshal struct with Perm field from JSON", func() {
			type Config struct {
				Mode Perm `json:"mode"`
			}
			data := []byte(`{"mode":"0755"}`)
			var cfg Config
			err := json.Unmarshal(data, &cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Mode.Uint64()).To(Equal(uint64(0755)))
		})
	})

	Describe("YAML", func() {
		It("should marshal to YAML", func() {
			perm := Perm(0644)
			data, err := yaml.Marshal(perm)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("0644"))
		})

		It("should unmarshal from YAML", func() {
			data := []byte("0644")
			var perm Perm
			err := yaml.Unmarshal(data, &perm)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should round-trip through YAML", func() {
			original := Perm(0755)
			data, err := yaml.Marshal(original)
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = yaml.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should handle permission with special bits in YAML", func() {
			original := Perm(04755)
			data, err := yaml.Marshal(original)
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = yaml.Unmarshal(data, &decoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should return error for invalid YAML", func() {
			data := []byte("invalid_perm")
			var perm Perm
			err := yaml.Unmarshal(data, &perm)
			Expect(err).To(HaveOccurred())
		})

		It("should marshal struct with Perm field to YAML", func() {
			type Config struct {
				Mode Perm `yaml:"mode"`
			}
			cfg := Config{Mode: Perm(0644)}
			data, err := yaml.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("mode"))
			Expect(string(data)).To(ContainSubstring("0644"))
		})

		It("should unmarshal struct with Perm field from YAML", func() {
			type Config struct {
				Mode Perm `yaml:"mode"`
			}
			data := []byte("mode: 0755")
			var cfg Config
			err := yaml.Unmarshal(data, &cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Mode.Uint64()).To(Equal(uint64(0755)))
		})
	})

	Describe("Text", func() {
		It("should marshal to text", func() {
			perm := Perm(0644)
			data, err := perm.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("0644"))
		})

		It("should unmarshal from text", func() {
			data := []byte("0644")
			var perm Perm
			err := perm.UnmarshalText(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should round-trip through text", func() {
			original := Perm(0755)
			data, err := original.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = decoded.UnmarshalText(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should return error for invalid text", func() {
			data := []byte("invalid")
			var perm Perm
			err := perm.UnmarshalText(data)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TOML", func() {
		It("should marshal to TOML", func() {
			perm := Perm(0644)
			data, err := perm.MarshalTOML()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("\"0644\""))
		})

		It("should unmarshal from TOML byte slice", func() {
			data := []byte("\"0644\"")
			var perm Perm
			err := perm.UnmarshalTOML(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should unmarshal from TOML string", func() {
			var perm Perm
			err := perm.UnmarshalTOML("0755")
			Expect(err).ToNot(HaveOccurred())
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should return error for invalid TOML type", func() {
			var perm Perm
			err := perm.UnmarshalTOML(123)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not in valid format"))
		})

		It("should round-trip through TOML", func() {
			original := Perm(0755)
			data, err := original.MarshalTOML()
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = decoded.UnmarshalTOML(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})
	})

	Describe("CBOR", func() {
		It("should marshal to CBOR", func() {
			perm := Perm(0644)
			data, err := perm.MarshalCBOR()
			Expect(err).ToNot(HaveOccurred())
			Expect(data).ToNot(BeEmpty())
		})

		It("should unmarshal from CBOR", func() {
			original := Perm(0644)
			data, err := cbor.Marshal("0644")
			Expect(err).ToNot(HaveOccurred())

			var perm Perm
			err = perm.UnmarshalCBOR(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(perm).To(Equal(original))
		})

		It("should round-trip through CBOR", func() {
			original := Perm(0755)
			data, err := original.MarshalCBOR()
			Expect(err).ToNot(HaveOccurred())

			var decoded Perm
			err = decoded.UnmarshalCBOR(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(original))
		})

		It("should return error for invalid CBOR", func() {
			data := []byte{0xFF, 0xFF} // Invalid CBOR
			var perm Perm
			err := perm.UnmarshalCBOR(data)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for CBOR with invalid permission value", func() {
			data, err := cbor.Marshal("invalid")
			Expect(err).ToNot(HaveOccurred())

			var perm Perm
			err = perm.UnmarshalCBOR(data)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Cross-format Compatibility", func() {
		It("should maintain value across different formats", func() {
			original := Perm(0644)

			// JSON
			jsonData, _ := json.Marshal(original)
			var fromJSON Perm
			json.Unmarshal(jsonData, &fromJSON)

			// YAML
			yamlData, _ := yaml.Marshal(original)
			var fromYAML Perm
			yaml.Unmarshal(yamlData, &fromYAML)

			// Text
			textData, _ := original.MarshalText()
			var fromText Perm
			fromText.UnmarshalText(textData)

			// CBOR
			cborData, _ := original.MarshalCBOR()
			var fromCBOR Perm
			fromCBOR.UnmarshalCBOR(cborData)

			// All should be equal
			Expect(fromJSON).To(Equal(original))
			Expect(fromYAML).To(Equal(original))
			Expect(fromText).To(Equal(original))
			Expect(fromCBOR).To(Equal(original))
		})
	})
})
