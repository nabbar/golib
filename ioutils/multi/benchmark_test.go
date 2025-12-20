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

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"

	"github.com/nabbar/golib/ioutils/multi"
)

// Performance benchmarks for Multi operations.
// These benchmarks measure the performance and memory allocation characteristics
// of various operations including construction, writes, reads, copies, and
// writer management. Uses gmeasure for statistical analysis.
//
// Benchmarks are organized following patterns from ioutils/delim:
//   - Aggregated experiments grouping related variations
//   - Systematic variations (data sizes, writer counts, modes)
//   - Real-world scenario testing
//   - Statistical analysis with gmeasure
//
// Run with: go test -v to see performance reports.
var _ = Describe("[TC-BC] Multi Performance Benchmarks", func() {
	Describe("Write operations", func() {
		It("[TC-BC-001] should benchmark Write with varying writer counts and data sizes", func() {
			experiment := gmeasure.NewExperiment("Write operations")
			AddReportEntry(experiment.Name, experiment)

			// Small data (10 bytes)
			smallData := []byte("test data")

			// Medium data (1KB)
			mediumData := make([]byte, 1024)

			// Large data (1MB)
			largeData := make([]byte, 1024*1024)

			experiment.SampleDuration("Single writer, small data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.Write(smallData)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("3 writers, small data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.Write(smallData)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Single writer, 1KB data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.Write(mediumData)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("3 writers, 1KB data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.Write(mediumData)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Single writer, 1MB data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.Write(largeData)
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("3 writers, 1MB data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.Write(largeData)
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})
		})
	})

	Describe("Read operations", func() {
		It("[TC-BC-002] should benchmark Read with varying data sizes", func() {
			experiment := gmeasure.NewExperiment("Read operations")
			AddReportEntry(experiment.Name, experiment)

			smallData := strings.Repeat("x", 100)
			mediumData := strings.Repeat("x", 1024)
			largeData := strings.Repeat("x", 1024*1024)

			experiment.SampleDuration("Read 100B", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				m.SetInput(io.NopCloser(strings.NewReader(smallData)))
				buf := make([]byte, 100)
				m.Read(buf)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Read 1KB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				m.SetInput(io.NopCloser(strings.NewReader(mediumData)))
				buf := make([]byte, 1024)
				m.Read(buf)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Read 1MB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				m.SetInput(io.NopCloser(strings.NewReader(largeData)))
				buf := make([]byte, 1024*1024)
				m.Read(buf)
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})
		})
	})

	Describe("Copy operations", func() {
		It("[TC-BC-003] should benchmark Copy with varying writer counts and data sizes", func() {
			experiment := gmeasure.NewExperiment("Copy operations")
			AddReportEntry(experiment.Name, experiment)

			smallData := "test data"
			mediumData := strings.Repeat("x", 1024)
			largeData := strings.Repeat("x", 1024*1024)

			experiment.SampleDuration("Single writer, small data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.SetInput(io.NopCloser(strings.NewReader(smallData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("3 writers, small data", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.SetInput(io.NopCloser(strings.NewReader(smallData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Single writer, 1KB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.SetInput(io.NopCloser(strings.NewReader(mediumData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("3 writers, 1KB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.SetInput(io.NopCloser(strings.NewReader(mediumData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Single writer, 1MB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.SetInput(io.NopCloser(strings.NewReader(largeData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("3 writers, 1MB", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.SetInput(io.NopCloser(strings.NewReader(largeData)))
				m.Copy()
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})
		})
	})

	Describe("Mode comparison", func() {
		It("[TC-BC-004] should compare Sequential vs Parallel modes", func() {
			experiment := gmeasure.NewExperiment("Sequential vs Parallel mode")
			AddReportEntry(experiment.Name, experiment)

			data := make([]byte, 1024)

			experiment.SampleDuration("Sequential mode", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3, buf4 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3, &buf4)
				m.Write(data)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Parallel mode", func(idx int) {
				m := multi.New(false, true, multi.DefaultConfig())
				var buf1, buf2, buf3, buf4 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3, &buf4)
				m.Write(data)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Adaptive mode", func(idx int) {
				m := multi.New(true, false, multi.DefaultConfig())
				var buf1, buf2, buf3, buf4 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3, &buf4)
				m.Write(data)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})
		})
	})

	Describe("Writer management", func() {
		It("[TC-BC-005] should benchmark AddWriter, Clean, and SetInput", func() {
			experiment := gmeasure.NewExperiment("Writer management operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Constructor default", func(idx int) {
				_ = multi.New(false, false, multi.DefaultConfig())
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Constructor adaptive", func(idx int) {
				_ = multi.New(true, false, multi.DefaultConfig())
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("AddWriter single", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("AddWriter multiple", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3, buf4, buf5 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3, &buf4, &buf5)
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("SetInput", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				m.SetInput(io.NopCloser(strings.NewReader("data")))
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("Clean", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf1, buf2, buf3 bytes.Buffer
				m.AddWriter(&buf1, &buf2, &buf3)
				m.Clean()
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})

			experiment.SampleDuration("WriteString", func(idx int) {
				m := multi.New(false, false, multi.DefaultConfig())
				var buf bytes.Buffer
				m.AddWriter(&buf)
				m.WriteString("test string")
			}, gmeasure.SamplingConfig{N: 1000, Duration: 0})
		})
	})

	Describe("Real-world scenarios", func() {
		It("[TC-BC-006] should benchmark log broadcasting to multiple destinations", func() {
			experiment := gmeasure.NewExperiment("Log broadcasting")

			experiment.Sample(func(idx int) {
				logLine := "[2024-12-21 15:35:00] INFO: Application event with contextual data\n"
				data := strings.Repeat(logLine, 10000)

				experiment.MeasureDuration("log-broadcast", func() {
					m := multi.New(false, false, multi.DefaultConfig())
					var stdout, file, network bytes.Buffer
					m.AddWriter(&stdout, &file, &network)
					m.Write([]byte(data))
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})

		It("[TC-BC-007] should benchmark stream replication to backup destinations", func() {
			experiment := gmeasure.NewExperiment("Stream replication")

			experiment.Sample(func(idx int) {
				streamData := strings.Repeat("data chunk\n", 50000)

				experiment.MeasureDuration("stream-replicate", func() {
					m := multi.New(false, false, multi.DefaultConfig())
					var primary, backup1, backup2 bytes.Buffer
					m.AddWriter(&primary, &backup1, &backup2)
					m.SetInput(io.NopCloser(strings.NewReader(streamData)))
					m.Copy()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})

		It("[TC-BC-008] should benchmark adaptive mode under varying load", func() {
			experiment := gmeasure.NewExperiment("Adaptive mode under load")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("adaptive-load", func() {
					m := multi.New(true, false, multi.DefaultConfig())
					var buf1, buf2, buf3, buf4, buf5 bytes.Buffer
					m.AddWriter(&buf1, &buf2, &buf3, &buf4, &buf5)

					// Simulate varying data sizes
					for i := 0; i < 100; i++ {
						size := 128 + (i * 10)
						data := make([]byte, size)
						m.Write(data)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})
	})
})
