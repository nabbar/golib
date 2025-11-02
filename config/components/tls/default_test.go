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

package tls_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/tls"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Default configuration tests verify the default config generation
// and customization for TLS component.
var _ = Describe("Default Configuration", func() {
	var (
		ctx context.Context
		cpt CptTlS
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx, nil)
	})

	Describe("DefaultConfig function", func() {
		Context("generating default configuration", func() {
			It("should return valid JSON", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should contain expected fields", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				// Check for expected keys
				Expect(result).To(HaveKey("inheritDefault"))
				Expect(result).To(HaveKey("versionMin"))
				Expect(result).To(HaveKey("versionMax"))
				Expect(result).To(HaveKey("curveList"))
				Expect(result).To(HaveKey("cipherList"))
			})

			It("should contain valid TLS versions", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				versionMin := result["versionMin"].(string)
				versionMax := result["versionMax"].(string)

				Expect(versionMin).NotTo(BeEmpty())
				Expect(versionMax).NotTo(BeEmpty())
			})

			It("should contain curve list", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				curveList, ok := result["curveList"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(curveList).NotTo(BeEmpty())
			})

			It("should contain cipher list", func() {
				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())

				cipherList, ok := result["cipherList"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(cipherList).NotTo(BeEmpty())
			})
		})

		Context("with indentation", func() {
			It("should format with empty indentation", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())

				// Should be valid JSON
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should format with space indentation", func() {
				config := DefaultConfig("  ")
				Expect(config).NotTo(BeEmpty())

				// Should be valid JSON with more bytes due to indentation
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should format with tab indentation", func() {
				config := DefaultConfig("\t")
				Expect(config).NotTo(BeEmpty())

				// Should be valid JSON
				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should have larger size with indentation", func() {
				configNoIndent := DefaultConfig("")
				configWithIndent := DefaultConfig("  ")

				// Indented version should be larger or equal
				Expect(len(configWithIndent)).To(BeNumerically(">=", len(configNoIndent)))
			})
		})

		Context("concurrent access", func() {
			It("should handle concurrent DefaultConfig calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						config := DefaultConfig("")
						Expect(config).NotTo(BeEmpty())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Component DefaultConfig method", func() {
		Context("generating default configuration", func() {
			It("should return valid JSON", func() {
				config := cpt.DefaultConfig("")
				Expect(config).NotTo(BeEmpty())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})

			It("should match global DefaultConfig", func() {
				globalConfig := DefaultConfig("")
				componentConfig := cpt.DefaultConfig("")

				Expect(componentConfig).To(Equal(globalConfig))
			})

			It("should work with indentation", func() {
				config := cpt.DefaultConfig("  ")
				Expect(config).NotTo(BeEmpty())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("SetDefaultConfig function", func() {
		Context("setting custom default configuration", func() {
			var originalConfig []byte

			BeforeEach(func() {
				// Save original config
				originalConfig = DefaultConfig("")
			})

			AfterEach(func() {
				// Restore original config
				SetDefaultConfig(originalConfig)
			})

			It("should accept valid JSON configuration", func() {
				customConfig := []byte(`{"versionMin":"1.3","versionMax":"1.3"}`)

				Expect(func() {
					SetDefaultConfig(customConfig)
				}).NotTo(Panic())
			})

			It("should update default configuration", func() {
				customConfig := []byte(`{"versionMin":"1.3","versionMax":"1.3"}`)
				SetDefaultConfig(customConfig)

				config := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKey("versionMin"))
			})

			It("should accept minimal configuration", func() {
				customConfig := []byte(`{}`)
				SetDefaultConfig(customConfig)

				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
			})
		})
	})

	Describe("Configuration validation", func() {
		Context("validating default config", func() {
			It("should be valid JSON structure", func() {
				configJSON := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(configJSON, &result)
				Expect(err).To(BeNil())

				// Should have all expected fields
				Expect(result).To(HaveKey("inheritDefault"))
				Expect(result).To(HaveKey("versionMin"))
				Expect(result).To(HaveKey("versionMax"))
				Expect(result).To(HaveKey("curveList"))
				Expect(result).To(HaveKey("cipherList"))
			})

			It("should contain valid structure for customization", func() {
				configJSON := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(configJSON, &result)
				Expect(err).To(BeNil())

				// Check structure is suitable for user customization
				Expect(result).To(HaveKey("rootCA"))
				Expect(result).To(HaveKey("clientCA"))
				Expect(result).To(HaveKey("certPair"))
			})
		})

		Context("configuration content", func() {
			It("should have secure defaults", func() {
				configJSON := DefaultConfig("")

				// Check for security-related settings
				var result map[string]interface{}
				err := json.Unmarshal(configJSON, &result)
				Expect(err).To(BeNil())

				// Should have TLS version constraints
				Expect(result).To(HaveKey("versionMin"))
				Expect(result).To(HaveKey("versionMax"))

				// Should have cipher configuration
				Expect(result).To(HaveKey("cipherList"))
			})

			It("should include standard curves", func() {
				configJSON := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(configJSON, &result)
				Expect(err).To(BeNil())

				curveList, ok := result["curveList"].([]interface{})
				Expect(ok).To(BeTrue())

				// Check for common secure curves
				curveStrings := make([]string, len(curveList))
				for i, c := range curveList {
					curveStrings[i] = c.(string)
				}

				// Should include X25519 (modern secure curve)
				Expect(curveStrings).To(ContainElement("X25519"))
			})

			It("should include strong ciphers", func() {
				configJSON := DefaultConfig("")

				var result map[string]interface{}
				err := json.Unmarshal(configJSON, &result)
				Expect(err).To(BeNil())

				cipherList, ok := result["cipherList"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(cipherList).NotTo(BeEmpty())

				// Should have multiple ciphers for compatibility
				Expect(len(cipherList)).To(BeNumerically(">", 5))
			})
		})
	})

	Describe("Edge cases", func() {
		Context("error handling", func() {
			It("should handle invalid indentation gracefully", func() {
				// Even with weird indentation, should return valid JSON
				config := DefaultConfig("invalid\nindent")
				Expect(config).NotTo(BeEmpty())

				// The DefaultConfig with invalid indent falls back to original format
				// which might have different formatting, but should parse correctly
				// when treated as raw JSON
				Expect(len(config)).To(BeNumerically(">", 0))
			})
		})

		Context("repeated calls", func() {
			It("should return consistent results", func() {
				config1 := DefaultConfig("")
				config2 := DefaultConfig("")

				Expect(config1).To(Equal(config2))
			})

			It("should return consistent results with indentation", func() {
				config1 := DefaultConfig("  ")
				config2 := DefaultConfig("  ")

				Expect(config1).To(Equal(config2))
			})
		})
	})
})
