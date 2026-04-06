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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libatm "github.com/nabbar/golib/atomic"
)

var _ = Describe("Map implementations", func() {
	Context("MapAny[K]", func() {
		It("supports Store/Load/Delete/LoadAndDelete/LoadOrStore/Swap/CompareAndSwap/CompareAndDelete/Range", func() {
			m := libatm.NewMapAny[string]()

			// Store & Load
			m.Store("a", 1)
			v, ok := m.Load("a")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(1))

			// LoadOrStore existing
			act, loaded := m.LoadOrStore("a", 2)
			Expect(loaded).To(BeTrue())
			Expect(act).To(Equal(1))
			// LoadOrStore new
			act, loaded = m.LoadOrStore("b", 3)
			Expect(loaded).To(BeFalse())
			Expect(act).To(Equal(3))

			// CompareAndSwap
			Expect(m.CompareAndSwap("a", 1, 10)).To(BeTrue())
			v, _ = m.Load("a")
			Expect(v).To(Equal(10))

			// Swap
			prev, loaded := m.Swap("b", 30)
			Expect(loaded).To(BeTrue())
			Expect(prev).To(Equal(3))
			v, _ = m.Load("b")
			Expect(v).To(Equal(30))

			// CompareAndDelete
			Expect(m.CompareAndDelete("b", 30)).To(BeTrue())
			_, ok = m.Load("b")
			Expect(ok).To(BeFalse())

			// LoadAndDelete
			vv, loaded := m.LoadAndDelete("a")
			Expect(loaded).To(BeTrue())
			Expect(vv).To(Equal(10))
			_, ok = m.Load("a")
			Expect(ok).To(BeFalse())

			// Range with bad key type should auto-delete
			m.Store("x", 99)

			// Inject bad key type using export_test functionality
			if mi, ok := m.(interface{ GetUnderlyingMap() *sync.Map }); ok {
				mi.GetUnderlyingMap().Store(123, "bad key type")
			}

			var seenKeys []string
			m.Range(func(k string, val any) bool {
				seenKeys = append(seenKeys, k)
				return true
			})
			Expect(seenKeys).To(ContainElement("x"))
			Expect(seenKeys).ToNot(ContainElement(123))

			// Verify bad key was deleted
			if mi, ok := m.(interface{ GetUnderlyingMap() *sync.Map }); ok {
				_, ok = mi.GetUnderlyingMap().Load(123)
				Expect(ok).To(BeFalse())
			}

			// Delete
			m.Delete("x")
			_, ok = m.Load("x")
			Expect(ok).To(BeFalse())
		})

		It("is safe under concurrency for basic operations", func() {
			m := libatm.NewMapAny[int]()
			var wg sync.WaitGroup
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					m.Store(i, i)
					_, _ = m.Load(i)
					m.CompareAndSwap(i, i, i+1)
					m.Delete(i + 1)
				}(i)
			}
			wg.Wait()
		})
	})

	Context("MapTyped[K,V]", func() {
		It("wraps MapAny with typed API and Range casting", func() {
			m := libatm.NewMapTyped[string, int]()
			// Store & Load
			m.Store("a", 1)
			v, ok := m.Load("a")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(1))

			// LoadOrStore
			act, loaded := m.LoadOrStore("a", 2)
			Expect(loaded).To(BeTrue())
			Expect(act).To(Equal(1))
			act, loaded = m.LoadOrStore("b", 3)
			Expect(loaded).To(BeFalse())
			Expect(act).To(Equal(3))

			// Swap
			prev, loaded := m.Swap("b", 30)
			Expect(loaded).To(BeTrue())
			Expect(prev).To(Equal(3))

			// CompareAndSwap / CompareAndDelete
			Expect(m.CompareAndSwap("a", 1, 10)).To(BeTrue())
			Expect(m.CompareAndDelete("a", 10)).To(BeTrue())
			_, ok = m.Load("a")
			Expect(ok).To(BeFalse())

			// LoadAndDelete
			val, ok := m.LoadAndDelete("b")
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal(30))

			// Range
			m.Store("x", 5)

			// Inject bad value type using export_test functionality
			if mi, ok := m.(interface{ GetUnderlyingMap() libatm.Map[string] }); ok {
				mi.GetUnderlyingMap().Store("bad", "not an int")
			}

			var seen []string
			m.Range(func(k string, v int) bool {
				seen = append(seen, k)
				return true
			})
			Expect(seen).To(ContainElement("x"))
			Expect(seen).ToNot(ContainElement("bad"))

			// Verify bad value was deleted
			_, ok = m.Load("bad")
			Expect(ok).To(BeFalse())
		})
	})
})
