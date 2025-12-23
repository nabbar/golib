/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package types_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/nabbar/golib/httpserver/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-HT] Handler Types", func() {
	Describe("BadHandler", func() {
		It("[TC-HT-001] should create bad handler", func() {
			handler := NewBadHandler()

			Expect(handler).ToNot(BeNil())
		})

		It("[TC-HT-002] should return 500 status code", func() {
			handler := NewBadHandler()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("[TC-HT-003] should handle different HTTP methods", func() {
			handler := NewBadHandler()

			methods := []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodPatch,
			}

			for _, method := range methods {
				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			}
		})

		It("[TC-HT-004] should handle different paths", func() {
			handler := NewBadHandler()

			paths := []string{
				"/",
				"/api",
				"/api/v1",
				"/some/deep/path",
			}

			for _, path := range paths {
				req := httptest.NewRequest(http.MethodGet, path, nil)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			}
		})
	})

	Describe("FuncHandler Type", func() {
		It("[TC-HT-007] should define handler function returning map", func() {
			var handlerFunc FuncHandler

			handlerFunc = func() map[string]http.Handler {
				return map[string]http.Handler{
					"test": http.NotFoundHandler(),
				}
			}

			result := handlerFunc()
			Expect(result).ToNot(BeNil())
			Expect(result).To(HaveKey("test"))
		})

		It("[TC-HT-009] should allow returning empty map", func() {
			var handlerFunc FuncHandler

			handlerFunc = func() map[string]http.Handler {
				return map[string]http.Handler{}
			}

			result := handlerFunc()
			Expect(result).ToNot(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("[TC-HT-010] should allow returning nil", func() {
			var handlerFunc FuncHandler

			handlerFunc = func() map[string]http.Handler {
				return nil
			}

			result := handlerFunc()
			Expect(result).To(BeNil())
		})

		It("[TC-HT-011] should support multiple handler keys", func() {
			var handlerFunc FuncHandler

			handlerFunc = func() map[string]http.Handler {
				return map[string]http.Handler{
					"api-v1":  http.NotFoundHandler(),
					"api-v2":  http.NotFoundHandler(),
					"web":     http.NotFoundHandler(),
					"default": NewBadHandler(),
				}
			}

			result := handlerFunc()
			Expect(result).To(HaveLen(4))
			Expect(result).To(HaveKey("api-v1"))
			Expect(result).To(HaveKey("api-v2"))
			Expect(result).To(HaveKey("web"))
			Expect(result).To(HaveKey("default"))
		})
	})

	Describe("BadHandler Direct Usage", func() {
		It("[TC-HT-005] should work with http.Handler interface", func() {
			var handler http.Handler = &BadHandler{}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("[TC-HT-006] should create multiple handler instances", func() {
			handler1 := NewBadHandler()
			handler2 := NewBadHandler()

			Expect(handler1).ToNot(BeNil())
			Expect(handler2).ToNot(BeNil())

			req1 := httptest.NewRequest(http.MethodGet, "/", nil)
			w1 := httptest.NewRecorder()
			handler1.ServeHTTP(w1, req1)

			req2 := httptest.NewRequest(http.MethodPost, "/", nil)
			w2 := httptest.NewRecorder()
			handler2.ServeHTTP(w2, req2)

			Expect(w1.Code).To(Equal(http.StatusInternalServerError))
			Expect(w2.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
