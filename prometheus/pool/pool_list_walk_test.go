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
	"fmt"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metric Pool List and Walk Operations", func() {
	Describe("List", func() {
		var pool prmpool.MetricPool
		var metrics []prmmet.Metric

		BeforeEach(func() {
			pool = newPool()
			metrics = make([]prmmet.Metric, 0)
		})

		AfterEach(func() {
			for _, m := range metrics {
				cleanupMetric(m)
			}
		})

		Context("with empty pool", func() {
			It("should return empty slice", func() {
				list := pool.List()
				Expect(list).To(BeEmpty())
				Expect(list).ToNot(BeNil())
			})
		})

		Context("with single metric", func() {
			It("should return slice with one entry", func() {
				name := uniqueMetricName("single_metric")
				m := addMetricToPool(pool, name, prmtps.Counter)
				metrics = append(metrics, m)

				list := pool.List()
				Expect(list).To(HaveLen(1))
				Expect(list).To(ContainElement(name))
			})
		})

		Context("with multiple metrics", func() {
			It("should return all metric names", func() {
				names := make([]string, 4)
				for i := 0; i < 4; i++ {
					names[i] = uniqueMetricName(fmt.Sprintf("metric_%d", i))
					m := addMetricToPool(pool, names[i], prmtps.Counter)
					metrics = append(metrics, m)
				}

				list := pool.List()
				Expect(list).To(HaveLen(4))
				for _, name := range names {
					Expect(list).To(ContainElement(name))
				}
			})

			It("should return correct count after additions", func() {
				n1 := uniqueMetricName("m1")
				m := addMetricToPool(pool, n1, prmtps.Counter)
				metrics = append(metrics, m)
				Expect(pool.List()).To(HaveLen(1))

				n2 := uniqueMetricName("m2")
				m = addMetricToPool(pool, n2, prmtps.Gauge)
				metrics = append(metrics, m)
				Expect(pool.List()).To(HaveLen(2))

				n3 := uniqueMetricName("m3")
				m = addMetricToPool(pool, n3, prmtps.Histogram)
				metrics = append(metrics, m)
				Expect(pool.List()).To(HaveLen(3))
			})

			It("should return correct count after deletions", func() {
				n1 := uniqueMetricName("m1")
				m1 := addMetricToPool(pool, n1, prmtps.Counter)
				metrics = append(metrics, m1)

				n2 := uniqueMetricName("m2")
				m2 := addMetricToPool(pool, n2, prmtps.Gauge)
				metrics = append(metrics, m2)

				n3 := uniqueMetricName("m3")
				m3 := addMetricToPool(pool, n3, prmtps.Histogram)
				metrics = append(metrics, m3)

				Expect(pool.List()).To(HaveLen(3))

				pool.Del(n2)
				list := pool.List()
				Expect(list).To(HaveLen(2))
				Expect(list).To(ContainElement(n1))
				Expect(list).To(ContainElement(n3))
				Expect(list).ToNot(ContainElement(n2))
			})
		})

		Context("with different metric types", func() {
			It("should list all types correctly", func() {
				names := make([]string, 4)
				names[0] = uniqueMetricName("counter1")
				m := addMetricToPool(pool, names[0], prmtps.Counter)
				metrics = append(metrics, m)

				names[1] = uniqueMetricName("gauge1")
				m = addMetricToPool(pool, names[1], prmtps.Gauge)
				metrics = append(metrics, m)

				names[2] = uniqueMetricName("histogram1")
				m = addMetricToPool(pool, names[2], prmtps.Histogram)
				metrics = append(metrics, m)

				names[3] = uniqueMetricName("summary1")
				m = addMetricToPool(pool, names[3], prmtps.Summary)
				metrics = append(metrics, m)

				list := pool.List()
				Expect(list).To(HaveLen(4))
				Expect(list).To(ConsistOf(names[0], names[1], names[2], names[3]))
			})
		})
	})

	Describe("Walk", func() {
		var pool prmpool.MetricPool
		var metrics []prmmet.Metric

		BeforeEach(func() {
			pool = newPool()
			metrics = make([]prmmet.Metric, 0)
		})

		AfterEach(func() {
			for _, m := range metrics {
				cleanupMetric(m)
			}
		})

		Context("with empty pool", func() {
			It("should complete without calling function", func() {
				called := false
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					called = true
					return true
				}

				pool.Walk(walkFunc)
				Expect(called).To(BeFalse())
			})
		})

		Context("with single metric", func() {
			It("should call function once", func() {
				name := uniqueMetricName("single")
				m := addMetricToPool(pool, name, prmtps.Counter)
				metrics = append(metrics, m)

				callCount := 0
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					Expect(key).To(Equal(name))
					Expect(val).ToNot(BeNil())
					Expect(val.GetName()).To(Equal(name))
					return true
				}

				pool.Walk(walkFunc)
				Expect(callCount).To(Equal(1))
			})
		})

		Context("with multiple metrics", func() {
			It("should iterate over all metrics", func() {
				names := make([]string, 4)
				for i := 0; i < 4; i++ {
					names[i] = uniqueMetricName(fmt.Sprintf("m_%d", i))
					m := addMetricToPool(pool, names[i], prmtps.Counter)
					metrics = append(metrics, m)
				}

				visited := make(map[string]bool)
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					visited[key] = true
					Expect(val).ToNot(BeNil())
					Expect(val.GetName()).To(Equal(key))
					return true
				}

				pool.Walk(walkFunc)
				Expect(visited).To(HaveLen(4))
				for _, name := range names {
					Expect(visited[name]).To(BeTrue())
				}
			})

			It("should provide access to pool reference", func() {
				name := uniqueMetricName("test")
				m := addMetricToPool(pool, name, prmtps.Counter)
				metrics = append(metrics, m)

				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					Expect(p).ToNot(BeNil())
					Expect(p).To(Equal(pool))
					// Should be able to use pool operations
					Expect(p.Get(key)).ToNot(BeNil())
					return true
				}

				pool.Walk(walkFunc)
			})

			It("should handle different metric types", func() {
				names := make([]string, 4)
				names[0] = uniqueMetricName("counter")
				m := addMetricToPool(pool, names[0], prmtps.Counter)
				metrics = append(metrics, m)

				names[1] = uniqueMetricName("gauge")
				m = addMetricToPool(pool, names[1], prmtps.Gauge)
				metrics = append(metrics, m)

				names[2] = uniqueMetricName("histogram")
				m = addMetricToPool(pool, names[2], prmtps.Histogram)
				metrics = append(metrics, m)

				names[3] = uniqueMetricName("summary")
				m = addMetricToPool(pool, names[3], prmtps.Summary)
				metrics = append(metrics, m)

				types := make(map[prmtps.MetricType]int)
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					types[val.GetType()]++
					return true
				}

				pool.Walk(walkFunc)
				Expect(types).To(HaveLen(4))
				Expect(types[prmtps.Counter]).To(Equal(1))
				Expect(types[prmtps.Gauge]).To(Equal(1))
				Expect(types[prmtps.Histogram]).To(Equal(1))
				Expect(types[prmtps.Summary]).To(Equal(1))
			})
		})

		Context("with early termination", func() {
			It("should stop when function returns false", func() {
				for i := 1; i <= 5; i++ {
					name := uniqueMetricName(fmt.Sprintf("early_term_%d", i))
					m := addMetricToPool(pool, name, prmtps.Counter)
					metrics = append(metrics, m)
				}

				callCount := 0
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					return callCount < 3 // Stop after 3 calls
				}

				pool.Walk(walkFunc)
				// Verify early termination happened by checking call count
				Expect(callCount).To(Equal(3))
			})

			It("should stop immediately when function returns false on first call", func() {
				n1 := uniqueMetricName("m1")
				m1 := addMetricToPool(pool, n1, prmtps.Counter)
				metrics = append(metrics, m1)

				n2 := uniqueMetricName("m2")
				m2 := addMetricToPool(pool, n2, prmtps.Gauge)
				metrics = append(metrics, m2)

				callCount := 0
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					return false // Always stop immediately
				}

				pool.Walk(walkFunc)
				// Should only be called once
				Expect(callCount).To(Equal(1))
			})
		})

		Context("with limit parameter", func() {
			var metricNames map[string]string // map of key to actual metric name

			BeforeEach(func() {
				// Add multiple metrics with unique names
				metricNames = make(map[string]string)

				metricNames["a"] = uniqueMetricName("metric_a")
				m := addMetricToPool(pool, metricNames["a"], prmtps.Counter)
				metrics = append(metrics, m)

				metricNames["b"] = uniqueMetricName("metric_b")
				m = addMetricToPool(pool, metricNames["b"], prmtps.Gauge)
				metrics = append(metrics, m)

				metricNames["c"] = uniqueMetricName("metric_c")
				m = addMetricToPool(pool, metricNames["c"], prmtps.Histogram)
				metrics = append(metrics, m)

				metricNames["x"] = uniqueMetricName("other_x")
				m = addMetricToPool(pool, metricNames["x"], prmtps.Summary)
				metrics = append(metrics, m)

				metricNames["y"] = uniqueMetricName("other_y")
				m = addMetricToPool(pool, metricNames["y"], prmtps.Counter)
				metrics = append(metrics, m)
			})

			It("should walk only specified metrics when limit is provided", func() {
				visited := make([]string, 0)
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					visited = append(visited, key)
					return true
				}

				pool.Walk(walkFunc, metricNames["a"], metricNames["c"])
				Expect(visited).To(ConsistOf(metricNames["a"], metricNames["c"]))
			})

			It("should walk single metric when single limit provided", func() {
				callCount := 0
				var visitedKey string
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					visitedKey = key
					return true
				}

				pool.Walk(walkFunc, metricNames["b"])
				Expect(callCount).To(Equal(1))
				Expect(visitedKey).To(Equal(metricNames["b"]))
			})

			It("should handle non-existing keys in limit", func() {
				visited := make([]string, 0)
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					visited = append(visited, key)
					return true
				}

				pool.Walk(walkFunc, metricNames["a"], "does_not_exist", metricNames["c"])
				// Should only visit existing metrics
				Expect(visited).To(ConsistOf(metricNames["a"], metricNames["c"]))
			})

			It("should walk all metrics when no limit provided", func() {
				visited := make([]string, 0)
				walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					visited = append(visited, key)
					return true
				}

				pool.Walk(walkFunc)
				Expect(visited).To(HaveLen(5))
			})
		})
	})
})
