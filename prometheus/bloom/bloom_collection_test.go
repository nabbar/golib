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

package bloom_test

import (
	"github.com/nabbar/golib/prometheus/bloom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collection", func() {
	var col bloom.Collection

	BeforeEach(func() {
		col = bloom.New()
	})

	Describe("New", func() {
		It("should create a new collection", func() {
			Expect(col).ToNot(BeNil())
		})
	})

	Describe("Add and Contains", func() {
		Context("with single metric", func() {
			It("should add and retrieve a value for a metric", func() {
				col.Add("metric1", "value1")
				Expect(col.Contains("metric1", "value1")).To(BeTrue())
			})

			It("should return false for non-existent value in existing metric", func() {
				col.Add("metric1", "value1")
				Expect(col.Contains("metric1", "value2")).To(BeFalse())
			})

			It("should return false for non-existent metric", func() {
				col.Add("metric1", "value1")
				Expect(col.Contains("metric2", "value1")).To(BeFalse())
			})

			It("should handle multiple values for same metric", func() {
				col.Add("metric1", "value1")
				col.Add("metric1", "value2")
				col.Add("metric1", "value3")

				Expect(col.Contains("metric1", "value1")).To(BeTrue())
				Expect(col.Contains("metric1", "value2")).To(BeTrue())
				Expect(col.Contains("metric1", "value3")).To(BeTrue())
			})
		})

		Context("with multiple metrics", func() {
			It("should manage separate Bloom filters per metric", func() {
				col.Add("metric1", "value1")
				col.Add("metric2", "value2")
				col.Add("metric3", "value3")

				Expect(col.Contains("metric1", "value1")).To(BeTrue())
				Expect(col.Contains("metric2", "value2")).To(BeTrue())
				Expect(col.Contains("metric3", "value3")).To(BeTrue())

				// Verify isolation between metrics
				Expect(col.Contains("metric1", "value2")).To(BeFalse())
				Expect(col.Contains("metric2", "value1")).To(BeFalse())
				Expect(col.Contains("metric3", "value1")).To(BeFalse())
			})

			It("should handle many metrics", func() {
				// Add 100 different metrics with values
				for i := 0; i < 100; i++ {
					metricName := "metric-" + string(rune('0'+i%10))
					value := "value-" + string(rune('a'+i%26))
					col.Add(metricName, value)
				}

				// Verify some random samples
				col.Add("metric-5", "value-x")
				Expect(col.Contains("metric-5", "value-x")).To(BeTrue())

				col.Add("metric-9", "value-z")
				Expect(col.Contains("metric-9", "value-z")).To(BeTrue())
			})
		})

		Context("with duplicate additions", func() {
			It("should handle adding same metric-value pair multiple times", func() {
				col.Add("metric1", "value1")
				col.Add("metric1", "value1")
				col.Add("metric1", "value1")

				Expect(col.Contains("metric1", "value1")).To(BeTrue())
			})
		})

		Context("with metric name variations", func() {
			It("should treat metric names as case-sensitive", func() {
				col.Add("metric", "value")
				col.Add("Metric", "value")
				col.Add("METRIC", "value")

				Expect(col.Contains("metric", "value")).To(BeTrue())
				Expect(col.Contains("Metric", "value")).To(BeTrue())
				Expect(col.Contains("METRIC", "value")).To(BeTrue())
				Expect(col.Contains("MeTrIc", "value")).To(BeFalse())
			})

			It("should handle special characters in metric names", func() {
				col.Add("metric.name", "value1")
				col.Add("metric:name", "value2")
				col.Add("metric_name", "value3")
				col.Add("metric-name", "value4")

				Expect(col.Contains("metric.name", "value1")).To(BeTrue())
				Expect(col.Contains("metric:name", "value2")).To(BeTrue())
				Expect(col.Contains("metric_name", "value3")).To(BeTrue())
				Expect(col.Contains("metric-name", "value4")).To(BeTrue())
			})
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty metric name", func() {
			col.Add("", "value")
			Expect(col.Contains("", "value")).To(BeTrue())
		})

		It("should handle empty value", func() {
			col.Add("metric", "")
			Expect(col.Contains("metric", "")).To(BeFalse()) // Empty strings return false in BloomFilter.Contains
		})

		It("should handle both empty metric and value", func() {
			col.Add("", "")
			Expect(col.Contains("", "")).To(BeFalse())
		})

		It("should not panic on nil checks", func() {
			// The implementation checks for nil, ensure it doesn't panic
			Expect(func() {
				col.Contains("nonexistent", "value")
			}).ToNot(Panic())
		})
	})

	Describe("Real-world Scenarios", func() {
		Context("with Prometheus metric labels", func() {
			It("should handle typical metric label combinations", func() {
				// Simulate typical Prometheus metric scenarios
				col.Add("http_requests_total", "method=GET,status=200,path=/api/users")
				col.Add("http_requests_total", "method=POST,status=201,path=/api/users")
				col.Add("http_requests_total", "method=GET,status=404,path=/api/missing")

				Expect(col.Contains("http_requests_total", "method=GET,status=200,path=/api/users")).To(BeTrue())
				Expect(col.Contains("http_requests_total", "method=POST,status=201,path=/api/users")).To(BeTrue())
				Expect(col.Contains("http_requests_total", "method=GET,status=404,path=/api/missing")).To(BeTrue())
				Expect(col.Contains("http_requests_total", "method=DELETE,status=200,path=/api/users")).To(BeFalse())
			})
		})

		Context("with cardinality tracking", func() {
			It("should efficiently track high-cardinality data", func() {
				metricName := "high_cardinality_metric"

				// Add 1000 unique label combinations
				for i := 0; i < 1000; i++ {
					labels := "user_id=" + string(rune('0'+i%10)) + ",session=" + string(rune('a'+i%26))
					col.Add(metricName, labels)
				}

				// Verify some known combinations exist
				col.Add(metricName, "user_id=5,session=x")
				Expect(col.Contains(metricName, "user_id=5,session=x")).To(BeTrue())

				// Verify unknown combination doesn't exist
				Expect(col.Contains(metricName, "user_id=999,session=zzz")).To(BeFalse())
			})
		})
	})

	Describe("Collection Isolation", func() {
		It("should isolate metrics from each other", func() {
			// Add same value to different metrics
			col.Add("metric1", "shared-value")
			col.Add("metric2", "shared-value")
			col.Add("metric3", "shared-value")

			// All should find their own value
			Expect(col.Contains("metric1", "shared-value")).To(BeTrue())
			Expect(col.Contains("metric2", "shared-value")).To(BeTrue())
			Expect(col.Contains("metric3", "shared-value")).To(BeTrue())

			// But adding to one shouldn't affect queries to another with different value
			col.Add("metric1", "unique-to-metric1")
			Expect(col.Contains("metric1", "unique-to-metric1")).To(BeTrue())
			Expect(col.Contains("metric2", "unique-to-metric1")).To(BeFalse())
		})
	})

	Describe("Performance Characteristics", func() {
		It("should handle rapid sequential additions", func() {
			metricName := "perf-test-metric"

			// Rapidly add many values
			testValues := make([]string, 0, 10000)
			for i := 0; i < 10000; i++ {
				value := "value-" + string(rune('0'+i%10)) + string(rune('a'+i%26))
				testValues = append(testValues, value)
				col.Add(metricName, value)
			}

			// Spot check some values that we actually added
			Expect(col.Contains(metricName, testValues[0])).To(BeTrue())
			Expect(col.Contains(metricName, testValues[5])).To(BeTrue())
			Expect(col.Contains(metricName, testValues[100])).To(BeTrue())
			Expect(col.Contains(metricName, testValues[9999])).To(BeTrue())
		})
	})
})
