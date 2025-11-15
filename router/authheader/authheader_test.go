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

package authheader_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	ginsdk "github.com/gin-gonic/gin"
	rtrhdr "github.com/nabbar/golib/router/authheader"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AuthHeader", func() {
	var (
		engine *ginsdk.Engine
	)

	BeforeEach(func() {
		ginsdk.SetMode(ginsdk.TestMode)
		engine = ginsdk.New()
	})

	Describe("AuthCode Constants", func() {
		It("should have correct AuthCode values", func() {
			Expect(uint8(rtrhdr.AuthCodeSuccess)).To(Equal(uint8(0)))
			Expect(uint8(rtrhdr.AuthCodeRequire)).To(Equal(uint8(1)))
			Expect(uint8(rtrhdr.AuthCodeForbidden)).To(Equal(uint8(2)))
		})
	})

	Describe("Header Constants", func() {
		It("should have correct header names", func() {
			Expect(rtrhdr.HeaderAuthRequire).To(Equal("WWW-Authenticate"))
			Expect(rtrhdr.HeaderAuthSend).To(Equal("Authorization"))
			Expect(rtrhdr.HeaderAuthReal).To(Equal("Basic realm=LDAP Authorization Required"))
		})
	})

	Describe("AuthRequire", func() {
		It("should set 401 status and WWW-Authenticate header", func() {
			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthRequire(c, nil)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			Expect(w.Header().Get(rtrhdr.HeaderAuthRequire)).To(Equal(rtrhdr.HeaderAuthReal))
		})

		It("should add error to context when error is provided", func() {
			var contextErrors []*ginsdk.Error

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthRequire(c, errors.New("auth failed"))
				contextErrors = c.Errors
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			Expect(contextErrors).To(HaveLen(1))
			Expect(contextErrors[0].Err.Error()).To(Equal("auth failed"))
			Expect(contextErrors[0].Type).To(Equal(ginsdk.ErrorTypePrivate))
		})

		It("should not add error when error is nil", func() {
			var contextErrors []*ginsdk.Error

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthRequire(c, nil)
				contextErrors = c.Errors
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			Expect(contextErrors).To(HaveLen(0))
		})

		It("should abort handler chain", func() {
			nextCalled := false

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthRequire(c, nil)
			}, func(c *ginsdk.Context) {
				nextCalled = true
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			Expect(nextCalled).To(BeFalse())
		})
	})

	Describe("AuthForbidden", func() {
		It("should set 403 status", func() {
			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthForbidden(c, nil)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusForbidden))
		})

		It("should add error to context when error is provided", func() {
			var contextErrors []*ginsdk.Error

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthForbidden(c, errors.New("forbidden"))
				contextErrors = c.Errors
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusForbidden))
			Expect(contextErrors).To(HaveLen(1))
			Expect(contextErrors[0].Err.Error()).To(Equal("forbidden"))
			Expect(contextErrors[0].Type).To(Equal(ginsdk.ErrorTypePrivate))
		})

		It("should not add error when error is nil", func() {
			var contextErrors []*ginsdk.Error

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthForbidden(c, nil)
				contextErrors = c.Errors
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusForbidden))
			Expect(contextErrors).To(HaveLen(0))
		})

		It("should abort handler chain", func() {
			nextCalled := false

			engine.GET("/test", func(c *ginsdk.Context) {
				rtrhdr.AuthForbidden(c, nil)
			}, func(c *ginsdk.Context) {
				nextCalled = true
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusForbidden))
			Expect(nextCalled).To(BeFalse())
		})
	})

	Describe("Integration", func() {
		It("should work in authentication flow", func() {
			engine.GET("/protected", func(c *ginsdk.Context) {
				authHeader := c.GetHeader("Authorization")
				if authHeader == "" {
					rtrhdr.AuthRequire(c, errors.New("missing auth"))
					return
				}
				if authHeader != "Bearer valid-token" {
					rtrhdr.AuthForbidden(c, errors.New("invalid token"))
					return
				}
				c.String(http.StatusOK, "authorized")
			})

			// Test missing auth
			w1 := httptest.NewRecorder()
			req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			engine.ServeHTTP(w1, req1)
			Expect(w1.Code).To(Equal(http.StatusUnauthorized))

			// Test invalid auth
			w2 := httptest.NewRecorder()
			req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			req2.Header.Set("Authorization", "Bearer invalid-token")
			engine.ServeHTTP(w2, req2)
			Expect(w2.Code).To(Equal(http.StatusForbidden))

			// Test valid auth
			w3 := httptest.NewRecorder()
			req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			req3.Header.Set("Authorization", "Bearer valid-token")
			engine.ServeHTTP(w3, req3)
			Expect(w3.Code).To(Equal(http.StatusOK))
			Expect(w3.Body.String()).To(Equal("authorized"))
		})
	})
})
