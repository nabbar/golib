/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package fields_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
)

var _ = Describe("Fields Management Operations", func() {
	Describe("Clean", func() {
		Context("on fields with data", func() {
			It("should remove all fields", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")

				Expect(f.Logrus()).To(HaveLen(3))

				f.Clean()

				Expect(f.Logrus()).To(HaveLen(0))
			})
		})

		Context("on empty fields", func() {
			It("should remain empty", func() {
				f := logfld.New(nil)

				f.Clean()

				Expect(f.Logrus()).To(HaveLen(0))
			})
		})
	})

	Describe("Get", func() {
		Context("with existing key", func() {
			It("should return value and true", func() {
				f := logfld.New(nil)
				f.Add("test_key", "test_value")

				val, ok := f.Get("test_key")

				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("test_value"))
			})

			It("should return correct type", func() {
				f := logfld.New(nil)
				f.Add("int_key", 42)
				f.Add("bool_key", true)
				f.Add("float_key", 3.14)

				intVal, ok := f.Get("int_key")
				Expect(ok).To(BeTrue())
				Expect(intVal).To(Equal(42))

				boolVal, ok := f.Get("bool_key")
				Expect(ok).To(BeTrue())
				Expect(boolVal).To(BeTrue())

				floatVal, ok := f.Get("float_key")
				Expect(ok).To(BeTrue())
				Expect(floatVal).To(Equal(3.14))
			})
		})

		Context("with non-existing key", func() {
			It("should return nil and false", func() {
				f := logfld.New(nil)

				val, ok := f.Get("non_existent")

				Expect(ok).To(BeFalse())
				Expect(val).To(BeNil())
			})
		})
	})

	Describe("Delete", func() {
		Context("with existing key", func() {
			It("should remove the key", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				result := f.Delete("key1")

				Expect(result).To(Equal(f))
				Expect(f.Logrus()).To(HaveLen(1))
				Expect(f.Logrus()).ToNot(HaveKey("key1"))
				Expect(f.Logrus()).To(HaveKey("key2"))
			})

			It("should allow method chaining", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")

				result := f.Delete("key1").Delete("key2")

				Expect(result).To(Equal(f))
				Expect(f.Logrus()).To(HaveLen(1))
				Expect(f.Logrus()).To(HaveKey("key3"))
			})
		})

		Context("with non-existing key", func() {
			It("should be no-op", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")

				result := f.Delete("non_existent")

				Expect(result).To(Equal(f))
				Expect(f.Logrus()).To(HaveLen(1))
			})
		})
	})

	Describe("Merge", func() {
		Context("with valid source fields", func() {
			It("should merge all fields", func() {
				base := logfld.New(nil)
				base.Add("base1", "value1")
				base.Add("base2", "value2")

				extra := logfld.New(nil)
				extra.Add("extra1", "value3")
				extra.Add("extra2", "value4")

				result := base.Merge(extra)

				Expect(result).To(Equal(base))
				Expect(base.Logrus()).To(HaveLen(4))
				Expect(base.Logrus()).To(HaveKey("base1"))
				Expect(base.Logrus()).To(HaveKey("base2"))
				Expect(base.Logrus()).To(HaveKey("extra1"))
				Expect(base.Logrus()).To(HaveKey("extra2"))
			})

			It("should overwrite existing keys", func() {
				base := logfld.New(nil)
				base.Add("key", "original")

				extra := logfld.New(nil)
				extra.Add("key", "updated")

				base.Merge(extra)

				val, ok := base.Get("key")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("updated"))
			})

			It("should not modify source fields", func() {
				base := logfld.New(nil)
				base.Add("base_key", "base_value")

				extra := logfld.New(nil)
				extra.Add("extra_key", "extra_value")

				base.Merge(extra)

				Expect(extra.Logrus()).To(HaveLen(1))
				Expect(extra.Logrus()).To(HaveKey("extra_key"))
				Expect(extra.Logrus()).ToNot(HaveKey("base_key"))
			})
		})

		Context("with nil source", func() {
			It("should return unchanged", func() {
				base := logfld.New(nil)
				base.Add("key", "value")

				result := base.Merge(nil)

				Expect(result).To(Equal(base))
				Expect(base.Logrus()).To(HaveLen(1))
			})
		})
	})

	Describe("Walk", func() {
		Context("with multiple fields", func() {
			It("should iterate all fields", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")

				visited := make(map[string]interface{})
				result := f.Walk(func(key string, val interface{}) bool {
					visited[key] = val
					return true
				})

				Expect(result).To(Equal(f))
				Expect(visited).To(HaveLen(3))
				Expect(visited).To(HaveKeyWithValue("key1", "value1"))
				Expect(visited).To(HaveKeyWithValue("key2", "value2"))
				Expect(visited).To(HaveKeyWithValue("key3", "value3"))
			})

			It("should stop iteration when callback returns false", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")

				count := 0
				f.Walk(func(key string, val interface{}) bool {
					count++
					return false // Stop after first
				})

				Expect(count).To(Equal(1))
			})

			It("should allow modification during walk", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				f.Walk(func(key string, val interface{}) bool {
					// Transform values during walk
					return true
				})

				Expect(f.Logrus()).To(HaveLen(2))
			})
		})

		Context("with empty fields", func() {
			It("should not call callback", func() {
				f := logfld.New(nil)

				called := false
				f.Walk(func(key string, val interface{}) bool {
					called = true
					return true
				})

				Expect(called).To(BeFalse())
			})
		})
	})

	Describe("WalkLimit", func() {
		Context("with specified keys", func() {
			It("should only iterate specified keys", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")
				f.Add("key3", "value3")
				f.Add("key4", "value4")

				visited := make(map[string]interface{})
				result := f.WalkLimit(func(key string, val interface{}) bool {
					visited[key] = val
					return true
				}, "key1", "key3")

				Expect(result).To(Equal(f))
				Expect(visited).To(HaveLen(2))
				Expect(visited).To(HaveKey("key1"))
				Expect(visited).To(HaveKey("key3"))
				Expect(visited).ToNot(HaveKey("key2"))
				Expect(visited).ToNot(HaveKey("key4"))
			})

			It("should skip non-existent keys", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")

				visited := make(map[string]interface{})
				f.WalkLimit(func(key string, val interface{}) bool {
					visited[key] = val
					return true
				}, "key1", "non_existent", "also_missing")

				Expect(visited).To(HaveLen(1))
				Expect(visited).To(HaveKey("key1"))
			})
		})

		Context("with empty key list", func() {
			It("should not iterate", func() {
				f := logfld.New(nil)

				called := false
				f.WalkLimit(func(key string, val interface{}) bool {
					called = true
					return true
				})

				Expect(called).To(BeFalse())
			})
		})
	})

	Describe("LoadOrStore", func() {
		Context("with existing key", func() {
			It("should return existing value", func() {
				f := logfld.New(nil)
				f.Add("key", "original")

				val, loaded := f.LoadOrStore("key", "new")

				Expect(loaded).To(BeTrue())
				Expect(val).To(Equal("original"))

				// Verify original value unchanged
				storedVal, _ := f.Get("key")
				Expect(storedVal).To(Equal("original"))
			})
		})

		Context("with non-existing key", func() {
			It("should store and return new value", func() {
				f := logfld.New(nil)

				val, loaded := f.LoadOrStore("key", "value")

				Expect(loaded).To(BeFalse())
				Expect(val).To(Equal("value"))

				// Verify value was stored
				storedVal, ok := f.Get("key")
				Expect(ok).To(BeTrue())
				Expect(storedVal).To(Equal("value"))
			})
		})
	})

	Describe("LoadAndDelete", func() {
		Context("with existing key", func() {
			It("should return value and delete key", func() {
				f := logfld.New(nil)
				f.Add("key", "value")

				val, loaded := f.LoadAndDelete("key")

				Expect(loaded).To(BeTrue())
				Expect(val).To(Equal("value"))

				// Verify key was deleted
				_, ok := f.Get("key")
				Expect(ok).To(BeFalse())
			})
		})

		Context("with non-existing key", func() {
			It("should return nil and false", func() {
				f := logfld.New(nil)

				val, loaded := f.LoadAndDelete("non_existent")

				Expect(loaded).To(BeFalse())
				Expect(val).To(BeNil())
			})
		})
	})
})
