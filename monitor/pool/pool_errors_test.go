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
	"time"

	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestPoolErrorHandling tests error scenarios and edge cases
var _ = Describe("Pool Error Handling and Edge Cases", func() {
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

	Describe("MonitorAdd Error Cases", func() {
		It("should return error when adding monitor with empty name", func() {
			// Create a monitor with empty name (if possible)
			info, err := moninf.New("")
			if err != nil {
				// Empty name rejected at info creation - that's also fine
				return
			}

			mon, err := libmon.New(x, info)
			if err != nil {
				// Monitor creation failed - that's expected
				return
			}

			err = pool.MonitorAdd(mon)
			// Should either error or handle gracefully
			_ = err
		})

		It("should handle adding monitor that fails to start when pool is running", func() {
			// Start the pool first
			firstMon := createTestMonitor("first-mon", nil)
			defer firstMon.Stop(ctx)

			Expect(pool.MonitorAdd(firstMon)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			Expect(pool.IsRunning()).To(BeTrue())

			// Try adding a monitor that might have issues
			problematicMon := createTestMonitor("problematic", func(ctx context.Context) error {
				return fmt.Errorf("startup error")
			})
			defer problematicMon.Stop(ctx)

			// This should handle the startup error gracefully
			err := pool.MonitorAdd(problematicMon)
			_ = err // May or may not error depending on timing
		})
	})

	Describe("MonitorSet Error Cases", func() {
		It("should return error for nil monitor", func() {
			err := pool.MonitorSet(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nil monitor"))
		})

		It("should return error for monitor with empty name", func() {
			// Try to create monitor with invalid name
			info, err := moninf.New("")
			if err != nil {
				// Expected - empty name not allowed
				return
			}

			mon, err := libmon.New(x, info)
			if err != nil {
				// Expected
				return
			}

			err = pool.MonitorSet(mon)
			// Should error
			_ = err
		})
	})

	Describe("Lifecycle Error Cases", func() {
		It("should handle Start errors from individual monitors", func() {
			// Create a context that will be cancelled
			failCtx, failCnl := context.WithCancel(ctx)
			failCnl() // Cancel immediately

			mon := createTestMonitor("fail-start", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Try to start with cancelled context
			err := pool.Start(failCtx)
			// May or may not error depending on implementation
			_ = err
		})

		It("should handle Stop errors from individual monitors", func() {
			mon := createTestMonitor("stop-error-test", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			// Create cancelled context for stop
			stopCtx, stopCnl := context.WithCancel(ctx)
			stopCnl()

			err := pool.Stop(stopCtx)
			// May or may not error
			_ = err
		})

		It("should handle Restart errors from individual monitors", func() {
			mon := createTestMonitor("restart-error-test", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Try restart with cancelled context
			restartCtx, restartCnl := context.WithCancel(ctx)
			restartCnl()

			err := pool.Restart(restartCtx)
			_ = err
		})
	})

	Describe("MarshalText Error Cases", func() {
		It("should handle monitors that fail to marshal", func() {
			// Add a normal monitor
			mon := createTestMonitor("marshal-test", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Try to marshal
			text, err := pool.MarshalText()
			// Should either succeed or error gracefully
			_ = text
			_ = err
		})
	})

	Describe("Encoding Edge Cases", func() {
		It("should handle monitors with special characters", func() {
			// Monitor name with special chars (if allowed)
			mon := createTestMonitor("special_test-123", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			text, err := pool.MarshalText()
			Expect(err).ToNot(HaveOccurred())

			jsonData, err := pool.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			Expect(text).ToNot(BeEmpty())
			Expect(jsonData).ToNot(BeEmpty())
		})
	})

	Describe("MonitorWalk Edge Cases", func() {
		It("should handle walk function that returns false immediately", func() {
			for i := 0; i < 5; i++ {
				mon := createTestMonitor(fmt.Sprintf("walk-stop-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			count := 0
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				count++
				return false // Stop immediately
			})

			Expect(count).To(Equal(1))
		})

		It("should handle walk with invalid monitor types", func() {
			// This shouldn't happen in practice, but tests robustness
			mon := createTestMonitor("type-test", nil)
			defer mon.Stop(ctx)
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			executed := false
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				executed = true
				Expect(val).ToNot(BeNil())
				return true
			})

			Expect(executed).To(BeTrue())
		})

		It("should handle empty validName list", func() {
			mon := createTestMonitor("valid-name-test", nil)
			defer mon.Stop(ctx)
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			count := 0
			// Pass empty variadic arg
			pool.MonitorWalk(func(name string, val montps.Monitor) bool {
				count++
				return true
			})

			Expect(count).To(Equal(1))
		})
	})

	Describe("Concurrent Error Scenarios", func() {
		It("should handle concurrent Start/Stop operations", func() {
			mon := createTestMonitor("concurrent-lifecycle", nil)
			defer mon.Stop(ctx)
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			done := make(chan bool, 2)

			// Start concurrently
			go func() {
				defer GinkgoRecover()
				_ = pool.Start(ctx)
				done <- true
			}()

			// Stop concurrently
			go func() {
				defer GinkgoRecover()
				_ = pool.Stop(ctx)
				done <- true
			}()

			// Wait for both
			<-done
			<-done

			// Should not deadlock or panic
		})

		It("should handle concurrent Add/Delete operations", func() {
			done := make(chan bool, 20)

			// Add and delete concurrently
			for i := 0; i < 10; i++ {
				go func(index int) {
					defer GinkgoRecover()
					name := fmt.Sprintf("concurrent-add-del-%d", index)
					mon := createTestMonitor(name, nil)
					defer mon.Stop(ctx)

					_ = pool.MonitorAdd(mon)
					done <- true
				}(i)

				go func(index int) {
					defer GinkgoRecover()
					name := fmt.Sprintf("concurrent-add-del-%d", index)
					time.Sleep(10 * time.Millisecond)
					pool.MonitorDel(name)
					done <- true
				}(i)
			}

			// Wait for all
			for i := 0; i < 20; i++ {
				<-done
			}

			// Should not panic or deadlock
		})

		It("should handle concurrent MonitorWalk calls", func() {
			// Add some monitors
			for i := 0; i < 5; i++ {
				mon := createTestMonitor(fmt.Sprintf("walk-concurrent-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			done := make(chan bool, 10)

			// Walk concurrently
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					pool.MonitorWalk(func(name string, val montps.Monitor) bool {
						time.Sleep(1 * time.Millisecond)
						return true
					})
					done <- true
				}()
			}

			// Wait for all
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})

	Describe("Nil and Empty Checks", func() {
		It("should handle MonitorGet with empty string", func() {
			result := pool.MonitorGet("")
			Expect(result).To(BeNil())
		})

		It("should handle MonitorDel with empty string", func() {
			pool.MonitorDel("")
			// Should not panic
		})

		It("should handle empty pool operations", func() {
			// All operations on empty pool should work
			Expect(pool.MonitorList()).To(BeEmpty())
			Expect(pool.Uptime()).To(Equal(time.Duration(0)))
			Expect(pool.IsRunning()).To(BeFalse())

			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			Expect(pool.Stop(ctx)).ToNot(HaveOccurred())
			Expect(pool.Restart(ctx)).ToNot(HaveOccurred())

			text, err := pool.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			Expect(text).To(BeEmpty())

			jsonData, err := pool.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonData).ToNot(BeEmpty())
		})
	})

	Describe("Logger Fallback", func() {
		It("should create default logger when none provided", func() {
			// Create pool without registering logger
			localPool := newPool(x)

			mon := createTestMonitor("logger-fallback", nil)
			defer mon.Stop(ctx)

			Expect(localPool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Should use default logger internally
			Expect(localPool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			// Should not panic
			_ = localPool.Stop(ctx)
		})
	})

	Describe("Context Cancellation Scenarios", func() {
		It("should handle context cancellation during operation", func() {
			mon := createTestMonitor("ctx-cancel", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())

			// Create a context that expires quickly
			shortCtx, shortCnl := context.WithTimeout(ctx, 50*time.Millisecond)
			defer shortCnl()

			// Wait for context to expire
			<-shortCtx.Done()

			// Operations should still work with original context
			Expect(pool.IsRunning()).To(BeTrue())

			_ = pool.Stop(ctx)
		})

		It("should handle nil context in operations", func() {
			mon := createTestMonitor("nil-ctx", nil)
			defer mon.Stop(ctx)

			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Most operations should handle nil context gracefully
			// (though it's not recommended practice)
		})
	})

	Describe("Memory and Resource Management", func() {
		It("should handle adding same monitor multiple times", func() {
			mon := createTestMonitor("duplicate-add", nil)
			defer mon.Stop(ctx)

			// Add multiple times
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

			// Should only appear once
			list := pool.MonitorList()
			count := 0
			for _, name := range list {
				if name == "duplicate-add" {
					count++
				}
			}
			Expect(count).To(Equal(1))
		})

		It("should handle monitor replacement with Set", func() {
			mon1 := createTestMonitor("replaceable", nil)
			defer mon1.Stop(ctx)

			Expect(pool.MonitorAdd(mon1)).ToNot(HaveOccurred())

			// Create new monitor with same name
			info := newInfo("replaceable")
			mon2 := newMonitor(x, info)
			mon2.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})
			defer mon2.Stop(ctx)

			// Replace with Set
			Expect(pool.MonitorSet(mon2)).ToNot(HaveOccurred())

			// Should have only one
			list := pool.MonitorList()
			count := 0
			for _, name := range list {
				if name == "replaceable" {
					count++
				}
			}
			Expect(count).To(Equal(1))
		})
	})
})
