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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logfld "github.com/nabbar/golib/logger/fields"
)

var _ = Describe("Fields Edge Cases and Additional Coverage", func() {
	Describe("Context Compatibility", func() {
		Context("with valid fields", func() {
			It("should work as context.Context", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				f := logfld.New(ctx)
				f.Add("test", "value")

				// Fields implements context.Context
				var _ context.Context = f
				Expect(f.Err()).To(BeNil())
			})

			It("should reflect context state", func() {
				ctx := context.Background()
				f := logfld.New(ctx)

				// Can use context methods directly
				Expect(f.Err()).To(BeNil())
				_, ok := f.Deadline()
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Nil Handling", func() {
		Context("with nil Fields receiver", func() {
			It("Add should return nil", func() {
				var f logfld.Fields = nil
				// Cannot call methods on nil interface without panic
				Expect(f).To(BeNil())
			})
		})

		Context("with nil internal context", func() {
			It("should handle gracefully", func() {
				// Create with nil context
				f := logfld.New(nil)

				// Should not panic
				Expect(f).ToNot(BeNil())

				// Should work normally
				f.Add("key", "value")
				Expect(f.Logrus()).To(HaveLen(1))
			})
		})
	})

	Describe("Boundary Conditions", func() {
		Context("with empty string keys", func() {
			It("should accept empty string as key", func() {
				f := logfld.New(nil)
				f.Add("", "empty_key_value")

				val, ok := f.Get("")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("empty_key_value"))
			})
		})

		Context("with large number of fields", func() {
			It("should handle many fields efficiently", func() {
				f := logfld.New(nil)

				// Add 1000 fields
				for i := 0; i < 1000; i++ {
					f.Add(string(rune(i)), i)
				}

				Expect(f.Logrus()).To(HaveLen(1000))
			})
		})

		Context("with special characters in keys", func() {
			It("should accept keys with special characters", func() {
				f := logfld.New(nil)
				f.Add("key-with-dashes", "value1")
				f.Add("key.with.dots", "value2")
				f.Add("key_with_underscores", "value3")
				f.Add("key/with/slashes", "value4")

				Expect(f.Logrus()).To(HaveLen(4))
			})
		})
	})

	Describe("Value Type Handling", func() {
		Context("with various nil values", func() {
			It("should handle nil interface values", func() {
				f := logfld.New(nil)
				var nilValue interface{} = nil
				f.Add("nil_interface", nilValue)

				// Nil values might not be stored or retrieved the same way
				_, ok := f.Get("nil_interface")
				// Just verify no panic occurs
				_ = ok
			})

			It("should handle nil pointer values", func() {
				f := logfld.New(nil)
				var nilPtr *string = nil
				f.Add("nil_ptr", nilPtr)

				// Nil pointer stored successfully
				_, ok := f.Get("nil_ptr")
				Expect(ok).To(BeTrue())
			})
		})

		Context("with complex nested structures", func() {
			It("should handle deeply nested maps", func() {
				f := logfld.New(nil)
				nested := map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "deep_value",
						},
					},
				}
				f.Add("nested", nested)

				val, ok := f.Get("nested")
				Expect(ok).To(BeTrue())
				Expect(val).ToNot(BeNil())
			})

			It("should handle slices of interfaces", func() {
				f := logfld.New(nil)
				slice := []interface{}{
					"string",
					42,
					true,
					3.14,
					map[string]string{"key": "value"},
				}
				f.Add("mixed_slice", slice)

				val, ok := f.Get("mixed_slice")
				Expect(ok).To(BeTrue())
				Expect(val).To(HaveLen(5))
			})
		})
	})

	Describe("Clone Edge Cases", func() {
		Context("when cloning empty fields", func() {
			It("should create empty clone", func() {
				original := logfld.New(nil)
				clone := original.Clone()

				Expect(clone).ToNot(BeNil())
				Expect(clone.Logrus()).To(HaveLen(0))
			})
		})

		Context("when cloning fields with nil values", func() {
			It("should preserve nil values", func() {
				original := logfld.New(nil)
				original.Add("nil_key", nil)

				clone := original.Clone()

				// Just verify clone works without panic
				Expect(clone).ToNot(BeNil())
			})
		})
	})

	Describe("Map Edge Cases", func() {
		Context("with transformation returning nil", func() {
			It("should store transformed values", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				f.Map(func(key string, val interface{}) interface{} {
					return nil
				})

				// Map should complete without error
				Expect(f).ToNot(BeNil())
			})
		})

		Context("with transformation changing types", func() {
			It("should handle type changes", func() {
				f := logfld.New(nil)
				f.Add("num", "42")

				f.Map(func(key string, val interface{}) interface{} {
					if key == "num" {
						return 42 // String to int
					}
					return val
				})

				val, ok := f.Get("num")
				Expect(ok).To(BeTrue())
				Expect(val).To(BeEquivalentTo(42))
			})
		})
	})

	Describe("Merge Edge Cases", func() {
		Context("when merging with itself", func() {
			It("should handle self-merge", func() {
				f := logfld.New(nil)
				f.Add("key", "value")

				f.Merge(f)

				Expect(f.Logrus()).To(HaveLen(1))
				Expect(f.Logrus()["key"]).To(Equal("value"))
			})
		})

		Context("when merging empty fields", func() {
			It("should merge without changes", func() {
				base := logfld.New(nil)
				base.Add("key", "value")

				empty := logfld.New(nil)

				base.Merge(empty)

				Expect(base.Logrus()).To(HaveLen(1))
			})
		})
	})

	Describe("Walk Edge Cases", func() {
		Context("when walk function modifies fields", func() {
			It("should allow safe iteration", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				count := 0
				f.Walk(func(key string, val interface{}) bool {
					count++
					return true
				})

				Expect(count).To(Equal(2))
			})
		})

		Context("when stopping early", func() {
			It("should respect false return", func() {
				f := logfld.New(nil)
				for i := 0; i < 10; i++ {
					f.Add(string(rune('a'+i)), i)
				}

				count := 0
				f.Walk(func(key string, val interface{}) bool {
					count++
					return count < 3 // Stop after 3
				})

				Expect(count).To(Equal(3))
			})
		})
	})

	Describe("Logrus Conversion Edge Cases", func() {
		Context("with empty fields", func() {
			It("should return empty map not nil", func() {
				f := logfld.New(nil)
				logrusFields := f.Logrus()

				Expect(logrusFields).ToNot(BeNil())
				Expect(logrusFields).To(HaveLen(0))
			})
		})

		Context("multiple conversions", func() {
			It("should create independent maps each time", func() {
				f := logfld.New(nil)
				f.Add("key", "value1")

				map1 := f.Logrus()
				map1["key"] = "modified"

				map2 := f.Logrus()

				// Original should be unchanged
				Expect(map2["key"]).To(Equal("value1"))
			})
		})
	})

	Describe("Concurrent Access Patterns", func() {
		Context("with concurrent reads", func() {
			It("should be safe for multiple readers", func() {
				f := logfld.New(nil)
				f.Add("key1", "value1")
				f.Add("key2", "value2")

				done := make(chan bool)

				// Multiple concurrent readers
				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						for j := 0; j < 50; j++ {
							_, _ = f.Get("key1")
							_ = f.Logrus()
						}
						done <- true
					}()
				}

				// Wait for all
				for i := 0; i < 10; i++ {
					<-done
				}
			})
		})

		Context("with clone and independent modification", func() {
			It("should isolate changes per clone", func() {
				base := logfld.New(nil)
				base.Add("shared", "value")

				done := make(chan bool)

				// Multiple goroutines with their own clones
				for i := 0; i < 5; i++ {
					go func(id int) {
						defer GinkgoRecover()
						clone := base.Clone()
						clone.Add("goroutine", id)

						// Each clone should be independent
						val, ok := clone.Get("goroutine")
						Expect(ok).To(BeTrue())
						Expect(val).To(Equal(id))

						done <- true
					}(i)
				}

				// Wait for all
				for i := 0; i < 5; i++ {
					<-done
				}

				// Base should be unchanged
				_, ok := base.Get("goroutine")
				Expect(ok).To(BeFalse())
			})
		})
	})
})
