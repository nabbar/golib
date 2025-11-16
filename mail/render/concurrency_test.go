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
	"sync"

	"github.com/go-hermes/hermes/v2"
	"github.com/nabbar/golib/mail/render"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrency", func() {
	Describe("Concurrent Clone Operations", func() {
		Context("when cloning from multiple goroutines", func() {
			It("should safely clone mailer concurrently", func() {
				mailer := render.New()
				mailer.SetName("Original")
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Intro 1", "Intro 2"},
				}
				mailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 100

				results := make([]render.Mailer, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer GinkgoRecover()
						defer wg.Done()

						clone := mailer.Clone()
						results[index] = clone

						// Verify clone has correct values
						Expect(clone.GetName()).To(Equal("Original"))
						Expect(clone.GetBody().Name).To(Equal("User"))
						Expect(clone.GetBody().Intros).To(HaveLen(2))
					}(i)
				}

				wg.Wait()

				// Verify all clones were created
				for _, clone := range results {
					Expect(clone).ToNot(BeNil())
				}
			})

			It("should create independent clones under concurrent load", func() {
				mailer := render.New()
				body := &hermes.Body{
					Intros: []string{"Original"},
				}
				mailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 50

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer GinkgoRecover()
						defer wg.Done()

						clone := mailer.Clone()

						// Modify the clone
						cloneBody := clone.GetBody()
						cloneBody.Intros = append(cloneBody.Intros, "Modified")

						// Clone should have 2 intros
						Expect(len(cloneBody.Intros)).To(BeNumerically(">=", 2))
					}(i)
				}

				wg.Wait()

				// Original should still have 1 intro
				originalBody := mailer.GetBody()
				Expect(originalBody.Intros).To(HaveLen(1))
				Expect(originalBody.Intros[0]).To(Equal("Original"))
			})
		})
	})

	Describe("Concurrent Read Operations", func() {
		Context("when reading from multiple goroutines", func() {
			It("should safely read mailer properties concurrently", func() {
				mailer := render.New()
				mailer.SetName("Test Company")
				mailer.SetLink("https://example.com")
				mailer.SetLogo("https://example.com/logo.png")
				mailer.SetCopyright("© 2024")
				mailer.SetTroubleText("Contact us")
				mailer.SetTheme(render.ThemeFlat)
				mailer.SetTextDirection(render.RightToLeft)

				var wg sync.WaitGroup
				const numGoroutines = 100

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						// Read all properties
						name := mailer.GetName()
						link := mailer.GetLink()
						logo := mailer.GetLogo()
						copyright := mailer.GetCopyright()
						troubleText := mailer.GetTroubleText()
						theme := mailer.GetTheme()
						direction := mailer.GetTextDirection()
						body := mailer.GetBody()

						// Verify values
						Expect(name).To(Equal("Test Company"))
						Expect(link).To(Equal("https://example.com"))
						Expect(logo).To(Equal("https://example.com/logo.png"))
						Expect(copyright).To(Equal("© 2024"))
						Expect(troubleText).To(Equal("Contact us"))
						Expect(theme).To(Equal(render.ThemeFlat))
						Expect(direction).To(Equal(render.RightToLeft))
						Expect(body).ToNot(BeNil())
					}()
				}

				wg.Wait()
			})
		})
	})

	Describe("Concurrent Generation", func() {
		Context("when generating HTML from multiple goroutines", func() {
			It("should safely generate HTML concurrently using independent clones", func() {
				mailer := render.New()
				mailer.SetName("Test Company")
				mailer.SetLink("https://example.com")
				mailer.SetLogo("https://example.com/logo.png")
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Welcome!"},
				}
				mailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 50

				results := make(chan bool, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						// Clone to avoid concurrent modifications
						clone := mailer.Clone()
						buf, err := clone.GenerateHTML()

						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
						Expect(buf.Len()).To(BeNumerically(">", 0))

						results <- true
					}()
				}

				wg.Wait()
				close(results)

				count := 0
				for range results {
					count++
				}

				Expect(count).To(Equal(numGoroutines))
			})

			It("should safely generate plain text concurrently using independent clones", func() {
				mailer := render.New()
				mailer.SetName("Test Company")
				body := &hermes.Body{
					Name:   "User",
					Intros: []string{"Test"},
				}
				mailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 50

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						clone := mailer.Clone()
						buf, err := clone.GeneratePlainText()

						Expect(err).To(BeNil())
						Expect(buf).ToNot(BeNil())
					}()
				}

				wg.Wait()
			})
		})
	})

	Describe("Concurrent Configuration Updates", func() {
		Context("when multiple goroutines update different mailers", func() {
			It("should handle concurrent updates on independent mailers", func() {
				var wg sync.WaitGroup
				const numGoroutines = 50

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer GinkgoRecover()
						defer wg.Done()

						mailer := render.New()
						mailer.SetName("Mailer " + string(rune(index)))
						mailer.SetTheme(render.ThemeFlat)

						body := &hermes.Body{
							Name: "User " + string(rune(index)),
						}
						mailer.SetBody(body)

						// Generate to ensure no race conditions
						_, err := mailer.GenerateHTML()
						Expect(err).To(BeNil())
					}(i)
				}

				wg.Wait()
			})
		})
	})

	Describe("Concurrent ParseData", func() {
		Context("when parsing data in multiple goroutines", func() {
			It("should safely parse data on independent clones", func() {
				baseMailer := render.New()
				body := &hermes.Body{
					Name:   "{{user}}",
					Intros: []string{"Hello {{user}}"},
				}
				baseMailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 50

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer GinkgoRecover()
						defer wg.Done()

						// Clone for independent operation
						mailer := baseMailer.Clone()

						data := map[string]string{
							"{{user}}": "User " + string(rune(index)),
						}

						mailer.ParseData(data)

						body := mailer.GetBody()
						Expect(body.Name).To(ContainSubstring("User"))
					}(i)
				}

				wg.Wait()

				// Original should still have template variables
				originalBody := baseMailer.GetBody()
				Expect(originalBody.Name).To(Equal("{{user}}"))
			})
		})
	})

	Describe("Mixed Concurrent Operations", func() {
		Context("when performing various operations concurrently", func() {
			It("should handle mixed operations safely", func() {
				mailer := render.New()
				mailer.SetName("Base")
				body := &hermes.Body{
					Name:   "{{user}}",
					Intros: []string{"Welcome"},
				}
				mailer.SetBody(body)

				var wg sync.WaitGroup
				const numGoroutines = 100

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)

					// Distribute different operations
					switch i % 4 {
					case 0:
						// Clone operation
						go func() {
							defer GinkgoRecover()
							defer wg.Done()
							clone := mailer.Clone()
							Expect(clone).ToNot(BeNil())
						}()
					case 1:
						// Read operation
						go func() {
							defer GinkgoRecover()
							defer wg.Done()
							name := mailer.GetName()
							Expect(name).To(Equal("Base"))
						}()
					case 2:
						// Generate HTML
						go func() {
							defer GinkgoRecover()
							defer wg.Done()
							clone := mailer.Clone()
							_, err := clone.GenerateHTML()
							Expect(err).To(BeNil())
						}()
					case 3:
						// Parse data on clone
						go func() {
							defer GinkgoRecover()
							defer wg.Done()
							clone := mailer.Clone()
							clone.ParseData(map[string]string{
								"{{user}}": "Test",
							})
							Expect(clone.GetBody().Name).To(Equal("Test"))
						}()
					}
				}

				wg.Wait()

				// Original should remain unchanged
				Expect(mailer.GetName()).To(Equal("Base"))
				Expect(mailer.GetBody().Name).To(Equal("{{user}}"))
			})
		})
	})

	Describe("High Load Stress Test", func() {
		Context("under high concurrent load", func() {
			It("should remain stable with many concurrent operations", func() {
				config := render.Config{
					Theme:       "default",
					Direction:   "ltr",
					Name:        "Stress Test",
					Link:        "https://example.com",
					Logo:        "https://example.com/logo.png",
					Copyright:   "© 2024",
					TroubleText: "Help",
					Body: hermes.Body{
						Name:   "{{user}}",
						Intros: []string{"Stress test email"},
					},
				}

				baseMailer := config.NewMailer()

				var wg sync.WaitGroup
				const highLoad = 200

				successCount := make(chan bool, highLoad)

				for i := 0; i < highLoad; i++ {
					wg.Add(1)
					go func(index int) {
						defer GinkgoRecover()
						defer wg.Done()

						clone := baseMailer.Clone()
						clone.ParseData(map[string]string{
							"{{user}}": "User" + string(rune(index)),
						})

						htmlBuf, htmlErr := clone.GenerateHTML()
						textBuf, textErr := clone.GeneratePlainText()

						if htmlErr == nil && textErr == nil &&
							htmlBuf != nil && textBuf != nil {
							successCount <- true
						}
					}(i)
				}

				wg.Wait()
				close(successCount)

				count := 0
				for range successCount {
					count++
				}

				// Should have high success rate
				Expect(count).To(BeNumerically(">=", highLoad*0.95))
			})
		})
	})
})
