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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus Basic Operations", func() {
	Describe("New", func() {
		Context("when creating a new Prometheus instance", func() {
			It("should create successfully", func() {
				p := newPrometheus()
				Expect(p).ToNot(BeNil())
			})

			It("should have default slow time", func() {
				p := newPrometheus()
				Expect(p.GetSlowTime()).To(Equal(libprm.DefaultSlowTime))
			})

			It("should have default duration buckets", func() {
				p := newPrometheus()
				durations := p.GetDuration()
				Expect(durations).To(HaveLen(5))
				Expect(durations).To(Equal([]float64{0.1, 0.3, 1.2, 5, 10}))
			})

			It("should have empty metric list initially", func() {
				p := newPrometheus()
				Expect(p.ListMetric()).To(BeEmpty())
			})
		})
	})

	Describe("SlowTime Configuration", func() {
		var p libprm.Prometheus

		BeforeEach(func() {
			p = newPrometheus()
		})

		Context("when setting slow time", func() {
			It("should update slow time successfully", func() {
				p.SetSlowTime(10)
				Expect(p.GetSlowTime()).To(Equal(int32(10)))
			})

			It("should handle zero value", func() {
				p.SetSlowTime(0)
				Expect(p.GetSlowTime()).To(Equal(int32(0)))
			})

			It("should handle negative values", func() {
				p.SetSlowTime(-5)
				Expect(p.GetSlowTime()).To(Equal(int32(-5)))
			})

			It("should handle large values", func() {
				p.SetSlowTime(3600)
				Expect(p.GetSlowTime()).To(Equal(int32(3600)))
			})

			It("should be thread-safe", func() {
				done := make(chan bool)

				for i := 0; i < 10; i++ {
					go func(val int32) {
						p.SetSlowTime(val)
						done <- true
					}(int32(i))
				}

				for i := 0; i < 10; i++ {
					<-done
				}

				// Should have one of the values set
				slowTime := p.GetSlowTime()
				Expect(slowTime).To(BeNumerically(">=", 0))
				Expect(slowTime).To(BeNumerically("<", 10))
			})
		})

		Context("when getting slow time", func() {
			It("should return current value", func() {
				p.SetSlowTime(15)
				Expect(p.GetSlowTime()).To(Equal(int32(15)))

				p.SetSlowTime(25)
				Expect(p.GetSlowTime()).To(Equal(int32(25)))
			})
		})
	})

	Describe("Duration Configuration", func() {
		var p libprm.Prometheus

		BeforeEach(func() {
			p = newPrometheus()
		})

		Context("when setting duration buckets", func() {
			It("should add new duration buckets", func() {
				initialLen := len(p.GetDuration())
				p.SetDuration([]float64{15.0, 30.0})

				durations := p.GetDuration()
				Expect(durations).To(HaveLen(initialLen + 2))
				Expect(durations).To(ContainElement(15.0))
				Expect(durations).To(ContainElement(30.0))
			})

			It("should append to existing buckets", func() {
				p.SetDuration([]float64{0.5})
				p.SetDuration([]float64{1.5})

				durations := p.GetDuration()
				Expect(durations).To(ContainElement(0.5))
				Expect(durations).To(ContainElement(1.5))
			})

			It("should handle empty slice", func() {
				initialDurations := p.GetDuration()
				p.SetDuration([]float64{})

				Expect(p.GetDuration()).To(Equal(initialDurations))
			})

			It("should handle single value", func() {
				p.SetDuration([]float64{2.5})
				durations := p.GetDuration()
				Expect(durations).To(ContainElement(2.5))
			})

			It("should handle multiple values", func() {
				newBuckets := []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}
				p.SetDuration(newBuckets)

				durations := p.GetDuration()
				for _, bucket := range newBuckets {
					Expect(durations).To(ContainElement(bucket))
				}
			})

			It("should preserve order", func() {
				// Get initial state
				initial := p.GetDuration()
				newBuckets := []float64{20.0, 40.0, 60.0}

				p.SetDuration(newBuckets)
				durations := p.GetDuration()

				// Check that new buckets are appended in order
				startIdx := len(initial)
				for i, bucket := range newBuckets {
					Expect(durations[startIdx+i]).To(Equal(bucket))
				}
			})
		})

		Context("when getting duration buckets", func() {
			It("should return all configured buckets", func() {
				p.SetDuration([]float64{100.0})
				durations := p.GetDuration()

				Expect(durations).To(ContainElement(100.0))
			})

			It("should return a copy not affecting internal state", func() {
				durations := p.GetDuration()
				originalLen := len(durations)

				// Modify the returned slice
				durations = append(durations, 999.0)

				// Internal state should not change
				newDurations := p.GetDuration()
				Expect(newDurations).To(HaveLen(originalLen))
			})
		})
	})

	Describe("Thread Safety", func() {
		Context("when accessing configuration concurrently", func() {
			It("should handle concurrent SlowTime operations", func() {
				p := newPrometheus()
				done := make(chan bool, 100)

				// Concurrent writers
				for i := 0; i < 50; i++ {
					go func(val int32) {
						p.SetSlowTime(val)
						done <- true
					}(int32(i))
				}

				// Concurrent readers
				for i := 0; i < 50; i++ {
					go func() {
						_ = p.GetSlowTime()
						done <- true
					}()
				}

				// Wait for all goroutines
				for i := 0; i < 100; i++ {
					<-done
				}

				// Should complete without panic
				Expect(p.GetSlowTime()).To(BeNumerically(">=", 0))
			})

			It("should handle concurrent Duration operations", func() {
				p := newPrometheus()
				done := make(chan bool, 100)

				// Concurrent writers
				for i := 0; i < 50; i++ {
					go func(val float64) {
						p.SetDuration([]float64{val})
						done <- true
					}(float64(i))
				}

				// Concurrent readers
				for i := 0; i < 50; i++ {
					go func() {
						_ = p.GetDuration()
						done <- true
					}()
				}

				// Wait for all goroutines
				for i := 0; i < 100; i++ {
					<-done
				}

				// Should complete without panic
				Expect(p.GetDuration()).ToNot(BeEmpty())
			})
		})
	})
})
