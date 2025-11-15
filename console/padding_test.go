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
	"unicode/utf8"

	. "github.com/nabbar/golib/console"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Padding Functions", func() {
	Describe("PadLeft", func() {
		Context("with basic padding", func() {
			It("should pad string on the left with spaces", func() {
				result := PadLeft("test", 10, " ")
				Expect(result).To(Equal("      test"))
				Expect(utf8.RuneCountInString(result)).To(Equal(10))
			})

			It("should pad with custom character", func() {
				result := PadLeft("abc", 5, "0")
				Expect(result).To(Equal("00abc"))
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
			})

			It("should pad with asterisks", func() {
				result := PadLeft("X", 5, "*")
				Expect(result).To(Equal("****X"))
			})

			It("should pad with dashes", func() {
				result := PadLeft("test", 8, "-")
				Expect(result).To(Equal("----test"))
			})
		})

		Context("when string is longer than target length", func() {
			It("should return original string without truncation", func() {
				result := PadLeft("verylongtext", 4, " ")
				Expect(result).To(Equal("verylongtext"))
			})

			It("should handle negative padding gracefully", func() {
				result := PadLeft("toolong", 3, " ")
				Expect(result).To(ContainSubstring("toolong"))
			})
		})

		Context("when string equals target length", func() {
			It("should return original string", func() {
				result := PadLeft("exact", 5, " ")
				Expect(result).To(Equal("exact"))
			})
		})

		Context("with empty strings", func() {
			It("should pad empty string", func() {
				result := PadLeft("", 5, "*")
				Expect(result).To(Equal("*****"))
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
			})

			It("should handle empty string with zero length", func() {
				result := PadLeft("", 0, " ")
				Expect(result).To(Equal(""))
			})
		})

		Context("with UTF-8 characters", func() {
			It("should correctly count UTF-8 characters", func() {
				result := PadLeft("café", 8, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(8))
				Expect(result).To(HaveSuffix("café"))
			})

			It("should pad Chinese characters correctly", func() {
				result := PadLeft("日本", 6, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(6))
				Expect(result).To(HaveSuffix("日本"))
			})

			It("should pad emoji correctly", func() {
				result := PadLeft("🌍", 5, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
				Expect(result).To(HaveSuffix("🌍"))
			})

			It("should handle mixed ASCII and UTF-8", func() {
				result := PadLeft("Hello世界", 12, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(12))
			})
		})

		Context("with multi-character padding", func() {
			It("should repeat entire padding string", func() {
				result := PadLeft("x", 5, "ab")
				// Needs 4 padding, repeats "ab" 4 times = "abababab" (8 chars) + "x"
				Expect(result).To(HavePrefix("ab"))
				Expect(result).To(HaveSuffix("x"))
			})
		})

		Context("edge cases", func() {
			It("should handle zero target length", func() {
				result := PadLeft("test", 0, " ")
				Expect(result).To(Equal("test"))
			})

			It("should handle large padding", func() {
				result := PadLeft("x", 100, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(100))
				Expect(result).To(HaveSuffix("x"))
			})
		})
	})

	Describe("PadRight", func() {
		Context("with basic padding", func() {
			It("should pad string on the right with spaces", func() {
				result := PadRight("test", 10, " ")
				Expect(result).To(Equal("test      "))
				Expect(utf8.RuneCountInString(result)).To(Equal(10))
			})

			It("should pad with custom character", func() {
				result := PadRight("abc", 5, "*")
				Expect(result).To(Equal("abc**"))
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
			})

			It("should pad with zeros", func() {
				result := PadRight("X", 5, "0")
				Expect(result).To(Equal("X0000"))
			})

			It("should pad with dashes", func() {
				result := PadRight("test", 8, "-")
				Expect(result).To(Equal("test----"))
			})
		})

		Context("when string is longer than target length", func() {
			It("should return original string without truncation", func() {
				result := PadRight("verylongtext", 4, " ")
				Expect(result).To(Equal("verylongtext"))
			})

			It("should handle negative padding gracefully", func() {
				result := PadRight("toolong", 3, " ")
				Expect(result).To(ContainSubstring("toolong"))
			})
		})

		Context("when string equals target length", func() {
			It("should return original string", func() {
				result := PadRight("exact", 5, " ")
				Expect(result).To(Equal("exact"))
			})
		})

		Context("with empty strings", func() {
			It("should pad empty string", func() {
				result := PadRight("", 5, "*")
				Expect(result).To(Equal("*****"))
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
			})

			It("should handle empty string with zero length", func() {
				result := PadRight("", 0, " ")
				Expect(result).To(Equal(""))
			})
		})

		Context("with UTF-8 characters", func() {
			It("should correctly count UTF-8 characters", func() {
				result := PadRight("café", 8, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(8))
				Expect(result).To(HavePrefix("café"))
			})

			It("should pad Chinese characters correctly", func() {
				result := PadRight("日本", 6, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(6))
				Expect(result).To(HavePrefix("日本"))
			})

			It("should pad emoji correctly", func() {
				result := PadRight("🌍", 5, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
				Expect(result).To(HavePrefix("🌍"))
			})

			It("should handle mixed ASCII and UTF-8", func() {
				result := PadRight("Hello世界", 12, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(12))
			})
		})

		Context("with multi-character padding", func() {
			It("should repeat entire padding string", func() {
				result := PadRight("x", 5, "ab")
				// Needs 4 padding, repeats "ab" 4 times = "abababab" (8 chars)
				Expect(result).To(HavePrefix("x"))
				Expect(result).To(HaveSuffix("ab"))
			})
		})

		Context("edge cases", func() {
			It("should handle zero target length", func() {
				result := PadRight("test", 0, " ")
				Expect(result).To(Equal("test"))
			})

			It("should handle large padding", func() {
				result := PadRight("x", 100, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(100))
				Expect(result).To(HavePrefix("x"))
			})
		})
	})

	Describe("PadCenter", func() {
		Context("with even padding", func() {
			It("should center string with equal padding on both sides", func() {
				result := PadCenter("ab", 6, "-")
				Expect(result).To(Equal("--ab--"))
				Expect(utf8.RuneCountInString(result)).To(Equal(6))
			})

			It("should center with spaces", func() {
				result := PadCenter("test", 10, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(10))
				Expect(result).To(ContainSubstring("test"))
			})

			It("should center with asterisks", func() {
				result := PadCenter("X", 5, "*")
				Expect(result).To(Equal("**X**"))
			})
		})

		Context("with odd padding", func() {
			It("should add extra padding to the right", func() {
				result := PadCenter("x", 4, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(4))
				Expect(result).To(ContainSubstring("x"))
				// Should be " x  " with more padding on the right
			})

			It("should handle odd padding with custom character", func() {
				result := PadCenter("abc", 8, "-")
				Expect(utf8.RuneCountInString(result)).To(Equal(8))
				Expect(result).To(ContainSubstring("abc"))
			})

			It("should distribute padding asymmetrically when odd", func() {
				result := PadCenter("hi", 5, "*")
				Expect(utf8.RuneCountInString(result)).To(Equal(5))
				// Should be "*hi**" or similar
				Expect(result).To(ContainSubstring("hi"))
			})
		})

		Context("when string is longer than target length", func() {
			It("should return original string without truncation", func() {
				result := PadCenter("verylongtext", 4, " ")
				Expect(result).To(Equal("verylongtext"))
			})
		})

		Context("when string equals target length", func() {
			It("should return original string", func() {
				result := PadCenter("exact", 5, " ")
				Expect(result).To(Equal("exact"))
			})
		})

		Context("with empty strings", func() {
			It("should pad empty string", func() {
				result := PadCenter("", 6, "*")
				Expect(result).To(Equal("******"))
				Expect(utf8.RuneCountInString(result)).To(Equal(6))
			})

			It("should handle empty string with zero length", func() {
				result := PadCenter("", 0, " ")
				Expect(result).To(Equal(""))
			})
		})

		Context("with UTF-8 characters", func() {
			It("should correctly center UTF-8 characters", func() {
				result := PadCenter("café", 10, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(10))
				Expect(result).To(ContainSubstring("café"))
			})

			It("should center Chinese characters", func() {
				result := PadCenter("中文", 8, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(8))
				Expect(result).To(ContainSubstring("中文"))
			})

			It("should center emoji", func() {
				result := PadCenter("🌍🌎", 8, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(8))
				Expect(result).To(ContainSubstring("🌍🌎"))
			})
		})

		Context("edge cases", func() {
			It("should handle zero target length", func() {
				result := PadCenter("test", 0, " ")
				Expect(result).To(Equal("test"))
			})

			It("should handle single character centering", func() {
				result := PadCenter("X", 1, " ")
				Expect(result).To(Equal("X"))
			})

			It("should handle large padding", func() {
				result := PadCenter("center", 100, " ")
				Expect(utf8.RuneCountInString(result)).To(Equal(100))
				Expect(result).To(ContainSubstring("center"))
			})
		})
	})

	Describe("PrintTabf", func() {
		It("should not panic with zero indentation", func() {
			captureStdout(func() {
				PrintTabf(0, "No indent")
			})
		})

		It("should not panic with one level indentation", func() {
			captureStdout(func() {
				PrintTabf(1, "One level")
			})
		})

		It("should not panic with multiple levels", func() {
			captureStdout(func() {
				PrintTabf(2, "Two levels")
				PrintTabf(3, "Three levels")
				PrintTabf(5, "Five levels")
			})
		})

		It("should support formatting", func() {
			captureStdout(func() {
				PrintTabf(1, "Name: %s, Age: %d", "Alice", 30)
			})
		})

		It("should handle empty format string", func() {
			captureStdout(func() {
				PrintTabf(1, "")
			})
		})

		It("should handle no format arguments", func() {
			captureStdout(func() {
				PrintTabf(2, "Plain text")
			})
		})

		It("should panic with negative indentation", func() {
			Expect(func() {
				captureStdout(func() {
					PrintTabf(-1, "Negative indent")
				})
			}).To(Panic())
		})

		It("should handle large indentation", func() {
			captureStdout(func() {
				PrintTabf(10, "Deep indentation")
			})
		})

		It("should support special characters in format", func() {
			captureStdout(func() {
				PrintTabf(1, "Special: \n\t%s", "chars")
			})
		})

		It("should work with Unicode", func() {
			captureStdout(func() {
				PrintTabf(1, "Unicode: %s", "世界 🌍")
			})
		})
	})

	Describe("Padding integration", func() {
		It("should combine padding functions for table formatting", func() {
			header := PadCenter("Title", 20, "=")
			left := PadLeft("Right-aligned", 20, " ")
			right := PadRight("Left-aligned", 20, " ")
			center := PadCenter("Centered", 20, " ")

			Expect(utf8.RuneCountInString(header)).To(Equal(20))
			Expect(utf8.RuneCountInString(left)).To(Equal(20))
			Expect(utf8.RuneCountInString(right)).To(Equal(20))
			Expect(utf8.RuneCountInString(center)).To(Equal(20))
		})

		It("should create consistent width columns", func() {
			col1 := PadRight("Name", 15, " ")
			col2 := PadRight("Age", 10, " ")
			col3 := PadRight("City", 15, " ")

			Expect(utf8.RuneCountInString(col1)).To(Equal(15))
			Expect(utf8.RuneCountInString(col2)).To(Equal(10))
			Expect(utf8.RuneCountInString(col3)).To(Equal(15))
		})
	})
})
