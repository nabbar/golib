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

package types_test

import (
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// mockMetric is a test implementation of the Metric interface
type mockMetric struct {
	name       string
	metricType prmtps.MetricType
	desc       string
	labels     []string
	buckets    []float64
	objectives map[float64]float64
}

func (m *mockMetric) GetName() string                    { return m.name }
func (m *mockMetric) GetType() prmtps.MetricType         { return m.metricType }
func (m *mockMetric) GetDesc() string                    { return m.desc }
func (m *mockMetric) GetLabel() []string                 { return m.labels }
func (m *mockMetric) GetBuckets() []float64              { return m.buckets }
func (m *mockMetric) GetObjectives() map[float64]float64 { return m.objectives }

// Helper function to create a mock counter metric
func newMockCounter(name string, labels ...string) *mockMetric {
	return &mockMetric{
		name:       name,
		metricType: prmtps.Counter,
		desc:       "Test counter metric",
		labels:     labels,
	}
}

// Helper function to create a mock gauge metric
func newMockGauge(name string, labels ...string) *mockMetric {
	return &mockMetric{
		name:       name,
		metricType: prmtps.Gauge,
		desc:       "Test gauge metric",
		labels:     labels,
	}
}

// Helper function to create a mock histogram metric
func newMockHistogram(name string, buckets []float64, labels ...string) *mockMetric {
	return &mockMetric{
		name:       name,
		metricType: prmtps.Histogram,
		desc:       "Test histogram metric",
		labels:     labels,
		buckets:    buckets,
	}
}

// Helper function to create a mock summary metric
func newMockSummary(name string, objectives map[float64]float64, labels ...string) *mockMetric {
	return &mockMetric{
		name:       name,
		metricType: prmtps.Summary,
		desc:       "Test summary metric",
		labels:     labels,
		objectives: objectives,
	}
}

var _ = Describe("MetricType Registration", func() {
	Describe("Counter Registration", func() {
		Context("when registering a counter metric", func() {
			It("should successfully create a CounterVec", func() {
				metric := newMockCounter("test_counter_total")

				collector, err := prmtps.Counter.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
				Expect(collector).To(BeAssignableToTypeOf(&prmsdk.CounterVec{}))
			})

			It("should create a counter with labels", func() {
				metric := newMockCounter("test_counter_with_labels", "method", "status")

				collector, err := prmtps.Counter.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())

				// Verify it's a CounterVec
				counterVec, ok := collector.(*prmsdk.CounterVec)
				Expect(ok).To(BeTrue())
				Expect(counterVec).ToNot(BeNil())
			})

			It("should create a counter without labels", func() {
				metric := newMockCounter("test_counter_no_labels")

				collector, err := prmtps.Counter.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})

			It("should handle multiple label dimensions", func() {
				metric := newMockCounter("multi_label_counter", "label1", "label2", "label3", "label4")

				collector, err := prmtps.Counter.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})
		})
	})

	Describe("Gauge Registration", func() {
		Context("when registering a gauge metric", func() {
			It("should successfully create a GaugeVec", func() {
				metric := newMockGauge("test_gauge")

				collector, err := prmtps.Gauge.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
				Expect(collector).To(BeAssignableToTypeOf(&prmsdk.GaugeVec{}))
			})

			It("should create a gauge with labels", func() {
				metric := newMockGauge("test_gauge_with_labels", "server", "region")

				collector, err := prmtps.Gauge.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())

				gaugeVec, ok := collector.(*prmsdk.GaugeVec)
				Expect(ok).To(BeTrue())
				Expect(gaugeVec).ToNot(BeNil())
			})

			It("should create a gauge without labels", func() {
				metric := newMockGauge("test_gauge_simple")

				collector, err := prmtps.Gauge.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})
		})
	})

	Describe("Histogram Registration", func() {
		Context("when registering a histogram metric with valid buckets", func() {
			It("should successfully create a HistogramVec", func() {
				buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
				metric := newMockHistogram("test_histogram_seconds", buckets)

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
				Expect(collector).To(BeAssignableToTypeOf(&prmsdk.HistogramVec{}))
			})

			It("should create a histogram with labels", func() {
				buckets := prmsdk.DefBuckets
				metric := newMockHistogram("test_histogram_with_labels", buckets, "endpoint", "method")

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())

				histogramVec, ok := collector.(*prmsdk.HistogramVec)
				Expect(ok).To(BeTrue())
				Expect(histogramVec).ToNot(BeNil())
			})

			It("should handle custom bucket configurations", func() {
				buckets := []float64{1, 10, 100, 1000, 10000}
				metric := newMockHistogram("test_custom_buckets", buckets)

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})

			It("should handle single bucket", func() {
				buckets := []float64{1.0}
				metric := newMockHistogram("test_single_bucket", buckets)

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})
		})

		Context("when registering a histogram without buckets", func() {
			It("should return an error", func() {
				metric := newMockHistogram("test_histogram_no_buckets", nil)

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("histogram type"))
				Expect(err.Error()).To(ContainSubstring("bucket param"))
				Expect(err.Error()).To(ContainSubstring("test_histogram_no_buckets"))
				Expect(collector).To(BeNil())
			})

			It("should return an error with empty bucket slice", func() {
				metric := newMockHistogram("test_histogram_empty_buckets", []float64{})

				collector, err := prmtps.Histogram.Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bucket param"))
				Expect(collector).To(BeNil())
			})
		})
	})

	Describe("Summary Registration", func() {
		Context("when registering a summary metric with valid objectives", func() {
			It("should successfully create a SummaryVec", func() {
				objectives := map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.99: 0.001,
				}
				metric := newMockSummary("test_summary_seconds", objectives)

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
				Expect(collector).To(BeAssignableToTypeOf(&prmsdk.SummaryVec{}))
			})

			It("should create a summary with labels", func() {
				objectives := map[float64]float64{0.5: 0.05, 0.95: 0.01}
				metric := newMockSummary("test_summary_with_labels", objectives, "service", "endpoint")

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())

				summaryVec, ok := collector.(*prmsdk.SummaryVec)
				Expect(ok).To(BeTrue())
				Expect(summaryVec).ToNot(BeNil())
			})

			It("should handle single objective", func() {
				objectives := map[float64]float64{0.99: 0.001}
				metric := newMockSummary("test_single_objective", objectives)

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})

			It("should handle multiple objectives", func() {
				objectives := map[float64]float64{
					0.25: 0.1,
					0.5:  0.05,
					0.75: 0.025,
					0.9:  0.01,
					0.95: 0.005,
					0.99: 0.001,
				}
				metric := newMockSummary("test_multi_objectives", objectives)

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).ToNot(HaveOccurred())
				Expect(collector).ToNot(BeNil())
			})
		})

		Context("when registering a summary without objectives", func() {
			It("should return an error", func() {
				metric := newMockSummary("test_summary_no_objectives", nil)

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("summary type"))
				Expect(err.Error()).To(ContainSubstring("objectives param"))
				Expect(err.Error()).To(ContainSubstring("test_summary_no_objectives"))
				Expect(collector).To(BeNil())
			})

			It("should return an error with empty objectives map", func() {
				metric := newMockSummary("test_summary_empty_objectives", map[float64]float64{})

				collector, err := prmtps.Summary.Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("objectives param"))
				Expect(collector).To(BeNil())
			})
		})
	})

	Describe("Invalid Type Registration", func() {
		Context("when registering a metric with None type", func() {
			It("should return an error", func() {
				metric := &mockMetric{
					name:       "test_none_type",
					metricType: prmtps.None,
					desc:       "Invalid metric type",
				}

				collector, err := prmtps.None.Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not compatible"))
				Expect(collector).To(BeNil())
			})
		})

		Context("when using an undefined metric type value", func() {
			It("should return an error for invalid type", func() {
				metric := &mockMetric{
					name:       "test_invalid_type",
					metricType: prmtps.MetricType(999), // Invalid type
					desc:       "Invalid metric type",
				}

				collector, err := metric.GetType().Register(metric)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not compatible"))
				Expect(collector).To(BeNil())
			})
		})
	})
})
