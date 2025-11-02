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

package types_test

import (
	prmtps "github.com/nabbar/golib/prometheus/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricType Values", func() {
	Describe("MetricType Constants", func() {
		Context("when checking type values", func() {
			It("should have None as zero value", func() {
				var mt prmtps.MetricType
				Expect(mt).To(Equal(prmtps.None))
				Expect(int(prmtps.None)).To(Equal(0))
			})

			It("should have Counter as next value", func() {
				Expect(int(prmtps.Counter)).To(Equal(1))
			})

			It("should have Gauge as next value", func() {
				Expect(int(prmtps.Gauge)).To(Equal(2))
			})

			It("should have Histogram as next value", func() {
				Expect(int(prmtps.Histogram)).To(Equal(3))
			})

			It("should have Summary as next value", func() {
				Expect(int(prmtps.Summary)).To(Equal(4))
			})

			It("should have distinct values for all types", func() {
				types := []prmtps.MetricType{
					prmtps.None,
					prmtps.Counter,
					prmtps.Gauge,
					prmtps.Histogram,
					prmtps.Summary,
				}

				// Check all are distinct
				seen := make(map[prmtps.MetricType]bool)
				for _, t := range types {
					Expect(seen[t]).To(BeFalse(), "Type %v should be unique", t)
					seen[t] = true
				}
			})
		})

		Context("when comparing types", func() {
			It("should allow equality comparisons", func() {
				Expect(prmtps.Counter).To(Equal(prmtps.Counter))
				Expect(prmtps.Gauge).To(Equal(prmtps.Gauge))
				Expect(prmtps.Histogram).To(Equal(prmtps.Histogram))
				Expect(prmtps.Summary).To(Equal(prmtps.Summary))
				Expect(prmtps.None).To(Equal(prmtps.None))
			})

			It("should distinguish different types", func() {
				Expect(prmtps.Counter).ToNot(Equal(prmtps.Gauge))
				Expect(prmtps.Counter).ToNot(Equal(prmtps.Histogram))
				Expect(prmtps.Counter).ToNot(Equal(prmtps.Summary))
				Expect(prmtps.Counter).ToNot(Equal(prmtps.None))

				Expect(prmtps.Gauge).ToNot(Equal(prmtps.Histogram))
				Expect(prmtps.Gauge).ToNot(Equal(prmtps.Summary))
				Expect(prmtps.Gauge).ToNot(Equal(prmtps.None))

				Expect(prmtps.Histogram).ToNot(Equal(prmtps.Summary))
				Expect(prmtps.Histogram).ToNot(Equal(prmtps.None))

				Expect(prmtps.Summary).ToNot(Equal(prmtps.None))
			})

			It("should work in switch statements", func() {
				testType := func(t prmtps.MetricType) string {
					switch t {
					case prmtps.None:
						return "none"
					case prmtps.Counter:
						return "counter"
					case prmtps.Gauge:
						return "gauge"
					case prmtps.Histogram:
						return "histogram"
					case prmtps.Summary:
						return "summary"
					default:
						return "unknown"
					}
				}

				Expect(testType(prmtps.None)).To(Equal("none"))
				Expect(testType(prmtps.Counter)).To(Equal("counter"))
				Expect(testType(prmtps.Gauge)).To(Equal("gauge"))
				Expect(testType(prmtps.Histogram)).To(Equal("histogram"))
				Expect(testType(prmtps.Summary)).To(Equal("summary"))
			})

			It("should work in map keys", func() {
				typeMap := map[prmtps.MetricType]string{
					prmtps.None:      "none",
					prmtps.Counter:   "counter",
					prmtps.Gauge:     "gauge",
					prmtps.Histogram: "histogram",
					prmtps.Summary:   "summary",
				}

				Expect(typeMap[prmtps.Counter]).To(Equal("counter"))
				Expect(typeMap[prmtps.Gauge]).To(Equal("gauge"))
				Expect(typeMap[prmtps.Histogram]).To(Equal("histogram"))
				Expect(typeMap[prmtps.Summary]).To(Equal("summary"))
				Expect(typeMap[prmtps.None]).To(Equal("none"))
			})
		})

		Context("when using in arrays and slices", func() {
			It("should work in slices", func() {
				validTypes := []prmtps.MetricType{
					prmtps.Counter,
					prmtps.Gauge,
					prmtps.Histogram,
					prmtps.Summary,
				}

				Expect(validTypes).To(HaveLen(4))
				Expect(validTypes).To(ContainElement(prmtps.Counter))
				Expect(validTypes).To(ContainElement(prmtps.Gauge))
				Expect(validTypes).To(ContainElement(prmtps.Histogram))
				Expect(validTypes).To(ContainElement(prmtps.Summary))
				Expect(validTypes).ToNot(ContainElement(prmtps.None))
			})

			It("should support range iteration", func() {
				types := []prmtps.MetricType{
					prmtps.Counter,
					prmtps.Gauge,
					prmtps.Histogram,
					prmtps.Summary,
				}

				count := 0
				for _, t := range types {
					Expect(t).ToNot(Equal(prmtps.None))
					count++
				}
				Expect(count).To(Equal(4))
			})
		})
	})

	Describe("Type Safety", func() {
		Context("when using metric types", func() {
			It("should maintain type safety", func() {
				var mt prmtps.MetricType = prmtps.Counter

				// Type should be preserved
				Expect(mt).To(BeAssignableToTypeOf(prmtps.MetricType(0)))
				Expect(mt).To(Equal(prmtps.Counter))
			})

			It("should allow casting from int", func() {
				mt := prmtps.MetricType(1)
				Expect(mt).To(Equal(prmtps.Counter))

				mt = prmtps.MetricType(2)
				Expect(mt).To(Equal(prmtps.Gauge))
			})

			It("should allow casting to int", func() {
				Expect(int(prmtps.Counter)).To(Equal(1))
				Expect(int(prmtps.Gauge)).To(Equal(2))
				Expect(int(prmtps.Histogram)).To(Equal(3))
				Expect(int(prmtps.Summary)).To(Equal(4))
			})
		})
	})
})
