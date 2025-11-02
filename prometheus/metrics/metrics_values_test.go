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

package metrics_test

import (
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Metrics Value Operations", func() {
	Describe("SetGaugeValue", func() {
		Context("with valid Gauge metric", func() {
			var m = newGaugeMetric("test_gauge_set_value", "method")

			BeforeEach(func() {
				m = newGaugeMetric("test_gauge_set_value", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should set gauge value successfully", func() {
				err := m.SetGaugeValue([]string{"GET"}, 42.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should set multiple different label values", func() {
				Expect(m.SetGaugeValue([]string{"GET"}, 10.0)).ToNot(HaveOccurred())
				Expect(m.SetGaugeValue([]string{"POST"}, 20.0)).ToNot(HaveOccurred())
				Expect(m.SetGaugeValue([]string{"DELETE"}, 30.0)).ToNot(HaveOccurred())
			})

			It("should allow overwriting existing value", func() {
				Expect(m.SetGaugeValue([]string{"GET"}, 10.0)).ToNot(HaveOccurred())
				Expect(m.SetGaugeValue([]string{"GET"}, 99.9)).ToNot(HaveOccurred())
			})

			It("should handle zero value", func() {
				err := m.SetGaugeValue([]string{"GET"}, 0.0)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle negative values", func() {
				err := m.SetGaugeValue([]string{"GET"}, -123.45)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle very large values", func() {
				err := m.SetGaugeValue([]string{"GET"}, 1e20)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle very small values", func() {
				err := m.SetGaugeValue([]string{"GET"}, 1e-20)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with multiple labels", func() {
			var m = newGaugeMetric("test_gauge_multi_label", "method", "status", "endpoint")

			BeforeEach(func() {
				m = newGaugeMetric("test_gauge_multi_label", "method", "status", "endpoint")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should set value with all labels", func() {
				err := m.SetGaugeValue([]string{"GET", "200", "/api/v1"}, 100.0)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle different label combinations", func() {
				Expect(m.SetGaugeValue([]string{"GET", "200", "/api/v1"}, 100.0)).ToNot(HaveOccurred())
				Expect(m.SetGaugeValue([]string{"POST", "201", "/api/v2"}, 200.0)).ToNot(HaveOccurred())
				Expect(m.SetGaugeValue([]string{"DELETE", "204", "/api/v3"}, 300.0)).ToNot(HaveOccurred())
			})
		})

		Context("with invalid metric types", func() {
			It("should return error for Counter type", func() {
				m := newCounterMetric("test_counter_set", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.SetGaugeValue([]string{"GET"}, 42.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Gauge type"))
			})

			It("should return error for Histogram type", func() {
				m := newHistogramMetric("test_histogram_set", []float64{0.1, 0.5, 1.0}, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.SetGaugeValue([]string{"GET"}, 42.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Gauge type"))
			})

			It("should return error for Summary type", func() {
				m := newSummaryMetric("test_summary_set", nil, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.SetGaugeValue([]string{"GET"}, 42.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Gauge type"))
			})

			It("should return error for None type", func() {
				m := newMetricWithRegistration("test_none_set", prmtps.None)
				err := m.SetGaugeValue([]string{"GET"}, 42.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not existed"))
			})

			It("should return error for unregistered metric", func() {
				m := prmmet.NewMetrics("test_unregistered_set", prmtps.Gauge)
				m.AddLabel("method")
				// Not registering the metric
				err := m.SetGaugeValue([]string{"GET"}, 42.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not registred"))
			})
		})
	})

	Describe("Inc", func() {
		Context("with Counter metric", func() {
			var m = newCounterMetric("test_counter_inc", "method")

			BeforeEach(func() {
				m = newCounterMetric("test_counter_inc", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should increment counter successfully", func() {
				err := m.Inc([]string{"GET"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("should increment multiple times", func() {
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())
			})

			It("should increment different label values independently", func() {
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"POST"})).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"DELETE"})).ToNot(HaveOccurred())
			})
		})

		Context("with Gauge metric", func() {
			var m = newGaugeMetric("test_gauge_inc", "method")

			BeforeEach(func() {
				m = newGaugeMetric("test_gauge_inc", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should increment gauge successfully", func() {
				err := m.Inc([]string{"GET"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("should allow inc after set", func() {
				Expect(m.SetGaugeValue([]string{"GET"}, 10.0)).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())
			})
		})

		Context("with invalid metric types", func() {
			It("should return error for Histogram type", func() {
				m := newHistogramMetric("test_histogram_inc", []float64{0.1, 0.5, 1.0}, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Inc([]string{"GET"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Counter or Gauge type"))
			})

			It("should return error for Summary type", func() {
				m := newSummaryMetric("test_summary_inc", nil, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Inc([]string{"GET"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Counter or Gauge type"))
			})

			It("should return error for None type", func() {
				m := newMetricWithRegistration("test_none_inc", prmtps.None)
				err := m.Inc([]string{"GET"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not existed"))
			})

			It("should return error for unregistered metric", func() {
				m := prmmet.NewMetrics("test_unregistered_inc", prmtps.Counter)
				m.AddLabel("method")
				// Not registering the metric
				err := m.Inc([]string{"GET"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not registred"))
			})
		})
	})

	Describe("Add", func() {
		Context("with Counter metric", func() {
			var m = newCounterMetric("test_counter_add", "method")

			BeforeEach(func() {
				m = newCounterMetric("test_counter_add", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should add positive value to counter", func() {
				err := m.Add([]string{"GET"}, 10.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should add multiple values", func() {
				Expect(m.Add([]string{"GET"}, 5.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"GET"}, 3.5)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"GET"}, 1.5)).ToNot(HaveOccurred())
			})

			It("should add to different label values", func() {
				Expect(m.Add([]string{"GET"}, 10.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"POST"}, 20.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"DELETE"}, 30.0)).ToNot(HaveOccurred())
			})

			It("should handle zero value", func() {
				err := m.Add([]string{"GET"}, 0.0)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle very large values", func() {
				err := m.Add([]string{"GET"}, 1e10)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle fractional values", func() {
				err := m.Add([]string{"GET"}, 0.123456789)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with Gauge metric", func() {
			var m = newGaugeMetric("test_gauge_add", "method")

			BeforeEach(func() {
				m = newGaugeMetric("test_gauge_add", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should add positive value to gauge", func() {
				err := m.Add([]string{"GET"}, 10.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should add negative value to gauge", func() {
				err := m.Add([]string{"GET"}, -5.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should allow add after set", func() {
				Expect(m.SetGaugeValue([]string{"GET"}, 100.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"GET"}, 50.0)).ToNot(HaveOccurred())
			})

			It("should handle positive and negative adds", func() {
				Expect(m.Add([]string{"GET"}, 100.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"GET"}, -50.0)).ToNot(HaveOccurred())
				Expect(m.Add([]string{"GET"}, 25.0)).ToNot(HaveOccurred())
			})
		})

		Context("with invalid metric types", func() {
			It("should return error for Histogram type", func() {
				m := newHistogramMetric("test_histogram_add", []float64{0.1, 0.5, 1.0}, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Add([]string{"GET"}, 10.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Counter or Gauge type"))
			})

			It("should return error for Summary type", func() {
				m := newSummaryMetric("test_summary_add", nil, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Add([]string{"GET"}, 10.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Counter or Gauge type"))
			})

			It("should return error for None type", func() {
				m := newMetricWithRegistration("test_none_add", prmtps.None)
				err := m.Add([]string{"GET"}, 10.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not existed"))
			})

			It("should return error for unregistered metric", func() {
				m := prmmet.NewMetrics("test_unregistered_add", prmtps.Counter)
				m.AddLabel("method")
				// Not registering the metric
				err := m.Add([]string{"GET"}, 5.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not registred"))
			})
		})
	})

	Describe("Observe", func() {
		Context("with Histogram metric", func() {
			var m = newHistogramMetric("test_histogram_observe", prmsdk.DefBuckets, "method")

			BeforeEach(func() {
				m = newHistogramMetric("test_histogram_observe", prmsdk.DefBuckets, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should observe values successfully", func() {
				err := m.Observe([]string{"GET"}, 0.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should observe multiple values", func() {
				Expect(m.Observe([]string{"GET"}, 0.1)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 1.0)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 5.0)).ToNot(HaveOccurred())
			})

			It("should observe different label values", func() {
				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"POST"}, 1.0)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"DELETE"}, 2.0)).ToNot(HaveOccurred())
			})

			It("should handle zero value", func() {
				err := m.Observe([]string{"GET"}, 0.0)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle negative values", func() {
				err := m.Observe([]string{"GET"}, -1.0)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle very large values", func() {
				err := m.Observe([]string{"GET"}, 1e10)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle very small values", func() {
				err := m.Observe([]string{"GET"}, 1e-10)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with Summary metric", func() {
			var m = newSummaryMetric("test_summary_observe", nil, "method")

			BeforeEach(func() {
				m = newSummaryMetric("test_summary_observe", nil, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				cleanupMetric(m)
			})

			It("should observe values successfully", func() {
				err := m.Observe([]string{"GET"}, 0.5)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should observe multiple values", func() {
				Expect(m.Observe([]string{"GET"}, 0.1)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 1.0)).ToNot(HaveOccurred())
			})

			It("should observe different label values", func() {
				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"POST"}, 1.0)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"DELETE"}, 2.0)).ToNot(HaveOccurred())
			})

			It("should handle distribution of values", func() {
				for i := 0; i < 100; i++ {
					Expect(m.Observe([]string{"GET"}, float64(i))).ToNot(HaveOccurred())
				}
			})
		})

		Context("with invalid metric types", func() {
			It("should return error for Counter type", func() {
				m := newCounterMetric("test_counter_observe", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Observe([]string{"GET"}, 1.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Histogram or Summary type"))
			})

			It("should return error for Gauge type", func() {
				m := newGaugeMetric("test_gauge_observe", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				err := m.Observe([]string{"GET"}, 1.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not Histogram or Summary type"))
			})

			It("should return error for None type", func() {
				m := newMetricWithRegistration("test_none_observe", prmtps.None)
				err := m.Observe([]string{"GET"}, 1.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not existed"))
			})

			It("should return error for unregistered metric", func() {
				m := prmmet.NewMetrics("test_unregistered_observe", prmtps.Histogram)
				m.AddLabel("method")
				m.AddBuckets(prmsdk.DefBuckets...)
				// Not registering the metric
				err := m.Observe([]string{"GET"}, 1.0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not registred"))
			})
		})
	})
})
