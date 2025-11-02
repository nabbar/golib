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
	monpol "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/Pool", func() {
	var (
		status libsts.Status
		pool   monpol.Pool
	)

	BeforeEach(func() {
		status = libsts.New(globalCtx)
		pool = newPool()
		status.RegisterPool(func() montps.Pool { return pool })
	})

	Describe("MonitorList", func() {
		Context("with empty pool", func() {
			It("should return empty list", func() {
				status = libsts.New(globalCtx)
				pool = newPool()
				status.RegisterPool(func() montps.Pool { return pool })
				list := status.MonitorList()
				Expect(list).To(BeEmpty())
			})
		})

		Context("with monitors", func() {
			BeforeEach(func() {
				// Add monitors with explicit names
				names := []string{"monitor-1", "monitor-2", "monitor-3"}
				for _, name := range names {
					m := newHealthyMonitor(name)
					err := pool.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should return list of monitors", func() {
				list := status.MonitorList()
				Expect(list).To(HaveLen(3))
				Expect(list).To(ContainElement("monitor-1"))
				Expect(list).To(ContainElement("monitor-2"))
				Expect(list).To(ContainElement("monitor-3"))
			})
		})
	})

	Describe("MonitorWalk", func() {
		BeforeEach(func() {
			// Add monitors with explicit names
			names := []string{"walk-1", "walk-2", "walk-3"}
			for _, name := range names {
				m := newHealthyMonitor(name)
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should walk through all monitors", func() {
			count := 0
			status.MonitorWalk(func(name string, cpt montps.Monitor) bool {
				count++
				Expect(name).ToNot(BeEmpty())
				Expect(cpt).ToNot(BeNil())
				return true
			})
			Expect(count).To(Equal(3))
		})

		It("should stop walking when function returns false", func() {
			count := 0
			status.MonitorWalk(func(name string, cpt montps.Monitor) bool {
				count++
				return count < 2 // Stop after 2 iterations
			})
			Expect(count).To(Equal(2))
		})
	})

	Describe("MonitorAdd", func() {
		It("should add a monitor", func() {
			m := newHealthyMonitor("new-monitor")
			err := status.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())

			list := status.MonitorList()
			Expect(list).To(ContainElement("new-monitor"))
		})
	})

	Describe("MonitorGet", func() {
		BeforeEach(func() {
			m := newHealthyMonitor("test-monitor")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should get an existing monitor", func() {
			mon := status.MonitorGet("test-monitor")
			Expect(mon).ToNot(BeNil())
		})

		It("should return nil for non-existent monitor", func() {
			mon := status.MonitorGet("non-existent")
			Expect(mon).To(BeNil())
		})
	})

	Describe("MonitorSet", func() {
		It("should set/update a monitor", func() {
			m := newHealthyMonitor("update-monitor")
			err := status.MonitorSet(m)
			Expect(err).ToNot(HaveOccurred())

			list := status.MonitorList()
			Expect(list).To(ContainElement("update-monitor"))
		})
	})

	Describe("MonitorDel", func() {
		BeforeEach(func() {
			m := newHealthyMonitor("delete-monitor")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete a monitor", func() {
			status.MonitorDel("delete-monitor")

			list := status.MonitorList()
			Expect(list).ToNot(ContainElement("delete-monitor"))
		})
	})
})
