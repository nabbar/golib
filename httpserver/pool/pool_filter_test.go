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

package pool_test

import (
	libhtp "github.com/nabbar/golib/httpserver"
	. "github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-FL] Pool Filtering", func() {
	var pool Pool

	BeforeEach(func() {
		pool = New(nil, nil)

		// Create test servers with different attributes
		cfgs := []libhtp.Config{
			makeTestConfig("api-server", "127.0.0.1:8080", "http://localhost:8080"),
			makeTestConfig("web-server", "127.0.0.1:8081", "http://localhost:8081"),
			makeTestConfig("admin-server", "192.168.1.1:8080", "http://admin.example.com:8080"),
			makeTestConfig("api-v2-server", "127.0.0.1:9000", "http://api.example.com:9000"),
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
		It("[TC-FL-001] should filter by exact name", func() {
			filtered := pool.Filter(srvtps.FieldName, "api-server", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))

			srv := filtered.Load("127.0.0.1:8080")
			Expect(srv).ToNot(BeNil())
			Expect(srv.GetName()).To(Equal("api-server"))
		})

		It("[TC-FL-002] should filter by name regex", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "^api-.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})

		It("[TC-FL-003] should return empty pool for no match", func() {
			filtered := pool.Filter(srvtps.FieldName, "non-existent", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})
	})

	Describe("Filter by Bind Address", func() {
		It("[TC-FL-004] should filter by exact bind address", func() {
			filtered := pool.Filter(srvtps.FieldBind, "127.0.0.1:8080", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})

		It("[TC-FL-005] should filter by bind address regex", func() {
			filtered := pool.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(3))
		})

		It("[TC-FL-006] should filter by specific network interface", func() {
			filtered := pool.Filter(srvtps.FieldBind, "", "^192\\.168\\..*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})
	})

	Describe("Filter by Expose Address", func() {
		It("[TC-FL-007] should filter by exact expose address", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "localhost:8080", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(1))
		})

		It("[TC-FL-008] should filter by expose regex", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "", ".*example\\.com.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})

		It("[TC-FL-009] should filter localhost servers", func() {
			filtered := pool.Filter(srvtps.FieldExpose, "", "localhost.*")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(2))
		})
	})

	Describe("List Operations", func() {
		It("[TC-FL-010] should list all server names", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "", ".*")

			Expect(names).To(HaveLen(4))
			Expect(names).To(ContainElements("api-server", "web-server", "admin-server", "api-v2-server"))
		})

		It("[TC-FL-011] should list filtered server names", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "", "^api-.*")

			Expect(names).To(HaveLen(2))
			Expect(names).To(ContainElements("api-server", "api-v2-server"))
		})

		It("[TC-FL-012] should list bind addresses", func() {
			binds := pool.List(srvtps.FieldBind, srvtps.FieldBind, "", ".*")

			Expect(binds).To(HaveLen(4))
			Expect(binds).To(ContainElements("127.0.0.1:8080", "127.0.0.1:8081", "192.168.1.1:8080", "127.0.0.1:9000"))
		})

		It("[TC-FL-013] should list expose addresses", func() {
			exposes := pool.List(srvtps.FieldExpose, srvtps.FieldExpose, "", ".*")

			Expect(exposes).To(HaveLen(4))
		})

		It("[TC-FL-014] should list names for filtered bind addresses", func() {
			names := pool.List(srvtps.FieldBind, srvtps.FieldName, "", "^127\\.0\\.0\\.1:808.*")

			Expect(names).To(HaveLen(2))
			Expect(names).To(ContainElements("api-server", "web-server"))
		})
	})

	Describe("Filter Edge Cases", func() {
		It("[TC-FL-015] should handle empty pattern and regex", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})

		It("[TC-FL-016] should handle invalid regex gracefully", func() {
			filtered := pool.Filter(srvtps.FieldName, "", "[invalid(regex")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})

		It("[TC-FL-017] should filter on empty pool", func() {
			emptyPool := New(nil, nil)
			filtered := emptyPool.Filter(srvtps.FieldName, "test", "")

			Expect(filtered).ToNot(BeNil())
			Expect(filtered.Len()).To(Equal(0))
		})
	})

	Describe("List with Empty Results", func() {
		It("[TC-FL-018] should return empty list for no matches", func() {
			names := pool.List(srvtps.FieldName, srvtps.FieldName, "non-existent", "")

			Expect(names).To(BeEmpty())
		})

		It("[TC-FL-019] should return empty list for empty pool", func() {
			emptyPool := New(nil, nil)
			names := emptyPool.List(srvtps.FieldName, srvtps.FieldName, "", ".*")

			Expect(names).To(BeEmpty())
		})
	})

	Describe("Complex Filtering", func() {
		It("[TC-FL-020] should chain filters", func() {
			// First filter by bind address
			filtered1 := pool.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")
			Expect(filtered1.Len()).To(Equal(3))

			// Then filter result by name
			filtered2 := filtered1.Filter(srvtps.FieldName, "", "^api-.*")
			Expect(filtered2.Len()).To(Equal(2))
		})

		It("[TC-FL-021] should filter and list in combination", func() {
			// Filter by bind address, list names
			names := pool.List(srvtps.FieldBind, srvtps.FieldName, "127.0.0.1:8080", "")

			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal("api-server"))
		})
	})

	Describe("Case Sensitivity", func() {
		It("[TC-FL-022] should be case-insensitive for exact pattern match", func() {
			filtered := pool.Filter(srvtps.FieldName, "API-SERVER", "")

			Expect(filtered.Len()).To(Equal(1))
		})
	})
})
