/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libctx "github.com/nabbar/golib/context"
)

var _ = Describe("Config Advanced Operations", func() {
	Describe("Clone", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
			cfg.Store("key2", 42)
			cfg.Store("key3", true)
		})

		It("should create an independent copy", func() {
			clone := cfg.Clone(nil)
			Expect(clone).ToNot(BeNil())

			// Verify cloned data
			val1, ok1 := clone.Load("key1")
			val2, ok2 := clone.Load("key2")
			val3, ok3 := clone.Load("key3")

			Expect(ok1).To(BeTrue())
			Expect(val1).To(Equal("value1"))
			Expect(ok2).To(BeTrue())
			Expect(val2).To(Equal(42))
			Expect(ok3).To(BeTrue())
			Expect(val3).To(Equal(true))
		})

		It("should not share data with original", func() {
			clone := cfg.Clone(nil)

			// Modify clone
			clone.Store("key1", "modified")
			clone.Store("new_key", "new_value")

			// Original should be unchanged
			val, ok := cfg.Load("key1")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("value1"))

			val2, ok2 := cfg.Load("new_key")
			Expect(ok2).To(BeFalse())
			Expect(val2).To(BeNil())
		})

		It("should use custom context if provided", func() {
			customCtx := context.WithValue(context.Background(), "custom", "value")
			clone := cfg.Clone(customCtx)

			Expect(clone.Value("custom")).To(Equal("value"))
		})

		It("should preserve context function if nil context provided", func() {
			funcCtx := context.WithValue(context.Background(), "func", "value")
			cfg2 := libctx.New[string](funcCtx)
			cfg2.Store("key", "value")

			clone := cfg2.Clone(nil)
			Expect(clone).ToNot(BeNil())
		})

		It("should handle empty config", func() {
			emptyCfg := libctx.NewConfig[string](nil)
			clone := emptyCfg.Clone(nil)

			Expect(clone).ToNot(BeNil())
		})

		Context("with cancelled context", func() {
			It("should return nil when cloning cancelled config", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancelledCfg := libctx.New[string](ctx)
				cancelledCfg.Store("key", "value")
				cancel()

				clone := cancelledCfg.Clone(nil)
				Expect(clone).To(BeNil())
			})
		})
	})

	Describe("Merge", func() {
		var cfg1, cfg2 libctx.Config[string]

		BeforeEach(func() {
			cfg1 = libctx.NewConfig[string](nil)
			cfg1.Store("key1", "value1")
			cfg1.Store("key2", "value2")

			cfg2 = libctx.NewConfig[string](nil)
			cfg2.Store("key3", "value3")
			cfg2.Store("key4", "value4")
		})

		It("should merge values from another config", func() {
			result := cfg1.Merge(cfg2)
			Expect(result).To(BeTrue())

			// Original values should still exist
			val1, ok1 := cfg1.Load("key1")
			val2, ok2 := cfg1.Load("key2")
			Expect(ok1).To(BeTrue())
			Expect(val1).To(Equal("value1"))
			Expect(ok2).To(BeTrue())
			Expect(val2).To(Equal("value2"))

			// Merged values should be added
			val3, ok3 := cfg1.Load("key3")
			val4, ok4 := cfg1.Load("key4")
			Expect(ok3).To(BeTrue())
			Expect(val3).To(Equal("value3"))
			Expect(ok4).To(BeTrue())
			Expect(val4).To(Equal("value4"))
		})

		It("should overwrite existing keys", func() {
			cfg2.Store("key1", "new_value")
			cfg1.Merge(cfg2)

			val, ok := cfg1.Load("key1")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("new_value"))
		})

		It("should not modify source config", func() {
			cfg1.Merge(cfg2)

			// cfg2 should remain unchanged
			val3, ok3 := cfg2.Load("key3")
			val4, ok4 := cfg2.Load("key4")
			Expect(ok3).To(BeTrue())
			Expect(val3).To(Equal("value3"))
			Expect(ok4).To(BeTrue())
			Expect(val4).To(Equal("value4"))

			// cfg2 should not have cfg1's values
			val1, ok1 := cfg2.Load("key1")
			Expect(ok1).To(BeFalse())
			Expect(val1).To(BeNil())
		})

		It("should return false for nil config", func() {
			result := cfg1.Merge(nil)
			Expect(result).To(BeFalse())
		})

		It("should handle empty config merge", func() {
			emptyCfg := libctx.NewConfig[string](nil)
			result := cfg1.Merge(emptyCfg)
			Expect(result).To(BeTrue())

			// Original values should still exist
			val1, ok1 := cfg1.Load("key1")
			Expect(ok1).To(BeTrue())
			Expect(val1).To(Equal("value1"))
		})

		Context("with cancelled context", func() {
			It("should return false when merging into cancelled config", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancelledCfg := libctx.New[string](ctx)
				cancel()

				result := cancelledCfg.Merge(cfg2)
				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("Walk", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
			cfg.Store("key2", 42)
			cfg.Store("key3", true)
		})

		It("should iterate over all stored values", func() {
			keys := []string{}
			values := []interface{}{}

			cfg.Walk(func(key string, val interface{}) bool {
				keys = append(keys, key)
				values = append(values, val)
				return true
			})

			Expect(keys).To(ConsistOf("key1", "key2", "key3"))
			Expect(values).To(ConsistOf("value1", 42, true))
		})

		It("should stop iteration when function returns false", func() {
			count := 0
			cfg.Walk(func(key string, val interface{}) bool {
				count++
				return count < 2
			})

			Expect(count).To(Equal(2))
		})

		It("should handle empty config", func() {
			emptyCfg := libctx.NewConfig[string](nil)
			count := 0

			emptyCfg.Walk(func(key string, val interface{}) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})

		It("should allow modification during walk", func() {
			cfg.Walk(func(key string, val interface{}) bool {
				if key == "key1" {
					cfg.Store("new_key", "new_value")
				}
				return true
			})

			val, ok := cfg.Load("new_key")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("new_value"))
		})
	})

	Describe("WalkLimit", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
			cfg.Store("key2", "value2")
			cfg.Store("key3", "value3")
			cfg.Store("key4", "value4")
		})

		It("should iterate only over specified keys", func() {
			keys := []string{}

			cfg.WalkLimit(func(key string, val interface{}) bool {
				keys = append(keys, key)
				return true
			}, "key1", "key3")

			Expect(keys).To(ConsistOf("key1", "key3"))
		})

		It("should iterate over all keys when no valid keys specified", func() {
			keys := []string{}

			cfg.WalkLimit(func(key string, val interface{}) bool {
				keys = append(keys, key)
				return true
			})

			Expect(keys).To(ConsistOf("key1", "key2", "key3", "key4"))
		})

		It("should handle non-existent keys in valid keys list", func() {
			keys := []string{}

			cfg.WalkLimit(func(key string, val interface{}) bool {
				keys = append(keys, key)
				return true
			}, "key1", "nonexistent", "key2")

			Expect(keys).To(ConsistOf("key1", "key2"))
		})

		It("should stop iteration when function returns false", func() {
			count := 0

			cfg.WalkLimit(func(key string, val interface{}) bool {
				count++
				return count < 2
			}, "key1", "key2", "key3")

			Expect(count).To(Equal(2))
		})

		It("should handle empty valid keys list", func() {
			keys := []string{}

			cfg.WalkLimit(func(key string, val interface{}) bool {
				keys = append(keys, key)
				return true
			})

			Expect(len(keys)).To(Equal(4))
		})
	})

	Describe("GetContext", func() {
		It("should return background context when function is nil", func() {
			cfg := libctx.NewConfig[string](nil)
			ctx := cfg.GetContext()
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Err()).To(BeNil())
		})

		It("should return context from function", func() {
			customCtx := context.WithValue(context.Background(), "test", "value")
			cfg := libctx.NewConfig[string](customCtx)

			ctx := cfg.GetContext()
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Value("test")).To(Equal("value"))
		})

		It("should return background context when function returns nil", func() {
			cfg := libctx.New[string](nil)

			ctx := cfg.GetContext()
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Err()).To(BeNil())
		})
	})
})
