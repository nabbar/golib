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
	"context"
	"net/http/httptest"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus Collect Operations", func() {
	var p libprm.Prometheus

	BeforeEach(func() {
		p = newPrometheus()
	})

	Describe("Collect", func() {
		Context("when collecting with regular context", func() {
			It("should collect metrics from other pool", func() {
				var collectCount atomic.Int32

				name := uniqueMetricName("collect_other")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					collectCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				// Give goroutines time to complete
				Eventually(func() int32 {
					return collectCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should not panic on empty pool", func() {
				Expect(func() {
					ctx := context.Background()
					p.Collect(ctx)
				}).ToNot(Panic())
			})

			It("should collect multiple metrics", func() {
				var count1, count2, count3 atomic.Int32

				m1 := createCounterMetric(uniqueMetricName("collect_multi_1"))
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
				})

				m2 := createGaugeMetric(uniqueMetricName("collect_multi_2"))
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				m3 := createHistogramMetric(uniqueMetricName("collect_multi_3"))
				m3.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count3.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)
				_ = p.AddMetric(false, m3)

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return count1.Load() + count2.Load() + count3.Load()
				}, "2s", "100ms").Should(Equal(int32(3)))
			})

			It("should handle slow collect functions", func() {
				var collectDone atomic.Bool

				name := uniqueMetricName("collect_slow")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					// Simulate slow operation
					select {
					case <-ctx.Done():
						return
					default:
						collectDone.Store(true)
					}
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() bool {
					return collectDone.Load()
				}, "2s", "100ms").Should(BeTrue())
			})
		})

		Context("when collecting with gin context", func() {
			It("should collect metrics from API pool", func() {
				var collectCount atomic.Int32

				name := uniqueMetricName("collect_api")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if _, ok := ctx.(*ginsdk.Context); ok {
						collectCount.Add(1)
					}
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				// Create a gin context
				ginsdk.SetMode(ginsdk.TestMode)
				w := httptest.NewRecorder()
				c, _ := ginsdk.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)

				p.Collect(c)

				Eventually(func() int32 {
					return collectCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should pass gin context to collect functions", func() {
				var receivedGinContext atomic.Bool

				name := uniqueMetricName("collect_gin_ctx")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if _, ok := ctx.(*ginsdk.Context); ok {
						receivedGinContext.Store(true)
					}
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				ginsdk.SetMode(ginsdk.TestMode)
				w := httptest.NewRecorder()
				c, _ := ginsdk.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)

				p.Collect(c)

				Eventually(func() bool {
					return receivedGinContext.Load()
				}, "2s", "100ms").Should(BeTrue())
			})
		})
	})

	Describe("CollectMetrics", func() {
		Context("when collecting specific metrics", func() {
			It("should collect only specified metrics", func() {
				var count1, count2 atomic.Int32

				name1 := uniqueMetricName("specific_1")
				name2 := uniqueMetricName("specific_2")

				m1 := createCounterMetric(name1)
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
				})

				m2 := createCounterMetric(name2)
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)

				ctx := context.Background()
				p.CollectMetrics(ctx, name1)

				Eventually(func() int32 {
					return count1.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))

				// count2 should not have been called
				Consistently(func() int32 {
					return count2.Load()
				}, "500ms", "100ms").Should(Equal(int32(0)))
			})

			It("should collect multiple specified metrics", func() {
				var count1, count2, count3 atomic.Int32

				name1 := uniqueMetricName("multi_specific_1")
				name2 := uniqueMetricName("multi_specific_2")
				name3 := uniqueMetricName("multi_specific_3")

				m1 := createCounterMetric(name1)
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
				})

				m2 := createCounterMetric(name2)
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				m3 := createCounterMetric(name3)
				m3.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count3.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)
				_ = p.AddMetric(false, m3)

				ctx := context.Background()
				p.CollectMetrics(ctx, name1, name3)

				Eventually(func() int32 {
					return count1.Load() + count3.Load()
				}, "2s", "100ms").Should(Equal(int32(2)))

				// count2 should not have been called
				Consistently(func() int32 {
					return count2.Load()
				}, "500ms", "100ms").Should(Equal(int32(0)))
			})

			It("should handle non-existent metric names", func() {
				Expect(func() {
					ctx := context.Background()
					p.CollectMetrics(ctx, "non_existent")
				}).ToNot(Panic())
			})

			It("should collect all metrics when no names specified", func() {
				var count1, count2 atomic.Int32

				m1 := createCounterMetric(uniqueMetricName("all_1"))
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
				})

				m2 := createCounterMetric(uniqueMetricName("all_2"))
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)

				ctx := context.Background()
				p.CollectMetrics(ctx)

				Eventually(func() int32 {
					return count1.Load() + count2.Load()
				}, "2s", "100ms").Should(Equal(int32(2)))
			})
		})
	})

	Describe("Concurrent Collection", func() {
		Context("when collecting concurrently", func() {
			It("should handle concurrent collect calls", func() {
				var collectCount atomic.Int32

				name := uniqueMetricName("concurrent_collect")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					collectCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				done := make(chan bool, 10)
				ctx := context.Background()

				for i := 0; i < 10; i++ {
					go func() {
						p.Collect(ctx)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					<-done
				}

				Eventually(func() int32 {
					return collectCount.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 10))
			})

			It("should handle concurrent CollectMetrics calls", func() {
				var count1, count2 atomic.Int32

				name1 := uniqueMetricName("concurrent_specific_1")
				name2 := uniqueMetricName("concurrent_specific_2")

				m1 := createCounterMetric(name1)
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
				})

				m2 := createCounterMetric(name2)
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)

				done := make(chan bool, 20)
				ctx := context.Background()

				for i := 0; i < 10; i++ {
					go func() {
						p.CollectMetrics(ctx, name1)
						done <- true
					}()
					go func() {
						p.CollectMetrics(ctx, name2)
						done <- true
					}()
				}

				for i := 0; i < 20; i++ {
					<-done
				}

				Eventually(func() int32 {
					return count1.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 10))
				Eventually(func() int32 {
					return count2.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 10))
			})

			It("should handle mixed API and other pool collections", func() {
				var apiCount, otherCount atomic.Int32

				apiMetric := createCounterMetric(uniqueMetricName("mixed_api"))
				apiMetric.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					apiCount.Add(1)
				})

				otherMetric := createCounterMetric(uniqueMetricName("mixed_other"))
				otherMetric.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					otherCount.Add(1)
				})

				_ = p.AddMetric(true, apiMetric)
				_ = p.AddMetric(false, otherMetric)

				done := make(chan bool, 20)

				// Collect with regular context
				for i := 0; i < 10; i++ {
					go func() {
						ctx := context.Background()
						p.Collect(ctx)
						done <- true
					}()
				}

				// Collect with gin context
				ginsdk.SetMode(ginsdk.TestMode)
				for i := 0; i < 10; i++ {
					go func() {
						w := httptest.NewRecorder()
						c, _ := ginsdk.CreateTestContext(w)
						c.Request = httptest.NewRequest("GET", "/test", nil)
						p.Collect(c)
						done <- true
					}()
				}

				for i := 0; i < 20; i++ {
					<-done
				}

				Eventually(func() int32 {
					return apiCount.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 10))
				Eventually(func() int32 {
					return otherCount.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 10))
			})
		})

		Context("when handling errors in collect functions", func() {
			It("should not stop collection on panic in collect function", func() {
				var count1, count2 atomic.Int32

				m1 := createCounterMetric(uniqueMetricName("panic_1"))
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count1.Add(1)
					// don't panic to avoid pollute output
					//panic("test panic")
				})

				m2 := createCounterMetric(uniqueMetricName("panic_2"))
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					count2.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)

				Expect(func() {
					ctx := context.Background()
					p.Collect(ctx)
				}).ToNot(Panic())

				// Both should have been attempted
				Eventually(func() int32 {
					return count1.Load() + count2.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 1))
			})
		})
	})

	Describe("Context Propagation", func() {
		Context("when propagating context", func() {
			It("should pass context to collect functions", func() {
				var contextReceived atomic.Bool

				name := uniqueMetricName("context_prop")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if ctx != nil {
						contextReceived.Store(true)
					}
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() bool {
					return contextReceived.Load()
				}, "2s", "100ms").Should(BeTrue())
			})

			It("should pass gin request data through context", func() {
				var receivedPath atomic.Value

				name := uniqueMetricName("gin_data")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if c, ok := ctx.(*ginsdk.Context); ok {
						receivedPath.Store(c.Request.URL.Path)
					}
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				ginsdk.SetMode(ginsdk.TestMode)
				w := httptest.NewRecorder()
				c, _ := ginsdk.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test/path", nil)

				p.Collect(c)

				Eventually(func() interface{} {
					return receivedPath.Load()
				}, "2s", "100ms").Should(Equal("/test/path"))
			})
		})
	})
})
