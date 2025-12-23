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

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-MON] Server Monitoring", func() {
	Describe("Monitor Name", func() {
		It("[TC-MON-001] should return monitor name for server", func() {
			cfg := Config{
				Name:   "monitor-test-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Monitor name should be based on server name or bind address
			monitorName := srv.MonitorName()
			Expect(monitorName).ToNot(BeEmpty())
			// Monitor name contains either the server name or bind address
			Expect(monitorName).To(Or(
				ContainSubstring("monitor-test-server"),
				ContainSubstring("127.0.0.1:8080"),
			))
		})

		It("[TC-MON-002] should return unique monitor names for different servers", func() {
			cfg1 := Config{
				Name:   "server-1",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg1.RegisterHandlerFunc(defaultHandler)

			cfg2 := Config{
				Name:   "server-2",
				Listen: "127.0.0.1:8081",
				Expose: "http://localhost:8081",
			}
			cfg2.RegisterHandlerFunc(defaultHandler)

			srv1, err := New(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			srv2, err := New(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			name1 := srv1.MonitorName()
			name2 := srv2.MonitorName()

			// Monitor names should be different
			Expect(name1).ToNot(Equal(name2))
		})
	})

	Describe("Monitor Interface", func() {
		It("[TC-MON-003] should have monitor method available", func() {
			cfg := Config{
				Name:   "monitor-interface-test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Verify server has monitor name available
			monitorName := srv.MonitorName()
			Expect(monitorName).ToNot(BeEmpty())
		})

		It("[TC-MON-004] should handle monitor with custom configuration", func() {
			cfg := Config{
				Name:   "custom-monitor-test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("OK"))
					}),
				}
			})

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Check that monitor name is available
			monitorName := srv.MonitorName()
			Expect(monitorName).ToNot(BeEmpty())
		})
	})

	Describe("Server Info for Monitoring", func() {
		It("[TC-MON-005] should provide complete server information", func() {
			cfg := Config{
				Name:     "info-monitor-test",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: false,
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// All info methods should return valid data
			Expect(srv.GetName()).To(Equal("info-monitor-test"))
			Expect(srv.GetBindable()).To(Equal("127.0.0.1:8080"))
			Expect(srv.GetExpose()).To(ContainSubstring("localhost:8080"))
			Expect(srv.IsDisable()).To(BeFalse())
			Expect(srv.IsTLS()).To(BeFalse())
			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.MonitorName()).ToNot(BeEmpty())
		})

		It("[TC-MON-006] should reflect server state changes", func() {
			cfg := Config{
				Name:   "state-monitor-test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Initial state
			Expect(srv.IsRunning()).To(BeFalse())

			// Update config to disabled
			newCfg := Config{
				Name:     "state-monitor-test",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}
			newCfg.RegisterHandlerFunc(defaultHandler)

			err = srv.SetConfig(newCfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// State should reflect change
			Expect(srv.IsDisable()).To(BeTrue())
		})
	})
})
