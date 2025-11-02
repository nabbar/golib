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

	monpool "github.com/nabbar/golib/monitor/pool"
	montps "github.com/nabbar/golib/monitor/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// BenchmarkTests provides performance benchmarks for pool operations
var _ = Describe("Pool Performance Benchmarks", func() {
	var (
		pool monpool.Pool
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(x, 30*time.Second)
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

	Describe("MonitorAdd Performance", func() {
		It("should add monitors efficiently", func() {
			start := time.Now()

			for i := 0; i < 100; i++ {
				mon := createTestMonitor(fmt.Sprintf("perf-add-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Added 100 monitors in %v\n", elapsed)

			// Should be reasonably fast
			Expect(elapsed).To(BeNumerically("<", 1*time.Second))
		})

		It("should handle sequential additions efficiently", func() {
			start := time.Now()

			for i := 0; i < 50; i++ {
				mon := createTestMonitor(fmt.Sprintf("seq-add-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())

				// Small operation between additions
				_ = pool.MonitorList()
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Sequential add+list 50 times in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})
	})

	Describe("MonitorGet Performance", func() {
		BeforeEach(func() {
			// Add monitors for retrieval
			for i := 0; i < 100; i++ {
				mon := createTestMonitor(fmt.Sprintf("get-perf-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should retrieve monitors quickly", func() {
			start := time.Now()

			for i := 0; i < 1000; i++ {
				idx := i % 100
				_ = pool.MonitorGet(fmt.Sprintf("get-perf-%d", idx))
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Retrieved monitors 1000 times in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 500*time.Millisecond))
		})
	})

	Describe("MonitorList Performance", func() {
		BeforeEach(func() {
			for i := 0; i < 100; i++ {
				mon := createTestMonitor(fmt.Sprintf("list-perf-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should list monitors efficiently", func() {
			start := time.Now()

			for i := 0; i < 100; i++ {
				list := pool.MonitorList()
				Expect(list).To(HaveLen(100))
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Listed 100 monitors 100 times in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 1*time.Second))
		})
	})

	Describe("MonitorWalk Performance", func() {
		BeforeEach(func() {
			for i := 0; i < 100; i++ {
				mon := createTestMonitor(fmt.Sprintf("walk-perf-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should walk through monitors efficiently", func() {
			start := time.Now()

			iterations := 50
			for i := 0; i < iterations; i++ {
				count := 0
				pool.MonitorWalk(func(name string, val montps.Monitor) bool {
					count++
					return true
				})
				Expect(count).To(Equal(100))
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Walked 100 monitors %d times in %v\n", iterations, elapsed)

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})
	})

	Describe("Lifecycle Performance", func() {
		BeforeEach(func() {
			for i := 0; i < 20; i++ {
				mon := createTestMonitor(fmt.Sprintf("lifecycle-perf-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}
		})

		It("should start monitors efficiently", func() {
			start := time.Now()
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			elapsed := time.Since(start)

			fmt.Fprintf(GinkgoWriter, "Started 20 monitors in %v\n", elapsed)

			time.Sleep(100 * time.Millisecond)
			Expect(pool.IsRunning()).To(BeTrue())

			// Startup should be fast
			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})

		It("should stop monitors efficiently", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			start := time.Now()
			Expect(pool.Stop(ctx)).ToNot(HaveOccurred())
			elapsed := time.Since(start)

			fmt.Fprintf(GinkgoWriter, "Stopped 20 monitors in %v\n", elapsed)

			time.Sleep(100 * time.Millisecond)
			Expect(pool.IsRunning()).To(BeFalse())

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})

		It("should restart monitors efficiently", func() {
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			start := time.Now()
			Expect(pool.Restart(ctx)).ToNot(HaveOccurred())
			elapsed := time.Since(start)

			fmt.Fprintf(GinkgoWriter, "Restarted 20 monitors in %v\n", elapsed)

			time.Sleep(100 * time.Millisecond)
			Expect(pool.IsRunning()).To(BeTrue())

			Expect(elapsed).To(BeNumerically("<", 3*time.Second))
		})
	})

	Describe("Encoding Performance", func() {
		BeforeEach(func() {
			for i := 0; i < 50; i++ {
				mon := createTestMonitor(fmt.Sprintf("encode-perf-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			time.Sleep(200 * time.Millisecond)
		})

		It("should marshal to text efficiently", func() {
			start := time.Now()

			for i := 0; i < 50; i++ {
				_, err := pool.MarshalText()
				Expect(err).ToNot(HaveOccurred())
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Marshaled to text 50 times with 50 monitors in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})

		It("should marshal to JSON efficiently", func() {
			start := time.Now()

			for i := 0; i < 50; i++ {
				_, err := pool.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Marshaled to JSON 50 times with 50 monitors in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})
	})

	Describe("Concurrent Operations Performance", func() {
		It("should handle concurrent reads efficiently", func() {
			// Add monitors
			for i := 0; i < 50; i++ {
				mon := createTestMonitor(fmt.Sprintf("concurrent-read-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			start := time.Now()

			done := make(chan bool, 100)

			// Concurrent reads
			for i := 0; i < 100; i++ {
				go func(index int) {
					defer GinkgoRecover()
					idx := index % 50
					_ = pool.MonitorGet(fmt.Sprintf("concurrent-read-%d", idx))
					done <- true
				}(i)
			}

			// Wait for all
			for i := 0; i < 100; i++ {
				<-done
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "100 concurrent reads in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 1*time.Second))
		})

		It("should handle concurrent writes efficiently", func() {
			start := time.Now()

			done := make(chan bool, 50)

			// Concurrent writes
			for i := 0; i < 50; i++ {
				go func(index int) {
					defer GinkgoRecover()
					mon := createTestMonitor(fmt.Sprintf("concurrent-write-%d", index), nil)
					defer mon.Stop(ctx)
					_ = pool.MonitorAdd(mon)
					done <- true
				}(i)
			}

			// Wait for all
			for i := 0; i < 50; i++ {
				<-done
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "50 concurrent writes in %v\n", elapsed)

			Expect(elapsed).To(BeNumerically("<", 2*time.Second))
		})
	})

	Describe("Scalability Tests", func() {
		It("should scale to large number of monitors", func() {
			start := time.Now()

			// Add many monitors
			numMonitors := 200
			for i := 0; i < numMonitors; i++ {
				mon := createTestMonitor(fmt.Sprintf("scale-%d", i), nil)
				defer mon.Stop(ctx)
				Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
			}

			addTime := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Added %d monitors in %v\n", numMonitors, addTime)

			// Start all
			start = time.Now()
			Expect(pool.Start(ctx)).ToNot(HaveOccurred())
			startTime := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Started %d monitors in %v\n", numMonitors, startTime)

			time.Sleep(200 * time.Millisecond)

			// List all
			start = time.Now()
			list := pool.MonitorList()
			listTime := time.Since(start)
			Expect(list).To(HaveLen(numMonitors))
			fmt.Fprintf(GinkgoWriter, "Listed %d monitors in %v\n", numMonitors, listTime)

			// Stop all
			start = time.Now()
			Expect(pool.Stop(ctx)).ToNot(HaveOccurred())
			stopTime := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "Stopped %d monitors in %v\n", numMonitors, stopTime)

			// All operations should scale reasonably
			Expect(addTime).To(BeNumerically("<", 5*time.Second))
			Expect(startTime).To(BeNumerically("<", 10*time.Second))
			Expect(listTime).To(BeNumerically("<", 500*time.Millisecond))
			Expect(stopTime).To(BeNumerically("<", 10*time.Second))
		})
	})

	Describe("Memory Efficiency", func() {
		It("should handle repeated add/remove cycles efficiently", func() {
			start := time.Now()

			cycles := 20
			monitorsPerCycle := 10

			for cycle := 0; cycle < cycles; cycle++ {
				// Add monitors
				for i := 0; i < monitorsPerCycle; i++ {
					mon := createTestMonitor(fmt.Sprintf("cycle-%d-%d", cycle, i), nil)
					defer mon.Stop(ctx)
					Expect(pool.MonitorAdd(mon)).ToNot(HaveOccurred())
				}

				// Remove monitors
				for i := 0; i < monitorsPerCycle; i++ {
					pool.MonitorDel(fmt.Sprintf("cycle-%d-%d", cycle, i))
				}
			}

			elapsed := time.Since(start)
			fmt.Fprintf(GinkgoWriter, "%d add/remove cycles in %v\n", cycles, elapsed)

			// Should handle cycles efficiently
			Expect(elapsed).To(BeNumerically("<", 5*time.Second))

			// Pool should be empty
			Expect(pool.MonitorList()).To(BeEmpty())
		})
	})
})
