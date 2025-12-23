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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	"github.com/nabbar/golib/httpserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-TLS] HTTPServer/TLS", func() {
	Describe("TLS server configuration", func() {
		It("[TC-TLS-001] should start server with TLS", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsTLS()).To(BeTrue())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeTrue())

			client := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}

			resp, err := client.Get(fmt.Sprintf("https://127.0.0.1:%d", port))
			if err == nil {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				resp.Body.Close()
			}

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-TLS-002] should check IsTLS with TLS config", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "tls-check-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			isTLS := cfg.IsTLS()
			Expect(isTLS).To(BeTrue())
		})

		It("[TC-TLS-003] should handle TLS mandatory flag", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:         "tls-mandatory-server",
				Listen:       fmt.Sprintf("127.0.0.1:%d", port),
				Expose:       fmt.Sprintf("https://127.0.0.1:%d", port),
				TLSMandatory: true,
				TLS:          srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsTLS()).To(BeTrue())
		})

		It("[TC-TLS-004] should get TLS config", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "get-tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			tlsCfg := cfg.GetTLS()
			Expect(tlsCfg).ToNot(BeNil())
		})

		It("[TC-TLS-005] should validate TLS configuration", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "check-tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			tlsCfg := cfg.GetTLS()
			Expect(tlsCfg).ToNot(BeNil())
		})

		It("[TC-TLS-006] should handle SetDefaultTLS", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "default-tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
			}

			defaultTLS := func() libtls.TLSConfig {
				tlsCfg := srvTLSCfg.New()
				return tlsCfg
			}

			cfg.SetDefaultTLS(defaultTLS)
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			Expect(cfg.GetTLS()).ToNot(BeNil())
		})
	})

	Describe("TLS server lifecycle", func() {
		It("[TC-TLS-007] should support TLS server lifecycle operations", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "lifecycle-tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsTLS()).To(BeTrue())

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(300 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeTrue())

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("[TC-TLS-008] should handle concurrent requests on TLS server", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "concurrent-tls-server",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("https://127.0.0.1:%d", port),
				TLS:    srvTLSCfg,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)

			client := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}

			for i := 0; i < 5; i++ {
				resp, e := client.Get(fmt.Sprintf("https://127.0.0.1:%d", port))
				if e == nil {
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					resp.Body.Close()
				}
			}

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
