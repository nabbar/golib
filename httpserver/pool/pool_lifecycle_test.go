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
	"context"
	"time"

	. "github.com/nabbar/golib/httpserver/pool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[TC-LC] Pool Lifecycle", func() {
	Describe("IsRunning", func() {
		It("[TC-LC-001] should return false for empty pool", func() {
			pool := New(nil, nil)
			Expect(pool.IsRunning()).To(BeFalse())
		})

		It("[TC-LC-002] should return false when no servers are running", func() {
			pool := New(nil, nil)
			cfg := makeTestConfig("test", "127.0.0.1:18080", "http://localhost:18080")
			pool.StoreNew(cfg, nil)
			Expect(pool.IsRunning()).To(BeFalse())
		})
	})

	Describe("Uptime", func() {
		It("[TC-LC-003] should return zero duration for empty pool", func() {
			pool := New(nil, nil)
			uptime := pool.Uptime()
			Expect(uptime).To(Equal(time.Duration(0)))
		})

		It("[TC-LC-004] should return zero duration when servers not started", func() {
			pool := New(nil, nil)
			cfg := makeTestConfig("test", "127.0.0.1:18081", "http://localhost:18081")
			pool.StoreNew(cfg, nil)
			uptime := pool.Uptime()
			Expect(uptime).To(Equal(time.Duration(0)))
		})
	})

	Describe("MonitorNames", func() {
		It("[TC-LC-005] should return empty slice for empty pool", func() {
			pool := New(nil, nil)
			names := pool.MonitorNames()
			Expect(names).To(BeEmpty())
		})

		It("[TC-LC-006] should return monitor names for all servers", func() {
			pool := New(nil, nil)
			cfg1 := makeTestConfig("server1", "127.0.0.1:18082", "http://localhost:18082")
			cfg2 := makeTestConfig("server2", "127.0.0.1:18083", "http://localhost:18083")
			pool.StoreNew(cfg1, nil)
			pool.StoreNew(cfg2, nil)

			names := pool.MonitorNames()
			Expect(names).To(HaveLen(2))
		})
	})

	Describe("Start/Stop/Restart", func() {
		It("[TC-LC-007] should handle Start on empty pool", func() {
			pool := New(nil, nil)
			ctx := context.Background()
			err := pool.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-LC-008] should handle Stop on empty pool", func() {
			pool := New(nil, nil)
			ctx := context.Background()
			err := pool.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("[TC-LC-009] should handle Restart on empty pool", func() {
			pool := New(nil, nil)
			ctx := context.Background()
			err := pool.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Context Integration", func() {
		It("[TC-LC-010] should create pool with context", func() {
			ctx := context.Background()
			pool := New(ctx, nil)
			Expect(pool).ToNot(BeNil())
		})

		It("[TC-LC-011] should clone pool with new context", func() {
			pool := New(nil, nil)
			cfg := makeTestConfig("test", "127.0.0.1:18084", "http://localhost:18084")
			pool.StoreNew(cfg, nil)

			newCtx := context.Background()
			cloned := pool.Clone(newCtx)
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.Len()).To(Equal(1))
		})

		It("[TC-LC-012] should clone pool with nil context", func() {
			pool := New(nil, nil)
			cloned := pool.Clone(nil)
			Expect(cloned).ToNot(BeNil())
		})
	})

	Describe("Config Operations", func() {
		It("[TC-LC-013] should set context on configs", func() {
			configs := Config{
				makeTestConfig("s1", "127.0.0.1:18085", "http://localhost:18085"),
			}
			ctx := context.Background()
			configs.SetContext(ctx)
			Expect(configs).To(HaveLen(1))
		})

		It("[TC-LC-014] should handle nil context", func() {
			configs := Config{
				makeTestConfig("s1", "127.0.0.1:18086", "http://localhost:18086"),
			}
			configs.SetContext(nil)
			Expect(configs).To(HaveLen(1))
		})

		It("[TC-LC-015] should set default TLS on configs", func() {
			configs := Config{
				makeTestConfig("s1", "127.0.0.1:18087", "http://localhost:18087"),
			}
			configs.SetDefaultTLS(nil)
			Expect(configs).To(HaveLen(1))
		})
	})
})
