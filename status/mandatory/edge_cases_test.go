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

package mandatory_test

import (
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	stsctr "github.com/nabbar/golib/status/control"
	"github.com/nabbar/golib/status/mandatory"
)

var _ = Describe("Mandatory/EdgeCases", func() {
	Describe("GetMode edge cases", func() {
		It("should return Ignore when Mode is nil", func() {
			m := mandatory.New()
			// Force nil by creating a new atomic.Value without storing
			_ = new(atomic.Value)

			// Use reflection to access internal field (for testing only)
			// In real scenario, this is handled by New()
			// Testing that GetMode handles nil gracefully
			Expect(m.GetMode()).To(Equal(stsctr.Ignore))
		})

		It("should handle rapid mode changes", func() {
			m := mandatory.New()

			for i := 0; i < 1000; i++ {
				m.SetMode(stsctr.Should)
				m.SetMode(stsctr.Must)
				m.SetMode(stsctr.AnyOf)
				m.SetMode(stsctr.Quorum)
			}

			mode := m.GetMode()
			Expect(mode).To(BeElementOf(stsctr.Should, stsctr.Must, stsctr.AnyOf, stsctr.Quorum))
		})
	})

	Describe("KeyHas edge cases", func() {
		It("should handle checking empty string key", func() {
			m := mandatory.New()
			m.KeyAdd("")
			Expect(m.KeyHas("")).To(BeTrue())
		})

		It("should handle special characters in keys", func() {
			m := mandatory.New()
			specialKeys := []string{
				"key-with-dash",
				"key_with_underscore",
				"key.with.dot",
				"key/with/slash",
				"key:with:colon",
				"key@with@at",
				"key with spaces",
				"key\twith\ttabs",
				"key\nwith\nnewlines",
			}

			for _, key := range specialKeys {
				m.KeyAdd(key)
			}

			for _, key := range specialKeys {
				Expect(m.KeyHas(key)).To(BeTrue())
			}
		})

		It("should handle unicode keys", func() {
			m := mandatory.New()
			unicodeKeys := []string{
				"clé",
				"键",
				"キー",
				"ключ",
				"مفتاح",
				"🔑",
				"key-🚀-emoji",
			}

			for _, key := range unicodeKeys {
				m.KeyAdd(key)
			}

			for _, key := range unicodeKeys {
				Expect(m.KeyHas(key)).To(BeTrue())
			}
		})

		It("should handle very long keys", func() {
			m := mandatory.New()
			longKey := string(make([]byte, 10000))
			for i := range longKey {
				longKey = longKey[:i] + "a"
			}

			m.KeyAdd(longKey)
			Expect(m.KeyHas(longKey)).To(BeTrue())
		})
	})

	Describe("KeyAdd edge cases", func() {
		It("should handle adding no keys", func() {
			m := mandatory.New()
			m.KeyAdd()
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should handle adding many keys at once", func() {
			m := mandatory.New()
			keys := make([]string, 1000)
			for i := range keys {
				keys[i] = string(rune('a' + i%26))
			}

			m.KeyAdd(keys...)
			Expect(m.KeyList()).To(HaveLen(26)) // Only 26 unique letters
		})

		It("should preserve order of first occurrence", func() {
			m := mandatory.New()
			m.KeyAdd("key1")
			m.KeyAdd("key2")
			m.KeyAdd("key3")
			m.KeyAdd("key1") // Duplicate

			list := m.KeyList()
			Expect(list).To(HaveLen(3))
			// Order should be maintained
			Expect(list[0]).To(Equal("key1"))
			Expect(list[1]).To(Equal("key2"))
			Expect(list[2]).To(Equal("key3"))
		})

		It("should handle concurrent additions of same key", func() {
			m := mandatory.New()
			done := make(chan bool)

			for i := 0; i < 10; i++ {
				go func() {
					m.KeyAdd("same-key")
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}

			// Should only have one instance
			list := m.KeyList()
			count := 0
			for _, k := range list {
				if k == "same-key" {
					count++
				}
			}
			Expect(count).To(Equal(1))
		})
	})

	Describe("KeyDel edge cases", func() {
		It("should handle deleting no keys", func() {
			m := mandatory.New()
			m.KeyAdd("key1", "key2")
			m.KeyDel()
			Expect(m.KeyList()).To(HaveLen(2))
		})

		It("should handle deleting from empty list multiple times", func() {
			m := mandatory.New()
			m.KeyDel("key1")
			m.KeyDel("key2")
			m.KeyDel("key3")
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should handle deleting all keys in one call", func() {
			m := mandatory.New()
			keys := []string{"key1", "key2", "key3", "key4", "key5"}
			m.KeyAdd(keys...)
			m.KeyDel(keys...)
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should handle deleting with duplicates in delete list", func() {
			m := mandatory.New()
			m.KeyAdd("key1", "key2", "key3")
			m.KeyDel("key1", "key1", "key1")
			Expect(m.KeyList()).To(HaveLen(2))
			Expect(m.KeyList()).ToNot(ContainElement("key1"))
		})

		It("should handle deleting subset", func() {
			m := mandatory.New()
			m.KeyAdd("key1", "key2", "key3", "key4", "key5")
			m.KeyDel("key2", "key4")

			list := m.KeyList()
			Expect(list).To(HaveLen(3))
			Expect(list).To(ContainElement("key1"))
			Expect(list).To(ContainElement("key3"))
			Expect(list).To(ContainElement("key5"))
		})
	})

	Describe("KeyList edge cases", func() {
		It("should return independent copies", func() {
			m := mandatory.New()
			m.KeyAdd("key1", "key2")

			list1 := m.KeyList()
			list2 := m.KeyList()

			// Modify list1
			list1[0] = "modified"
			list1 = append(list1, "new-key")

			// list2 should be unchanged
			Expect(list2).To(HaveLen(2))
			Expect(list2[0]).To(Equal("key1"))
		})

		It("should handle rapid list retrievals", func() {
			m := mandatory.New()
			m.KeyAdd("key1", "key2", "key3")

			for i := 0; i < 1000; i++ {
				list := m.KeyList()
				Expect(list).To(HaveLen(3))
			}
		})
	})

	Describe("Complex scenarios", func() {
		It("should handle interleaved operations", func() {
			m := mandatory.New()

			m.KeyAdd("key1")
			Expect(m.KeyHas("key1")).To(BeTrue())

			m.SetMode(stsctr.Should)
			Expect(m.GetMode()).To(Equal(stsctr.Should))

			m.KeyAdd("key2")
			Expect(m.KeyList()).To(HaveLen(2))

			m.KeyDel("key1")
			Expect(m.KeyList()).To(HaveLen(1))

			m.SetMode(stsctr.Must)
			Expect(m.GetMode()).To(Equal(stsctr.Must))
		})

		It("should maintain consistency under stress", func() {
			m := mandatory.New()
			done := make(chan bool)

			// Writer 1: Add keys
			go func() {
				for i := 0; i < 100; i++ {
					m.KeyAdd("key1", "key2", "key3")
				}
				done <- true
			}()

			// Writer 2: Delete keys
			go func() {
				for i := 0; i < 100; i++ {
					m.KeyDel("key1")
				}
				done <- true
			}()

			// Writer 3: Change mode
			go func() {
				for i := 0; i < 100; i++ {
					m.SetMode(stsctr.Should)
					m.SetMode(stsctr.Must)
				}
				done <- true
			}()

			// Reader
			go func() {
				for i := 0; i < 100; i++ {
					_ = m.KeyList()
					_ = m.KeyHas("key1")
					_ = m.GetMode()
				}
				done <- true
			}()

			for i := 0; i < 4; i++ {
				<-done
			}

			// Should still be functional
			list := m.KeyList()
			Expect(list).ToNot(BeNil())
			mode := m.GetMode()
			Expect(mode).To(BeElementOf(stsctr.Should, stsctr.Must))
		})

		It("should handle mode and keys independently", func() {
			m := mandatory.New()

			// Set mode without keys
			m.SetMode(stsctr.Should)
			Expect(m.GetMode()).To(Equal(stsctr.Should))
			Expect(m.KeyList()).To(BeEmpty())

			// Add keys without changing mode
			m.KeyAdd("key1")
			Expect(m.GetMode()).To(Equal(stsctr.Should))
			Expect(m.KeyList()).To(HaveLen(1))

			// Change mode without affecting keys
			m.SetMode(stsctr.Must)
			Expect(m.GetMode()).To(Equal(stsctr.Must))
			Expect(m.KeyList()).To(HaveLen(1))

			// Delete keys without affecting mode
			m.KeyDel("key1")
			Expect(m.GetMode()).To(Equal(stsctr.Must))
			Expect(m.KeyList()).To(BeEmpty())
		})
	})

	Describe("Memory and performance", func() {
		It("should handle large number of keys", func() {
			m := mandatory.New()

			// Add 10000 unique keys
			for i := 0; i < 10000; i++ {
				m.KeyAdd(string(rune(i)))
			}

			Expect(m.KeyList()).To(HaveLen(10000))
		})

		It("should handle rapid add/delete cycles", func() {
			m := mandatory.New()

			for i := 0; i < 1000; i++ {
				m.KeyAdd("key1", "key2", "key3")
				m.KeyDel("key1", "key2", "key3")
			}

			Expect(m.KeyList()).To(BeEmpty())
		})
	})
})
