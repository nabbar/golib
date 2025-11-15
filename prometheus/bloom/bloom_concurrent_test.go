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
	"sync"

	"github.com/nabbar/golib/prometheus/bloom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrent Operations", func() {
	Describe("BloomFilter Concurrency", func() {
		var bf bloom.BloomFilter

		BeforeEach(func() {
			bf = bloom.NewBloomFilter()
		})

		It("should handle concurrent additions", func() {
			var wg sync.WaitGroup
			numGoroutines := 100
			valuesPerGoroutine := 100

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < valuesPerGoroutine; j++ {
						value := "goroutine-" + string(rune('0'+goroutineID%10)) + "-value-" + string(rune('0'+j%10))
						bf.Add(value)
					}
				}(i)
			}

			wg.Wait()

			// Verify some values were added successfully
			bf.Add("test-after-concurrent")
			Expect(bf.Contains("test-after-concurrent")).To(BeTrue())
		})

		It("should handle concurrent reads", func() {
			// Pre-populate with some values
			for i := 0; i < 100; i++ {
				bf.Add("value-" + string(rune('0'+i%10)))
			}

			var wg sync.WaitGroup
			numGoroutines := 50

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 100; j++ {
						value := "value-" + string(rune('0'+j%10))
						_ = bf.Contains(value)
					}
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent adds and reads", func() {
			var wg sync.WaitGroup
			numWriters := 25
			numReaders := 25

			// Writers
			wg.Add(numWriters)
			for i := 0; i < numWriters; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 50; j++ {
						value := "writer-" + string(rune('0'+id%10)) + "-" + string(rune('0'+j%10))
						bf.Add(value)
					}
				}(i)
			}

			// Readers
			wg.Add(numReaders)
			for i := 0; i < numReaders; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 50; j++ {
						value := "reader-check-" + string(rune('0'+id%10)) + "-" + string(rune('0'+j%10))
						_ = bf.Contains(value)
					}
				}(i)
			}

			wg.Wait()

			// Verify filter still works after concurrent operations
			bf.Add("final-test")
			Expect(bf.Contains("final-test")).To(BeTrue())
		})

		It("should maintain correctness under concurrent stress", func() {
			knownValues := []string{
				"concurrent-test-1",
				"concurrent-test-2",
				"concurrent-test-3",
				"concurrent-test-4",
				"concurrent-test-5",
			}

			// Add known values
			for _, v := range knownValues {
				bf.Add(v)
			}

			var wg sync.WaitGroup
			numGoroutines := 100

			// Concurrent reads of known values
			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					for _, v := range knownValues {
						Expect(bf.Contains(v)).To(BeTrue(), "Known value should always be found: "+v)
					}
				}()
			}

			wg.Wait()
		})
	})

	Describe("Collection Concurrency", func() {
		var col bloom.Collection

		BeforeEach(func() {
			col = bloom.New()
		})

		It("should handle concurrent additions to same metric", func() {
			var wg sync.WaitGroup
			numGoroutines := 50
			metricName := "concurrent-metric"

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 50; j++ {
						value := "g" + string(rune('0'+goroutineID%10)) + "-v" + string(rune('0'+j%10))
						col.Add(metricName, value)
					}
				}(i)
			}

			wg.Wait()

			// Verify collection still works
			col.Add(metricName, "test-after-concurrent")
			Expect(col.Contains(metricName, "test-after-concurrent")).To(BeTrue())
		})

		It("should handle concurrent additions to different metrics", func() {
			var wg sync.WaitGroup
			numGoroutines := 50

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer GinkgoRecover()
					defer wg.Done()

					metricName := "metric-" + string(rune('0'+goroutineID%10))
					for j := 0; j < 50; j++ {
						value := "value-" + string(rune('0'+j%10))
						col.Add(metricName, value)
					}
				}(i)
			}

			wg.Wait()

			// Verify different metrics are isolated
			col.Add("metric-1", "test-value")
			Expect(col.Contains("metric-1", "test-value")).To(BeTrue())
			Expect(col.Contains("metric-2", "test-value")).To(BeFalse())
		})

		It("should handle concurrent reads from same metric", func() {
			metricName := "read-test-metric"

			// Pre-populate
			for i := 0; i < 100; i++ {
				col.Add(metricName, "value-"+string(rune('0'+i%10)))
			}

			var wg sync.WaitGroup
			numGoroutines := 50

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 100; j++ {
						value := "value-" + string(rune('0'+j%10))
						_ = col.Contains(metricName, value)
					}
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent reads and writes", func() {
			var wg sync.WaitGroup
			metricName := "mixed-ops-metric"
			numWriters := 25
			numReaders := 25

			// Writers
			wg.Add(numWriters)
			for i := 0; i < numWriters; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 50; j++ {
						value := "w" + string(rune('0'+id%10)) + "-" + string(rune('0'+j%10))
						col.Add(metricName, value)
					}
				}(i)
			}

			// Readers
			wg.Add(numReaders)
			for i := 0; i < numReaders; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					for j := 0; j < 50; j++ {
						value := "r" + string(rune('0'+id%10)) + "-" + string(rune('0'+j%10))
						_ = col.Contains(metricName, value)
					}
				}(i)
			}

			wg.Wait()

			// Verify collection is still functional
			col.Add(metricName, "final-check")
			Expect(col.Contains(metricName, "final-check")).To(BeTrue())
		})

		It("should handle concurrent operations across multiple metrics", func() {
			var wg sync.WaitGroup
			numGoroutines := 100

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer GinkgoRecover()
					defer wg.Done()

					// Each goroutine works with its own set of metrics
					for m := 0; m < 5; m++ {
						metricName := "g" + string(rune('0'+goroutineID%10)) + "-m" + string(rune('0'+m))
						for v := 0; v < 10; v++ {
							value := "v" + string(rune('0'+v))
							col.Add(metricName, value)

							// Immediately verify
							if !col.Contains(metricName, value) {
								Fail("Value should be found immediately after adding: " + metricName + " - " + value)
							}
						}
					}
				}(i)
			}

			wg.Wait()
		})

		It("should maintain data integrity under heavy concurrent load", func() {
			var wg sync.WaitGroup
			numGoroutines := 100
			metricsPerGoroutine := 10
			valuesPerMetric := 20

			// Track what we're adding
			type metricValue struct {
				metric string
				value  string
			}
			addedValues := make(chan metricValue, numGoroutines*metricsPerGoroutine*valuesPerMetric)

			// Add concurrently
			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer GinkgoRecover()
					defer wg.Done()

					for m := 0; m < metricsPerGoroutine; m++ {
						metric := "stress-m" + string(rune('0'+m%10))
						for v := 0; v < valuesPerMetric; v++ {
							value := "g" + string(rune('0'+goroutineID%10)) + "-v" + string(rune('0'+v%10))
							col.Add(metric, value)
							addedValues <- metricValue{metric, value}
						}
					}
				}(i)
			}

			wg.Wait()
			close(addedValues)

			// Verify a sample of added values
			count := 0
			for mv := range addedValues {
				if count%100 == 0 { // Sample 1% of values
					Expect(col.Contains(mv.metric, mv.value)).To(BeTrue(),
						"Value should be found: "+mv.metric+" - "+mv.value)
				}
				count++
			}
		})
	})
})
