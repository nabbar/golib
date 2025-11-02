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
	"bytes"
	"context"
	"io"
	"time"

	liblog "github.com/nabbar/golib/logger"
	libvpr "github.com/nabbar/golib/viper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Viper Creation and Getters", func() {
	var (
		ctx context.Context
		log liblog.FuncLog
		v   libvpr.Viper
	)

	BeforeEach(func() {
		// Create context and logger for tests
		ctx = context.Background()
		log = func() liblog.Logger {
			return liblog.New(ctx)
		}
		v = libvpr.New(ctx, log)
	})

	Describe("New", func() {
		Context("with valid context and logger", func() {
			It("should create a new Viper instance", func() {
				Expect(v).ToNot(BeNil())
				Expect(v.Viper()).ToNot(BeNil())
			})
		})

		Context("with nil logger", func() {
			It("should create a new Viper instance with default logger", func() {
				v := libvpr.New(ctx, nil)
				Expect(v).ToNot(BeNil())
				Expect(v.Viper()).ToNot(BeNil())
			})
		})
	})

	Describe("Getter Methods", func() {
		BeforeEach(func() {
			// Set up test data in viper
			v.Viper().Set("test.bool", true)
			v.Viper().Set("test.string", "hello")
			v.Viper().Set("test.int", 42)
			v.Viper().Set("test.int32", int32(32))
			v.Viper().Set("test.int64", int64(64))
			v.Viper().Set("test.uint", uint(100))
			v.Viper().Set("test.uint16", uint16(16))
			v.Viper().Set("test.uint32", uint32(32))
			v.Viper().Set("test.uint64", uint64(64))
			v.Viper().Set("test.float64", 3.14)
			v.Viper().Set("test.duration", "5s")
			v.Viper().Set("test.time", time.Now())
			v.Viper().Set("test.intslice", []int{1, 2, 3})
			v.Viper().Set("test.stringslice", []string{"a", "b", "c"})
			v.Viper().Set("test.stringmap", map[string]interface{}{"key": "value"})
			v.Viper().Set("test.stringmapstring", map[string]string{"key": "value"})
			v.Viper().Set("test.stringmapstringslice", map[string][]string{"key": {"a", "b"}})
		})

		It("should get bool value", func() {
			Expect(v.GetBool("test.bool")).To(BeTrue())
		})

		It("should get string value", func() {
			Expect(v.GetString("test.string")).To(Equal("hello"))
		})

		It("should get int value", func() {
			Expect(v.GetInt("test.int")).To(Equal(42))
		})

		It("should get int32 value", func() {
			Expect(v.GetInt32("test.int32")).To(Equal(int32(32)))
		})

		It("should get int64 value", func() {
			Expect(v.GetInt64("test.int64")).To(Equal(int64(64)))
		})

		It("should get uint value", func() {
			Expect(v.GetUint("test.uint")).To(Equal(uint(100)))
		})

		It("should get uint16 value", func() {
			Expect(v.GetUint16("test.uint16")).To(Equal(uint16(16)))
		})

		It("should get uint32 value", func() {
			Expect(v.GetUint32("test.uint32")).To(Equal(uint32(32)))
		})

		It("should get uint64 value", func() {
			Expect(v.GetUint64("test.uint64")).To(Equal(uint64(64)))
		})

		It("should get float64 value", func() {
			Expect(v.GetFloat64("test.float64")).To(Equal(3.14))
		})

		It("should get duration value", func() {
			Expect(v.GetDuration("test.duration")).To(Equal(5 * time.Second))
		})

		It("should get time value", func() {
			Expect(v.GetTime("test.time")).ToNot(BeZero())
		})

		It("should get int slice", func() {
			Expect(v.GetIntSlice("test.intslice")).To(Equal([]int{1, 2, 3}))
		})

		It("should get string slice", func() {
			Expect(v.GetStringSlice("test.stringslice")).To(Equal([]string{"a", "b", "c"}))
		})

		It("should get string map", func() {
			m := v.GetStringMap("test.stringmap")
			Expect(m).To(HaveKey("key"))
			Expect(m["key"]).To(Equal("value"))
		})

		It("should get string map string", func() {
			Expect(v.GetStringMapString("test.stringmapstring")).To(Equal(map[string]string{"key": "value"}))
		})

		It("should get string map string slice", func() {
			Expect(v.GetStringMapStringSlice("test.stringmapstringslice")).To(Equal(map[string][]string{"key": {"a", "b"}}))
		})

		Context("with non-existent keys", func() {
			It("should return zero values for missing keys", func() {
				Expect(v.GetBool("nonexistent")).To(BeFalse())
				Expect(v.GetString("nonexistent")).To(BeEmpty())
				Expect(v.GetInt("nonexistent")).To(Equal(0))
				Expect(v.GetFloat64("nonexistent")).To(Equal(0.0))
			})
		})
	})

	Describe("Setters", func() {
		It("should set remote provider", func() {
			v.SetRemoteProvider("etcd")
			// No error expected, just verify it doesn't panic
		})

		It("should set remote endpoint", func() {
			v.SetRemoteEndpoint("http://localhost:2379")
			// No error expected, just verify it doesn't panic
		})

		It("should set remote path", func() {
			v.SetRemotePath("/config/app")
			// No error expected, just verify it doesn't panic
		})

		It("should set remote secure key", func() {
			v.SetRemoteSecureKey("secret-key")
			// No error expected, just verify it doesn't panic
		})

		It("should set remote model", func() {
			type TestModel struct {
				Name string
			}
			v.SetRemoteModel(&TestModel{})
			// No error expected, just verify it doesn't panic
		})

		It("should set remote reload function", func() {
			v.SetRemoteReloadFunc(func() {
				// Reload function
			})
			// No error expected, just verify it doesn't panic
		})

		It("should set home base name", func() {
			v.SetHomeBaseName("myapp")
			// No error expected, just verify it doesn't panic
		})

		It("should set env vars prefix", func() {
			v.SetEnvVarsPrefix("MYAPP")
			// No error expected, just verify it doesn't panic
		})

		It("should set default config", func() {
			v.SetDefaultConfig(func() io.Reader {
				return bytes.NewReader([]byte(`{"test": "value"}`))
			})
			// No error expected, just verify it doesn't panic
		})
	})

	Describe("Concurrent Access", func() {
		It("should be safe for concurrent reads", func() {
			v.Viper().Set("concurrent.test", "value")

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					Expect(v.GetString("concurrent.test")).To(Equal("value"))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})
