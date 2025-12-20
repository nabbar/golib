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
		It("should efficiently read with Read", func() {
			data := strings.Repeat(strings.Repeat("x", 1023)+"\n", 20480)
			fctBytes := func(r iotdlm.BufferDelim, size int) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()

				buf := make([]byte, size) // Small read buffer
				for {
					_, err := r.Read(buf)
					if err == io.EOF {
						break
					}
				}
			}

			experiment := gmeasure.NewExperiment("Read 20MB data")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("128B Buffer Read", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 32*libsiz.SizeKilo, false), 128)
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("1KB Buffer Read", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 32*libsiz.SizeKilo, false), 1024)
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("4KB Buffer Read", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 32*libsiz.SizeKilo, false), 4*1024)
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("ReadBytes performance", func() {
		It("should efficiently read with ReadBytes", func() {
			dataSmall := strings.Repeat(strings.Repeat("x", 127)+"\n", 20480)
			dataMedium := strings.Repeat(strings.Repeat("x", 1023)+"\n", 20480)
			dataLarge := strings.Repeat(strings.Repeat("x", 4095)+"\n", 20480)
			fctBytes := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()

				for {
					_, err := r.ReadBytes()
					if err == io.EOF {
						break
					}
				}
			}

			experiment := gmeasure.NewExperiment("ReadBytes 20480 chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Chunk of 128B", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataSmall)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Chunk of 1KB", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataMedium)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Chunk of 4KB", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataLarge)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("WriteTo performance", func() {
		It("should efficiently write with WriteTo", func() {
			dataSmall := strings.Repeat(strings.Repeat("x", 127)+"\n", 20480)
			dataMedium := strings.Repeat(strings.Repeat("x", 1023)+"\n", 20480)
			dataLarge := strings.Repeat(strings.Repeat("x", 4095)+"\n", 20480)
			fctBytes := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()
				_, _ = r.WriteTo(&bytes.Buffer{})
			}

			experiment := gmeasure.NewExperiment("WriteTo 20480 chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Chunk of 128B", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataSmall)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Chunk of 1KB", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataMedium)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Chunk of 4KB", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(dataLarge)), '\n', 32*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Buffer size impact", func() {
		It("performance with default buffer", func() {
			data := strings.Repeat(strings.Repeat("x", 4095)+"\n", 20480)
			fctBytes := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()
				for {
					_, err := r.ReadBytes()
					if err == io.EOF {
						break
					}
				}
			}

			experiment := gmeasure.NewExperiment("Buffer size impact for 20MB data")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Default buffer", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("64B Buffer", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 64*libsiz.SizeUnit, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("1KB Buffer", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("64KB Buffer", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(io.NopCloser(strings.NewReader(data)), '\n', 64*libsiz.SizeKilo, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Different delimiters performance", func() {
		It("performance with newline delimiter", func() {
			data := func(mrk string) io.ReadCloser {
				return io.NopCloser(strings.NewReader(strings.Repeat(strings.Repeat("x", 4095)+mrk, 20480)))
			}

			fctBytes := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()
				for {
					_, err := r.ReadBytes()
					if err == io.EOF {
						break
					}
				}
			}

			experiment := gmeasure.NewExperiment("Different delimiters performance for 20480 chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Newline delimiter", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(data("\n"), '\n', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Comma delimiter", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(data(","), ',', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Pipe delimiter", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(data("|"), '|', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Null byte delimiter", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctBytes(iotdlm.New(data("\x00"), 0, 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Copy vs WriteTo performance", func() {
		It("Copy method performance", func() {
			data := func(mrk string) io.ReadCloser {
				return io.NopCloser(strings.NewReader(strings.Repeat(strings.Repeat("x", 4095)+mrk, 20480)))
			}

			fctCopy := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()
				_, _ = r.Copy(&bytes.Buffer{})
			}

			fctWriteTo := func(r iotdlm.BufferDelim) {
				defer GinkgoRecover()
				defer func() {
					_ = r.Close()
				}()
				_, _ = r.WriteTo(&bytes.Buffer{})
			}

			experiment := gmeasure.NewExperiment("Copy vs WriteTo with 20480 chunks")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("Copy method", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctCopy(iotdlm.New(data("\n"), '\n', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("WriteTo method", func(idx int) {
				// 20,000 lines, small size (~10 bytes)
				fctWriteTo(iotdlm.New(data("\n"), '\n', 0, false))
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Others performance", func() {
		It("Others performance", func() {
			data := io.NopCloser(strings.NewReader(strings.Repeat(strings.Repeat("x", 4095)+"\n", 20480)))

			experiment := gmeasure.NewExperiment("Others performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.SampleDuration("DiscardCloser Read", func(idx int) {
				dc := iotdlm.DiscardCloser{}
				buf := make([]byte, 1024)
				for i := 0; i < 50000; i++ {
					_, _ = dc.Read(buf)
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("DiscardCloser Write", func(idx int) {
				dc := iotdlm.DiscardCloser{}
				data := []byte("test data to discard")
				for i := 0; i < 50000; i++ {
					_, _ = dc.Write(data)
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("DiscardCloser Close", func(idx int) {
				dc := iotdlm.DiscardCloser{}
				for i := 0; i < 50000; i++ {
					_ = dc.Close()
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Constructor default", func(idx int) {
				for i := 0; i < 10000; i++ {
					bd := iotdlm.New(data, '\n', 0, false)
					_ = bd.Close()
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Constructor custom buffer", func(idx int) {
				for i := 0; i < 10000; i++ {
					bd := iotdlm.New(data, '\n', 4096, false)
					_ = bd.Close()
				}
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("UnRead operations", func(idx int) {
				bd := iotdlm.New(data, '\n', 1024, false)
				for i := 0; i < 50; i++ {
					_, _ = bd.ReadBytes()
					_, _ = bd.UnRead()
				}
				_ = bd.Close()
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("Read operations", func(idx int) {
				bd := iotdlm.New(data, '\n', 0, false)
				buf := make([]byte, 100)
				for {
					_, err := bd.Read(buf)
					if err == io.EOF {
						break
					}
				}
				_ = bd.Close()
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			experiment.SampleDuration("ReadBytes operations", func(idx int) {
				bd := iotdlm.New(data, '\n', 0, false)
				for {
					_, err := bd.ReadBytes()
					if err == io.EOF {
						break
					}
				}
				_ = bd.Close()
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Describe("Real-world scenarios", func() {
		It("CSV-like data parsing", func() {
			experiment := gmeasure.NewExperiment("CSV parsing")

			experiment.Sample(func(idx int) {
				data := strings.Repeat("col1,col2,col3,col4,col5\n", 50000)
				experiment.MeasureDuration("csv-parse", func() {
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0, false)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})

		It("Log file processing", func() {
			experiment := gmeasure.NewExperiment("Log processing")

			experiment.Sample(func(idx int) {
				logLine := "[2024-01-01 12:00:00] INFO: Sample log message with some data\n"
				data := strings.Repeat(logLine, 50000)
				experiment.MeasureDuration("log-process", func() {
					r := io.NopCloser(strings.NewReader(data))
					bd := iotdlm.New(r, '\n', 0, false)

					buf := &bytes.Buffer{}
					_, _ = bd.WriteTo(buf)
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 15, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})

		It("Stream processing with various line lengths", func() {
			experiment := gmeasure.NewExperiment("Variable length streams")

			experiment.Sample(func(idx int) {
				var data strings.Builder
				for i := 0; i < 10000; i++ {
					lineLen := (i % 100) + 10
					data.WriteString(strings.Repeat("x", lineLen))
					data.WriteString("\n")
				}
				dataStr := data.String()
				experiment.MeasureDuration("variable-stream", func() {
					r := io.NopCloser(strings.NewReader(dataStr))
					bd := iotdlm.New(r, '\n', 0, false)

					for {
						_, err := bd.ReadBytes()
						if err == io.EOF {
							break
						}
					}
					_ = bd.Close()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})

			AddReportEntry(experiment.Name, experiment)
		})
	})
})
