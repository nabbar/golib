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

var _ = Describe("Benchmarks", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Write Performance", func() {
		It("should measure write throughput", func() {
			experiment := gmeasure.NewExperiment("Write Throughput")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			numWrites := 1000
			data := []byte("benchmark data")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("write_latency", func() {
					for i := 0; i < numWrites; i++ {
						_, _ = agg.Write(data)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 5 * time.Second})

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			// Report stats
			stats := experiment.GetStats("write_latency")
			Expect(stats).ToNot(BeNil())
			AddReportEntry("Average Write Latency", stats.DurationFor(gmeasure.StatMean))
			AddReportEntry("Max Write Latency", stats.DurationFor(gmeasure.StatMax))
			AddReportEntry("Min Write Latency", stats.DurationFor(gmeasure.StatMin))
		})

		It("should measure write throughput with different buffer sizes", func() {
			bufferSizes := []int{10, 100, 1000, 10000}
			numWrites := 500

			for _, bufSize := range bufferSizes {
				experiment := gmeasure.NewExperiment("Buffer Size: " + string(rune(bufSize)))
				AddReportEntry(experiment.Name, experiment)

				writer := newTestWriter()
				cfg := aggregator.Config{
					BufWriter: bufSize,
					FctWriter: writer.Write,
				}

				agg, err := aggregator.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = agg.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				data := []byte("test")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("write_time", func() {
						for i := 0; i < numWrites; i++ {
							_, _ = agg.Write(data)
						}
					})
				}, gmeasure.SamplingConfig{N: 5})

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())

				stats := experiment.GetStats("write_time")
				AddReportEntry("Avg Time (Buffer="+string(rune(bufSize))+")", stats.DurationFor(gmeasure.StatMean))
			}
		})
	})

	Describe("Concurrent Write Performance", func() {
		It("should measure concurrent write throughput", func() {
			experiment := gmeasure.NewExperiment("Concurrent Write Throughput")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 10
			writesPerGoroutine := 100
			data := []byte("test")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("concurrent_writes", func() {
					var wg sync.WaitGroup
					for i := 0; i < numGoroutines; i++ {
						wg.Add(1)
						go func() {
							defer wg.Done()
							for j := 0; j < writesPerGoroutine; j++ {
								_, _ = agg.Write(data)
							}
						}()
					}
					wg.Wait()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 5 * time.Second})

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			stats := experiment.GetStats("concurrent_writes")
			AddReportEntry("Average Concurrent Write Time", stats.DurationFor(gmeasure.StatMean))
		})

		It("should measure scalability with different goroutine counts", func() {
			goroutineCounts := []int{1, 5, 10, 20, 50}
			writesPerGoroutine := 100

			for _, numGoroutines := range goroutineCounts {
				experiment := gmeasure.NewExperiment("Goroutines: " + string(rune(numGoroutines)))
				AddReportEntry(experiment.Name, experiment)

				writer := newTestWriter()
				cfg := aggregator.Config{
					BufWriter: 1000,
					FctWriter: writer.Write,
				}

				agg, err := aggregator.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = agg.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				data := []byte("test")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("execution_time", func() {
						var wg sync.WaitGroup
						for i := 0; i < numGoroutines; i++ {
							wg.Add(1)
							go func() {
								defer wg.Done()
								for j := 0; j < writesPerGoroutine; j++ {
									_, _ = agg.Write(data)
								}
							}()
						}
						wg.Wait()
					})
				}, gmeasure.SamplingConfig{N: 5})

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())

				stats := experiment.GetStats("execution_time")
				AddReportEntry("Avg Time ("+string(rune(numGoroutines))+" goroutines)", stats.DurationFor(gmeasure.StatMean))
			}
		})
	})

	Describe("Start/Stop Performance", func() {
		It("should measure start time", func() {
			experiment := gmeasure.NewExperiment("Start Time")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			experiment.Sample(func(idx int) {
				agg, err := aggregator.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				experiment.MeasureDuration("start_time", func() {
					err = agg.Start(ctx)
					Expect(err).ToNot(HaveOccurred())
				})

				_ = agg.Close()
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

			stats := experiment.GetStats("start_time")
			AddReportEntry("Average Start Time", stats.DurationFor(gmeasure.StatMean))
		})

		It("should measure stop time", func() {
			experiment := gmeasure.NewExperiment("Stop Time")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			experiment.Sample(func(idx int) {
				agg, err := aggregator.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = agg.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				experiment.MeasureDuration("stop_time", func() {
					err = agg.Stop(ctx)
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

			stats := experiment.GetStats("stop_time")
			AddReportEntry("Average Stop Time", stats.DurationFor(gmeasure.StatMean))
		})

		It("should measure restart time", func() {
			experiment := gmeasure.NewExperiment("Restart Time")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("restart_time", func() {
					err = agg.Restart(ctx)
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 5 * time.Second})

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			stats := experiment.GetStats("restart_time")
			AddReportEntry("Average Restart Time", stats.DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Memory Performance", func() {
		It("should measure memory usage under load", func() {
			experiment := gmeasure.NewExperiment("Memory Usage")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := aggregator.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Write with different data sizes
			dataSizes := []int{10, 100, 1000, 10000}

			for _, size := range dataSizes {
				data := make([]byte, size)
				label := "data_size_" + string(rune(size))

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration(label, func() {
						for i := 0; i < 100; i++ {
							_, _ = agg.Write(data)
						}
					})
				}, gmeasure.SamplingConfig{N: 10})
			}

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Async/Sync Function Performance", func() {
		It("should measure sync function overhead", func() {
			experiment := gmeasure.NewExperiment("Sync Function Overhead")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			counter := newTestCounter()

			cfg := aggregator.Config{
				SyncTimer: 10 * time.Millisecond,
				SyncFct: func(ctx context.Context) {
					counter.Inc()
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			experiment.Sample(func(idx int) {
				counter.Reset()
				startTime := time.Now()

				experiment.MeasureDuration("time_for_10_calls", func() {
					// Wait for 10 sync calls
					Eventually(func() int {
						return counter.Get()
					}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", 10))
				})

				elapsed := time.Since(startTime)
				experiment.RecordDuration("total_time", elapsed)
			}, gmeasure.SamplingConfig{N: 5})

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			stats := experiment.GetStats("time_for_10_calls")
			AddReportEntry("Avg Time for 10 Sync Calls", stats.DurationFor(gmeasure.StatMean))
		})
	})

	Describe("End-to-End Performance", func() {
		It("should measure complete aggregation cycle", func() {
			experiment := gmeasure.NewExperiment("Complete Aggregation Cycle")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				writer := newTestWriter()
				cfg := aggregator.Config{
					BufWriter: 100,
					FctWriter: writer.Write,
				}

				var agg aggregator.Aggregator
				var err error

				experiment.MeasureDuration("total_cycle", func() {
					// Create
					agg, err = aggregator.New(ctx, cfg)
					Expect(err).ToNot(HaveOccurred())

					// Start
					err = agg.Start(ctx)
					Expect(err).ToNot(HaveOccurred())

					// Write data
					for i := 0; i < 100; i++ {
						_, _ = agg.Write([]byte("data"))
					}

					// Wait for processing
					Eventually(func() int32 {
						return writer.GetCallCount()
					}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))

					// Close
					err = agg.Close()
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 10 * time.Second})

			stats := experiment.GetStats("total_cycle")
			AddReportEntry("Average Complete Cycle Time", stats.DurationFor(gmeasure.StatMean))
			AddReportEntry("Max Complete Cycle Time", stats.DurationFor(gmeasure.StatMax))
		})
	})
})
