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

// defaultHandler provides a minimal handler for tests
func defaultHandler() map[string]http.Handler {
	return map[string]http.Handler{
		"": http.NotFoundHandler(),
	}
}

var _ = Describe("Server Info", func() {
	Describe("Server Creation", func() {
		It("should create server from valid config", func() {
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

		It("should fail with invalid config", func() {
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

		It("should return correct server name", func() {
			name := srv.GetName()
			Expect(name).To(Equal("info-server"))
		})

		It("should return correct bind address", func() {
			bind := srv.GetBindable()
			Expect(bind).To(Equal("127.0.0.1:9000"))
		})

		It("should return correct expose address", func() {
			expose := srv.GetExpose()
			Expect(expose).To(Equal("localhost:9000"))
		})

		It("should not be disabled by default", func() {
			disabled := srv.IsDisable()
			Expect(disabled).To(BeFalse())
		})

		It("should not have TLS by default", func() {
			hasTLS := srv.IsTLS()
			Expect(hasTLS).To(BeFalse())
		})
	})

	Describe("Server Disabled Flag", func() {
		It("should respect disabled flag", func() {
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

		It("should not be disabled when flag is false", func() {
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
		It("should report TLS when TLSMandatory is true", func() {
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

		It("should not report TLS when TLSMandatory is false and no certificates", func() {
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
		It("should get server config", func() {
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

		It("should update server config", func() {
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

		It("should succeed updating compatible config", func() {
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
		It("should not be running initially", func() {
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
		It("should merge server configs", func() {
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
})
