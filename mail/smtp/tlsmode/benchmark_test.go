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

package tlsmode_test

import (
	"encoding/json"

	. "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("TLS Mode Benchmarks", func() {

	Describe("Parse Performance", func() {
		It("should benchmark Parse from string", func() {
			experiment := gmeasure.NewExperiment("Parse String")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-starttls", func() {
					Parse("starttls")
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Parse 'starttls' mean", experiment.GetStats("parse-starttls").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark Parse with whitespace", func() {
			experiment := gmeasure.NewExperiment("Parse with Whitespace")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-whitespace", func() {
					Parse("  starttls  ")
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Parse with whitespace mean", experiment.GetStats("parse-whitespace").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark ParseBytes", func() {
			experiment := gmeasure.NewExperiment("ParseBytes")
			AddReportEntry(experiment.Name, experiment)

			data := []byte("starttls")
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-bytes", func() {
					ParseBytes(data)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("ParseBytes mean", experiment.GetStats("parse-bytes").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark ParseInt64", func() {
			experiment := gmeasure.NewExperiment("ParseInt64")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-int", func() {
					ParseInt64(1)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("ParseInt64 mean", experiment.GetStats("parse-int").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark ParseFloat64", func() {
			experiment := gmeasure.NewExperiment("ParseFloat64")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-float", func() {
					ParseFloat64(1.0)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("ParseFloat64 mean", experiment.GetStats("parse-float").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Format Performance", func() {
		It("should benchmark String conversion", func() {
			experiment := gmeasure.NewExperiment("String Conversion")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("to-string", func() {
					_ = mode.String()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("String() mean", experiment.GetStats("to-string").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark numeric conversions", func() {
			experiment := gmeasure.NewExperiment("Numeric Conversions")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStrictTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("to-int", func() {
					_ = mode.Int()
				})
				experiment.MeasureDuration("to-uint64", func() {
					_ = mode.Uint64()
				})
				experiment.MeasureDuration("to-float64", func() {
					_ = mode.Float64()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Int() mean", experiment.GetStats("to-int").DurationFor(gmeasure.StatMean))
			AddReportEntry("Uint64() mean", experiment.GetStats("to-uint64").DurationFor(gmeasure.StatMean))
			AddReportEntry("Float64() mean", experiment.GetStats("to-float64").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Encoding Performance", func() {
		It("should benchmark JSON marshaling", func() {
			experiment := gmeasure.NewExperiment("JSON Marshal")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("json-marshal", func() {
					_, _ = json.Marshal(mode)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("JSON Marshal mean", experiment.GetStats("json-marshal").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark JSON unmarshaling", func() {
			experiment := gmeasure.NewExperiment("JSON Unmarshal")
			AddReportEntry(experiment.Name, experiment)

			data := []byte(`"starttls"`)
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("json-unmarshal", func() {
					var mode TLSMode
					_ = json.Unmarshal(data, &mode)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("JSON Unmarshal mean", experiment.GetStats("json-unmarshal").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark Text marshaling", func() {
			experiment := gmeasure.NewExperiment("Text Marshal")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStrictTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("text-marshal", func() {
					_, _ = mode.MarshalText()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Text Marshal mean", experiment.GetStats("text-marshal").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark Text unmarshaling", func() {
			experiment := gmeasure.NewExperiment("Text Unmarshal")
			AddReportEntry(experiment.Name, experiment)

			data := []byte("tls")
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("text-unmarshal", func() {
					var mode TLSMode
					_ = mode.UnmarshalText(data)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Text Unmarshal mean", experiment.GetStats("text-unmarshal").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark CBOR marshaling", func() {
			experiment := gmeasure.NewExperiment("CBOR Marshal")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("cbor-marshal", func() {
					_, _ = mode.MarshalCBOR()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("CBOR Marshal mean", experiment.GetStats("cbor-marshal").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Roundtrip Performance", func() {
		It("should benchmark string roundtrip", func() {
			experiment := gmeasure.NewExperiment("String Roundtrip")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("roundtrip", func() {
					str := mode.String()
					_ = Parse(str)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("String roundtrip mean", experiment.GetStats("roundtrip").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark int64 roundtrip", func() {
			experiment := gmeasure.NewExperiment("Int64 Roundtrip")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStrictTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("roundtrip", func() {
					i := mode.Int64()
					_ = ParseInt64(i)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("Int64 roundtrip mean", experiment.GetStats("roundtrip").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark JSON roundtrip", func() {
			experiment := gmeasure.NewExperiment("JSON Roundtrip")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("roundtrip", func() {
					data, _ := json.Marshal(mode)
					var decoded TLSMode
					_ = json.Unmarshal(data, &decoded)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			AddReportEntry("JSON roundtrip mean", experiment.GetStats("roundtrip").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Comparison Benchmarks", func() {
		It("should compare parsing methods", func() {
			experiment := gmeasure.NewExperiment("Parse Method Comparison")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("parse-string", func() {
					Parse("starttls")
				})
				experiment.MeasureDuration("parse-bytes", func() {
					ParseBytes([]byte("starttls"))
				})
				experiment.MeasureDuration("parse-int", func() {
					ParseInt64(1)
				})
			}, gmeasure.SamplingConfig{N: 500})

			AddReportEntry("Parse string mean", experiment.GetStats("parse-string").DurationFor(gmeasure.StatMean))
			AddReportEntry("Parse bytes mean", experiment.GetStats("parse-bytes").DurationFor(gmeasure.StatMean))
			AddReportEntry("Parse int mean", experiment.GetStats("parse-int").DurationFor(gmeasure.StatMean))
		})

		It("should compare TLS modes", func() {
			experiment := gmeasure.NewExperiment("TLS Mode Comparison")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("none", func() {
					Parse("")
				})
				experiment.MeasureDuration("starttls", func() {
					Parse("starttls")
				})
				experiment.MeasureDuration("tls", func() {
					Parse("tls")
				})
			}, gmeasure.SamplingConfig{N: 500})

			AddReportEntry("TLSNone parse mean", experiment.GetStats("none").DurationFor(gmeasure.StatMean))
			AddReportEntry("TLSStartTLS parse mean", experiment.GetStats("starttls").DurationFor(gmeasure.StatMean))
			AddReportEntry("TLSStrictTLS parse mean", experiment.GetStats("tls").DurationFor(gmeasure.StatMean))
		})
	})

	Describe("Stress Tests", func() {
		It("should benchmark rapid parsing", func() {
			experiment := gmeasure.NewExperiment("Rapid Parsing")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("burst", func() {
					for i := 0; i < 100; i++ {
						Parse("starttls")
					}
				})
			}, gmeasure.SamplingConfig{N: 50})

			AddReportEntry("100 parses mean", experiment.GetStats("burst").DurationFor(gmeasure.StatMean))
		})

		It("should benchmark rapid conversions", func() {
			experiment := gmeasure.NewExperiment("Rapid Conversions")
			AddReportEntry(experiment.Name, experiment)

			mode := TLSStartTLS
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("burst", func() {
					for i := 0; i < 100; i++ {
						_ = mode.String()
						_ = mode.Int64()
						_ = mode.Float64()
					}
				})
			}, gmeasure.SamplingConfig{N: 50})

			AddReportEntry("300 conversions mean", experiment.GetStats("burst").DurationFor(gmeasure.StatMean))
		})
	})
})
