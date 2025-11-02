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

var _ = Describe("Config Basic Operations", func() {
	Describe("NewConfig", func() {
		Context("with nil context function", func() {
			It("should create config with background context", func() {
				cfg := libctx.NewConfig[string](nil)
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.Err()).To(BeNil())
			})
		})

		Context("with custom context function", func() {
			It("should create config with custom context", func() {
				customCtx := context.WithValue(context.Background(), "test", "value")
				cfg := libctx.New[string](customCtx)
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.Value("test")).To(Equal("value"))
			})
		})

		Context("with different key types", func() {
			It("should work with string keys", func() {
				cfg := libctx.NewConfig[string](nil)
				Expect(cfg).ToNot(BeNil())
			})

			It("should work with int keys", func() {
				cfg := libctx.NewConfig[int](nil)
				Expect(cfg).ToNot(BeNil())
			})

			It("should work with custom type keys", func() {
				type CustomKey string
				cfg := libctx.NewConfig[CustomKey](nil)
				Expect(cfg).ToNot(BeNil())
			})
		})
	})

	Describe("Store and Load", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
		})

		It("should store and load a value", func() {
			cfg.Store("key1", "value1")
			val, ok := cfg.Load("key1")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("value1"))
		})

		It("should return false for non-existent key", func() {
			val, ok := cfg.Load("nonexistent")
			Expect(ok).To(BeFalse())
			Expect(val).To(BeNil())
		})

		It("should store multiple values", func() {
			cfg.Store("key1", "value1")
			cfg.Store("key2", 42)
			cfg.Store("key3", true)

			val1, ok1 := cfg.Load("key1")
			val2, ok2 := cfg.Load("key2")
			val3, ok3 := cfg.Load("key3")

			Expect(ok1).To(BeTrue())
			Expect(val1).To(Equal("value1"))
			Expect(ok2).To(BeTrue())
			Expect(val2).To(Equal(42))
			Expect(ok3).To(BeTrue())
			Expect(val3).To(Equal(true))
		})

		It("should overwrite existing value", func() {
			cfg.Store("key1", "value1")
			cfg.Store("key1", "value2")
			val, ok := cfg.Load("key1")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("value2"))
		})

		It("should handle nil values", func() {
			cfg.Store("key1", nil)
			val, ok := cfg.Load("key1")
			Expect(ok).To(BeFalse())
			Expect(val).To(BeNil())
		})

		It("should be thread-safe", func() {
			done := make(chan bool)

			// Concurrent writes
			for i := 0; i < 100; i++ {
				go func(index int) {
					cfg.Store("key", index)
					done <- true
				}(i)
			}

			// Wait for all writes
			for i := 0; i < 100; i++ {
				<-done
			}

			// Should have some value
			val, ok := cfg.Load("key")
			Expect(ok).To(BeTrue())
			Expect(val).ToNot(BeNil())
		})
	})

	Describe("Delete", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
			cfg.Store("key2", "value2")
		})

		It("should delete an existing key", func() {
			cfg.Delete("key1")
			val, ok := cfg.Load("key1")
			Expect(ok).To(BeFalse())
			Expect(val).To(BeNil())
		})

		It("should not affect other keys", func() {
			cfg.Delete("key1")
			val, ok := cfg.Load("key2")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("value2"))
		})

		It("should handle deleting non-existent key", func() {
			Expect(func() {
				cfg.Delete("nonexistent")
			}).ToNot(Panic())
		})
	})

	Describe("LoadOrStore", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
		})

		It("should load existing value", func() {
			cfg.Store("key1", "existing")
			val, loaded := cfg.LoadOrStore("key1", "new")
			Expect(loaded).To(BeTrue())
			Expect(val).To(Equal("existing"))
		})

		It("should store new value if key doesn't exist", func() {
			val, loaded := cfg.LoadOrStore("key1", "new")
			Expect(loaded).To(BeFalse())
			Expect(val).To(Equal("new"))

			// Verify it was stored
			storedVal, ok := cfg.Load("key1")
			Expect(ok).To(BeTrue())
			Expect(storedVal).To(Equal("new"))
		})

		It("should be atomic", func() {
			done := make(chan bool)
			results := make([]bool, 100)

			for i := 0; i < 100; i++ {
				go func(index int) {
					_, loaded := cfg.LoadOrStore("key", index)
					results[index] = loaded
					done <- true
				}(i)
			}

			for i := 0; i < 100; i++ {
				<-done
			}

			// Exactly one should have stored (loaded = false)
			storeCount := 0
			for _, loaded := range results {
				if !loaded {
					storeCount++
				}
			}
			Expect(storeCount).To(Equal(1))
		})
	})

	Describe("LoadAndDelete", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
		})

		It("should load and delete existing key", func() {
			val, loaded := cfg.LoadAndDelete("key1")
			Expect(loaded).To(BeTrue())
			Expect(val).To(Equal("value1"))

			// Verify it was deleted
			val2, ok := cfg.Load("key1")
			Expect(ok).To(BeFalse())
			Expect(val2).To(BeNil())
		})

		It("should return false for non-existent key", func() {
			val, loaded := cfg.LoadAndDelete("nonexistent")
			Expect(loaded).To(BeFalse())
			Expect(val).To(BeNil())
		})

		It("should be atomic", func() {
			cfg.Store("key", "value")
			done := make(chan bool)
			results := make([]bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					_, loaded := cfg.LoadAndDelete("key")
					results[index] = loaded
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			// Exactly one should have loaded (others already deleted)
			loadedCount := 0
			for _, loaded := range results {
				if loaded {
					loadedCount++
				}
			}
			Expect(loadedCount).To(Equal(1))
		})
	})

	Describe("Clean", func() {
		var cfg libctx.Config[string]

		BeforeEach(func() {
			cfg = libctx.NewConfig[string](nil)
			cfg.Store("key1", "value1")
			cfg.Store("key2", "value2")
			cfg.Store("key3", "value3")
		})

		It("should remove all stored values", func() {
			cfg.Clean()

			val1, ok1 := cfg.Load("key1")
			val2, ok2 := cfg.Load("key2")
			val3, ok3 := cfg.Load("key3")

			Expect(ok1).To(BeFalse())
			Expect(val1).To(BeNil())
			Expect(ok2).To(BeFalse())
			Expect(val2).To(BeNil())
			Expect(ok3).To(BeFalse())
			Expect(val3).To(BeNil())
		})

		It("should allow storing after clean", func() {
			cfg.Clean()
			cfg.Store("new_key", "new_value")

			val, ok := cfg.Load("new_key")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("new_value"))
		})
	})
})
