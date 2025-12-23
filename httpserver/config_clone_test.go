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
	"net/http"

	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-CF] Config Helper Methods", func() {
	Describe("Config Clone", func() {
		It("[TC-CF-020] should clone config successfully", func() {
			original := Config{
				Name:       "original",
				Listen:     "127.0.0.1:8080",
				Expose:     "http://localhost:8080",
				HandlerKey: "api",
				Disabled:   false,
			}

			cloned := original.Clone()

			Expect(cloned.Name).To(Equal(original.Name))
			Expect(cloned.Listen).To(Equal(original.Listen))
			Expect(cloned.Expose).To(Equal(original.Expose))
			Expect(cloned.HandlerKey).To(Equal(original.HandlerKey))
			Expect(cloned.Disabled).To(Equal(original.Disabled))
		})

		It("[TC-CF-021] should create independent clone", func() {
			original := Config{
				Name:   "original",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			cloned := original.Clone()
			cloned.Name = "modified"

			// Original should remain unchanged
			Expect(original.Name).To(Equal("original"))
			Expect(cloned.Name).To(Equal("modified"))
		})

		It("[TC-CF-022] should clone disabled flag", func() {
			original := Config{
				Name:     "original",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}

			cloned := original.Clone()

			Expect(cloned.Disabled).To(BeTrue())
		})

		It("[TC-CF-023] should clone TLS mandatory flag", func() {
			original := Config{
				Name:         "original",
				Listen:       "127.0.0.1:8443",
				Expose:       "https://localhost:8443",
				TLSMandatory: true,
			}

			cloned := original.Clone()

			Expect(cloned.TLSMandatory).To(BeTrue())
		})
	})

	Describe("Config RegisterHandlerFunc", func() {
		It("[TC-CF-024] should register handler function", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"test": http.NotFoundHandler(),
				}
			}

			cfg.RegisterHandlerFunc(handlerFunc)

			// Config should accept handler registration
			Expect(cfg.Name).To(Equal("test"))
		})

		It("[TC-CF-025] should allow nil handler function", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			// Should not panic
			cfg.RegisterHandlerFunc(nil)
		})
	})

	Describe("Config SetContext", func() {
		It("[TC-CF-026] should set context function", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			cfg.SetContext(context.Background())

			// Config should accept context function
			Expect(cfg.Name).To(Equal("test"))
		})

		It("[TC-CF-027] should allow nil context function", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			// Should not panic
			cfg.SetContext(nil)
		})
	})

	Describe("Config Validation Edge Cases", func() {
		It("[TC-CF-028] should validate with all optional fields", func() {
			cfg := Config{
				Name:         "complete-config",
				Listen:       "127.0.0.1:8080",
				Expose:       "http://localhost:8080",
				HandlerKey:   "api-v1",
				Disabled:     false,
				TLSMandatory: false,
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-029] should validate disabled server", func() {
			cfg := Config{
				Name:     "disabled",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-030] should fail with empty name", func() {
			cfg := Config{
				Name:   "",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-031] should fail with empty listen", func() {
			cfg := Config{
				Name:   "test",
				Listen: "",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-032] should fail with empty expose", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:8080",
				Expose: "",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-033] should fail with invalid port in listen", func() {
			cfg := Config{
				Name:   "test",
				Listen: "127.0.0.1:99999",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-034] should validate numeric ports", func() {
			cfg := Config{
				Name:   "port-server",
				Listen: "127.0.0.1:65535",
				Expose: "http://localhost:65535",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Config Server Creation", func() {
		It("[TC-CF-035] should create server from config", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}
			cfg.RegisterHandlerFunc(defaultHandler)

			srv, err := cfg.Server(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
			Expect(srv.GetName()).To(Equal("test-server"))
		})

		It("[TC-CF-036] should fail to create server from invalid config", func() {
			cfg := Config{
				Name: "invalid",
			}

			srv, err := cfg.Server(nil)
			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})
	})
})
