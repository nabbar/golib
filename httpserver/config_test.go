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
	. "github.com/nabbar/golib/httpserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-CF] Config", func() {
	Describe("Config Validation", func() {
		It("[TC-CF-001] should fail validation without name", func() {
			cfg := Config{
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-002] should fail validation without listen address", func() {
			cfg := Config{
				Name:   "test-server",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-003] should fail validation without expose URL", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-004] should validate valid config", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-005] should fail validation with invalid listen format", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "invalid format",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("[TC-CF-006] should fail validation with invalid expose URL", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
				Expose: "not a url",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Config Fields", func() {
		It("[TC-CF-007] should set server name", func() {
			cfg := Config{
				Name:   "my-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080",
			}

			Expect(cfg.Name).To(Equal("my-server"))
		})

		It("[TC-CF-008] should set listen address with port", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "192.168.1.100:9000",
				Expose: "http://localhost:9000",
			}

			Expect(cfg.Listen).To(Equal("192.168.1.100:9000"))
		})

		It("[TC-CF-009] should set expose URL", func() {
			cfg := Config{
				Name:   "test-server",
				Listen: "127.0.0.1:8080",
				Expose: "https://api.example.com",
			}

			Expect(cfg.Expose).To(Equal("https://api.example.com"))
		})

		It("[TC-CF-010] should set handler key", func() {
			cfg := Config{
				Name:       "test-server",
				Listen:     "127.0.0.1:8080",
				Expose:     "http://localhost:8080",
				HandlerKey: "api-v1",
			}

			Expect(cfg.HandlerKey).To(Equal("api-v1"))
		})

		It("[TC-CF-011] should set disabled flag", func() {
			cfg := Config{
				Name:     "test-server",
				Listen:   "127.0.0.1:8080",
				Expose:   "http://localhost:8080",
				Disabled: true,
			}

			Expect(cfg.Disabled).To(BeTrue())
		})

		It("[TC-CF-012] should set TLS mandatory flag", func() {
			cfg := Config{
				Name:         "test-server",
				Listen:       "127.0.0.1:8443",
				Expose:       "https://localhost:8443",
				TLSMandatory: true,
			}

			Expect(cfg.TLSMandatory).To(BeTrue())
		})
	})

	Describe("Config with Different Listen Formats", func() {
		It("[TC-CF-013] should accept IPv4 address", func() {
			cfg := Config{
				Name:   "ipv4-server",
				Listen: "192.168.1.1:8080",
				Expose: "http://192.168.1.1:8080",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-014] should accept localhost", func() {
			cfg := Config{
				Name:   "localhost-server",
				Listen: "localhost:8080",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-015] should accept all interfaces binding", func() {
			cfg := Config{
				Name:   "all-interfaces",
				Listen: "0.0.0.0:8080",
				Expose: "http://localhost:8080",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Config with Different Expose URLs", func() {
		It("[TC-CF-016] should accept HTTP URL", func() {
			cfg := Config{
				Name:   "http-server",
				Listen: "127.0.0.1:8080",
				Expose: "http://api.example.com",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-017] should accept HTTPS URL", func() {
			cfg := Config{
				Name:   "https-server",
				Listen: "127.0.0.1:8443",
				Expose: "https://secure.example.com",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-018] should accept URL with port", func() {
			cfg := Config{
				Name:   "custom-port",
				Listen: "127.0.0.1:9000",
				Expose: "http://localhost:9000",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-CF-019] should accept URL with path", func() {
			cfg := Config{
				Name:   "with-path",
				Listen: "127.0.0.1:8080",
				Expose: "http://localhost:8080/api/v1",
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
