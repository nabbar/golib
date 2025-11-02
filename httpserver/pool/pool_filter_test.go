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
	"net/http"

	libhtp "github.com/nabbar/golib/httpserver"
	. "github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// filterDefaultHandler provides a minimal handler for tests
func filterDefaultHandler() map[string]http.Handler {
	return map[string]http.Handler{
		"": http.NotFoundHandler(),
	}
}

// makeFilterConfig creates a config with handler for testing
func makeFilterConfig(name, listen, expose string) libhtp.Config {
	cfg := libhtp.Config{
		Name:   name,
		Listen: listen,
		Expose: expose,
	}
	cfg.RegisterHandlerFunc(filterDefaultHandler)
	return cfg
}

var _ = Describe("Pool Filtering", func() {
	var pool Pool

	BeforeEach(func() {
		pool = New(nil, nil)

		// Create test servers with different attributes
		cfgs := []libhtp.Config{
			makeFilterConfig("api-server", "127.0.0.1:8080", "http://localhost:8080"),
			makeFilterConfig("web-server", "127.0.0.1:8081", "http://localhost:8081"),
			makeFilterConfig("admin-server", "192.168.1.1:8080", "http://admin.example.com:8080"),
			makeFilterConfig("api-v2-server", "127.0.0.1:9000", "http://api.example.com:9000"),
		}

		for _, cfg := range cfgs {
			err := pool.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		pool.Clean()
	})

	Describe("Filter by Name", func() {
		It("should filter by exact name", func() {
			filtered := pool.Filter(srvtps.FieldName, "api-server", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))

			srv := filtered.Load("127.0.0.1:8080")
			Expect(srv).ToNot(BeNil())
			Expect(srv.GetName()).To(Equal("api-server"))
		})

		It("should filter by name regex", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "^api-.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})

		It("should return empty pool for no match", func() {
			filtered := pool.Filter(srvtps.FieldName, "non-existent", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})
	})

	Describe("Filter by Bind Address", func() {
		It("should filter by exact bind address", func() {
			filtered := pool.Filter(srvtps.FieldBind, "127.0.0.1:8080", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})

		It("should filter by bind address regex", func() {
			filtered := pool.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(3))
		})

		It("should filter by specific network interface", func() {
			filtered := pool.Filter(srvtps.FieldBind, "", "^192\\.168\\..*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})
	})

	Describe("Filter by Expose Address", func() {
		It("should filter by exact expose address", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "localhost:8080", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})

		It("should filter by expose regex", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "", ".*example\\.com.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})

		It("should filter localhost servers", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "", "localhost.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})
	})

	Describe("List Operations", func() {
		It("should list all server names", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "", ".*")

			Expect(names).To(HaveLen(4))
			Expect(names).To(ContainElements("api-server", "web-server", "admin-server", "api-v2-server"))
		})

		It("should list filtered server names", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "", "^api-.*")

			Expect(names).To(HaveLen(2))
			Expect(names).To(ContainElements("api-server", "api-v2-server"))
		})

		It("should list bind addresses", func() {
			binds := pool.List(srvtps.FieldBind, srvtps.FieldBind, "", ".*")

			Expect(binds).To(HaveLen(4))
			Expect(binds).To(ContainElements("127.0.0.1:8080", "127.0.0.1:8081", "192.168.1.1:8080", "127.0.0.1:9000"))
		})

		It("should list expose addresses", func() {
			exposes := pool.List(srvtps.FieldExpose, srvtps.FieldExpose, "", ".*")

			Expect(exposes).To(HaveLen(4))
		})

		It("should list names for filtered bind addresses", func() {
			names := pool.List(srvtps.FieldBind, srvtps.FieldName, "", "^127\\.0\\.0\\.1:808.*")

			Expect(names).To(HaveLen(2))
			Expect(names).To(ContainElements("api-server", "web-server"))
		})
	})

	Describe("Filter Edge Cases", func() {
		It("should handle empty pattern and regex", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})

		It("should handle invalid regex gracefully", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "[invalid(regex")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})

		It("should filter on empty pool", func() {
			emptyPool := New(nil, nil)
			filtered := emptyPool.Filter(srvtps.FieldName, "test", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})
	})

	Describe("List with Empty Results", func() {
		It("should return empty list for no matches", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "non-existent", "")

			Expect(names).To(BeEmpty())
		})

		It("should return empty list for empty pool", func() {
			emptyPool := New(nil, nil)
			names := emptyPool.List(srvtps.FieldName, srvtps.FieldName, "", ".*")

			Expect(names).To(BeEmpty())
		})
	})

	Describe("Complex Filtering", func() {
		It("should chain filters", func() {
			// First filter by bind address
			filtered1 := pool.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")
			Expect(filtered1.Len()).To(Equal(3))

			// Then filter result by name
			filtered2 := filtered1.Filter(srvtps.FieldName, "", "^api-.*")
			Expect(filtered2.Len()).To(Equal(2))
		})

		It("should filter and list in combination", func() {
			// Filter by bind address, list names
			names := pool.List(srvtps.FieldBind, srvtps.FieldName, "127.0.0.1:8080", "")

			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal("api-server"))
		})
	})

	Describe("Case Sensitivity", func() {
		It("should be case-insensitive for exact pattern match", func() {
			filtered := pool.Filter(srvtps.FieldName, "API-SERVER", "")

			Expect(filtered.Len()).To(Equal(1))
		})
	})
})
