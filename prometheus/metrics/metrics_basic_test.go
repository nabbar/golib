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

var _ = Describe("Metrics Basic Operations", func() {
	Describe("NewMetrics", func() {
		Context("when creating a new metric", func() {
			It("should create a Counter metric successfully", func() {
				m := prmmet.NewMetrics("test_counter", prmtps.Counter)
				Expect(m).ToNot(BeNil())
				Expect(m.GetName()).To(Equal("test_counter"))
				Expect(m.GetType()).To(Equal(prmtps.Counter))
			})

			It("should create a Gauge metric successfully", func() {
				m := prmmet.NewMetrics("test_gauge", prmtps.Gauge)
				Expect(m).ToNot(BeNil())
				Expect(m.GetName()).To(Equal("test_gauge"))
				Expect(m.GetType()).To(Equal(prmtps.Gauge))
			})

			It("should create a Histogram metric successfully", func() {
				m := prmmet.NewMetrics("test_histogram", prmtps.Histogram)
				Expect(m).ToNot(BeNil())
				Expect(m.GetName()).To(Equal("test_histogram"))
				Expect(m.GetType()).To(Equal(prmtps.Histogram))
			})

			It("should create a Summary metric successfully", func() {
				m := prmmet.NewMetrics("test_summary", prmtps.Summary)
				Expect(m).ToNot(BeNil())
				Expect(m.GetName()).To(Equal("test_summary"))
				Expect(m.GetType()).To(Equal(prmtps.Summary))
			})

			It("should create a None type metric", func() {
				m := prmmet.NewMetrics("test_none", prmtps.None)
				Expect(m).ToNot(BeNil())
				Expect(m.GetName()).To(Equal("test_none"))
				Expect(m.GetType()).To(Equal(prmtps.None))
			})
		})

		Context("when metric is initialized", func() {
			It("should have empty description by default", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				Expect(m.GetDesc()).To(BeEmpty())
			})

			It("should have empty labels by default", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				Expect(m.GetLabel()).To(BeEmpty())
			})

			It("should have empty buckets by default", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Histogram)
				Expect(m.GetBuckets()).To(BeEmpty())
			})

			It("should have empty objectives by default", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Summary)
				Expect(m.GetObjectives()).To(BeEmpty())
			})

			It("should have nil collect function by default", func() {
				m := prmmet.NewMetrics("test_metric", prmtps.Counter)
				Expect(m.GetCollect()).To(BeNil())
			})
		})
	})

	Describe("GetName", func() {
		It("should return the correct metric name", func() {
			m := prmmet.NewMetrics("my_custom_metric", prmtps.Counter)
			Expect(m.GetName()).To(Equal("my_custom_metric"))
		})

		It("should handle empty name", func() {
			m := prmmet.NewMetrics("", prmtps.Counter)
			Expect(m.GetName()).To(BeEmpty())
		})

		It("should handle special characters in name", func() {
			m := prmmet.NewMetrics("test_metric_123", prmtps.Counter)
			Expect(m.GetName()).To(Equal("test_metric_123"))
		})
	})

	Describe("GetType", func() {
		It("should return Counter type", func() {
			m := prmmet.NewMetrics("test", prmtps.Counter)
			Expect(m.GetType()).To(Equal(prmtps.Counter))
		})

		It("should return Gauge type", func() {
			m := prmmet.NewMetrics("test", prmtps.Gauge)
			Expect(m.GetType()).To(Equal(prmtps.Gauge))
		})

		It("should return Histogram type", func() {
			m := prmmet.NewMetrics("test", prmtps.Histogram)
			Expect(m.GetType()).To(Equal(prmtps.Histogram))
		})

		It("should return Summary type", func() {
			m := prmmet.NewMetrics("test", prmtps.Summary)
			Expect(m.GetType()).To(Equal(prmtps.Summary))
		})

		It("should return None type", func() {
			m := prmmet.NewMetrics("test", prmtps.None)
			Expect(m.GetType()).To(Equal(prmtps.None))
		})
	})

	Describe("SetDesc and GetDesc", func() {
		It("should set and get description", func() {
			m := prmmet.NewMetrics("test_metric", prmtps.Counter)
			desc := "This is a test metric description"
			m.SetDesc(desc)
			Expect(m.GetDesc()).To(Equal(desc))
		})

		It("should allow updating description", func() {
			m := prmmet.NewMetrics("test_metric", prmtps.Counter)
			m.SetDesc("Initial description")
			Expect(m.GetDesc()).To(Equal("Initial description"))

			m.SetDesc("Updated description")
			Expect(m.GetDesc()).To(Equal("Updated description"))
		})

		It("should handle empty description", func() {
			m := prmmet.NewMetrics("test_metric", prmtps.Counter)
			m.SetDesc("")
			Expect(m.GetDesc()).To(BeEmpty())
		})

		It("should handle long descriptions", func() {
			m := prmmet.NewMetrics("test_metric", prmtps.Counter)
			longDesc := "This is a very long description that contains many words and should still work properly with the metric system without any issues whatsoever"
			m.SetDesc(longDesc)
			Expect(m.GetDesc()).To(Equal(longDesc))
		})

		It("should handle special characters in description", func() {
			m := prmmet.NewMetrics("test_metric", prmtps.Counter)
			desc := "Test with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
			m.SetDesc(desc)
			Expect(m.GetDesc()).To(Equal(desc))
		})
	})
})
