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

package control_test

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/nabbar/golib/status/control"
)

var _ = Describe("Mode/Encoding", func() {
	Describe("JSON", func() {
		Context("MarshalJSON", func() {
			It("should marshal Should to JSON", func() {
				data, err := json.Marshal(control.Should)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"Should"`))
			})

			It("should marshal Must to JSON", func() {
				data, err := json.Marshal(control.Must)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"Must"`))
			})

			It("should marshal AnyOf to JSON", func() {
				data, err := json.Marshal(control.AnyOf)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"AnyOf"`))
			})

			It("should marshal Quorum to JSON", func() {
				data, err := json.Marshal(control.Quorum)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"Quorum"`))
			})

			It("should marshal Ignore to JSON", func() {
				data, err := json.Marshal(control.Ignore)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`""`))
			})
		})

		Context("UnmarshalJSON", func() {
			It("should unmarshal 'Should' from JSON", func() {
				var mode control.Mode
				// JSON unmarshaling includes quotes in the bytes
				err := mode.UnmarshalJSON([]byte(`"Should"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'should' from JSON", func() {
				var mode control.Mode
				err := mode.UnmarshalJSON([]byte(`"should"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'Must' from JSON", func() {
				var mode control.Mode
				err := mode.UnmarshalJSON([]byte(`"Must"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Must))
			})

			It("should unmarshal 'AnyOf' from JSON", func() {
				var mode control.Mode
				err := mode.UnmarshalJSON([]byte(`"AnyOf"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.AnyOf))
			})

			It("should unmarshal 'Quorum' from JSON", func() {
				var mode control.Mode
				err := mode.UnmarshalJSON([]byte(`"Quorum"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Quorum))
			})

			It("should unmarshal invalid value to Ignore", func() {
				var mode control.Mode
				err := mode.UnmarshalJSON([]byte(`"invalid"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Ignore))
			})
		})

		Context("Round-trip", func() {
			It("should marshal and unmarshal Should", func() {
				original := control.Should
				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var decoded control.Mode
				err = json.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			})

			It("should marshal and unmarshal Must", func() {
				original := control.Must
				data, err := json.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var decoded control.Mode
				err = json.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			})
		})
	})

	Describe("YAML", func() {
		Context("MarshalYAML", func() {
			It("should marshal Should to YAML", func() {
				data, err := yaml.Marshal(control.Should)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Should\n"))
			})

			It("should marshal Must to YAML", func() {
				data, err := yaml.Marshal(control.Must)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Must\n"))
			})

			It("should marshal AnyOf to YAML", func() {
				data, err := yaml.Marshal(control.AnyOf)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("AnyOf\n"))
			})

			It("should marshal Quorum to YAML", func() {
				data, err := yaml.Marshal(control.Quorum)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Quorum\n"))
			})
		})

		Context("UnmarshalYAML", func() {
			It("should unmarshal 'Should' from YAML", func() {
				var mode control.Mode
				err := yaml.Unmarshal([]byte("Should"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'must' from YAML", func() {
				var mode control.Mode
				err := yaml.Unmarshal([]byte("must"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Must))
			})

			It("should unmarshal 'AnyOf' from YAML", func() {
				var mode control.Mode
				err := yaml.Unmarshal([]byte("AnyOf"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.AnyOf))
			})

			It("should unmarshal 'Quorum' from YAML", func() {
				var mode control.Mode
				err := yaml.Unmarshal([]byte("Quorum"), &mode)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Quorum))
			})
		})

		Context("Round-trip", func() {
			It("should marshal and unmarshal Should", func() {
				original := control.Should
				data, err := yaml.Marshal(original)
				Expect(err).ToNot(HaveOccurred())

				var decoded control.Mode
				err = yaml.Unmarshal(data, &decoded)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			})
		})
	})

	Describe("Text", func() {
		Context("MarshalText", func() {
			It("should marshal Should to text", func() {
				data, err := control.Should.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Should"))
			})

			It("should marshal Must to text", func() {
				data, err := control.Must.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Must"))
			})

			It("should marshal AnyOf to text", func() {
				data, err := control.AnyOf.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("AnyOf"))
			})

			It("should marshal Quorum to text", func() {
				data, err := control.Quorum.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal("Quorum"))
			})

			It("should marshal Ignore to empty text", func() {
				data, err := control.Ignore.MarshalText()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(""))
			})
		})

		Context("UnmarshalText", func() {
			It("should unmarshal 'Should' from text", func() {
				var mode control.Mode
				err := mode.UnmarshalText([]byte("Should"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'must' from text", func() {
				var mode control.Mode
				err := mode.UnmarshalText([]byte("must"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Must))
			})

			It("should unmarshal 'AnyOf' from text", func() {
				var mode control.Mode
				err := mode.UnmarshalText([]byte("AnyOf"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.AnyOf))
			})

			It("should unmarshal 'Quorum' from text", func() {
				var mode control.Mode
				err := mode.UnmarshalText([]byte("Quorum"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Quorum))
			})
		})
	})

	Describe("CBOR", func() {
		Context("MarshalCBOR", func() {
			It("should marshal Should to CBOR", func() {
				data, err := control.Should.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeEmpty())
			})

			It("should marshal Must to CBOR", func() {
				data, err := control.Must.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())
				Expect(data).ToNot(BeEmpty())
			})
		})

		Context("UnmarshalCBOR", func() {
			It("should unmarshal 'Should' from CBOR", func() {
				data, err := cbor.Marshal("Should")
				Expect(err).ToNot(HaveOccurred())

				var mode control.Mode
				err = mode.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'must' from CBOR", func() {
				data, err := cbor.Marshal("must")
				Expect(err).ToNot(HaveOccurred())

				var mode control.Mode
				err = mode.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Must))
			})
		})

		Context("Round-trip", func() {
			It("should marshal and unmarshal Should", func() {
				original := control.Should
				data, err := original.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())

				var decoded control.Mode
				err = decoded.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			})

			It("should marshal and unmarshal Must", func() {
				original := control.Must
				data, err := original.MarshalCBOR()
				Expect(err).ToNot(HaveOccurred())

				var decoded control.Mode
				err = decoded.UnmarshalCBOR(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(decoded).To(Equal(original))
			})
		})
	})

	Describe("TOML", func() {
		Context("MarshalTOML", func() {
			It("should marshal Should to TOML", func() {
				data, err := control.Should.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"Should"`))
			})

			It("should marshal Must to TOML", func() {
				data, err := control.Must.MarshalTOML()
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(`"Must"`))
			})
		})

		Context("UnmarshalTOML", func() {
			It("should unmarshal 'Should' from TOML (bytes)", func() {
				var mode control.Mode
				err := mode.UnmarshalTOML([]byte("Should"))
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Should))
			})

			It("should unmarshal 'must' from TOML (string)", func() {
				var mode control.Mode
				err := mode.UnmarshalTOML("must")
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Equal(control.Must))
			})

			It("should return error for invalid type", func() {
				var mode control.Mode
				err := mode.UnmarshalTOML(123)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not in valid format"))
			})
		})
	})
})
