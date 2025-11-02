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

package viper_test

import (
	"context"

	liblog "github.com/nabbar/golib/logger"
	libvpr "github.com/nabbar/golib/viper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cleaner Operations (Unset)", func() {
	var (
		ctx context.Context
		log liblog.FuncLog
		v   libvpr.Viper
	)

	BeforeEach(func() {
		ctx = context.Background()
		log = func() liblog.Logger {
			return liblog.New(ctx)
		}
		v = libvpr.New(ctx, log)
	})

	Describe("Unset", func() {
		Context("with no keys", func() {
			It("should return nil when no keys provided", func() {
				err := v.Unset()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with single top-level key", func() {
			BeforeEach(func() {
				v.Viper().Set("key1", "value1")
				v.Viper().Set("key2", "value2")
			})

			It("should unset single key", func() {
				err := v.Unset("key1")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("key1")).To(BeEmpty())
				Expect(v.GetString("key2")).To(Equal("value2"))
			})
		})

		Context("with nested keys", func() {
			BeforeEach(func() {
				v.Viper().Set("app.name", "myapp")
				v.Viper().Set("app.version", "1.0.0")
				v.Viper().Set("app.settings.debug", true)
				v.Viper().Set("app.settings.timeout", 30)
			})

			It("should unset nested key", func() {
				err := v.Unset("app.settings.debug")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetBool("app.settings.debug")).To(BeFalse())
				Expect(v.GetInt("app.settings.timeout")).To(Equal(30))
				Expect(v.GetString("app.name")).To(Equal("myapp"))
			})

			It("should unset entire nested section", func() {
				err := v.Unset("app.settings")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetBool("app.settings.debug")).To(BeFalse())
				Expect(v.GetInt("app.settings.timeout")).To(Equal(0))
				Expect(v.GetString("app.name")).To(Equal("myapp"))
			})
		})

		Context("with multiple keys", func() {
			BeforeEach(func() {
				v.Viper().Set("key1", "value1")
				v.Viper().Set("key2", "value2")
				v.Viper().Set("key3", "value3")
			})

			It("should unset multiple keys at once", func() {
				err := v.Unset("key1", "key3")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("key1")).To(BeEmpty())
				Expect(v.GetString("key2")).To(Equal("value2"))
				Expect(v.GetString("key3")).To(BeEmpty())
			})
		})

		Context("with non-existent keys", func() {
			BeforeEach(func() {
				v.Viper().Set("existing", "value")
			})

			It("should handle non-existent keys gracefully", func() {
				err := v.Unset("nonexistent")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("existing")).To(Equal("value"))
			})
		})

		Context("with deeply nested structure", func() {
			BeforeEach(func() {
				v.Viper().Set("level1.level2.level3.level4.value", "deep")
				v.Viper().Set("level1.level2.other", "keep")
			})

			It("should unset deeply nested key", func() {
				err := v.Unset("level1.level2.level3.level4.value")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("level1.level2.level3.level4.value")).To(BeEmpty())
				Expect(v.GetString("level1.level2.other")).To(Equal("keep"))
			})

			It("should unset intermediate level", func() {
				err := v.Unset("level1.level2.level3")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("level1.level2.level3.level4.value")).To(BeEmpty())
				Expect(v.GetString("level1.level2.other")).To(Equal("keep"))
			})
		})

		Context("with complex data types", func() {
			BeforeEach(func() {
				v.Viper().Set("array", []string{"a", "b", "c"})
				v.Viper().Set("map", map[string]interface{}{"key": "value"})
				v.Viper().Set("number", 42)
			})

			It("should unset array", func() {
				err := v.Unset("array")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetStringSlice("array")).To(BeEmpty())
			})

			It("should unset map", func() {
				err := v.Unset("map")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetStringMap("map")).To(BeEmpty())
			})

			It("should unset number", func() {
				err := v.Unset("number")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetInt("number")).To(Equal(0))
			})
		})

		Context("with mixed nested and top-level keys", func() {
			BeforeEach(func() {
				v.Viper().Set("top", "value")
				v.Viper().Set("nested.key", "value")
				v.Viper().Set("deep.nested.key", "value")
			})

			It("should unset mixed keys", func() {
				err := v.Unset("top", "nested.key", "deep.nested.key")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("top")).To(BeEmpty())
				Expect(v.GetString("nested.key")).To(BeEmpty())
				Expect(v.GetString("deep.nested.key")).To(BeEmpty())
			})
		})

		Context("preserving other config", func() {
			BeforeEach(func() {
				v.Viper().Set("app.name", "myapp")
				v.Viper().Set("app.version", "1.0.0")
				v.Viper().Set("app.port", 8080)
				v.Viper().Set("database.host", "localhost")
				v.Viper().Set("database.port", 5432)
			})

			It("should preserve unrelated config when unsetting", func() {
				err := v.Unset("app.port")
				Expect(err).ToNot(HaveOccurred())

				// Verify unset key is gone
				Expect(v.GetInt("app.port")).To(Equal(0))

				// Verify other app config is preserved
				Expect(v.GetString("app.name")).To(Equal("myapp"))
				Expect(v.GetString("app.version")).To(Equal("1.0.0"))

				// Verify database config is preserved
				Expect(v.GetString("database.host")).To(Equal("localhost"))
				Expect(v.GetInt("database.port")).To(Equal(5432))
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("with empty string key", func() {
			It("should handle empty key gracefully", func() {
				v.Viper().Set("test", "value")
				err := v.Unset("")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("test")).To(Equal("value"))
			})
		})

		Context("with keys containing dots in value", func() {
			BeforeEach(func() {
				v.Viper().Set("url", "http://example.com")
				v.Viper().Set("config.url", "http://config.com")
			})

			It("should handle keys correctly", func() {
				err := v.Unset("config.url")
				Expect(err).ToNot(HaveOccurred())
				Expect(v.GetString("config.url")).To(BeEmpty())
				Expect(v.GetString("url")).To(Equal("http://example.com"))
			})
		})
	})
})
