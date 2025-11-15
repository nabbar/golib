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
	"fmt"
	"sync/atomic"
	"time"

	monpool "github.com/nabbar/golib/monitor/pool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool Metrics Operations", func() {
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

	Describe("RegisterFctProm", func() {
		It("should register Prometheus function without error", func() {
			promFunc := newPromFunc()
			pool.RegisterFctProm(promFunc)
			// Should not panic or error
		})

		It("should allow nil Prometheus function", func() {
			pool.RegisterFctProm(nil)
			// Should not panic
		})
	})

	Describe("RegisterFctLogger", func() {
		It("should register logger function without error", func() {
			pool.RegisterFctLogger(fl)
			// Should not panic or error
		})

		It("should allow nil logger function", func() {
			pool.RegisterFctLogger(nil)
			// Should not panic
		})
	})

	Describe("InitMetrics", func() {
		It("should initialize metrics with both functions", func() {
			localPool := newPool(ctx)
			promFunc := newPromFunc()

			err := localPool.RegisterMetrics(promFunc, fl)
			defer localPool.UnregisterMetrics()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil functions gracefully", func() {
			localPool := newPool(ctx)
			err := localPool.RegisterMetrics(nil, nil)
			defer localPool.UnregisterMetrics()

			// Should complete without panic
			_ = err
		})
	})

	Describe("TriggerCollectMetrics", func() {
		It("should trigger metrics collection periodically", func() {
			// This is a long-running goroutine test
			// We'll use a short-lived context
			metricsCtx, metricsCnl := context.WithTimeout(ctx, 500*time.Millisecond)
			defer metricsCnl()

			// Add a monitor to collect metrics from
			monitor := createTestMonitor("metrics-trigger", nil)
			defer monitor.Stop(ctx)

			Expect(pool.MonitorAdd(monitor)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Register Prometheus with the pool
			promFunc := newPromFunc()
			pool.RegisterFctProm(promFunc)

			// Trigger collection every 100ms
			done := make(chan bool)
			go func() {
				pool.TriggerCollectMetrics(metricsCtx, 100*time.Millisecond)
				done <- true
			}()

			// Wait for context to expire or collection to complete
			select {
			case <-done:
				// Completed
			case <-time.After(1 * time.Second):
				// Timeout
			}

			// Should complete without errors

			// collect Metrics
			prm := promFunc()
			prm.Collect(ctx)
		})

		It("should stop when context is cancelled", func() {
			metricsCtx, metricsCnl := context.WithTimeout(ctx, 200*time.Millisecond)
			defer metricsCnl()

			done := make(chan bool, 1)
			go func() {
				pool.TriggerCollectMetrics(metricsCtx, 50*time.Millisecond)
				done <- true
			}()

			// Wait for goroutine to complete
			select {
			case <-done:
				// Should complete after context cancellation
			case <-time.After(1 * time.Second):
				Fail("TriggerCollectMetrics did not stop after context cancellation")
			}
		})

		It("should handle nil Prometheus gracefully", func() {
			metricsCtx, metricsCnl := context.WithTimeout(ctx, 200*time.Millisecond)
			defer metricsCnl()

			// Don't register Prometheus function
			done := make(chan bool, 1)
			go func() {
				pool.TriggerCollectMetrics(metricsCtx, 50*time.Millisecond)
				done <- true
			}()

			// Should complete without panic
			select {
			case <-done:
				// OK
			case <-time.After(1 * time.Second):
				// Timeout - but should not panic
			}
		})
	})

	Describe("Metrics Integration", func() {
		It("should collect metrics from running monitors", func() {
			// Create monitors with health checks
			checkCount1 := new(atomic.Int64)
			checkCount2 := new(atomic.Int64)

			mon1 := createTestMonitor("metrics-int-1", func(ctx context.Context) error {
				checkCount1.Add(1)
				return nil
			})

			mon2 := createTestMonitor("metrics-int-2", func(ctx context.Context) error {
				checkCount2.Add(1)
				return nil
			})

			defer func() {
				mon1.Stop(ctx)
				mon2.Stop(ctx)
			}()

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon2)).ToNot(HaveOccurred())

			// Start monitors
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Wait for health checks to execute
			time.Sleep(200 * time.Millisecond)

			// Verify health checks were called
			Expect(checkCount1.Load()).To(BeNumerically(">", 0))
			Expect(checkCount2.Load()).To(BeNumerically(">", 0))

			// Pool uptime should be greater than zero
			Expect(pool.Uptime()).To(BeNumerically(">", 0))
		})
	})

	Describe("Concurrent Metrics Operations", func() {
		It("should handle concurrent metric registrations", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()

					promFunc := newPromFunc()
					pool.RegisterFctProm(promFunc)
					pool.RegisterFctLogger(fl)

					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				<-done
			}

			// Should complete without data races
		})
	})

	Describe("Metrics Collection Coverage", func() {
		It("should collect all metric types from monitors", func() {
			// Create a dedicated pool for this test
			localCtx, localCnl := context.WithTimeout(x, 30*time.Second)
			defer localCnl()

			localPool := newPool(localCtx)

			// Create test monitors
			mon1 := createTestMonitor("aaa-metrics-coverage-1", func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			Expect(mon1.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				_ = mon1.Stop(context.Background())
			}()

			mon2 := createTestMonitor("bbb-metrics-coverage-2", func(ctx context.Context) error {
				time.Sleep(5 * time.Millisecond)
				return nil
			})
			Expect(mon2.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				_ = mon2.Stop(context.Background())
			}()

			// Add monitors to pool
			Expect(localPool.MonitorAdd(mon1)).ToNot(HaveOccurred())
			Expect(localPool.MonitorAdd(mon2)).ToNot(HaveOccurred())

			// Initialize metrics with Prometheus
			Expect(localPool.InitMetrics(newPromFunc(), fl)).ToNot(HaveOccurred())
			defer func() {
				localPool.ShutDown()
			}()

			// Start the pool
			Expect(localPool.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				Expect(localPool.Stop(localCtx)).ToNot(HaveOccurred())
			}()

			// Wait for monitors to run and collect some data
			time.Sleep(300 * time.Millisecond)

			// Collect metrics - this should trigger all collect functions
			newProm().CollectMetrics(localCtx)

			// Verify pool is running
			Expect(localPool.IsRunning()).To(BeTrue())
			Expect(localPool.Uptime()).To(BeNumerically(">", 0))
		})

		It("should handle metrics collection with failing monitors", func() {
			// Create a dedicated pool for this test
			localCtx, localCnl := context.WithTimeout(x, 30*time.Second)
			defer localCnl()

			localPool := newPool(localCtx)

			// Create a failing monitor
			failCount := new(atomic.Int64)
			monFail := createTestMonitor("ccc-metrics-fail", func(ctx context.Context) error {
				if failCount.Add(1) > 2 {
					return nil // Succeed after 2 failures
				}
				return fmt.Errorf("simulated failure")
			})
			Expect(monFail.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				_ = monFail.Stop(context.Background())
			}()

			// Create a monitor that transitions states
			checkCount := new(atomic.Int64)
			monTransition := createTestMonitor("ddd-metrics-transition", func(ctx context.Context) error {
				count := checkCount.Add(1)
				// Fail on specific checks to trigger state transitions
				if count == 3 || count == 6 {
					return fmt.Errorf("transition failure")
				}
				return nil
			})
			Expect(monTransition.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				_ = monTransition.Stop(context.Background())
			}()

			// Add monitors to pool
			Expect(localPool.MonitorAdd(monFail)).ToNot(HaveOccurred())
			Expect(localPool.MonitorAdd(monTransition)).ToNot(HaveOccurred())

			// Initialize metrics with Prometheus
			Expect(localPool.InitMetrics(newPromFunc(), fl)).ToNot(HaveOccurred())
			defer func() {
				localPool.ShutDown()
			}()

			// Start the pool
			Expect(localPool.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				Expect(localPool.Stop(localCtx)).ToNot(HaveOccurred())
			}()

			// Wait for monitors to run and experience state transitions
			time.Sleep(800 * time.Millisecond)

			// Collect metrics multiple times to cover different states
			for i := 0; i < 3; i++ {
				newProm().CollectMetrics(localCtx)
				time.Sleep(100 * time.Millisecond)
			}

			// Verify pool is running
			Expect(localPool.IsRunning()).To(BeTrue())
		})

		It("should collect metrics without logger", func() {
			// Create a dedicated pool WITHOUT logger
			localCtx, localCnl := context.WithTimeout(x, 30*time.Second)
			defer localCnl()

			localPool := newPool(localCtx)

			mon := createTestMonitor("eee-metrics-nolog", func(ctx context.Context) error {
				return nil
			})
			Expect(mon.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				_ = mon.Stop(context.Background())
			}()

			Expect(localPool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Initialize metrics WITHOUT logger (nil)
			Expect(localPool.InitMetrics(newPromFunc(), nil)).ToNot(HaveOccurred())
			defer func() {
				localPool.ShutDown()
			}()

			Expect(localPool.Start(localCtx)).ToNot(HaveOccurred())
			defer func() {
				Expect(localPool.Stop(localCtx)).ToNot(HaveOccurred())
			}()

			time.Sleep(200 * time.Millisecond)

			// This should trigger setDefaultLog() path
			newProm().CollectMetrics(localCtx)

			Expect(localPool.IsRunning()).To(BeTrue())
		})
	})
})
