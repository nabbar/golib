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

package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	ginsdk "github.com/gin-gonic/gin"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	rtrauth "github.com/nabbar/golib/router/auth"
	rtrhdr "github.com/nabbar/golib/router/authheader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth/Authorization", func() {
	var (
		engine *ginsdk.Engine
		log    liblog.Logger
		auth   rtrauth.Authorization
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		engine = ginsdk.New()
		log = liblog.New(context.Background())
	})

	Describe("NewAuthorization", func() {
		It("should create a new Authorization instance", func() {
			checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
				return rtrhdr.AuthCodeSuccess, nil
			}

			logFunc := func() liblog.Logger {
				return log
			}

			auth = rtrauth.NewAuthorization(logFunc, "Bearer", checkFunc)
			Expect(auth).ToNot(BeNil())
		})

		It("should create Authorization with nil logger", func() {
			checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
				return rtrhdr.AuthCodeSuccess, nil
			}

			auth = rtrauth.NewAuthorization(nil, "Bearer", checkFunc)
			Expect(auth).ToNot(BeNil())
		})
	})

	Describe("Handler", func() {
		Context("when authorization header is missing", func() {
			It("should return 401 Unauthorized", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeSuccess, nil
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "Bearer", checkFunc)
				engine.GET("/test", auth.Handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				Expect(w.Header().Get(rtrhdr.HeaderAuthRequire)).To(Equal(rtrhdr.HeaderAuthReal))
			})
		})

		Context("when authorization header is empty", func() {
			It("should return 401 Unauthorized", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeSuccess, nil
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "Bearer", checkFunc)
				engine.GET("/test", auth.Handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when authorization type does not match", func() {
			It("should return 401 Unauthorized", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeSuccess, nil
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "Bearer", checkFunc)
				engine.GET("/test", auth.Handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "Basic dGVzdDp0ZXN0")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when authorization is successful", func() {
			It("should call registered handlers", func() {
				called := false
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					Expect(authHeader).To(Equal("token123"))
					return rtrhdr.AuthCodeSuccess, nil
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)
				handler := auth.Register(func(c *ginsdk.Context) {
					called = true
					c.String(http.StatusOK, "success")
				})

				engine.GET("/test", handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer token123")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(called).To(BeTrue())
				Expect(w.Body.String()).To(Equal("success"))
			})

			It("should handle case-insensitive auth type", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeSuccess, nil
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)
				handler := auth.Register(func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				engine.GET("/test", handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "bearer token123")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when authorization check returns AuthCodeRequire", func() {
			It("should return 401 Unauthorized", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeRequire, liberr.Error(nil)
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "Bearer", checkFunc)
				handler := auth.Register(func(c *ginsdk.Context) {
					c.String(http.StatusOK, "should not reach")
				})

				engine.GET("/test", handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer invalid")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when authorization check returns AuthCodeForbidden", func() {
			It("should return 403 Forbidden", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCodeForbidden, liberr.Error(nil)
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)
				handler := auth.Register(func(c *ginsdk.Context) {
					c.String(http.StatusOK, "should not reach")
				})

				engine.GET("/test", handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer forbidden")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("when authorization check returns unknown code", func() {
			It("should return 500 Internal Server Error", func() {
				checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
					return rtrhdr.AuthCode(99), liberr.Error(nil)
				}

				logFunc := func() liblog.Logger {
					return log
				}

				auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)
				handler := auth.Register(func(c *ginsdk.Context) {
					c.String(http.StatusOK, "should not reach")
				})

				engine.GET("/test", handler)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer token")
				engine.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("Register", func() {
		It("should register handlers and return Handler function", func() {
			checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
				return rtrhdr.AuthCodeSuccess, nil
			}

			logFunc := func() liblog.Logger {
				return log
			}

			auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)

			handler1 := func(c *ginsdk.Context) {
				c.Set("handler1", "called")
				c.Next()
			}
			handler2 := func(c *ginsdk.Context) {
				val, _ := c.Get("handler1")
				c.String(http.StatusOK, val.(string))
			}

			authHandler := auth.Register(handler1, handler2)
			Expect(authHandler).ToNot(BeNil())

			engine.GET("/test", authHandler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer token")
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("called"))
		})
	})

	Describe("Append", func() {
		It("should append handlers to existing list", func() {
			checkFunc := func(authHeader string) (rtrhdr.AuthCode, liberr.Error) {
				return rtrhdr.AuthCodeSuccess, nil
			}

			logFunc := func() liblog.Logger {
				return log
			}

			auth = rtrauth.NewAuthorization(logFunc, "BEARER", checkFunc)

			handler1 := func(c *ginsdk.Context) {
				c.Set("count", 1)
				c.Next()
			}

			auth.Register(handler1)

			handler2 := func(c *ginsdk.Context) {
				count, _ := c.Get("count")
				c.Set("count", count.(int)+1)
				c.Next()
			}
			handler3 := func(c *ginsdk.Context) {
				count, _ := c.Get("count")
				c.String(http.StatusOK, "%d", count.(int))
			}

			auth.Append(handler2, handler3)

			engine.GET("/test", auth.Handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(rtrhdr.HeaderAuthSend, "Bearer token")
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("2"))
		})
	})
})
