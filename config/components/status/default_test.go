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
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Default configuration tests verify the DefaultConfig function and related behaviors.
var _ = Describe("Default Configuration", func() {
	Describe("DefaultConfig function", func() {
		Context("with no indentation", func() {
			It("should return a valid, non-empty JSON byte slice", func() {
				config := DefaultConfig("")
				Expect(config).NotTo(BeEmpty())

				// Verify it's valid JSON
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should contain the 'ReturnCode' key", func() {
				config := DefaultConfig("")
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(data).To(HaveKey("return-code"))
			})

			It("should return non-empty data", func() {
				config := DefaultConfig("")
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(len(data)).To(BeNumerically(">", 0))
			})
		})

		Context("with indentation", func() {
			It("should format correctly with 2 spaces", func() {
				config := DefaultConfig("  ")
				Expect(config).NotTo(BeEmpty())

				// Should be valid JSON
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format correctly with 4 spaces", func() {
				config := DefaultConfig("    ")
				Expect(config).NotTo(BeEmpty())

				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should format correctly with tabs", func() {
				config := DefaultConfig("\t")
				Expect(config).NotTo(BeEmpty())

				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
			})

			It("should contain newlines when indentation is used", func() {
				config := DefaultConfig("  ")
				configStr := string(config)

				// Indented JSON should have newlines
				Expect(configStr).To(ContainSubstring("\n"))
			})
		})
	})

	Describe("Component's DefaultConfig method", func() {
		var (
			ctx context.Context
			cpt CptStatus
		)

		BeforeEach(func() {
			ctx = context.Background()
			cpt = New(ctx)
		})

		Context("method behavior", func() {
			It("should return the same content as the global function", func() {
				globalConfig := DefaultConfig("")
				componentConfig := cpt.DefaultConfig("")

				var globalData map[string]interface{}
				var componentData map[string]interface{}

				json.Unmarshal(globalConfig, &globalData)
				json.Unmarshal(componentConfig, &componentData)

				Expect(componentData).To(Equal(globalData))
			})

			It("should work correctly with different indentation settings", func() {
				config1 := cpt.DefaultConfig("")
				config2 := cpt.DefaultConfig("  ")
				config3 := cpt.DefaultConfig("\t")

				// All should be valid JSON
				var data1, data2, data3 map[string]interface{}
				Expect(json.Unmarshal(config1, &data1)).To(Succeed())
				Expect(json.Unmarshal(config2, &data2)).To(Succeed())
				Expect(json.Unmarshal(config3, &data3)).To(Succeed())

				// All should have the same data content
				Expect(data1).To(Equal(data2))
				Expect(data2).To(Equal(data3))
			})
		})

		Context("consistency", func() {
			It("should return consistent results on multiple calls", func() {
				config1 := cpt.DefaultConfig("  ")
				config2 := cpt.DefaultConfig("  ")

				Expect(config1).To(Equal(config2))
			})

			It("should return the same default config across multiple component instances", func() {
				cpt2 := New(ctx)

				config1 := cpt.DefaultConfig("")
				config2 := cpt2.DefaultConfig("")

				Expect(config1).To(Equal(config2))
			})
		})
	})

	Describe("SetDefaultConfig function", func() {
		Context("when updating the default configuration", func() {
			It("should allow setting a custom default configuration", func() {
				customConfig := []byte(`{"return-code":{"OK":200}}`)
				SetDefaultConfig(customConfig)

				// Get the new default
				config := DefaultConfig("")
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())

				Expect(data).To(HaveKey("return-code"))
			})

			It("should replace the previous default configuration", func() {
				// Set first config
				config1 := []byte(`{"return-code":{"OK":200}}`)
				SetDefaultConfig(config1)

				result1 := DefaultConfig("")
				var data1 map[string]interface{}
				json.Unmarshal(result1, &data1)
				Expect(data1).To(HaveKey("return-code"))

				// Set second config
				config2 := []byte(`{"Component":[]}`)
				SetDefaultConfig(config2)

				result2 := DefaultConfig("")
				var data2 map[string]interface{}
				json.Unmarshal(result2, &data2)
				Expect(data2).To(HaveKey("Component"))
				Expect(data2).NotTo(HaveKey("return-code"))
			})

			It("should handle an empty JSON object as config", func() {
				emptyConfig := []byte(`{}`)
				SetDefaultConfig(emptyConfig)

				config := DefaultConfig("")
				var data map[string]interface{}
				err := json.Unmarshal(config, &data)
				Expect(err).To(BeNil())
				Expect(data).To(BeEmpty())
			})

			It("should handle a nil config gracefully without panicking", func() {
				Expect(func() {
					SetDefaultConfig(nil)
				}).NotTo(Panic())
			})
		})

		Context("with malformed JSON", func() {
			It("should still return the (invalid) byte slice without crashing", func() {
				// Set invalid JSON
				invalidConfig := []byte(`{invalid-json}`)
				SetDefaultConfig(invalidConfig)

				// DefaultConfig should still return the byte slice it was given.
				// The json.Indent call will fail and return the original slice.
				config := DefaultConfig("  ")
				Expect(config).To(Equal(invalidConfig))
			})
		})
	})

	Describe("Format Handling", func() {
		Context("with various indentation formats", func() {
			It("should handle various indent strings without error", func() {
				indents := []string{"", " ", "  ", "    ", "\t", "\t\t"}

				for _, indent := range indents {
					config := DefaultConfig(indent)
					Expect(config).NotTo(BeEmpty(), "Failed with indent: "+indent)
				}
			})
		})

		Context("with edge case indent strings", func() {
			It("should handle a very long indent string", func() {
				longIndent := ""
				for i := 0; i < 100; i++ {
					longIndent += " "
				}

				config := DefaultConfig(longIndent)
				Expect(config).NotTo(BeEmpty())
			})

			It("should handle special characters in indent string without panicking", func() {
				// Most special chars won't make valid JSON indents, but shouldn't crash.
				specialIndents := []string{"\n", "\r"}

				for _, indent := range specialIndents {
					Expect(func() {
						DefaultConfig(indent)
					}).NotTo(Panic())
				}
			})
		})
	})
})
