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

package mandatory_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	stsmdt "github.com/nabbar/golib/status/mandatory"
)

var _ = Describe("Tools", func() {

	Describe("GetDefaultName", func() {
		It("should return a string starting with mandatory-", func() {
			n := stsmdt.GetDefaultName()
			Expect(n).To(HavePrefix("mandatory-"))
		})

		It("should return unique names on subsequent calls", func() {
			n1 := stsmdt.GetDefaultName()
			n2 := stsmdt.GetDefaultName()
			Expect(n1).ToNot(Equal(n2))
		})
	})

	Describe("FilterName", func() {
		It("should return unchanged string for valid characters", func() {
			s := "valid-name_123"
			Expect(stsmdt.FilterName(s)).To(Equal(s))
		})

		It("should remove uppercase letters", func() {
			// Based on rgx `[^a-z0-9\-_]+`, uppercase are invalid.
			// FilterName does NOT automatically lowercase.
			s := "InvalidName"
			Expect(stsmdt.FilterName(s)).To(Equal("nvalidame"))
		})

		It("should remove special characters", func() {
			s := "invalid!@#name"
			Expect(stsmdt.FilterName(s)).To(Equal("invalidname"))
		})

		It("should return empty string if all characters are invalid", func() {
			s := "!@#$%"
			Expect(stsmdt.FilterName(s)).To(BeEmpty())
		})
	})

	Describe("GetNameOrDefault", func() {
		It("should return the sanitized name if valid", func() {
			s := "My-Name_123"
			// GetNameOrDefault calls strings.ToLower first, then FilterName.
			// "My-Name_123" -> "my-name_123" -> "my-name_123"
			Expect(stsmdt.GetNameOrDefault(s)).To(Equal("my-name_123"))
		})

		It("should return default name if input is empty", func() {
			Expect(stsmdt.GetNameOrDefault("")).To(HavePrefix("mandatory-"))
		})

		It("should return default name if input becomes empty after sanitization", func() {
			// "!!!", ToLower -> "!!!", FilterName -> ""
			Expect(stsmdt.GetNameOrDefault("!!!")).To(HavePrefix("mandatory-"))
		})
	})
})
