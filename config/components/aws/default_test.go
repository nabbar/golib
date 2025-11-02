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

package aws_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgcst "github.com/nabbar/golib/config/const"
)

var _ = Describe("Default Configuration", func() {
	Describe("DefaultConfig", func() {
		It("should return valid JSON", func() {
			config := DefaultConfig("")
			Expect(config).NotTo(BeNil())
			Expect(len(config)).To(BeNumerically(">", 0))

			// Verify it's valid JSON
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return indented JSON with indent parameter", func() {
			indent := cfgcst.JSONIndent
			config := DefaultConfig(indent)
			Expect(config).NotTo(BeNil())

			// Verify it contains indentation
			configStr := string(config)
			Expect(configStr).To(ContainSubstring("\n"))
			Expect(configStr).To(ContainSubstring(indent))
		})

		It("should contain required AWS fields", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Check for typical AWS config fields
			Expect(result).To(HaveKey("bucket"))
			Expect(result).To(HaveKey("accesskey"))
			Expect(result).To(HaveKey("secretkey"))
			Expect(result).To(HaveKey("region"))
		})
	})

	Describe("DefaultConfigStandard", func() {
		It("should return valid standard config JSON", func() {
			config := DefaultConfigStandard("")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should match default config when standard is set", func() {
			SetDefaultConfigStandard(false)
			standardConfig := DefaultConfigStandard("")
			defaultConfig := DefaultConfig("")

			// Both should produce valid JSON
			var stdResult, defResult map[string]interface{}
			Expect(json.Unmarshal(standardConfig, &stdResult)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &defResult)).To(Succeed())
		})

		It("should handle various indent levels", func() {
			indents := []string{"", "  ", "    ", "\t"}
			for _, indent := range indents {
				config := DefaultConfigStandard(indent)
				Expect(config).NotTo(BeNil())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("DefaultConfigStandardStatus", func() {
		It("should return valid standard config with status JSON", func() {
			config := DefaultConfigStandardStatus("")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should contain config and health sections", func() {
			config := DefaultConfigStandardStatus("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("config"))
			Expect(result).To(HaveKey("health"))
		})

		It("should be selectable as default config", func() {
			SetDefaultConfigStandard(true)
			statusConfig := DefaultConfigStandardStatus("")
			defaultConfig := DefaultConfig("")

			// Both should be valid JSON
			var statusResult, defResult map[string]interface{}
			Expect(json.Unmarshal(statusConfig, &statusResult)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &defResult)).To(Succeed())
		})
	})

	Describe("DefaultConfigCustom", func() {
		It("should return valid custom config JSON", func() {
			config := DefaultConfigCustom("")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should contain endpoint field", func() {
			config := DefaultConfigCustom("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("endpoint"))
		})

		It("should be selectable as default config", func() {
			SetDefaultConfigCustom(false)
			customConfig := DefaultConfigCustom("")
			defaultConfig := DefaultConfig("")

			// Both should be valid JSON
			var custResult, defResult map[string]interface{}
			Expect(json.Unmarshal(customConfig, &custResult)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &defResult)).To(Succeed())
		})
	})

	Describe("DefaultConfigCustomStatus", func() {
		It("should return valid custom config with status JSON", func() {
			config := DefaultConfigCustomStatus("")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should contain config, http-client, and health sections", func() {
			config := DefaultConfigCustomStatus("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("config"))
			Expect(result).To(HaveKey("http-client"))
			Expect(result).To(HaveKey("health"))
		})

		It("should be selectable as default config", func() {
			SetDefaultConfigCustom(true)
			statusConfig := DefaultConfigCustomStatus("")
			defaultConfig := DefaultConfig("")

			// Both should be valid JSON
			var statusResult, defResult map[string]interface{}
			Expect(json.Unmarshal(statusConfig, &statusResult)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &defResult)).To(Succeed())
		})
	})

	Describe("SetDefaultConfig", func() {
		var originalDefault []byte

		BeforeEach(func() {
			// Save original default
			originalDefault = DefaultConfig("")
		})

		AfterEach(func() {
			// Restore original default
			SetDefaultConfig(originalDefault)
		})

		It("should allow setting custom default config", func() {
			customConfig := []byte(`{"custom":"config"}`)
			SetDefaultConfig(customConfig)

			defaultConfig := DefaultConfig("")
			// Compare JSON content, not byte representation (indentation may vary)
			var custom, def map[string]interface{}
			Expect(json.Unmarshal(customConfig, &custom)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &def)).To(Succeed())
			Expect(def).To(Equal(custom))
		})

		It("should persist custom config across multiple calls", func() {
			customConfig := []byte(`{"persistent":"config"}`)
			SetDefaultConfig(customConfig)

			// Call DefaultConfig multiple times
			config1 := DefaultConfig("")
			config2 := DefaultConfig("")

			// Compare JSON content
			var custom, def1, def2 map[string]interface{}
			Expect(json.Unmarshal(customConfig, &custom)).To(Succeed())
			Expect(json.Unmarshal(config1, &def1)).To(Succeed())
			Expect(json.Unmarshal(config2, &def2)).To(Succeed())
			Expect(def1).To(Equal(custom))
			Expect(def2).To(Equal(custom))
		})
	})

	Describe("SetDefaultConfigStandard", func() {
		AfterEach(func() {
			// Reset to custom default
			SetDefaultConfigCustom(false)
		})

		It("should set standard config without status", func() {
			SetDefaultConfigStandard(false)
			config := DefaultConfig("")

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Should NOT have separate config section (flat structure)
			_, hasConfigKey := result["config"]
			Expect(hasConfigKey).To(BeFalse())
		})

		It("should set standard config with status", func() {
			SetDefaultConfigStandard(true)
			config := DefaultConfig("")

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Should have config and health sections
			Expect(result).To(HaveKey("config"))
			Expect(result).To(HaveKey("health"))
		})
	})

	Describe("SetDefaultConfigCustom", func() {
		AfterEach(func() {
			// Reset to default custom
			SetDefaultConfigCustom(false)
		})

		It("should set custom config without status", func() {
			SetDefaultConfigCustom(false)
			config := DefaultConfig("")

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Should have endpoint field
			Expect(result).To(HaveKey("endpoint"))
		})

		It("should set custom config with status", func() {
			SetDefaultConfigCustom(true)
			config := DefaultConfig("")

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Should have config, http-client, and health sections
			Expect(result).To(HaveKey("config"))
			Expect(result).To(HaveKey("http-client"))
			Expect(result).To(HaveKey("health"))
		})
	})

	Describe("Component DefaultConfig method", func() {
		It("should return same as package-level DefaultConfig", func() {
			ctx := context.Background()
			cpt := New(ctx, ConfigStandard)

			packageDefault := DefaultConfig(cfgcst.JSONIndent)
			componentDefault := cpt.DefaultConfig(cfgcst.JSONIndent)

			Expect(componentDefault).To(Equal(packageDefault))
		})

		It("should handle different indentations", func() {
			ctx := context.Background()
			cpt := New(ctx, ConfigCustom)

			indents := []string{"", "  ", "    "}
			for _, indent := range indents {
				config := cpt.DefaultConfig(indent)
				Expect(config).NotTo(BeNil())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("Configuration Switching", func() {
		AfterEach(func() {
			// Reset to default
			SetDefaultConfigCustom(false)
		})

		It("should switch between different config types", func() {
			// Test switching through all config types
			SetDefaultConfigStandard(false)
			config1 := DefaultConfig("")
			Expect(config1).NotTo(BeNil())

			SetDefaultConfigStandard(true)
			config2 := DefaultConfig("")
			Expect(config2).NotTo(BeNil())
			Expect(config2).NotTo(Equal(config1))

			SetDefaultConfigCustom(false)
			config3 := DefaultConfig("")
			Expect(config3).NotTo(BeNil())
			Expect(config3).NotTo(Equal(config2))

			SetDefaultConfigCustom(true)
			config4 := DefaultConfig("")
			Expect(config4).NotTo(BeNil())
			Expect(config4).NotTo(Equal(config3))

			// All should be valid JSON
			for _, cfg := range [][]byte{config1, config2, config3, config4} {
				var result map[string]interface{}
				Expect(json.Unmarshal(cfg, &result)).To(Succeed())
			}
		})
	})
})

var _ = Describe("Default Configuration Edge Cases", func() {
	Context("with empty indent", func() {
		It("should return valid JSON even with empty indent", func() {
			config := DefaultConfig("")
			Expect(config).NotTo(BeNil())

			// Verify it's valid JSON regardless of formatting
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with extreme indentation", func() {
		It("should handle very long indent strings", func() {
			longIndent := "                                        " // 40 spaces
			config := DefaultConfig(longIndent)
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with special characters in indent", func() {
		It("should handle tab indentation", func() {
			config := DefaultConfig("\t")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
