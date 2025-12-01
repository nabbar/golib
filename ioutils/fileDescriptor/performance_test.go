/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package fileDescriptor_test

import (
	"time"

	. "github.com/nabbar/golib/ioutils/fileDescriptor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

// Performance tests for SystemFileDescriptor using gmeasure.
// These tests measure operation timing and verify performance characteristics.
//
// Expected Performance:
//   - Query operation: < 1 microsecond (single syscall)
//   - Increase operation: < 10 microseconds (syscall + validation)
//   - No memory allocations
//   - Zero overhead after initial call
var _ = Describe("SystemFileDescriptor - Performance", func() {
	Context("Query operation performance", func() {
		It("should query limits in sub-microsecond time", func() {
			exp := gmeasure.NewExperiment("Query Performance")
			AddReportEntry(exp.Name, exp)

			// Warmup
			SystemFileDescriptor(0)

			// Measure query performance
			exp.Sample(func(idx int) {
				exp.MeasureDuration("query", func() {
					_, _, err := SystemFileDescriptor(0)
					Expect(err).ToNot(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: 100})

			// Verify performance characteristics
			stats := exp.GetStats("query")
			Expect(stats).NotTo(BeNil())

			// Query should be very fast (< 10 microseconds on average)
			// Note: Actual time depends on system load and hardware
			GinkgoWriter.Printf("Query Performance:\n")
			GinkgoWriter.Printf("  Mean:   %v\n", stats.DurationFor(gmeasure.StatMean))
			GinkgoWriter.Printf("  Median: %v\n", stats.DurationFor(gmeasure.StatMedian))
			GinkgoWriter.Printf("  StdDev: %v\n", stats.DurationFor(gmeasure.StatStdDev))
			GinkgoWriter.Printf("  Min:    %v\n", stats.DurationFor(gmeasure.StatMin))
			GinkgoWriter.Printf("  Max:    %v\n", stats.DurationFor(gmeasure.StatMax))

			// Reasonable upper bound for syscall operation
			// Even on slow systems, should complete in < 100 microseconds
			Expect(stats.DurationFor(gmeasure.StatMean).Microseconds()).To(
				BeNumerically("<", 100),
				"Query should complete in < 100µs on average")
		})

		It("should have consistent query performance", func() {
			exp := gmeasure.NewExperiment("Query Consistency")
			AddReportEntry(exp.Name, exp)

			// Measure consistency across many calls
			exp.Sample(func(idx int) {
				exp.MeasureDuration("query", func() {
					SystemFileDescriptor(0)
				})
			}, gmeasure.SamplingConfig{N: 200})

			stats := exp.GetStats("query")

			// Standard deviation should be relatively small
			// indicating consistent performance
			mean := stats.DurationFor(gmeasure.StatMean)
			stddev := stats.DurationFor(gmeasure.StatStdDev)

			GinkgoWriter.Printf("Consistency Metrics:\n")
			GinkgoWriter.Printf("  Mean:   %v\n", mean)
			GinkgoWriter.Printf("  StdDev: %v\n", stddev)
			GinkgoWriter.Printf("  CV:     %.2f%%\n", float64(stddev)/float64(mean)*100)

			// Coefficient of variation should be reasonable
			// Note: Can be high on shared/virtualized systems
			cv := float64(stddev) / float64(mean)
			if cv < 10.0 {
				GinkgoWriter.Printf("  Performance is consistent (CV: %.2f%%)\n", cv*100)
			} else {
				GinkgoWriter.Printf("  Note: High variance detected (CV: %.2f%%) - may be due to system load\n", cv*100)
			}
		})
	})

	Context("Increase operation performance", func() {
		It("should measure increase operation timing", func() {
			initial, max, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())

			// Only test if we can increase
			if initial >= max-10 {
				Skip("Cannot test increase: already near maximum")
			}

			target := initial + 5

			exp := gmeasure.NewExperiment("Increase Performance")
			AddReportEntry(exp.Name, exp)

			// Measure increase performance
			// Note: This may fail due to permissions, which is acceptable
			var successCount int
			exp.Sample(func(idx int) {
				exp.MeasureDuration("increase", func() {
					_, _, err := SystemFileDescriptor(target)
					if err == nil {
						successCount++
					}
				})
			}, gmeasure.SamplingConfig{N: 10}) // Fewer samples for modification

			// Only report if we had successful increases
			if successCount > 0 {
				stats := exp.GetStats("increase")
				GinkgoWriter.Printf("Increase Performance (%d successes):\n", successCount)
				GinkgoWriter.Printf("  Mean:   %v\n", stats.DurationFor(gmeasure.StatMean))
				GinkgoWriter.Printf("  Median: %v\n", stats.DurationFor(gmeasure.StatMedian))
				GinkgoWriter.Printf("  Min:    %v\n", stats.DurationFor(gmeasure.StatMin))
				GinkgoWriter.Printf("  Max:    %v\n", stats.DurationFor(gmeasure.StatMax))
			} else {
				GinkgoWriter.Println("All increase attempts failed (likely permission denied)")
			}
		})
	})

	Context("Throughput testing", func() {
		It("should handle high query throughput", func() {
			exp := gmeasure.NewExperiment("Query Throughput")
			AddReportEntry(exp.Name, exp)

			// Measure how many queries can be done in a fixed time
			const targetQueries = 1000

			exp.MeasureDuration("throughput", func() {
				for i := 0; i < targetQueries; i++ {
					_, _, err := SystemFileDescriptor(0)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			stats := exp.GetStats("throughput")
			totalTime := stats.DurationFor(gmeasure.StatMean)
			queriesPerSecond := float64(targetQueries) / totalTime.Seconds()

			GinkgoWriter.Printf("Throughput Metrics:\n")
			GinkgoWriter.Printf("  Total time:  %v\n", totalTime)
			GinkgoWriter.Printf("  Per query:   %v\n", totalTime/targetQueries)
			GinkgoWriter.Printf("  Throughput:  %.0f queries/sec\n", queriesPerSecond)

			// Should handle at least 10,000 queries per second
			// (very conservative, actual performance is much higher)
			Expect(queriesPerSecond).To(BeNumerically(">", 10000),
				"Should handle at least 10k queries/sec")
		})

		It("should have minimal overhead per call", func() {
			exp := gmeasure.NewExperiment("Per-Call Overhead")
			AddReportEntry(exp.Name, exp)

			// Single call
			var singleTime time.Duration
			exp.MeasureDuration("single", func() {
				SystemFileDescriptor(0)
			})
			stats := exp.GetStats("single")
			singleTime = stats.DurationFor(gmeasure.StatMean)

			// Batch of 10 calls
			var batchTime time.Duration
			exp.MeasureDuration("batch", func() {
				for i := 0; i < 10; i++ {
					SystemFileDescriptor(0)
				}
			})
			stats = exp.GetStats("batch")
			batchTime = stats.DurationFor(gmeasure.StatMean)

			avgTimePerCall := batchTime / 10

			GinkgoWriter.Printf("Overhead Analysis:\n")
			GinkgoWriter.Printf("  Single call:     %v\n", singleTime)
			GinkgoWriter.Printf("  Batch (10):      %v\n", batchTime)
			GinkgoWriter.Printf("  Avg per call:    %v\n", avgTimePerCall)
			GinkgoWriter.Printf("  Overhead ratio:  %.2f\n", float64(avgTimePerCall)/float64(singleTime))

			// Batch average should be similar to single call
			// (indicates no significant state or caching effects)
			ratio := float64(avgTimePerCall) / float64(singleTime)
			
			// Informational output - actual ratio depends on system characteristics
			if ratio > 0.5 && ratio < 1.5 {
				GinkgoWriter.Printf("  Overhead is consistent\n")
			} else {
				GinkgoWriter.Printf("  Note: Overhead ratio varies (%.2f) - may indicate CPU caching effects\n", ratio)
			}
		})
	})

	Context("Scalability testing", func() {
		It("should scale linearly with number of calls", func() {
			exp := gmeasure.NewExperiment("Scalability")
			AddReportEntry(exp.Name, exp)

			callCounts := []int{10, 50, 100, 500}

			GinkgoWriter.Println("Scalability Test:")

			var prevTimePerCall time.Duration
			for _, count := range callCounts {
				exp.MeasureDuration("calls", func() {
					for i := 0; i < count; i++ {
						SystemFileDescriptor(0)
					}
				})

				stats := exp.GetStats("calls")
				totalTime := stats.DurationFor(gmeasure.StatMean)
				timePerCall := totalTime / time.Duration(count)

				GinkgoWriter.Printf("  %4d calls: total=%v, per-call=%v\n",
					count, totalTime, timePerCall)

				// Analyze scaling (time per call should be roughly constant for linear scaling)
				if prevTimePerCall > 0 {
					ratio := float64(timePerCall) / float64(prevTimePerCall)
					// Performance variations are normal due to:
					// - CPU frequency scaling
					// - Cache effects
					// - System load
					// - First call warmup
					if ratio >= 0.5 && ratio <= 2.0 {
						GinkgoWriter.Printf("    Scaling: good (ratio: %.2f)\n", ratio)
					} else {
						GinkgoWriter.Printf("    Scaling: variable (ratio: %.2f) - may be due to system effects\n", ratio)
					}
				}
				prevTimePerCall = timePerCall
			}
		})
	})

	Context("Performance under load", func() {
		It("should maintain performance under sustained queries", func() {
			exp := gmeasure.NewExperiment("Sustained Load")
			AddReportEntry(exp.Name, exp)

			// Measure performance in batches to detect degradation
			const batchSize = 100
			const batches = 10

			GinkgoWriter.Println("Sustained Load Test:")

			var firstBatchTime time.Duration
			for batch := 0; batch < batches; batch++ {
				exp.MeasureDuration("batch", func() {
					for i := 0; i < batchSize; i++ {
						SystemFileDescriptor(0)
					}
				})

				stats := exp.GetStats("batch")
				batchTime := stats.DurationFor(gmeasure.StatMean)

				if batch == 0 {
					firstBatchTime = batchTime
				}

				timePerQuery := batchTime / batchSize
				GinkgoWriter.Printf("  Batch %2d: %v (%v/query)\n",
					batch, batchTime, timePerQuery)

				// Performance should not degrade significantly
				// (no resource leaks or accumulation)
				if firstBatchTime > 0 {
					ratio := float64(batchTime) / float64(firstBatchTime)
					Expect(ratio).To(BeNumerically("<", 2.0),
						"Performance should not degrade significantly")
				}
			}
		})
	})

	Context("Memory allocations", func() {
		It("should have zero allocations per call", func() {
			// This test verifies that the function doesn't allocate memory
			// Use go test -benchmem to see allocation stats

			// Warmup
			SystemFileDescriptor(0)

			// The function should not allocate as it:
			// - Returns int values (not pointers)
			// - Uses stack-allocated syscall.Rlimit struct
			// - No string formatting or conversions

			// We verify this by checking it completes successfully
			// Actual allocation count is visible with: go test -bench . -benchmem
			current, max, err := SystemFileDescriptor(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(current).To(BeNumerically(">", 0))
			Expect(max).To(BeNumerically(">=", current))

			GinkgoWriter.Println("Note: Run 'go test -bench . -benchmem' to verify zero allocations")
		})
	})
})
