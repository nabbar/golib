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
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor Metrics", func() {
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

		Expect(mon.SetConfig(x, newConfig(nfo))).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if mon != nil && mon.IsRunning() {
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		}
		if cnl != nil {
			cnl()
		}
	})

	Describe("Latency", func() {
		It("should track health check latency", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			Eventually(func() time.Duration {
				return mon.Latency()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">=", 10*time.Millisecond))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should update latency on each check", func() {
			checkCount := &atomic.Int32{}
			mon.SetHealthCheck(func(ctx context.Context) error {
				count := checkCount.Add(1)
				if count == 1 {
					time.Sleep(10 * time.Millisecond)
				} else {
					time.Sleep(50 * time.Millisecond)
				}
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for second check with longer latency
			Eventually(func() time.Duration {
				if checkCount.Load() >= 2 {
					return mon.Latency()
				}
				return 0
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">=", 50*time.Millisecond))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should be zero initially", func() {
			Expect(mon.Latency()).To(Equal(time.Duration(0)))
		})
	})

	Describe("Uptime", func() {
		It("should track time in OK status", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 1
			cfg.RiseCountWarn = 1
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK status
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Wait and verify uptime increases
			time.Sleep(100 * time.Millisecond)
			uptime := mon.Uptime()
			Expect(uptime).To(BeNumerically(">", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should not increase when not in OK status", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return ErrorMockTest
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			uptime := mon.Uptime()
			Expect(uptime).To(Equal(time.Duration(0)))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Downtime", func() {
		It("should track time in KO/Warn status", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return ErrorMockTest
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait and verify downtime increases
			Eventually(func() time.Duration {
				return mon.Downtime()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should increase in Warn status", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(false)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 1
			cfg.RiseCountWarn = 1
			cfg.FallCountWarn = 1
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Start failing
			shouldFail.Store(true)

			// Wait for Warn
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.Warn))

			// Downtime should start increasing
			time.Sleep(100 * time.Millisecond)
			Expect(mon.Downtime()).To(BeNumerically(">", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Rise and Fall Times", func() {
		It("should track rise time during KO to OK transition", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 2
			cfg.RiseCountWarn = 2
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for rising state
			Eventually(func() bool {
				return mon.IsRise()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			// CollectRiseTime should return non-zero
			Eventually(func() time.Duration {
				return mon.CollectRiseTime()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should track fall time during OK to KO transition", func() {
			shouldFail := &atomic.Bool{}
			shouldFail.Store(false)

			mon.SetHealthCheck(func(ctx context.Context) error {
				if shouldFail.Load() {
					return ErrorMockTest
				}
				return nil
			})

			cfg := newConfig(nfo)
			cfg.RiseCountKO = 1
			cfg.RiseCountWarn = 1
			cfg.FallCountWarn = 2
			cfg.FallCountKO = 2
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			// Start failing
			shouldFail.Store(true)

			// Wait for falling state
			Eventually(func() bool {
				return mon.IsFall()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			// CollectFallTime should return non-zero
			Eventually(func() time.Duration {
				return mon.CollectFallTime()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeNumerically(">", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("Prometheus Metrics Collection", func() {
		It("should register and collect metrics", func() {
			metricsCollected := &atomic.Bool{}
			collectedNames := &atomic.Value{}

			mon.RegisterMetricsName("test_monitor_health", "test_monitor_status")
			mon.RegisterCollectMetrics(func(ctx context.Context, names ...string) {
				metricsCollected.Store(true)
				collectedNames.Store(names)
			})

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for metrics to be collected
			Eventually(func() bool {
				return metricsCollected.Load()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			names := collectedNames.Load().([]string)
			Expect(names).To(ContainElement("test_monitor_health"))
			Expect(names).To(ContainElement("test_monitor_status"))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should add metrics names incrementally", func() {
			mon.RegisterMetricsName("metric1", "metric2")
			mon.RegisterMetricsAddName("metric3")
			mon.RegisterMetricsAddName("metric2") // Duplicate, should be ignored

			metricsCollected := &atomic.Bool{}
			collectedNames := &atomic.Value{}

			mon.RegisterCollectMetrics(func(ctx context.Context, names ...string) {
				metricsCollected.Store(true)
				collectedNames.Store(names)
			})

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			Eventually(func() bool {
				return metricsCollected.Load()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(BeTrue())

			names := collectedNames.Load().([]string)
			Expect(names).To(HaveLen(3))
			Expect(names).To(ContainElement("metric1"))
			Expect(names).To(ContainElement("metric2"))
			Expect(names).To(ContainElement("metric3"))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should provide CollectStatus metrics", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			cfg := newConfig(nfo)
			cfg.CheckTimeout = libdur.ParseDuration(5 * time.Second)
			cfg.IntervalCheck = libdur.ParseDuration(200 * time.Millisecond)
			cfg.RiseCountKO = 1
			cfg.RiseCountWarn = 1
			Expect(mon.SetConfig(x, cfg)).ToNot(HaveOccurred())

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for OK status
			Eventually(func() monsts.Status {
				return mon.Status()
			}, 500*time.Millisecond, 20*time.Millisecond).Should(Equal(monsts.OK))

			status, rise, fall := mon.CollectStatus()
			Expect(status).To(Equal(monsts.OK))
			Expect(rise).To(BeFalse())
			Expect(fall).To(BeFalse())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should provide timing metrics", func() {
			mon.SetHealthCheck(func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())

			// Wait for at least one check
			time.Sleep(500 * time.Millisecond)

			latency := mon.CollectLatency()
			uptime := mon.CollectUpTime()
			downtime := mon.CollectDownTime()
			riseTime := mon.CollectRiseTime()
			fallTime := mon.CollectFallTime()

			Expect(latency).To(BeNumerically(">=", 0))
			Expect(uptime).To(BeNumerically(">=", 0))
			Expect(downtime).To(BeNumerically(">=", 0))
			Expect(riseTime).To(BeNumerically(">=", 0))
			Expect(fallTime).To(BeNumerically(">=", 0))

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should not collect metrics if no names registered", func() {
			metricsCollected := &atomic.Bool{}

			mon.RegisterCollectMetrics(func(ctx context.Context, names ...string) {
				metricsCollected.Store(true)
			})

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(500 * time.Millisecond)

			// Should not collect without names
			Expect(metricsCollected.Load()).To(BeFalse())

			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})

		It("should not collect metrics if no function registered", func() {
			mon.RegisterMetricsName("test_metric")

			mon.SetHealthCheck(func(ctx context.Context) error {
				return nil
			})

			// Should not panic without collection function
			Expect(mon.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(500 * time.Millisecond)
			Expect(mon.Stop(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("InfoMap and InfoName", func() {
		It("should return info metadata", func() {
			infoMap := mon.InfoMap()
			Expect(infoMap).ToNot(BeNil())
			Expect(infoMap).To(HaveKey("version"))
			Expect(infoMap["version"]).To(Equal("1.0.0"))
		})

		It("should return info name", func() {
			Expect(mon.InfoName()).To(Equal(key))
		})

		It("should allow updating info", func() {
			updatedInfo := newInfoWithName("updated-name", func() (map[string]interface{}, error) {
				return map[string]interface{}{
					"version": "2.0.0",
				}, nil
			})

			mon.InfoUpd(updatedInfo)

			Expect(mon.InfoName()).To(Equal("updated-name"))
			Expect(mon.InfoMap()).To(HaveKeyWithValue("version", "2.0.0"))
		})
	})
})
