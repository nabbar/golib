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
	"github.com/go-hermes/hermes/v2"
	"github.com/nabbar/golib/mail/render"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mailer Interface", func() {
	var mailer render.Mailer

	BeforeEach(func() {
		mailer = render.New()
	})

	Describe("Creation", func() {
		Context("when creating a new mailer", func() {
			It("should create with default values", func() {
				Expect(mailer).ToNot(BeNil())
				Expect(mailer.GetTheme()).To(Equal(render.ThemeDefault))
				Expect(mailer.GetTextDirection()).To(Equal(render.LeftToRight))
				Expect(mailer.GetName()).To(BeEmpty())
				Expect(mailer.GetLink()).To(BeEmpty())
				Expect(mailer.GetLogo()).To(BeEmpty())
				Expect(mailer.GetCopyright()).To(BeEmpty())
				Expect(mailer.GetTroubleText()).To(BeEmpty())
				Expect(mailer.GetBody()).ToNot(BeNil())
			})
		})
	})

	Describe("Theme Management", func() {
		Context("when setting and getting theme", func() {
			It("should set and get default theme", func() {
				mailer.SetTheme(render.ThemeDefault)
				Expect(mailer.GetTheme()).To(Equal(render.ThemeDefault))
			})

			It("should set and get flat theme", func() {
				mailer.SetTheme(render.ThemeFlat)
				Expect(mailer.GetTheme()).To(Equal(render.ThemeFlat))
			})
		})
	})

	Describe("Text Direction Management", func() {
		Context("when setting and getting text direction", func() {
			It("should set and get left to right", func() {
				mailer.SetTextDirection(render.LeftToRight)
				Expect(mailer.GetTextDirection()).To(Equal(render.LeftToRight))
			})

			It("should set and get right to left", func() {
				mailer.SetTextDirection(render.RightToLeft)
				Expect(mailer.GetTextDirection()).To(Equal(render.RightToLeft))
			})
		})
	})

	Describe("Product Information", func() {
		Context("when setting product details", func() {
			It("should set and get name", func() {
				mailer.SetName("Test Company")
				Expect(mailer.GetName()).To(Equal("Test Company"))
			})

			It("should set and get link", func() {
				mailer.SetLink("https://example.com")
				Expect(mailer.GetLink()).To(Equal("https://example.com"))
			})

			It("should set and get logo", func() {
				mailer.SetLogo("https://example.com/logo.png")
				Expect(mailer.GetLogo()).To(Equal("https://example.com/logo.png"))
			})

			It("should set and get copyright", func() {
				mailer.SetCopyright("© 2024 Test Company")
				Expect(mailer.GetCopyright()).To(Equal("© 2024 Test Company"))
			})

			It("should set and get trouble text", func() {
				mailer.SetTroubleText("Having trouble? Contact support")
				Expect(mailer.GetTroubleText()).To(Equal("Having trouble? Contact support"))
			})
		})

		Context("when setting empty values", func() {
			It("should handle empty strings", func() {
				mailer.SetName("")
				mailer.SetLink("")
				mailer.SetLogo("")
				mailer.SetCopyright("")
				mailer.SetTroubleText("")

				Expect(mailer.GetName()).To(BeEmpty())
				Expect(mailer.GetLink()).To(BeEmpty())
				Expect(mailer.GetLogo()).To(BeEmpty())
				Expect(mailer.GetCopyright()).To(BeEmpty())
				Expect(mailer.GetTroubleText()).To(BeEmpty())
			})
		})
	})

	Describe("Body Management", func() {
		Context("when setting and getting body", func() {
			It("should set and get body", func() {
				body := &hermes.Body{
					Name:      "John Doe",
					Intros:    []string{"Welcome!"},
					Greeting:  "Hello",
					Signature: "Best regards",
					Title:     "Test Email",
				}

				mailer.SetBody(body)
				retrievedBody := mailer.GetBody()

				Expect(retrievedBody).ToNot(BeNil())
				Expect(retrievedBody.Name).To(Equal("John Doe"))
				Expect(retrievedBody.Intros).To(HaveLen(1))
				Expect(retrievedBody.Intros[0]).To(Equal("Welcome!"))
				Expect(retrievedBody.Greeting).To(Equal("Hello"))
				Expect(retrievedBody.Signature).To(Equal("Best regards"))
				Expect(retrievedBody.Title).To(Equal("Test Email"))
			})
		})
	})

	Describe("CSS Inline", func() {
		Context("when setting CSS inline option", func() {
			It("should set disable CSS inline to true", func() {
				mailer.SetCSSInline(true)
				// No getter for this, but it should not panic
			})

			It("should set disable CSS inline to false", func() {
				mailer.SetCSSInline(false)
				// No getter for this, but it should not panic
			})
		})
	})

	Describe("Clone", func() {
		Context("when cloning a mailer", func() {
			BeforeEach(func() {
				mailer.SetName("Original Name")
				mailer.SetLink("https://original.com")
				mailer.SetLogo("https://original.com/logo.png")
				mailer.SetCopyright("© Original")
				mailer.SetTroubleText("Original trouble")
				mailer.SetTheme(render.ThemeFlat)
				mailer.SetTextDirection(render.RightToLeft)

				body := &hermes.Body{
					Name:   "Original Body",
					Intros: []string{"Intro 1", "Intro 2"},
					Outros: []string{"Outro 1"},
					Dictionary: []hermes.Entry{
						{Key: "Key1", Value: "Value1"},
					},
					Actions: []hermes.Action{
						{
							Instructions: "Click the button",
							Button: hermes.Button{
								Text: "Click me",
								Link: "https://example.com",
							},
						},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{{Key: "Col1", Value: "Val1"}},
						},
						Columns: hermes.Columns{
							CustomWidth: map[string]string{"col1": "100px"},
						},
					}},
				}
				mailer.SetBody(body)
			})

			It("should create an independent copy", func() {
				clone := mailer.Clone()

				Expect(clone).ToNot(BeNil())
				Expect(clone.GetName()).To(Equal(mailer.GetName()))
				Expect(clone.GetLink()).To(Equal(mailer.GetLink()))
				Expect(clone.GetLogo()).To(Equal(mailer.GetLogo()))
				Expect(clone.GetCopyright()).To(Equal(mailer.GetCopyright()))
				Expect(clone.GetTroubleText()).To(Equal(mailer.GetTroubleText()))
				Expect(clone.GetTheme()).To(Equal(mailer.GetTheme()))
				Expect(clone.GetTextDirection()).To(Equal(mailer.GetTextDirection()))
			})

			It("should have independent body slices", func() {
				clone := mailer.Clone()

				// Modify clone's intros
				cloneBody := clone.GetBody()
				cloneBody.Intros = append(cloneBody.Intros, "New Intro")

				// Original should not be affected
				originalBody := mailer.GetBody()
				Expect(originalBody.Intros).To(HaveLen(2))
				Expect(cloneBody.Intros).To(HaveLen(3))
			})

			It("should have independent dictionary entries", func() {
				clone := mailer.Clone()

				// Modify clone's dictionary
				cloneBody := clone.GetBody()
				cloneBody.Dictionary[0].Value = "Modified Value"

				// Original should not be affected
				originalBody := mailer.GetBody()
				Expect(originalBody.Dictionary[0].Value).To(Equal("Value1"))
				Expect(cloneBody.Dictionary[0].Value).To(Equal("Modified Value"))
			})

			It("should have independent table data", func() {
				clone := mailer.Clone()

				// Modify clone's table data
				cloneBody := clone.GetBody()
				cloneBody.Tables[0].Data[0][0].Value = "Modified"

				// Original should not be affected
				originalBody := mailer.GetBody()
				Expect(originalBody.Tables[0].Data[0][0].Value).To(Equal("Val1"))
				Expect(cloneBody.Tables[0].Data[0][0].Value).To(Equal("Modified"))
			})

			It("should handle nil slices in body", func() {
				emptyMailer := render.New()
				emptyMailer.SetBody(&hermes.Body{
					Name: "Test",
				})

				clone := emptyMailer.Clone()
				Expect(clone).ToNot(BeNil())
				Expect(clone.GetBody()).ToNot(BeNil())
			})
		})
	})
})
