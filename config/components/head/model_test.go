/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package head_test

import (
	"context"
	"fmt"

	. "github.com/nabbar/golib/config/components/head"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	librtr "github.com/nabbar/golib/router/header"
)

// Model tests verify GetHeaders and SetHeaders operations
var _ = Describe("Model Operations", func() {
	var (
		ctx context.Context
		cpt CptHead
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	Describe("GetHeaders", func() {
		Context("on new component", func() {
			It("should return non-nil headers", func() {
				headers := cpt.GetHeaders()
				Expect(headers).NotTo(BeNil())
			})

			It("should return empty headers initially", func() {
				headers := cpt.GetHeaders()
				Expect(headers.Header()).To(BeEmpty())
			})

			It("should return consistent headers", func() {
				headers1 := cpt.GetHeaders()
				headers2 := cpt.GetHeaders()
				// Both should be empty initially
				Expect(headers1.Header()).To(BeEmpty())
				Expect(headers2.Header()).To(BeEmpty())
			})
		})

		Context("after setting headers", func() {
			It("should return set headers", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Test-Header", "test-value")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Test-Header")).To(Equal("test-value"))
			})

			It("should maintain multiple headers", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Header-1", "value-1")
				headers.Set("X-Header-2", "value-2")
				headers.Set("X-Header-3", "value-3")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Header-1")).To(Equal("value-1"))
				Expect(retrieved.Get("X-Header-2")).To(Equal("value-2"))
				Expect(retrieved.Get("X-Header-3")).To(Equal("value-3"))
			})
		})

		Context("with special header values", func() {
			It("should handle empty values", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Empty-Header", "")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Empty-Header")).To(Equal(""))
			})

			It("should handle long values", func() {
				longValue := ""
				for i := 0; i < 1000; i++ {
					longValue += "a"
				}

				headers := librtr.NewHeaders()
				headers.Set("X-Long-Header", longValue)
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Long-Header")).To(Equal(longValue))
			})

			It("should handle special characters", func() {
				specialValue := "value with spaces, special chars: !@#$%^&*()"
				headers := librtr.NewHeaders()
				headers.Set("X-Special-Header", specialValue)
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Special-Header")).To(Equal(specialValue))
			})
		})
	})

	Describe("SetHeaders", func() {
		Context("with valid headers", func() {
			It("should set headers successfully", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Custom-Header", "custom-value")

				Expect(func() {
					cpt.SetHeaders(headers)
				}).NotTo(Panic())

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Custom-Header")).To(Equal("custom-value"))
			})

			It("should replace existing headers", func() {
				// Set initial headers
				headers1 := librtr.NewHeaders()
				headers1.Set("X-Header-1", "value-1")
				cpt.SetHeaders(headers1)

				// Replace with new headers
				headers2 := librtr.NewHeaders()
				headers2.Set("X-Header-2", "value-2")
				cpt.SetHeaders(headers2)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Header-1")).To(BeEmpty())
				Expect(retrieved.Get("X-Header-2")).To(Equal("value-2"))
			})

			It("should handle multiple sets", func() {
				for i := 0; i < 10; i++ {
					headers := librtr.NewHeaders()
					headers.Set("X-Header", fmt.Sprintf("value-%d", i))
					cpt.SetHeaders(headers)

					retrieved := cpt.GetHeaders()
					Expect(retrieved.Get("X-Header")).To(Equal(fmt.Sprintf("value-%d", i)))
				}
			})
		})

		Context("with nil headers", func() {
			It("should handle nil headers", func() {
				Expect(func() {
					cpt.SetHeaders(nil)
				}).NotTo(Panic())
			})

			It("should initialize empty headers when nil", func() {
				cpt.SetHeaders(nil)
				headers := cpt.GetHeaders()
				Expect(headers).NotTo(BeNil())
				Expect(headers.Header()).To(BeEmpty())
			})

			It("should clear existing headers when set to nil", func() {
				// Set some headers
				headers := librtr.NewHeaders()
				headers.Set("X-Test", "value")
				cpt.SetHeaders(headers)

				// Set to nil
				cpt.SetHeaders(nil)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Header()).To(BeEmpty())
			})
		})

		Context("with complex header configurations", func() {
			It("should handle security headers", func() {
				headers := librtr.NewHeaders()
				headers.Set("Content-Security-Policy", "default-src 'self'")
				headers.Set("X-Frame-Options", "DENY")
				headers.Set("X-Content-Type-Options", "nosniff")
				headers.Set("Strict-Transport-Security", "max-age=31536000")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("Content-Security-Policy")).To(Equal("default-src 'self'"))
				Expect(retrieved.Get("X-Frame-Options")).To(Equal("DENY"))
				Expect(retrieved.Get("X-Content-Type-Options")).To(Equal("nosniff"))
				Expect(retrieved.Get("Strict-Transport-Security")).To(Equal("max-age=31536000"))
			})

			It("should handle CORS headers", func() {
				headers := librtr.NewHeaders()
				headers.Set("Access-Control-Allow-Origin", "*")
				headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
				headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("Access-Control-Allow-Origin")).To(Equal("*"))
				Expect(retrieved.Get("Access-Control-Allow-Methods")).To(Equal("GET, POST, PUT, DELETE"))
				Expect(retrieved.Get("Access-Control-Allow-Headers")).To(Equal("Content-Type, Authorization"))
			})

			It("should handle cache control headers", func() {
				headers := librtr.NewHeaders()
				headers.Set("Cache-Control", "no-cache, no-store, must-revalidate")
				headers.Set("Pragma", "no-cache")
				headers.Set("Expires", "0")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("Cache-Control")).To(Equal("no-cache, no-store, must-revalidate"))
				Expect(retrieved.Get("Pragma")).To(Equal("no-cache"))
				Expect(retrieved.Get("Expires")).To(Equal("0"))
			})
		})
	})

	Describe("Header Manipulation", func() {
		Context("working with header interface", func() {
			It("should support Add operation", func() {
				headers := cpt.GetHeaders()
				headers.Add("X-Test", "value1")
				headers.Add("X-Test", "value2")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				// Get returns first value
				Expect(retrieved.Get("X-Test")).To(Equal("value1"))
			})

			It("should support Set operation", func() {
				headers := cpt.GetHeaders()
				headers.Set("X-Test", "value1")
				headers.Set("X-Test", "value2")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Test")).To(Equal("value2"))
			})

			It("should support Del operation", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Test", "value")
				headers.Set("X-Keep", "keep-value")
				cpt.SetHeaders(headers)

				// Get headers, delete one, and set back
				headers = cpt.GetHeaders()
				headers.Del("X-Test")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Test")).To(BeEmpty())
				Expect(retrieved.Get("X-Keep")).To(Equal("keep-value"))
			})

			It("should support Get operation", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Test", "value")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				value := retrieved.Get("X-Test")
				Expect(value).To(Equal("value"))
			})

			It("should return empty string for non-existent headers", func() {
				headers := librtr.NewHeaders()
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				value := retrieved.Get("X-Non-Existent")
				Expect(value).To(BeEmpty())
			})
		})

		Context("case insensitivity", func() {
			It("should handle header name case insensitivity", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Test-Header", "value")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				// HTTP headers are case-insensitive
				value1 := retrieved.Get("X-Test-Header")
				value2 := retrieved.Get("x-test-header")
				value3 := retrieved.Get("X-TEST-HEADER")

				Expect(value1).To(Equal("value"))
				Expect(value2).To(Equal("value"))
				Expect(value3).To(Equal("value"))
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("with unusual inputs", func() {
			It("should handle very long header names", func() {
				longName := "X-"
				for i := 0; i < 500; i++ {
					longName += "A"
				}

				headers := librtr.NewHeaders()
				headers.Set(longName, "value")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get(longName)).To(Equal("value"))
			})

			It("should handle many headers", func() {
				headers := librtr.NewHeaders()
				for i := 0; i < 100; i++ {
					headers.Set(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("value-%d", i))
				}
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				for i := 0; i < 100; i++ {
					Expect(retrieved.Get(fmt.Sprintf("X-Header-%d", i))).To(Equal(fmt.Sprintf("value-%d", i)))
				}
			})

			It("should handle Unicode in header values", func() {
				headers := librtr.NewHeaders()
				headers.Set("X-Unicode", "Hello 世界 🌍")
				cpt.SetHeaders(headers)

				retrieved := cpt.GetHeaders()
				Expect(retrieved.Get("X-Unicode")).To(Equal("Hello 世界 🌍"))
			})
		})
	})
})

