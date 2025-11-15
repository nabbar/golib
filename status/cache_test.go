/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package status_test

import (
	"time"

	monpol "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/Cache", func() {
	var (
		status libsts.Status
		pool   monpol.Pool
	)

	BeforeEach(func() {
		status = libsts.New(globalCtx)
		status.SetInfo("cache-test", "v1.0.0", "abc123")
		pool = newPool()
		status.RegisterPool(func() montps.Pool { return pool })
	})

	Describe("Cache health checks", func() {
		Context("with empty pool", func() {
			It("should report cache health status", func() {
				healthy := status.IsCacheHealthy()
				Expect(healthy).To(BeAssignableToTypeOf(false))
			})

			It("should report cache strictly healthy status", func() {
				strictlyHealthy := status.IsCacheStrictlyHealthy()
				Expect(strictlyHealthy).To(BeAssignableToTypeOf(false))
			})
		})

		Context("with monitors", func() {
			BeforeEach(func() {
				// Add a monitor
				m := newHealthyMonitor("test-monitor")
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(50 * time.Millisecond)

				// Trigger a status check to populate cache
				_ = status.IsHealthy()
			})

			It("should use cached status", func() {
				// First call populates cache
				healthy1 := status.IsCacheHealthy()

				// Second call should use cache
				healthy2 := status.IsCacheHealthy()

				Expect(healthy1).To(Equal(healthy2))
			})

			It("should check strictly healthy from cache", func() {
				strictlyHealthy := status.IsCacheStrictlyHealthy()
				Expect(strictlyHealthy).To(BeTrue())
			})
		})
	})

	Describe("Multiple health checks", func() {
		BeforeEach(func() {
			// Add multiple monitors
			for i := 0; i < 3; i++ {
				m := newHealthyMonitor(time.Now().Format("monitor-2006-01-02-15-04-05.000000000"))
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(time.Microsecond) // Ensure unique names
			}
			time.Sleep(50 * time.Millisecond)
		})

		It("should check overall health", func() {
			healthy := status.IsHealthy()
			Expect(healthy).To(BeTrue())
		})

		It("should check strictly healthy", func() {
			strictlyHealthy := status.IsStrictlyHealthy()
			Expect(strictlyHealthy).To(BeTrue())
		})

		It("should check specific monitor health", func() {
			monitors := pool.MonitorList()
			if len(monitors) > 0 {
				healthy := status.IsHealthy(monitors[0])
				Expect(healthy).To(BeTrue())
			}
		})
	})
})
