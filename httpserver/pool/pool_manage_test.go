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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-MG] Pool Management", func() {
	Describe("Store and Load Operations", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)
		})

		AfterEach(func() {
			pool.Clean()
		})

		It("[TC-MG-001] should store and load server", func() {
			cfg := makeTestConfig("test-server", "127.0.0.1:8080", "http://localhost:8080")

			err := pool.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			srv := pool.Load("127.0.0.1:8080")
			Expect(srv).ToNot(BeNil())
			Expect(srv.GetName()).To(Equal("test-server"))
		})

		It("[TC-MG-002] should return nil for non-existent server", func() {
			srv := pool.Load("non-existent:9999")
			Expect(srv).To(BeNil())
		})

		It("[TC-MG-003] should store multiple servers", func() {
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			cfg2 := makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081")

			err := pool.StoreNew(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			err = pool.StoreNew(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(pool.Len()).To(Equal(2))
		})

		It("[TC-MG-004] should overwrite server with same bind address", func() {
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")

			err := pool.StoreNew(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			cfg2 := makeTestConfig("server2", "127.0.0.1:8080", "http://localhost:8080")

			err = pool.StoreNew(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			srv := pool.Load("127.0.0.1:8080")
			Expect(srv.GetName()).To(Equal("server2"))
			Expect(pool.Len()).To(Equal(1))
		})
	})

	Describe("Delete Operations", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)
		})

		AfterEach(func() {
			pool.Clean()
		})

		It("[TC-MG-005] should delete existing server", func() {
			cfg := makeTestConfig("delete-test", "127.0.0.1:8080", "http://localhost:8080")

			err := pool.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool.Len()).To(Equal(1))

			pool.Delete("127.0.0.1:8080")
			Expect(pool.Len()).To(Equal(0))
		})

		It("[TC-MG-006] should handle deleting non-existent server", func() {
			// Should not panic
			pool.Delete("non-existent:9999")
			Expect(pool.Len()).To(Equal(0))
		})

		It("[TC-MG-007] should load and delete server", func() {
			cfg := makeTestConfig("load-delete-test", "127.0.0.1:8080", "http://localhost:8080")

			err := pool.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			srv, loaded := pool.LoadAndDelete("127.0.0.1:8080")
			Expect(loaded).To(BeTrue())
			Expect(srv).ToNot(BeNil())
			Expect(srv.GetName()).To(Equal("load-delete-test"))
			Expect(pool.Len()).To(Equal(0))
		})

		It("[TC-MG-008] should return false for load and delete non-existent", func() {
			srv, loaded := pool.LoadAndDelete("non-existent:9999")
			Expect(loaded).To(BeFalse())
			Expect(srv).To(BeNil())
		})
	})

	Describe("Walk Operations", func() {
		var pol Pool

		BeforeEach(func() {
			pol = New(nil, nil)

			// Add test servers
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			cfg2 := makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081")
			cfg3 := makeTestConfig("server3", "127.0.0.1:8082", "http://localhost:8082")

			_ = pol.StoreNew(cfg1, nil)
			_ = pol.StoreNew(cfg2, nil)
			_ = pol.StoreNew(cfg3, nil)
		})

		AfterEach(func() {
			pol.Clean()
		})

		It("[TC-MG-009] should walk all servers", func() {
			var count int
			var names []string

			pol.Walk(func(bindAddress string, srv libhtp.Server) bool {
				count++
				names = append(names, srv.GetName())
				return true
			})

			Expect(count).To(Equal(3))
			Expect(names).To(ContainElements("server1", "server2", "server3"))
		})

		It("[TC-MG-010] should stop walking when callback returns false", func() {
			var count int

			pol.Walk(func(bindAddress string, srv libhtp.Server) bool {
				count++
				return count < 2
			})

			// Should stop after 2 iterations when callback returns false
			Expect(count).To(Equal(2))
		})

		It("[TC-MG-011] should walk with bind address filter", func() {
			var names []string

			pol.WalkLimit(func(bindAddress string, srv libhtp.Server) bool {
				names = append(names, srv.GetName())
				return true
			}, "127.0.0.1:8080", "127.0.0.1:8082")

			Expect(names).To(ConsistOf("server1", "server3"))
		})
	})

	Describe("Has Operation", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)

			cfg := makeTestConfig("test-server", "127.0.0.1:8080", "http://localhost:8080")
			_ = pool.StoreNew(cfg, nil)
		})

		AfterEach(func() {
			pool.Clean()
		})

		It("[TC-MG-012] should return true for existing server", func() {
			exists := pool.Has("127.0.0.1:8080")
			Expect(exists).To(BeTrue())
		})

		It("[TC-MG-013] should return false for non-existent server", func() {
			exists := pool.Has("127.0.0.1:9999")
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Clean Operation", func() {
		It("[TC-MG-014] should remove all servers", func() {
			pool := New(nil, nil)

			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			cfg2 := makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081")

			_ = pool.StoreNew(cfg1, nil)
			_ = pool.StoreNew(cfg2, nil)
			Expect(pool.Len()).To(Equal(2))

			pool.Clean()
			Expect(pool.Len()).To(Equal(0))
		})
	})

	Describe("StoreNew Error Handling", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)
		})

		AfterEach(func() {
			pool.Clean()
		})

		It("[TC-MG-015] should fail with invalid config", func() {
			cfg := libhtp.Config{
				Name: "invalid",
				// Missing Listen and Expose
			}

			err := pool.StoreNew(cfg, nil)
			Expect(err).To(HaveOccurred())
			Expect(pool.Len()).To(Equal(0))
		})
	})
})
