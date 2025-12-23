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
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-HD] Server Handlers", func() {
	var (
		srv      Server
		err      error
		testPort string
	)

	BeforeEach(func() {
		testPort = fmt.Sprintf("127.0.0.1:%d", GetFreePort())
	})

	AfterEach(func() {
		if srv != nil {
			srv.Stop(context.Background())
		}
	})

	Describe("Handler Registration", func() {
		It("[TC-HD-010] should register and use custom handlers", func() {
			cfg := Config{
				Name:   "handler-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}

			// Register handler that returns specific content
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("test-response"))
				})
				mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("hello-world"))
				})
				return map[string]http.Handler{"": mux}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start(context.Background())
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			// Test /test endpoint
			resp, err := http.Get("http://" + testPort + "/test")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("test-response"))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Test /hello endpoint
			resp2, err := http.Get("http://" + testPort + "/hello")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()

			body2, err := io.ReadAll(resp2.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body2)).To(Equal("hello-world"))
		})

		It("[TC-HD-011] should handle multiple handler keys", func() {
			cfg := Config{
				Name:       "multi-handler-test",
				Listen:     testPort,
				Expose:     "http://" + testPort,
				HandlerKey: "handler1", // Specify which handler to use
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				mux1 := http.NewServeMux()
				mux1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("handler1"))
				})

				mux2 := http.NewServeMux()
				mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("handler2"))
				})

				return map[string]http.Handler{
					"handler1": mux1,
					"handler2": mux2,
				}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			// Verify the handler key
			Expect(cfg.GetHandlerKey()).To(Equal("handler1"))
		})

		It("[TC-HD-012] should update handlers dynamically", func() {
			cfg := Config{
				Name:   "dynamic-handler-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}

			initialHandler := func() map[string]http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("initial"))
				})
				return map[string]http.Handler{"": mux}
			}

			cfg.RegisterHandlerFunc(initialHandler)
			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			// Register new handler
			newHandler := func() map[string]http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("updated"))
				})
				return map[string]http.Handler{"": mux}
			}

			srv.Handler(newHandler)
			Expect(srv).NotTo(BeNil())
		})
	})

	Describe("Handler Validation", func() {
		It("[TC-HD-013] should handle requests with different methods", func() {
			cfg := Config{
				Name:   "method-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPost {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("post-ok"))
					} else {
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				})
				return map[string]http.Handler{"": mux}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start(context.Background())
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			// Test POST request
			resp, err := http.Post("http://"+testPort+"/post", "text/plain", nil)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("post-ok"))

			// Test GET request (should fail)
			resp2, err := http.Get("http://" + testPort + "/post")
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()

			Expect(resp2.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		})

		It("[TC-HD-014] should handle 404 for unknown paths", func() {
			cfg := Config{
				Name:   "404-test",
				Listen: testPort,
				Expose: "http://" + testPort,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NewServeMux()}
			})

			srv, err = New(cfg, nil)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start(context.Background())
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			resp, err := http.Get("http://" + testPort + "/nonexistent")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
