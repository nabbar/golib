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
	"sync/atomic"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Lifecycle Operations", func() {
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

	Describe("Start", func() {
		It("should start an empty pool without error", func() {
			err := pool.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(pool.IsRunning()).To(BeFalse()) // Empty pool is not running
		})

		It("should start all monitors in the pool", func() {
			// Add monitors
			monitors := []montps.Monitor{
				createTestMonitor("start-1", nil),
				createTestMonitor("start-2", nil),
				createTestMonitor("start-3", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			// Start the pool
			err := pool.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Give monitors time to start
			time.Sleep(100 * time.Millisecond)

			// Verify all monitors are running
			for _, name := range []string{"start-1", "start-2", "start-3"} {
				mon := pool.MonitorGet(name)
				Expect(mon).ToNot(BeNil())
				Expect(mon.IsRunning()).To(BeTrue())
			}

			Expect(pool.IsRunning()).To(BeTrue())
		})

		It("should handle monitors with health checks", func() {
			checkCount := new(atomic.Int64)

			monitor := createTestMonitor("health-check-test", func(ctx context.Context) error {
				checkCount.Add(1)
				return nil
			})
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Wait for some health checks to execute
			time.Sleep(200 * time.Millisecond)

			// Verify health checks were executed
			Expect(checkCount.Load()).To(BeNumerically(">", 0))
		})

		It("should be idempotent (calling Start twice should not error)", func() {
			monitor := createTestMonitor("idempotent-start", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// First start
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			// Second start
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			Expect(pool.IsRunning()).To(BeTrue())
		})
	})

	Describe("Stop", func() {
		It("should stop an empty pool without error", func() {
			err := pool.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop all running monitors", func() {
			// Add and start monitors
			monitors := []montps.Monitor{
				createTestMonitor("stop-1", nil),
				createTestMonitor("stop-2", nil),
				createTestMonitor("stop-3", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			// Stop the pool
			err := pool.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Give monitors time to stop
			time.Sleep(100 * time.Millisecond)

			// Verify all monitors are stopped
			for _, name := range []string{"stop-1", "stop-2", "stop-3"} {
				mon := pool.MonitorGet(name)
				Expect(mon).ToNot(BeNil())
				Expect(mon.IsRunning()).To(BeFalse())
			}

			Expect(pool.IsRunning()).To(BeFalse())
		})

		It("should be safe to call Stop on non-running pool", func() {
			monitor := createTestMonitor("stop-not-running", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// Stop without starting
			err := pool.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Restart", func() {
		It("should restart all monitors in the pool", func() {
			checkCount := new(atomic.Int64)

			monitor := createTestMonitor("restart-test", func(ctx context.Context) error {
				checkCount.Add(1)
				return nil
			})
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Wait for some checks
			time.Sleep(150 * time.Millisecond)
			firstCount := checkCount.Load()

			// Restart
			err := pool.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for more checks after restart
			time.Sleep(150 * time.Millisecond)
			secondCount := checkCount.Load()

			// Should have more checks after restart
			Expect(secondCount).To(BeNumerically(">", firstCount))
			Expect(pool.IsRunning()).To(BeTrue())
		})

		It("should work on a stopped pool", func() {
			monitor := createTestMonitor("restart-stopped", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())

			// Restart without starting
			err := pool.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Give it time to start
			time.Sleep(50 * time.Millisecond)

			retrieved := pool.MonitorGet("restart-stopped")
			Expect(retrieved).ToNot(BeNil())
			Expect(retrieved.IsRunning()).To(BeTrue())
		})
	})

	Describe("IsRunning", func() {
		It("should return false for empty pool", func() {
			Expect(pool.IsRunning()).To(BeFalse())
		})

		It("should return false for pool with stopped monitors", func() {
			monitor := createTestMonitor("not-running", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.IsRunning()).To(BeFalse())
		})

		It("should return true if at least one monitor is running", func() {
			mon1 := createTestMonitor("running-1", nil)
			mon2 := createTestMonitor("running-2", nil)

			defer func() {
				mon1.Stop(ctx)
				mon2.Stop(ctx)
			}()

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())

			// Start only one monitor
			Expect(mon1.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			Expect(pool.IsRunning()).To(BeTrue())
		})

		It("should return true when all monitors are running", func() {
			monitors := []montps.Monitor{
				createTestMonitor("all-running-1", nil),
				createTestMonitor("all-running-2", nil),
				createTestMonitor("all-running-3", nil),
			}

			for _, mon := range monitors {
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			Expect(pool.IsRunning()).To(BeTrue())
		})
	})

	Describe("Uptime", func() {
		It("should return zero for empty pool", func() {
			Expect(pool.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should return zero for pool with newly added monitors", func() {
			monitor := createTestMonitor("new-uptime", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should return maximum uptime among all monitors", func() {
			mon1 := createTestMonitor("uptime-1", nil)
			mon2 := createTestMonitor("uptime-2", nil)

			defer func() {
				mon1.Stop(ctx)
				mon2.Stop(ctx)
			}()

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())

			// Start monitors and let them run
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(200 * time.Millisecond)

			uptime := pool.Uptime()
			// Uptime should be greater than 0 after running
			Expect(uptime).To(BeNumerically(">", 0))
		})
	})

	Describe("Context Cancellation", func() {
		It("should stop monitors when context is cancelled", func() {
			localCtx, localCnl := context.WithTimeout(ctx, 500*time.Millisecond)
			defer localCnl()

			monitor := createTestMonitor("context-cancel", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(monitor.Start(localCtx)).ToNot(HaveOccurred())

			// Wait for context to be cancelled
			<-localCtx.Done()

			// Give it time to clean up
			time.Sleep(100 * time.Millisecond)

			// Monitor should eventually stop
			Eventually(func() bool {
				return monitor.IsRunning()
			}, "2s", "100ms").Should(BeFalse())
		})
	})
})
