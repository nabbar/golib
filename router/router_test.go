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

var _ = Describe("Router/RouterList", func() {
	var (
		routerList librtr.RouterList
		engine     *ginsdk.Engine
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		routerList = librtr.NewRouterList(librtr.DefaultGinInit)
		engine = routerList.Engine()
	})

	Describe("NewRouterList", func() {
		It("should create a new RouterList instance", func() {
			Expect(routerList).ToNot(BeNil())
		})

		It("should create RouterList with custom init function", func() {
			customInit := func() *ginsdk.Engine {
				e := ginsdk.New()
				e.Use(ginsdk.Recovery())
				return e
			}
			rl := librtr.NewRouterList(customInit)
			Expect(rl).ToNot(BeNil())
			Expect(rl.Engine()).ToNot(BeNil())
		})
	})

	Describe("Register", func() {
		It("should register a route without group", func() {
			called := false
			handler := func(c *ginsdk.Context) {
				called = true
				c.String(http.StatusOK, "test")
			}

			routerList.Register(http.MethodGet, "/test", handler)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(called).To(BeTrue())
			Expect(w.Body.String()).To(Equal("test"))
		})

		It("should register multiple routes", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "route1")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "route2")
			}

			routerList.Register(http.MethodGet, "/route1", handler1)
			routerList.Register(http.MethodPost, "/route2", handler2)
			routerList.Handler(engine)

			// Test route 1
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/route1", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w1.Body.String()).To(Equal("route1"))

			// Test route 2
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodPost, "/route2", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))
			Expect(w2.Body.String()).To(Equal("route2"))
		})

		It("should register route with multiple handlers", func() {
			middleware := func(c *ginsdk.Context) {
				c.Set("middleware", "called")
				c.Next()
			}
			handler := func(c *ginsdk.Context) {
				val, _ := c.Get("middleware")
				c.String(http.StatusOK, val.(string))
			}

			routerList.Register(http.MethodGet, "/multi", middleware, handler)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/multi", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("called"))
		})
	})

	Describe("RegisterInGroup", func() {
		It("should register route in a group", func() {
			handler := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "grouped")
			}

			routerList.RegisterInGroup("/api", http.MethodGet, "/test", handler)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("grouped"))
		})

		It("should register multiple routes in same group", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "user")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "profile")
			}

			routerList.RegisterInGroup("/api", http.MethodGet, "/user", handler1)
			routerList.RegisterInGroup("/api", http.MethodGet, "/profile", handler2)
			routerList.Handler(engine)

			// Test /api/user
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/api/user", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w1.Body.String()).To(Equal("user"))

			// Test /api/profile
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodGet, "/api/profile", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))
			Expect(w2.Body.String()).To(Equal("profile"))
		})

		It("should register routes in different groups", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "v1")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "v2")
			}

			routerList.RegisterInGroup("/api/v1", http.MethodGet, "/test", handler1)
			routerList.RegisterInGroup("/api/v2", http.MethodGet, "/test", handler2)
			routerList.Handler(engine)

			// Test v1
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/api/v1/test", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w1.Body.String()).To(Equal("v1"))

			// Test v2
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodGet, "/api/v2/test", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))
			Expect(w2.Body.String()).To(Equal("v2"))
		})

		It("should treat empty group as no group", func() {
			handler := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "no-group")
			}

			routerList.RegisterInGroup("", http.MethodGet, "/test", handler)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("no-group"))
		})
	})

	Describe("RegisterMergeInGroup", func() {
		It("should register new route when not exists", func() {
			handler := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "new")
			}

			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test", handler)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("new"))
		})

		It("should merge handlers when route exists", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "original")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "merged")
			}

			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test", handler1)
			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test", handler2)
			routerList.Handler(engine)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			// Should use merged handler (handler2)
			Expect(w.Body.String()).To(Equal("merged"))
		})

		It("should not merge different methods", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "get")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "post")
			}

			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test", handler1)
			routerList.RegisterMergeInGroup("/api", http.MethodPost, "/test", handler2)
			routerList.Handler(engine)

			// Test GET
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w1.Body.String()).To(Equal("get"))

			// Test POST
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodPost, "/api/test", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))
			Expect(w2.Body.String()).To(Equal("post"))
		})

		It("should not merge different paths", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "path1")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "path2")
			}

			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test1", handler1)
			routerList.RegisterMergeInGroup("/api", http.MethodGet, "/test2", handler2)
			routerList.Handler(engine)

			// Test path1
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/api/test1", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))
			Expect(w1.Body.String()).To(Equal("path1"))

			// Test path2
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodGet, "/api/test2", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))
			Expect(w2.Body.String()).To(Equal("path2"))
		})
	})

	Describe("Engine", func() {
		It("should return a valid Gin engine", func() {
			eng := routerList.Engine()
			Expect(eng).ToNot(BeNil())
		})

		It("should use default init when init is nil", func() {
			rl := librtr.NewRouterList(nil)
			eng := rl.Engine()
			Expect(eng).ToNot(BeNil())
		})
	})

	Describe("Handler", func() {
		It("should apply all registered routes to engine", func() {
			handler1 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "route1")
			}
			handler2 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "route2")
			}
			handler3 := func(c *ginsdk.Context) {
				c.String(http.StatusOK, "grouped")
			}

			routerList.Register(http.MethodGet, "/route1", handler1)
			routerList.Register(http.MethodPost, "/route2", handler2)
			routerList.RegisterInGroup("/api", http.MethodGet, "/test", handler3)

			routerList.Handler(engine)

			// Verify all routes work
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/route1", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusOK))

			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodPost, "/route2", nil)
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusOK))

			w3 := httptest.NewRecorder()
			req3, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
			engine.ServeHTTP(w3, req3)
			Expect(w3.Code).To(Equal(http.StatusOK))
		})
	})
})
