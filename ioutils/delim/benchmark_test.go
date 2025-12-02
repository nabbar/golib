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

package delim_test

import (
	"bytes"
	"io"
	"strings"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
)

// This test file provides performance benchmarks using gmeasure.
// It measures:
//   - Read performance with various data sizes (small, medium, large)
//   - ReadBytes performance across different scenarios
//   - WriteTo performance for data copying
//   - Constructor overhead with different buffer configurations
//   - UnRead operation performance
//   - Memory allocation patterns
//   - Real-world scenarios (CSV parsing, log processing, variable streams)
//
// Benchmarks use gmeasure.Experiment for statistical analysis including:
//   - Minimum, median, mean, max, and standard deviation
//   - Multiple sample iterations for reliability
//   - Performance reports integrated with test output
//
// Run with: go test -v to see performance reports.

var _ = Describe("BufferDelim Benchmarks", func() {
	Describe("Read performance", func() {
		It("should efficiently read small chunks", func() {
			experiment := gmeasure.NewExperiment("Read small chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("read-small", func() {
					data := strings.Repeat("small line\n", 100)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := make([]byte, 100)
					for {
						_, err := bd.Read(buf)
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently read medium chunks", func() {
			experiment := gmeasure.NewExperiment("Read medium chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("read-medium", func() {
					data := strings.Repeat("medium length line with more content\n", 500)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := make([]byte, 200)
					for {
						_, err := bd.Read(buf)
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently read large chunks", func() {
			experiment := gmeasure.NewExperiment("Read large chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("read-large", func() {
					data := strings.Repeat(strings.Repeat("x", 1000)+"\n", 100)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := make([]byte, 2000)
					for {
						_, err := bd.Read(buf)
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})
	})

	Describe("ReadBytes performance", func() {
		It("should efficiently read with ReadBytes - small data", func() {
			experiment := gmeasure.NewExperiment("ReadBytes small data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("readbytes-small", func() {
					data := strings.Repeat("line\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently read with ReadBytes - medium data", func() {
			experiment := gmeasure.NewExperiment("ReadBytes medium data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("readbytes-medium", func() {
					data := strings.Repeat(strings.Repeat("x", 100)+"\n", 500)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently read with ReadBytes - large data", func() {
			experiment := gmeasure.NewExperiment("ReadBytes large data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("readbytes-large", func() {
					data := strings.Repeat(strings.Repeat("x", 1000)+"\n", 100)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})
	})

	Describe("WriteTo performance", func() {
		It("should efficiently write small data", func() {
			experiment := gmeasure.NewExperiment("WriteTo small data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("writeto-small", func() {
					data := strings.Repeat("line\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently write medium data", func() {
			experiment := gmeasure.NewExperiment("WriteTo medium data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("writeto-medium", func() {
					data := strings.Repeat(strings.Repeat("x", 100)+"\n", 500)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("should efficiently write large data", func() {
			experiment := gmeasure.NewExperiment("WriteTo large data")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("writeto-large", func() {
					data := strings.Repeat(strings.Repeat("x", 1000)+"\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Buffer size impact", func() {
		It("performance with default buffer", func() {
			experiment := gmeasure.NewExperiment("Default buffer")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("default-buffer", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with small buffer (64 bytes)", func() {
			experiment := gmeasure.NewExperiment("Small buffer 64B")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("buffer-64", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 64*libsiz.SizeUnit)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with medium buffer (1KB)", func() {
			experiment := gmeasure.NewExperiment("Medium buffer 1KB")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("buffer-1kb", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', libsiz.SizeKilo)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with large buffer (64KB)", func() {
			experiment := gmeasure.NewExperiment("Large buffer 64KB")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("buffer-64kb", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 64*libsiz.SizeKilo)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})
	})

	Describe("Different delimiters performance", func() {
		It("performance with newline delimiter", func() {
			experiment := gmeasure.NewExperiment("Newline delimiter")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("delim-newline", func() {
					data := strings.Repeat("test\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with comma delimiter", func() {
			experiment := gmeasure.NewExperiment("Comma delimiter")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("delim-comma", func() {
					data := strings.Repeat("test,", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, ',', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with pipe delimiter", func() {
			experiment := gmeasure.NewExperiment("Pipe delimiter")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("delim-pipe", func() {
					data := strings.Repeat("test|", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '|', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("performance with null byte delimiter", func() {
			experiment := gmeasure.NewExperiment("Null byte delimiter")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("delim-null", func() {
					data := strings.Repeat("test\x00", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, 0, 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})
	})

	Describe("Copy vs WriteTo performance", func() {
		It("Copy method performance", func() {
			experiment := gmeasure.NewExperiment("Copy method")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("copy", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.Copy(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("WriteTo method performance", func() {
			experiment := gmeasure.NewExperiment("WriteTo method")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("writeto", func() {
					data := strings.Repeat("test line\n", 5000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})
	})

	Describe("DiscardCloser performance", func() {
		It("Read performance", func() {
			experiment := gmeasure.NewExperiment("DiscardCloser Read")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("discard-read", func() {
					dc := iotdlm.DiscardCloser{}
					buf := make([]byte, 1024)
					for i := 0; i < 10000; i++ {
						_, _ = dc.Read(buf)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})

		It("Write performance", func() {
			experiment := gmeasure.NewExperiment("DiscardCloser Write")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("discard-write", func() {
					dc := iotdlm.DiscardCloser{}
					data := []byte("test data to discard")
					for i := 0; i < 10000; i++ {
						_, _ = dc.Write(data)
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})

		It("Close performance", func() {
			experiment := gmeasure.NewExperiment("DiscardCloser Close")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("discard-close", func() {
					dc := iotdlm.DiscardCloser{}
					for i := 0; i < 10000; i++ {
						_ = dc.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Construction overhead", func() {
		It("New constructor performance", func() {
			experiment := gmeasure.NewExperiment("Constructor default")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("new-default", func() {
					for i := 0; i < 1000; i++ {
						data := "test\n"
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 0)
						_ = bd.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})

		It("New constructor with custom buffer", func() {
			experiment := gmeasure.NewExperiment("Constructor custom buffer")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("new-custom", func() {
					for i := 0; i < 1000; i++ {
						data := "test\n"
						r := io.NopCloser(strings.NewReader(data))
						bd := iotdlm.New(r, '\n', 4096)
						_ = bd.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("UnRead performance", func() {
		It("UnRead call performance", func() {
			experiment := gmeasure.NewExperiment("UnRead operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("unread", func() {
					data := strings.Repeat("test\n", 100)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 1024)

					for i := 0; i < 50; i++ {
						_, _ = bd.ReadBytes()
						_, _ = bd.UnRead()
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})
	})

	Describe("Memory allocation patterns", func() {
		It("allocation in Read operations", func() {
			experiment := gmeasure.NewExperiment("Read allocations")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("read-alloc", func() {
					data := strings.Repeat("line\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := make([]byte, 100)
					for {
						_, err := bd.Read(buf)
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("allocation in ReadBytes operations", func() {
			experiment := gmeasure.NewExperiment("ReadBytes allocations")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("readbytes-alloc", func() {
					data := strings.Repeat("line\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})
	})

	Describe("Real-world scenarios", func() {
		It("CSV-like data parsing", func() {
			experiment := gmeasure.NewExperiment("CSV parsing")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("csv-parse", func() {
					data := strings.Repeat("col1,col2,col3,col4,col5\n", 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("Log file processing", func() {
			experiment := gmeasure.NewExperiment("Log processing")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("log-process", func() {
					logLine := "[2024-01-01 12:00:00] INFO: Sample log message with some data\n"
					data := strings.Repeat(logLine, 1000)
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})
		})

		It("Stream processing with various line lengths", func() {
			experiment := gmeasure.NewExperiment("Variable length streams")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("variable-stream", func() {
					var data strings.Builder
					for i := 0; i < 1000; i++ {
						lineLen := (i % 100) + 10
						data.WriteString(strings.Repeat("x", lineLen))
						data.WriteString("\n")
					}
					r := io.NopCloser(strings.NewReader(data.String()))
					bd := iotdlm.New(r, '\n', 0)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})
})
