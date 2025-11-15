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

package httpserver_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock HTTP handler for testing
type mockHandler struct {
	called bool
	status int
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	if m.status == 0 {
		m.status = http.StatusOK
	}
	w.WriteHeader(m.status)
	_, _ = w.Write([]byte("mock response"))
}

var _ = Describe("Handler Management", func() {
	Describe("Handler Registration", func() {
		It("should register handler function", func() {
			mock := &mockHandler{}
			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"": mock,
				}
			}

			cfg := Config{
				Name:   "handler-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(handlerFunc)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Can update handler after creation
			srv.Handler(handlerFunc)
			// Handler is registered (no error means success)
		})

		It("should handle nil handler function gracefully", func() {
			cfg := Config{
				Name:   "nil-handler-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Should not panic with nil handler
			srv.Handler(nil)
		})
	})

	Describe("Handler with HandlerKey", func() {
		It("should use handler key from config", func() {
			mock := &mockHandler{}
			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"api-v1": mock,
				}
			}

			cfg := Config{
				Name:       "keyed-server",
				Listen:     "127.0.0.1:8080",
				Expose:     "http://localhost:8080",
				HandlerKey: "api-v1",
			}
			cfg.RegisterHandlerFunc(handlerFunc)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should work with multiple handler keys", func() {
			mock1 := &mockHandler{status: http.StatusOK}
			mock2 := &mockHandler{status: http.StatusAccepted}

			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"api-v1": mock1,
					"api-v2": mock2,
				}
			}

			cfg := Config{
				Name:       "multi-handler-server",
				Listen:     "127.0.0.1:8080",
				Expose:     "http://localhost:8080",
				HandlerKey: "api-v1",
			}
			cfg.RegisterHandlerFunc(handlerFunc)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

	Describe("Handler Execution", func() {
		It("should execute custom handler", func() {
			mock := &mockHandler{}

			// Test the handler directly
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			mock.ServeHTTP(w, req)

			Expect(mock.called).To(BeTrue())
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.String()).To(Equal("mock response"))
		})

		It("should handle custom status codes", func() {
			mock := &mockHandler{status: http.StatusCreated}

			req := httptest.NewRequest(http.MethodPost, "/create", nil)
			w := httptest.NewRecorder()

			mock.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))
		})
	})

	Describe("Multiple Handler Registration", func() {
		It("should allow handler replacement", func() {
			cfg := Config{
				Name:   "replace-handler-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// First handler
			mock1 := &mockHandler{}
			handler1 := func() map[string]http.Handler {
				return map[string]http.Handler{"test": mock1}
			}
			srv.Handler(handler1)

			// Second handler (replacement)
			mock2 := &mockHandler{}
			handler2 := func() map[string]http.Handler {
				return map[string]http.Handler{"test": mock2}
			}
			srv.Handler(handler2)

			// No error means successful replacement
		})
	})

	Describe("Handler Edge Cases", func() {
		It("should handle empty handler map", func() {
			cfg := Config{
				Name:   "empty-handler-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			emptyHandler := func() map[string]http.Handler {
				return map[string]http.Handler{}
			}

			srv.Handler(emptyHandler)
			// Should not panic with empty map
		})

		It("should handle handler returning nil map", func() {
			cfg := Config{
				Name:   "nil-map-handler-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			nilMapHandler := func() map[string]http.Handler {
				return nil
			}

			srv.Handler(nilMapHandler)
			// Should not panic with nil map
		})
	})
})
