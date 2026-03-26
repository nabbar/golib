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

package info_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/info"
)

// This suite tests the manual data manipulation methods of the Info component,
// such as SetName, SetData, AddData, and DelData. It also covers edge cases
// like nil receivers and unregistering functions.
var _ = Describe("Manual Data Manipulation", func() {
	var i info.Info

	BeforeEach(func() {
		var err error
		i, err = info.New("manual-test")
		Expect(err).NotTo(HaveOccurred())
	})

	// Tests for the SetName method.
	Context("SetName", func() {
		// TC-MAN-001
		It("should set a new name", func() {
			i.SetName("new-name")
			Expect(i.Name()).To(Equal("new-name"))
		})

		// TC-MAN-002
		It("should revert to default name if empty string is set", func() {
			i.SetName("new-name")
			Expect(i.Name()).To(Equal("new-name"))
			i.SetName("")
			Expect(i.Name()).To(Equal("manual-test"))
		})
	})

	// Tests for the SetData method.
	Context("SetData", func() {
		// TC-MAN-003
		It("should replace all existing data", func() {
			i.SetData(map[string]interface{}{"a": 1})
			i.SetData(map[string]interface{}{"b": 2})
			Expect(i.Data()).To(Equal(map[string]interface{}{"b": 2}))
		})

		// TC-MAN-004
		It("should handle nil and empty maps", func() {
			i.SetData(map[string]interface{}{"a": 1})
			i.SetData(nil)
			Expect(i.Data()).To(BeEmpty())

			i.SetData(map[string]interface{}{"a": 1})
			i.SetData(map[string]interface{}{})
			Expect(i.Data()).To(BeEmpty())
		})
	})

	// Tests for the AddData method.
	Context("AddData", func() {
		// TC-MAN-005
		It("should add new data or update existing", func() {
			i.AddData("a", 1)
			Expect(i.Data()).To(HaveKeyWithValue("a", 1))
			i.AddData("a", 2)
			Expect(i.Data()).To(HaveKeyWithValue("a", 2))
		})

		// TC-MAN-006
		It("should delete data if value is nil", func() {
			i.AddData("a", 1)
			i.AddData("a", nil)
			Expect(i.Data()).NotTo(HaveKey("a"))
		})

		// TC-MAN-007
		It("should ignore empty key", func() {
			i.AddData("", 1)
			Expect(i.Data()).To(BeEmpty())
		})
	})

	// Tests for the DelData method.
	Context("DelData", func() {
		// TC-MAN-008
		It("should delete data by key", func() {
			i.AddData("a", 1)
			i.AddData("b", 2)
			i.DelData("a")
			Expect(i.Data()).NotTo(HaveKey("a"))
			Expect(i.Data()).To(HaveKey("b"))
		})

		// TC-MAN-009
		It("should ignore empty key", func() {
			i.AddData("a", 1)
			i.DelData("")
			Expect(i.Data()).To(HaveKey("a"))
		})
	})

	// Tests for unregistering a name function by passing nil.
	Context("RegisterName with nil", func() {
		// TC-MAN-010
		It("should unregister name function", func() {
			i.RegisterName(func() (string, error) {
				return "dynamic", nil
			})
			Expect(i.Name()).To(Equal("dynamic"))

			i.RegisterName(nil)
			// Since caching is disabled, Name() should fallback to default.
			Expect(i.Name()).To(Equal("manual-test"))
		})
	})

	// Tests for unregistering an info function by passing nil.
	Context("RegisterData with nil", func() {
		// TC-MAN-011
		It("should unregister info function", func() {
			i.RegisterData(func() (map[string]interface{}, error) {
				return map[string]interface{}{"k": "v"}, nil
			})
			Expect(i.Data()).To(HaveKey("k"))

			i.RegisterData(nil)
			Expect(i.Data()).To(BeEmpty())
		})
	})

	// This context tests the behavior of methods when called on a nil receiver.
	// It uses reflection to create a nil pointer to the internal struct type
	// without breaking the black-box testing approach.
	Context("Nil Receiver Checks", func() {
		var nilInfo info.Info

		BeforeEach(func() {
			// Create a nil pointer to the internal struct type using reflection.
			// This allows testing the nil-guard clauses in the methods.
			var i info.Info
			var err error
			i, err = info.New("temp")
			Expect(err).NotTo(HaveOccurred())

			// Get the type of the implementation (*inf).
			t := reflect.TypeOf(i)
			// Create a zero value of that type (which is a nil pointer).
			v := reflect.Zero(t)
			// Convert back to the interface. nilInfo now holds a nil *inf.
			nilInfo = v.Interface().(info.Info)
		})

		// TC-MAN-012
		It("should handle Name() on nil instance", func() {
			Expect(nilInfo.Name()).To(BeEmpty())
		})

		// TC-MAN-013
		It("should handle data() on nil instance", func() {
			Expect(nilInfo.Data()).To(BeNil())
		})

		// TC-MAN-014
		It("should handle SetName() on nil instance", func() {
			Expect(func() { nilInfo.SetName("test") }).ShouldNot(Panic())
		})

		// TC-MAN-015
		It("should handle SetData() on nil instance", func() {
			Expect(func() { nilInfo.SetData(nil) }).ShouldNot(Panic())
		})

		// TC-MAN-016
		It("should handle AddData() on nil instance", func() {
			Expect(func() { nilInfo.AddData("key", "val") }).ShouldNot(Panic())
		})

		// TC-MAN-017
		It("should handle DelData() on a nil instance", func() {
			Expect(func() { nilInfo.DelData("key") }).ShouldNot(Panic())
		})

		// TC-MAN-018
		It("should handle RegisterName() on nil instance", func() {
			Expect(func() { nilInfo.RegisterName(nil) }).ShouldNot(Panic())
		})

		// TC-MAN-019
		It("should handle RegisterData() on nil instance", func() {
			Expect(func() { nilInfo.RegisterData(nil) }).ShouldNot(Panic())
		})
	})
})
