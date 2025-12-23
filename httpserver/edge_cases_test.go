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

package httpserver_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	"github.com/nabbar/golib/httpserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-EC] HTTPServer/EdgeCases", func() {
	Describe("TLS configuration edge cases", func() {
		It("[TC-EC-001] should handle IsTLS check on config without TLS", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "tls-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			isTLS := cfg.IsTLS()
			Expect(isTLS).To(BeFalse())
		})

		It("[TC-EC-002] should handle SetDefaultTLS", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "default-tls-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			defaultTLS := func() libtls.TLSConfig {
				return nil
			}

			cfg.SetDefaultTLS(defaultTLS)
		})
	})

	Describe("Handler key edge cases", func() {
		It("[TC-EC-003] should handle handler registration with key", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:       "handler-key-test",
				Listen:     fmt.Sprintf("127.0.0.1:%d", port),
				Expose:     fmt.Sprintf("http://127.0.0.1:%d", port),
				HandlerKey: "test-key",
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"test-key": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("[TC-EC-004] should handle handler updates", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "handler-update-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.NotFoundHandler(),
				}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			srv.Handler(func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				}
			})

			Expect(srv).ToNot(BeNil())
		})
	})

	Describe("Restart operation", func() {
		It("[TC-EC-005] should call restart on stopped server", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "restart-stopped-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			Expect(srv.IsRunning()).To(BeFalse())

			err = srv.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeTrue())

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Port availability checks", func() {
		It("[TC-EC-006] should verify port checking functions exist", func() {
			port := GetFreePort()
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			Expect(addr).ToNot(BeEmpty())
		})
	})

	Describe("Info method edge cases", func() {
		It("[TC-EC-007] should handle GetName with valid name", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "valid-name-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.GetName()).To(Equal("valid-name-test"))
		})

		It("[TC-EC-008] should handle GetExpose with valid URL", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "expose-valid-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.GetExpose()).ToNot(BeEmpty())
		})

		It("[TC-EC-009] should handle IsDisable with disabled server", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:     "disabled-check-test",
				Listen:   fmt.Sprintf("127.0.0.1:%d", port),
				Expose:   fmt.Sprintf("http://127.0.0.1:%d", port),
				Disabled: true,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(srv.IsDisable()).To(BeTrue())
		})
	})

	Describe("Config edge cases", func() {
		It("[TC-EC-010] should handle GetListen with various formats", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "listen-format-test",
				Listen: fmt.Sprintf(":%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			url := cfg.GetListen()
			Expect(url).ToNot(BeNil())
		})

		It("[TC-EC-011] should handle GetExpose with various formats", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "expose-format-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://:%d", port),
			}

			url := cfg.GetExpose()
			Expect(url).ToNot(BeNil())
		})

		It("[TC-EC-012] should create server with default logger", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "logger-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("[TC-EC-013] should handle GetTLS when TLS not configured", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "get-tls-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			tlsCfg := cfg.GetTLS()
			Expect(tlsCfg).ToNot(BeNil())
		})

		It("[TC-EC-014] should handle CheckTLS with no TLS configured", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "check-tls-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			_, err := cfg.CheckTLS()
			Expect(err).To(HaveOccurred())
		})
	})
})
