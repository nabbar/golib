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

package router_test

import (
	"net/http"
	"net/http/httptest"

	ginsdk "github.com/gin-gonic/gin"
	librtr "github.com/nabbar/golib/router"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router/Default", func() {
	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
	})

	Describe("DefaultGinInit", func() {
		It("should create a Gin engine with default middleware", func() {
			engine := librtr.DefaultGinInit()
			Expect(engine).ToNot(BeNil())
		})

		It("should have Logger and Recovery middleware", func() {
			engine := librtr.DefaultGinInit()
			Expect(engine).ToNot(BeNil())

			// Test that engine works
			engine.GET("/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("DefaultGinWithTrustyProxy", func() {
		It("should create engine without proxies", func() {
			engine := librtr.DefaultGinWithTrustyProxy(nil)
			Expect(engine).ToNot(BeNil())
		})

		It("should create engine with trusted proxies", func() {
			proxies := []string{"127.0.0.1", "192.168.1.1"}
			engine := librtr.DefaultGinWithTrustyProxy(proxies)
			Expect(engine).ToNot(BeNil())
		})

		It("should work with empty proxy list", func() {
			engine := librtr.DefaultGinWithTrustyProxy([]string{})
			Expect(engine).ToNot(BeNil())
		})
	})

	Describe("DefaultGinWithTrustedPlatform", func() {
		It("should create engine without platform", func() {
			engine := librtr.DefaultGinWithTrustedPlatform("")
			Expect(engine).ToNot(BeNil())
		})

		It("should create engine with trusted platform", func() {
			engine := librtr.DefaultGinWithTrustedPlatform("X-CDN-IP")
			Expect(engine).ToNot(BeNil())
			Expect(engine.TrustedPlatform).To(Equal("X-CDN-IP"))
		})
	})

	Describe("GinEngine", func() {
		It("should create engine without configuration", func() {
			engine, err := librtr.GinEngine("", nil...)
			Expect(err).ToNot(HaveOccurred())
			Expect(engine).ToNot(BeNil())
		})

		It("should create engine with trusted platform", func() {
			engine, err := librtr.GinEngine("X-Real-IP")
			Expect(err).ToNot(HaveOccurred())
			Expect(engine).ToNot(BeNil())
			Expect(engine.TrustedPlatform).To(Equal("X-Real-IP"))
		})

		It("should create engine with trusted proxies", func() {
			engine, err := librtr.GinEngine("", "127.0.0.1", "192.168.1.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(engine).ToNot(BeNil())
		})

		It("should create engine with both platform and proxies", func() {
			engine, err := librtr.GinEngine("X-Forwarded-For", "127.0.0.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(engine).ToNot(BeNil())
			Expect(engine.TrustedPlatform).To(Equal("X-Forwarded-For"))
		})
	})

	Describe("GinAddGlobalMiddleware", func() {
		It("should add middleware to engine", func() {
			engine := ginsdk.New()
			middleware := func(c *ginsdk.Context) {
				c.Set("middleware", "called")
				c.Next()
			}

			result := librtr.GinAddGlobalMiddleware(engine, middleware)
			Expect(result).To(Equal(engine))

			engine.GET("/test", func(c *ginsdk.Context) {
				val, _ := c.Get("middleware")
				c.String(http.StatusOK, val.(string))
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("called"))
		})

		It("should add multiple middlewares", func() {
			engine := ginsdk.New()

			middleware1 := func(c *ginsdk.Context) {
				c.Set("m1", "1")
				c.Next()
			}
			middleware2 := func(c *ginsdk.Context) {
				c.Set("m2", "2")
				c.Next()
			}

			librtr.GinAddGlobalMiddleware(engine, middleware1, middleware2)

			engine.GET("/test", func(c *ginsdk.Context) {
				v1, _ := c.Get("m1")
				v2, _ := c.Get("m2")
				c.String(http.StatusOK, v1.(string)+v2.(string))
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Body.String()).To(Equal("12"))
		})
	})

	Describe("SetGinHandler", func() {
		It("should convert function to HandlerFunc", func() {
			handler := librtr.SetGinHandler(func(c *ginsdk.Context) {
				c.String(http.StatusOK, "test")
			})

			Expect(handler).ToNot(BeNil())

			engine := ginsdk.New()
			engine.GET("/test", handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("test"))
		})
	})

	Describe("Handler", func() {
		It("should create HTTP handler from RouterList", func() {
			routerList := librtr.NewRouterList(librtr.DefaultGinInit)
			routerList.Register(http.MethodGet, "/test", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "ok")
			})

			handler := librtr.Handler(routerList)
			Expect(handler).ToNot(BeNil())

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			handler.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("ok"))
		})
	})

	Describe("Global Router Functions", func() {
		var engine *ginsdk.Engine

		BeforeEach(func() {
			engine = ginsdk.New()
		})

		It("should register route using RoutersRegister", func() {
			librtr.RoutersRegister(http.MethodGet, "/global", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "global")
			})

			librtr.RoutersHandler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/global", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("global"))
		})

		It("should register route in group using RoutersRegisterInGroup", func() {
			librtr.RoutersRegisterInGroup("/api", http.MethodGet, "/grouped", func(c *ginsdk.Context) {
				c.String(http.StatusOK, "grouped")
			})

			librtr.RoutersHandler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/grouped", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("grouped"))
		})
	})

	Describe("Constants", func() {
		It("should have EmptyHandlerGroup constant", func() {
			Expect(librtr.EmptyHandlerGroup).To(Equal("<nil>"))
		})

		It("should have GinContextStartUnixNanoTime constant", func() {
			Expect(librtr.GinContextStartUnixNanoTime).To(Equal("gin-ctx-start-unix-nano-time"))
		})

		It("should have GinContextRequestPath constant", func() {
			Expect(librtr.GinContextRequestPath).To(Equal("gin-ctx-request-path"))
		})

		It("should have GinContextRequestUser constant", func() {
			Expect(librtr.GinContextRequestUser).To(Equal("gin-ctx-request-user"))
		})
	})
})
