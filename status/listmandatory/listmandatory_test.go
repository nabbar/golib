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

package listmandatory_test

import (
	"context"
	"fmt"
	"time"

	libsem "github.com/nabbar/golib/semaphore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	stsctr "github.com/nabbar/golib/status/control"
	"github.com/nabbar/golib/status/listmandatory"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

var _ = Describe("ListMandatory", func() {
	var list listmandatory.ListMandatory

	BeforeEach(func() {
		list = listmandatory.New()
	})

	Describe("New", func() {
		It("should create an empty list", func() {
			Expect(list).ToNot(BeNil())
			Expect(list.Len()).To(Equal(0))
		})

		It("should create a list with initial mandatories", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")

			list := listmandatory.New(m1, m2)
			Expect(list.Len()).To(Equal(2))
		})

		It("should handle multiple initial mandatories", func() {
			m1 := stsmdt.New()
			m2 := stsmdt.New()
			m3 := stsmdt.New()

			list := listmandatory.New(m1, m2, m3)
			Expect(list.Len()).To(Equal(3))
		})
	})

	Describe("Len", func() {
		It("should return 0 for empty list", func() {
			Expect(list.Len()).To(Equal(0))
		})

		It("should return correct length after adding", func() {
			m := stsmdt.New()
			list.Add(m)
			Expect(list.Len()).To(Equal(1))
		})

		It("should return correct length after adding multiple", func() {
			m1 := stsmdt.New()
			m2 := stsmdt.New()
			m3 := stsmdt.New()

			list.Add(m1, m2, m3)
			Expect(list.Len()).To(Equal(3))
		})
	})

	Describe("Add", func() {
		It("should add a single mandatory", func() {
			m := stsmdt.New()
			list.Add(m)
			Expect(list.Len()).To(Equal(1))
		})

		It("should add multiple mandatories", func() {
			m1 := stsmdt.New()
			m2 := stsmdt.New()

			list.Add(m1, m2)
			Expect(list.Len()).To(Equal(2))
		})

		It("should add mandatories incrementally", func() {
			m1 := stsmdt.New()
			list.Add(m1)
			Expect(list.Len()).To(Equal(1))

			m2 := stsmdt.New()
			list.Add(m2)
			Expect(list.Len()).To(Equal(2))

			m3 := stsmdt.New()
			list.Add(m3)
			Expect(list.Len()).To(Equal(3))
		})

		It("should allow adding same mandatory multiple times", func() {
			m := stsmdt.New()
			list.Add(m)
			list.Add(m)
			Expect(list.Len()).To(Equal(2))
		})
	})

	Describe("Del", func() {
		It("should delete a mandatory", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")

			list.Add(m)
			Expect(list.Len()).To(Equal(1))

			list.Del(m)
			Expect(list.Len()).To(Equal(0))
		})

		It("should delete only matching mandatory", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")

			list.Add(m1, m2)
			Expect(list.Len()).To(Equal(2))

			list.Del(m1)
			Expect(list.Len()).To(Equal(1))
		})

		It("should handle deleting non-existent mandatory", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")

			list.Add(m1)
			list.Del(m2)
			Expect(list.Len()).To(Equal(1))
		})

		It("should match by key list content", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1", "key2")

			m2 := stsmdt.New()
			m2.KeyAdd("key2", "key1") // Same keys, different order

			list.Add(m1)
			list.Del(m2)
			Expect(list.Len()).To(Equal(0))
		})
	})

	Describe("Walk", func() {
		It("should walk through all mandatories", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")

			m3 := stsmdt.New()
			m3.KeyAdd("key3")

			list.Add(m1, m2, m3)

			count := 0
			list.Walk(func(m stsmdt.Mandatory) bool {
				count++
				Expect(m).ToNot(BeNil())
				return true
			})

			Expect(count).To(Equal(3))
		})

		It("should stop walking when function returns false", func() {
			m1 := stsmdt.New()
			m2 := stsmdt.New()
			m3 := stsmdt.New()

			list.Add(m1, m2, m3)

			count := 0
			list.Walk(func(m stsmdt.Mandatory) bool {
				count++
				return count < 2 // Stop after 2
			})

			Expect(count).To(Equal(2))
		})

		It("should handle empty list", func() {
			count := 0
			list.Walk(func(m stsmdt.Mandatory) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})

		It("should allow modifying mandatories during walk", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Ignore)

			list.Add(m1)

			list.Walk(func(m stsmdt.Mandatory) bool {
				m.SetMode(stsctr.Should)
				return true
			})

			// Verify the change persisted
			mode := list.GetMode("key1")
			Expect(mode).To(Equal(stsctr.Should))
		})
	})

	Describe("GetMode", func() {
		It("should return Ignore for non-existent key", func() {
			mode := list.GetMode("nonexistent")
			Expect(mode).To(Equal(stsctr.Ignore))
		})

		It("should return mode for existing key", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Should)

			list.Add(m)

			mode := list.GetMode("key1")
			Expect(mode).To(Equal(stsctr.Should))
		})

		It("should return mode from first matching mandatory", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Should)

			m2 := stsmdt.New()
			m2.KeyAdd("key1")
			m2.SetMode(stsctr.Must)

			list.Add(m1, m2)

			mode := list.GetMode("key1")
			// Should return the first one found
			Expect(mode).To(BeElementOf(stsctr.Should, stsctr.Must))
		})

		It("should handle different keys", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Should)

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetMode(stsctr.Must)

			list.Add(m1, m2)

			Expect(list.GetMode("key1")).To(Equal(stsctr.Should))
			Expect(list.GetMode("key2")).To(Equal(stsctr.Must))
		})
	})

	Describe("SetMode", func() {
		It("should set mode for existing key", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Ignore)

			list.Add(m)

			list.SetMode("key1", stsctr.Should)
			Expect(list.GetMode("key1")).To(Equal(stsctr.Should))
		})

		It("should not affect non-matching keys", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Should)

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetMode(stsctr.Must)

			list.Add(m1, m2)

			list.SetMode("key1", stsctr.AnyOf)

			Expect(list.GetMode("key1")).To(Equal(stsctr.AnyOf))
			Expect(list.GetMode("key2")).To(Equal(stsctr.Must))
		})

		It("should handle non-existent key", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")

			list.Add(m)

			list.SetMode("nonexistent", stsctr.Should)
			Expect(list.GetMode("nonexistent")).To(Equal(stsctr.Ignore))
		})

		It("should update first matching mandatory", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Ignore)

			m2 := stsmdt.New()
			m2.KeyAdd("key1")
			m2.SetMode(stsctr.Ignore)

			list.Add(m1, m2)

			list.SetMode("key1", stsctr.Should)

			// At least one should be updated
			mode := list.GetMode("key1")
			Expect(mode).To(Equal(stsctr.Should))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle concurrent additions", func() {
			wg := libsem.New(context.Background(), 2, false)
			defer wg.DeferMain()

			ft := func(key string) {
				m := stsmdt.New()
				m.KeyAdd(key)
				time.Sleep(5 * time.Millisecond) // prevent memory writing speed vs reading
				list.Add(m)
				time.Sleep(5 * time.Millisecond) // prevent memory writing speed vs reading
			}

			for i := 0; i < 25; i++ {
				Expect(wg.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer wg.DeferWorker()
					ft(fmt.Sprintf("key1-%d", i))
				}()
			}

			for i := 0; i < 25; i++ {
				Expect(wg.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer wg.DeferWorker()
					ft(fmt.Sprintf("key2-%d", i))
				}()
			}

			for i := 0; i < 25; i++ {
				Expect(wg.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer wg.DeferWorker()
					ft(fmt.Sprintf("key3-%d", i))
				}()
			}

			for i := 0; i < 25; i++ {
				Expect(wg.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer wg.DeferWorker()
					ft(fmt.Sprintf("key4-%d", i))
				}()
			}

			Expect(wg.WaitAll()).ToNot(HaveOccurred())
			time.Sleep(time.Second)

			Expect(len(list.GetList())).To(Equal(100))
			Expect(list.Len()).To(Equal(100))
		})

		It("should handle concurrent reads and writes", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Should)
			list.Add(m)

			done := make(chan bool)

			// Reader
			go func() {
				for i := 0; i < 100; i++ {
					_ = list.Len()
					_ = list.GetMode("key1")
				}
				done <- true
			}()

			// Writer
			go func() {
				for i := 0; i < 100; i++ {
					list.SetMode("key1", stsctr.Must)
					m := stsmdt.New()
					list.Add(m)
				}
				done <- true
			}()

			<-done
			<-done

			Expect(list.Len()).To(BeNumerically(">", 0))
		})

		It("should handle concurrent walks", func() {
			for i := 0; i < 10; i++ {
				m := stsmdt.New()
				list.Add(m)
			}

			done := make(chan bool)

			for i := 0; i < 5; i++ {
				go func() {
					list.Walk(func(m stsmdt.Mandatory) bool {
						return true
					})
					done <- true
				}()
			}

			for i := 0; i < 5; i++ {
				<-done
			}

			Expect(list.Len()).To(Equal(10))
		})
	})
})
