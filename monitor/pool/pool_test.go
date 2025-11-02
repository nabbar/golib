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

package pool_test

import (
	"context"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Basic Operations", func() {
	var (
		pool monpool.Pool
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 10*time.Second)
		pool = newPool(ctx)
	})

	AfterEach(func() {
		if pool != nil && pool.IsRunning() {
			_ = pool.Stop(ctx)
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("Pool Creation", func() {
		It("should create a new pool instance", func() {
			Expect(pool).ToNot(BeNil())
		})

		It("should not be running initially", func() {
			Expect(pool.IsRunning()).To(BeFalse())
		})

		It("should have empty monitor list initially", func() {
			Expect(pool.MonitorList()).To(BeEmpty())
		})

		It("should have zero uptime initially", func() {
			Expect(pool.Uptime()).To(Equal(time.Duration(0)))
		})
	})

	Describe("MonitorAdd", func() {
		var monitor montps.Monitor

		BeforeEach(func() {
			monitor = createTestMonitor("test-monitor", nil)
		})

		AfterEach(func() {
			if monitor != nil && monitor.IsRunning() {
				_ = monitor.Stop(ctx)
			}
		})

		It("should add a monitor to the pool", func() {
			err := pool.MonitorAdd(monitor)
			Expect(err).ToNot(HaveOccurred())

			retrieved := pool.MonitorGet("test-monitor")
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.Name()).To(Equal("test-monitor"))
		})

		It("should return nil when adding nil monitor", func() {
			err := pool.MonitorAdd(nil)
			Expect(err).To(BeNil())
		})

		It("should add multiple monitors", func() {
			mon1 := createTestMonitor("monitor-1", nil)
			mon2 := createTestMonitor("monitor-2", nil)
			mon3 := createTestMonitor("monitor-3", nil)

			defer func() {
				_ = mon1.Stop(ctx)
				_ = mon2.Stop(ctx)
				_ = mon3.Stop(ctx)
			}()

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon3)).ToNot(HaveOccurred())

			list := pool.MonitorList()
			Expect(list).To(HaveLen(3))
			Expect(list).To(ContainElements("monitor-1", "monitor-2", "monitor-3"))
		})

		It("should start monitor if pool is running", func() {
			// Add a first monitor to make the pool "running"
			firstMonitor := createTestMonitor("first-monitor", nil)
			defer firstMonitor.Stop(ctx)

			Expect(pool.MonitorAdd(firstMonitor)).ToNot(HaveOccurred())

			// Start the pool
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Give it time to start
			time.Sleep(100 * time.Millisecond)

			// Verify pool is running
			Expect(pool.IsRunning()).To(BeTrue())

			// Add a second monitor - it should be automatically started
			err := pool.MonitorAdd(monitor)
			Expect(err).ToNot(HaveOccurred())

			// Give it time to start
			time.Sleep(100 * time.Millisecond)

			retrieved := pool.MonitorGet("test-monitor")
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.IsRunning()).To(BeTrue())
		})
	})

	Describe("MonitorGet", func() {
		It("should retrieve existing monitor", func() {
			monitor := createTestMonitor("get-test", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			retrieved := pool.MonitorGet("get-test")
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.Name()).To(Equal("get-test"))
		})

		It("should return nil for non-existent monitor", func() {
			retrieved := pool.MonitorGet("non-existent")
			Expect(retrieved).To(BeNil())
		})

		It("should return nil for empty name", func() {
			retrieved := pool.MonitorGet("")
			Expect(retrieved).To(BeNil())
		})
	})

	Describe("MonitorSet", func() {
		It("should update existing monitor", func() {
			monitor1 := createTestMonitor("set-test", nil)
			defer monitor1.Stop(ctx)

			Expect(pool.MonitorAdd(monitor1)).ToNot(HaveOccurred())

			// Create new monitor with same name
			info := newInfo("set-test")
			monitor2 := newMonitor(x, info)
			monitor2.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})
			defer monitor2.Stop(ctx)

			Expect(pool.MonitorSet(monitor2)).ToNot(HaveOccurred())

			retrieved := pool.MonitorGet("set-test")
			Expect(retrieved).ToNot(BeNil())
		})

		It("should add monitor if it doesn't exist", func() {
			monitor := createTestMonitor("new-monitor", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorSet(monitor)).ToNot(HaveOccurred())

			retrieved := pool.MonitorGet("new-monitor")
			Expect(retrieved).ToNot(BeNil())
		})

		It("should return error for nil monitor", func() {
			err := pool.MonitorSet(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nil monitor"))
		})
	})

	Describe("MonitorDel", func() {
		It("should delete existing monitor", func() {
			monitor := createTestMonitor("del-test", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// Verify it exists
			Expect(pool.MonitorGet("del-test")).ToNot(BeNil())

			// Delete it
			pool.MonitorDel("del-test")

			// Verify it's gone
			Expect(pool.MonitorGet("del-test")).To(BeNil())
		})

		It("should handle deleting non-existent monitor gracefully", func() {
			pool.MonitorDel("non-existent")
			// Should not panic or error
		})

		It("should handle empty name gracefully", func() {
			pool.MonitorDel("")
			// Should not panic or error
		})
	})

	Describe("MonitorList", func() {
		It("should return empty list when no monitors", func() {
			list := pool.MonitorList()
			Expect(list).To(BeEmpty())
		})

		It("should list all monitor names", func() {
			monitors := []montps.Monitor{
				createTestMonitor("list-1", nil),
				createTestMonitor("list-2", nil),
				createTestMonitor("list-3", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			list := pool.MonitorList()
			Expect(list).To(HaveLen(3))
			Expect(list).To(ConsistOf("list-1", "list-2", "list-3"))
		})
	})

	Describe("MonitorWalk", func() {
		BeforeEach(func() {
			// Add test monitors
			for i := 1; i <= 3; i++ {
				mon := createTestMonitor(GinkgoT().Name()+"-"+string(rune('0'+i)), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should iterate over all monitors", func() {
			count := 0
			names := make([]string, 0)

			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				count++
				names = append(names, name)
				Expect(val).ToNot(BeNil())
				return true
			})

			Expect(count).To(Equal(3))
			Expect(names).To(HaveLen(3))
		})

		It("should stop iteration when function returns false", func() {
			count := 0

			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				count++
				return false // Stop after first
			})

			Expect(count).To(Equal(1))
		})

		It("should filter by specific names", func() {
			// Add monitors with predictable names
			mon1 := createTestMonitor("walk-filter-1", nil)
			mon2 := createTestMonitor("walk-filter-2", nil)
			mon3 := createTestMonitor("walk-filter-3", nil)

			defer func() {
				mon1.Stop(ctx)
				mon2.Stop(ctx)
				mon3.Stop(ctx)
			}()

			// Clear previous monitors
			for _, name := range pool.MonitorList() {
				pool.MonitorDel(name)
			}

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon3)).ToNot(HaveOccurred())

			count := 0
			names := make([]string, 0)

			// Walk only specific monitors
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				count++
				names = append(names, name)
				return true
			}, "walk-filter-1", "walk-filter-3")

			Expect(count).To(Equal(2))
			Expect(names).To(ConsistOf("walk-filter-1", "walk-filter-3"))
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent additions", func() {
			done := make(chan bool, 10)

			// Add monitors concurrently
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					name := "concurrent-" + string(rune('0'+index))
					mon := createTestMonitor(name, nil)
					defer mon.Stop(ctx)

					err := pool.MonitorAdd(mon)
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				<-done
			}

			// Verify all monitors were added
			list := pool.MonitorList()
			Expect(len(list)).To(BeNumerically(">=", 10))
		})

		It("should handle concurrent reads", func() {
			// Add a monitor
			mon := createTestMonitor("concurrent-read", nil)
			defer mon.Stop(ctx)
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			done := make(chan bool, 20)

			// Read concurrently
			for i := 0; i < 20; i++ {
				go func() {
					defer GinkgoRecover()
					retrieved := pool.MonitorGet("concurrent-read")
					Expect(retrieved).ToNot(BeNil())
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 20; i++ {
				<-done
			}
		})
	})
})
