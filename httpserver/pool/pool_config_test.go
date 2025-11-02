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

package pool_test

import (
	"context"
	"net/http"

	libhtp "github.com/nabbar/golib/httpserver"
	. "github.com/nabbar/golib/httpserver/pool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// configDefaultHandler provides a minimal handler for tests
func configDefaultHandler() map[string]http.Handler {
	return map[string]http.Handler{
		"": http.NotFoundHandler(),
	}
}

// makeConfigConfig creates a config with handler for testing
func makeConfigConfig(name, listen, expose string) libhtp.Config {
	cfg := libhtp.Config{
		Name:   name,
		Listen: listen,
		Expose: expose,
	}
	cfg.RegisterHandlerFunc(configDefaultHandler)
	return cfg
}

var _ = Describe("Pool Config", func() {
	Describe("Config Validation", func() {
		It("should validate all valid configs", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation with invalid config", func() {
			cfg := Config{
				makeConfigConfig("valid-server", "127.0.0.1:8080", "http://localhost:8080"),
				{
					Name: "invalid-server",
					// Missing Listen and Expose
				},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should validate empty config", func() {
			cfg := Config{}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Config Pool Creation", func() {
		It("should create pool from valid configs", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			pool, err := cfg.Pool(nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool).ToNot(BeNil())
			Expect(pool.Len()).To(Equal(2))
		})

		It("should fail to create pool with invalid configs", func() {
			cfg := Config{
				{
					Name: "invalid",
					// Missing required fields
				},
			}

			pool, err := cfg.Pool(nil, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(pool).ToNot(BeNil())
			Expect(pool.Len()).To(Equal(0))
		})

		It("should create empty pool from empty config", func() {
			cfg := Config{}

			pool, err := cfg.Pool(nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool).ToNot(BeNil())
			Expect(pool.Len()).To(Equal(0))
		})
	})

	Describe("Config Walk", func() {
		It("should walk all configs", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			var count int
			var names []string

			cfg.Walk(func(c libhtp.Config) bool {
				count++
				names = append(names, c.Name)
				return true
			})

			Expect(count).To(Equal(2))
			Expect(names).To(ContainElements("server1", "server2"))
		})

		It("should stop walking when callback returns false", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
				makeConfigConfig("server3", "127.0.0.1:8082", "http://localhost:8082"),
			}

			var count int

			cfg.Walk(func(c libhtp.Config) bool {
				count++
				return count < 2
			})

			Expect(count).To(Equal(2))
		})

		It("should handle nil walk function", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
			}

			// Should not panic
			cfg.Walk(nil)
		})

		It("should walk empty config", func() {
			cfg := Config{}

			var count int
			cfg.Walk(func(c libhtp.Config) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})
	})

	Describe("Config SetHandlerFunc", func() {
		It("should set handler function for all configs", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"default": http.NotFoundHandler(),
				}
			}

			cfg.SetHandlerFunc(handlerFunc)

			// Verify all configs can still be validated
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil handler function", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
			}

			// Should not panic
			cfg.SetHandlerFunc(nil)
		})

		It("should work on empty config", func() {
			cfg := Config{}

			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{}
			}

			// Should not panic
			cfg.SetHandlerFunc(handlerFunc)
		})
	})

	Describe("Config SetContext", func() {
		It("should set context function for all configs", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeConfigConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			cfg.SetContext(context.Background())

			// Verify all configs can still be validated
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil context function", func() {
			cfg := Config{
				makeConfigConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
			}

			// Should not panic
			cfg.SetContext(nil)
		})
	})

	Describe("Config with Multiple Operations", func() {
		It("should handle all config operations in sequence", func() {
			// Create configs without handler first
			cfg := Config{
				{
					Name:   "server1",
					Listen: "127.0.0.1:8080",
					Expose: "http://localhost:8080",
				},
				{
					Name:   "server2",
					Listen: "127.0.0.1:8081",
					Expose: "http://localhost:8081",
				},
			}

			// Set handler
			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"": http.NotFoundHandler(),
				}
			}
			cfg.SetHandlerFunc(handlerFunc)

			// Set context
			cfg.SetContext(context.Background())

			// Validate
			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())

			// Create pool
			pool, err := cfg.Pool(nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool.Len()).To(Equal(2))
		})
	})

	Describe("Config Partial Validation", func() {
		It("should report all validation errors", func() {
			cfg := Config{
				{
					Name: "invalid1",
					// Missing required fields
				},
				{
					Name: "invalid2",
					// Missing required fields
				},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should create pool with valid configs only", func() {
			cfg := Config{
				makeConfigConfig("valid", "127.0.0.1:8080", "http://localhost:8080"),
				{
					Name: "invalid",
					// Missing required fields
				},
			}

			pool, err := cfg.Pool(nil, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(pool).ToNot(BeNil())
			// Only one valid config should be added
			Expect(pool.Len()).To(Equal(1))
		})
	})
})
