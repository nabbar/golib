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
 */

package render_test

import (
	"github.com/nabbar/golib/mail/render"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Themes", func() {
	Describe("Theme Constants", func() {
		Context("when using theme constants", func() {
			It("should have correct default theme value", func() {
				Expect(render.ThemeDefault).To(BeNumerically("==", 0))
			})

			It("should have correct flat theme value", func() {
				Expect(render.ThemeFlat).To(BeNumerically("==", 1))
			})
		})
	})

	Describe("Theme String Conversion", func() {
		Context("when converting theme to string", func() {
			It("should convert default theme to string", func() {
				str := render.ThemeDefault.String()
				Expect(str).To(Equal("default"))
			})

			It("should convert flat theme to string", func() {
				str := render.ThemeFlat.String()
				Expect(str).To(Equal("flat"))
			})
		})
	})

	Describe("ParseTheme", func() {
		Context("with valid theme names", func() {
			It("should parse 'default' to ThemeDefault", func() {
				theme := render.ParseTheme("default")
				Expect(theme).To(Equal(render.ThemeDefault))
			})

			It("should parse 'Default' (mixed case) to ThemeDefault", func() {
				theme := render.ParseTheme("Default")
				Expect(theme).To(Equal(render.ThemeDefault))
			})

			It("should parse 'DEFAULT' (upper case) to ThemeDefault", func() {
				theme := render.ParseTheme("DEFAULT")
				Expect(theme).To(Equal(render.ThemeDefault))
			})

			It("should parse 'flat' to ThemeFlat", func() {
				theme := render.ParseTheme("flat")
				Expect(theme).To(Equal(render.ThemeFlat))
			})

			It("should parse 'Flat' (mixed case) to ThemeFlat", func() {
				theme := render.ParseTheme("Flat")
				Expect(theme).To(Equal(render.ThemeFlat))
			})

			It("should parse 'FLAT' (upper case) to ThemeFlat", func() {
				theme := render.ParseTheme("FLAT")
				Expect(theme).To(Equal(render.ThemeFlat))
			})
		})

		Context("with invalid theme names", func() {
			It("should default to ThemeDefault for unknown theme", func() {
				theme := render.ParseTheme("unknown")
				Expect(theme).To(Equal(render.ThemeDefault))
			})

			It("should default to ThemeDefault for empty string", func() {
				theme := render.ParseTheme("")
				Expect(theme).To(Equal(render.ThemeDefault))
			})

			It("should default to ThemeDefault for random string", func() {
				theme := render.ParseTheme("random123")
				Expect(theme).To(Equal(render.ThemeDefault))
			})
		})
	})
})

var _ = Describe("Text Direction", func() {
	Describe("Direction Constants", func() {
		Context("when using direction constants", func() {
			It("should have correct left to right value", func() {
				Expect(render.LeftToRight).To(BeNumerically("==", 0))
			})

			It("should have correct right to left value", func() {
				Expect(render.RightToLeft).To(BeNumerically("==", 1))
			})
		})
	})

	Describe("Direction String Conversion", func() {
		Context("when converting direction to string", func() {
			It("should convert LeftToRight to string", func() {
				str := render.LeftToRight.String()
				Expect(str).To(Equal("Left->Right"))
			})

			It("should convert RightToLeft to string", func() {
				str := render.RightToLeft.String()
				Expect(str).To(Equal("Right->Left"))
			})
		})
	})

	Describe("ParseTextDirection", func() {
		Context("with left-to-right variations", func() {
			It("should parse 'ltr' to LeftToRight", func() {
				dir := render.ParseTextDirection("ltr")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should parse 'left-to-right' to LeftToRight", func() {
				dir := render.ParseTextDirection("left-to-right")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should parse 'left' to LeftToRight", func() {
				dir := render.ParseTextDirection("left")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should parse 'Left->Right' to LeftToRight", func() {
				dir := render.ParseTextDirection("Left->Right")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should parse 'LTR' (upper case) to LeftToRight", func() {
				dir := render.ParseTextDirection("LTR")
				Expect(dir).To(Equal(render.LeftToRight))
			})
		})

		Context("with right-to-left variations", func() {
			It("should parse 'rtl' to RightToLeft", func() {
				dir := render.ParseTextDirection("rtl")
				Expect(dir).To(Equal(render.RightToLeft))
			})

			It("should parse 'right-to-left' to RightToLeft", func() {
				dir := render.ParseTextDirection("right-to-left")
				Expect(dir).To(Equal(render.RightToLeft))
			})

			It("should parse 'right' to RightToLeft", func() {
				dir := render.ParseTextDirection("right")
				Expect(dir).To(Equal(render.RightToLeft))
			})

			It("should parse 'Right->Left' to RightToLeft", func() {
				dir := render.ParseTextDirection("Right->Left")
				Expect(dir).To(Equal(render.RightToLeft))
			})

			It("should parse 'RTL' (upper case) to RightToLeft", func() {
				dir := render.ParseTextDirection("RTL")
				Expect(dir).To(Equal(render.RightToLeft))
			})
		})

		Context("with edge cases", func() {
			It("should default to LeftToRight for empty string", func() {
				dir := render.ParseTextDirection("")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should default to LeftToRight for unknown string", func() {
				dir := render.ParseTextDirection("unknown")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should parse correctly when both 'left' and 'right' are present", func() {
				// "right->left" should be RightToLeft (right comes first)
				dir := render.ParseTextDirection("right->left")
				Expect(dir).To(Equal(render.RightToLeft))

				// "left->right" should be LeftToRight (left comes first)
				dir = render.ParseTextDirection("left->right")
				Expect(dir).To(Equal(render.LeftToRight))
			})

			It("should handle strings with only 'right' as RightToLeft", func() {
				dir := render.ParseTextDirection("right")
				Expect(dir).To(Equal(render.RightToLeft))
			})

			It("should handle strings with only 'left' as LeftToRight", func() {
				dir := render.ParseTextDirection("left")
				Expect(dir).To(Equal(render.LeftToRight))
			})
		})
	})
})
