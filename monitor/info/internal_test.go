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

package info

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These are white-box tests that access internal structure
// to test edge cases that are hard to trigger through the public API

var _ = Describe("Internal State Edge Cases", func() {
	Describe("Corrupted internal state scenarios", func() {
		Context("with non-string key in sync.Map", func() {
			It("should handle non-string keys in Name()", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				// Cast to internal type to access fields
				impl := i.(*inf)

				// Store a non-string key to test the type assertion path
				impl.v.Store(123, "non-string-key")
				impl.v.Store(keyName, "valid-name")

				// Name should handle this and return the valid name
				name := i.Name()
				Expect(name).To(Equal("valid-name"))
			})

			It("should handle non-string value for keyName", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Store a non-string value for keyName
				impl.v.Store(keyName, 12345)

				// Should fall back to getName() which returns default
				name := i.Name()
				Expect(name).To(Equal("test"))
			})

			It("should handle non-string keys in Info()", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Store non-string keys
				impl.v.Store(123, "integer-key")
				impl.v.Store(true, "bool-key")
				impl.v.Store("valid", "valid-value")

				// Info should skip non-string keys and internal keys
				info := i.Info()
				Expect(info).NotTo(BeNil())
				Expect(info).To(HaveKey("valid"))
				Expect(info["valid"]).To(Equal("valid-value"))
				// Non-string keys should be skipped
				Expect(info).NotTo(HaveKey(123))
				Expect(info).NotTo(HaveKey(true))
			})

			It("should handle non-string keys in RegisterInfo", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Populate with some non-string keys
				impl.v.Store(456, "another-integer")
				impl.v.Store("oldkey", "oldvalue")

				// RegisterInfo should clean up non-internal keys
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"new": "data"}, nil
				})

				// After registration, non-string keys should be removed
				info := i.Info()
				Expect(info).NotTo(BeNil())
				Expect(info["new"]).To(Equal("data"))
			})
		})

		Context("with missing or corrupted default name", func() {
			It("should handle missing default name", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Delete the default name
				impl.v.Delete(keyDefName)

				// defaultName() should return empty string
				name := impl.defaultName()
				Expect(name).To(Equal(""))
			})

			It("should handle non-string default name", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Store non-string as default name
				impl.v.Store(keyDefName, 999)

				// defaultName() should return empty string
				name := impl.defaultName()
				Expect(name).To(Equal(""))
			})
		})

		Context("with internal key filtering", func() {
			It("should filter out internal keys from Info()", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Store various keys including internal ones
				impl.v.Store(keyDefName, "default")
				impl.v.Store(keyName, "name")
				impl.v.Store("public", "public-value")

				// Info should only return public keys
				info := i.Info()
				Expect(info).NotTo(BeNil())
				Expect(info).To(HaveLen(1))
				Expect(info).To(HaveKey("public"))
				Expect(info).NotTo(HaveKey(keyDefName))
				Expect(info).NotTo(HaveKey(keyName))
			})

			It("should preserve internal keys during RegisterInfo cleanup", func() {
				i, err := New("test")
				Expect(err).NotTo(HaveOccurred())

				impl := i.(*inf)

				// Set up initial state
				impl.v.Store(keyDefName, "default")
				impl.v.Store(keyName, "cached-name")
				impl.v.Store("old", "old-value")

				// Register new info
				i.RegisterInfo(func() (map[string]interface{}, error) {
					return map[string]interface{}{"new": "new-value"}, nil
				})

				// Internal keys should still exist
				defName, ok := impl.v.Load(keyDefName)
				Expect(ok).To(BeTrue())
				Expect(defName).To(Equal("default"))

				// Old public keys should be cleaned
				// New info should be accessible
				info := i.Info()
				Expect(info["new"]).To(Equal("new-value"))
				Expect(info).NotTo(HaveKey("old"))
			})
		})
	})
})
