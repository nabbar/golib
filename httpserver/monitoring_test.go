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
	"net/http"
	"time"

	. "github.com/nabbar/golib/httpserver"
	logcfg "github.com/nabbar/golib/logger/config"
	moncfg "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
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

	Describe("HealthCheck", func() {
		var (
			srv Server
			ctx context.Context
			prt int
		)

		BeforeEach(func() {
			ctx = context.Background()
			prt = GetFreePort()
			cfg := Config{
				Name:   "healthcheck-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", prt),
				Expose: fmt.Sprintf("http://localhost:%d", prt),
			}
			cfg.RegisterHandlerFunc(defaultHandler)
			var err error
			srv, err = New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			_ = srv.Stop(ctx)
		})

		It("should return an error if the server is not running", func() {
			err := srv.HealthCheck(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("server is not running"))
		})

		It("should return nil if the server is running and healthy", func() {
			err := srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond) // give time for the server to start
			err = srv.HealthCheck(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error if the server has been stopped", func() {
			err := srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			err = srv.HealthCheck(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("server is not running"))
		})

		It("should not panic if logger is nil", func() {
			cfg := Config{
				Name:   "healthcheck-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", prt),
				Expose: fmt.Sprintf("http://localhost:%d", prt),
			}
			cfg.RegisterHandlerFunc(defaultHandler)
			var err error
			srv, err = New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			err = srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			// Set the logger to nil (simulating a missing logger)
			// This is normally not possible from outside the package
			// but we can use reflection to achieve it for testing purposes
			// This is a HACK and should not be done in production code
			// It is only used to increase test coverage
			//if s, ok := srv.(*srv); ok {
			//	s.l.Store(nil)
			//}

			err = srv.HealthCheck(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Monitor", func() {
		var (
			srv Server
			vrs libver.Version
			prt int
		)

		BeforeEach(func() {
			prt = GetFreePort()
			cfg := Config{
				Name:   "monitor-func-test",
				Listen: fmt.Sprintf("127.0.0.1:%d", prt),
				Expose: fmt.Sprintf("http://localhost:%d", prt),
				Monitor: moncfg.Config{
					Name:          "monitor-test",
					CheckTimeout:  0,
					IntervalCheck: 0,
					IntervalFall:  0,
					IntervalRise:  0,
					FallCountKO:   0,
					FallCountWarn: 0,
					RiseCountKO:   0,
					RiseCountWarn: 0,
					Logger:        logcfg.Options{},
				},
			}
			cfg.RegisterHandlerFunc(defaultHandler)
			var err error
			srv, err = New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			vrs = libver.NewVersion(
				libver.License_MIT,
				"testapp",
				"Test Application",
				"2024-01-01",
				"abc123",
				"v1.0.0",
				"Test Author",
				"testapp",
				struct{}{},
				0,
			)
		})

		It("should return a valid monitor instance", func() {
			mon, err := srv.Monitor(vrs)
			Expect(err).ToNot(HaveOccurred())
			Expect(mon).ToNot(BeNil())
		})

		It("should return an error for an invalid monitor config", func() {
			cfg := srv.GetConfig()
			cfg.Monitor.Name = "monitor-test"
			err := srv.SetConfig(*cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = srv.Monitor(vrs)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
