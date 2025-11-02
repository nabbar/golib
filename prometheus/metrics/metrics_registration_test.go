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
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Metrics Registration", func() {
	Describe("Register", func() {
		Context("with Counter metric", func() {
			It("should register successfully", func() {
				m := newCounterMetric("test_counter_register", "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(vec).ToNot(BeNil())

				err = m.Register(testRegistry, vec)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should allow re-registration of same metric", func() {
				m := newCounterMetric("test_counter_reregister", "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				// Re-register should work (it unregisters first internally)
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())
			})

			It("should work with labels", func() {
				m := newCounterMetric("test_counter_with_labels", "method", "status")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())

				err = m.Register(testRegistry, vec)
				Expect(err).ToNot(HaveOccurred())

				// Should be able to use the metric
				Expect(m.Inc([]string{"GET", "200"})).ToNot(HaveOccurred())
			})
		})

		Context("with Gauge metric", func() {
			It("should register successfully", func() {
				m := newGaugeMetric("test_gauge_register", "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(vec).ToNot(BeNil())

				err = m.Register(testRegistry, vec)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should work after registration", func() {
				m := newGaugeMetric("test_gauge_functional", "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				Expect(m.SetGaugeValue([]string{"GET"}, 42.0)).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"POST"})).ToNot(HaveOccurred())
				Expect(m.Add([]string{"DELETE"}, 10.5)).ToNot(HaveOccurred())
			})
		})

		Context("with Histogram metric", func() {
			It("should register successfully with buckets", func() {
				m := newHistogramMetric("test_histogram_register", []float64{0.1, 0.5, 1.0}, "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(vec).ToNot(BeNil())

				err = m.Register(testRegistry, vec)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail registration without buckets", func() {
				m := newMetricWithRegistration("test_histogram_no_buckets", prmtps.Histogram)
				m.SetDesc("Histogram without buckets")
				m.AddLabel("method")

				vec, err := m.GetType().Register(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot lose bucket param"))
				Expect(vec).To(BeNil())
			})

			It("should work after registration", func() {
				m := newHistogramMetric("test_histogram_functional", prmsdk.DefBuckets, "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"POST"}, 1.5)).ToNot(HaveOccurred())
			})
		})

		Context("with Summary metric", func() {
			It("should register successfully with objectives", func() {
				objectives := map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.99: 0.001,
				}
				m := newSummaryMetric("test_summary_register", objectives, "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(vec).ToNot(BeNil())

				err = m.Register(testRegistry, vec)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should fail registration without objectives", func() {
				m := newMetricWithRegistration("test_summary_no_objectives", prmtps.Summary)
				m.SetDesc("Summary without objectives")
				m.AddLabel("method")

				vec, err := m.GetType().Register(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot lose objectives param"))
				Expect(vec).To(BeNil())
			})

			It("should work after registration", func() {
				m := newSummaryMetric("test_summary_functional", nil, "method")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"POST"}, 1.5)).ToNot(HaveOccurred())
			})
		})

		Context("with None metric type", func() {
			It("should fail to register None type", func() {
				m := newMetricWithRegistration("test_none_register", prmtps.None)
				m.SetDesc("None type metric")

				vec, err := m.GetType().Register(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not compatible"))
				Expect(vec).To(BeNil())
			})
		})

		Context("with metrics without labels", func() {
			It("should register counter without labels", func() {
				m := newCounterMetric("test_counter_no_labels")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				// Should be able to use with empty label values
				Expect(m.Inc([]string{})).ToNot(HaveOccurred())
			})

			It("should register gauge without labels", func() {
				m := newGaugeMetric("test_gauge_no_labels")
				defer cleanupMetric(m)

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				Expect(m.SetGaugeValue([]string{}, 100.0)).ToNot(HaveOccurred())
			})
		})
	})

	Describe("UnRegister", func() {
		Context("with registered metrics", func() {
			It("should unregister counter successfully", func() {
				m := newCounterMetric("test_counter_unregister", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())

				result := m.UnRegister(testRegistry)
				Expect(result).ToNot(HaveOccurred())
			})

			It("should unregister gauge successfully", func() {
				m := newGaugeMetric("test_gauge_unregister", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())

				result := m.UnRegister(testRegistry)
				Expect(result).ToNot(HaveOccurred())
			})

			It("should unregister histogram successfully", func() {
				m := newHistogramMetric("test_histogram_unregister", []float64{0.1, 0.5, 1.0}, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())

				result := m.UnRegister(testRegistry)
				Expect(result).ToNot(HaveOccurred())
			})

			It("should unregister summary successfully", func() {
				m := newSummaryMetric("test_summary_unregister", nil, "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())

				result := m.UnRegister(testRegistry)
				Expect(result).ToNot(HaveOccurred())
			})

			It("should allow re-registration after unregister", func() {
				m := newCounterMetric("test_counter_rereg_after_unreg", "method")

				vec, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec)).ToNot(HaveOccurred())

				Expect(m.UnRegister(testRegistry)).ToNot(HaveOccurred())

				// Re-register should work
				vec2, err := m.GetType().Register(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m.Register(testRegistry, vec2)).ToNot(HaveOccurred())

				cleanupMetric(m)
			})
		})

		Context("with unregistered metrics", func() {
			It("should return false for already unregistered metric", func() {
				m := newCounterMetric("test_counter_double_unreg", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())

				Expect(m.UnRegister(testRegistry)).ToNot(HaveOccurred())

				// Second unregister should return false
				Expect(m.UnRegister(testRegistry)).To(HaveOccurred())

			})

			It("should not panic on unregister without registration", func() {
				m := newCounterMetric("test_counter_unreg_never_reg", "method")

				// Should not panic, just return false
				result := m.UnRegister(testRegistry)
				Expect(result).To(HaveOccurred())
			})
		})

		Context("with multiple metrics lifecycle", func() {
			It("should handle register-unregister-register cycle", func() {
				name := "test_counter_lifecycle"

				// First registration
				m1 := newCounterMetric(name, "method")
				Expect(registerMetric(m1)).ToNot(HaveOccurred())
				Expect(m1.Inc([]string{"GET"})).ToNot(HaveOccurred())
				Expect(m1.UnRegister(testRegistry)).ToNot(HaveOccurred())

				// Second registration with same name
				m2 := newCounterMetric(name, "method")
				Expect(registerMetric(m2)).ToNot(HaveOccurred())
				Expect(m2.Inc([]string{"POST"})).ToNot(HaveOccurred())
				cleanupMetric(m2)
			})

			It("should maintain independence between different metrics", func() {
				m1 := newCounterMetric("test_counter_independent_1", "method")
				m2 := newCounterMetric("test_counter_independent_2", "method")

				Expect(registerMetric(m1)).ToNot(HaveOccurred())
				Expect(registerMetric(m2)).ToNot(HaveOccurred())

				Expect(m1.Inc([]string{"GET"})).ToNot(HaveOccurred())
				Expect(m2.Inc([]string{"POST"})).ToNot(HaveOccurred())

				Expect(m1.UnRegister(testRegistry)).ToNot(HaveOccurred())
				// m2 should still work
				Expect(m2.Inc([]string{"DELETE"})).ToNot(HaveOccurred())

				cleanupMetric(m2)
			})
		})
	})

	Describe("Registration edge cases", func() {
		Context("with special metric configurations", func() {
			It("should handle histogram with single bucket", func() {
				m := newHistogramMetric("test_histogram_single_bucket", []float64{1.0}, "method")
				defer cleanupMetric(m)

				Expect(registerMetric(m)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 0.5)).ToNot(HaveOccurred())
			})

			It("should handle histogram with many buckets", func() {
				buckets := make([]float64, 100)
				for i := range buckets {
					buckets[i] = float64(i + 1)
				}
				m := newHistogramMetric("test_histogram_many_buckets", buckets, "method")
				defer cleanupMetric(m)

				Expect(registerMetric(m)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 50.0)).ToNot(HaveOccurred())
			})

			It("should handle summary with single objective", func() {
				objectives := map[float64]float64{0.5: 0.05}
				m := newSummaryMetric("test_summary_single_objective", objectives, "method")
				defer cleanupMetric(m)

				Expect(registerMetric(m)).ToNot(HaveOccurred())
				Expect(m.Observe([]string{"GET"}, 1.0)).ToNot(HaveOccurred())
			})

			It("should handle summary with many objectives", func() {
				objectives := map[float64]float64{
					0.1:    0.01,
					0.25:   0.01,
					0.5:    0.05,
					0.75:   0.01,
					0.9:    0.01,
					0.95:   0.005,
					0.99:   0.001,
					0.999:  0.0001,
					0.9999: 0.00001,
				}
				m := newSummaryMetric("test_summary_many_objectives", objectives, "method")
				defer cleanupMetric(m)

				Expect(registerMetric(m)).ToNot(HaveOccurred())
				for i := 0; i < 50; i++ {
					Expect(m.Observe([]string{"GET"}, float64(i))).ToNot(HaveOccurred())
				}
			})

			It("should handle metrics with many labels", func() {
				m := newCounterMetric("test_counter_many_labels",
					"label1", "label2", "label3", "label4", "label5",
					"label6", "label7", "label8", "label9", "label10")
				defer cleanupMetric(m)

				Expect(registerMetric(m)).ToNot(HaveOccurred())
				Expect(m.Inc([]string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"})).ToNot(HaveOccurred())
			})
		})
	})
})
