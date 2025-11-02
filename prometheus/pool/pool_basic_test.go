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

package pool_test

import (
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metric Pool Basic Operations", func() {
	Describe("Pool Creation", func() {
		Context("with valid context function", func() {
			It("should create a new pool successfully", func() {
				pool := newPool()
				Expect(pool).ToNot(BeNil())
			})

			It("should start with empty pool", func() {
				pool := newPool()
				list := pool.List()
				Expect(list).To(BeEmpty())
			})
		})
	})

	Describe("Add", func() {
		var pool prmpool.MetricPool

		BeforeEach(func() {
			pool = newPool()
		})

		Context("with valid metrics", func() {
			It("should add a counter metric", func() {
				m := createCounterMetric("test_counter", "method")
				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := pool.Get("test_counter")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal("test_counter"))
			})

			It("should add a gauge metric", func() {
				m := createGaugeMetric("test_gauge", "service")
				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := pool.Get("test_gauge")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Gauge))
			})

			It("should add a histogram metric", func() {
				m := createHistogramMetric("test_histogram", "endpoint")
				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := pool.Get("test_histogram")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Histogram))
			})

			It("should add a summary metric", func() {
				m := createSummaryMetric("test_summary", "handler")
				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := pool.Get("test_summary")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Summary))
			})

			It("should add multiple different metrics", func() {
				metrics := []struct {
					name       string
					metricType prmtps.MetricType
				}{
					{"counter1", prmtps.Counter},
					{"gauge1", prmtps.Gauge},
					{"histogram1", prmtps.Histogram},
					{"summary1", prmtps.Summary},
				}

				for _, tc := range metrics {
					m := addMetricToPool(pool, tc.name, tc.metricType, "label")
					Expect(m).ToNot(BeNil())
				}

				list := pool.List()
				Expect(list).To(HaveLen(4))
			})
		})

		Context("with invalid metrics", func() {
			It("should return error for metric without name", func() {
				m := createCounterMetric("", "method")
				err := pool.Add(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("metric name cannot be empty"))
			})

			It("should return error for metric without collect function", func() {
				m := prmmet.NewMetrics("test_no_collect", prmtps.Counter)
				m.SetDesc("Test metric")
				m.AddLabel("method")
				// Not setting collect function

				err := pool.Add(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("metric collect func cannot be empty"))
			})
		})

		Context("with duplicate metrics", func() {
			It("should return error when adding duplicate metric name", func() {
				m1 := createCounterMetric("duplicate_test", "method")
				err := pool.Add(m1)
				Expect(err).ToNot(HaveOccurred())

				m2 := createCounterMetric("duplicate_test", "method")
				err = pool.Add(m2)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Get", func() {
		var pool prmpool.MetricPool

		BeforeEach(func() {
			pool = newPool()
		})

		Context("with existing metrics", func() {
			It("should retrieve an existing metric", func() {
				original := addMetricToPool(pool, "get_test", prmtps.Counter, "method")

				retrieved := pool.Get("get_test")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(original.GetName()))
				Expect(retrieved.GetType()).To(Equal(original.GetType()))
			})

			It("should retrieve correct metric among multiple", func() {
				addMetricToPool(pool, "metric1", prmtps.Counter)
				addMetricToPool(pool, "metric2", prmtps.Gauge)
				target := addMetricToPool(pool, "metric3", prmtps.Histogram)
				addMetricToPool(pool, "metric4", prmtps.Summary)

				retrieved := pool.Get("metric3")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(target.GetName()))
				Expect(retrieved.GetType()).To(Equal(prmtps.Histogram))
			})
		})

		Context("with non-existing metrics", func() {
			It("should return nil for non-existing metric", func() {
				retrieved := pool.Get("does_not_exist")
				Expect(retrieved).To(BeNil())
			})

			It("should return nil for empty pool", func() {
				retrieved := pool.Get("any_name")
				Expect(retrieved).To(BeNil())
			})
		})
	})

	Describe("Set", func() {
		var pool prmpool.MetricPool

		BeforeEach(func() {
			pool = newPool()
		})

		Context("setting new metrics", func() {
			It("should set a new metric directly", func() {
				m := createCounterMetric("set_test", "method")
				pool.Set("set_test", m)

				retrieved := pool.Get("set_test")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal("set_test"))
			})

			It("should set metric with different key than metric name", func() {
				m := createGaugeMetric("original_name", "label")
				pool.Set("custom_key", m)

				retrieved := pool.Get("custom_key")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal("original_name"))

				byOriginalName := pool.Get("original_name")
				Expect(byOriginalName).To(BeNil())
			})
		})

		Context("replacing existing metrics", func() {
			It("should replace an existing metric", func() {
				original := createCounterMetric("replace_test", "method")
				pool.Set("replace_test", original)

				replacement := createGaugeMetric("new_metric", "label")
				pool.Set("replace_test", replacement)

				retrieved := pool.Get("replace_test")
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal("new_metric"))
				Expect(retrieved.GetType()).To(Equal(prmtps.Gauge))
			})
		})
	})

	Describe("Del", func() {
		var pool prmpool.MetricPool

		BeforeEach(func() {
			pool = newPool()
		})

		Context("deleting existing metrics", func() {
			It("should delete an existing metric", func() {
				addMetricToPool(pool, "delete_test", prmtps.Counter, "method")

				Expect(pool.Get("delete_test")).ToNot(BeNil())

				pool.Del("delete_test")

				Expect(pool.Get("delete_test")).To(BeNil())
			})

			It("should maintain other metrics after deletion", func() {
				n1 := uniqueMetricName("keep1")
				addMetricToPool(pool, n1, prmtps.Counter)

				n2 := uniqueMetricName("delete_me")
				addMetricToPool(pool, n2, prmtps.Gauge)

				n3 := uniqueMetricName("keep2")
				addMetricToPool(pool, n3, prmtps.Histogram)

				pool.Del(n2)

				Expect(pool.Get(n1)).ToNot(BeNil())
				Expect(pool.Get(n2)).To(BeNil())
				Expect(pool.Get(n3)).ToNot(BeNil())

				list := pool.List()
				Expect(list).To(HaveLen(2))
			})

			It("should delete multiple metrics sequentially", func() {
				n1 := uniqueMetricName("m1")
				addMetricToPool(pool, n1, prmtps.Counter)

				n2 := uniqueMetricName("m2")
				addMetricToPool(pool, n2, prmtps.Gauge)

				n3 := uniqueMetricName("m3")
				addMetricToPool(pool, n3, prmtps.Histogram)

				pool.Del(n1)
				pool.Del(n3)

				Expect(pool.Get(n1)).To(BeNil())
				Expect(pool.Get(n2)).ToNot(BeNil())
				Expect(pool.Get(n3)).To(BeNil())

				list := pool.List()
				Expect(list).To(HaveLen(1))
			})
		})

		Context("deleting non-existing metrics", func() {
			It("should not error when deleting non-existing metric", func() {
				Expect(func() {
					pool.Del("does_not_exist")
				}).ToNot(Panic())
			})

			It("should not affect pool when deleting from empty pool", func() {
				pool.Del("any_name")

				list := pool.List()
				Expect(list).To(BeEmpty())
			})
		})
	})
})
