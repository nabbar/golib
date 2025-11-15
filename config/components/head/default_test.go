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

package head_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/head"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Default configuration tests verify DefaultConfig functionality
var _ = Describe("Default Configuration", func() {
	Describe("DefaultConfig function", func() {
		Context("with no indentation", func() {
			It("should return valid JSON", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())

				// Verify it's valid JSON
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should contain security headers", func() {
				config := DefaultConfig("")
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				// Check for common security headers
				Expect(data).To(HaveKey("X-Frame-Options"))
				Expect(data).To(HaveKey("X-Content-Type-Options"))
			})

			It("should return non-empty data", func() {
				config := DefaultConfig("")
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(len(data)).To(BeNumerically(">", 0))
			})
		})

		Context("with indentation", func() {
			It("should format with 2 spaces", func() {
				config := DefaultConfig("  ")
				Expect(config).NotTo(BeEmpty())

				// Should be valid JSON
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format with 4 spaces", func() {
				config := DefaultConfig("    ")
				Expect(config).NotTo(BeEmpty())

				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format with tabs", func() {
				config := DefaultConfig("\t")
				Expect(config).NotTo(BeEmpty())

				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should contain newlines with indentation", func() {
				config := DefaultConfig("  ")
				configStr := string(config)

				// Indented JSON should have newlines
				Expect(configStr).To(ContainSubstring("\n"))
			})
		})

		Context("default header values", func() {
			It("should have X-Frame-Options", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				Expect(data).To(HaveKey("X-Frame-Options"))
				Expect(data["X-Frame-Options"]).NotTo(BeEmpty())
			})

			It("should have X-Content-Type-Options", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				Expect(data).To(HaveKey("X-Content-Type-Options"))
				Expect(data["X-Content-Type-Options"]).To(Equal("nosniff"))
			})

			It("should have security-focused defaults", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				// Common security headers
				expectedHeaders := []string{
					"X-Frame-Options",
					"X-Content-Type-Options",
					"Referrer-Policy",
				}

				for _, header := range expectedHeaders {
					Expect(data).To(HaveKey(header),
						"Expected default config to have header: "+header)
				}
			})
		})
	})

	Describe("Component DefaultConfig method", func() {
		var (
			ctx context.Context
			cpt CptHead
		)

		BeforeEach(func() {
			ctx = context.Background()
			cpt = New(ctx)
		})

		Context("method behavior", func() {
			It("should return same as global function", func() {
				globalConfig := DefaultConfig("")
				componentConfig := cpt.DefaultConfig("")

				var globalData map[string]string
				var componentData map[string]string

				json.Unmarshal(globalConfig, &globalData)
				json.Unmarshal(componentConfig, &componentData)

				Expect(componentData).To(Equal(globalData))
			})

			It("should work with different indentation", func() {
				config1 := cpt.DefaultConfig("")
				config2 := cpt.DefaultConfig("  ")
				config3 := cpt.DefaultConfig("\t")

				// All should be valid JSON
				var data1, data2, data3 map[string]string
				Expect(json.Unmarshal(config1, &data1)).To(Succeed())
				Expect(json.Unmarshal(config2, &data2)).To(Succeed())
				Expect(json.Unmarshal(config3, &data3)).To(Succeed())

				// All should have same data
				Expect(data1).To(Equal(data2))
				Expect(data2).To(Equal(data3))
			})
		})

		Context("consistency", func() {
			It("should return consistent results", func() {
				config1 := cpt.DefaultConfig("  ")
				config2 := cpt.DefaultConfig("  ")

				Expect(config1).To(Equal(config2))
			})

			It("should work across multiple components", func() {
				cpt2 := New(ctx)

				config1 := cpt.DefaultConfig("")
				config2 := cpt2.DefaultConfig("")

				Expect(config1).To(Equal(config2))
			})
		})
	})

	Describe("SetDefaultConfig", func() {
		Context("updating default configuration", func() {
			It("should allow setting custom default", func() {
				customConfig := []byte(`{"X-Custom-Header":"custom-value"}`)
				SetDefaultConfig(customConfig)

				// Get the new default
				config := DefaultConfig("")
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(data).To(HaveKey("X-Custom-Header"))
				Expect(data["X-Custom-Header"]).To(Equal("custom-value"))
			})

			It("should replace previous default", func() {
				// Set first config
				config1 := []byte(`{"X-Header-1":"value-1"}`)
				SetDefaultConfig(config1)

				result1 := DefaultConfig("")
				var data1 map[string]string
				json.Unmarshal(result1, &data1)
				Expect(data1).To(HaveKey("X-Header-1"))

				// Set second config
				config2 := []byte(`{"X-Header-2":"value-2"}`)
				SetDefaultConfig(config2)

				result2 := DefaultConfig("")
				var data2 map[string]string
				json.Unmarshal(result2, &data2)
				Expect(data2).To(HaveKey("X-Header-2"))
				Expect(data2).NotTo(HaveKey("X-Header-1"))
			})

			It("should handle empty config", func() {
				emptyConfig := []byte(`{}`)
				SetDefaultConfig(emptyConfig)

				config := DefaultConfig("")
				var data map[string]string
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
				Expect(data).To(BeEmpty())
			})

			It("should handle nil config gracefully", func() {
				Expect(func() {
					SetDefaultConfig(nil)
				}).NotTo(Panic())
			})
		})

		Context("with malformed JSON", func() {
			It("should still return something for invalid JSON", func() {
				// Set invalid JSON
				invalidConfig := []byte(`{invalid-json}`)
				SetDefaultConfig(invalidConfig)

				// DefaultConfig should still return something
				// (it might return the invalid JSON or handle it)
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
			})
		})
	})

	Describe("Default Configuration Content", func() {
		Context("security best practices", func() {
			It("should have restrictive X-Frame-Options", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				if val, ok := data["X-Frame-Options"]; ok {
					// Should be DENY or SAMEORIGIN
					Expect([]string{"DENY", "SAMEORIGIN"}).To(ContainElement(val))
				}
			})

			It("should have nosniff for X-Content-Type-Options", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				if val, ok := data["X-Content-Type-Options"]; ok {
					Expect(val).To(Equal("nosniff"))
				}
			})

			It("should have strict Referrer-Policy", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				if val, ok := data["Referrer-Policy"]; ok {
					// Should be restrictive
					restrictivePolicies := []string{
						"no-referrer",
						"same-origin",
						"strict-origin",
						"strict-origin-when-cross-origin",
					}
					Expect(restrictivePolicies).To(ContainElement(val))
				}
			})
		})

		Context("completeness", func() {
			It("should provide reasonable defaults", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
				Expect(len(config)).To(BeNumerically(">", 10))
			})

			It("should have valid header values", func() {
				config := DefaultConfig("")
				var data map[string]string
				json.Unmarshal(config, &data)

				// All values should be non-empty strings
				for key, value := range data {
					Expect(key).NotTo(BeEmpty())
					Expect(value).NotTo(BeEmpty())
				}
			})
		})
	})

	Describe("Format Handling", func() {
		Context("indentation formats", func() {
			It("should handle various indent strings", func() {
				indents := []string{"", " ", "  ", "    ", "\t", "\t\t"}

				for _, indent := range indents {
					config := DefaultConfig(indent)
					Expect(config).NotTo(BeEmpty(), "Failed with indent: "+indent)
				}
			})

			It("should produce different output with different indents", func() {
				config1 := DefaultConfig("")
				config2 := DefaultConfig("  ")

				// Unless the JSON is very simple, these should differ
				// (one should be formatted, one compact)
				if len(config1) > 20 && len(config2) > 20 {
					// If both are substantial, formatted one should be longer
					Expect(len(config2)).To(BeNumerically(">=", len(config1)))
				}
			})
		})

		Context("edge cases", func() {
			It("should handle very long indent string", func() {
				longIndent := ""
				for i := 0; i < 100; i++ {
					longIndent += " "
				}

				config := DefaultConfig(longIndent)
				Expect(config).NotTo(BeEmpty())
			})

			It("should handle special characters in indent", func() {
				// Most special chars won't make valid JSON indents, but shouldn't crash
				specialIndents := []string{"  ", "\t", "\n", "\r"}

				for _, indent := range specialIndents {
					Expect(func() {
						DefaultConfig(indent)
					}).NotTo(Panic())
				}
			})
		})
	})
})