// Concurrent access tests for model operations
var _ = Describe("Model Concurrent Access", func() {
	var (
		ctx context.Context
		cpt CptHead
	)

	BeforeEach(func() {
		ctx = context.Background()
		cpt = New(ctx)
	})

	Context("concurrent SetHeaders", func() {
		It("should handle concurrent sets", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					headers := librtr.NewHeaders()
					headers.Set("X-Test", fmt.Sprintf("value-%d", index))
					cpt.SetHeaders(headers)
					done <- true
				}(i)
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}

			// Final state should be valid
			headers := cpt.GetHeaders()
			Expect(headers).NotTo(BeNil())
		})
	})

	Context("concurrent GetHeaders", func() {
		It("should handle concurrent gets", func() {
			// Set initial headers
			headers := librtr.NewHeaders()
			headers.Set("X-Shared", "shared-value")
			cpt.SetHeaders(headers)

			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					retrieved := cpt.GetHeaders()
					Expect(retrieved).NotTo(BeNil())
					Expect(retrieved.Get("X-Shared")).To(Equal("shared-value"))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Context("mixed concurrent operations", func() {
		It("should handle concurrent get and set", func() {
			done := make(chan bool, 20)

			// Concurrent sets
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					headers := librtr.NewHeaders()
					headers.Set("X-Test", fmt.Sprintf("value-%d", index))
					cpt.SetHeaders(headers)
					done <- true
				}(i)
			}

			// Concurrent gets
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					headers := cpt.GetHeaders()
					Expect(headers).NotTo(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 20; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})
