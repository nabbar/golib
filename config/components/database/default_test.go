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

package database_test

import (
	"context"
	"encoding/json"

	. "github.com/nabbar/golib/config/components/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cfgcst "github.com/nabbar/golib/config/const"
)

// Default configuration tests verify the DefaultConfig function behavior,
// including JSON validity, indentation handling, and field presence.
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

		It("should contain required database fields", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Check for typical database config fields
			Expect(result).To(HaveKey("driver"))
			Expect(result).To(HaveKey("name"))
			Expect(result).To(HaveKey("dsn"))
			Expect(result).To(HaveKey("disabled"))
		})

		It("should contain GORM configuration fields", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Check GORM-specific fields
			Expect(result).To(HaveKey("skip-default-transaction"))
			Expect(result).To(HaveKey("full-save-associations"))
			Expect(result).To(HaveKey("dry-run"))
			Expect(result).To(HaveKey("prepare-stmt"))
		})

		It("should contain connection pool fields", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Check connection pool fields
			Expect(result).To(HaveKey("enable-connection-pool"))
			Expect(result).To(HaveKey("pool-max-idle-conns"))
			Expect(result).To(HaveKey("pool-max-open-conns"))
			Expect(result).To(HaveKey("pool-conn-max-lifetime"))
		})

		It("should contain monitor configuration", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("monitor"))
		})

		It("should have default values set correctly", func() {
			config := DefaultConfig("")
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify default boolean values are false
			Expect(result["skip-default-transaction"]).To(BeFalse())
			Expect(result["full-save-associations"]).To(BeFalse())
			Expect(result["dry-run"]).To(BeFalse())
			Expect(result["prepare-stmt"]).To(BeFalse())
			Expect(result["disabled"]).To(BeFalse())
			Expect(result["enable-connection-pool"]).To(BeFalse())

			// Verify numeric defaults are 0
			Expect(result["create-batch-size"]).To(BeNumerically("==", 0))
			Expect(result["pool-max-idle-conns"]).To(BeNumerically("==", 0))
			Expect(result["pool-max-open-conns"]).To(BeNumerically("==", 0))
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
			customConfig := []byte(`{"driver":"sqlite","name":"custom"}`)
			SetDefaultConfig(customConfig)

			defaultConfig := DefaultConfig("")
			// Compare JSON content, not byte representation (indentation may vary)
			var custom, def map[string]interface{}
			Expect(json.Unmarshal(customConfig, &custom)).To(Succeed())
			Expect(json.Unmarshal(defaultConfig, &def)).To(Succeed())
			Expect(def).To(Equal(custom))
		})

		It("should persist custom config across multiple calls", func() {
			customConfig := []byte(`{"driver":"mysql","persistent":"config"}`)
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

	Describe("Component DefaultConfig method", func() {
		It("should return same as package-level DefaultConfig", func() {
			ctx := context.Background()
			cpt := New(ctx)

			packageDefault := DefaultConfig(cfgcst.JSONIndent)
			componentDefault := cpt.DefaultConfig(cfgcst.JSONIndent)

			Expect(componentDefault).To(Equal(packageDefault))
		})

		It("should handle different indentations", func() {
			ctx := context.Background()
			cpt := New(ctx)

			indents := []string{"", "  ", "    ", "\t"}
			for _, indent := range indents {
				config := cpt.DefaultConfig(indent)
				Expect(config).NotTo(BeNil())

				var result map[string]interface{}
				err := json.Unmarshal(config, &result)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})

// Edge cases for default configuration
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

		It("should handle mixed tab and space indentation", func() {
			config := DefaultConfig("\t  ")
			Expect(config).NotTo(BeNil())

			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with invalid custom config", func() {
		var originalDefault []byte

		BeforeEach(func() {
			originalDefault = DefaultConfig("")
		})

		AfterEach(func() {
			SetDefaultConfig(originalDefault)
		})

		It("should handle invalid JSON gracefully", func() {
			// Set invalid JSON as default
			invalidJSON := []byte(`{invalid json}`)
			SetDefaultConfig(invalidJSON)

			// DefaultConfig will return the invalid JSON
			config := DefaultConfig("")
			Expect(config).To(Equal(invalidJSON))

			// Unmarshaling will fail, which is expected
			var result map[string]interface{}
			err := json.Unmarshal(config, &result)
			Expect(err).To(HaveOccurred())
		})
	})
})
