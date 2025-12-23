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

	"github.com/nabbar/golib/httpserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-MON] HTTPServer/HealthMonitor", func() {
	Describe("Server lifecycle tracking", func() {
		It("[TC-MON-007] should track server state through lifecycle", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "health-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("[TC-MON-008] should detect running state correctly", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "health-test-running",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeTrue())

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("MonitorName", func() {
		It("[TC-MON-009] should return unique monitor name", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "monitor-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			name := srv.MonitorName()
			Expect(name).ToNot(BeEmpty())
			Expect(name).To(ContainSubstring(fmt.Sprintf("%d", port)))
		})

		It("[TC-MON-010] should include bind address in monitor name", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "monitor-state-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			name := srv.MonitorName()
			Expect(name).To(ContainSubstring("127.0.0.1"))
		})

		It("[TC-MON-011] should work when server is disabled", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:     "monitor-disabled-test",
				Listen:   fmt.Sprintf("127.0.0.1:%d", port),
				Expose:   fmt.Sprintf("http://127.0.0.1:%d", port),
				Disabled: true,
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			name := srv.MonitorName()
			Expect(name).ToNot(BeEmpty())
		})
	})

	Describe("Uptime tracking", func() {
		It("[TC-MON-012] should return zero uptime for non-running server", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "uptime-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			uptime := srv.Uptime()
			Expect(uptime).To(Equal(time.Duration(0)))
		})

		It("[TC-MON-013] should track uptime for running server", func() {
			port := GetFreePort()
			cfg := httpserver.Config{
				Name:   "uptime-running-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", port),
				Expose: fmt.Sprintf("http://127.0.0.1:%d", port),
			}

			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{"": http.NotFoundHandler()}
			})

			srv, err := httpserver.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)

			uptime := srv.Uptime()
			Expect(uptime).To(BeNumerically(">", 100*time.Millisecond))

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
