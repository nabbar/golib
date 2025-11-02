/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package console_test

import (
	"github.com/fatih/color"

	. "github.com/nabbar/golib/console"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Color Management", func() {
	Describe("ColorType Constants", func() {
		It("should define ColorPrint with correct value", func() {
			Expect(ColorPrint).To(Equal(ColorType(0)))
		})

		It("should define ColorPrompt with correct value", func() {
			Expect(ColorPrompt).To(Equal(ColorType(1)))
		})

		It("should maintain order of color types", func() {
			Expect(uint8(ColorPrint)).To(BeNumerically("<", uint8(ColorPrompt)))
		})
	})

	Describe("GetColorType", func() {
		It("should convert uint8(0) to ColorPrint", func() {
			ct := GetColorType(0)
			Expect(ct).To(Equal(ColorPrint))
		})

		It("should convert uint8(1) to ColorPrompt", func() {
			ct := GetColorType(1)
			Expect(ct).To(Equal(ColorPrompt))
		})

		It("should handle arbitrary uint8 values", func() {
			ct := GetColorType(255)
			Expect(ct).To(Equal(ColorType(255)))
		})
	})

	Describe("SetColor and GetColor", func() {
		BeforeEach(func() {
			// Clean up colors before each test
			DelColor(ColorPrint)
			DelColor(ColorPrompt)
		})

		Context("when setting single color attribute", func() {
			It("should set foreground color", func() {
				SetColor(ColorPrint, int(color.FgRed))
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})

			It("should set background color", func() {
				SetColor(ColorPrint, int(color.BgBlue))
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})

			It("should set text attributes", func() {
				SetColor(ColorPrint, int(color.Bold))
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})
		})

		Context("when setting multiple color attributes", func() {
			It("should combine foreground color and bold", func() {
				SetColor(ColorPrint, int(color.FgRed), int(color.Bold))
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})

			It("should combine multiple attributes", func() {
				SetColor(ColorPrint, int(color.FgYellow), int(color.BgBlue), int(color.Bold), int(color.Underline))
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})
		})

		Context("when setting with empty attributes", func() {
			It("should handle empty attribute list", func() {
				SetColor(ColorPrint)
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})
		})

		Context("when getting color that doesn't exist", func() {
			It("should return empty color object", func() {
				DelColor(ColorPrint)
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})
		})

		Context("when using different color types", func() {
			It("should maintain separate colors for different types", func() {
				SetColor(ColorPrint, int(color.FgRed))
				SetColor(ColorPrompt, int(color.FgGreen))

				cPrint := GetColor(ColorPrint)
				cPrompt := GetColor(ColorPrompt)

				Expect(cPrint).ToNot(Equal(cPrompt))
			})
		})
	})

	Describe("ColorType.SetColor", func() {
		BeforeEach(func() {
			DelColor(ColorPrint)
			DelColor(ColorPrompt)
		})

		Context("when setting with valid color object", func() {
			It("should set color from color.Color pointer", func() {
				c := color.New(color.FgGreen)
				ColorPrint.SetColor(c)
				retrieved := GetColor(ColorPrint)
				Expect(retrieved).To(Equal(c))
			})

			It("should set color with multiple attributes", func() {
				c := color.New(color.FgCyan, color.Bold)
				ColorPrompt.SetColor(c)
				retrieved := GetColor(ColorPrompt)
				Expect(retrieved).To(Equal(c))
			})
		})

		Context("when setting with nil color", func() {
			It("should store empty color object", func() {
				ColorPrint.SetColor(nil)
				c := GetColor(ColorPrint)
				Expect(c).ToNot(BeNil())
			})
		})
	})

	Describe("DelColor", func() {
		It("should remove color from storage", func() {
			SetColor(ColorPrint, int(color.FgRed))
			DelColor(ColorPrint)
			// After deletion, should return empty color
			c := GetColor(ColorPrint)
			Expect(c).ToNot(BeNil())
		})

		It("should not panic when deleting non-existent color", func() {
			Expect(func() {
				DelColor(ColorType(99))
			}).ToNot(Panic())
		})

		It("should allow setting color again after deletion", func() {
			SetColor(ColorPrint, int(color.FgRed))
			DelColor(ColorPrint)
			SetColor(ColorPrint, int(color.FgBlue))
			c := GetColor(ColorPrint)
			Expect(c).ToNot(BeNil())
		})
	})

	Describe("Color isolation between types", func() {
		BeforeEach(func() {
			DelColor(ColorPrint)
			DelColor(ColorPrompt)
		})

		It("should not affect other color types when setting one", func() {
			SetColor(ColorPrint, int(color.FgRed))
			DelColor(ColorPrompt)

			cPrint := GetColor(ColorPrint)
			cPrompt := GetColor(ColorPrompt)

			Expect(cPrint).ToNot(BeNil())
			Expect(cPrompt).ToNot(BeNil())
		})

		It("should maintain independence when updating colors", func() {
			SetColor(ColorPrint, int(color.FgRed))
			SetColor(ColorPrompt, int(color.FgGreen))

			// Update one color
			SetColor(ColorPrint, int(color.FgBlue))

			cPrint := GetColor(ColorPrint)
			cPrompt := GetColor(ColorPrompt)

			Expect(cPrint).ToNot(BeNil())
			Expect(cPrompt).ToNot(BeNil())
		})
	})
})
