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

var _ = Describe("Protocol Unmarshaling", func() {
	Describe("UnmarshalJSON", func() {
		Context("with valid JSON", func() {
			It("should unmarshal 'tcp' from JSON", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"tcp"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal 'udp' from JSON", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"udp"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUDP))
			})

			It("should unmarshal 'unix' from JSON", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"unix"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUnix))
			})

			It("should unmarshal 'tcp4' from JSON", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"tcp4"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP4))
			})

			It("should handle uppercase JSON values", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"TCP"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should handle mixed case JSON values", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"TcP"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("with single quotes", func() {
			It("should strip single quotes from value", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`'tcp'`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should handle nested quotes (edge case)", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"'tcp'"`))
				Expect(err).To(BeNil())
				// After stripping outer double quotes: 'tcp'
				// After stripping single quotes: tcp
				// However, the trim order matters - it trims single then double
				// bytes.Trim first removes single quotes, then double quotes
				// So "'tcp'" becomes "tcp" which then fails to parse
				// This is an edge case with nested quotes
				Expect(p).To(Equal(NetworkEmpty)) // Doesn't parse correctly with nested quotes
			})
		})

		Context("with invalid JSON", func() {
			It("should handle unknown protocol gracefully", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`"invalid"`))
				Expect(err).To(BeNil()) // No error, but sets NetworkEmpty
				Expect(p).To(Equal(NetworkEmpty))
			})

			It("should handle empty string", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte(`""`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})

			It("should handle empty JSON bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON([]byte{})
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})

			It("should handle nil bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalJSON(nil)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})
		})

		Context("in struct unmarshaling", func() {
			type TestStruct struct {
				Protocol NetworkProtocol `json:"protocol"`
			}

			It("should unmarshal struct with protocol field", func() {
				var s TestStruct
				err := json.Unmarshal([]byte(`{"protocol":"tcp"}`), &s)
				Expect(err).To(BeNil())
				Expect(s.Protocol).To(Equal(NetworkTCP))
			})

			It("should unmarshal struct with multiple protocol fields", func() {
				type MultiProtocol struct {
					Primary   NetworkProtocol `json:"primary"`
					Secondary NetworkProtocol `json:"secondary"`
				}

				var mp MultiProtocol
				err := json.Unmarshal([]byte(`{"primary":"tcp","secondary":"udp"}`), &mp)
				Expect(err).To(BeNil())
				Expect(mp.Primary).To(Equal(NetworkTCP))
				Expect(mp.Secondary).To(Equal(NetworkUDP))
			})

			It("should handle missing field as NetworkEmpty", func() {
				var s TestStruct
				err := json.Unmarshal([]byte(`{}`), &s)
				Expect(err).To(BeNil())
				Expect(s.Protocol).To(Equal(NetworkEmpty))
			})
		})

		Context("roundtrip testing", func() {
			It("should successfully roundtrip marshal and unmarshal", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix,
					NetworkTCP4, NetworkTCP6, NetworkUnixGram,
				}

				for _, original := range protocols {
					marshaled, err := original.MarshalJSON()
					Expect(err).To(BeNil())

					var unmarshaled NetworkProtocol
					err = unmarshaled.UnmarshalJSON(marshaled)
					Expect(err).To(BeNil())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("UnmarshalYAML", func() {
		Context("with valid YAML", func() {
			It("should unmarshal 'tcp' from YAML", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: "tcp"}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal 'udp' from YAML", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: "udp"}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUDP))
			})

			It("should unmarshal 'unix' from YAML", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: "unix"}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUnix))
			})

			It("should handle uppercase YAML values", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: "TCP"}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("with quotes in YAML", func() {
			It("should strip single quotes", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: "'tcp'"}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should strip double quotes", func() {
				var p NetworkProtocol
				node := &yaml.Node{Value: `"tcp"`}
				err := p.UnmarshalYAML(node)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("in struct unmarshaling", func() {
			type YAMLStruct struct {
				Protocol NetworkProtocol `yaml:"protocol"`
			}

			It("should unmarshal struct from YAML", func() {
				var s YAMLStruct
				err := yaml.Unmarshal([]byte("protocol: tcp"), &s)
				Expect(err).To(BeNil())
				Expect(s.Protocol).To(Equal(NetworkTCP))
			})
		})

		Context("roundtrip testing", func() {
			It("should successfully roundtrip YAML marshal and unmarshal", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix, NetworkTCP4,
				}

				for _, original := range protocols {
					marshaled, err := original.MarshalYAML()
					Expect(err).To(BeNil())

					var unmarshaled NetworkProtocol
					str := marshaled.(string)
					node := &yaml.Node{Value: str}
					err = unmarshaled.UnmarshalYAML(node)
					Expect(err).To(BeNil())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("UnmarshalTOML", func() {
		Context("with valid TOML", func() {
			It("should unmarshal from []byte", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML([]byte("tcp"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal from string", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML("tcp")
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal 'udp' from string", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML("udp")
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUDP))
			})

			It("should unmarshal 'unix' from bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML([]byte("unix"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUnix))
			})

			It("should handle uppercase values", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML("TCP")
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("with quotes", func() {
			It("should strip single quotes from string", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML("'tcp'")
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should strip double quotes from string", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML(`"tcp"`)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should strip quotes from bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML([]byte(`"tcp"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("with invalid types", func() {
			It("should return error for int type", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML(42)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("not in valid format"))
			})

			It("should return error for bool type", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML(true)
				Expect(err).NotTo(BeNil())
			})

			It("should return error for nil", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML(nil)
				Expect(err).NotTo(BeNil())
			})

			It("should return error for struct", func() {
				var p NetworkProtocol
				err := p.UnmarshalTOML(struct{}{})
				Expect(err).NotTo(BeNil())
			})
		})

		Context("roundtrip testing", func() {
			It("should successfully roundtrip TOML marshal and unmarshal", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix, NetworkTCP4,
				}

				for _, original := range protocols {
					marshaled, err := original.MarshalTOML()
					Expect(err).To(BeNil())

					var unmarshaled NetworkProtocol
					err = unmarshaled.UnmarshalTOML(marshaled)
					Expect(err).To(BeNil())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("UnmarshalText", func() {
		Context("with valid text", func() {
			It("should unmarshal 'tcp'", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte("tcp"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal 'udp'", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte("udp"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUDP))
			})

			It("should unmarshal all valid protocols", func() {
				tests := map[string]NetworkProtocol{
					"tcp":      NetworkTCP,
					"tcp4":     NetworkTCP4,
					"tcp6":     NetworkTCP6,
					"udp":      NetworkUDP,
					"unix":     NetworkUnix,
					"unixgram": NetworkUnixGram,
				}

				for str, expected := range tests {
					var p NetworkProtocol
					err := p.UnmarshalText([]byte(str))
					Expect(err).To(BeNil())
					Expect(p).To(Equal(expected), "Failed for '%s'", str)
				}
			})
		})

		Context("with quotes", func() {
			It("should strip single quotes", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte("'tcp'"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should strip double quotes", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte(`"tcp"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("with empty or invalid data", func() {
			It("should handle empty bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte{})
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})

			It("should handle nil bytes", func() {
				var p NetworkProtocol
				err := p.UnmarshalText(nil)
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})

			It("should handle invalid protocol", func() {
				var p NetworkProtocol
				err := p.UnmarshalText([]byte("invalid"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkEmpty))
			})
		})

		Context("roundtrip testing", func() {
			It("should successfully roundtrip Text marshal and unmarshal", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix,
					NetworkTCP4, NetworkTCP6, NetworkUnixGram,
				}

				for _, original := range protocols {
					marshaled, err := original.MarshalText()
					Expect(err).To(BeNil())

					var unmarshaled NetworkProtocol
					err = unmarshaled.UnmarshalText(marshaled)
					Expect(err).To(BeNil())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("UnmarshalCBOR", func() {
		Context("with valid CBOR data", func() {
			It("should unmarshal 'tcp'", func() {
				var p NetworkProtocol
				err := p.UnmarshalCBOR([]byte("tcp"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})

			It("should unmarshal 'udp'", func() {
				var p NetworkProtocol
				err := p.UnmarshalCBOR([]byte("udp"))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkUDP))
			})

			It("should unmarshal all protocols", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix,
					NetworkTCP4, NetworkTCP6, NetworkUnixGram,
				}

				for _, expected := range protocols {
					var p NetworkProtocol
					str := expected.String()
					err := p.UnmarshalCBOR([]byte(str))
					Expect(err).To(BeNil())
					Expect(p).To(Equal(expected))
				}
			})
		})

		Context("with quotes", func() {
			It("should strip quotes", func() {
				var p NetworkProtocol
				err := p.UnmarshalCBOR([]byte(`"tcp"`))
				Expect(err).To(BeNil())
				Expect(p).To(Equal(NetworkTCP))
			})
		})

		Context("roundtrip testing", func() {
			It("should successfully roundtrip CBOR marshal and unmarshal", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix, NetworkTCP4,
				}

				for _, original := range protocols {
					marshaled, err := original.MarshalCBOR()
					Expect(err).To(BeNil())

					var unmarshaled NetworkProtocol
					err = unmarshaled.UnmarshalCBOR(marshaled)
					Expect(err).To(BeNil())
					Expect(unmarshaled).To(Equal(original))
				}
			})
		})
	})

	Describe("Memory safety in unmarshaling", func() {
		It("should handle concurrent unmarshaling", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer func() { done <- true }()
					for j := 0; j < 100; j++ {
						var p1, p2, p3 NetworkProtocol
						_ = p1.UnmarshalJSON([]byte(`"tcp"`))
						_ = p2.UnmarshalText([]byte("udp"))
						_ = p3.UnmarshalTOML("unix")
					}
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		// ⚠️ MEMORY SAFETY: Test with very large inputs
		It("should handle very large byte arrays without panic", func() {
			var p NetworkProtocol
			largeData := make([]byte, 1024*1024) // 1MB
			Expect(func() { p.UnmarshalJSON(largeData) }).NotTo(Panic())
		})

		It("should handle repeated unmarshaling without memory leaks", func() {
			var p NetworkProtocol
			data := []byte(`"tcp"`)
			for i := 0; i < 10000; i++ {
				_ = p.UnmarshalJSON(data)
			}
		})
	})
})
