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
 *
 */

package errors_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gin Integration", func() {
	var (
		router *gin.Engine
		w      *httptest.ResponseRecorder
		ctx    *gin.Context
	)

	BeforeEach(func() {
		// Register test error messages
		if !ExistInMapMessage(TestErrorCode1) {
			RegisterIdFctMessage(TestErrorCode1, func(code CodeError) string {
				switch code {
				case TestErrorCode1:
					return "test error 1"
				case TestErrorCode2:
					return "test error 2"
				case TestErrorCode3:
					return "test error 3"
				default:
					return ""
				}
			})
		}

		// Set Gin to test mode
		gin.SetMode(gin.TestMode)
		router = gin.New()
		w = httptest.NewRecorder()
	})

	Describe("GinTonicAbort", func() {
		It("should abort with JSON response", func() {
			router.GET("/test", func(c *gin.Context) {
				r := NewDefaultReturn()
				r.SetError(404, "not found", "handler.go", 42)
				r.GinTonicAbort(c, http.StatusNotFound)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusNotFound))
			Expect(w.Body.String()).To(ContainSubstring("not found"))
		})

		It("should handle zero httpCode", func() {
			router.GET("/test", func(c *gin.Context) {
				r := NewDefaultReturn()
				r.SetError(500, "server error", "handler.go", 10)
				r.GinTonicAbort(c, 0) // 0 should default to 500
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should handle custom HTTP codes", func() {
			router.GET("/test", func(c *gin.Context) {
				r := NewDefaultReturn()
				r.SetError(400, "bad request", "handler.go", 20)
				r.GinTonicAbort(c, http.StatusBadRequest)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("should handle already aborted context", func() {
			ctx, _ = gin.CreateTestContext(w)
			ctx.Abort()

			r := NewDefaultReturn()
			r.SetError(500, "error", "file.go", 1)

			// Should not panic even if already aborted
			Expect(func() {
				r.GinTonicAbort(ctx, 500)
			}).ToNot(Panic())
		})

		It("should set response body with error details", func() {
			router.GET("/test", func(c *gin.Context) {
				r := NewDefaultReturn()
				r.SetError(403, "forbidden", "auth.go", 55)
				r.AddParent(401, "unauthorized", "auth.go", 30)
				r.GinTonicAbort(c, http.StatusForbidden)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusForbidden))
			body := w.Body.String()
			Expect(body).To(ContainSubstring("forbidden"))
		})
	})

	Describe("GinTonicErrorAbort", func() {
		It("should abort with errors added to gin context", func() {
			router.GET("/test", func(c *gin.Context) {
				r := NewDefaultReturn()
				r.SetError(500, "internal error", "server.go", 100)
				r.GinTonicErrorAbort(c, http.StatusInternalServerError)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should add error to gin context errors", func() {
			var capturedContext *gin.Context
			router.GET("/test", func(c *gin.Context) {
				capturedContext = c
				r := NewDefaultReturn()
				r.SetError(400, "validation error", "validator.go", 25)
				r.GinTonicErrorAbort(c, http.StatusBadRequest)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// Verify error was added to context
			Expect(capturedContext.Errors).ToNot(BeEmpty())
			Expect(capturedContext.Errors.Last().Error()).To(ContainSubstring("validation error"))
		})

		It("should add parent errors to gin context", func() {
			var capturedContext *gin.Context
			router.GET("/test", func(c *gin.Context) {
				capturedContext = c
				r := NewDefaultReturn()
				r.SetError(500, "main error", "main.go", 10)
				r.AddParent(400, "parent error 1", "parent.go", 20)
				r.AddParent(300, "parent error 2", "parent2.go", 30)
				r.GinTonicErrorAbort(c, http.StatusInternalServerError)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			// Should have multiple errors
			Expect(len(capturedContext.Errors)).To(BeNumerically(">", 1))
		})

		It("should handle already aborted context", func() {
			ctx, _ = gin.CreateTestContext(w)
			ctx.Abort()

			r := NewDefaultReturn()
			r.SetError(500, "error", "file.go", 1)

			// Should not panic
			Expect(func() {
				r.GinTonicErrorAbort(ctx, 500)
			}).ToNot(Panic())
		})

		It("should work without parent errors", func() {
			var capturedContext *gin.Context
			router.GET("/test", func(c *gin.Context) {
				capturedContext = c
				r := NewDefaultReturn()
				r.SetError(404, "not found", "handler.go", 5)
				r.GinTonicErrorAbort(c, http.StatusNotFound)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			Expect(capturedContext.Errors).ToNot(BeEmpty())
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})
	})

	Describe("ReturnGin Interface", func() {
		It("should implement ReturnGin interface", func() {
			r := NewDefaultReturn()
			var _ ReturnGin = r
			Expect(r).ToNot(BeNil())
		})

		It("should implement Return interface", func() {
			r := NewDefaultReturn()
			var _ Return = r
			Expect(r).ToNot(BeNil())
		})

		It("should work in real handler scenario", func() {
			router.POST("/api/resource", func(c *gin.Context) {
				// Simulate an error occurring
				err := TestErrorCode1.Error(nil)

				if err != nil {
					r := NewDefaultReturn()
					err.Return(r)
					r.GinTonicAbort(c, http.StatusInternalServerError)
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req, _ := http.NewRequest("POST", "/api/resource", nil)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
