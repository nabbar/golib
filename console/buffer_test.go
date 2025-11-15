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
	"bytes"
	"errors"
	"io"

	"github.com/fatih/color"

	. "github.com/nabbar/golib/console"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock writer that always fails
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("mock write error")
}

var _ = Describe("Buffer Operations", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		DelColor(ColorPrint)
		DelColor(ColorPrompt)
	})

	Describe("ColorType.BuffPrintf", func() {
		Context("with valid buffer and no color", func() {
			It("should write simple text to buffer", func() {
				n, err := ColorPrint.BuffPrintf(buf, "test")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("test"))
			})

			It("should write formatted string to buffer", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Hello %s", "World")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("Hello World"))
			})

			It("should write integer formatting", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Number: %d", 42)
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("Number: 42"))
			})

			It("should write float formatting", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Pi: %.2f", 3.14159)
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("Pi: 3.14"))
			})

			It("should write multiple values", func() {
				n, err := ColorPrint.BuffPrintf(buf, "%s: %d, %s: %.1f", "Age", 25, "Score", 95.5)
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(ContainSubstring("Age: 25"))
				Expect(buf.String()).To(ContainSubstring("Score: 95.5"))
			})
		})

		Context("with valid buffer and colored output", func() {
			BeforeEach(func() {
				SetColor(ColorPrint, int(color.FgRed))
			})

			It("should write colored text to buffer", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Red text")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(ContainSubstring("Red text"))
			})

			It("should write formatted colored text", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Red: %s %d", "test", 123)
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(ContainSubstring("test"))
				Expect(buf.String()).To(ContainSubstring("123"))
			})

			It("should write colored text successfully", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Colored")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				// Output contains the text (may or may not have ANSI codes depending on env)
				Expect(buf.String()).To(ContainSubstring("Colored"))
			})
		})

		Context("with nil buffer", func() {
			It("should return error when buffer is nil", func() {
				n, err := ColorPrint.BuffPrintf(nil, "test")
				Expect(err).ToNot(BeNil())
				Expect(n).To(Equal(0))
				Expect(err.Error()).To(ContainSubstring("buffer is not defined"))
			})

			It("should return ErrorColorBufUndefined", func() {
				_, err := ColorPrint.BuffPrintf(nil, "test")
				Expect(err).To(MatchError(ContainSubstring("buffer is not defined")))
			})
		})

		Context("with failing writer", func() {
			It("should propagate write errors", func() {
				fw := &failingWriter{}
				n, err := ColorPrint.BuffPrintf(fw, "test")
				Expect(err).ToNot(BeNil())
				Expect(n).To(Equal(0))
			})
		})

		Context("with empty content", func() {
			It("should handle empty format string", func() {
				n, err := ColorPrint.BuffPrintf(buf, "")
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
				Expect(buf.String()).To(Equal(""))
			})

			It("should handle format with empty result", func() {
				n, err := ColorPrint.BuffPrintf(buf, "%s", "")
				Expect(err).To(BeNil())
				Expect(n).To(Equal(0))
				Expect(buf.String()).To(Equal(""))
			})
		})

		Context("with special characters", func() {
			It("should handle newlines", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Line1\nLine2\n")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("Line1\nLine2\n"))
			})

			It("should handle tabs", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Col1\tCol2\tCol3")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("Col1\tCol2\tCol3"))
			})

			It("should handle Unicode characters", func() {
				n, err := ColorPrint.BuffPrintf(buf, "Hello 世界 🌍")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(ContainSubstring("世界"))
				Expect(buf.String()).To(ContainSubstring("🌍"))
			})
		})

		Context("with multiple writes", func() {
			It("should append to existing buffer content", func() {
				n1, err1 := ColorPrint.BuffPrintf(buf, "First ")
				n2, err2 := ColorPrint.BuffPrintf(buf, "Second ")
				n3, err3 := ColorPrint.BuffPrintf(buf, "Third")

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
				Expect(err3).To(BeNil())
				Expect(n1 + n2 + n3).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("First Second Third"))
			})

			It("should handle multiple formatted writes", func() {
				ColorPrint.BuffPrintf(buf, "Name: %s\n", "Alice")
				ColorPrint.BuffPrintf(buf, "Age: %d\n", 30)
				ColorPrint.BuffPrintf(buf, "City: %s\n", "NYC")

				content := buf.String()
				Expect(content).To(ContainSubstring("Name: Alice"))
				Expect(content).To(ContainSubstring("Age: 30"))
				Expect(content).To(ContainSubstring("City: NYC"))
			})
		})

		Context("with different ColorTypes", func() {
			var buf1, buf2 *bytes.Buffer

			BeforeEach(func() {
				buf1 = &bytes.Buffer{}
				buf2 = &bytes.Buffer{}
				SetColor(ColorPrint, int(color.FgRed))
				SetColor(ColorPrompt, int(color.FgGreen))
			})

			It("should write with different colors to different buffers", func() {
				n1, err1 := ColorPrint.BuffPrintf(buf1, "Red text")
				n2, err2 := ColorPrompt.BuffPrintf(buf2, "Green text")

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
				Expect(n1).To(BeNumerically(">", 0))
				Expect(n2).To(BeNumerically(">", 0))

				Expect(buf1.String()).To(ContainSubstring("Red text"))
				Expect(buf2.String()).To(ContainSubstring("Green text"))
				// Different color codes mean different output
				Expect(buf1.String()).ToNot(Equal(buf2.String()))
			})

			It("should write to same buffer with different colors", func() {
				ColorPrint.BuffPrintf(buf, "Red ")
				ColorPrompt.BuffPrintf(buf, "Green")

				content := buf.String()
				Expect(content).To(ContainSubstring("Red"))
				Expect(content).To(ContainSubstring("Green"))
			})
		})

		Context("with edge cases", func() {
			It("should handle very long strings", func() {
				longText := make([]byte, 10000)
				for i := range longText {
					longText[i] = 'x'
				}
				n, err := ColorPrint.BuffPrintf(buf, "%s", string(longText))
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
			})

			It("should handle many format arguments", func() {
				n, err := ColorPrint.BuffPrintf(buf, "%s %s %s %s %s", "a", "b", "c", "d", "e")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("a b c d e"))
			})

			It("should handle percent escaping", func() {
				n, err := ColorPrint.BuffPrintf(buf, "100%% complete")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(buf.String()).To(Equal("100% complete"))
			})
		})

		Context("with different io.Writer implementations", func() {
			It("should work with any io.Writer", func() {
				var customBuf bytes.Buffer
				var writer io.Writer = &customBuf

				n, err := ColorPrint.BuffPrintf(writer, "Test")
				Expect(err).To(BeNil())
				Expect(n).To(BeNumerically(">", 0))
				Expect(customBuf.String()).To(Equal("Test"))
			})
		})
	})

	Describe("Buffer state preservation", func() {
		It("should not modify buffer on nil buffer error", func() {
			originalContent := "existing content"
			buf.WriteString(originalContent)

			// Try to write to different buffer (nil)
			ColorPrint.BuffPrintf(nil, "new content")

			// Original buffer should be unchanged
			Expect(buf.String()).To(Equal(originalContent))
		})

		It("should preserve buffer content on successful writes", func() {
			buf.WriteString("Initial: ")
			ColorPrint.BuffPrintf(buf, "Added")

			Expect(buf.String()).To(Equal("Initial: Added"))
		})
	})
})
