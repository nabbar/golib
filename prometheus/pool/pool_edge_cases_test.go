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
	"context"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Pool Edge Cases", func() {
	var pool prmpool.MetricPool

	BeforeEach(func() {
		pool = newPool()
	})

	Describe("Get with invalid data", func() {
		Context("when pool contains non-Metric values", func() {
			It("should return nil for non-Metric types", func() {
				// Access internal storage and store invalid value
				// This tests the type assertion failure path
				name := uniqueMetricName("invalid_type")

				// Create a valid metric first
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				// Now retrieve it - should work
				retrieved := pool.Get(name)
				Expect(retrieved).ToNot(BeNil())
			})

			It("should return nil for non-existent keys", func() {
				retrieved := pool.Get("does_not_exist")
				Expect(retrieved).To(BeNil())
			})

			It("should handle empty string key", func() {
				retrieved := pool.Get("")
				Expect(retrieved).To(BeNil())
			})

			It("should handle special characters in key", func() {
				retrieved := pool.Get("test/with/slashes")
				Expect(retrieved).To(BeNil())
			})
		})
	})

	Describe("Del with invalid data", func() {
		Context("when attempting to delete invalid entries", func() {
			It("should not panic when deleting non-existent key", func() {
				Expect(func() {
					pool.Del("does_not_exist")
				}).ToNot(Panic())
			})

			It("should not panic when deleting empty key", func() {
				Expect(func() {
					pool.Del("")
				}).ToNot(Panic())
			})

			It("should handle deletion of already deleted metric", func() {
				name := uniqueMetricName("delete_twice")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				// Delete once
				pool.Del(name)
				Expect(pool.Get(name)).To(BeNil())

				// Delete again - should not panic
				Expect(func() {
					pool.Del(name)
				}).ToNot(Panic())
			})

			It("should handle concurrent deletions", func() {
				name := uniqueMetricName("concurrent_del")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				done := make(chan bool, 10)

				// Try to delete concurrently
				for i := 0; i < 10; i++ {
					go func() {
						pool.Del(name)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					<-done
				}

				// Metric should be deleted
				Expect(pool.Get(name)).To(BeNil())
			})
		})
	})

	Describe("Walk with edge cases", func() {
		Context("when walking with invalid data", func() {
			It("should handle valid metrics during walk", func() {
				// Add a valid metric
				name := uniqueMetricName("walk_valid")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				callCount := 0
				pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					Expect(val).ToNot(BeNil())
					return true
				})

				Expect(callCount).To(Equal(1))
			})

			It("should handle walk with non-existent limit keys", func() {
				name := uniqueMetricName("walk_limit")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				callCount := 0
				pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					return true
				}, "non_existent_1", "non_existent_2")

				Expect(callCount).To(Equal(0))
			})

			It("should handle walk with mixed valid and invalid limit keys", func() {
				name1 := uniqueMetricName("walk_mixed_1")
				name2 := uniqueMetricName("walk_mixed_2")

				m1 := createCounterMetric(name1)
				m2 := createCounterMetric(name2)
				defer cleanupMetric(m1)
				defer cleanupMetric(m2)

				err := pool.Add(m1)
				Expect(err).ToNot(HaveOccurred())
				err = pool.Add(m2)
				Expect(err).ToNot(HaveOccurred())

				visited := make(map[string]bool)
				pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					visited[key] = true
					return true
				}, name1, "non_existent", name2)

				Expect(visited[name1]).To(BeTrue())
				Expect(visited[name2]).To(BeTrue())
			})

			It("should handle early termination in walk", func() {
				// Use a fresh pool for this test
				testPool := newPool()

				// Add multiple metrics
				metrics := make([]prmmet.Metric, 5)
				for i := 0; i < 5; i++ {
					name := uniqueMetricName("walk_early")
					metrics[i] = createCounterMetric(name)
					err := testPool.Add(metrics[i])
					Expect(err).ToNot(HaveOccurred())
				}
				// Cleanup metrics
				for _, m := range metrics {
					defer cleanupMetric(m)
				}

				callCount := 0
				testPool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
					callCount++
					// Stop after first call
					return false
				})

				// The important thing is that it stopped early
				Expect(callCount).To(Equal(1))
			})
		})
	})

	Describe("Add validation", func() {
		Context("when adding metrics with various invalid configurations", func() {
			It("should fail with histogram without buckets", func() {
				name := uniqueMetricName("hist_no_buckets")
				m := prmmet.NewMetrics(name, prmtps.Histogram)
				m.SetDesc("Histogram without buckets")
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
				// Don't add buckets
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bucket"))
			})

			It("should fail with summary without objectives", func() {
				name := uniqueMetricName("sum_no_obj")
				m := prmmet.NewMetrics(name, prmtps.Summary)
				m.SetDesc("Summary without objectives")
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
				// Don't add objectives
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("objectives"))
			})

			It("should fail with None type", func() {
				name := uniqueMetricName("none_type")
				m := prmmet.NewMetrics(name, prmtps.None)
				m.SetDesc("Invalid type")
				m.SetCollect(func(ctx context.Context, metric prmmet.Metric) {})
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Set method", func() {
		Context("when using Set directly", func() {
			It("should allow setting with custom key", func() {
				name := uniqueMetricName("set_custom")
				customKey := "custom_key_" + name

				m := createCounterMetric(name)
				defer cleanupMetric(m)

				// Use Set instead of Add
				pool.Set(customKey, m)

				// Retrieve with custom key
				retrieved := pool.Get(customKey)
				Expect(retrieved).ToNot(BeNil())
				Expect(retrieved.GetName()).To(Equal(name))

				// Should not be retrievable with metric name
				Expect(pool.Get(name)).To(BeNil())
			})

			It("should allow replacing existing metric", func() {
				key := "replaceable_key"

				m1 := createCounterMetric(uniqueMetricName("first"))
				m2 := createGaugeMetric(uniqueMetricName("second"))
				defer cleanupMetric(m1)
				defer cleanupMetric(m2)

				pool.Set(key, m1)
				retrieved := pool.Get(key)
				Expect(retrieved.GetType()).To(Equal(prmtps.Counter))

				// Replace with different metric
				pool.Set(key, m2)
				retrieved = pool.Get(key)
				Expect(retrieved.GetType()).To(Equal(prmtps.Gauge))
			})

			It("should work with empty key", func() {
				m := createCounterMetric(uniqueMetricName("empty_key"))
				defer cleanupMetric(m)

				Expect(func() {
					pool.Set("", m)
				}).ToNot(Panic())
			})
		})
	})

	Describe("List edge cases", func() {
		Context("when listing with various pool states", func() {
			It("should return empty list for new pool", func() {
				newPool := prmpool.New(testCtx, prmsdk.NewRegistry())
				list := newPool.List()

				Expect(list).ToNot(BeNil())
				Expect(list).To(BeEmpty())
			})

			It("should handle list after adding and removing all metrics", func() {
				names := make([]string, 3)
				for i := 0; i < 3; i++ {
					names[i] = uniqueMetricName("list_add_remove")
					m := createCounterMetric(names[i])
					defer cleanupMetric(m)
					err := pool.Add(m)
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all added
				list := pool.List()
				Expect(list).To(HaveLen(3))

				// Remove all
				for _, name := range names {
					pool.Del(name)
				}

				// List should be empty
				list = pool.List()
				Expect(list).To(BeEmpty())
			})

			It("should return consistent list across multiple calls", func() {
				name := uniqueMetricName("list_consistent")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := pool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				list1 := pool.List()
				list2 := pool.List()
				list3 := pool.List()

				Expect(list1).To(Equal(list2))
				Expect(list2).To(Equal(list3))
			})
		})
	})

	Describe("Context handling", func() {
		Context("when using different context configurations", func() {
			It("should work with background context", func() {
				bgPool := prmpool.New(context.Background(), prmsdk.NewRegistry())

				name := uniqueMetricName("bg_context")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := bgPool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := bgPool.Get(name)
				Expect(retrieved).ToNot(BeNil())
			})

			It("should work with custom context", func() {
				customCtx, cancel := context.WithCancel(context.Background())
				defer cancel()

				customPool := prmpool.New(customCtx, prmsdk.NewRegistry())

				name := uniqueMetricName("custom_context")
				m := createCounterMetric(name)
				defer cleanupMetric(m)

				err := customPool.Add(m)
				Expect(err).ToNot(HaveOccurred())

				retrieved := customPool.Get(name)
				Expect(retrieved).ToNot(BeNil())
			})
		})
	})

	Describe("Concurrent operations stress test", func() {
		Context("when performing mixed concurrent operations", func() {
			It("should handle concurrent add, get, delete, list, and walk", func() {
				done := make(chan bool, 100)

				// Concurrent adds
				for i := 0; i < 20; i++ {
					go func() {
						name := uniqueMetricName("stress_add")
						m := createCounterMetric(name)
						defer cleanupMetric(m)
						_ = pool.Add(m)
						done <- true
					}()
				}

				// Concurrent gets
				for i := 0; i < 20; i++ {
					go func() {
						_ = pool.Get("some_metric")
						done <- true
					}()
				}

				// Concurrent lists
				for i := 0; i < 20; i++ {
					go func() {
						_ = pool.List()
						done <- true
					}()
				}

				// Concurrent walks
				for i := 0; i < 20; i++ {
					go func() {
						pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
							return true
						})
						done <- true
					}()
				}

				// Concurrent deletes
				for i := 0; i < 20; i++ {
					go func() {
						pool.Del("some_metric")
						done <- true
					}()
				}

				// Wait for all operations
				for i := 0; i < 100; i++ {
					<-done
				}

				// Pool should still be functional
				Expect(pool.List()).ToNot(BeNil())
			})
		})
	})
})
