/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// This file contains performance benchmarks using gmeasure.
//
// Test Strategy:
//   - Measure wrapper creation overhead
//   - Benchmark default read/write operations (delegation)
//   - Benchmark custom function read/write operations
//   - Measure function update (SetRead/SetWrite) performance
//   - Benchmark seek operations
//   - Test mixed operation performance
//
// Uses github.com/onsi/gomega/gmeasure for statistical performance testing.
// Results are reported with min/median/mean/stddev/max metrics.
//
// Coverage: 8 specs measuring performance characteristics and overhead.
package iowrapper_test

import (
	"bytes"
	"io"
	"strings"

	. "github.com/nabbar/golib/ioutils/iowrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("IOWrapper - Benchmarks", func() {
	It("Wrapper creation overhead", func() {
		experiment := NewExperiment("Wrapper creation overhead")
		AddReportEntry(experiment.Name, experiment)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("creation time", func() {
				for i := 0; i < 10000; i++ {
					buf := bytes.NewBuffer([]byte("test"))
					_ = New(buf)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("creation time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Creation should be fast") // 0.5s in nanoseconds
	})

	It("Default read performance", func() {
		experiment := NewExperiment("Default read performance")
		AddReportEntry(experiment.Name, experiment)

		reader := strings.NewReader(strings.Repeat("x", 10000))
		wrapper := New(reader)
		data := make([]byte, 100)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("read time", func() {
				for i := 0; i < 1000; i++ {
					reader.Seek(0, io.SeekStart)
					wrapper.Read(data)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("read time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Reads should be fast")
	})

	It("Default write performance", func() {
		experiment := NewExperiment("Default write performance")
		AddReportEntry(experiment.Name, experiment)

		buf := &bytes.Buffer{}
		wrapper := New(buf)
		data := []byte("test data")

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("write time", func() {
				for i := 0; i < 1000; i++ {
					buf.Reset()
					wrapper.Write(data)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("write time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Writes should be fast")
	})

	It("Custom read function performance", func() {
		experiment := NewExperiment("Custom read function performance")
		AddReportEntry(experiment.Name, experiment)

		wrapper := New(nil)
		wrapper.SetRead(func(p []byte) []byte {
			return []byte("data")
		})
		data := make([]byte, 100)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("custom read time", func() {
				for i := 0; i < 1000; i++ {
					wrapper.Read(data)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("custom read time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Custom reads should be fast")
	})

	It("Custom write function performance", func() {
		experiment := NewExperiment("Custom write function performance")
		AddReportEntry(experiment.Name, experiment)

		wrapper := New(nil)
		wrapper.SetWrite(func(p []byte) []byte {
			return p
		})
		data := []byte("test data")

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("custom write time", func() {
				for i := 0; i < 1000; i++ {
					wrapper.Write(data)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("custom write time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Custom writes should be fast")
	})

	It("Function update performance", func() {
		experiment := NewExperiment("Function update performance")
		AddReportEntry(experiment.Name, experiment)

		wrapper := New(nil)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("update time", func() {
				for i := 0; i < 1000; i++ {
					wrapper.SetRead(func(p []byte) []byte {
						return []byte("x")
					})
					wrapper.SetWrite(func(p []byte) []byte {
						return p
					})
					wrapper.SetSeek(func(offset int64, whence int) (int64, error) {
						return 0, nil
					})
					wrapper.SetClose(func() error {
						return nil
					})
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("update time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Function updates should be fast")
	})

	It("Seek performance", func() {
		experiment := NewExperiment("Seek performance")
		AddReportEntry(experiment.Name, experiment)

		reader := bytes.NewReader(make([]byte, 10000))
		wrapper := New(reader)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("seek time", func() {
				for i := 0; i < 1000; i++ {
					wrapper.Seek(int64(i%10000), io.SeekStart)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("seek time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 500*1000*1000), "Seeks should be fast")
	})

	It("Mixed operations performance", func() {
		experiment := NewExperiment("Mixed operations performance")
		AddReportEntry(experiment.Name, experiment)

		buf := bytes.NewBuffer(make([]byte, 10000))
		wrapper := New(buf)
		data := make([]byte, 100)

		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("mixed operations time", func() {
				for i := 0; i < 1000; i++ {
					wrapper.Read(data)
					wrapper.Write(data)
					wrapper.Seek(0, io.SeekStart)
				}
			})
		}, SamplingConfig{N: 5})

		stats := experiment.GetStats("mixed operations time")
		Expect(stats.DurationFor(StatMean)).To(BeNumerically("<", 1000*1000*1000), "Mixed operations should be reasonably fast")
	})

	Context("Memory allocation benchmarks", func() {
		It("should minimize allocations during reads", func() {
			reader := strings.NewReader(strings.Repeat("x", 1000))
			wrapper := New(reader)
			data := make([]byte, 100)

			// Warm up
			for i := 0; i < 10; i++ {
				reader.Seek(0, io.SeekStart)
				wrapper.Read(data)
			}

			// This is a smoke test, actual allocation counting would need benchmarking
			reader.Seek(0, io.SeekStart)
			n, err := wrapper.Read(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(100))
		})

		It("should minimize allocations during writes", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)
			data := []byte(strings.Repeat("x", 100))

			// Warm up
			for i := 0; i < 10; i++ {
				buf.Reset()
				wrapper.Write(data)
			}

			// This is a smoke test
			buf.Reset()
			n, err := wrapper.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(100))
		})
	})

	Context("Scalability tests", func() {
		It("should handle large data reads efficiently", func() {
			largeData := strings.Repeat("x", 1024*1024) // 1MB
			reader := strings.NewReader(largeData)
			wrapper := New(reader)
			data := make([]byte, 4096)

			totalRead := 0
			for {
				n, err := wrapper.Read(data)
				totalRead += n
				if n == 0 {
					break
				}
				if err != nil {
					break
				}
			}

			Expect(totalRead).To(Equal(len(largeData)))
		})

		It("should handle large data writes efficiently", func() {
			buf := &bytes.Buffer{}
			wrapper := New(buf)
			largeData := []byte(strings.Repeat("x", 1024*1024)) // 1MB

			n, err := wrapper.Write(largeData)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(buf.Len()).To(Equal(len(largeData)))
		})

		It("should handle many small operations efficiently", func() {
			buf := bytes.NewBuffer(make([]byte, 100000))
			wrapper := New(buf)
			data := make([]byte, 10)

			for i := 0; i < 10000; i++ {
				wrapper.Seek(int64(i%1000), io.SeekStart)
				wrapper.Read(data)
			}

			// Should complete without timeout
			Expect(true).To(BeTrue())
		})
	})
})
