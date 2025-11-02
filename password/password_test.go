/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package password_test

import (
	"regexp"

	. "github.com/nabbar/golib/password"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("password", func() {
	It("Generate should create password of specified length", func() {
		lengths := []int{1, 8, 16, 32, 64, 128}
		for _, n := range lengths {
			pwd := Generate(n)
			Expect(len(pwd)).To(Equal(n))
		}
	})

	It("Generate should use characters from LetterBytes", func() {
		pwd := Generate(100)
		// All characters should be from the LetterBytes constant
		validChars := regexp.MustCompile(`^[a-zA-Z0-9,;:!?.\/*%^$&"'(\-_)=+~#{[\|` + "`" + `\\^@\]}]+$`)
		Expect(validChars.MatchString(pwd)).To(BeTrue())
	})

	It("Generate should create different passwords on multiple calls", func() {
		pwd1 := Generate(32)
		pwd2 := Generate(32)
		pwd3 := Generate(32)

		// With 93 possible characters and 32 length, collision is extremely unlikely
		Expect(pwd1).ToNot(Equal(pwd2))
		Expect(pwd2).ToNot(Equal(pwd3))
		Expect(pwd1).ToNot(Equal(pwd3))
	})

	It("Generate should handle edge cases", func() {
		// Zero length
		pwd0 := Generate(0)
		Expect(len(pwd0)).To(Equal(0))

		// Very large password
		pwd := Generate(1000)
		Expect(len(pwd)).To(Equal(1000))
	})

	It("Generate should produce non-empty passwords for positive lengths", func() {
		for i := 1; i <= 20; i++ {
			pwd := Generate(i)
			Expect(pwd).ToNot(BeEmpty())
			Expect(len(pwd)).To(Equal(i))
		}
	})

	It("randIdx should return valid indices within LetterBytes", func() {
		// Test indirectly by generating many short passwords
		for i := 0; i < 100; i++ {
			pwd := Generate(10)
			for _, c := range pwd {
				// Each character must be in LetterBytes
				Expect(LetterBytes).To(ContainSubstring(string(c)))
			}
		}
	})
})
