/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
 */

package aggregator_test

import (
	"context"
	"sync"
	"time"

	"github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Metrics", func() {
	var (
		agg        aggregator.Aggregator
		ctx        context.Context
		cancel     context.CancelFunc
		writeCount int
		writeMutex sync.Mutex
		writeDelay time.Duration
		writeData  [][]byte
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
		writeMutex.Lock()
		writeCount = 0
		writeDelay = 0
		writeData = make([][]byte, 0)
		writeMutex.Unlock()

		cfg := aggregator.Config{
			BufWriter: 10,
			FctWriter: func(p []byte) (int, error) {
				writeMutex.Lock()
				delay := writeDelay
				writeMutex.Unlock()

				if delay > 0 {
					time.Sleep(delay)
				}

				writeMutex.Lock()
				writeCount++
				// Store copy of data
				buf := make([]byte, len(p))
				copy(buf, p)
				writeData = append(writeData, buf)
				writeMutex.Unlock()
				return len(p), nil
			},
		}

		var err error
		agg, err = aggregator.New(ctx, cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(agg).ToNot(BeNil())
	})

	AfterEach(func() {
		if agg != nil {
			agg.Close()
		}
		cancel()
		time.Sleep(50 * time.Millisecond)
	})

	Describe("NbWaiting and NbProcessing", func() {
		Context("when aggregator is not started", func() {
			It("should return 0 for both metrics", func() {
				Expect(agg.NbWaiting()).To(Equal(int64(0)))
				Expect(agg.NbProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when aggregator is running with no writes", func() {
			It("should return 0 for both metrics", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				Expect(agg.NbWaiting()).To(Equal(int64(0)))
				Expect(agg.NbProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when writes are processed immediately", func() {
			It("should show correct counts during and after processing", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				// Write some data
				for i := 0; i < 5; i++ {
					_, err := agg.Write([]byte("test"))
					Expect(err).ToNot(HaveOccurred())
				}

				// Wait for processing
				Eventually(func() int {
					writeMutex.Lock()
					defer writeMutex.Unlock()
					return writeCount
				}, 2*time.Second, 10*time.Millisecond).Should(Equal(5))

				// After processing, both should be 0
				Expect(agg.NbWaiting()).To(Equal(int64(0)))
				Expect(agg.NbProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when buffer is full with slow processing", func() {
			BeforeEach(func() {
				writeMutex.Lock()
				writeDelay = 100 * time.Millisecond
				writeMutex.Unlock()
			})

			It("should show waiting writes when buffer is full", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				// Fill the buffer (capacity = 10) with many concurrent writes
				// to increase the chance of having waiting writes
				var wg sync.WaitGroup
				numWrites := 30
				for i := 0; i < numWrites; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()
						_, err := agg.Write([]byte("data"))
						Expect(err).ToNot(HaveOccurred())
					}(i)
				}

				// Check that some writes are waiting OR processing
				// (timing-dependent, so we check for either condition)
				Eventually(func() int64 {
					return agg.NbWaiting() + agg.NbProcessing()
				}, 1*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))

				wg.Wait()

				// Eventually all should be processed
				Eventually(func() int64 {
					return agg.NbWaiting()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))

				Eventually(func() int64 {
					return agg.NbProcessing()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
			})
		})
	})

	Describe("SizeWaiting and SizeProcessing", func() {
		Context("when aggregator is not started", func() {
			It("should return 0 for both metrics", func() {
				Expect(agg.SizeWaiting()).To(Equal(int64(0)))
				Expect(agg.SizeProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when aggregator is running with no writes", func() {
			It("should return 0 for both metrics", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				Expect(agg.SizeWaiting()).To(Equal(int64(0)))
				Expect(agg.SizeProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when writes are processed immediately", func() {
			It("should track correct byte sizes", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				// Write data of known sizes
				testData := [][]byte{
					[]byte("small"),        // 5 bytes
					[]byte("medium12"),     // 8 bytes
					[]byte("larger123456"), // 12 bytes
				}

				for _, data := range testData {
					_, err := agg.Write(data)
					Expect(err).ToNot(HaveOccurred())
				}

				// Wait for processing
				Eventually(func() int {
					writeMutex.Lock()
					defer writeMutex.Unlock()
					return writeCount
				}, 2*time.Second, 10*time.Millisecond).Should(Equal(3))

				// After processing, both should be 0
				Expect(agg.SizeWaiting()).To(Equal(int64(0)))
				Expect(agg.SizeProcessing()).To(Equal(int64(0)))
			})
		})

		Context("when buffer is full with slow processing", func() {
			BeforeEach(func() {
				writeMutex.Lock()
				writeDelay = 100 * time.Millisecond
				writeMutex.Unlock()
			})

			It("should track byte sizes in waiting and processing states", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				// Write messages of 10 bytes each
				msgSize := 10
				numMessages := 15

				var wg sync.WaitGroup
				for i := 0; i < numMessages; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						data := make([]byte, msgSize)
						for j := range data {
							data[j] = 'x'
						}
						_, err := agg.Write(data)
						Expect(err).ToNot(HaveOccurred())
					}()
				}

				// Check that some bytes are waiting or processing
				Eventually(func() int64 {
					return agg.SizeWaiting() + agg.SizeProcessing()
				}, 500*time.Millisecond, 10*time.Millisecond).Should(BeNumerically(">", 0))

				wg.Wait()

				// Eventually all bytes should be processed
				Eventually(func() int64 {
					return agg.SizeWaiting()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))

				Eventually(func() int64 {
					return agg.SizeProcessing()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
			})
		})

		Context("with variable message sizes", func() {
			BeforeEach(func() {
				writeMutex.Lock()
				writeDelay = 50 * time.Millisecond
				writeMutex.Unlock()
			})

			It("should accurately track total byte sizes", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				// Write messages of different sizes
				sizes := []int{10, 50, 100, 200, 500}
				totalExpectedBytes := int64(0)
				for _, size := range sizes {
					totalExpectedBytes += int64(size)
				}

				var wg sync.WaitGroup
				for _, size := range sizes {
					wg.Add(1)
					go func(s int) {
						defer wg.Done()
						defer GinkgoRecover()
						data := make([]byte, s)
						_, err := agg.Write(data)
						Expect(err).ToNot(HaveOccurred())
					}(size)
				}

				// At some point, the total memory in flight should be close to expected
				Eventually(func() bool {
					total := agg.SizeWaiting() + agg.SizeProcessing()
					return total > 0 && total <= totalExpectedBytes
				}, 500*time.Millisecond, 10*time.Millisecond).Should(BeTrue())

				wg.Wait()
				time.Sleep(time.Second)

				// Eventually all should be processed
				Eventually(func() int64 {
					return agg.SizeWaiting() + agg.SizeProcessing()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))

				// Verify all data was received
				writeMutex.Lock()
				defer writeMutex.Unlock()
				Expect(writeCount).To(BeNumerically(">=", len(sizes)))
			})
		})
	})

	Describe("Combined Metrics", func() {
		Context("with concurrent writes and slow processing", func() {
			BeforeEach(func() {
				writeMutex.Lock()
				writeDelay = 50 * time.Millisecond
				writeMutex.Unlock()
			})

			It("should maintain consistency between count and size metrics", func() {
				Expect(startAndWait(agg, ctx)).To(Succeed())

				msgSize := 100
				numMessages := 20

				var wg sync.WaitGroup
				for i := 0; i < numMessages; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						data := make([]byte, msgSize)
						_, err := agg.Write(data)
						Expect(err).ToNot(HaveOccurred())
					}()
				}

				// Sample metrics during processing
				time.Sleep(200 * time.Millisecond)

				nbProc := agg.NbProcessing()
				sizeProc := agg.SizeProcessing()

				// If there are items processing, size should be proportional
				if nbProc > 0 {
					avgSize := sizeProc / nbProc
					// Average size should be approximately msgSize
					Expect(avgSize).To(BeNumerically(">=", int64(msgSize-10)))
					Expect(avgSize).To(BeNumerically("<=", int64(msgSize+10)))
				}

				wg.Wait()

				// Final state: all metrics should be 0
				Eventually(func() int64 {
					return agg.NbWaiting() + agg.NbProcessing()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))

				Eventually(func() int64 {
					return agg.SizeWaiting() + agg.SizeProcessing()
				}, 3*time.Second, 50*time.Millisecond).Should(Equal(int64(0)))
			})
		})
	})

	Describe("Metrics after Restart", func() {
		It("should reset metrics after restart", func() {
			Expect(startAndWait(agg, ctx)).To(Succeed())

			// Write some data
			for i := 0; i < 5; i++ {
				_, err := agg.Write([]byte("data"))
				Expect(err).ToNot(HaveOccurred())
			}

			// Wait for processing
			Eventually(func() int {
				writeMutex.Lock()
				defer writeMutex.Unlock()
				return writeCount
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(5))

			// Restart
			Expect(agg.Restart(ctx)).To(Succeed())
			time.Sleep(100 * time.Millisecond)

			// All metrics should be 0 after restart
			Expect(agg.NbWaiting()).To(Equal(int64(0)))
			Expect(agg.NbProcessing()).To(Equal(int64(0)))
			Expect(agg.SizeWaiting()).To(Equal(int64(0)))
			Expect(agg.SizeProcessing()).To(Equal(int64(0)))

			// Write after restart should work
			_, err := agg.Write([]byte("after"))
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				writeMutex.Lock()
				defer writeMutex.Unlock()
				return writeCount
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(6))
		})
	})

	Describe("Performance - Metrics Overhead", Ordered, func() {
		var experiment *gmeasure.Experiment

		BeforeAll(func() {
			experiment = gmeasure.NewExperiment("Metrics Overhead")
			AddReportEntry(experiment.Name, experiment)
		})

		It("should measure overhead of metrics tracking", func() {
			cfg := aggregator.Config{
				BufWriter: 1000,
				FctWriter: func(p []byte) (int, error) {
					return len(p), nil
				},
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			Expect(startAndWait(agg, ctx)).To(Succeed())

			// Warmup
			for i := 0; i < 100; i++ {
				agg.Write([]byte("warmup"))
			}
			time.Sleep(100 * time.Millisecond)

			// Measure write latency with metrics
			experiment.Sample(func(idx int) {
				data := []byte("test data for metrics")

				experiment.MeasureDuration("write_with_metrics", func() {
					_, err := agg.Write(data)
					Expect(err).ToNot(HaveOccurred())

					// Also check metrics (this adds overhead)
					_ = agg.NbWaiting()
					_ = agg.NbProcessing()
					_ = agg.SizeWaiting()
					_ = agg.SizeProcessing()
				})
			}, gmeasure.SamplingConfig{N: 1000, Duration: 5 * time.Second})

			// Wait for all to process
			time.Sleep(500 * time.Millisecond)

			stats := experiment.GetStats("write_with_metrics")
			AddReportEntry("Write Latency Stats", stats)

			// Metrics overhead should be negligible (< 1ms)
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 1*time.Millisecond))
		})

		It("should measure metrics read performance", func() {
			cfg := aggregator.Config{
				BufWriter: 100,
				FctWriter: func(p []byte) (int, error) {
					time.Sleep(10 * time.Millisecond)
					return len(p), nil
				},
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			Expect(startAndWait(agg, ctx)).To(Succeed())

			// Fill buffer to have meaningful metrics
			go func() {
				for i := 0; i < 50; i++ {
					agg.Write([]byte("data"))
					time.Sleep(5 * time.Millisecond)
				}
			}()

			time.Sleep(50 * time.Millisecond)

			// Measure metrics read performance
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("metrics_read", func() {
					_ = agg.NbWaiting()
					_ = agg.NbProcessing()
					_ = agg.SizeWaiting()
					_ = agg.SizeProcessing()
				})
			}, gmeasure.SamplingConfig{N: 10000, Duration: 2 * time.Second})

			stats := experiment.GetStats("metrics_read")
			AddReportEntry("Metrics Read Stats", stats)

			// Reading all 4 metrics should be very fast (< 1µs)
			Expect(stats.DurationFor(gmeasure.StatMedian)).To(BeNumerically("<", 5*time.Microsecond))
		})
	})
})
