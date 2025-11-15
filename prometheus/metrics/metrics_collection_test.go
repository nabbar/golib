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
	"context"
	"sync/atomic"
	"time"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics Collection", func() {
	Describe("SetCollect and GetCollect", func() {
		Context("when setting collect function", func() {
			It("should set and get collect function", func() {
				m := newCounterMetric("test_collect_set_get", "method")
				called := atomic.Bool{}

				collectFunc := func(ctx context.Context, metric prmmet.Metric) {
					called.Store(true)
				}

				m.SetCollect(collectFunc)
				retrievedFunc := m.GetCollect()

				Expect(retrievedFunc).ToNot(BeNil())

				// Test that the function works
				ctx := context.Background()
				retrievedFunc(ctx, m)
				Expect(called.Load()).To(BeTrue())
			})

			It("should allow updating collect function", func() {
				m := newCounterMetric("test_collect_update", "method")
				call1 := atomic.Bool{}
				call2 := atomic.Bool{}

				func1 := func(ctx context.Context, metric prmmet.Metric) {
					call1.Store(true)
				}

				func2 := func(ctx context.Context, metric prmmet.Metric) {
					call2.Store(true)
				}

				m.SetCollect(func1)
				m.GetCollect()(context.Background(), m)
				Expect(call1.Load()).To(BeTrue())

				m.SetCollect(func2)
				m.GetCollect()(context.Background(), m)
				Expect(call2.Load()).To(BeTrue())
			})

			It("should handle nil function", func() {
				m := newCounterMetric("test_collect_nil", "method")
				m.SetCollect(nil)
				f := m.GetCollect()
				Expect(f).ToNot(BeNil())
			})
		})

		Context("when collect function is not set", func() {
			It("should return nil", func() {
				m := newCounterMetric("test_collect_not_set", "method")
				Expect(m.GetCollect()).To(BeNil())
			})
		})
	})

	Describe("Collect", func() {
		Context("with collect function set", func() {
			It("should call collect function on Collect", func() {
				m := newCounterMetric("test_collect_invoke", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				called := atomic.Bool{}
				var capturedMetric prmmet.Metric

				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					called.Store(true)
					capturedMetric = metric
					// Perform some metric operation
					_ = metric.Inc([]string{"GET"})
				})

				ctx := context.Background()
				m.Collect(ctx)

				Expect(called.Load()).To(BeTrue())
				Expect(capturedMetric).To(Equal(m))
			})

			It("should pass context to collect function", func() {
				m := newCounterMetric("test_collect_context", "method")
				var capturedCtx context.Context

				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					capturedCtx = ctx
				})

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				m.Collect(ctx)
				Expect(capturedCtx).ToNot(BeNil())
				Expect(capturedCtx).To(Equal(ctx))
			})

			It("should allow metric updates in collect function", func() {
				m := newGaugeMetric("test_collect_metric_update", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				collectCount := atomic.Int32{}

				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count := collectCount.Add(1)
					_ = metric.SetGaugeValue([]string{"GET"}, float64(count))
				})

				m.Collect(context.Background())
				Expect(collectCount.Load()).To(Equal(int32(1)))

				m.Collect(context.Background())
				Expect(collectCount.Load()).To(Equal(int32(2)))

				m.Collect(context.Background())
				Expect(collectCount.Load()).To(Equal(int32(3)))
			})

			It("should work with different metric types", func() {
				testCases := []struct {
					name       string
					metricType prmtps.MetricType
					setup      func(prmmet.Metric)
				}{
					{
						name:       "counter",
						metricType: prmtps.Counter,
						setup: func(m prmmet.Metric) {
							m.AddLabel("method")
						},
					},
					{
						name:       "gauge",
						metricType: prmtps.Gauge,
						setup: func(m prmmet.Metric) {
							m.AddLabel("method")
						},
					},
					{
						name:       "histogram",
						metricType: prmtps.Histogram,
						setup: func(m prmmet.Metric) {
							m.AddLabel("method")
							m.AddBuckets(0.1, 0.5, 1.0)
						},
					},
					{
						name:       "summary",
						metricType: prmtps.Summary,
						setup: func(m prmmet.Metric) {
							m.AddLabel("method")
							m.AddObjective(0.5, 0.05)
							m.AddObjective(0.9, 0.01)
						},
					},
				}

				for _, tc := range testCases {
					m := prmmet.NewMetrics("test_collect_type_"+tc.name, tc.metricType)
					m.SetDesc("Test metric for " + tc.name)
					tc.setup(m)

					called := atomic.Bool{}
					m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
						called.Store(true)
					})

					if tc.metricType != prmtps.None {
						Expect(registerMetric(m)).ToNot(HaveOccurred())
					}

					m.Collect(context.Background())
					Expect(called.Load()).To(BeTrue(), "Collect should be called for "+tc.name)

					if tc.metricType != prmtps.None {
						cleanupMetric(m)
					}
				}
			})
		})

		Context("without collect function set", func() {
			It("should not panic when calling Collect", func() {
				m := newCounterMetric("test_collect_no_func", "method")

				Expect(func() {
					m.Collect(context.Background())
				}).ToNot(Panic())
			})

			It("should be a no-op", func() {
				m := newCounterMetric("test_collect_noop", "method")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				// Add some value
				Expect(m.Inc([]string{"GET"})).ToNot(HaveOccurred())

				// Calling Collect without function should do nothing
				m.Collect(context.Background())

				// Should still be able to use the metric
				Expect(m.Inc([]string{"POST"})).ToNot(HaveOccurred())
			})
		})

		Context("with context cancellation", func() {
			It("should handle canceled context", func() {
				m := newCounterMetric("test_collect_canceled_context", "method")

				contextReceived := atomic.Bool{}
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					contextReceived.Store(true)
					// Check if context is already canceled
					select {
					case <-ctx.Done():
						// Context is canceled
					default:
						// Context is not canceled
					}
				})

				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				m.Collect(ctx)
				Expect(contextReceived.Load()).To(BeTrue())
			})

			It("should handle timeout context", func() {
				m := newCounterMetric("test_collect_timeout_context", "method")

				contextReceived := atomic.Bool{}
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					contextReceived.Store(true)
					deadline, ok := ctx.Deadline()
					Expect(ok).To(BeTrue())
					Expect(deadline).ToNot(BeZero())
				})

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				m.Collect(ctx)
				Expect(contextReceived.Load()).To(BeTrue())
			})
		})
	})

	Describe("Collect function patterns", func() {
		Context("common collection patterns", func() {
			It("should support periodic metric updates", func() {
				m := newGaugeMetric("test_collect_periodic", "resource")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				callCount := atomic.Int32{}
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count := callCount.Add(1)
					// Simulate collecting resource metrics
					_ = metric.SetGaugeValue([]string{"cpu"}, float64(count*10))
					_ = metric.SetGaugeValue([]string{"memory"}, float64(count*20))
				})

				// Simulate periodic collection
				for i := 0; i < 5; i++ {
					m.Collect(context.Background())
				}
				Expect(callCount.Load()).To(Equal(int32(5)))
			})

			It("should support conditional metric collection", func() {
				m := newCounterMetric("test_collect_conditional", "event")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				shouldCollect := atomic.Bool{}
				shouldCollect.Store(false)

				collectedCount := atomic.Int32{}
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if shouldCollect.Load() {
						collectedCount.Add(1)
						_ = metric.Inc([]string{"processed"})
					}
				})

				// Should not collect
				m.Collect(context.Background())
				Expect(collectedCount.Load()).To(Equal(int32(0)))

				// Enable collection
				shouldCollect.Store(true)
				m.Collect(context.Background())
				Expect(collectedCount.Load()).To(Equal(int32(1)))
			})

			It("should support aggregated metrics collection", func() {
				m := newHistogramMetric("test_collect_aggregated", []float64{0.1, 0.5, 1.0, 5.0}, "operation")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				observations := []float64{0.05, 0.15, 0.45, 0.75, 1.5, 2.5, 4.5}
				currentIndex := atomic.Int32{}

				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					// Collect next observation
					idx := int(currentIndex.Add(1)) - 1
					if idx < len(observations) {
						_ = metric.Observe([]string{"api_call"}, observations[idx])
					}
				})

				// Collect all observations
				for i := 0; i < len(observations); i++ {
					m.Collect(context.Background())
				}
				Expect(currentIndex.Load()).To(Equal(int32(len(observations))))
			})

			It("should support error handling in collect function", func() {
				m := newCounterMetric("test_collect_error_handling", "request")
				Expect(registerMetric(m)).ToNot(HaveOccurred())
				defer cleanupMetric(m)

				errorCount := atomic.Int32{}
				successCount := atomic.Int32{}

				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					// Simulate some operation that might fail
					if time.Now().UnixNano()%2 == 0 {
						successCount.Add(1)
						_ = metric.Inc([]string{"success"})
					} else {
						errorCount.Add(1)
						_ = metric.Inc([]string{"error"})
					}
				})

				// Collect multiple times
				for i := 0; i < 10; i++ {
					m.Collect(context.Background())
				}

				total := errorCount.Load() + successCount.Load()
				Expect(total).To(Equal(int32(10)))
			})
		})
	})
})
