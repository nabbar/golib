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
 *
 */

package multi_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/ioutils/multi"
)

// Performance benchmarks for Multi operations.
// These benchmarks measure the performance and memory allocation characteristics
// of various operations including construction, writes, reads, copies, and
// writer management. Uses gmeasure for statistical analysis.
var _ = Describe("Multi Performance Benchmarks", func() {
	Describe("Constructor benchmarks", func() {
		It("should benchmark New() creation", func() {
			experiment := gmeasure.NewExperiment("New() Constructor")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("creation", func() {
					_ = multi.New()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("creation").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 100*time.Microsecond), "New() should be fast")
		})
	})

	Describe("Write benchmarks", func() {
		It("should benchmark Write to single writer", func() {
			experiment := gmeasure.NewExperiment("Write Single")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)
			data := []byte("test data")

			experiment.Sample(func(idx int) {
				buf.Reset()
				experiment.MeasureDuration("write", func() {
					m.Write(data)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("write").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should benchmark Write to multiple writers", func() {
			experiment := gmeasure.NewExperiment("Write Multiple")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf1, buf2, buf3 bytes.Buffer
			m.AddWriter(&buf1, &buf2, &buf3)
			data := []byte("broadcast data")

			experiment.Sample(func(idx int) {
				buf1.Reset()
				buf2.Reset()
				buf3.Reset()
				experiment.MeasureDuration("write", func() {
					m.Write(data)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("write").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 50*time.Microsecond))
		})

		It("should benchmark WriteString", func() {
			experiment := gmeasure.NewExperiment("WriteString")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)
			str := "test string"

			experiment.Sample(func(idx int) {
				buf.Reset()
				experiment.MeasureDuration("write-string", func() {
					m.WriteString(str)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("write-string").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should benchmark large Write operations", func() {
			experiment := gmeasure.NewExperiment("Write Large")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)
			largeData := make([]byte, 1024*1024) // 1MB

			experiment.Sample(func(idx int) {
				buf.Reset()
				experiment.MeasureDuration("write-large", func() {
					m.Write(largeData)
				})
			}, gmeasure.SamplingConfig{N: 100})

			Expect(experiment.GetStats("write-large").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Read benchmarks", func() {
		It("should benchmark Read operations", func() {
			experiment := gmeasure.NewExperiment("Read")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			buf := make([]byte, 1024)

			experiment.Sample(func(idx int) {
				input := io.NopCloser(strings.NewReader(strings.Repeat("x", 1024)))
				m.SetInput(input)

				experiment.MeasureDuration("read", func() {
					m.Read(buf)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("read").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should benchmark large Read operations", func() {
			experiment := gmeasure.NewExperiment("Read Large")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			buf := make([]byte, 1024*1024) // 1MB

			experiment.Sample(func(idx int) {
				largeData := strings.Repeat("x", 1024*1024)
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				experiment.MeasureDuration("read-large", func() {
					m.Read(buf)
				})
			}, gmeasure.SamplingConfig{N: 100})

			Expect(experiment.GetStats("read-large").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("Copy benchmarks", func() {
		It("should benchmark Copy to single writer", func() {
			experiment := gmeasure.NewExperiment("Copy Single")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)

			experiment.Sample(func(idx int) {
				buf.Reset()
				input := io.NopCloser(strings.NewReader("test data for copy"))
				m.SetInput(input)

				experiment.MeasureDuration("copy", func() {
					m.Copy()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("copy").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 50*time.Microsecond))
		})

		It("should benchmark Copy to multiple writers", func() {
			experiment := gmeasure.NewExperiment("Copy Multiple")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf1, buf2, buf3 bytes.Buffer
			m.AddWriter(&buf1, &buf2, &buf3)

			experiment.Sample(func(idx int) {
				buf1.Reset()
				buf2.Reset()
				buf3.Reset()
				input := io.NopCloser(strings.NewReader("test data for copy"))
				m.SetInput(input)

				experiment.MeasureDuration("copy", func() {
					m.Copy()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("copy").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should benchmark large Copy operations", func() {
			experiment := gmeasure.NewExperiment("Copy Large")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)

			experiment.Sample(func(idx int) {
				buf.Reset()
				largeData := strings.Repeat("x", 1024*1024) // 1MB
				input := io.NopCloser(strings.NewReader(largeData))
				m.SetInput(input)

				experiment.MeasureDuration("copy-large", func() {
					m.Copy()
				})
			}, gmeasure.SamplingConfig{N: 100})

			Expect(experiment.GetStats("copy-large").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 50*time.Millisecond))
		})
	})

	Describe("AddWriter benchmarks", func() {
		It("should benchmark AddWriter operations", func() {
			experiment := gmeasure.NewExperiment("AddWriter")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				m := multi.New()
				var buf bytes.Buffer

				experiment.MeasureDuration("add-writer", func() {
					m.AddWriter(&buf)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("add-writer").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 50*time.Microsecond))
		})

		It("should benchmark AddWriter with multiple writers", func() {
			experiment := gmeasure.NewExperiment("AddWriter Multiple")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				m := multi.New()
				var buf1, buf2, buf3, buf4, buf5 bytes.Buffer

				experiment.MeasureDuration("add-multiple", func() {
					m.AddWriter(&buf1, &buf2, &buf3, &buf4, &buf5)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("add-multiple").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 100*time.Microsecond))
		})
	})

	Describe("Clean benchmarks", func() {
		It("should benchmark Clean operations", func() {
			experiment := gmeasure.NewExperiment("Clean")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				m := multi.New()
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)

				experiment.MeasureDuration("clean", func() {
					m.Clean()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("clean").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 50*time.Microsecond))
		})
	})

	Describe("SetInput benchmarks", func() {
		It("should benchmark SetInput operations", func() {
			experiment := gmeasure.NewExperiment("SetInput")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()

			experiment.Sample(func(idx int) {
				input := io.NopCloser(strings.NewReader("data"))
				experiment.MeasureDuration("set-input", func() {
					m.SetInput(input)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			Expect(experiment.GetStats("set-input").DurationFor(gmeasure.StatMean)).
				To(BeNumerically("<", 10*time.Microsecond))
		})
	})

	Describe("Memory allocation benchmarks", func() {
		It("should measure Write allocations", func() {
			experiment := gmeasure.NewExperiment("Write Allocations")
			AddReportEntry(experiment.Name, experiment)

			m := multi.New()
			var buf bytes.Buffer
			m.AddWriter(&buf)
			data := []byte("test")

			experiment.Sample(func(idx int) {
				buf.Reset()
				experiment.RecordValue("allocations", float64(testing.AllocsPerRun(100, func() {
					m.Write(data)
				})))
			}, gmeasure.SamplingConfig{N: 10})

			// Write should have minimal allocations
			Expect(experiment.GetStats("allocations").FloatFor(gmeasure.StatMean)).
				To(BeNumerically("<", 5))
		})
	})
})
