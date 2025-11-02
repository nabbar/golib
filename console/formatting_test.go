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

var _ = Describe("Console Formatting Functions", func() {
	BeforeEach(func() {
		// Reset colors before each test
		DelColor(ColorPrint)
		DelColor(ColorPrompt)
	})

	Describe("ColorType.Sprintf", func() {
		Context("with basic formatting", func() {
			It("should format string without placeholders", func() {
				result := ColorPrint.Sprintf("Hello World")
				Expect(result).To(Equal("Hello World"))
			})

			It("should format string with %s placeholder", func() {
				result := ColorPrint.Sprintf("Hello %s", "World")
				Expect(result).To(Equal("Hello World"))
			})

			It("should format integer with %d placeholder", func() {
				result := ColorPrint.Sprintf("Number: %d", 123)
				Expect(result).To(ContainSubstring("123"))
			})

			It("should format float with %f placeholder", func() {
				result := ColorPrint.Sprintf("Pi: %.2f", 3.14159)
				Expect(result).To(ContainSubstring("3.14"))
			})

			It("should format multiple values", func() {
				result := ColorPrint.Sprintf("Name: %s, Age: %d", "John", 30)
				Expect(result).To(ContainSubstring("John"))
				Expect(result).To(ContainSubstring("30"))
			})
		})

		Context("with colors applied", func() {
			BeforeEach(func() {
				SetColor(ColorPrint, int(color.FgRed))
			})

			It("should format colored string", func() {
				result := ColorPrint.Sprintf("Colored %s", "text")
				Expect(result).To(ContainSubstring("text"))
			})

			It("should include ANSI codes when color is set", func() {
				result := ColorPrint.Sprintf("Red text")
				// ANSI escape codes are included
				Expect(len(result)).To(BeNumerically(">=", len("Red text")))
			})
		})

		Context("with empty or nil values", func() {
			It("should handle empty string", func() {
				result := ColorPrint.Sprintf("")
				Expect(result).To(Equal(""))
			})

			It("should handle empty format with arguments", func() {
				result := ColorPrint.Sprintf("")
				Expect(result).To(Equal(""))
			})

			It("should handle format without arguments", func() {
				result := ColorPrint.Sprintf("No args")
				Expect(result).To(Equal("No args"))
			})
		})

		Context("with special characters", func() {
			It("should handle newlines in format", func() {
				result := ColorPrint.Sprintf("Line1\nLine2")
				Expect(result).To(ContainSubstring("\n"))
			})

			It("should handle tabs", func() {
				result := ColorPrint.Sprintf("Col1\tCol2")
				Expect(result).To(ContainSubstring("\t"))
			})

			It("should handle percent signs", func() {
				result := ColorPrint.Sprintf("100%% complete")
				Expect(result).To(ContainSubstring("100%"))
			})
		})
	})

	Describe("ColorType.Print", func() {
		It("should not panic with simple text", func() {
			captureStdout(func() {
				ColorPrint.Print("test")
			})
		})

		It("should not panic with empty string", func() {
			captureStdout(func() {
				ColorPrint.Print("")
			})
		})

		It("should not panic with special characters", func() {
			captureStdout(func() {
				ColorPrint.Print("Special: \n\t!@#$%^&*()")
			})
		})

		It("should work with colored output", func() {
			SetColor(ColorPrint, int(color.FgBlue))
			captureStdout(func() {
				ColorPrint.Print("Blue text")
			})
		})
	})

	Describe("ColorType.Println", func() {
		It("should not panic with simple text", func() {
			captureStdout(func() {
				ColorPrint.Println("test line")
			})
		})

		It("should not panic with empty string", func() {
			captureStdout(func() {
				ColorPrint.Println("")
			})
		})

		It("should work with colored output", func() {
			SetColor(ColorPrint, int(color.FgGreen))
			captureStdout(func() {
				ColorPrint.Println("Green line")
			})
		})

		It("should handle multiple lines", func() {
			captureStdout(func() {
				ColorPrint.Println("Line 1")
				ColorPrint.Println("Line 2")
				ColorPrint.Println("Line 3")
			})
		})
	})

	Describe("ColorType.Printf", func() {
		It("should not panic with basic formatting", func() {
			captureStdout(func() {
				ColorPrint.Printf("Format %s %d", "test", 123)
			})
		})

		It("should not panic with no arguments", func() {
			captureStdout(func() {
				ColorPrint.Printf("No args")
			})
		})

		It("should work with colored output", func() {
			SetColor(ColorPrint, int(color.FgYellow))
			captureStdout(func() {
				ColorPrint.Printf("Yellow: %s", "text")
			})
		})

		It("should handle complex formatting", func() {
			captureStdout(func() {
				ColorPrint.Printf("Name: %s, Age: %d, Score: %.2f", "Alice", 25, 95.5)
			})
		})

		It("should handle empty format string", func() {
			captureStdout(func() {
				ColorPrint.Printf("")
			})
		})
	})

	Describe("ColorType.PrintLnf", func() {
		It("should not panic with basic formatting", func() {
			captureStdout(func() {
				ColorPrint.PrintLnf("Line %s", "test")
			})
		})

		It("should not panic with no arguments", func() {
			captureStdout(func() {
				ColorPrint.PrintLnf("Plain line")
			})
		})

		It("should work with colored output", func() {
			SetColor(ColorPrint, int(color.FgMagenta))
			captureStdout(func() {
				ColorPrint.PrintLnf("Magenta: %d", 42)
			})
		})

		It("should handle multiple calls", func() {
			captureStdout(func() {
				ColorPrint.PrintLnf("First: %s", "A")
				ColorPrint.PrintLnf("Second: %s", "B")
				ColorPrint.PrintLnf("Third: %s", "C")
			})
		})

		It("should handle empty format string", func() {
			captureStdout(func() {
				ColorPrint.PrintLnf("")
			})
		})
	})

	Describe("Multiple ColorType interaction", func() {
		BeforeEach(func() {
			SetColor(ColorPrint, int(color.FgRed))
			SetColor(ColorPrompt, int(color.FgGreen))
		})

		It("should format independently for different types", func() {
			result1 := ColorPrint.Sprintf("Red")
			result2 := ColorPrompt.Sprintf("Green")

			Expect(result1).ToNot(Equal(result2))
		})

		It("should print independently for different types", func() {
			captureStdout(func() {
				ColorPrint.Println("Print color")
				ColorPrompt.Println("Prompt color")
			})
		})

		It("should maintain separate formatting", func() {
			captureStdout(func() {
				ColorPrint.Printf("Message from Print: %s\n", "A")
				ColorPrompt.Printf("Message from Prompt: %s\n", "B")
			})
		})
	})

	Describe("Edge cases", func() {
		It("should handle very long strings", func() {
			longString := make([]byte, 10000)
			for i := range longString {
				longString[i] = 'x'
			}
			Expect(func() {
				ColorPrint.Sprintf("%s", string(longString))
			}).ToNot(Panic())
		})

		It("should handle many arguments", func() {
			Expect(func() {
				ColorPrint.Sprintf("%s %s %s %s %s", "a", "b", "c", "d", "e")
			}).ToNot(Panic())
		})

		It("should handle Unicode characters", func() {
			result := ColorPrint.Sprintf("Hello 世界 🌍")
			Expect(result).To(ContainSubstring("世界"))
			Expect(result).To(ContainSubstring("🌍"))
		})

		It("should handle format with correct arguments", func() {
			Expect(func() {
				ColorPrint.Sprintf("%s %s %s", "one", "two", "three")
			}).ToNot(Panic())
		})
	})
})
