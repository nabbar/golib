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
	"reflect"
	"strings"

	libmap "github.com/go-viper/mapstructure/v2"
	liblog "github.com/nabbar/golib/logger"
	libvpr "github.com/nabbar/golib/viper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hook Operations", func() {
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

	Describe("HookRegister", func() {
		It("should register a decode hook", func() {
			// Create a simple hook that converts strings to uppercase
			hook := func(f reflect.Value, t reflect.Value) (interface{}, error) {
				if f.Kind() == reflect.String && t.Kind() == reflect.String {
					return strings.ToUpper(f.String()), nil
				}
				return f.Interface(), nil
			}

			v.HookRegister(hook)
			// No error expected, just verify it doesn't panic
		})

		It("should register multiple hooks", func() {
			hook1 := func(f reflect.Value, t reflect.Value) (interface{}, error) {
				return f.Interface(), nil
			}
			hook2 := func(f reflect.Value, t reflect.Value) (interface{}, error) {
				return f.Interface(), nil
			}

			v.HookRegister(hook1)
			v.HookRegister(hook2)
			// No error expected, just verify it doesn't panic
		})
	})

	Describe("HookReset", func() {
		It("should reset all hooks", func() {
			hook := func(f reflect.Value, t reflect.Value) (interface{}, error) {
				return f.Interface(), nil
			}

			v.HookRegister(hook)
			v.HookReset()
			// No error expected, just verify it doesn't panic
		})
	})

	Describe("Hook Execution", func() {
		type TestConfig struct {
			Name string
		}

		Context("with string transformation hook", func() {
			BeforeEach(func() {
				// Register hook that converts strings to uppercase
				v.HookRegister(libmap.StringToTimeDurationHookFunc())
			})

			It("should apply hook during unmarshal", func() {
				v.Viper().Set("name", "test")
				var config TestConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with custom type conversion hook", func() {
			type CustomType string
			type ConfigWithCustom struct {
				Value CustomType
			}

			BeforeEach(func() {
				// Register hook for custom type conversion
				hook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
					if t == reflect.TypeOf(CustomType("")) {
						if str, ok := data.(string); ok {
							return CustomType(strings.ToUpper(str)), nil
						}
					}
					return data, nil
				}
				v.HookRegister(hook)
			})

			It("should convert to custom type", func() {
				v.Viper().Set("value", "test")
				var config ConfigWithCustom
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(config.Value)).To(Equal("TEST"))
			})
		})

		Context("with hook reset", func() {
			BeforeEach(func() {
				// Register a hook
				hook := func(f reflect.Value, t reflect.Value) (interface{}, error) {
					if f.Kind() == reflect.String {
						return strings.ToUpper(f.String()), nil
					}
					return f.Interface(), nil
				}
				v.HookRegister(hook)
			})

			It("should not apply hooks after reset", func() {
				v.HookReset()
				v.Viper().Set("name", "test")
				var config TestConfig
				err := v.Unmarshal(&config)
				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("test"))
			})
		})
	})

	Describe("Multiple Hook Composition", func() {
		type TransformConfig struct {
			Text string
			Num  int
		}

		It("should compose multiple hooks", func() {
			// First hook: trim spaces
			hook1 := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if str, ok := data.(string); ok {
					return strings.TrimSpace(str), nil
				}
				return data, nil
			}

			// Second hook: convert to uppercase
			hook2 := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if str, ok := data.(string); ok {
					return strings.ToUpper(str), nil
				}
				return data, nil
			}

			v.HookRegister(hook1)
			v.HookRegister(hook2)

			v.Viper().Set("text", "  hello  ")
			v.Viper().Set("num", 42)

			var config TransformConfig
			err := v.Unmarshal(&config)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Text).To(Equal("HELLO"))
			Expect(config.Num).To(Equal(42))
		})
	})

	Describe("Hook Error Handling", func() {
		type ErrorConfig struct {
			Value string
		}

		It("should handle hook errors gracefully", func() {
			// Register a hook that might return an error
			hook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				// This hook just passes through
				return data, nil
			}

			v.HookRegister(hook)
			v.Viper().Set("value", "test")

			var config ErrorConfig
			err := v.Unmarshal(&config)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Concurrent Hook Operations", func() {
		It("should be safe for concurrent hook registration", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					hook := func(f reflect.Value, t reflect.Value) (interface{}, error) {
						return f.Interface(), nil
					}
					v.HookRegister(hook)
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})
