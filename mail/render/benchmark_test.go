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
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmarks", func() {
	Describe("Creation Benchmarks", func() {
		Context("mailer creation performance", func() {
			It("should benchmark New() creation", func() {
				experiment := gmeasure.NewExperiment("Mailer Creation")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("new", func() {
						mailer := render.New()
						Expect(mailer).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

				stats := experiment.GetStats("new")
				AddReportEntry("Mean creation time", stats.DurationFor(gmeasure.StatMean))
			})

			It("should benchmark Config.NewMailer() creation", func() {
				experiment := gmeasure.NewExperiment("Config Creation")
				AddReportEntry(experiment.Name, experiment)

				config := render.Config{
					Theme:       "default",
					Direction:   "ltr",
					Name:        "Test",
					Link:        "https://example.com",
					Logo:        "https://example.com/logo.png",
					Copyright:   "© 2024",
					TroubleText: "Help",
					Body:        hermes.Body{Name: "User"},
				}

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("config-new", func() {
						mailer := config.NewMailer()
						Expect(mailer).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 1000})

				stats := experiment.GetStats("config-new")
				AddReportEntry("Mean config creation time", stats.DurationFor(gmeasure.StatMean))
			})
		})
	})

	Describe("Clone Benchmarks", func() {
		Context("cloning performance", func() {
			It("should benchmark simple clone", func() {
				experiment := gmeasure.NewExperiment("Simple Clone")
				AddReportEntry(experiment.Name, experiment)

				mailer := render.New()
				mailer.SetName("Test")
				body := &hermes.Body{Name: "User"}
				mailer.SetBody(body)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("clone", func() {
						clone := mailer.Clone()
						Expect(clone).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 1000})

				stats := experiment.GetStats("clone")
				AddReportEntry("Mean simple clone time", stats.DurationFor(gmeasure.StatMean))
			})

			It("should benchmark complex clone", func() {
				experiment := gmeasure.NewExperiment("Complex Clone")
				AddReportEntry(experiment.Name, experiment)

				mailer := render.New()
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Intro 1", "Intro 2", "Intro 3"},
					Outros: []string{"Outro 1", "Outro 2"},
					Dictionary: []hermes.Entry{
						{Key: "K1", Value: "V1"},
						{Key: "K2", Value: "V2"},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{{Key: "C1", Value: "V1"}, {Key: "C2", Value: "V2"}},
							{{Key: "C1", Value: "V3"}, {Key: "C2", Value: "V4"}},
						},
					}},
					Actions: []hermes.Action{
						{
							Instructions: "Test",
							Button:       hermes.Button{Text: "Click", Link: "https://example.com"},
						},
					},
				}
				mailer.SetBody(body)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("clone-complex", func() {
						clone := mailer.Clone()
						Expect(clone).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 500})

				stats := experiment.GetStats("clone-complex")
				AddReportEntry("Mean complex clone time", stats.DurationFor(gmeasure.StatMean))
			})
		})
	})

	Describe("Generation Benchmarks", func() {
		var mailer render.Mailer

		BeforeEach(func() {
			mailer = render.New()
			mailer.SetName("Test Company")
			mailer.SetLink("https://example.com")
			mailer.SetLogo("https://example.com/logo.png")
			mailer.SetCopyright("© 2024")
			mailer.SetTroubleText("Contact us")
		})

		Context("HTML generation performance", func() {
			It("should benchmark simple HTML generation", func() {
				experiment := gmeasure.NewExperiment("Simple HTML")
				AddReportEntry(experiment.Name, experiment)

				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Welcome!"},
				}
				mailer.SetBody(body)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("html-simple", func() {
						buf, err := mailer.GenerateHTML()
						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				stats := experiment.GetStats("html-simple")
				AddReportEntry("Mean simple HTML time", stats.DurationFor(gmeasure.StatMean))
				AddReportEntry("Max simple HTML time", stats.DurationFor(gmeasure.StatMax))
			})

			It("should benchmark complex HTML generation", func() {
				experiment := gmeasure.NewExperiment("Complex HTML")
				AddReportEntry(experiment.Name, experiment)

				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Welcome!", "Here's your report"},
					Dictionary: []hermes.Entry{
						{Key: "ID", Value: "123"},
						{Key: "Date", Value: "2024-01-01"},
					},
					Tables: []hermes.Table{{
						Data: [][]hermes.Entry{
							{{Key: "Item", Value: "A"}, {Key: "Price", Value: "$10"}},
							{{Key: "Item", Value: "B"}, {Key: "Price", Value: "$20"}},
							{{Key: "Item", Value: "C"}, {Key: "Price", Value: "$30"}},
						},
					}},
					Actions: []hermes.Action{
						{
							Instructions: "View details",
							Button:       hermes.Button{Text: "View", Link: "https://example.com"},
						},
					},
					Outros: []string{"Thank you", "Best regards"},
				}
				mailer.SetBody(body)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("html-complex", func() {
						buf, err := mailer.GenerateHTML()
						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				stats := experiment.GetStats("html-complex")
				AddReportEntry("Mean complex HTML time", stats.DurationFor(gmeasure.StatMean))
				AddReportEntry("Max complex HTML time", stats.DurationFor(gmeasure.StatMax))
			})
		})

		Context("plain text generation performance", func() {
			It("should benchmark plain text generation", func() {
				experiment := gmeasure.NewExperiment("Plain Text")
				AddReportEntry(experiment.Name, experiment)

				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Welcome!"},
				}
				mailer.SetBody(body)

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("text-simple", func() {
						buf, err := mailer.GeneratePlainText()
						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				stats := experiment.GetStats("text-simple")
				AddReportEntry("Mean plain text time", stats.DurationFor(gmeasure.StatMean))
			})
		})

		Context("theme comparison", func() {
			It("should compare generation time between themes", func() {
				experiment := gmeasure.NewExperiment("Theme Comparison")
				AddReportEntry(experiment.Name, experiment)

				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Test email"},
					Actions: []hermes.Action{
						{Button: hermes.Button{Text: "Click", Link: "https://example.com"}},
					},
				}

				experiment.Sample(func(idx int) {
					mailer.SetTheme(render.ThemeDefault)
					mailer.SetBody(body)
					experiment.MeasureDuration("theme-default", func() {
						buf, err := mailer.GenerateHTML()
						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				experiment.Sample(func(idx int) {
					mailer.SetTheme(render.ThemeFlat)
					mailer.SetBody(body)
					experiment.MeasureDuration("theme-flat", func() {
						buf, err := mailer.GenerateHTML()
						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				defaultStats := experiment.GetStats("theme-default")
				flatStats := experiment.GetStats("theme-flat")
				AddReportEntry("Mean default theme time", defaultStats.DurationFor(gmeasure.StatMean))
				AddReportEntry("Mean flat theme time", flatStats.DurationFor(gmeasure.StatMean))
			})
		})
	})

	Describe("ParseData Benchmarks", func() {
		Context("template parsing performance", func() {
			It("should benchmark simple ParseData", func() {
				experiment := gmeasure.NewExperiment("Simple ParseData")
				AddReportEntry(experiment.Name, experiment)

				data := map[string]string{
					"{{user}}":    "John",
					"{{company}}": "Test Inc",
				}

				experiment.Sample(func(idx int) {
					mailer := render.New()
					body := &hermes.Body{
						Name:   "{{user}}",
						Intros: []string{"Hello {{user}}"},
					}
					mailer.SetBody(body)
					mailer.SetName("{{company}}")

					experiment.MeasureDuration("parse-simple", func() {
						mailer.ParseData(data)
					})
				}, gmeasure.SamplingConfig{N: 500})

				stats := experiment.GetStats("parse-simple")
				AddReportEntry("Mean simple parse time", stats.DurationFor(gmeasure.StatMean))
			})

			It("should benchmark complex ParseData", func() {
				experiment := gmeasure.NewExperiment("Complex ParseData")
				AddReportEntry(experiment.Name, experiment)

				data := map[string]string{
					"{{user}}":    "John",
					"{{company}}": "Test Inc",
					"{{code}}":    "123456",
					"{{link}}":    "https://example.com",
					"{{email}}":   "user@example.com",
				}

				experiment.Sample(func(idx int) {
					mailer := render.New()
					body := &hermes.Body{
						Name:   "{{user}}",
						Intros: []string{"Hello {{user}}", "Your code: {{code}}"},
						Dictionary: []hermes.Entry{
							{Key: "Company", Value: "{{company}}"},
							{Key: "Email", Value: "{{email}}"},
						},
						Tables: []hermes.Table{{
							Data: [][]hermes.Entry{
								{{Key: "Link", Value: "{{link}}"}},
							},
						}},
						Actions: []hermes.Action{
							{
								Instructions: "Visit {{link}}",
								Button:       hermes.Button{Text: "Visit", Link: "{{link}}"},
							},
						},
					}
					mailer.SetBody(body)
					mailer.SetName("{{company}}")

					experiment.MeasureDuration("parse-complex", func() {
						mailer.ParseData(data)
					})
				}, gmeasure.SamplingConfig{N: 500})

				stats := experiment.GetStats("parse-complex")
				AddReportEntry("Mean complex parse time", stats.DurationFor(gmeasure.StatMean))
			})
		})
	})

	Describe("Complete Workflow Benchmark", func() {
		Context("end-to-end performance", func() {
			It("should benchmark complete email generation workflow", func() {
				experiment := gmeasure.NewExperiment("Complete Workflow")
				AddReportEntry(experiment.Name, experiment)

				config := render.Config{
					Theme:       "default",
					Direction:   "ltr",
					Name:        "{{company}}",
					Link:        "https://example.com",
					Logo:        "https://example.com/logo.png",
					Copyright:   "© 2024 {{company}}",
					TroubleText: "Contact us",
					Body: hermes.Body{
						Name:   "{{user}}",
						Intros: []string{"Welcome {{user}}!"},
						Actions: []hermes.Action{
							{
								Instructions: "Verify your account:",
								Button:       hermes.Button{Text: "Verify", Link: "{{link}}"},
							},
						},
					},
				}

				data := map[string]string{
					"{{company}}": "Test Inc",
					"{{user}}":    "John Doe",
					"{{link}}":    "https://example.com/verify",
				}

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("workflow", func() {
						mailer := config.NewMailer()
						mailer.ParseData(data)
						htmlBuf, htmlErr := mailer.GenerateHTML()
						textBuf, textErr := mailer.GeneratePlainText()

						Expect(htmlErr).To(BeNil())
						Expect(textErr).To(BeNil())
						Expect(htmlBuf).ToNot(BeNil())
						Expect(textBuf).ToNot(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 100})

				stats := experiment.GetStats("workflow")
				AddReportEntry("Mean workflow time", stats.DurationFor(gmeasure.StatMean))
				AddReportEntry("Max workflow time", stats.DurationFor(gmeasure.StatMax))
			})
		})
	})

	Describe("Parsing Benchmarks", func() {
		Context("theme and direction parsing", func() {
			It("should benchmark ParseTheme", func() {
				experiment := gmeasure.NewExperiment("ParseTheme")
				AddReportEntry(experiment.Name, experiment)

				themes := []string{"default", "flat", "Default", "FLAT", "unknown"}

				experiment.Sample(func(idx int) {
					theme := themes[idx%len(themes)]
					experiment.MeasureDuration("parse-theme", func() {
						_ = render.ParseTheme(theme)
					})
				}, gmeasure.SamplingConfig{N: 1000})

				stats := experiment.GetStats("parse-theme")
				AddReportEntry("Mean theme parse time", stats.DurationFor(gmeasure.StatMean))
			})

			It("should benchmark ParseTextDirection", func() {
				experiment := gmeasure.NewExperiment("ParseTextDirection")
				AddReportEntry(experiment.Name, experiment)

				directions := []string{"ltr", "rtl", "left-to-right", "right-to-left", "LTR"}

				experiment.Sample(func(idx int) {
					dir := directions[idx%len(directions)]
					experiment.MeasureDuration("parse-direction", func() {
						_ = render.ParseTextDirection(dir)
					})
				}, gmeasure.SamplingConfig{N: 1000})

				stats := experiment.GetStats("parse-direction")
				AddReportEntry("Mean direction parse time", stats.DurationFor(gmeasure.StatMean))
			})
		})
	})

	Describe("Validation Benchmark", func() {
		Context("config validation performance", func() {
			It("should benchmark Config.Validate()", func() {
				experiment := gmeasure.NewExperiment("Config Validation")
				AddReportEntry(experiment.Name, experiment)

				validConfig := render.Config{
					Theme:       "default",
					Direction:   "ltr",
					Name:        "Test",
					Link:        "https://example.com",
					Logo:        "https://example.com/logo.png",
					Copyright:   "© 2024",
					TroubleText: "Help",
					Body:        hermes.Body{Name: "User"},
				}

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("validate", func() {
						err := validConfig.Validate()
						Expect(err).To(BeNil())
					})
				}, gmeasure.SamplingConfig{N: 500})

				stats := experiment.GetStats("validate")
				AddReportEntry("Mean validation time", stats.DurationFor(gmeasure.StatMean))
			})
		})
	})
})
