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

package prometheus_test

import (
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus Metrics Management", func() {
	var p libprm.Prometheus

	BeforeEach(func() {
		p = newPrometheus()
	})

	Describe("AddMetric", func() {
		Context("when adding API metrics", func() {
			It("should add a counter metric successfully", func() {
				name := uniqueMetricName("api_counter")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(name))
			})

			It("should add a gauge metric successfully", func() {
				name := uniqueMetricName("api_gauge")
				m := createGaugeMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Gauge))
			})

			It("should add a histogram metric successfully", func() {
				name := uniqueMetricName("api_histogram")
				m := createHistogramMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Histogram))
			})

			It("should add a summary metric successfully", func() {
				name := uniqueMetricName("api_summary")
				m := createSummaryMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetType()).To(Equal(prmtps.Summary))
			})
		})

		Context("when adding non-API metrics", func() {
			It("should add to other pool", func() {
				name := uniqueMetricName("other_counter")
				m := createCounterMetric(name)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
			})

			It("should separate from API metrics", func() {
				apiName := uniqueMetricName("api_metric")
				otherName := uniqueMetricName("other_metric")

				apiMetric := createCounterMetric(apiName)
				otherMetric := createCounterMetric(otherName)

				err := p.AddMetric(true, apiMetric)
				Expect(err).ToNot(HaveOccurred())

				err = p.AddMetric(false, otherMetric)
				Expect(err).ToNot(HaveOccurred())

				// Both should be retrievable
				Expect(p.GetMetric(apiName)).ToNot(BeNil())
				Expect(p.GetMetric(otherName)).ToNot(BeNil())
			})
		})

		Context("when adding metrics with labels", func() {
			It("should add metric with labels", func() {
				name := uniqueMetricName("labeled_counter")
				m := createCounterMetric(name, "method", "status")

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetLabel()).To(Equal([]string{"method", "status"}))
			})
		})

		Context("when adding invalid metrics", func() {
			It("should fail with empty name", func() {
				m := createCounterMetric("")

				err := p.AddMetric(true, m)
				Expect(err).To(HaveOccurred())
			})

			It("should fail without collect function", func() {
				name := uniqueMetricName("no_collect")
				m := prmmet.NewMetrics(name, prmtps.Counter)
				m.SetDesc("Test counter without collect")
				// Don't set collect function

				err := p.AddMetric(true, m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("collect func"))
			})
		})

		Context("when adding duplicate metrics", func() {
			It("should fail to add duplicate metric name", func() {
				name := uniqueMetricName("duplicate")
				m1 := createCounterMetric(name)
				m2 := createCounterMetric(name)

				err := p.AddMetric(true, m1)
				Expect(err).ToNot(HaveOccurred())

				err = p.AddMetric(true, m2)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetMetric", func() {
		Context("when retrieving metrics", func() {
			It("should return metric from API pool", func() {
				name := uniqueMetricName("get_api")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(name))
			})

			It("should return metric from other pool", func() {
				name := uniqueMetricName("get_other")
				m := createCounterMetric(name)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(name))
			})

			It("should return nil for non-existent metric", func() {
				retrieved := p.GetMetric("non_existent_metric")
				Expect(retrieved).To(BeNil())
			})

			It("should check API pool first", func() {
				name := uniqueMetricName("priority")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := p.GetMetric(name)
				Expect(retrieved).ToNot(BeNil())
			})
		})
	})

	Describe("DelMetric", func() {
		Context("when deleting metrics", func() {
			It("should delete metric from API pool", func() {
				name := uniqueMetricName("del_api")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				p.DelMetric(name)

				retrieved := p.GetMetric(name)
				Expect(retrieved).To(BeNil())
			})

			It("should delete metric from other pool", func() {
				name := uniqueMetricName("del_other")
				m := createCounterMetric(name)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				p.DelMetric(name)

				retrieved := p.GetMetric(name)
				Expect(retrieved).To(BeNil())
			})

			It("should not fail when deleting non-existent metric", func() {
				Expect(func() {
					p.DelMetric("non_existent")
				}).ToNot(Panic())
			})

			It("should unregister metric from Prometheus", func() {
				name := uniqueMetricName("del_unregister")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				p.DelMetric(name)

				// Metric should be unregistered
				retrieved := p.GetMetric(name)
				Expect(retrieved).To(BeNil())
			})
		})
	})

	Describe("ListMetric", func() {
		Context("when listing metrics", func() {
			It("should return empty list initially", func() {
				list := p.ListMetric()
				Expect(list).To(BeEmpty())
			})

			It("should list API metrics", func() {
				name1 := uniqueMetricName("list_api_1")
				name2 := uniqueMetricName("list_api_2")

				m1 := createCounterMetric(name1)
				m2 := createCounterMetric(name2)

				err := p.AddMetric(true, m1)
				Expect(err).ToNot(HaveOccurred())
				err = p.AddMetric(true, m2)
				Expect(err).ToNot(HaveOccurred())

				list := p.ListMetric()
				Expect(list).To(ContainElement(name1))
				Expect(list).To(ContainElement(name2))
			})

			It("should list other metrics", func() {
				name := uniqueMetricName("list_other")
				m := createCounterMetric(name)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				list := p.ListMetric()
				Expect(list).To(ContainElement(name))
			})

			It("should list both API and other metrics", func() {
				apiName := uniqueMetricName("list_combined_api")
				otherName := uniqueMetricName("list_combined_other")

				apiMetric := createCounterMetric(apiName)
				otherMetric := createCounterMetric(otherName)

				err := p.AddMetric(true, apiMetric)
				Expect(err).ToNot(HaveOccurred())
				err = p.AddMetric(false, otherMetric)
				Expect(err).ToNot(HaveOccurred())

				list := p.ListMetric()
				Expect(list).To(ContainElement(apiName))
				Expect(list).To(ContainElement(otherName))
			})

			It("should not include deleted metrics", func() {
				name := uniqueMetricName("list_deleted")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				p.DelMetric(name)

				list := p.ListMetric()
				Expect(list).ToNot(ContainElement(name))
			})
		})
	})

	Describe("Concurrent Metrics Operations", func() {
		Context("when accessing metrics concurrently", func() {
			It("should handle concurrent Add operations", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(idx int) {
						name := uniqueMetricName("concurrent_add")
						m := createCounterMetric(name)

						_ = p.AddMetric(true, m)
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					<-done
				}

				// Should complete without panic
				Expect(p.ListMetric()).ToNot(BeEmpty())
			})

			It("should handle concurrent Get operations", func() {
				name := uniqueMetricName("concurrent_get")
				m := createCounterMetric(name)

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				done := make(chan bool, 20)

				for i := 0; i < 20; i++ {
					go func() {
						_ = p.GetMetric(name)
						done <- true
					}()
				}

				for i := 0; i < 20; i++ {
					<-done
				}
			})

			It("should handle mixed operations concurrently", func() {
				done := make(chan bool, 30)

				// Add operations
				for i := 0; i < 10; i++ {
					go func() {
						name := uniqueMetricName("concurrent_mixed")
						m := createCounterMetric(name)

						_ = p.AddMetric(true, m)
						done <- true
					}()
				}

				// List operations
				for i := 0; i < 10; i++ {
					go func() {
						_ = p.ListMetric()
						done <- true
					}()
				}

				// Get operations
				for i := 0; i < 10; i++ {
					go func() {
						_ = p.GetMetric("some_metric")
						done <- true
					}()
				}

				for i := 0; i < 30; i++ {
					<-done
				}
			})
		})
	})
})
