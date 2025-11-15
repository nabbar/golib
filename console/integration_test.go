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
	"strings"

	"github.com/fatih/color"

	. "github.com/nabbar/golib/console"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Console Integration Tests", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		DelColor(ColorPrint)
		DelColor(ColorPrompt)
	})

	Describe("Complete workflow scenarios", func() {
		Context("colored output to buffer", func() {
			It("should support full workflow with color", func() {
				SetColor(ColorPrint, int(color.FgCyan), int(color.Bold))

				n1, err1 := ColorPrint.BuffPrintf(buf, "First line\n")
				n2, err2 := ColorPrint.BuffPrintf(buf, "Second line\n")
				n3, err3 := ColorPrint.BuffPrintf(buf, "Third line")

				Expect(err1).To(BeNil())
				Expect(err2).To(BeNil())
				Expect(err3).To(BeNil())
				Expect(n1).To(BeNumerically(">", 0))
				Expect(n2).To(BeNumerically(">", 0))
				Expect(n3).To(BeNumerically(">", 0))

				content := buf.String()
				Expect(content).To(ContainSubstring("First line"))
				Expect(content).To(ContainSubstring("Second line"))
				Expect(content).To(ContainSubstring("Third line"))
			})

			It("should handle multiple colors in same buffer", func() {
				buf1 := &bytes.Buffer{}
				buf2 := &bytes.Buffer{}

				SetColor(ColorPrint, int(color.FgRed))
				SetColor(ColorPrompt, int(color.FgGreen))

				ColorPrint.BuffPrintf(buf1, "Red text\n")
				ColorPrompt.BuffPrintf(buf2, "Green text\n")

				Expect(buf1.String()).To(ContainSubstring("Red text"))
				Expect(buf2.String()).To(ContainSubstring("Green text"))
				Expect(buf1.String()).ToNot(Equal(buf2.String()))
			})
		})

		Context("formatted table output", func() {
			It("should create aligned table with padding", func() {
				header1 := PadRight("Name", 20, " ")
				header2 := PadRight("Age", 10, " ")
				header3 := PadRight("City", 15, " ")

				row1_1 := PadRight("Alice", 20, " ")
				row1_2 := PadRight("30", 10, " ")
				row1_3 := PadRight("NYC", 15, " ")

				row2_1 := PadRight("Bob", 20, " ")
				row2_2 := PadRight("25", 10, " ")
				row2_3 := PadRight("LA", 15, " ")

				ColorPrint.BuffPrintf(buf, "%s%s%s\n", header1, header2, header3)
				ColorPrint.BuffPrintf(buf, "%s\n", strings.Repeat("-", 45))
				ColorPrint.BuffPrintf(buf, "%s%s%s\n", row1_1, row1_2, row1_3)
				ColorPrint.BuffPrintf(buf, "%s%s%s\n", row2_1, row2_2, row2_3)

				content := buf.String()
				Expect(content).To(ContainSubstring("Name"))
				Expect(content).To(ContainSubstring("Alice"))
				Expect(content).To(ContainSubstring("Bob"))
			})

			It("should create centered title with decorations", func() {
				title := PadCenter("Report Title", 50, " ")
				border := strings.Repeat("=", 50)

				ColorPrint.BuffPrintf(buf, "%s\n", border)
				ColorPrint.BuffPrintf(buf, "%s\n", title)
				ColorPrint.BuffPrintf(buf, "%s\n", border)

				content := buf.String()
				Expect(content).To(ContainSubstring("Report Title"))
				Expect(content).To(ContainSubstring("="))
			})
		})

		Context("hierarchical output with PrintTabf", func() {
			It("should create nested structure output", func() {
				captureStdout(func() {
					PrintTabf(0, "Root\n")
					PrintTabf(1, "Child 1\n")
					PrintTabf(2, "Grandchild 1.1\n")
					PrintTabf(2, "Grandchild 1.2\n")
					PrintTabf(1, "Child 2\n")
					PrintTabf(2, "Grandchild 2.1\n")
				})
			})

			It("should format hierarchical data", func() {
				captureStdout(func() {
					PrintTabf(0, "Configuration:\n")
					PrintTabf(1, "Database:\n")
					PrintTabf(2, "Host: %s\n", "localhost")
					PrintTabf(2, "Port: %d\n", 5432)
					PrintTabf(1, "Cache:\n")
					PrintTabf(2, "Enabled: %t\n", true)
					PrintTabf(2, "Size: %d MB\n", 512)
				})
			})
		})

		Context("combining features", func() {
			It("should combine color, padding, and formatting", func() {
				SetColor(ColorPrint, int(color.FgCyan))
				SetColor(ColorPrompt, int(color.FgYellow))

				title := PadCenter("System Status", 40, "=")
				ColorPrint.BuffPrintf(buf, "%s\n", title)

				label1 := PadRight("CPU Usage", 15, " ")
				value1 := PadLeft("45%", 10, " ")
				ColorPrint.BuffPrintf(buf, "%s:%s\n", label1, value1)

				label2 := PadRight("Memory", 15, " ")
				value2 := PadLeft("2.5GB", 10, " ")
				ColorPrompt.BuffPrintf(buf, "%s:%s\n", label2, value2)

				content := buf.String()
				Expect(content).To(ContainSubstring("System Status"))
				Expect(content).To(ContainSubstring("CPU Usage"))
				Expect(content).To(ContainSubstring("Memory"))
			})

			It("should create progress indicator with padding", func() {
				for i := 0; i <= 100; i += 20 {
					filled := i / 5
					empty := 20 - filled
					bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
					label := PadRight("Progress", 12, " ")

					buf.Reset()
					ColorPrint.BuffPrintf(buf, "%s [%s] %3d%%", label, bar, i)

					content := buf.String()
					Expect(content).To(ContainSubstring("Progress"))
					Expect(content).ToNot(BeEmpty())
				}
			})
		})

		Context("multi-language support", func() {
			It("should handle UTF-8 text with padding and colors", func() {
				SetColor(ColorPrint, int(color.FgGreen))

				texts := map[string]string{
					"English":  "Hello World",
					"French":   "Bonjour le monde",
					"Spanish":  "Hola Mundo",
					"Chinese":  "你好世界",
					"Japanese": "こんにちは世界",
					"Arabic":   "مرحبا بالعالم",
					"Emoji":    "Hello 🌍 World 🌎",
				}

				for lang, text := range texts {
					label := PadRight(lang, 12, " ")
					ColorPrint.BuffPrintf(buf, "%s: %s\n", label, text)
				}

				content := buf.String()
				Expect(content).To(ContainSubstring("English"))
				Expect(content).To(ContainSubstring("你好世界"))
				Expect(content).To(ContainSubstring("🌍"))
			})
		})

		Context("error handling in workflows", func() {
			It("should gracefully handle nil buffer in workflow", func() {
				SetColor(ColorPrint, int(color.FgRed))

				_, err := ColorPrint.BuffPrintf(nil, "This should fail")
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError(ContainSubstring("buffer is not defined")))

				// Should still work with valid buffer
				_, err2 := ColorPrint.BuffPrintf(buf, "This should work")
				Expect(err2).To(BeNil())
			})

			It("should continue after errors", func() {
				// First call fails
				n1, err1 := ColorPrint.BuffPrintf(nil, "fail")
				Expect(err1).ToNot(BeNil())
				Expect(n1).To(Equal(0))

				// Second call succeeds
				n2, err2 := ColorPrint.BuffPrintf(buf, "success")
				Expect(err2).To(BeNil())
				Expect(n2).To(BeNumerically(">", 0))

				// Third call succeeds
				n3, err3 := ColorPrint.BuffPrintf(buf, " continued")
				Expect(err3).To(BeNil())
				Expect(n3).To(BeNumerically(">", 0))

				Expect(buf.String()).To(Equal("success continued"))
			})
		})

		Context("color management across operations", func() {
			It("should maintain color state", func() {
				SetColor(ColorPrint, int(color.FgRed))

				ColorPrint.BuffPrintf(buf, "Red ")
				ColorPrint.BuffPrintf(buf, "Still Red ")

				SetColor(ColorPrint, int(color.FgBlue))
				ColorPrint.BuffPrintf(buf, "Now Blue")

				content := buf.String()
				Expect(content).To(ContainSubstring("Red"))
				Expect(content).To(ContainSubstring("Blue"))
			})

			It("should allow dynamic color changes", func() {
				colors := []color.Attribute{
					color.FgRed,
					color.FgGreen,
					color.FgYellow,
					color.FgBlue,
					color.FgMagenta,
					color.FgCyan,
				}

				for _, c := range colors {
					buf.Reset()
					SetColor(ColorPrint, int(c))
					ColorPrint.BuffPrintf(buf, "Colored text")
					Expect(buf.Len()).To(BeNumerically(">", 0))
				}
			})

			It("should support removing and re-adding colors", func() {
				SetColor(ColorPrint, int(color.FgRed))
				ColorPrint.BuffPrintf(buf, "Red ")

				DelColor(ColorPrint)
				ColorPrint.BuffPrintf(buf, "No Color ")

				SetColor(ColorPrint, int(color.FgGreen))
				ColorPrint.BuffPrintf(buf, "Green")

				Expect(buf.String()).To(ContainSubstring("Red"))
				Expect(buf.String()).To(ContainSubstring("Green"))
			})
		})

		Context("complex formatting scenarios", func() {
			It("should format data report with mixed features", func() {
				SetColor(ColorPrint, int(color.FgCyan), int(color.Bold))

				// Title
				title := PadCenter("SALES REPORT", 60, "=")
				ColorPrint.BuffPrintf(buf, "\n%s\n\n", title)

				// Headers
				h1 := PadRight("Product", 25, " ")
				h2 := PadLeft("Units", 10, " ")
				h3 := PadLeft("Revenue", 15, " ")
				ColorPrint.BuffPrintf(buf, "%s %s %s\n", h1, h2, h3)
				ColorPrint.BuffPrintf(buf, "%s\n", strings.Repeat("-", 60))

				// Data rows
				products := []struct {
					name    string
					units   int
					revenue float64
				}{
					{"Widget A", 150, 15000.50},
					{"Widget B", 89, 8900.00},
					{"Widget C", 234, 23400.75},
				}

				for _, p := range products {
					c1 := PadRight(p.name, 25, " ")
					c2 := PadLeft(ColorPrint.Sprintf("%d", p.units), 10, " ")
					c3 := PadLeft(ColorPrint.Sprintf("$%.2f", p.revenue), 15, " ")
					ColorPrint.BuffPrintf(buf, "%s %s %s\n", c1, c2, c3)
				}

				content := buf.String()
				Expect(content).To(ContainSubstring("SALES REPORT"))
				Expect(content).To(ContainSubstring("Widget A"))
				Expect(content).To(ContainSubstring("150"))
				Expect(content).To(ContainSubstring("15000.50"))
			})

			It("should create config dump with indentation", func() {
				captureStdout(func() {
					PrintTabf(0, "Application Configuration\n")
					PrintTabf(0, "%s\n", strings.Repeat("=", 50))
					PrintTabf(0, "\n")
					PrintTabf(0, "Server:\n")
					PrintTabf(1, "Host: %s\n", "0.0.0.0")
					PrintTabf(1, "Port: %d\n", 8080)
					PrintTabf(1, "TLS:\n")
					PrintTabf(2, "Enabled: %t\n", true)
					PrintTabf(2, "CertFile: %s\n", "/etc/ssl/cert.pem")
					PrintTabf(0, "\n")
					PrintTabf(0, "Database:\n")
					PrintTabf(1, "Driver: %s\n", "postgres")
					PrintTabf(1, "MaxConnections: %d\n", 100)
					PrintTabf(1, "Pool:\n")
					PrintTabf(2, "MinIdle: %d\n", 10)
					PrintTabf(2, "MaxIdle: %d\n", 20)
				})
			})
		})
	})

	Describe("Performance and stress tests", func() {
		It("should handle many sequential writes", func() {
			SetColor(ColorPrint, int(color.FgGreen))

			for i := 0; i < 1000; i++ {
				_, err := ColorPrint.BuffPrintf(buf, "Line %d\n", i)
				Expect(err).To(BeNil())
			}

			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("should handle rapid color changes", func() {
			colors := []color.Attribute{
				color.FgRed, color.FgGreen, color.FgBlue,
				color.FgYellow, color.FgMagenta, color.FgCyan,
			}

			for i := 0; i < 100; i++ {
				SetColor(ColorPrint, int(colors[i%len(colors)]))
				ColorPrint.BuffPrintf(buf, "T")
			}

			Expect(buf.Len()).To(BeNumerically(">", 0))
		})

		It("should handle large padded strings", func() {
			longStr := strings.Repeat("x", 500)
			result := PadCenter(longStr, 1000, " ")
			Expect(len([]rune(result))).To(Equal(1000))
		})
	})

	Describe("Thread-safety scenarios", func() {
		It("should handle concurrent color types independently", func() {
			buf1 := &bytes.Buffer{}
			buf2 := &bytes.Buffer{}

			SetColor(ColorPrint, int(color.FgRed))
			SetColor(ColorPrompt, int(color.FgGreen))

			done := make(chan bool, 2)

			go func() {
				for i := 0; i < 100; i++ {
					ColorPrint.BuffPrintf(buf1, "P%d ", i)
				}
				done <- true
			}()

			go func() {
				for i := 0; i < 100; i++ {
					ColorPrompt.BuffPrintf(buf2, "R%d ", i)
				}
				done <- true
			}()

			<-done
			<-done

			Expect(buf1.Len()).To(BeNumerically(">", 0))
			Expect(buf2.Len()).To(BeNumerically(">", 0))
		})
	})
})
