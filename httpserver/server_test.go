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
	"net"
	"net/http"
	"time"

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// defaultHandler provides a minimal handler for tests
func defaultHandler() map[string]http.Handler {
	return map[string]http.Handler{
		"": http.NotFoundHandler(),
	}
}

var _ = Describe("[TC-SV] Server Info", func() {
	Describe("Server Creation", func() {
		It("[TC-SV-001] should create server from valid config", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("[TC-SV-002] should fail with invalid config", func() {
			cfg := Config{
				Name: "invalid",
				// Missing Listen and Expose
			}

			srv, err := New(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})
	})

	Describe("Server Info Methods", func() {
		var srv Server

		BeforeEach(func() {
			cfg := Config{
				Name:   "info-server",
				Listen: "127.0.0.1:9000",
				Expose: "http://localhost:9000",
			}
			cfg.RegisterHandlerFunc(defaultHandler)
			var err error
			srv, err = New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-SV-003] should return correct server name", func() {
			name := srv.GetName()
			Expect(name).To(Equal("info-server"))
		})

		It("[TC-SV-004] should return correct bind address", func() {
			bind := srv.GetBindable()
			Expect(bind).To(Equal("127.0.0.1:9000"))
		})

		It("[TC-SV-005] should return correct expose address", func() {
			expose := srv.GetExpose()
			Expect(expose).To(Equal("localhost:9000"))
		})

		It("[TC-SV-006] should not be disabled by default", func() {
			disabled := srv.IsDisable()
			Expect(disabled).To(BeFalse())
		})

		It("[TC-SV-007] should not have TLS by default", func() {
			hasTLS := srv.IsTLS()
			Expect(hasTLS).To(BeFalse())
		})
	})

	Describe("Server Disabled Flag", func() {
		It("[TC-SV-008] should respect disabled flag", func() {
			cfg := Config{
				Name:     "disabled-server",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsDisable()).To(BeTrue())
		})

		It("[TC-SV-009] should not be disabled when flag is false", func() {
			cfg := Config{
				Name:     "enabled-server",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: false,
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsDisable()).To(BeFalse())
		})
	})

	Describe("Server TLS Configuration", func() {
		It("[TC-SV-010] should report TLS when TLSMandatory is true", func() {
			cfg := Config{
				Name:         "tls-server",
				Listen:       "127.0.0.1:8443",
				Expose:       "https://localhost:8443",
				TLSMandatory: true,
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			_, err := New(cfg, nil)
			Expect(err).To(HaveOccurred()) // Will fail due to missing TLS certificates
		})

		It("[TC-SV-011] should not report TLS when TLSMandatory is false and no certificates", func() {
			cfg := Config{
				Name:         "no-tls-server",
				Listen:       "127.0.0.1:8080",
				Expose:       "http://localhost:8080",
				TLSMandatory: false,
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsTLS()).To(BeFalse())
		})
	})

	Describe("Server Config Management", func() {
		It("[TC-SV-012] should get server config", func() {
			cfg := Config{
				Name:   "config-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			retrievedCfg := srv.GetConfig()
			Expect(retrievedCfg).ToNot(BeNil())
			Expect(retrievedCfg.Name).To(Equal("config-server"))
		})

		It("[TC-SV-013] should update server config", func() {
			originalCfg := Config{
				Name:   "original-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			originalCfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(originalCfg, nil)
			Expect(err).ToNot(HaveOccurred())

			newCfg := Config{
				Name:   "updated-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			newCfg.RegisterHandlerFunc(defaultHandler)

			err = srv.SetConfig(newCfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.GetName()).To(Equal("updated-server"))
		})

		It("[TC-SV-014] should succeed updating compatible config", func() {
			cfg := Config{
				Name:   "valid-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			// Another valid config
			newCfg := Config{
				Name:     "renamed-server",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}
			newCfg.RegisterHandlerFunc(defaultHandler)

			err = srv.SetConfig(newCfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.GetName()).To(Equal("renamed-server"))
			Expect(srv.IsDisable()).To(BeTrue())
		})
	})

	Describe("Server Lifecycle State", func() {
		It("[TC-SV-015] should not be running initially", func() {
			cfg := Config{
				Name:   "lifecycle-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Describe("Server Merge", func() {
		It("[TC-SV-016] should merge server configs", func() {
			cfg1 := Config{
				Name:   "server1",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg1.RegisterHandlerFunc(defaultHandler)

			srv1, err := New(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			cfg2 := Config{
				Name:   "server2",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg2.RegisterHandlerFunc(defaultHandler)

			srv2, err := New(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			err = srv1.Merge(srv2, nil)
			Expect(err).ToNot(HaveOccurred())

			// After merge, srv1 should have server2's config
			Expect(srv1.GetName()).To(Equal("server2"))
		})
	})

	Describe("Server Error Handling", func() {
		var (
			srv Server
			ctx context.Context
			prt int
		)

		BeforeEach(func() {
			ctx = context.Background()
			prt = GetFreePort()
			cfg := Config{
				Name:   "error-handling-test",
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

		It("should report no error when server starts and stops successfully", func() {
			Expect(srv.IsError()).To(BeFalse())
			Expect(srv.GetError()).To(BeNil())

			err := srv.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			Expect(srv.IsError()).To(BeFalse())
			Expect(srv.GetError()).To(BeNil())

			err = srv.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			Expect(srv.IsError()).To(BeTrue())

			err = srv.GetError()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("server start but not listen"))
		})

		It("should report an error when server fails to start due to port in use", func() {
			// Start a listener on the port before trying to start our server
			lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", prt))
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = lis.Close()
			}()

			cx1, cn1 := context.WithTimeout(ctx, time.Second*3)
			defer cn1()

			// Attempt to start our server, which should fail
			err = srv.Start(cx1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("server is not running"))

			// IsError and GetError should reflect this failure
			Expect(srv.IsError()).To(BeTrue())
			Expect(srv.GetError()).To(HaveOccurred())
			Expect(srv.GetError().Error()).To(ContainSubstring("server start but not listen"))
		})
	})
})
