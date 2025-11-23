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
 */

package static_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration", func() {
	var handler any

	BeforeEach(func() {
		handler = newTestStatic()
	})

	Describe("Download", func() {
		Context("when setting download flag", func() {
			It("should mark file as downloadable", func() {
				h := handler.(staticDownload)
				h.SetDownload("testdata/test.txt", true)
				Expect(h.IsDownload("testdata/test.txt")).To(BeTrue())
			})

			It("should mark file as not downloadable", func() {
				h := handler.(staticDownload)
				h.SetDownload("testdata/test.txt", false)
				Expect(h.IsDownload("testdata/test.txt")).To(BeFalse())
			})

			It("should toggle download flag", func() {
				h := handler.(staticDownload)

				h.SetDownload("testdata/test.txt", true)
				Expect(h.IsDownload("testdata/test.txt")).To(BeTrue())

				h.SetDownload("testdata/test.txt", false)
				Expect(h.IsDownload("testdata/test.txt")).To(BeFalse())
			})
		})

		Context("when file does not exist", func() {
			It("should not set download flag", func() {
				h := handler.(staticDownload)

				h.SetDownload("testdata/nonexistent.txt", true)
				Expect(h.IsDownload("testdata/nonexistent.txt")).To(BeFalse())
			})
		})

		Context("when checking download flag", func() {
			It("should return false for unset files", func() {
				h := handler.(staticDownload)

				Expect(h.IsDownload("testdata/test.txt")).To(BeFalse())
			})

			It("should return false for nonexistent files", func() {
				h := handler.(staticDownload)

				Expect(h.IsDownload("testdata/nonexistent.txt")).To(BeFalse())
			})
		})

		Context("when path is empty", func() {
			It("should not set download flag", func() {
				h := handler.(staticDownload)

				h.SetDownload("", true)
				Expect(h.IsDownload("")).To(BeFalse())
			})
		})
	})

	Describe("Index", func() {
		Context("when setting index file", func() {
			It("should set index for route", func() {
				h := handler.(staticIndex)

				h.SetIndex("", "/", "testdata/index.html")
				Expect(h.GetIndex("", "/")).To(Equal("testdata/index.html"))
				Expect(h.IsIndex("testdata/index.html")).To(BeTrue())
			})

			It("should set index for route with group", func() {
				h := handler.(staticIndex)

				h.SetIndex("/api", "/v1", "testdata/index.html")
				Expect(h.GetIndex("/api", "/v1")).To(Equal("testdata/index.html"))
			})

			It("should support multiple routes for same file", func() {
				h := handler.(staticIndex)

				h.SetIndex("", "/", "testdata/index.html")
				h.SetIndex("", "/home", "testdata/index.html")

				Expect(h.GetIndex("", "/")).To(Equal("testdata/index.html"))
				Expect(h.GetIndex("", "/home")).To(Equal("testdata/index.html"))
				Expect(h.IsIndexForRoute("testdata/index.html", "", "/")).To(BeTrue())
				Expect(h.IsIndexForRoute("testdata/index.html", "", "/home")).To(BeTrue())
			})
		})

		Context("when file does not exist", func() {
			It("should not set index", func() {
				h := handler.(staticIndex)

				h.SetIndex("", "/", "testdata/nonexistent.html")
				Expect(h.GetIndex("", "/")).To(BeEmpty())
				Expect(h.IsIndex("testdata/nonexistent.html")).To(BeFalse())
			})
		})

		Context("when checking index", func() {
			It("should return false for non-index files", func() {
				h := handler.(staticIndex)

				Expect(h.IsIndex("testdata/test.txt")).To(BeFalse())
			})

			It("should return empty string for unset routes", func() {
				h := handler.(staticIndex)

				Expect(h.GetIndex("", "/notfound")).To(BeEmpty())
			})
		})

		Context("when path is empty", func() {
			It("should not set index", func() {
				h := handler.(staticIndex)

				h.SetIndex("", "/", "")
				Expect(h.GetIndex("", "/")).To(BeEmpty())
			})
		})
	})

	Describe("Redirect", func() {
		Context("when setting redirect", func() {
			It("should redirect route to destination", func() {
				h := handler.(staticRedirect)

				h.SetRedirect("", "/old", "", "/new")
				Expect(h.GetRedirect("", "/old")).To(Equal("/new"))
				Expect(h.IsRedirect("", "/old")).To(BeTrue())
			})

			It("should redirect with groups", func() {
				h := handler.(staticRedirect)
				h.SetRedirect("/api", "/v1", "/api", "/v2")
				Expect(h.GetRedirect("/api", "/v1")).To(Equal("/api/v2"))
			})

			It("should handle cross-group redirects", func() {
				h := handler.(staticRedirect)
				h.SetRedirect("/old", "/path", "/new", "/path")
				Expect(h.GetRedirect("/old", "/path")).To(Equal("/new/path"))
			})

			It("should update redirect destination", func() {
				h := handler.(staticRedirect)

				h.SetRedirect("", "/path", "", "/dest1")
				Expect(h.GetRedirect("", "/path")).To(Equal("/dest1"))

				h.SetRedirect("", "/path", "", "/dest2")
				Expect(h.GetRedirect("", "/path")).To(Equal("/dest2"))
			})
		})

		Context("when checking redirect", func() {
			It("should return false for non-redirect routes", func() {
				h := handler.(staticRedirect)
				Expect(h.IsRedirect("", "/notredirect")).To(BeFalse())
			})

			It("should return empty string for unset routes", func() {
				h := handler.(staticRedirect)
				Expect(h.GetRedirect("", "/notfound")).To(BeEmpty())
			})
		})
	})

	Describe("Specific", func() {
		Context("when setting specific handler", func() {
			It("should set handler for route", func() {
				h := handler.(staticSpecific)
				h.SetSpecific("", "/custom", customMiddlewareOK("custom", nil))

				retrieved := h.GetSpecific("", "/custom")
				Expect(retrieved).ToNot(BeNil())
			})

			It("should set handler with group", func() {
				h := handler.(staticSpecific)
				h.SetSpecific("/api", "/custom", customMiddlewareOK("custom", nil))

				retrieved := h.GetSpecific("/api", "/custom")
				Expect(retrieved).ToNot(BeNil())
			})

			It("should update specific handler", func() {
				h := handler.(staticSpecific)

				h.SetSpecific("", "/path", customMiddlewareOK("handler1", nil))
				Expect(h.GetSpecific("", "/path")).ToNot(BeNil())

				h.SetSpecific("", "/path", customMiddlewareOK("handler2", nil))
				Expect(h.GetSpecific("", "/path")).ToNot(BeNil())
			})
		})

		Context("when checking specific handler", func() {
			It("should return nil for unset routes", func() {
				h := handler.(staticSpecific)
				Expect(h.GetSpecific("", "/notfound")).To(BeNil())
			})

			It("should return nil for different groups", func() {
				h := handler.(staticSpecific)
				h.SetSpecific("/api", "/path", customMiddlewareOK("custom", nil))
				Expect(h.GetSpecific("/other", "/path")).To(BeNil())
			})
		})
	})

	Describe("Multiple Configurations", func() {
		Context("when combining configurations", func() {
			It("should support multiple configurations on same file", func() {
				h := handler.(staticConfig)

				h.SetDownload("testdata/index.html", true)
				h.SetIndex("", "/", "testdata/index.html")

				Expect(h.IsDownload("testdata/index.html")).To(BeTrue())
				Expect(h.IsIndex("testdata/index.html")).To(BeTrue())
				Expect(h.GetIndex("", "/")).To(Equal("testdata/index.html"))
			})

			It("should keep configurations independent", func() {
				h := handler.(staticConfig)

				h.SetDownload("testdata/test.txt", true)
				h.SetIndex("", "/", "testdata/index.html")

				Expect(h.IsDownload("testdata/test.txt")).To(BeTrue())
				Expect(h.IsDownload("testdata/index.html")).To(BeFalse())
				Expect(h.IsIndex("testdata/test.txt")).To(BeFalse())
				Expect(h.IsIndex("testdata/index.html")).To(BeTrue())
			})
		})
	})
})
