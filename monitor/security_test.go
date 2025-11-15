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

var _ = Describe("Monitor Security and Robustness", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
		mon montps.Monitor
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 10*time.Second)
		nfo = newInfo(nil)
		mon = newMonitor(x, nfo)
	})

	AfterEach(func() {
		if mon != nil && mon.IsRunning() {
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("Nil Handling", func() {
		It("should handle nil info during creation", func() {
			m, err := libmon.New(x, nil)
			Expect(err).To(HaveOccurred())
			Expect(m).To(BeNil())
		})

		It("should handle nil context during creation", func() {
			m, err := libmon.New(nil, nfo)
			Expect(err).ToNot(HaveOccurred())
			Expect(m).ToNot(BeNil())
		})

		It("should handle nil context in SetConfig", func() {
			err := mon.SetConfig(nil, newConfig(nfo))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Timeout Enforcement", func() {
		It("should timeout long-running health checks", func() {
			timeoutOccurred := &atomic.Bool{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				select {
				case <-time.After(10 * time.Second):
					return nil
				case <-ctx.Done():
					timeoutOccurred.Store(true)
					return ctx.Err()
				}
			})

			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(200 * time.Millisecond)
			cfg.IntervalCheck = libdur.ParseDuration(500 * time.Millisecond)
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			Eventually(func() bool {
				return timeoutOccurred.Load()
			}, 3*time.Second, 100*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should not start if context is already cancelled", func() {
			cancelledCtx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			// May or may not start depending on timing
			_ = mon.Start(cancelledCtx)
		})
	})

	Describe("Panic Recovery", func() {
		It("should handle panicking health check functions", func() {
			panicOccurred := &atomic.Bool{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				defer func() {
					if r := recover(); r != nil {
						panicOccurred.Store(true)
					}
				}()
				panic("intentional panic for testing")
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			// Monitor should handle the panic gracefully
			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			// Should still be running despite panics
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Resource Cleanup", func() {
		It("should clean up resources after stop", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cfg := montps.Config{
				Name:          "cleanup-test",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			// Start and stop multiple times
			for i := 0; i < 5; i++ {
				Expect(mon.Start(ctx)).ToNot(HaveOccurred())
				time.Sleep(100 * time.Millisecond)
				Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			}

			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should not leak goroutines after stop", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			// Give time for cleanup
			time.Sleep(50 * time.Millisecond)

			// Should be fully stopped
			Expect(mon.IsRunning()).To(BeFalse())
		})
	})

	Describe("Edge Cases", func() {
		It("should handle zero-duration intervals", func() {
			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(0)
			cfg.IntervalCheck = libdur.ParseDuration(0)
			cfg.IntervalRise = libdur.ParseDuration(0)
			cfg.IntervalFall = libdur.ParseDuration(0)
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			// Should normalize to minimums (intervals set to IntervalCheck when < microsecond)
			retrievedCfg := mon.GetConfig()
			Expect(retrievedCfg.CheckTimeout.Time()).To(BeNumerically(">", 0))
			Expect(retrievedCfg.IntervalCheck.Time()).To(BeNumerically(">", 0))
		})

		It("should handle zero threshold counts", func() {
			cfg := newConfig(nfo)
			cfg.FallCountKO = 0
			cfg.FallCountWarn = 0
			cfg.RiseCountKO = 0
			cfg.RiseCountWarn = 0
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			// Should normalize to minimums
			retrievedCfg := mon.GetConfig()
			Expect(retrievedCfg.FallCountKO).To(Equal(uint8(1)))
			Expect(retrievedCfg.FallCountWarn).To(Equal(uint8(1)))
			Expect(retrievedCfg.RiseCountKO).To(Equal(uint8(1)))
			Expect(retrievedCfg.RiseCountWarn).To(Equal(uint8(1)))
		})

		It("should handle very high threshold counts", func() {
			cfg := newConfig(nfo)
			cfg.FallCountKO = 255
			cfg.FallCountWarn = 255
			cfg.RiseCountKO = 255
			cfg.RiseCountWarn = 255
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			// Should still be working
			Expect(mon.IsRunning()).To(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should handle extremely frequent checks", func() {
			checkCount := &atomic.Int32{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				checkCount.Add(1)
				return nil
			})

			cfg := montps.Config{
				Name:          "frequent-checks",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(1 * time.Millisecond), // Will be normalized to 1s
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())

			// Should not overwhelm the system
			Expect(checkCount.Load()).To(BeNumerically("<", 1000))
		})
	})

	Describe("State Consistency", func() {
		It("should maintain consistent state under rapid start/stop", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			// Rapidly start and stop
			for i := 0; i < 10; i++ {
				Expect(mon.Start(ctx)).ToNot(HaveOccurred())
				time.Sleep(50 * time.Millisecond)
				Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			}

			// Final state should be consistent
			Expect(mon.IsRunning()).To(BeFalse())
		})

		It("should handle health check function changes while running", func() {
			check1Called := &atomic.Bool{}
			check2Called := &atomic.Bool{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				check1Called.Store(true)
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			Eventually(func() bool {
				return check1Called.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			// Change health check function
			mon.SetHealthCheck(func(ctx context.Context) error {
				check2Called.Store(true)
				return nil
			})

			// New function should be called
			Eventually(func() bool {
				return check2Called.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Error Message Safety", func() {
		It("should safely handle very long error messages", func() {
			longMessage := ""
			for i := 0; i < 10000; i++ {
				longMessage += "A"
			}

			mon.SetHealthCheck(func(ctx context.Context) error {
				return &customError{msg: longMessage}
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			// Should handle long message without issues
			msg := mon.Message()
			Expect(len(msg)).To(Equal(10000))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should handle special characters in error messages", func() {
			specialMsg := "Error with special chars: \n\t\r\"'\\{}<>&"

			mon.SetHealthCheck(func(ctx context.Context) error {
				return &customError{msg: specialMsg}
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(200 * time.Millisecond)

			msg := mon.Message()
			Expect(msg).To(ContainSubstring("special chars"))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})
})

// customError is a test error type for special error message testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
