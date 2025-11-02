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

package status_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/fxamacker/cbor/v2"
	"github.com/nabbar/golib/monitor/status"
)

var _ = Describe("Status Encoding Formats", func() {
	Describe("Text Marshaling", func() {
		Context("MarshalText", func() {
			It("should marshal OK to text", func() {
				s := status.OK
				text, err := s.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(text)).To(Equal("OK"))
			})

			It("should marshal Warn to text", func() {
				s := status.Warn
				text, err := s.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(text)).To(Equal("Warn"))
			})

			It("should marshal KO to text", func() {
				s := status.KO
				text, err := s.MarshalText()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(text)).To(Equal("KO"))
			})
		})

		Context("UnmarshalText", func() {
			It("should unmarshal text to OK", func() {
				var s status.Status
				err := s.UnmarshalText([]byte("OK"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal text to Warn", func() {
				var s status.Status
				err := s.UnmarshalText([]byte("Warn"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal text to KO", func() {
				var s status.Status
				err := s.UnmarshalText([]byte("KO"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should handle lowercase", func() {
				var s status.Status
				err := s.UnmarshalText([]byte("ok"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should handle unknown values", func() {
				var s status.Status
				err := s.UnmarshalText([]byte("unknown"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("Round-trip", func() {
			It("should maintain status through text round-trip", func() {
				for _, original := range []status.Status{status.KO, status.Warn, status.OK} {
					text, err := original.MarshalText()
					Expect(err).NotTo(HaveOccurred())

					var unmarshaled status.Status
					err = unmarshaled.UnmarshalText(text)
					Expect(err).NotTo(HaveOccurred())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("YAML Marshaling", func() {
		Context("MarshalYAML", func() {
			It("should marshal OK to YAML", func() {
				s := status.OK
				data, err := s.MarshalYAML()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal("OK"))
			})

			It("should marshal Warn to YAML", func() {
				s := status.Warn
				data, err := s.MarshalYAML()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal("Warn"))
			})

			It("should marshal KO to YAML", func() {
				s := status.KO
				data, err := s.MarshalYAML()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal("KO"))
			})
		})

		Context("UnmarshalYAML", func() {
			It("should unmarshal YAML to OK", func() {
				var s status.Status
				node := &yaml.Node{Value: "OK"}
				err := s.UnmarshalYAML(node)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal YAML to Warn", func() {
				var s status.Status
				node := &yaml.Node{Value: "Warn"}
				err := s.UnmarshalYAML(node)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal YAML to KO", func() {
				var s status.Status
				node := &yaml.Node{Value: "KO"}
				err := s.UnmarshalYAML(node)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should handle lowercase", func() {
				var s status.Status
				node := &yaml.Node{Value: "ok"}
				err := s.UnmarshalYAML(node)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})
		})

		Context("with yaml.Marshal/Unmarshal", func() {
			It("should work with standard YAML marshaling", func() {
				type TestStruct struct {
					Status status.Status `yaml:"status"`
					Name   string        `yaml:"name"`
				}

				ts := TestStruct{
					Status: status.OK,
					Name:   "test",
				}

				data, err := yaml.Marshal(ts)
				Expect(err).NotTo(HaveOccurred())

				var result TestStruct
				err = yaml.Unmarshal(data, &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Status).To(Equal(status.OK))
				Expect(result.Name).To(Equal("test"))
			})
		})
	})

	Describe("TOML Marshaling", func() {
		Context("MarshalTOML", func() {
			It("should marshal OK to TOML", func() {
				s := status.OK
				data, err := s.MarshalTOML()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"OK"`))
			})

			It("should marshal Warn to TOML", func() {
				s := status.Warn
				data, err := s.MarshalTOML()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"Warn"`))
			})

			It("should marshal KO to TOML", func() {
				s := status.KO
				data, err := s.MarshalTOML()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`"KO"`))
			})
		})

		Context("UnmarshalTOML", func() {
			It("should unmarshal byte slice to OK", func() {
				var s status.Status
				err := s.UnmarshalTOML([]byte("OK"))
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal string to Warn", func() {
				var s status.Status
				err := s.UnmarshalTOML("Warn")
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal string to KO", func() {
				var s status.Status
				err := s.UnmarshalTOML("KO")
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should return error for invalid type", func() {
				var s status.Status
				err := s.UnmarshalTOML(123)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not in valid format"))
			})
		})
	})

	Describe("CBOR Marshaling", func() {
		Context("MarshalCBOR", func() {
			It("should marshal OK to CBOR", func() {
				s := status.OK
				data, err := s.MarshalCBOR()
				Expect(err).NotTo(HaveOccurred())
				Expect(data).NotTo(BeEmpty())

				// Verify it can be unmarshaled back
				var str string
				err = cbor.Unmarshal(data, &str)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).To(Equal("OK"))
			})

			It("should marshal Warn to CBOR", func() {
				s := status.Warn
				data, err := s.MarshalCBOR()
				Expect(err).NotTo(HaveOccurred())

				var str string
				err = cbor.Unmarshal(data, &str)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).To(Equal("Warn"))
			})

			It("should marshal KO to CBOR", func() {
				s := status.KO
				data, err := s.MarshalCBOR()
				Expect(err).NotTo(HaveOccurred())

				var str string
				err = cbor.Unmarshal(data, &str)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).To(Equal("KO"))
			})
		})

		Context("UnmarshalCBOR", func() {
			It("should unmarshal CBOR to OK", func() {
				data, err := cbor.Marshal("OK")
				Expect(err).NotTo(HaveOccurred())

				var s status.Status
				err = s.UnmarshalCBOR(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})

			It("should unmarshal CBOR to Warn", func() {
				data, err := cbor.Marshal("Warn")
				Expect(err).NotTo(HaveOccurred())

				var s status.Status
				err = s.UnmarshalCBOR(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.Warn))
			})

			It("should unmarshal CBOR to KO", func() {
				data, err := cbor.Marshal("KO")
				Expect(err).NotTo(HaveOccurred())

				var s status.Status
				err = s.UnmarshalCBOR(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.KO))
			})

			It("should handle lowercase in CBOR", func() {
				data, err := cbor.Marshal("ok")
				Expect(err).NotTo(HaveOccurred())

				var s status.Status
				err = s.UnmarshalCBOR(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal(status.OK))
			})
		})

		Context("Round-trip", func() {
			It("should maintain status through CBOR round-trip", func() {
				for _, original := range []status.Status{status.KO, status.Warn, status.OK} {
					data, err := original.MarshalCBOR()
					Expect(err).NotTo(HaveOccurred())

					var unmarshaled status.Status
					err = unmarshaled.UnmarshalCBOR(data)
					Expect(err).NotTo(HaveOccurred())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})
})
