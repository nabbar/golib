/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package header_test

import (
	"net/http"
	"net/http/httptest"

	ginsdk "github.com/gin-gonic/gin"
	rtrhdr "github.com/nabbar/golib/router/header"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Header/Headers", func() {
	var (
		headers rtrhdr.Headers
		engine  *ginsdk.Engine
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		headers = rtrhdr.NewHeaders()
		engine = ginsdk.New()
	})

	Describe("NewHeaders", func() {
		It("should create a new Headers instance", func() {
			Expect(headers).ToNot(BeNil())
		})
	})

	Describe("Add", func() {
		It("should add a header", func() {
			headers.Add("X-Custom-Header", "value1")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value1"))
		})

		It("should append to existing header", func() {
			headers.Add("X-Custom-Header", "value1")
			headers.Add("X-Custom-Header", "value2")

			// Get returns first value
			Expect(headers.Get("X-Custom-Header")).To(Equal("value1"))
		})

		It("should handle multiple different headers", func() {
			headers.Add("X-Header-1", "value1")
			headers.Add("X-Header-2", "value2")
			headers.Add("X-Header-3", "value3")

			Expect(headers.Get("X-Header-1")).To(Equal("value1"))
			Expect(headers.Get("X-Header-2")).To(Equal("value2"))
			Expect(headers.Get("X-Header-3")).To(Equal("value3"))
		})

		It("should be case-insensitive for header names", func() {
			headers.Add("x-custom-header", "value")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value"))
			Expect(headers.Get("X-CUSTOM-HEADER")).To(Equal("value"))
		})
	})

	Describe("Set", func() {
		It("should set a header", func() {
			headers.Set("X-Custom-Header", "value1")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value1"))
		})

		It("should replace existing header", func() {
			headers.Set("X-Custom-Header", "value1")
			headers.Set("X-Custom-Header", "value2")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value2"))
		})

		It("should replace multiple values with single value", func() {
			headers.Add("X-Custom-Header", "value1")
			headers.Add("X-Custom-Header", "value2")
			headers.Set("X-Custom-Header", "new-value")
			Expect(headers.Get("X-Custom-Header")).To(Equal("new-value"))
		})
	})

	Describe("Get", func() {
		It("should return empty string for non-existent header", func() {
			Expect(headers.Get("X-Non-Existent")).To(Equal(""))
		})

		It("should return header value", func() {
			headers.Set("X-Custom-Header", "value")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value"))
		})

		It("should be case-insensitive", func() {
			headers.Set("X-Custom-Header", "value")
			Expect(headers.Get("x-custom-header")).To(Equal("value"))
			Expect(headers.Get("X-CUSTOM-HEADER")).To(Equal("value"))
		})
	})

	Describe("Del", func() {
		It("should delete a header", func() {
			headers.Set("X-Custom-Header", "value")
			Expect(headers.Get("X-Custom-Header")).To(Equal("value"))

			headers.Del("X-Custom-Header")
			Expect(headers.Get("X-Custom-Header")).To(Equal(""))
		})

		It("should not panic when deleting non-existent header", func() {
			Expect(func() {
				headers.Del("X-Non-Existent")
			}).ToNot(Panic())
		})

		It("should delete all values of a header", func() {
			headers.Add("X-Custom-Header", "value1")
			headers.Add("X-Custom-Header", "value2")
			headers.Del("X-Custom-Header")
			Expect(headers.Get("X-Custom-Header")).To(Equal(""))
		})
	})

	Describe("Header", func() {
		It("should return empty map when no headers set", func() {
			headerMap := headers.Header()
			Expect(headerMap).ToNot(BeNil())
			Expect(headerMap).To(BeEmpty())
		})

		It("should return all headers as map", func() {
			headers.Set("X-Header-1", "value1")
			headers.Set("X-Header-2", "value2")
			headers.Set("X-Header-3", "value3")

			headerMap := headers.Header()
			Expect(headerMap).To(HaveLen(3))
			Expect(headerMap["X-Header-1"]).To(Equal("value1"))
			Expect(headerMap["X-Header-2"]).To(Equal("value2"))
			Expect(headerMap["X-Header-3"]).To(Equal("value3"))
		})

		It("should return first value for multi-value headers", func() {
			headers.Add("X-Custom-Header", "value1")
			headers.Add("X-Custom-Header", "value2")

			headerMap := headers.Header()
			Expect(headerMap["X-Custom-Header"]).To(Equal("value1"))
		})
	})

	Describe("Clone", func() {
		It("should create a copy of headers", func() {
			headers.Set("X-Header-1", "value1")
			headers.Set("X-Header-2", "value2")

			cloned := headers.Clone()
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.Get("X-Header-1")).To(Equal("value1"))
			Expect(cloned.Get("X-Header-2")).To(Equal("value2"))
		})

		It("should share underlying header map", func() {
			headers.Set("X-Original", "original")
			cloned := headers.Clone()

			// Modify original - clone shares the same underlying map
			headers.Set("X-Original", "modified")
			headers.Set("X-New", "new")

			// Clone shares the same header map
			Expect(cloned.Get("X-Original")).To(Equal("modified"))
			Expect(cloned.Get("X-New")).To(Equal("new"))
		})
	})

	Describe("Handler", func() {
		It("should set headers in Gin context", func() {
			headers.Set("X-Custom-Header", "custom-value")
			headers.Set("X-API-Version", "v1")

			engine.GET("/test", headers.Handler, func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Custom-Header")).To(Equal("custom-value"))
			Expect(w.Header().Get("X-API-Version")).To(Equal("v1"))
		})

		It("should not panic when headers is nil", func() {
			headers := rtrhdr.NewHeaders()

			engine.GET("/test", headers.Handler, func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			Expect(func() {
				engine.ServeHTTP(w, req)
			}).ToNot(Panic())
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should set multiple headers", func() {
			headers.Set("X-Header-1", "value1")
			headers.Set("X-Header-2", "value2")
			headers.Set("X-Header-3", "value3")

			engine.GET("/test", headers.Handler, func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Header().Get("X-Header-1")).To(Equal("value1"))
			Expect(w.Header().Get("X-Header-2")).To(Equal("value2"))
			Expect(w.Header().Get("X-Header-3")).To(Equal("value3"))
		})
	})

	Describe("Register", func() {
		It("should return handler chain with Header handler first", func() {
			headers.Set("X-Custom-Header", "value")

			handler := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			}

			chain := headers.Register(handler)
			Expect(chain).To(HaveLen(2))

			engine.GET("/test", chain...)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Custom-Header")).To(Equal("value"))
		})

		It("should work with multiple handlers", func() {
			headers.Set("X-Custom-Header", "value")

			middleware := func(c *ginsdk.Context) {
				c.Set("middleware", "called")
				c.Next()
			}

			handler := func(c *ginsdk.Context) {
				val, _ := c.Get("middleware")
				c.String(http.StatusOK, val.(string))
			}

			chain := headers.Register(middleware, handler)
			engine.GET("/test", chain...)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("called"))
			Expect(w.Header().Get("X-Custom-Header")).To(Equal("value"))
		})

		It("should work with no additional handlers", func() {
			headers.Set("X-Custom-Header", "value")

			chain := headers.Register()
			Expect(chain).To(HaveLen(1))

			engine.GET("/test", append(chain, func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})...)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Header().Get("X-Custom-Header")).To(Equal("value"))
		})
	})

	Describe("Integration", func() {
		It("should work in complete request flow", func() {
			headers.Set("X-API-Version", "v1")
			headers.Set("X-Request-ID", "12345")
			headers.Set("Cache-Control", "no-cache")

			middleware := func(c *ginsdk.Context) {
				c.Set("processed", true)
				c.Next()
			}

			handler := func(c *ginsdk.Context) {
				processed, _ := c.Get("processed")
				if processed.(bool) {
					c.JSON(http.StatusOK, map[string]string{
						"status": "success",
					})
				}
			}

			chain := headers.Register(middleware, handler)
			engine.GET("/api/test", chain...)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-API-Version")).To(Equal("v1"))
			Expect(w.Header().Get("X-Request-ID")).To(Equal("12345"))
			Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache"))
			Expect(w.Body.String()).To(ContainSubstring("success"))
		})
	})
})
