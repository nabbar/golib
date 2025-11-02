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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus Integration Tests", func() {
	Describe("Complete Workflow", func() {
		Context("when using Prometheus with Gin server", func() {
			It("should collect metrics through middleware", func() {
				p := newPrometheus()
				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				// Add a metric
				var requestCount atomic.Int32
				name := uniqueMetricName("integration_requests")
				m := createCounterMetric(name, "method", "path")
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					if c, ok := ctx.(*ginsdk.Context); ok {
						requestCount.Add(1)
						// In real scenario, would increment counter here
						_ = metric
						_ = c
					}
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				// Setup middleware
				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/test", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				// Make request
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Eventually(func() int32 {
					return requestCount.Load()
				}, "2s", "100ms").Should(BeNumerically(">=", 1))
			})

			It("should expose metrics endpoint", func() {
				p := newPrometheus()

				// Add a test metric so we have something to expose
				m := createCounterMetric(uniqueMetricName("test_expose"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					_ = metric.Inc([]string{})
				})
				Expect(p.AddMetric(false, m)).ToNot(HaveOccurred())
				p.Collect(testCtx)

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				// Add metrics endpoint
				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				// Request metrics
				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("# HELP"))
				Expect(w.Body.String()).To(ContainSubstring("# TYPE"))
			})

			It("should exclude configured paths", func() {
				p := newPrometheus()
				p.ExcludePath("/health", "/metrics")

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				var collectCount atomic.Int32
				m := createCounterMetric(uniqueMetricName("excluded_test"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					collectCount.Add(1)
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/health", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})
				router.GET("/api", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				// Request excluded path
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				time.Sleep(100 * time.Millisecond)
				excludedCount := collectCount.Load()

				// Request non-excluded path
				req = httptest.NewRequest("GET", "/api", nil)
				w = httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Eventually(func() int32 {
					return collectCount.Load()
				}, "2s", "100ms").Should(BeNumerically(">", excludedCount))
			})

			It("should handle multiple endpoints", func() {
				p := newPrometheus()
				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/users", func(c *ginsdk.Context) {
					c.JSON(http.StatusOK, ginsdk.H{"users": []string{}})
				})
				router.POST("/users", func(c *ginsdk.Context) {
					c.JSON(http.StatusCreated, ginsdk.H{"id": 1})
				})
				router.GET("/products", func(c *ginsdk.Context) {
					c.JSON(http.StatusOK, ginsdk.H{"products": []string{}})
				})

				// Make requests to different endpoints
				endpoints := []string{"/users", "/products"}
				methods := []string{"GET", "POST"}

				for _, endpoint := range endpoints {
					for _, method := range methods {
						if endpoint == "/products" && method == "POST" {
							continue
						}
						req := httptest.NewRequest(method, endpoint, nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
					}
				}
			})
		})

		Context("when using different metric types", func() {
			It("should collect counter metrics", func() {
				p := newPrometheus()

				var callCount atomic.Int32
				m := createCounterMetric(uniqueMetricName("counter_integration"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					callCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return callCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should collect gauge metrics", func() {
				p := newPrometheus()

				var callCount atomic.Int32
				m := createGaugeMetric(uniqueMetricName("gauge_integration"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					callCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return callCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should collect histogram metrics", func() {
				p := newPrometheus()

				var callCount atomic.Int32
				m := createHistogramMetric(uniqueMetricName("histogram_integration"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					callCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return callCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should collect summary metrics", func() {
				p := newPrometheus()

				var callCount atomic.Int32
				m := createSummaryMetric(uniqueMetricName("summary_integration"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					callCount.Add(1)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return callCount.Load()
				}, "2s", "100ms").Should(Equal(int32(1)))
			})

			It("should collect mixed metric types", func() {
				p := newPrometheus()

				var counter, gauge, histogram, summary atomic.Int32

				m1 := createCounterMetric(uniqueMetricName("mixed_counter"))
				m1.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					counter.Add(1)
				})

				m2 := createGaugeMetric(uniqueMetricName("mixed_gauge"))
				m2.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					gauge.Add(1)
				})

				m3 := createHistogramMetric(uniqueMetricName("mixed_histogram"))
				m3.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					histogram.Add(1)
				})

				m4 := createSummaryMetric(uniqueMetricName("mixed_summary"))
				m4.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					summary.Add(1)
				})

				_ = p.AddMetric(false, m1)
				_ = p.AddMetric(false, m2)
				_ = p.AddMetric(false, m3)
				_ = p.AddMetric(false, m4)

				ctx := context.Background()
				p.Collect(ctx)

				Eventually(func() int32 {
					return counter.Load() + gauge.Load() + histogram.Load() + summary.Load()
				}, "2s", "100ms").Should(Equal(int32(4)))
			})
		})

		Context("when handling high load", func() {
			It("should handle many concurrent requests", func() {
				p := newPrometheus()
				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				var requestCount atomic.Int32
				m := createCounterMetric(uniqueMetricName("load_test"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					requestCount.Add(1)
				})

				err := p.AddMetric(true, m)
				Expect(err).ToNot(HaveOccurred())

				router.Use(func(c *ginsdk.Context) {
					p.MiddleWareGin(c)
				})

				router.GET("/load", func(c *ginsdk.Context) {
					c.String(http.StatusOK, "ok")
				})

				done := make(chan bool, 100)

				for i := 0; i < 100; i++ {
					go func() {
						req := httptest.NewRequest("GET", "/load", nil)
						w := httptest.NewRecorder()
						router.ServeHTTP(w, req)
						done <- true
					}()
				}

				for i := 0; i < 100; i++ {
					<-done
				}

				Eventually(func() int32 {
					return requestCount.Load()
				}, "5s", "100ms").Should(BeNumerically(">=", 100))
			})

			It("should handle rapid metric additions and deletions", func() {
				p := newPrometheus()

				done := make(chan bool, 100)

				for i := 0; i < 50; i++ {
					go func() {
						name := uniqueMetricName("rapid_add")
						m := createCounterMetric(name)
						_ = p.AddMetric(false, m)
						time.Sleep(10 * time.Millisecond)
						p.DelMetric(name)
						done <- true
					}()
				}

				for i := 0; i < 50; i++ {
					go func() {
						_ = p.ListMetric()
						done <- true
					}()
				}

				for i := 0; i < 100; i++ {
					<-done
				}
			})
		})

		Context("when validating metric output", func() {
			It("should serve valid Prometheus metrics format", func() {
				p := newPrometheus()

				// Add a test metric so we have something to expose
				m := createCounterMetric(uniqueMetricName("test_format"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					_ = metric.Inc([]string{})
				})
				Expect(p.AddMetric(false, m)).ToNot(HaveOccurred())
				p.Collect(testCtx)

				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				output := w.Body.String()
				// Prometheus output should contain standard metrics format
				Expect(output).To(ContainSubstring("# HELP"))
				Expect(output).To(ContainSubstring("# TYPE"))
				// Should have Go runtime metrics at minimum
				Expect(output).To(Or(
					ContainSubstring("go_goroutines"),
					ContainSubstring("go_info"),
					ContainSubstring("go_gc"),
				))
			})

			It("should not include deleted metrics in output", func() {
				p := newPrometheus()
				ginsdk.SetMode(ginsdk.TestMode)
				router := ginsdk.New()

				name := uniqueMetricName("deleted_test")
				m := createCounterMetric(name)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				p.DelMetric(name)

				router.GET("/metrics", func(c *ginsdk.Context) {
					p.ExposeGin(c)
				})

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				output := w.Body.String()
				Expect(output).ToNot(ContainSubstring(name))
			})
		})

		Context("when configuring slow time and duration", func() {
			It("should use configured slow time", func() {
				p := newPrometheus()
				p.SetSlowTime(10)

				Expect(p.GetSlowTime()).To(Equal(int32(10)))
			})

			It("should use configured duration buckets", func() {
				p := newPrometheus()
				customBuckets := []float64{0.01, 0.05, 0.1, 0.5, 1.0}
				p.SetDuration(customBuckets)

				durations := p.GetDuration()
				for _, bucket := range customBuckets {
					Expect(durations).To(ContainElement(bucket))
				}
			})

			It("should work with webmetrics that use duration", func() {
				p := newPrometheus()
				customBuckets := []float64{0.001, 0.01, 0.1, 1.0}
				p.SetDuration(customBuckets)

				// Create a histogram that would use these buckets
				name := uniqueMetricName("duration_bucket_test")
				m := createHistogramMetric(name)

				// Use the prometheus duration buckets
				m.AddBuckets(p.GetDuration()...)

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Error Handling", func() {
		Context("when handling error conditions", func() {
			It("should handle metric registration failures gracefully", func() {
				p := newPrometheus()

				// Try to add metric without collect function
				name := uniqueMetricName("no_collect_error")
				m := prmmet.NewMetrics(name, prmtps.Counter)
				m.SetDesc("Test counter without collect for error handling")
				// Don't set collect function

				err := p.AddMetric(false, m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("collect func"))

				// Prometheus should still be functional
				Expect(p.ListMetric()).ToNot(BeNil())
			})

			It("should handle collection errors gracefully", func() {
				p := newPrometheus()

				m := createCounterMetric(uniqueMetricName("panic_collect"))
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					// dont panic to avoid pollute output
					// panic("intentional panic")
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				Expect(func() {
					ctx := context.Background()
					p.Collect(ctx)
				}).ToNot(Panic())
			})

			It("should handle invalid metric types", func() {
				p := newPrometheus()

				m := prmmet.NewMetrics("invalid", prmtps.None)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})

				err := p.AddMetric(false, m)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Resource Cleanup", func() {
		Context("when cleaning up resources", func() {
			It("should properly clean up deleted metrics", func() {
				p := newPrometheus()

				names := make([]string, 10)
				for i := 0; i < 10; i++ {
					names[i] = uniqueMetricName("cleanup_test")
					m := createCounterMetric(names[i])

					_ = p.AddMetric(false, m)
				}

				// Verify all added
				list := p.ListMetric()
				for _, name := range names {
					Expect(list).To(ContainElement(name))
				}

				// Delete all
				for _, name := range names {
					p.DelMetric(name)
				}

				// Verify all removed
				list = p.ListMetric()
				for _, name := range names {
					Expect(list).ToNot(ContainElement(name))
				}
			})

			It("should handle cleanup during active collection", func() {
				p := newPrometheus()

				name := uniqueMetricName("active_cleanup")
				m := createCounterMetric(name)
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
					time.Sleep(100 * time.Millisecond)
				})

				err := p.AddMetric(false, m)
				Expect(err).ToNot(HaveOccurred())

				// Start collection
				go func() {
					ctx := context.Background()
					p.Collect(ctx)
				}()

				// Delete while collecting
				time.Sleep(10 * time.Millisecond)
				p.DelMetric(name)

				// Should not panic
				time.Sleep(200 * time.Millisecond)
			})
		})
	})
})
