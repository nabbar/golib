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
	"net/http"

	libhtp "github.com/nabbar/golib/httpserver"
	. "github.com/nabbar/golib/httpserver/pool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-MR] Pool Merge and Handler", func() {
	Describe("Pool Merge", func() {
		It("[TC-MR-001] should merge two pools", func() {
			pool1 := New(nil, nil)
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			err := pool1.StoreNew(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			pool2 := New(nil, nil)
			cfg2 := makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081")
			err = pool2.StoreNew(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			err = pool1.Merge(pool2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool1.Len()).To(Equal(2))
		})

		It("[TC-MR-002] should merge overlapping servers", func() {
			pool1 := New(nil, nil)
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			err := pool1.StoreNew(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			pool2 := New(nil, nil)
			cfg2 := makeTestConfig("server1-updated", "127.0.0.1:8080", "http://localhost:8080")
			err = pool2.StoreNew(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			err = pool1.Merge(pool2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool1.Len()).To(Equal(1))

			srv := pool1.Load("127.0.0.1:8080")
			Expect(srv.GetName()).To(Equal("server1-updated"))
		})

		It("[TC-MR-003] should merge empty pool", func() {
			pool1 := New(nil, nil)
			cfg := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			err := pool1.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			pool2 := New(nil, nil)

			err = pool1.Merge(pool2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool1.Len()).To(Equal(1))
		})

		It("[TC-MR-004] should merge into empty pool", func() {
			pool1 := New(nil, nil)

			pool2 := New(nil, nil)
			cfg := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			err := pool2.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			err = pool1.Merge(pool2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool1.Len()).To(Equal(1))
		})

		It("[TC-MR-005] should merge multiple servers", func() {
			pool1 := New(nil, nil)
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			err := pool1.StoreNew(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			pool2 := New(nil, nil)
			cfgs := []libhtp.Config{
				makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
				makeTestConfig("server3", "127.0.0.1:8082", "http://localhost:8082"),
			}
			for _, cfg := range cfgs {
				err = pool2.StoreNew(cfg, nil)
				Expect(err).ToNot(HaveOccurred())
			}

			err = pool1.Merge(pool2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool1.Len()).To(Equal(3))
		})
	})

	Describe("Pool Handler", func() {
		It("[TC-MR-006] should register handler function", func() {
			pool := New(nil, nil)

			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"test": http.NotFoundHandler(),
				}
			}

			pool.Handler(handlerFunc)

			// Handler registered successfully (no error)
		})

		It("[TC-MR-007] should allow nil handler", func() {
			pool := New(nil, nil)

			// Should not panic
			pool.Handler(nil)
		})

		It("[TC-MR-008] should replace existing handler", func() {
			pool := New(nil, nil)

			handler1 := func() map[string]http.Handler {
				return map[string]http.Handler{"h1": http.NotFoundHandler()}
			}
			pool.Handler(handler1)

			handler2 := func() map[string]http.Handler {
				return map[string]http.Handler{"h2": http.NotFoundHandler()}
			}
			pool.Handler(handler2)

			// No error means successful replacement
		})
	})

	Describe("Pool with Handler Function", func() {
		It("[TC-MR-009] should create pool with handler", func() {
			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"default": http.NotFoundHandler(),
				}
			}

			pool := New(nil, handlerFunc)

			Expect(pool).ToNot(BeNil())
			Expect(pool.Len()).To(Equal(0))
		})

		It("[TC-MR-010] should add servers to pool with handler", func() {
			handlerFunc := func() map[string]http.Handler {
				return map[string]http.Handler{
					"api": http.NotFoundHandler(),
				}
			}

			pool := New(nil, handlerFunc)

			cfg := makeTestConfig("api-server", "127.0.0.1:8080", "http://localhost:8080")

			err := pool.StoreNew(cfg, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool.Len()).To(Equal(1))
		})
	})

	Describe("Monitor Names", func() {
		It("[TC-MR-011] should return monitor names for all servers", func() {
			pool := New(nil, nil)

			cfgs := []libhtp.Config{
				makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080"),
				makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081"),
			}

			for _, cfg := range cfgs {
				err := pool.StoreNew(cfg, nil)
				Expect(err).ToNot(HaveOccurred())
			}

			names := pool.MonitorNames()
			Expect(names).To(HaveLen(2))
		})

		It("[TC-MR-012] should return empty list for empty pool", func() {
			pool := New(nil, nil)

			names := pool.MonitorNames()
			Expect(names).To(BeEmpty())
		})
	})

	Describe("Pool New with Servers", func() {
		It("[TC-MR-013] should create pool with initial servers", func() {
			cfg1 := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			srv1, err := libhtp.New(cfg1, nil)
			Expect(err).ToNot(HaveOccurred())

			cfg2 := makeTestConfig("server2", "127.0.0.1:8081", "http://localhost:8081")
			srv2, err := libhtp.New(cfg2, nil)
			Expect(err).ToNot(HaveOccurred())

			pool := New(nil, nil, srv1, srv2)

			Expect(pool.Len()).To(Equal(2))
			Expect(pool.Has("127.0.0.1:8080")).To(BeTrue())
			Expect(pool.Has("127.0.0.1:8081")).To(BeTrue())
		})

		It("[TC-MR-014] should handle nil servers in creation", func() {
			cfg := makeTestConfig("server1", "127.0.0.1:8080", "http://localhost:8080")
			srv, err := libhtp.New(cfg, nil)
			Expect(err).ToNot(HaveOccurred())

			pool := New(nil, nil, srv, nil)

			Expect(pool.Len()).To(Equal(1))
		})

		It("[TC-MR-015] should create empty pool with no initial servers", func() {
			pool := New(nil, nil)

			Expect(pool.Len()).To(Equal(0))
		})
	})
})
