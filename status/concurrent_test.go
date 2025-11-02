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
	"fmt"
	"sync"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libsts "github.com/nabbar/golib/status"
	stsctr "github.com/nabbar/golib/status/control"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status/Concurrent", func() {
	Describe("Concurrent health checks", func() {
		It("should handle concurrent IsHealthy calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor-1")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup
			results := make([]bool, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()
					results[index] = status.IsHealthy()
				}(i)
			}

			wg.Wait()

			// All results should be consistent
			for _, result := range results {
				Expect(result).To(Equal(results[0]))
			}
		})

		It("should handle concurrent IsStrictlyHealthy calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor-2")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup
			results := make([]bool, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()
					results[index] = status.IsStrictlyHealthy()
				}(i)
			}

			wg.Wait()

			for _, result := range results {
				Expect(result).To(Equal(results[0]))
			}
		})

		It("should handle concurrent IsCacheHealthy calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor-3")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					_ = status.IsCacheHealthy()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent configuration updates", func() {
		It("should handle concurrent SetConfig calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					cfg := libsts.Config{
						ReturnCode: map[monsts.Status]int{
							monsts.OK:   200 + index,
							monsts.Warn: 207,
							monsts.KO:   500,
						},
					}
					status.SetConfig(cfg)
				}(i)
			}

			wg.Wait()

			// Should not panic or deadlock
			Expect(true).To(BeTrue())
		})

		It("should handle concurrent SetInfo calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					status.SetInfo(
						fmt.Sprintf("app-%d", index),
						fmt.Sprintf("v1.0.%d", index),
						fmt.Sprintf("hash-%d", index),
					)
				}(i)
			}

			wg.Wait()

			Expect(true).To(BeTrue())
		})
	})

	Describe("Concurrent monitor operations", func() {
		It("should handle concurrent MonitorAdd calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			var wg sync.WaitGroup

			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					m := newHealthyMonitor(fmt.Sprintf("monitor-%d", index))
					err := status.MonitorAdd(m)
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()

			list := status.MonitorList()
			Expect(len(list)).To(Equal(20))
		})

		It("should handle concurrent MonitorGet calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			// Add some monitors first
			for i := 0; i < 10; i++ {
				m := newHealthyMonitor(fmt.Sprintf("monitor-%d", i))
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
			}

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					monitorName := fmt.Sprintf("monitor-%d", index%10)
					_ = status.MonitorGet(monitorName)
				}(i)
			}

			wg.Wait()
		})

		It("should handle concurrent MonitorList calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			// Add some monitors
			for i := 0; i < 5; i++ {
				m := newHealthyMonitor(fmt.Sprintf("monitor-%d", i))
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
			}

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					list := status.MonitorList()
					Expect(len(list)).To(Equal(5))
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent MonitorWalk calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			// Add some monitors
			for i := 0; i < 5; i++ {
				m := newHealthyMonitor(fmt.Sprintf("monitor-%d", i))
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
			}

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					count := 0
					status.MonitorWalk(func(name string, mon montps.Monitor) bool {
						count++
						return true
					})
					Expect(count).To(Equal(5))
				}()
			}

			wg.Wait()
		})

		It("should handle mixed concurrent operations", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			var wg sync.WaitGroup

			// Add monitors
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					m := newHealthyMonitor(fmt.Sprintf("monitor-%d", index))
					_ = status.MonitorAdd(m)
				}(i)
			}

			// Get monitors
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					_ = status.MonitorGet(fmt.Sprintf("monitor-%d", index%10))
				}(i)
			}

			// List monitors
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = status.MonitorList()
				}()
			}

			// Check health
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = status.IsHealthy()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent marshaling", func() {
		It("should handle concurrent MarshalJSON calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor-4")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, err := status.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent MarshalText calls", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("test-monitor-5")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, err := status.MarshalText()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})
	})

	Describe("Concurrent with control modes", func() {
		It("should handle concurrent health checks with control modes", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			cfg := libsts.Config{
				MandatoryComponent: []libsts.Mandatory{
					{
						Mode: stsctr.Must,
						Keys: []string{"critical-1", "critical-2"},
					},
					{
						Mode: stsctr.AnyOf,
						Keys: []string{"anyof-1", "anyof-2", "anyof-3"},
					},
				},
			}
			status.SetConfig(cfg)

			// Add monitors
			for _, name := range []string{"critical-1", "critical-2", "anyof-1", "anyof-2", "anyof-3"} {
				m := newHealthyMonitor(name)
				err := pool.MonitorAdd(m)
				Expect(err).ToNot(HaveOccurred())
			}

			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = status.IsHealthy()
					_ = status.IsStrictlyHealthy()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Cache concurrent access", func() {
		It("should handle concurrent cache reads", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("cache-test")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			// Populate cache
			_ = status.IsCacheHealthy()

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_ = status.IsCacheHealthy()
					_ = status.IsCacheStrictlyHealthy()
				}()
			}

			wg.Wait()
		})

		It("should handle cache expiration during concurrent access", func() {
			status := libsts.New(globalCtx)
			status.SetInfo("concurrent-test", "v1.0.0", "abc123")

			pool := newPool()
			status.RegisterPool(func() montps.Pool { return pool })

			m := newHealthyMonitor("cache-test")
			err := pool.MonitorAdd(m)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(testMonitorStabilizeDelay)

			var wg sync.WaitGroup

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 10; j++ {
						_ = status.IsCacheHealthy()
						time.Sleep(5 * time.Millisecond)
					}
				}()
			}

			wg.Wait()
		})
	})
})
