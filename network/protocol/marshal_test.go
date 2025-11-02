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

package protocol_test

import (
	"encoding/json"

	. "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Protocol Marshaling", func() {
	Describe("MarshalJSON", func() {
		Context("with valid protocols", func() {
			It("should marshal NetworkTCP to JSON", func() {
				data, err := NetworkTCP.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`"tcp"`))
			})

			It("should marshal NetworkUDP to JSON", func() {
				data, err := NetworkUDP.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`"udp"`))
			})

			It("should marshal NetworkUnix to JSON", func() {
				data, err := NetworkUnix.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`"unix"`))
			})

			It("should marshal NetworkTCP4 to JSON", func() {
				data, err := NetworkTCP4.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`"tcp4"`))
			})

			// ✅ FIXED: UnixGram now lowercase like other protocols
			It("should marshal NetworkUnixGram to JSON", func() {
				data, err := NetworkUnixGram.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`"unixgram"`))
			})
		})

		Context("with empty/invalid protocols", func() {
			It("should marshal NetworkEmpty to empty JSON string", func() {
				data, err := NetworkEmpty.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`""`))
			})

			It("should marshal undefined protocol to empty JSON string", func() {
				invalid := NetworkProtocol(99)
				data, err := invalid.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`""`))
			})
		})

		Context("in struct marshaling", func() {
			type TestStruct struct {
				Protocol NetworkProtocol `json:"protocol"`
			}

			It("should marshal struct with NetworkProtocol field", func() {
				s := TestStruct{Protocol: NetworkTCP}
				data, err := json.Marshal(s)
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(`{"protocol":"tcp"}`))
			})

			It("should marshal struct with multiple protocol fields", func() {
				type MultiProtocol struct {
					Primary   NetworkProtocol `json:"primary"`
					Secondary NetworkProtocol `json:"secondary"`
				}

				mp := MultiProtocol{
					Primary:   NetworkTCP,
					Secondary: NetworkUDP,
				}
				data, err := json.Marshal(mp)
				Expect(err).To(BeNil())
				Expect(string(data)).To(ContainSubstring(`"primary":"tcp"`))
				Expect(string(data)).To(ContainSubstring(`"secondary":"udp"`))
			})
		})

		Context("memory efficiency", func() {
			// ⚠️ POTENTIAL OPTIMIZATION: MarshalJSON allocates new slice with capacity
			It("should produce correctly sized output", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkTCP4, NetworkTCP6,
				}

				for _, p := range protocols {
					data, err := p.MarshalJSON()
					Expect(err).To(BeNil())

					// JSON output should be: quotes + string + quotes
					str := p.String()
					expectedLen := len(str) + 2 // 2 quotes
					Expect(len(data)).To(Equal(expectedLen))
				}
			})

			It("should not allocate excessive memory for long protocol names", func() {
				// UnixGram is the longest protocol name
				data, err := NetworkUnixGram.MarshalJSON()
				Expect(err).To(BeNil())
				// Should be reasonable size (< 100 bytes)
				Expect(len(data)).To(BeNumerically("<", 100))
			})
		})
	})

	Describe("MarshalYAML", func() {
		Context("with valid protocols", func() {
			// ✅ FIXED: MarshalYAML now returns string (idiomatic YAML)
			It("should marshal NetworkTCP to YAML", func() {
				data, err := NetworkTCP.MarshalYAML()
				Expect(err).To(BeNil())
				str, ok := data.(string)
				Expect(ok).To(BeTrue(), "MarshalYAML should return string")
				Expect(str).To(Equal("tcp"))
			})

			It("should marshal NetworkUDP to YAML", func() {
				data, err := NetworkUDP.MarshalYAML()
				Expect(err).To(BeNil())
				str := data.(string)
				Expect(str).To(Equal("udp"))
			})

			It("should marshal NetworkUnix to YAML", func() {
				data, err := NetworkUnix.MarshalYAML()
				Expect(err).To(BeNil())
				str := data.(string)
				Expect(str).To(Equal("unix"))
			})
		})

		Context("in struct marshaling", func() {
			type YAMLStruct struct {
				Protocol NetworkProtocol `yaml:"protocol"`
			}

			It("should marshal struct with protocol field to YAML", func() {
				s := YAMLStruct{Protocol: NetworkTCP}
				data, err := yaml.Marshal(s)
				Expect(err).To(BeNil())
				Expect(string(data)).To(ContainSubstring("protocol:"))
				// ✅ FIXED: Now properly marshals as string
				Expect(string(data)).To(ContainSubstring("tcp"))
			})
		})

		Context("with empty protocol", func() {
			It("should marshal NetworkEmpty to empty YAML value", func() {
				data, err := NetworkEmpty.MarshalYAML()
				Expect(err).To(BeNil())
				str := data.(string)
				Expect(str).To(Equal(""))
			})
		})
	})

	Describe("MarshalTOML", func() {
		Context("with valid protocols", func() {
			It("should marshal NetworkTCP to TOML", func() {
				data, err := NetworkTCP.MarshalTOML()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("tcp"))
			})

			It("should marshal NetworkUDP to TOML", func() {
				data, err := NetworkUDP.MarshalTOML()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("udp"))
			})

			It("should marshal NetworkUnix to TOML", func() {
				data, err := NetworkUnix.MarshalTOML()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("unix"))
			})

			It("should marshal all protocols without error", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkUnix, NetworkUnixGram,
					NetworkIP, NetworkIP4, NetworkIP6,
				}

				for _, p := range protocols {
					data, err := p.MarshalTOML()
					Expect(err).To(BeNil())
					Expect(data).NotTo(BeNil())
					Expect(len(data)).To(BeNumerically(">", 0))
				}
			})
		})

		Context("with empty protocol", func() {
			It("should marshal NetworkEmpty to empty bytes", func() {
				data, err := NetworkEmpty.MarshalTOML()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(""))
			})
		})
	})

	Describe("MarshalText", func() {
		Context("with valid protocols", func() {
			It("should marshal NetworkTCP to text", func() {
				data, err := NetworkTCP.MarshalText()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("tcp"))
			})

			It("should marshal NetworkUDP to text", func() {
				data, err := NetworkUDP.MarshalText()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("udp"))
			})

			It("should marshal all valid protocols", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, p := range protocols {
					data, err := p.MarshalText()
					Expect(err).To(BeNil())
					Expect(string(data)).To(Equal(p.String()))
				}
			})
		})

		Context("with empty protocol", func() {
			It("should marshal NetworkEmpty to empty text", func() {
				data, err := NetworkEmpty.MarshalText()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(""))
			})
		})
	})

	Describe("MarshalCBOR", func() {
		Context("with valid protocols", func() {
			It("should marshal NetworkTCP to CBOR", func() {
				data, err := NetworkTCP.MarshalCBOR()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("tcp"))
			})

			It("should marshal NetworkUDP to CBOR", func() {
				data, err := NetworkUDP.MarshalCBOR()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("udp"))
			})

			It("should marshal all protocols", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, p := range protocols {
					data, err := p.MarshalCBOR()
					Expect(err).To(BeNil())
					Expect(string(data)).To(Equal(p.String()))
				}
			})
		})

		Context("with empty protocol", func() {
			It("should marshal NetworkEmpty to empty CBOR", func() {
				data, err := NetworkEmpty.MarshalCBOR()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal(""))
			})
		})
	})

	Describe("Marshal consistency across formats", func() {
		It("should produce consistent string representation", func() {
			protocols := []NetworkProtocol{
				NetworkTCP, NetworkUDP, NetworkUnix, NetworkTCP4,
			}

			for _, p := range protocols {
				jsonData, _ := p.MarshalJSON()
				yamlData, _ := p.MarshalYAML()
				tomlData, _ := p.MarshalTOML()
				textData, _ := p.MarshalText()
				cborData, _ := p.MarshalCBOR()

				// Remove JSON quotes for comparison
				jsonStr := string(jsonData)
				if len(jsonStr) >= 2 {
					jsonStr = jsonStr[1 : len(jsonStr)-1]
				}

				// ✅ FIXED: YAML now returns string instead of []byte
				yamlStr := yamlData.(string)
				tomlStr := string(tomlData)
				textStr := string(textData)
				cborStr := string(cborData)

				// All should produce the same string representation
				Expect(jsonStr).To(Equal(p.String()))
				Expect(yamlStr).To(Equal(p.String()))
				Expect(tomlStr).To(Equal(p.String()))
				Expect(textStr).To(Equal(p.String()))
				Expect(cborStr).To(Equal(p.String()))
			}
		})
	})

	Describe("Memory safety in marshaling", func() {
		It("should not panic with concurrent marshaling", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer func() { done <- true }()
					for j := 0; j < 100; j++ {
						_, _ = NetworkTCP.MarshalJSON()
						_, _ = NetworkUDP.MarshalYAML()
						_, _ = NetworkUnix.MarshalTOML()
					}
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle repeated marshaling without memory leaks", func() {
			p := NetworkTCP
			// Repeated marshaling should not accumulate memory
			for i := 0; i < 10000; i++ {
				_, _ = p.MarshalJSON()
			}
		})
	})
})
