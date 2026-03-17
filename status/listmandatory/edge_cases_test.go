/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package listmandatory_test

import (
	stsctr "github.com/nabbar/golib/status/control"
	"github.com/nabbar/golib/status/listmandatory"
	stsmdt "github.com/nabbar/golib/status/mandatory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListMandatory Edge Cases", func() {
	var list listmandatory.ListMandatory

	BeforeEach(func() {
		list = listmandatory.New()
	})

	Context("Empty List Operations", func() {
		It("GetMode should return Ignore when list is empty", func() {
			Expect(list.GetMode("any-key")).To(Equal(stsctr.Ignore))
		})

		It("SetMode should do nothing when list is empty", func() {
			// Should not panic
			list.SetMode("any-key", stsctr.Must)
			Expect(list.Len()).To(Equal(0))
		})

		It("Del should do nothing when list is empty", func() {
			m := stsmdt.New()
			m.KeyAdd("key")
			list.Del(m)
			Expect(list.Len()).To(Equal(0))
		})

		It("DelKey should do nothing when list is empty", func() {
			list.DelKey("any-group")
			Expect(list.Len()).To(Equal(0))
		})

		It("Walk should not execute callback", func() {
			called := false
			list.Walk(func(_ string, _ stsmdt.Mandatory) bool {
				called = true
				return true
			})
			Expect(called).To(BeFalse())
		})

		It("GetList should return empty slice", func() {
			Expect(list.GetList()).To(BeEmpty())
		})
	})

	Context("Invalid Inputs", func() {
		It("Del should handle nil input", func() {
			Expect(func() {
				list.Del(nil)
			}).ToNot(Panic())
		})

		It("DelKey should handle empty string", func() {
			m := stsmdt.New()
			m.KeyAdd("key")
			m.SetName("group")
			list.Add(m)
			Expect(list.Len()).To(Equal(1))

			list.DelKey("")
			Expect(list.Len()).To(Equal(1))
		})

		It("SetMode with non-matching key should not affect existing groups", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Should)
			list.Add(m)

			list.SetMode("key2", stsctr.Must)
			Expect(list.GetMode("key1")).To(Equal(stsctr.Should))
		})
	})

	Context("Self-Healing (Invalid Entries)", func() {
		It("Len should filter out groups that became empty", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			list.Add(m)
			Expect(list.Len()).To(Equal(1))

			// Make the group invalid by removing all keys
			m.KeyDel("key1")

			// Len() should cleanup invalid entries
			Expect(list.Len()).To(Equal(0))
		})

		It("Walk should filter out groups that became empty", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			list.Add(m)

			m.KeyDel("key1")

			count := 0
			list.Walk(func(_ string, _ stsmdt.Mandatory) bool {
				count++
				return true
			})
			Expect(count).To(Equal(0))
		})

		It("GetMode should skip groups that became empty", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Must)
			list.Add(m)

			m.KeyDel("key1")

			// Even though we ask for a key that *was* there, the group is now invalid
			// effectively removing it from consideration.
			// However, since we deleted the key from the group, the group itself
			// doesn't have the key anymore either.
			// A better test is if the group had *another* key?
			// No, the condition is `len(v.KeyList()) < 1`.

			Expect(list.GetMode("key1")).To(Equal(stsctr.Ignore))
		})

		It("GetList should filter out groups that became empty", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			list.Add(m)

			m.KeyDel("key1")

			Expect(list.GetList()).To(BeEmpty())
		})
	})
})
