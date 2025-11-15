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

var _ = Describe("Unmarshalling Operations", func() {
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

	Describe("UnmarshalKey", func() {
		type TestConfig struct {
			Name  string
			Value int
		}

		Context("with existing key", func() {
			BeforeEach(func() {
				v.Viper().Set("app.config", map[string]interface{}{
					"name":  "test",
					"value": 42,
				})
			})

			It("should unmarshal key successfully", func() {
				var config TestConfig
				err := v.UnmarshalKey("app.config", &config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("test"))
				Expect(config.Value).To(Equal(42))
			})
		})

		Context("with non-existent key", func() {
			It("should return error for missing key", func() {
				var config TestConfig
				err := v.UnmarshalKey("nonexistent.key", &config)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with nil value", func() {
			BeforeEach(func() {
				// Setting a key with actual data, then testing nil behavior
				v.Viper().Set("nil.key", map[string]interface{}{"data": nil})
			})

			It("should handle nil value in nested structure", func() {
				var config TestConfig
				err := v.UnmarshalKey("nil.key", &config)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with complex nested structure", func() {
			type NestedConfig struct {
				App struct {
					Name     string
					Version  string
					Settings struct {
						Debug   bool
						Timeout int
					}
				}
			}

			BeforeEach(func() {
				v.Viper().Set("nested", map[string]interface{}{
					"app": map[string]interface{}{
						"name":    "myapp",
						"version": "1.0.0",
						"settings": map[string]interface{}{
							"debug":   true,
							"timeout": 30,
						},
					},
				})
			})

			It("should unmarshal nested structure", func() {
				var config NestedConfig
				err := v.UnmarshalKey("nested", &config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.App.Name).To(Equal("myapp"))
				Expect(config.App.Version).To(Equal("1.0.0"))
				Expect(config.App.Settings.Debug).To(BeTrue())
				Expect(config.App.Settings.Timeout).To(Equal(30))
			})
		})
	})

	Describe("Unmarshal", func() {
		type AppConfig struct {
			Name    string
			Version string
			Port    int
		}

		Context("with valid config", func() {
			BeforeEach(func() {
				v.Viper().Set("name", "testapp")
				v.Viper().Set("version", "1.0.0")
				v.Viper().Set("port", 8080)
			})

			It("should unmarshal entire config", func() {
				var config AppConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("testapp"))
				Expect(config.Version).To(Equal("1.0.0"))
				Expect(config.Port).To(Equal(8080))
			})
		})

		Context("with extra fields in config", func() {
			BeforeEach(func() {
				v.Viper().Set("name", "testapp")
				v.Viper().Set("version", "1.0.0")
				v.Viper().Set("port", 8080)
				v.Viper().Set("extra", "field")
			})

			It("should ignore extra fields", func() {
				var config AppConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("testapp"))
			})
		})
	})

	Describe("UnmarshalExact", func() {
		type StrictConfig struct {
			Name string
			Port int
		}

		Context("with exact matching fields", func() {
			BeforeEach(func() {
				v.Viper().Set("name", "testapp")
				v.Viper().Set("port", 8080)
			})

			It("should unmarshal successfully", func() {
				var config StrictConfig
				err := v.UnmarshalExact(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("testapp"))
				Expect(config.Port).To(Equal(8080))
			})
		})

		Context("with extra fields in config", func() {
			BeforeEach(func() {
				v.Viper().Set("name", "testapp")
				v.Viper().Set("port", 8080)
				v.Viper().Set("extra", "field")
			})

			It("should return error for extra fields", func() {
				var config StrictConfig
				err := v.UnmarshalExact(&config)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Type Conversions", func() {
		type TypeConfig struct {
			StringVal string
			IntVal    int
			BoolVal   bool
			FloatVal  float64
		}

		Context("with string to int conversion", func() {
			BeforeEach(func() {
				v.Viper().Set("intval", "42")
			})

			It("should convert string to int", func() {
				var config TypeConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.IntVal).To(Equal(42))
			})
		})

		Context("with string to bool conversion", func() {
			BeforeEach(func() {
				v.Viper().Set("boolval", "true")
			})

			It("should convert string to bool", func() {
				var config TypeConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.BoolVal).To(BeTrue())
			})
		})
	})

	Describe("Array and Map Unmarshalling", func() {
		type CollectionConfig struct {
			Items []string
			Tags  map[string]string
		}

		BeforeEach(func() {
			v.Viper().Set("items", []string{"a", "b", "c"})
			v.Viper().Set("tags", map[string]string{"env": "prod", "region": "us"})
		})

		It("should unmarshal arrays and maps", func() {
			var config CollectionConfig
			err := v.Unmarshal(&config)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Items).To(Equal([]string{"a", "b", "c"}))
			Expect(config.Tags).To(HaveKeyWithValue("env", "prod"))
			Expect(config.Tags).To(HaveKeyWithValue("region", "us"))
		})
	})
})
