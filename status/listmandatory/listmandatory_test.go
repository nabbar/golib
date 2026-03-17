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
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("group2")

			list := listmandatory.New(m1, m2)
			Expect(list.Len()).To(Equal(2))
		})

		It("should ignore nil or empty mandatory groups during initialization", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetName("group1")
			m1.SetMode(stsctr.Should)

			emptyM := stsmdt.New() // No keys added

			list := listmandatory.New(m1, nil, emptyM)
			Expect(list.Len()).To(Equal(1))
			Expect(list.GetMode("key1")).ToNot(Equal(stsctr.Ignore))
		})
	})

	Describe("Len", func() {
		It("should return 0 for empty list", func() {
			Expect(list.Len()).To(Equal(0))
		})

		It("should return correct length after adding", func() {
			m := stsmdt.New()
			m.KeyAdd("key")
			list.Add(m)
			Expect(list.Len()).To(Equal(1))
		})
	})

	Describe("Add", func() {
		It("should add a single mandatory", func() {
			m := stsmdt.New()
			m.KeyAdd("key")
			list.Add(m)
			Expect(list.Len()).To(Equal(1))
		})

		It("should add multiple mandatories", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("group2")

			list.Add(m1, m2)
			Expect(list.Len()).To(Equal(2))
		})

		It("should not add nil mandatory groups", func() {
			list.Add(nil)
			Expect(list.Len()).To(Equal(0))

			m := stsmdt.New()
			m.KeyAdd("key1")
			list.Add(m, nil)
			Expect(list.Len()).To(Equal(1))
		})

		It("should not add mandatory groups with no keys", func() {
			emptyM := stsmdt.New() // No keys added
			list.Add(emptyM)
			Expect(list.Len()).To(Equal(0))

			m := stsmdt.New()
			m.KeyAdd("key1")
			list.Add(m, emptyM)
			Expect(list.Len()).To(Equal(1))
		})

		Context("when dealing with names", func() {
			It("should overwrite a mandatory if the name is the same", func() {
				m1 := stsmdt.New()
				m1.SetName("group-a")
				m1.KeyAdd("key1")
				m1.SetMode(stsctr.Must) // Set a non-Ignore mode for key1
				list.Add(m1)
				Expect(list.Len()).To(Equal(1))
				Expect(list.GetMode("key1")).To(Equal(stsctr.Must))

				m2 := stsmdt.New()
				m2.SetName("group-a")
				m2.KeyAdd("key2")         // Different key, same name
				m2.SetMode(stsctr.Should) // Set a non-Ignore mode for key2
				list.Add(m2)

				Expect(list.Len()).To(Equal(1))                       // Length should still be 1
				Expect(list.GetMode("key1")).To(Equal(stsctr.Ignore)) // Old key should be gone
				Expect(list.GetMode("key2")).To(Equal(stsctr.Should)) // New key should be present with its mode
			})

			It("should add as new if the name is different", func() {
				m1 := stsmdt.New()
				m1.SetName("group-a")
				m1.KeyAdd("key1")
				list.Add(m1)

				m2 := stsmdt.New()
				m2.SetName("group-b")
				m2.KeyAdd("key1") // Same key, different name
				list.Add(m2)

				Expect(list.Len()).To(Equal(2))
			})

			It("should generate a default name if none is provided", func() {
				m1 := stsmdt.New()
				m1.KeyAdd("key1")
				list.Add(m1) // No name

				m2 := stsmdt.New()
				m2.KeyAdd("key2")
				list.Add(m2) // No name

				// Assuming default names are unique based on content or a counter
				Expect(list.Len()).To(Equal(2))
			})
		})
	})

	Describe("Del (by content)", func() {
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
			m1.SetName("group1")
			m1.SetMode(stsctr.Must) // Set a non-Ignore mode

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("group2")
			m2.SetMode(stsctr.Should) // Set a non-Ignore mode

			list.Add(m1, m2)
			Expect(list.Len()).To(Equal(2))

			list.Del(m1)
			Expect(list.Len()).To(Equal(1))
			Expect(list.GetMode("key1")).To(Equal(stsctr.Ignore)) // key1 should be gone
			Expect(list.GetMode("key2")).To(Equal(stsctr.Should)) // key2 should still be there
		})

		It("should match by key list content, ignoring order", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1", "key2")
			list.Add(m1)

			m2 := stsmdt.New()
			m2.KeyAdd("key2", "key1") // Same keys, different order

			list.Del(m2)
			Expect(list.Len()).To(Equal(0))
		})
	})

	Describe("DelKey (by name)", func() {
		var m1, m2 stsmdt.Mandatory

		BeforeEach(func() {
			m1 = stsmdt.New()
			m1.SetName("group-one")
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Must) // Set a non-Ignore mode for m1

			m2 = stsmdt.New()
			m2.SetName("group-two")
			m2.KeyAdd("key2")
			m2.SetMode(stsctr.Should) // Set a non-Ignore mode for m2

			list.Add(m1, m2)
			Expect(list.Len()).To(Equal(2))
		})

		It("should delete a mandatory by its name", func() {
			list.DelKey("group-one")
			Expect(list.Len()).To(Equal(1))
			Expect(list.GetMode("key1")).To(Equal(stsctr.Ignore)) // key1 should be gone
			Expect(list.GetMode("key2")).To(Equal(stsctr.Should)) // key2 should still be there
		})

		It("should not fail when deleting a non-existent name", func() {
			list.DelKey("non-existent-group")
			Expect(list.Len()).To(Equal(2))
		})
	})

	Describe("Walk", func() {
		BeforeEach(func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("group2")

			m3 := stsmdt.New()
			m3.KeyAdd("key3")
			m3.SetName("group3")

			list.Add(m1, m2, m3)
		})

		It("should walk through all mandatories", func() {
			count := 0
			var names []string
			list.Walk(func(k string, m stsmdt.Mandatory) bool {
				count++
				names = append(names, k)
				Expect(m).ToNot(BeNil())
				return true
			})

			Expect(count).To(Equal(3))
			Expect(names).To(ContainElements("group1", "group2", "group3"))
		})

		It("should stop walking when function returns false", func() {
			count := 0
			list.Walk(func(_ string, m stsmdt.Mandatory) bool {
				count++
				return count < 2 // Stop after 2
			})

			Expect(count).To(Equal(2))
		})

		It("should allow modifying mandatories during walk", func() {
			list.Walk(func(k string, m stsmdt.Mandatory) bool {
				if k == "group1" {
					m.SetMode(stsctr.Should)
				}
				return true
			})

			// Verify the change persisted
			mode := list.GetMode("key1")
			Expect(mode).To(Equal(stsctr.Should))
		})
	})

	Describe("GetList", func() {
		It("should return an empty slice for an empty list", func() {
			Expect(list.GetList()).To(BeEmpty())
		})

		It("should return all added mandatories", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetName("m1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("m2")

			list.Add(m1, m2)

			result := list.GetList()
			Expect(len(result)).To(Equal(2))

			var names []string
			for _, m := range result {
				names = append(names, m.GetName())
			}
			Expect(names).To(ContainElements("m1", "m2"))
		})

		It("should return a snapshot that is not affected by later modifications", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetName("m1")
			list.Add(m1)

			snapshot := list.GetList()
			Expect(len(snapshot)).To(Equal(1))

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetName("m2")
			list.Add(m2)

			Expect(list.Len()).To(Equal(2))
			Expect(len(snapshot)).To(Equal(1)) // Snapshot should be unchanged
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
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key1")
			m2.SetMode(stsctr.Must)
			m2.SetName("group2")

			list.Add(m1, m2)

			mode := list.GetMode("key1")
			// The order is not guaranteed, so it could be either
			Expect(mode).To(BeElementOf(stsctr.Should, stsctr.Must))
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
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key2")
			m2.SetMode(stsctr.Must)
			m2.SetName("group2")

			list.Add(m1, m2)
			list.SetMode("key1", stsctr.AnyOf)

			Expect(list.GetMode("key1")).To(Equal(stsctr.AnyOf))
			Expect(list.GetMode("key2")).To(Equal(stsctr.Must))
		})

		It("should update first matching mandatory only", func() {
			m1 := stsmdt.New()
			m1.KeyAdd("key1")
			m1.SetMode(stsctr.Ignore)
			m1.SetName("group1")

			m2 := stsmdt.New()
			m2.KeyAdd("key1")
			m2.SetMode(stsctr.Ignore)
			m2.SetName("group2")

			list.Add(m1, m2)
			list.SetMode("key1", stsctr.Should)

			// Check that one was updated
			mode := list.GetMode("key1")
			Expect(mode).To(Equal(stsctr.Should))

			// This is tricky to test without knowing which one was updated.
			// We can walk and check that only one has been updated.
			var updatedModes []stsctr.Mode
			list.Walk(func(k string, m stsmdt.Mandatory) bool {
				if m.KeyHas("key1") {
					updatedModes = append(updatedModes, m.GetMode())
				}
				return true
			})

			Expect(updatedModes).To(ContainElement(stsctr.Should))
			Expect(updatedModes).To(ContainElement(stsctr.Ignore))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle concurrent additions", func() {
			wg := libsem.New(context.Background(), 10, false)
			defer wg.DeferMain()

			for i := 0; i < 100; i++ {
				Expect(wg.NewWorker()).ToNot(HaveOccurred())
				go func(i int) {
					defer wg.DeferWorker()
					m := stsmdt.New()
					m.KeyAdd(fmt.Sprintf("key-%d", i))
					m.SetName(fmt.Sprintf("group-%d", i))
					list.Add(m)
				}(i)
			}

			Expect(wg.WaitAll()).ToNot(HaveOccurred())
			Expect(list.Len()).To(Equal(100))
		})

		It("should handle concurrent reads and writes", func() {
			m := stsmdt.New()
			m.KeyAdd("key1")
			m.SetMode(stsctr.Should)
			m.SetName("group1")
			list.Add(m)

			done := make(chan bool)

			// Reader
			go func() {
				for i := 0; i < 100; i++ {
					_ = list.Len()
					_ = list.GetMode("key1")
					time.Sleep(1 * time.Millisecond)
				}
				done <- true
			}()

			// Writer
			go func() {
				for i := 0; i < 100; i++ {
					list.SetMode("key1", stsctr.Must)
					m := stsmdt.New()
					m.SetName(fmt.Sprintf("new-group-%d", i))
					m.KeyAdd(fmt.Sprintf("dynamic-key-%d", i)) // Added a key here
					list.Add(m)
					time.Sleep(1 * time.Millisecond)
				}
				done <- true
			}()

			<-done
			<-done

			Expect(list.Len()).To(BeNumerically(">", 100))
		})
	})
})
