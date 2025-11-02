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
)

var _ = Describe("Metrics Types and Attributes", func() {
	Describe("Labels", func() {
		Context("when adding labels", func() {
			It("should add a single label", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("method")
				labels := m.GetLabel()
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(ContainElement("method"))
			})

			It("should add multiple labels at once", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("method", "status", "endpoint")
				labels := m.GetLabel()
				Expect(labels).To(HaveLen(3))
				Expect(labels).To(ContainElements("method", "status", "endpoint"))
			})

			It("should add labels incrementally", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("method")
				m.AddLabel("status")
				m.AddLabel("endpoint")
				labels := m.GetLabel()
				Expect(labels).To(HaveLen(3))
				Expect(labels).To(ContainElements("method", "status", "endpoint"))
			})

			It("should preserve order of labels", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("first", "second", "third")
				labels := m.GetLabel()
				Expect(labels[0]).To(Equal("first"))
				Expect(labels[1]).To(Equal("second"))
				Expect(labels[2]).To(Equal("third"))
			})

			It("should handle empty label slice", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel()
				labels := m.GetLabel()
				Expect(labels).To(BeEmpty())
			})

			It("should allow duplicate labels", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("method", "method")
				labels := m.GetLabel()
				Expect(labels).To(HaveLen(2))
			})

			It("should handle empty string labels", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("", "valid", "")
				labels := m.GetLabel()
				Expect(labels).To(HaveLen(3))
			})
		})

		Context("when retrieving labels", func() {
			It("should return empty slice for metric without labels", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				labels := m.GetLabel()
				Expect(labels).To(BeEmpty())
			})

			It("should return all added labels", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				m.AddLabel("label1", "label2", "label3")
				labels := m.GetLabel()
				Expect(labels).To(Equal([]string{"label1", "label2", "label3"}))
			})
		})
	})

	Describe("Buckets", func() {
		Context("when adding buckets for Histogram", func() {
			It("should add a single bucket", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(0.5)
				buckets := m.GetBuckets()
				Expect(buckets).To(HaveLen(1))
				Expect(buckets).To(ContainElement(0.5))
			})

			It("should add multiple buckets at once", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(0.1, 0.5, 1.0, 5.0)
				buckets := m.GetBuckets()
				Expect(buckets).To(HaveLen(4))
				Expect(buckets).To(ContainElements(0.1, 0.5, 1.0, 5.0))
			})

			It("should add buckets incrementally", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(0.1)
				m.AddBuckets(0.5)
				m.AddBuckets(1.0)
				buckets := m.GetBuckets()
				Expect(buckets).To(HaveLen(3))
				Expect(buckets).To(ContainElements(0.1, 0.5, 1.0))
			})

			It("should preserve order of buckets", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(0.1, 0.5, 1.0, 5.0, 10.0)
				buckets := m.GetBuckets()
				Expect(buckets[0]).To(Equal(0.1))
				Expect(buckets[1]).To(Equal(0.5))
				Expect(buckets[2]).To(Equal(1.0))
				Expect(buckets[3]).To(Equal(5.0))
				Expect(buckets[4]).To(Equal(10.0))
			})

			It("should handle empty bucket slice", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets()
				buckets := m.GetBuckets()
				Expect(buckets).To(BeEmpty())
			})

			It("should allow duplicate buckets", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(1.0, 1.0, 2.0)
				buckets := m.GetBuckets()
				Expect(buckets).To(HaveLen(3))
			})

			It("should handle negative buckets", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(-1.0, 0.0, 1.0)
				buckets := m.GetBuckets()
				Expect(buckets).To(ContainElements(-1.0, 0.0, 1.0))
			})

			It("should handle very large bucket values", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(1e10, 1e20, 1e30)
				buckets := m.GetBuckets()
				Expect(buckets).To(ContainElements(1e10, 1e20, 1e30))
			})

			It("should handle very small bucket values", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(1e-10, 1e-20, 1e-30)
				buckets := m.GetBuckets()
				Expect(buckets).To(ContainElements(1e-10, 1e-20, 1e-30))
			})
		})

		Context("when retrieving buckets", func() {
			It("should return empty slice for metric without buckets", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				buckets := m.GetBuckets()
				Expect(buckets).To(BeEmpty())
			})

			It("should return all added buckets", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				m.AddBuckets(0.1, 0.5, 1.0, 5.0, 10.0)
				buckets := m.GetBuckets()
				Expect(buckets).To(Equal([]float64{0.1, 0.5, 1.0, 5.0, 10.0}))
			})
		})
	})

	Describe("Objectives", func() {
		Context("when adding objectives for Summary", func() {
			It("should add a single objective", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				m.AddObjective(0.5, 0.05)
				objectives := m.GetObjectives()
				Expect(objectives).To(HaveLen(1))
				Expect(objectives).To(HaveKeyWithValue(0.5, 0.05))
			})

			It("should add multiple objectives", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				m.AddObjective(0.5, 0.05)
				m.AddObjective(0.9, 0.01)
				m.AddObjective(0.99, 0.001)
				objectives := m.GetObjectives()
				Expect(objectives).To(HaveLen(3))
				Expect(objectives).To(HaveKeyWithValue(0.5, 0.05))
				Expect(objectives).To(HaveKeyWithValue(0.9, 0.01))
				Expect(objectives).To(HaveKeyWithValue(0.99, 0.001))
			})

			It("should update existing objective", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				m.AddObjective(0.5, 0.05)
				m.AddObjective(0.5, 0.1)
				objectives := m.GetObjectives()
				Expect(objectives).To(HaveLen(1))
				Expect(objectives).To(HaveKeyWithValue(0.5, 0.1))
			})

			It("should handle edge case quantiles", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				m.AddObjective(0.0, 0.0)
				m.AddObjective(1.0, 0.0)
				objectives := m.GetObjectives()
				Expect(objectives).To(HaveKeyWithValue(0.0, 0.0))
				Expect(objectives).To(HaveKeyWithValue(1.0, 0.0))
			})

			It("should handle high precision objectives", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				m.AddObjective(0.999, 0.0001)
				m.AddObjective(0.9999, 0.00001)
				objectives := m.GetObjectives()
				Expect(objectives).To(HaveKeyWithValue(0.999, 0.0001))
				Expect(objectives).To(HaveKeyWithValue(0.9999, 0.00001))
			})

			It("should handle empty objectives initially", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				objectives := m.GetObjectives()
				Expect(objectives).To(BeEmpty())
			})
		})

		Context("when retrieving objectives", func() {
			It("should return empty map for metric without objectives", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				objectives := m.GetObjectives()
				Expect(objectives).To(BeEmpty())
			})

			It("should return all added objectives", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				expected := map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.99: 0.001,
				}
				for k, v := range expected {
					m.AddObjective(k, v)
				}
				objectives := m.GetObjectives()
				Expect(objectives).To(Equal(expected))
			})
		})
	})

	Describe("Combined operations", func() {
		It("should handle all attributes together", func() {
			m := prmmet.NewMetrics("complete_metric", prmtps.Histogram)
			m.SetDesc("Complete metric with all attributes")
			m.AddLabel("method", "status")
			m.AddBuckets(0.1, 0.5, 1.0, 5.0)

			Expect(m.GetName()).To(Equal("complete_metric"))
			Expect(m.GetType()).To(Equal(prmtps.Histogram))
			Expect(m.GetDesc()).To(Equal("Complete metric with all attributes"))
			Expect(m.GetLabel()).To(Equal([]string{"method", "status"}))
			Expect(m.GetBuckets()).To(Equal([]float64{0.1, 0.5, 1.0, 5.0}))
		})

		It("should maintain independence between metrics", func() {
			m1 := prmmet.NewMetrics("metric1", prmtps.Counter)
			m1.AddLabel("label1")

			m2 := prmmet.NewMetrics("metric2", prmtps.Counter)
			m2.AddLabel("label2")

			Expect(m1.GetLabel()).To(Equal([]string{"label1"}))
			Expect(m2.GetLabel()).To(Equal([]string{"label2"}))
		})
	})
})
