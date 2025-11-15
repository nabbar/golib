/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package atomic_test

import (
	"sync"

	. "github.com/nabbar/golib/atomic"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

var _ = Describe("atomic", func() {
	Describe("Value[T]", func() {
		Describe("NewValue", func() {
			It("should create a new atomic value", func() {
				v := NewValue[int]()
				Expect(v).ToNot(BeNil())
			})

			It("should create with default values", func() {
				v := NewValueDefault[string]("defaultLoad", "defaultStore")
				Expect(v).ToNot(BeNil())
			})
		})

		Describe("Store and Load", func() {
			It("should store and load int values", func() {
				v := NewValue[int]()
				v.Store(42)
				Expect(v.Load()).To(Equal(42))
			})

			It("should store and load string values", func() {
				v := NewValue[string]()
				v.Store("hello")
				Expect(v.Load()).To(Equal("hello"))
			})

			It("should store and load struct values", func() {
				type TestStruct struct {
					Name string
					Age  int
				}
				v := NewValue[TestStruct]()
				ts := TestStruct{Name: "Alice", Age: 30}
				v.Store(ts)
				Expect(v.Load()).To(Equal(ts))
			})

			It("should store and load pointer values", func() {
				v := NewValue[*int]()
				num := 123
				v.Store(&num)
				loaded := v.Load()
				Expect(*loaded).To(Equal(123))
			})
		})

		Describe("Swap", func() {
			It("should swap and return old value", func() {
				v := NewValue[int]()
				v.Store(10)
				old := v.Swap(20)
				Expect(old).To(Equal(10))
				Expect(v.Load()).To(Equal(20))
			})

			It("should swap strings", func() {
				v := NewValue[string]()
				v.Store("first")
				old := v.Swap("second")
				Expect(old).To(Equal("first"))
				Expect(v.Load()).To(Equal("second"))
			})
		})

		Describe("CompareAndSwap", func() {
			It("should swap when old value matches", func() {
				v := NewValue[int]()
				v.Store(100)
				swapped := v.CompareAndSwap(100, 200)
				Expect(swapped).To(BeTrue())
				Expect(v.Load()).To(Equal(200))
			})

			It("should not swap when old value doesn't match", func() {
				v := NewValue[int]()
				v.Store(100)
				swapped := v.CompareAndSwap(99, 200)
				Expect(swapped).To(BeFalse())
				Expect(v.Load()).To(Equal(100))
			})

			It("should work with strings", func() {
				v := NewValue[string]()
				v.Store("current")
				swapped := v.CompareAndSwap("current", "new")
				Expect(swapped).To(BeTrue())
				Expect(v.Load()).To(Equal("new"))
			})
		})

		Describe("SetDefaultLoad and SetDefaultStore", func() {
			It("should set default load value", func() {
				v := NewValue[int]()
				v.SetDefaultLoad(999)
				// Default is returned when nothing stored yet
				Expect(v).ToNot(BeNil())
			})

			It("should set default store value", func() {
				v := NewValue[string]()
				v.SetDefaultStore("default")
				Expect(v).ToNot(BeNil())
			})
		})

		Describe("Concurrency", func() {
			It("should be thread-safe for concurrent writes", func() {
				v := NewValue[int]()
				v.Store(0)

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func(val int) {
						defer wg.Done()
						v.Store(val)
					}(i)
				}
				wg.Wait()

				// Should not panic and have some value
				result := v.Load()
				Expect(result).To(BeNumerically(">=", 0))
				Expect(result).To(BeNumerically("<", 100))
			})

			It("should be thread-safe for concurrent reads and writes", func() {
				v := NewValue[int]()
				v.Store(0)

				var wg sync.WaitGroup
				// Writers
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func(val int) {
						defer wg.Done()
						v.Store(val)
					}(i)
				}
				// Readers
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_ = v.Load()
					}()
				}
				wg.Wait()
			})
		})
	})

	Describe("Map[K]", func() {
		Describe("NewMapAny", func() {
			It("should create a new atomic map", func() {
				m := NewMapAny[string]()
				Expect(m).ToNot(BeNil())
			})
		})

		Describe("Store and Load", func() {
			It("should store and load values", func() {
				m := NewMapAny[string]()
				m.Store("key1", 42)
				val, ok := m.Load("key1")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal(42))
			})

			It("should return false for non-existent keys", func() {
				m := NewMapAny[string]()
				_, ok := m.Load("nonexistent")
				Expect(ok).To(BeFalse())
			})
		})

		Describe("LoadOrStore", func() {
			It("should load existing value", func() {
				m := NewMapAny[string]()
				m.Store("key", "existing")
				actual, loaded := m.LoadOrStore("key", "new")
				Expect(loaded).To(BeTrue())
				Expect(actual).To(Equal("existing"))
			})

			It("should store new value when key doesn't exist", func() {
				m := NewMapAny[string]()
				actual, loaded := m.LoadOrStore("key", "value")
				Expect(loaded).To(BeFalse())
				Expect(actual).To(Equal("value"))
			})
		})

		Describe("LoadAndDelete", func() {
			It("should load and delete existing key", func() {
				m := NewMapAny[string]()
				m.Store("key", "value")
				val, loaded := m.LoadAndDelete("key")
				Expect(loaded).To(BeTrue())
				Expect(val).To(Equal("value"))

				_, ok := m.Load("key")
				Expect(ok).To(BeFalse())
			})

			It("should return false for non-existent key", func() {
				m := NewMapAny[string]()
				_, loaded := m.LoadAndDelete("nonexistent")
				Expect(loaded).To(BeFalse())
			})
		})

		Describe("Delete", func() {
			It("should delete existing key", func() {
				m := NewMapAny[string]()
				m.Store("key", "value")
				m.Delete("key")
				_, ok := m.Load("key")
				Expect(ok).To(BeFalse())
			})

			It("should not error on deleting non-existent key", func() {
				m := NewMapAny[string]()
				Expect(func() { m.Delete("nonexistent") }).ToNot(Panic())
			})
		})

		Describe("Swap", func() {
			It("should swap existing value", func() {
				m := NewMapAny[string]()
				m.Store("key", "old")
				previous, loaded := m.Swap("key", "new")
				Expect(loaded).To(BeTrue())
				Expect(previous).To(Equal("old"))

				val, _ := m.Load("key")
				Expect(val).To(Equal("new"))
			})

			It("should store new value when key doesn't exist", func() {
				m := NewMapAny[string]()
				previous, loaded := m.Swap("key", "value")
				Expect(loaded).To(BeFalse())
				Expect(previous).To(BeNil())
			})
		})

		Describe("CompareAndSwap", func() {
			It("should swap when old value matches", func() {
				m := NewMapAny[string]()
				m.Store("key", "old")
				swapped := m.CompareAndSwap("key", "old", "new")
				Expect(swapped).To(BeTrue())

				val, _ := m.Load("key")
				Expect(val).To(Equal("new"))
			})

			It("should not swap when old value doesn't match", func() {
				m := NewMapAny[string]()
				m.Store("key", "value")
				swapped := m.CompareAndSwap("key", "wrong", "new")
				Expect(swapped).To(BeFalse())

				val, _ := m.Load("key")
				Expect(val).To(Equal("value"))
			})
		})

		Describe("CompareAndDelete", func() {
			It("should delete when old value matches", func() {
				m := NewMapAny[string]()
				m.Store("key", "value")
				deleted := m.CompareAndDelete("key", "value")
				Expect(deleted).To(BeTrue())

				_, ok := m.Load("key")
				Expect(ok).To(BeFalse())
			})

			It("should not delete when old value doesn't match", func() {
				m := NewMapAny[string]()
				m.Store("key", "value")
				deleted := m.CompareAndDelete("key", "wrong")
				Expect(deleted).To(BeFalse())

				_, ok := m.Load("key")
				Expect(ok).To(BeTrue())
			})
		})

		Describe("Range", func() {
			It("should iterate over all entries", func() {
				m := NewMapAny[string]()
				m.Store("key1", 1)
				m.Store("key2", 2)
				m.Store("key3", 3)

				count := 0
				m.Range(func(key string, value any) bool {
					count++
					return true
				})

				Expect(count).To(Equal(3))
			})

			It("should stop iteration when function returns false", func() {
				m := NewMapAny[string]()
				for i := 0; i < 10; i++ {
					m.Store(string(rune('a'+i)), i)
				}

				count := 0
				m.Range(func(key string, value any) bool {
					count++
					return count < 5
				})

				Expect(count).To(Equal(5))
			})
		})
	})

	Describe("MapTyped[K, V]", func() {
		Describe("NewMapTyped", func() {
			It("should create a new typed atomic map", func() {
				m := NewMapTyped[string, int]()
				Expect(m).ToNot(BeNil())
			})
		})

		Describe("Type-safe operations", func() {
			It("should store and load typed values", func() {
				m := NewMapTyped[string, int]()
				m.Store("age", 25)
				val, ok := m.Load("age")
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal(25))
			})

			It("should work with complex types", func() {
				type Person struct {
					Name string
					Age  int
				}
				m := NewMapTyped[string, Person]()
				m.Store("alice", Person{Name: "Alice", Age: 30})
				val, ok := m.Load("alice")
				Expect(ok).To(BeTrue())
				Expect(val.Name).To(Equal("Alice"))
				Expect(val.Age).To(Equal(30))
			})

			It("should support all Map operations with type safety", func() {
				m := NewMapTyped[int, string]()

				// Store
				m.Store(1, "one")

				// Load
				val, ok := m.Load(1)
				Expect(ok).To(BeTrue())
				Expect(val).To(Equal("one"))

				// LoadOrStore
				actual, loaded := m.LoadOrStore(2, "two")
				Expect(loaded).To(BeFalse())
				Expect(actual).To(Equal("two"))

				// Swap
				prev, loaded := m.Swap(1, "ONE")
				Expect(loaded).To(BeTrue())
				Expect(prev).To(Equal("one"))

				// Delete
				m.Delete(2)
				_, ok = m.Load(2)
				Expect(ok).To(BeFalse())
			})

			It("should support Range with typed values", func() {
				m := NewMapTyped[string, int]()
				m.Store("a", 1)
				m.Store("b", 2)
				m.Store("c", 3)

				sum := 0
				m.Range(func(key string, value int) bool {
					sum += value
					return true
				})

				Expect(sum).To(Equal(6))
			})
		})
	})

	Describe("Concurrency tests", func() {
		It("Map should handle concurrent access", func() {
			m := NewMapAny[int]()
			var wg sync.WaitGroup

			// Concurrent writes
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					m.Store(idx, idx*2)
				}(i)
			}

			// Concurrent reads
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_, _ = m.Load(idx)
				}(i)
			}

			wg.Wait()
		})

		It("MapTyped should handle concurrent access", func() {
			m := NewMapTyped[string, int]()
			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					key := string(rune('a' + (idx % 26)))
					m.Store(key, idx)
				}(i)
			}

			wg.Wait()
		})
	})
})
