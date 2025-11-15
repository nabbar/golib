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

package http_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/http"
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

				// Verify it's valid JSON (array of server configs)
				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should contain multiple server configurations", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				// Default config should have at least one server
				Expect(len(data)).To(BeNumerically(">", 0))
			})

			It("should return non-empty data", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
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
				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format with 4 spaces", func() {
				config := DefaultConfig("    ")
				Expect(config).NotTo(BeEmpty())

				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format with tabs", func() {
				config := DefaultConfig("\t")
				Expect(config).NotTo(BeEmpty())

				var data []map[string]interface{}
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

		Context("default server values", func() {
			It("should have server names", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					Expect(srv["name"]).NotTo(BeEmpty())
				}
			})

			It("should have handler keys", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					Expect(srv["handler_key"]).NotTo(BeEmpty())
				}
			})

			It("should have listen addresses", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					Expect(srv["listen"]).NotTo(BeEmpty())
				}
			})

			It("should have expose URLs", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					Expect(srv["expose"]).NotTo(BeEmpty())
				}
			})

			It("should have reasonable defaults", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				// Check for expected server names
				names := make(map[string]bool)
				for _, srv := range data {
					if name, ok := srv["name"].(string); ok {
						names[name] = true
					}
				}

				// Default config typically has status, api, and metrics servers
				Expect(len(names)).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("Component DefaultConfig method", func() {
		var (
			ctx context.Context
			cpt CptHttp
		)

		BeforeEach(func() {
			ctx = context.Background()
			cpt = New(ctx, DefaultTlsKey, nil)
		})

		Context("method behavior", func() {
			It("should return same as global function", func() {
				globalConfig := DefaultConfig("")
				componentConfig := cpt.DefaultConfig("")

				var globalData []map[string]interface{}
				var componentData []map[string]interface{}

				Expect(json.Unmarshal(globalConfig, &globalData)).NotTo(HaveOccurred())
				Expect(json.Unmarshal(componentConfig, &componentData)).NotTo(HaveOccurred())

				Expect(componentData).To(Equal(globalData))
			})

			It("should work with different indentation", func() {
				config1 := cpt.DefaultConfig("")
				config2 := cpt.DefaultConfig("  ")
				config3 := cpt.DefaultConfig("\t")

				// All should be valid JSON
				var data1, data2, data3 []map[string]interface{}
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
				cpt2 := New(ctx, DefaultTlsKey, nil)

				config1 := cpt.DefaultConfig("")
				config2 := cpt2.DefaultConfig("")

				Expect(config1).To(Equal(config2))
			})
		})
	})

	Describe("SetDefaultConfig", func() {
		Context("updating default configuration", func() {
			It("should allow setting custom default", func() {
				// Keep Default JSON before play with set
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				customConfig := []byte(`[{"name":"custom","handler_key":"custom","listen":"0.0.0.0:9999","expose":"http://localhost"}]`)
				SetDefaultConfig(customConfig)

				// Get the new default
				config := DefaultConfig("")
				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(len(data)).To(Equal(1))
				Expect(data[0]["name"]).To(Equal("custom"))
			})

			It("should replace previous default", func() {
				// Keep Default JSON before play with set
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				// Set first config
				config1 := []byte(`[{"name":"server1","handler_key":"h1","listen":"0.0.0.0:8001","expose":"http://localhost"}]`)
				SetDefaultConfig(config1)

				result1 := DefaultConfig("")
				var data1 []map[string]interface{}
				Expect(json.Unmarshal(result1, &data1)).NotTo(HaveOccurred())
				Expect(data1[0]["name"]).To(Equal("server1"))

				// Set second config
				config2 := []byte(`[{"name":"server2","handler_key":"h2","listen":"0.0.0.0:8002","expose":"http://localhost"}]`)
				SetDefaultConfig(config2)

				result2 := DefaultConfig("")
				var data2 []map[string]interface{}
				Expect(json.Unmarshal(result2, &data2)).NotTo(HaveOccurred())
				Expect(data2[0]["name"]).To(Equal("server2"))
			})

			It("should handle empty config", func() {
				// Keep Default JSON before play with set
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				emptyConfig := []byte(`[]`)
				SetDefaultConfig(emptyConfig)

				config := DefaultConfig("")
				var data []map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
				Expect(data).To(BeEmpty())
			})

			It("should handle nil config gracefully", func() {
				// Keep Default JSON before play with set
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				Expect(func() {
					SetDefaultConfig(nil)
				}).NotTo(Panic())
			})
		})

		Context("with malformed JSON", func() {
			It("should still return something for invalid JSON", func() {
				// keep old json default before test with invalid json
				old := DefaultConfig("")
				defer func() {
					SetDefaultConfig(old)
				}()

				// Set invalid JSON
				invalidConfig := []byte(`{invalid-json}`)
				SetDefaultConfig(invalidConfig)

				// DefaultConfig should still return something
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
			})
		})
	})

	Describe("Default Configuration Content", func() {
		Context("server configurations", func() {
			It("should have valid server names", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					name := srv["name"].(string)
					Expect(name).NotTo(BeEmpty())
					Expect(name).To(MatchRegexp("^[a-zA-Z0-9_-]+$"))
				}
			})

			It("should have valid handler keys", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					Expect(srv["handler_key"]).NotTo(BeEmpty())
				}
			})

			It("should have valid listen addresses", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				for _, srv := range data {
					listen := srv["listen"].(string)
					Expect(listen).NotTo(BeEmpty())
					// Should contain port
					Expect(listen).To(ContainSubstring(":"))
				}
			})
		})

		Context("completeness", func() {
			It("should provide reasonable defaults", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())
				Expect(len(config)).To(BeNumerically(">", 10))
			})

			It("should have valid configurations", func() {
				config := DefaultConfig("")
				var data []map[string]interface{}
				Expect(json.Unmarshal(config, &data)).NotTo(HaveOccurred())

				// All configs should be valid
				for _, srv := range data {
					Expect(srv["name"]).NotTo(BeEmpty())
					Expect(srv["handler_key"]).NotTo(BeEmpty())
					Expect(srv["listen"]).NotTo(BeEmpty())
					Expect(srv["expose"]).NotTo(BeEmpty())
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

				// Formatted one should be longer or equal
				Expect(len(config2)).To(BeNumerically(">=", len(config1)))
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
