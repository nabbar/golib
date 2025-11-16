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

var _ = Describe("Config", func() {
	var validConfig render.Config

	BeforeEach(func() {
		validConfig = render.Config{
			Theme:            "default",
			Direction:        "ltr",
			Name:             "Test Company",
			Link:             "https://example.com",
			Logo:             "https://example.com/logo.png",
			Copyright:        "© 2024 Test Company",
			TroubleText:      "Having trouble?",
			DisableCSSInline: false,
			Body: hermes.Body{
				Name:   "John Doe",
				Intros: []string{"Welcome to our service!"},
			},
		}
	})

	Describe("Validation", func() {
		Context("with valid configuration", func() {
			It("should validate successfully", func() {
				err := validConfig.Validate()
				Expect(err).To(BeNil())
			})
		})

		Context("with missing required fields", func() {
			It("should fail validation when theme is missing", func() {
				validConfig.Theme = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when direction is missing", func() {
				validConfig.Direction = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when name is missing", func() {
				validConfig.Name = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when link is missing", func() {
				validConfig.Link = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when logo is missing", func() {
				validConfig.Logo = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when copyright is missing", func() {
				validConfig.Copyright = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation when trouble text is missing", func() {
				validConfig.TroubleText = ""
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})
		})

		Context("with invalid URL format", func() {
			It("should fail validation for invalid link URL", func() {
				validConfig.Link = "not-a-valid-url"
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should fail validation for invalid logo URL", func() {
				validConfig.Logo = "not-a-valid-url"
				err := validConfig.Validate()
				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should accept valid HTTP URLs", func() {
				validConfig.Link = "http://example.com"
				validConfig.Logo = "http://example.com/logo.png"
				err := validConfig.Validate()
				Expect(err).To(BeNil())
			})

			It("should accept valid HTTPS URLs", func() {
				validConfig.Link = "https://example.com"
				validConfig.Logo = "https://example.com/logo.png"
				err := validConfig.Validate()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("NewMailer", func() {
		Context("when creating mailer from config", func() {
			It("should create mailer with correct values", func() {
				mailer := validConfig.NewMailer()

				Expect(mailer).ToNot(BeNil())
				Expect(mailer.GetName()).To(Equal(validConfig.Name))
				Expect(mailer.GetLink()).To(Equal(validConfig.Link))
				Expect(mailer.GetLogo()).To(Equal(validConfig.Logo))
				Expect(mailer.GetCopyright()).To(Equal(validConfig.Copyright))
				Expect(mailer.GetTroubleText()).To(Equal(validConfig.TroubleText))
				Expect(mailer.GetTheme()).To(Equal(render.ThemeDefault))
				Expect(mailer.GetTextDirection()).To(Equal(render.LeftToRight))
			})

			It("should create mailer with flat theme", func() {
				validConfig.Theme = "flat"
				mailer := validConfig.NewMailer()

				Expect(mailer.GetTheme()).To(Equal(render.ThemeFlat))
			})

			It("should create mailer with right to left direction", func() {
				validConfig.Direction = "rtl"
				mailer := validConfig.NewMailer()

				Expect(mailer.GetTextDirection()).To(Equal(render.RightToLeft))
			})

			It("should preserve body content", func() {
				mailer := validConfig.NewMailer()
				body := mailer.GetBody()

				Expect(body).ToNot(BeNil())
				Expect(body.Name).To(Equal("John Doe"))
				Expect(body.Intros).To(HaveLen(1))
				Expect(body.Intros[0]).To(Equal("Welcome to our service!"))
			})

			It("should handle CSS inline disable flag", func() {
				validConfig.DisableCSSInline = true
				mailer := validConfig.NewMailer()

				// Should not panic
				Expect(mailer).ToNot(BeNil())
			})
		})

		Context("with different configurations", func() {
			It("should handle minimal valid configuration", func() {
				minConfig := render.Config{
					Theme:       "default",
					Direction:   "ltr",
					Name:        "Company",
					Link:        "https://company.com",
					Logo:        "https://company.com/logo.png",
					Copyright:   "©",
					TroubleText: "Help",
					Body:        hermes.Body{},
				}

				err := minConfig.Validate()
				Expect(err).To(BeNil())

				mailer := minConfig.NewMailer()
				Expect(mailer).ToNot(BeNil())
			})

			It("should handle configuration with complex body", func() {
				validConfig.Body = hermes.Body{
					Name:   "User",
					Intros: []string{"Line 1", "Line 2"},
					Outros: []string{"Goodbye"},
					Actions: []hermes.Action{
						{
							Instructions: "Click below",
							Button: hermes.Button{
								Text: "Action",
								Link: "https://example.com/action",
							},
						},
					},
					Dictionary: []hermes.Entry{
						{Key: "Key", Value: "Value"},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{{Key: "Col1", Value: "Val1"}},
						},
					}},
				}

				mailer := validConfig.NewMailer()
				body := mailer.GetBody()

				Expect(body.Name).To(Equal("User"))
				Expect(body.Intros).To(HaveLen(2))
				Expect(body.Outros).To(HaveLen(1))
				Expect(body.Actions).To(HaveLen(1))
				Expect(body.Dictionary).To(HaveLen(1))
				Expect(body.Tables[0].Data).To(HaveLen(1))
			})
		})
	})

	Describe("Configuration Edge Cases", func() {
		Context("when handling special characters", func() {
			It("should handle special characters in text fields", func() {
				validConfig.Name = "Company™ & Co. <>"
				validConfig.Copyright = "© 2024 • All Rights Reserved"
				validConfig.TroubleText = "Questions? →"

				err := validConfig.Validate()
				Expect(err).To(BeNil())

				mailer := validConfig.NewMailer()
				Expect(mailer.GetName()).To(ContainSubstring("™"))
				Expect(mailer.GetCopyright()).To(ContainSubstring("•"))
				Expect(mailer.GetTroubleText()).To(ContainSubstring("→"))
			})
		})

		Context("when handling long strings", func() {
			It("should handle very long text values", func() {
				longString := string(make([]byte, 1000))
				for i := range longString {
					longString = longString[:i] + "a" + longString[i+1:]
				}

				validConfig.Copyright = longString
				mailer := validConfig.NewMailer()

				Expect(len(mailer.GetCopyright())).To(Equal(1000))
			})
		})
	})
})
