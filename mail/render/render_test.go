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
	"strings"

	"github.com/go-hermes/hermes/v2"
	"github.com/nabbar/golib/mail/render"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Email Rendering", func() {
	var mailer render.Mailer

	BeforeEach(func() {
		mailer = render.New()
		mailer.SetName("Test Company")
		mailer.SetLink("https://example.com")
		mailer.SetLogo("https://example.com/logo.png")
		mailer.SetCopyright("© 2024 Test Company")
		mailer.SetTroubleText("Having trouble? Contact us")
	})

	Describe("HTML Generation", func() {
		Context("with basic content", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "John Doe",
					Intros: []string{"Welcome to our service!"},
					Outros: []string{"Thank you for using our service."},
				}
				mailer.SetBody(body)
			})

			It("should generate valid HTML", func() {
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
				Expect(buf.Len()).To(BeNumerically(">", 0))

				html := buf.String()
				Expect(html).To(ContainSubstring("<!DOCTYPE html"))
				Expect(html).To(ContainSubstring("John Doe"))
				Expect(html).To(ContainSubstring("Welcome to our service!"))
			})

			It("should include product information in HTML", func() {
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("Test Company"))
				Expect(html).To(ContainSubstring("© 2024 Test Company"))
			})
		})

		Context("with complex content", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Here is your activity report:"},
					Dictionary: []hermes.Entry{
						{Key: "Transaction ID", Value: "123456"},
						{Key: "Amount", Value: "$99.99"},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{
								{Key: "Item", Value: "Product A"},
								{Key: "Price", Value: "$49.99"},
							},
							{
								{Key: "Item", Value: "Product B"},
								{Key: "Price", Value: "$49.99"},
							},
						},
						Columns: hermes.Columns{
							CustomWidth: map[string]string{
								"Item":  "80%",
								"Price": "20%",
							},
						},
					}},
					Actions: []hermes.Action{
						{
							Instructions: "To view your order, click below:",
							Button: hermes.Button{
								Text: "View Order",
								Link: "https://example.com/order/123456",
							},
						},
					},
					Outros: []string{"Need help? Reply to this email."},
				}
				mailer.SetBody(body)
			})

			It("should generate HTML with table data", func() {
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("Product A"))
				Expect(html).To(ContainSubstring("Product B"))
				Expect(html).To(ContainSubstring("$49.99"))
			})

			It("should generate HTML with action button", func() {
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("View Order"))
				Expect(html).To(ContainSubstring("https://example.com/order/123456"))
			})

			It("should generate HTML with dictionary entries", func() {
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("Transaction ID"))
				Expect(html).To(ContainSubstring("123456"))
				Expect(html).To(ContainSubstring("$99.99"))
			})
		})

		Context("with different themes", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Test email"},
				}
				mailer.SetBody(body)
			})

			It("should generate HTML with default theme", func() {
				mailer.SetTheme(render.ThemeDefault)
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				Expect(buf.Len()).To(BeNumerically(">", 0))
			})

			It("should generate HTML with flat theme", func() {
				mailer.SetTheme(render.ThemeFlat)
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				Expect(buf.Len()).To(BeNumerically(">", 0))
			})
		})

		Context("with CSS inline options", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Test"},
				}
				mailer.SetBody(body)
			})

			It("should generate HTML with CSS inline enabled", func() {
				mailer.SetCSSInline(false)
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				// With inline CSS, styles should be in style attributes
				Expect(html).To(ContainSubstring("style="))
			})

			It("should generate HTML with CSS inline disabled", func() {
				mailer.SetCSSInline(true)
				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
			})
		})
	})

	Describe("Plain Text Generation", func() {
		Context("with basic content", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "John Doe",
					Intros: []string{"Welcome to our service!"},
					Outros: []string{"Thank you."},
				}
				mailer.SetBody(body)
			})

			It("should generate valid plain text", func() {
				buf, err := mailer.GeneratePlainText()

				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
				Expect(buf.Len()).To(BeNumerically(">", 0))

				text := buf.String()
				Expect(text).To(ContainSubstring("John Doe"))
				Expect(text).To(ContainSubstring("Welcome to our service!"))
				Expect(text).ToNot(ContainSubstring("<html>"))
			})
		})

		Context("with complex content", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Report:"},
					Dictionary: []hermes.Entry{
						{Key: "ID", Value: "123"},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{{Key: "Item", Value: "A"}},
						},
					}},
					Actions: []hermes.Action{
						{
							Instructions: "Click here:",
							Button: hermes.Button{
								Text: "Action",
								Link: "https://example.com",
							},
						},
					},
				}
				mailer.SetBody(body)
			})

			It("should include all content in plain text", func() {
				buf, err := mailer.GeneratePlainText()

				Expect(err).To(BeNil())
				text := buf.String()
				Expect(text).To(ContainSubstring("User"))
				Expect(text).To(ContainSubstring("Report:"))
				Expect(text).To(ContainSubstring("https://example.com"))
			})
		})
	})

	Describe("ParseData", func() {
		Context("with template variables", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Name:      "{{username}}",
					Intros:    []string{"Hello {{username}}, your code is {{code}}"},
					Greeting:  "Hi {{username}}",
					Signature: "{{company}}",
					Title:     "Welcome {{username}}",
					Dictionary: []hermes.Entry{
						{Key: "{{key}}", Value: "{{value}}"},
					},
				}
				mailer.SetBody(body)
				mailer.SetName("{{company}}")
				mailer.SetCopyright("© {{year}} {{company}}")
			})

			It("should replace all template variables", func() {
				data := map[string]string{
					"{{username}}": "John Doe",
					"{{code}}":     "123456",
					"{{company}}":  "Test Inc",
					"{{year}}":     "2024",
					"{{key}}":      "Key1",
					"{{value}}":    "Value1",
				}

				mailer.ParseData(data)

				body := mailer.GetBody()
				Expect(body.Name).To(Equal("John Doe"))
				Expect(body.Intros[0]).To(Equal("Hello John Doe, your code is 123456"))
				Expect(body.Greeting).To(Equal("Hi John Doe"))
				Expect(body.Signature).To(Equal("Test Inc"))
				Expect(body.Title).To(Equal("Welcome John Doe"))
				Expect(body.Dictionary[0].Key).To(Equal("Key1"))
				Expect(body.Dictionary[0].Value).To(Equal("Value1"))

				Expect(mailer.GetName()).To(Equal("Test Inc"))
				Expect(mailer.GetCopyright()).To(Equal("© 2024 Test Inc"))
			})

			It("should handle multiple replacements in same string", func() {
				body := &hermes.Body{
					Intros: []string{"{{var1}} and {{var2}} and {{var1}}"},
				}
				mailer.SetBody(body)

				data := map[string]string{
					"{{var1}}": "First",
					"{{var2}}": "Second",
				}

				mailer.ParseData(data)

				body = mailer.GetBody()
				Expect(body.Intros[0]).To(Equal("First and Second and First"))
			})
		})

		Context("with table and action replacements", func() {
			BeforeEach(func() {
				body := &hermes.Body{
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{
								{Key: "{{header1}}", Value: "{{value1}}"},
								{Key: "{{header2}}", Value: "{{value2}}"},
							},
						},
					}},
					Actions: []hermes.Action{
						{
							Instructions: "{{instruction}}",
							InviteCode:   "{{code}}",
							Button: hermes.Button{
								Text:      "{{buttonText}}",
								Link:      "{{buttonLink}}",
								Color:     "{{color}}",
								TextColor: "{{textColor}}",
							},
						},
					},
				}
				mailer.SetBody(body)
			})

			It("should replace variables in table data", func() {
				data := map[string]string{
					"{{header1}}": "Name",
					"{{header2}}": "Email",
					"{{value1}}":  "John",
					"{{value2}}":  "john@example.com",
				}

				mailer.ParseData(data)

				body := mailer.GetBody()
				Expect(body.Tables[0].Data[0][0].Key).To(Equal("Name"))
				Expect(body.Tables[0].Data[0][0].Value).To(Equal("John"))
				Expect(body.Tables[0].Data[0][1].Key).To(Equal("Email"))
				Expect(body.Tables[0].Data[0][1].Value).To(Equal("john@example.com"))
			})

			It("should replace variables in actions", func() {
				data := map[string]string{
					"{{instruction}}": "Click below",
					"{{code}}":        "INVITE123",
					"{{buttonText}}":  "Accept",
					"{{buttonLink}}":  "https://example.com/accept",
					"{{color}}":       "#007bff",
					"{{textColor}}":   "#ffffff",
				}

				mailer.ParseData(data)

				body := mailer.GetBody()
				action := body.Actions[0]
				Expect(action.Instructions).To(Equal("Click below"))
				Expect(action.InviteCode).To(Equal("INVITE123"))
				Expect(action.Button.Text).To(Equal("Accept"))
				Expect(action.Button.Link).To(Equal("https://example.com/accept"))
				Expect(action.Button.Color).To(Equal("#007bff"))
				Expect(action.Button.TextColor).To(Equal("#ffffff"))
			})
		})

		Context("with markdown content", func() {
			It("should replace variables in free markdown", func() {
				body := &hermes.Body{
					FreeMarkdown: hermes.Markdown("# {{title}}\n\nHello {{name}}!"),
				}
				mailer.SetBody(body)

				data := map[string]string{
					"{{title}}": "Welcome",
					"{{name}}":  "User",
				}

				mailer.ParseData(data)

				body = mailer.GetBody()
				markdown := string(body.FreeMarkdown)
				Expect(markdown).To(ContainSubstring("# Welcome"))
				Expect(markdown).To(ContainSubstring("Hello User!"))
			})
		})

		Context("with empty or nil data", func() {
			It("should handle empty data map", func() {
				body := &hermes.Body{
					Name: "Original",
				}
				mailer.SetBody(body)

				mailer.ParseData(map[string]string{})

				// Should not modify anything
				Expect(mailer.GetBody().Name).To(Equal("Original"))
			})

			It("should handle nil slices gracefully", func() {
				body := &hermes.Body{
					Name: "Test",
				}
				mailer.SetBody(body)

				// Should not panic
				mailer.ParseData(map[string]string{"{{test}}": "value"})
				Expect(mailer.GetBody().Name).To(Equal("Test"))
			})
		})
	})

	Describe("Integration", func() {
		Context("when using complete workflow", func() {
			It("should create, configure, parse and generate email", func() {
				// Configure
				mailer.SetName("{{company}}")
				mailer.SetTheme(render.ThemeFlat)

				// Set body
				body := &hermes.Body{
					Name:   "{{user}}",
					Intros: []string{"Welcome {{user}}!"},
					Actions: []hermes.Action{
						{
							Instructions: "Verify your email:",
							Button: hermes.Button{
								Text: "Verify",
								Link: "{{link}}",
							},
						},
					},
				}
				mailer.SetBody(body)

				// Parse data
				mailer.ParseData(map[string]string{
					"{{company}}": "Test Inc",
					"{{user}}":    "John",
					"{{link}}":    "https://example.com/verify",
				})

				// Generate both formats
				htmlBuf, htmlErr := mailer.GenerateHTML()
				textBuf, textErr := mailer.GeneratePlainText()

				Expect(htmlErr).To(BeNil())
				Expect(textErr).To(BeNil())
				Expect(htmlBuf.Len()).To(BeNumerically(">", 0))
				Expect(textBuf.Len()).To(BeNumerically(">", 0))

				html := htmlBuf.String()
				text := textBuf.String()

				Expect(html).To(ContainSubstring("John"))
				Expect(html).To(ContainSubstring("Welcome John!"))
				Expect(html).To(ContainSubstring("https://example.com/verify"))

				Expect(text).To(ContainSubstring("John"))
				Expect(text).To(ContainSubstring("Welcome John!"))
			})
		})
	})

	Describe("Error Handling", func() {
		Context("with invalid configurations", func() {
			It("should handle generation with minimal setup", func() {
				minimalMailer := render.New()
				minimalMailer.SetBody(&hermes.Body{
					Name: "Test",
				})

				htmlBuf, htmlErr := minimalMailer.GenerateHTML()
				textBuf, textErr := minimalMailer.GeneratePlainText()

				// Should not error with minimal setup
				Expect(htmlErr).To(BeNil())
				Expect(textErr).To(BeNil())
				Expect(htmlBuf).ToNot(BeNil())
				Expect(textBuf).ToNot(BeNil())
			})
		})
	})

	Describe("Special Characters", func() {
		Context("when content contains special characters", func() {
			It("should handle HTML special characters", func() {
				body := &hermes.Body{
					Name:   "User <>&\"'",
					Intros: []string{"Test <script>alert('xss')</script>"},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				// Hermes should escape HTML entities
				Expect(html).ToNot(ContainSubstring("<script>"))
			})

			It("should handle unicode characters", func() {
				body := &hermes.Body{
					Name:   "用户 👤",
					Intros: []string{"Hello 世界 🌍!"},
				}
				mailer.SetBody(body)

				htmlBuf, htmlErr := mailer.GenerateHTML()
				textBuf, textErr := mailer.GeneratePlainText()

				Expect(htmlErr).To(BeNil())
				Expect(textErr).To(BeNil())
				Expect(htmlBuf.String()).To(ContainSubstring("用户"))
				Expect(textBuf.String()).To(ContainSubstring("世界"))
			})
		})
	})

	Describe("Text Direction", func() {
		Context("with different text directions", func() {
			It("should generate with left-to-right direction", func() {
				mailer.SetTextDirection(render.LeftToRight)
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Test"},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("ltr"))
			})

			It("should generate with right-to-left direction", func() {
				mailer.SetTextDirection(render.RightToLeft)
				body := &hermes.Body{
					Name:   "مستخدم",
					Intros: []string{"مرحبا"},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()

				Expect(err).To(BeNil())
				html := buf.String()
				Expect(html).To(ContainSubstring("rtl"))
			})
		})
	})

	Describe("Empty Content", func() {
		Context("with empty body sections", func() {
			It("should handle empty intros", func() {
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()
				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
			})

			It("should handle empty outros", func() {
				body := &hermes.Body{
					Name:   "User",
					Outros: []string{},
				}
				mailer.SetBody(body)

				buf, err := mailer.GeneratePlainText()
				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
			})
		})
	})

	Describe("Long Content", func() {
		Context("with very long strings", func() {
			It("should handle long intro text", func() {
				longText := strings.Repeat("Lorem ipsum dolor sit amet. ", 100)
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{longText},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()
				Expect(err).To(BeNil())
				Expect(buf.Len()).To(BeNumerically(">", len(longText)))
			})

			It("should handle many table rows", func() {
				data := make([][]hermes.Entry, 100)
				for i := range data {
					data[i] = []hermes.Entry{
						{Key: "Item", Value: "Row " + string(rune(i))},
					}
				}

				body := &hermes.Body{
					Name: "User",
					Tables: []hermes.Table{{
						Data: data,
					}},
				}
				mailer.SetBody(body)

				buf, err := mailer.GenerateHTML()
				Expect(err).To(BeNil())
				Expect(buf).ToNot(BeNil())
			})
		})
	})
})
