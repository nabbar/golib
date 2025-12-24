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
	"sync"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	moninf "github.com/nabbar/golib/monitor/info"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Integration Tests", func() {
	var (
		ctx context.Context
		cnl context.CancelFunc
		nfo montps.Info
		mon montps.Monitor
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 5*time.Second)
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

	Describe("Real-World Scenario: Database Connection Monitoring", func() {
		It("should handle flapping connection", func() {
			connectionUp := &atomic.Bool{}
			connectionUp.Store(false)
			checkCount := &atomic.Int32{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				count := checkCount.Add(1)
				// Simulate flapping: fail, fail, success, fail, fail, success
				if count%3 == 0 {
					connectionUp.Store(true)
					return nil
				}
				connectionUp.Store(false)
				return ErrorMockTest
			})

			cfg := montps.Config{
				Name:          "db-monitor",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
				RiseCountKO:   3, // Require stable connection
				FallCountWarn: 3,
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(500 * time.Millisecond)

			// Should remain KO due to flapping (never gets 3 consecutive successes)
			Expect(mon.Status()).To(Equal(monsts.KO))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should recover from temporary outage", func() {
			outageUntil := time.Now().Add(800 * time.Millisecond)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if time.Now().Before(outageUntil) {
					return ErrorMockTest
				}
				return nil
			})

			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(5 * time.Second)
			cfg.IntervalCheck = libdur.ParseDuration(200 * time.Millisecond)
			cfg.RiseCountKO = 2
			cfg.RiseCountWarn = 2

			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())
			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Should start KO
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Status()).To(Equal(monsts.KO))

			// Should recover after outage ends
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 1500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Real-World Scenario: API Endpoint Monitoring", func() {
		It("should track latency degradation", func() {
			latencies := []time.Duration{
				10 * time.Millisecond,
				20 * time.Millisecond,
				50 * time.Millisecond,
				100 * time.Millisecond,
			}
			checkCount := &atomic.Int32{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				count := int(checkCount.Add(1)) - 1
				if count < len(latencies) {
					time.Sleep(latencies[count])
				} else {
					time.Sleep(latencies[len(latencies)-1])
				}
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for multiple checks
			Eventually(func() int32 {
				return checkCount.Load()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">=", 4))

			// Latency should reflect the degradation
			latency := mon.Latency()
			Expect(latency).To(BeNumerically(">=", 50*time.Millisecond))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Real-World Scenario: Multiple Monitors", func() {
		It("should run multiple monitors independently", func() {
			// Create second monitor
			nfo2 := newInfoWithName("monitor-2", nil)
			mon2 := newMonitor(x, nfo2)
			defer func() {
				if mon2.IsRunning() {
					Expect(mon2.Stop(ctx)).ToNot(HaveOccurred())
				}
			}()

			check1Count := &atomic.Int32{}
			check2Count := &atomic.Int32{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				check1Count.Add(1)
				return nil
			})

			mon2.SetHealthCheck(func(ctx context.Context) error {
				check2Count.Add(1)
				return ErrorMockTest
			})

			cfg1 := newConfig(nfo)
			cfg1.RiseCountKO = 1
			cfg1.RiseCountWarn = 1
			Expect(mon.SetConfig(x, cfg1)).ToNot(HaveOccurred())

			cfg2 := newConfig(nfo2)
			Expect(mon2.SetConfig(x, cfg2)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			Expect(mon2.Start(ctx)).ToNot(HaveOccurred())

			time.Sleep(200 * time.Millisecond)

			// Both should be running independently
			Expect(check1Count.Load()).To(BeNumerically(">", 0))
			Expect(check2Count.Load()).To(BeNumerically(">", 0))

			// mon should be OK, mon2 should be KO
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 2*time.Second, 100*time.Millisecond).Should(Equal(monsts.OK))
			Expect(mon2.Status()).To(Equal(monsts.KO))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			Expect(mon2.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Real-World Scenario: Dynamic Info Updates", func() {
		It("should use updated info in encoding", func() {
			mon.InfoUpd(newInfoWithName("test-integration", nil))
			cfg := montps.Config{
				Name:          "dynamic-info",
				CheckTimeout:  libdur.ParseDuration(5 * time.Second),
				IntervalCheck: libdur.ParseDuration(200 * time.Millisecond),
				Logger:        lo.Clone(),
			}
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			// Check initial info
			Expect(mon.InfoName()).To(Equal("test-integration"))

			// Update info with dynamic function
			newInfo, err := moninf.New("updated-service")
			Expect(err).ToNot(HaveOccurred())

			mon.InfoUpd(newInfo)

			// Verify updated info
			Expect(mon.InfoName()).To(Equal("updated-service"))

			// Verify encoding uses new info
			text, err := mon.MarshalText()
			Expect(err).ToNot(HaveOccurred())
			// The text encoding should contain the monitor name
			Expect(string(text)).To(ContainSubstring("dynamic-info"))
		})
	})

	Describe("Real-World Scenario: Graceful Shutdown", func() {
		It("should handle graceful shutdown during check", func() {
			checkStarted := &atomic.Bool{}
			checkCompleted := &atomic.Bool{}

			mon.SetHealthCheck(func(ctx context.Context) error {
				checkStarted.Store(true)
				select {
				case <-time.After(500 * time.Millisecond):
					checkCompleted.Store(true)
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(2 * time.Second)
			cfg.IntervalCheck = libdur.ParseDuration(100 * time.Millisecond)
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for check to start
			Eventually(func() bool {
				return checkStarted.Load()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			// Stop while check is running
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
			Expect(mon.IsRunning()).To(BeFalse())
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent reads safely", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						_ = mon.Status()
						_ = mon.Latency()
						_ = mon.Uptime()
						_ = mon.Downtime()
						_ = mon.IsRise()
						_ = mon.IsFall()
						_ = mon.Message()
					}
				}()
			}

			wg.Wait()
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should handle concurrent config updates safely", func() {
			var wg sync.WaitGroup
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					cfg := newConfig(nfo)
					cfg.Name = "concurrent-config"
					cfg.IntervalCheck = libdur.ParseDuration(time.Duration(100+index*10) * time.Millisecond)
					_ = mon.SetConfig(x, cfg)
				}(i)
			}

			wg.Wait()

			// Should have some valid config
			cfg := mon.GetConfig()
			Expect(cfg.Name).To(Equal("concurrent-config"))
		})
	})
})
