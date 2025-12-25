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

package helper_test

import (
	"bytes"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	arccmp "github.com/nabbar/golib/archive/compress"
	"github.com/nabbar/golib/archive/helper"
)

var _ = Describe("TC-BC-001: Benchmark Tests", func() {
	Context("TC-BC-010: Constructor performance", func() {
		It("TC-BC-011: should benchmark NewReader creation", func() {
			experiment := gmeasure.NewExperiment("NewReader creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create", func() {
					h, err := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader("test"))
					Expect(err).ToNot(HaveOccurred())
					h.Close()
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})
		})

		It("TC-BC-012: should benchmark NewWriter creation", func() {
			experiment := gmeasure.NewExperiment("NewWriter creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create", func() {
					var buf bytes.Buffer
					h, err := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					Expect(err).ToNot(HaveOccurred())
					h.Close()
				})
			}, gmeasure.SamplingConfig{N: 100, Duration: 0})
		})
	})

	Context("TC-BC-020: Compress reader performance", func() {
		It("TC-BC-021: should benchmark small data compression", func() {
			experiment := gmeasure.NewExperiment("Compress reader - small data")
			AddReportEntry(experiment.Name, experiment)

			data := "Hello, World!"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					h, _ := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
					defer h.Close()
					io.ReadAll(h)
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 0})
		})

		It("TC-BC-022: should benchmark medium data compression", func() {
			experiment := gmeasure.NewExperiment("Compress reader - medium data")
			AddReportEntry(experiment.Name, experiment)

			data := strings.Repeat("test data ", 1000)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					h, _ := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
					defer h.Close()
					io.ReadAll(h)
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("TC-BC-023: should benchmark large data compression", func() {
			experiment := gmeasure.NewExperiment("Compress reader - large data")
			AddReportEntry(experiment.Name, experiment)

			data := strings.Repeat("test data with variety ", 10000)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					h, _ := helper.NewReader(arccmp.Gzip, helper.Compress, strings.NewReader(data))
					defer h.Close()
					io.ReadAll(h)
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Context("TC-BC-030: Compress writer performance", func() {
		It("TC-BC-031: should benchmark small data writes", func() {
			experiment := gmeasure.NewExperiment("Compress writer - small data")
			AddReportEntry(experiment.Name, experiment)

			data := []byte("Hello, World!")

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					var buf bytes.Buffer
					h, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					h.Write(data)
					h.Close()
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 0})
		})

		It("TC-BC-032: should benchmark medium data writes", func() {
			experiment := gmeasure.NewExperiment("Compress writer - medium data")
			AddReportEntry(experiment.Name, experiment)

			data := bytes.Repeat([]byte("test data "), 1000)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					var buf bytes.Buffer
					h, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					h.Write(data)
					h.Close()
				})
			}, gmeasure.SamplingConfig{N: 20, Duration: 0})
		})

		It("TC-BC-033: should benchmark large data writes", func() {
			experiment := gmeasure.NewExperiment("Compress writer - large data")
			AddReportEntry(experiment.Name, experiment)

			data := bytes.Repeat([]byte("test data with variety "), 10000)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("compress", func() {
					var buf bytes.Buffer
					h, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &buf)
					h.Write(data)
					h.Close()
				})
			}, gmeasure.SamplingConfig{N: 10, Duration: 0})
		})
	})

	Context("TC-BC-040: Decompress performance", func() {
		It("TC-BC-041: should benchmark decompression", func() {
			experiment := gmeasure.NewExperiment("Decompress reader")
			AddReportEntry(experiment.Name, experiment)

			compressed := []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd, 0xc9, 0xc9, 0x07,
				0x00, 0x82, 0x89, 0xd1, 0xf7, 0x05, 0x00, 0x00,
				0x00,
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("decompress", func() {
					h, _ := helper.NewReader(arccmp.Gzip, helper.Decompress, bytes.NewReader(compressed))
					defer h.Close()
					io.ReadAll(h)
				})
			}, gmeasure.SamplingConfig{N: 50, Duration: 0})
		})
	})

	Context("TC-BC-050: Round-trip performance", func() {
		It("TC-BC-051: should benchmark compress and decompress cycle", func() {
			experiment := gmeasure.NewExperiment("Round-trip compression")
			AddReportEntry(experiment.Name, experiment)

			original := "Test data for round trip performance"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("roundtrip", func() {
					var compressed bytes.Buffer
					cw, _ := helper.NewWriter(arccmp.Gzip, helper.Compress, &compressed)
					cw.Write([]byte(original))
					cw.Close()

					var decompressed bytes.Buffer
					dw, _ := helper.NewWriter(arccmp.Gzip, helper.Decompress, &decompressed)
					dw.Write(compressed.Bytes())
					dw.Close()
				})
			}, gmeasure.SamplingConfig{N: 30, Duration: 0})
		})
	})
})
