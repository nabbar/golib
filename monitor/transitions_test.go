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
	"fmt"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Status Transitions", func() {
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

	Describe("Initial State", func() {
		It("should start with KO status when no check has run", func() {
			Expect(mon.Status()).To(Equal(monsts.KO))
		})

		It("should have error message before first check", func() {
			msg := mon.Message()
			Expect(msg).ToNot(BeEmpty())
			Expect(msg).To(ContainSubstring("healcheck"))
		})

		It("should not be rising or falling initially", func() {
			Expect(mon.IsRise()).To(BeFalse())
			Expect(mon.IsFall()).To(BeFalse())
		})
	})

	Describe("Rise Transitions (KO -> Warn -> OK)", func() {
		BeforeEach(func() {
			cfg := newConfig(nfo)
			cfg.IntervalRise = libdur.ParseDuration(50 * time.Millisecond)
			cfg.RiseCountKO = 2   // KO -> Warn after 2 successes
			cfg.RiseCountWarn = 2 // Warn -> OK after 2 successes
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())
		})

		It("should transition from KO to Warn after RiseCountKO successes", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil // Always succeed
			})
			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(10 * time.Millisecond)

			// Should still be KO initially
			Expect(mon.Status()).To(Equal(monsts.KO))

			// Wait for enough checks to transition to Warn
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.Warn))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should transition from Warn to OK after RiseCountWarn successes", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil // Always succeed
			})
			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for KO -> Warn
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.Warn))

			// Wait for Warn -> OK
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should set IsRise flag during rising transitions", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})
			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Should detect rising state
			Eventually(func() bool {
				return mon.IsRise()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Fall Transitions (OK -> Warn -> KO)", func() {
		BeforeEach(func() {
			cfg := newConfig(nfo)
			cfg.IntervalFall = libdur.ParseDuration(50 * time.Millisecond)
			cfg.FallCountWarn = 2 // OK -> Warn after 2 failures
			cfg.FallCountKO = 2   // Warn -> KO after 2 failures
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())
		})

		It("should transition from OK to Warn after FallCountWarn failures", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(false)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK status
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 800*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Start failing
			shouldFail.Store(true)

			// Wait for transition to Warn
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.Warn))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should transition from Warn to KO after FallCountKO failures", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(false)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK status
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 800*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Start failing
			shouldFail.Store(true)

			// Wait for Warn
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.Warn))

			// Wait for KO
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.KO))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should set IsFall flag during falling transitions", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(false)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 800*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Start failing
			shouldFail.Store(true)

			// Should detect falling state
			Eventually(func() bool {
				return mon.IsFall()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Threshold Enforcement", func() {
		BeforeEach(func() {
			// Using default newConfig which has appropriate counts
			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())
		})

		It("should not transition with insufficient successes", func() {
			checkCount := &atomic.Int32{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				count := checkCount.Add(1)
				if count < 3 {
					return nil
				}
				return ErrorMockTest // Fail before reaching threshold
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)

			// Should still be KO (didn't reach threshold)
			Expect(mon.Status()).To(Equal(monsts.KO))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should reset rise counter on failure", func() {
			checkCount := &atomic.Int32{}

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 3
			cfg.RiseCountWarn = 3
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			mon.SetHealthCheck(func(ctx context.Context) error {
				count := checkCount.Add(1)
				// Success, Success, Fail, Success, Success, Fail - never reaches threshold
				if count%3 == 0 {
					return ErrorMockTest
				}
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(800 * time.Millisecond)

			// Should still be KO (counter keeps resetting)
			Expect(mon.Status()).To(Equal(monsts.KO))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Message Updates", func() {
		It("should clear message on successful check", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(true)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 1
			cfg.RiseCountWarn = 1
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for error message
			Eventually(func() string {
				return mon.Message()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(ContainSubstring("mock test error"))

			// Start succeeding
			shouldFail.Store(false)

			// Message should be cleared
			Eventually(func() string {
				return mon.Message()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeEmpty())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should update message on new failure", func() {
			errorMsg := &atomic.Value{}
			errorMsg.Store("first error")

			mon.SetHealthCheck(func(ctx context.Context) error {
				return fmt.Errorf("%s", errorMsg.Load().(string))
			})

			cfg := montps.Config{
				Name:          "message-test",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for first error
			Eventually(func() string {
				return mon.Message()
			}, 2*time.Second, 100*time.Millisecond).Should(ContainSubstring("first error"))

			// Change error message
			errorMsg.Store("second error")

			// Message should update
			Eventually(func() string {
				return mon.Message()
			}, 2*time.Second, 100*time.Millisecond).Should(ContainSubstring("second error"))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})
})
