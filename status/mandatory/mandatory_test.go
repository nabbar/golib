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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	stsctr "github.com/nabbar/golib/status/control"
	stsmdt "github.com/nabbar/golib/status/mandatory"
)

var _ = Describe("Mandatory", func() {
	var m stsmdt.Mandatory

	BeforeEach(func() {
		m = stsmdt.New()
	})

	Describe("New", func() {
		It("should create a new mandatory instance", func() {
			Expect(m).ToNot(BeNil())
		})

		It("should initialize with Ignore mode", func() {
			Expect(m.GetMode()).To(Equal(stsctr.Ignore))
		})

		It("should initialize with empty keys", func() {
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should initialize with empty name", func() {
			Expect(m.GetName()).To(BeEmpty())
		})

		It("should initialize with empty info", func() {
			Expect(m.GetInfo()).To(BeEmpty())
		})
	})

	Describe("SetMode/GetMode", func() {
		It("should set and get Should mode", func() {
			m.SetMode(stsctr.Should)
			Expect(m.GetMode()).To(Equal(stsctr.Should))
		})

		It("should set and get Must mode", func() {
			m.SetMode(stsctr.Must)
			Expect(m.GetMode()).To(Equal(stsctr.Must))
		})

		It("should set and get AnyOf mode", func() {
			m.SetMode(stsctr.AnyOf)
			Expect(m.GetMode()).To(Equal(stsctr.AnyOf))
		})

		It("should set and get Quorum mode", func() {
			m.SetMode(stsctr.Quorum)
			Expect(m.GetMode()).To(Equal(stsctr.Quorum))
		})

		It("should set and get Ignore mode", func() {
			m.SetMode(stsctr.Must)
			m.SetMode(stsctr.Ignore)
			Expect(m.GetMode()).To(Equal(stsctr.Ignore))
		})

		It("should allow multiple mode changes", func() {
			m.SetMode(stsctr.Should)
			Expect(m.GetMode()).To(Equal(stsctr.Should))

			m.SetMode(stsctr.Must)
			Expect(m.GetMode()).To(Equal(stsctr.Must))

			m.SetMode(stsctr.AnyOf)
			Expect(m.GetMode()).To(Equal(stsctr.AnyOf))
		})
	})

	Describe("SetName/GetName", func() {
		It("should set and get a valid name", func() {
			name := "my-valid-group-1"
			m.SetName(name)
			Expect(m.GetName()).To(Equal(name))
		})

		It("should sanitize invalid characters and uppercase", func() {
			m.SetName("My Group #1")
			// "My Group #1" -> toLower: "my group #1" -> filter: "mygroup1"
			Expect(m.GetName()).To(Equal("mygroup1"))
		})

		It("should set a default name if input is empty", func() {
			m.SetName("")
			Expect(m.GetName()).To(MatchRegexp(`^mandatory-\d+$`))
		})

		It("should set a default name if input becomes empty after sanitization", func() {
			m.SetName("!!!")
			Expect(m.GetName()).To(MatchRegexp(`^mandatory-\d+$`))
		})

		It("should overwrite existing name", func() {
			m.SetName("first-name")
			Expect(m.GetName()).To(Equal("first-name"))

			m.SetName("second-name")
			Expect(m.GetName()).To(Equal("second-name"))
		})
	})

	Describe("Info Metadata", func() {
		It("should set and get info map", func() {
			info := map[string]interface{}{
				"description": "test description",
				"version":     1,
			}
			m.SetInfo(info)
			Expect(m.GetInfo()).To(Equal(info))
		})

		It("should add single info entry", func() {
			m.AddInfo("author", "tester")
			info := m.GetInfo()
			Expect(info).To(HaveKeyWithValue("author", "tester"))
		})

		It("should overwrite existing info with SetInfo", func() {
			m.SetInfo(map[string]interface{}{"key1": "value1"})
			m.SetInfo(map[string]interface{}{"key2": "value2"})
			info := m.GetInfo()
			Expect(info).To(HaveKey("key1")) // SetInfo merges/adds, it doesn't clear the map in current implementation?
			// Checking implementation logic: SetInfo iterates and stores. It doesn't clear previous.
			// Let's verify this assumption with a test.
			Expect(info).To(HaveKey("key2"))
		})

		It("should overwrite existing key with AddInfo", func() {
			m.AddInfo("key", "value1")
			m.AddInfo("key", "value2")
			Expect(m.GetInfo()).To(HaveKeyWithValue("key", "value2"))
		})

		It("should ignore empty keys in SetInfo", func() {
			m.SetInfo(map[string]interface{}{"": "value"})
			Expect(m.GetInfo()).To(BeEmpty())
		})

		It("should ignore nil values in SetInfo", func() {
			m.SetInfo(map[string]interface{}{"key": nil})
			Expect(m.GetInfo()).To(BeEmpty())
		})

		It("should ignore empty key in AddInfo", func() {
			m.AddInfo("", "value")
			Expect(m.GetInfo()).To(BeEmpty())
		})

		It("should ignore nil value in AddInfo", func() {
			m.AddInfo("key", nil)
			Expect(m.GetInfo()).To(BeEmpty())
		})

		It("should handle mixed types in info", func() {
			m.AddInfo("string", "val")
			m.AddInfo("int", 123)
			m.AddInfo("bool", true)
			info := m.GetInfo()
			Expect(info["string"]).To(Equal("val"))
			Expect(info["int"]).To(Equal(123))
			Expect(info["bool"]).To(BeTrue())
		})
	})

	Describe("KeyAdd", func() {
		It("should add a single key", func() {
			m.KeyAdd("key1")
			Expect(m.KeyList()).To(ContainElement("key1"))
			Expect(m.KeyList()).To(HaveLen(1))
		})

		It("should add multiple keys", func() {
			m.KeyAdd("key1", "key2", "key3")
			Expect(m.KeyList()).To(ContainElement("key1"))
			Expect(m.KeyList()).To(ContainElement("key2"))
			Expect(m.KeyList()).To(ContainElement("key3"))
			Expect(m.KeyList()).To(HaveLen(3))
		})

		It("should not add duplicate keys", func() {
			m.KeyAdd("key1")
			m.KeyAdd("key1")
			Expect(m.KeyList()).To(HaveLen(1))
		})

		It("should handle adding duplicate keys in same call", func() {
			m.KeyAdd("key1", "key2", "key1", "key3", "key2")
			Expect(m.KeyList()).To(HaveLen(3))
			Expect(m.KeyList()).To(ContainElement("key1"))
			Expect(m.KeyList()).To(ContainElement("key2"))
			Expect(m.KeyList()).To(ContainElement("key3"))
		})

		It("should add keys incrementally", func() {
			m.KeyAdd("key1")
			Expect(m.KeyList()).To(HaveLen(1))

			m.KeyAdd("key2")
			Expect(m.KeyList()).To(HaveLen(2))

			m.KeyAdd("key3")
			Expect(m.KeyList()).To(HaveLen(3))
		})

		It("should handle empty string keys", func() {
			m.KeyAdd("")
			Expect(m.KeyList()).To(ContainElement(""))
			Expect(m.KeyList()).To(HaveLen(1))
		})
	})

	Describe("KeyHas", func() {
		BeforeEach(func() {
			m.KeyAdd("key1", "key2", "key3")
		})

		It("should return true for existing key", func() {
			Expect(m.KeyHas("key1")).To(BeTrue())
			Expect(m.KeyHas("key2")).To(BeTrue())
			Expect(m.KeyHas("key3")).To(BeTrue())
		})

		It("should return false for non-existing key", func() {
			Expect(m.KeyHas("key4")).To(BeFalse())
			Expect(m.KeyHas("nonexistent")).To(BeFalse())
		})

		It("should return false for empty keys list", func() {
			m2 := stsmdt.New()
			Expect(m2.KeyHas("key1")).To(BeFalse())
		})

		It("should be case-sensitive", func() {
			Expect(m.KeyHas("Key1")).To(BeFalse())
			Expect(m.KeyHas("KEY1")).To(BeFalse())
		})
	})

	Describe("KeyDel", func() {
		BeforeEach(func() {
			m.KeyAdd("key1", "key2", "key3", "key4")
		})

		It("should delete a single key", func() {
			m.KeyDel("key1")
			Expect(m.KeyList()).ToNot(ContainElement("key1"))
			Expect(m.KeyList()).To(HaveLen(3))
		})

		It("should delete multiple keys", func() {
			m.KeyDel("key1", "key2")
			Expect(m.KeyList()).ToNot(ContainElement("key1"))
			Expect(m.KeyList()).ToNot(ContainElement("key2"))
			Expect(m.KeyList()).To(HaveLen(2))
		})

		It("should handle deleting non-existing key", func() {
			m.KeyDel("nonexistent")
			Expect(m.KeyList()).To(HaveLen(4))
		})

		It("should handle deleting all keys", func() {
			m.KeyDel("key1", "key2", "key3", "key4")
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should handle deleting from empty list", func() {
			m2 := stsmdt.New()
			m2.KeyDel("key1")
			Expect(m2.KeyList()).To(BeEmpty())
		})

		It("should preserve remaining keys", func() {
			m.KeyDel("key2")
			Expect(m.KeyList()).To(ContainElement("key1"))
			Expect(m.KeyList()).To(ContainElement("key3"))
			Expect(m.KeyList()).To(ContainElement("key4"))
		})
	})

	Describe("KeyList", func() {
		It("should return empty list initially", func() {
			Expect(m.KeyList()).To(BeEmpty())
		})

		It("should return all added keys", func() {
			m.KeyAdd("key1", "key2", "key3")
			list := m.KeyList()
			Expect(list).To(HaveLen(3))
			Expect(list).To(ContainElement("key1"))
			Expect(list).To(ContainElement("key2"))
			Expect(list).To(ContainElement("key3"))
		})

		It("should return a copy of the list", func() {
			m.KeyAdd("key1", "key2")
			list1 := m.KeyList()
			list2 := m.KeyList()

			// Modifying one list should not affect the other
			list1 = append(list1, "key3")
			Expect(list2).To(HaveLen(2))
			Expect(m.KeyList()).To(HaveLen(2))
		})

		It("should reflect changes after add", func() {
			m.KeyAdd("key1")
			Expect(m.KeyList()).To(HaveLen(1))

			m.KeyAdd("key2")
			Expect(m.KeyList()).To(HaveLen(2))
		})

		It("should reflect changes after delete", func() {
			m.KeyAdd("key1", "key2", "key3")
			Expect(m.KeyList()).To(HaveLen(3))

			m.KeyDel("key2")
			Expect(m.KeyList()).To(HaveLen(2))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle concurrent mode changes", func() {
			done := make(chan bool)

			go func() {
				for i := 0; i < 100; i++ {
					m.SetMode(stsctr.Should)
				}
				done <- true
			}()

			go func() {
				for i := 0; i < 100; i++ {
					m.SetMode(stsctr.Must)
				}
				done <- true
			}()

			<-done
			<-done

			mode := m.GetMode()
			Expect(mode).To(BeElementOf(stsctr.Should, stsctr.Must))
		})

		It("should handle concurrent key additions", func() {
			done := make(chan bool)

			go func() {
				for i := 0; i < 50; i++ {
					m.KeyAdd("key1")
				}
				done <- true
			}()

			go func() {
				for i := 0; i < 50; i++ {
					m.KeyAdd("key2")
				}
				done <- true
			}()

			<-done
			<-done

			list := m.KeyList()
			Expect(list).To(ContainElement("key1"))
			Expect(list).To(ContainElement("key2"))
		})

		It("should handle concurrent reads and writes", func() {
			m.KeyAdd("key1", "key2", "key3")
			done := make(chan bool)

			// Reader
			go func() {
				for i := 0; i < 100; i++ {
					_ = m.KeyList()
					_ = m.KeyHas("key1")
					_ = m.GetMode()
					_ = m.GetInfo()
				}
				done <- true
			}()

			// Writer
			go func() {
				for i := 0; i < 100; i++ {
					m.KeyAdd("key4")
					m.SetMode(stsctr.Should)
					m.AddInfo("status", "running")
				}
				done <- true
			}()

			<-done
			<-done

			Expect(m.KeyList()).ToNot(BeEmpty())
		})
	})
})
