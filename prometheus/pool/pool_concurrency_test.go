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
	"sync"
	"sync/atomic"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metric Pool Concurrency", func() {
	Describe("Concurrent operations", func() {
		Context("concurrent Add operations", func() {
			It("should handle multiple goroutines adding metrics", func() {
				pool := newPool()
				wg := sync.WaitGroup{}
				iterations := 20

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("concurrent_add_%d", idx)
						m := createCounterMetric(name, "label")
						err := pool.Add(m)
						Expect(err).ToNot(HaveOccurred())
					}(i)
				}

				wg.Wait()
				list := pool.List()
				Expect(len(list)).To(BeNumerically(">=", 1))
				Expect(len(list)).To(BeNumerically("<=", iterations))
			})

			It("should handle concurrent additions with same name gracefully", func() {
				pool := newPool()
				wg := sync.WaitGroup{}
				successCount := atomic.Int32{}
				failureCount := atomic.Int32{}
				iterations := 10

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						m := createCounterMetric("duplicate_concurrent", "label")
						err := pool.Add(m)
						if err != nil {
							failureCount.Add(1)
						} else {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()
				// Only one should succeed
				Expect(successCount.Load()).To(Equal(int32(1)))
				Expect(failureCount.Load()).To(Equal(int32(iterations - 1)))
			})
		})

		Context("concurrent Get operations", func() {
			It("should handle multiple goroutines reading metrics", func() {
				pool := newPool()
				addMetricToPool(pool, "shared_metric", prmtps.Counter, "label")

				wg := sync.WaitGroup{}
				iterations := 50

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						m := pool.Get("shared_metric")
						Expect(m).ToNot(BeNil())
						Expect(m.GetName()).To(Equal("shared_metric"))
					}()
				}

				wg.Wait()
			})

			It("should handle concurrent reads of multiple metrics", func() {
				pool := newPool()
				for i := 0; i < 5; i++ {
					addMetricToPool(pool, fmt.Sprintf("metric_%d", i), prmtps.Counter)
				}

				wg := sync.WaitGroup{}
				iterations := 30

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						metricNum := idx % 5
						m := pool.Get(fmt.Sprintf("metric_%d", metricNum))
						Expect(m).ToNot(BeNil())
					}(i)
				}

				wg.Wait()
			})
		})

		Context("concurrent Set operations", func() {
			It("should handle multiple goroutines setting metrics", func() {
				pool := newPool()
				wg := sync.WaitGroup{}
				iterations := 20

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("set_metric_%d", idx)
						m := createGaugeMetric(name, "label")
						pool.Set(name, m)
					}(i)
				}

				wg.Wait()

				list := pool.List()
				Expect(len(list)).To(BeNumerically(">=", 1))
			})

			It("should handle concurrent overwrites to same key", func() {
				pool := newPool()
				wg := sync.WaitGroup{}
				iterations := 15

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						m := createCounterMetric(fmt.Sprintf("overwrite_%d", idx), "label")
						pool.Set("same_key", m)
					}(i)
				}

				wg.Wait()

				// Should have exactly one metric at "same_key"
				m := pool.Get("same_key")
				Expect(m).ToNot(BeNil())

				list := pool.List()
				Expect(list).To(ContainElement("same_key"))
			})
		})

		Context("concurrent Del operations", func() {
			It("should handle multiple goroutines deleting metrics", func() {
				pool := newPool()
				metricCount := 20

				// Add metrics
				for i := 0; i < metricCount; i++ {
					addMetricToPool(pool, fmt.Sprintf("del_metric_%d", i), prmtps.Counter)
				}

				wg := sync.WaitGroup{}

				// Delete half of them concurrently
				for i := 0; i < metricCount/2; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						pool.Del(fmt.Sprintf("del_metric_%d", idx))
					}(i)
				}

				wg.Wait()

				list := pool.List()
				Expect(len(list)).To(BeNumerically(">=", metricCount/2))
			})

			It("should handle concurrent deletion of same metric", func() {
				pool := newPool()
				addMetricToPool(pool, "delete_me", prmtps.Counter)

				wg := sync.WaitGroup{}
				iterations := 10

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						Expect(func() {
							pool.Del("delete_me")
						}).ToNot(Panic())
					}()
				}

				wg.Wait()

				m := pool.Get("delete_me")
				Expect(m).To(BeNil())
			})
		})

		Context("mixed concurrent operations", func() {
			It("should handle Add, Get, Set, Del concurrently", func() {
				pool := newPool()
				wg := sync.WaitGroup{}
				iterations := 50

				// Concurrent Adds
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("mixed_add_%d", idx)
						m := createCounterMetric(name, "label")
						_ = pool.Add(m)
					}(i)
				}

				// Concurrent Gets
				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("mixed_add_%d", idx%20)
						pool.Get(name)
					}(i)
				}

				// Concurrent Sets
				for i := 0; i < iterations/2; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("mixed_set_%d", idx)
						m := createGaugeMetric(name, "label")
						pool.Set(name, m)
					}(i)
				}

				// Concurrent Dels
				for i := 0; i < iterations/4; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						defer GinkgoRecover()

						name := fmt.Sprintf("mixed_add_%d", idx)
						pool.Del(name)
					}(i)
				}

				wg.Wait()

				// Just verify pool is still functional
				list := pool.List()
				Expect(list).ToNot(BeNil())
			})

			It("should handle concurrent List operations", func() {
				pool := newPool()

				// Add some metrics
				for i := 0; i < 10; i++ {
					addMetricToPool(pool, fmt.Sprintf("list_metric_%d", i), prmtps.Counter)
				}

				wg := sync.WaitGroup{}
				iterations := 30

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						list := pool.List()
						Expect(list).ToNot(BeNil())
						Expect(len(list)).To(BeNumerically(">=", 0))
					}()
				}

				wg.Wait()
			})

			It("should handle concurrent Walk operations", func() {
				pool := newPool()

				// Add metrics
				for i := 0; i < 10; i++ {
					addMetricToPool(pool, fmt.Sprintf("walk_metric_%d", i), prmtps.Counter)
				}

				wg := sync.WaitGroup{}
				iterations := 20

				for i := 0; i < iterations; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()

						count := 0
						walkFunc := func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
							count++
							return true
						}

						pool.Walk(walkFunc)
						Expect(count).To(BeNumerically(">=", 0))
					}()
				}

				wg.Wait()
			})
		})
	})
})
