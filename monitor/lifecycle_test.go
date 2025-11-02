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

package monitor_test

import (
	"context"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libmon "github.com/nabbar/golib/monitor"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Lifecycle", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
		mon montps.Monitor
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 3*time.Second)
		nfo = newInfo(nil)
		mon = newMonitor(x, nfo)

		cfg := newConfig(nfo)
		cfg.CheckTimeout = libdur.ParseDuration(5 * time.Second)
		cfg.IntervalCheck = libdur.ParseDuration(200 * time.Millisecond)
		Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if mon != nil && mon.IsRunning() {
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("Start", func() {
		It("should start the monitor successfully", func() {
			called := &atomic.Bool{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				called.Store(true)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			// Wait for at least one health check
			Eventually(func() bool {
				return called.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should execute health checks periodically", func() {
			callCount := &atomic.Int32{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				callCount.Add(1)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for multiple checks
			Eventually(func() int32 {
				return callCount.Load()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">=", int32(3)))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should handle missing health check function", func() {
			// Don't set a health check function
			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			// Should have error message
			Eventually(func() string {
				return mon.Message()
			}, 1*time.Second, 50*time.Millisecond).Should(ContainSubstring("healthcheck"))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should stop existing monitor before starting new one", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			// Start again without explicit stop
			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Stop", func() {
		It("should stop the monitor successfully", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should be idempotent", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			// Stop again
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should stop health check execution", func() {
			callCount := &atomic.Int32{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				callCount.Add(1)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for some checks
			time.Sleep(500 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			// Record count after stop
			countAfterStop := callCount.Load()

			// Wait and verify no more checks
			time.Sleep(500 * time.Millisecond)
			Expect(callCount.Load()).To(Equal(countAfterStop))
		})
	})

	Describe("Restart", func() {
		It("should restart the monitor successfully", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Restart(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should continue health checks after restart", func() {
			callCount := &atomic.Int32{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				callCount.Add(1)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(300 * time.Millisecond)

			Expect(mon.Restart(ctx)).ToNot(HaveOccurred())

			// Reset counter
			callCount.Store(0)

			// Verify checks continue
			Eventually(func() int32 {
				return callCount.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should work when monitor is not running", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Restart(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("IsRunning", func() {
		It("should return false initially", func() {
			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should return true after start", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should return false after stop", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeFalse())
		})
	})

	Describe("Context Cancellation", func() {
		It("should handle context cancellation during start", func() {
			localCtx, localCnl := context.WithTimeout(ctx, 100*time.Millisecond)
			defer localCnl()

			mon.SetHealthCheck(func(ctx context.Context) error {
				time.Sleep(200 * time.Millisecond) // Longer than context timeout
				return nil
			})

			// This might fail or succeed depending on timing
			_ = mon.Start(localCtx)
		})

		It("should handle context cancellation during health check", func() {
			checkCtxCancelled := &atomic.Bool{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					checkCtxCancelled.Store(true)
					return ctx.Err()
				case <-time.After(10 * time.Second):
					return nil
				}
			})

			cfg := montps.Config{
				Name:          "ctx-test",
				CheckTimeout:  libdur.ParseDuration(100 * time.Millisecond),
				IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for timeout to trigger
			Eventually(func() bool {
				return checkCtxCancelled.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Clone", func() {
		It("should create a copy of the monitor", func() {
			cloneCtx, cloneCnl := context.WithTimeout(context.Background(), 5*time.Second)
			defer cloneCnl()

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cloned, err := mon.Clone(cloneCtx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
		})

		It("should start cloned monitor if original is running", func() {
			cloneCtx, cloneCnl := context.WithTimeout(context.Background(), 5*time.Second)
			defer cloneCnl()

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeTrue())

			cloned, err := mon.Clone(cloneCtx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())

			// Give it time to start
			time.Sleep(200 * time.Millisecond)
			Expect(cloned.IsRunning()).To(BeTrue())

			Expect(cloned.Stop(cloneCtx)).ToNot(HaveOccurred())
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should not start cloned monitor if original is not running", func() {
			cloneCtx, cloneCnl := context.WithTimeout(context.Background(), 5*time.Second)
			defer cloneCnl()

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cloned, err := mon.Clone(cloneCtx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.IsRunning()).To(BeFalse())
		})
	})

	Describe("Creation", func() {
		It("should create monitor with valid info", func() {
			inf := newInfo(nil)
			m, err := libmon.New(x, inf)
			Expect(err).ToNot(HaveOccurred())
			Expect(m).ToNot(BeNil())
		})

		It("should fail with nil info", func() {
			m, err := libmon.New(x, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("info cannot be nil"))
			Expect(m).To(BeNil())
		})

		It("should use default context when nil", func() {
			inf := newInfo(nil)
			m, err := libmon.New(nil, inf)
			Expect(err).ToNot(HaveOccurred())
			Expect(m).ToNot(BeNil())
		})
	})
})
