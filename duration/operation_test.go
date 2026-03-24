/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package duration_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

var _ = Describe("Duration Range Operations", func() {
	Describe("RangeTo", func() {
		It("should create range from smaller to larger duration", func() {
			// TC-OP-001
			start := libdur.Seconds(10)
			end := libdur.Seconds(100)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should ensure start is included", func() {
			// TC-OP-002
			start := libdur.Seconds(50)
			end := libdur.Seconds(100)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang[0]).To(Equal(start))
		})

		It("should ensure end is included", func() {
			// TC-OP-003
			start := libdur.Seconds(10)
			end := libdur.Seconds(50)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle equal start and end", func() {
			// TC-OP-004
			start := libdur.Seconds(50)
			end := libdur.Seconds(50)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).To(ContainElement(start))
		})

		It("should create at least 2 elements", func() {
			// TC-OP-005
			start := libdur.Seconds(10)
			end := libdur.Seconds(15)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(len(rang)).To(BeNumerically(">=", 2))
		})

		It("should have increasing values", func() {
			// TC-OP-006
			start := libdur.Seconds(10)
			end := libdur.Seconds(100)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			for i := 1; i < len(rang); i++ {
				Expect(rang[i]).To(BeNumerically(">=", rang[i-1]))
			}
		})
	})

	Describe("RangeDefTo", func() {
		It("should use default rates", func() {
			// TC-OP-007
			start := libdur.Seconds(10)
			end := libdur.Seconds(100)

			rang := start.RangeDefTo(end)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should create valid range with defaults", func() {
			// TC-OP-008
			start := libdur.Minutes(1)
			end := libdur.Minutes(10)

			rang := start.RangeDefTo(end)

			Expect(len(rang)).To(BeNumerically(">", 0))
		})
	})

	Describe("RangeFrom", func() {
		It("should create range from larger to smaller duration", func() {
			// TC-OP-009
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should ensure end is first", func() {
			// TC-OP-010
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang[0]).To(Equal(end))
		})

		It("should ensure start is last", func() {
			// TC-OP-011
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should handle equal start and end", func() {
			// TC-OP-012
			start := libdur.Seconds(50)
			end := libdur.Seconds(50)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang).To(ContainElement(start))
		})

		It("should create at least 2 elements", func() {
			// TC-OP-013
			start := libdur.Seconds(15)
			end := libdur.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(len(rang)).To(BeNumerically(">=", 2))
		})

		It("should have increasing values", func() {
			// TC-OP-014
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			for i := 1; i < len(rang); i++ {
				Expect(rang[i]).To(BeNumerically(">=", rang[i-1]))
			}
		})
	})

	Describe("RangeDefFrom", func() {
		It("should use default rates", func() {
			// TC-OP-015
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeDefFrom(end)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should create valid range with defaults", func() {
			// TC-OP-016
			start := libdur.Minutes(10)
			end := libdur.Minutes(1)

			rang := start.RangeDefFrom(end)

			Expect(len(rang)).To(BeNumerically(">", 0))
		})
	})

	Describe("Range Edge Cases", func() {
		It("should handle zero duration", func() {
			// TC-OP-017
			start := libdur.Seconds(0)
			end := libdur.Seconds(10)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
		})

		It("should handle negative duration", func() {
			// TC-OP-018
			start := libdur.Seconds(-10)
			end := libdur.Seconds(10)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
		})

		It("should handle very small range", func() {
			// TC-OP-019
			start := libdur.Seconds(10)
			end := libdur.Seconds(11)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(len(rang)).To(BeNumerically(">=", 2))
		})

		It("should handle very large range", func() {
			// TC-OP-020
			start := libdur.Seconds(1)
			end := libdur.Hours(24)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})
	})

	Describe("Default Rate Constants", func() {
		It("should have valid default rates", func() {
			// TC-OP-021
			Expect(libdur.DefaultRateProportional).To(Equal(0.1))
			Expect(libdur.DefaultRateIntegral).To(Equal(0.01))
			Expect(libdur.DefaultRateDerivative).To(Equal(0.05))
		})
	})

	Describe("RangeCtxTo with Context", func() {
		It("should respect context timeout", func() {
			// TC-OP-022
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := libdur.Seconds(1)
			end := libdur.Hours(1000) // Very large range

			rang := start.RangeCtxTo(ctx, end, 0.001, 0.0001, 0.0005)

			// Should complete even with timeout (may return partial or fallback results)
			Expect(rang).ToNot(BeNil())
		})

		It("should work with valid context", func() {
			// TC-OP-023
			ctx := context.Background()
			start := libdur.Seconds(10)
			end := libdur.Seconds(100)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle cancelled context gracefully", func() {
			// TC-OP-024
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			start := libdur.Seconds(10)
			end := libdur.Seconds(100)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			// Should still return valid range (fallback behavior)
			Expect(rang).ToNot(BeNil())
		})

		It("should ensure minimum 2 elements even with short timeout", func() {
			// TC-OP-025
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			start := libdur.Seconds(10)
			end := libdur.Seconds(20)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			// Fallback ensures at least start and end
			Expect(len(rang)).To(BeNumerically(">=", 2))
		})
	})

	Describe("RangeCtxFrom with Context", func() {
		It("should respect context timeout", func() {
			// TC-OP-026
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := libdur.Hours(1000) // Very large value
			end := libdur.Seconds(1)

			rang := start.RangeCtxFrom(ctx, end, 0.001, 0.0001, 0.0005)

			// Should complete even with timeout
			Expect(rang).ToNot(BeNil())
		})

		It("should work with valid context", func() {
			// TC-OP-027
			ctx := context.Background()
			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeCtxFrom(ctx, end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should handle cancelled context gracefully", func() {
			// TC-OP-028
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			start := libdur.Seconds(100)
			end := libdur.Seconds(10)

			rang := start.RangeCtxFrom(ctx, end, 0.1, 0.01, 0.05)

			// Should still return valid range (fallback behavior)
			Expect(rang).ToNot(BeNil())
		})
	})

	Describe("Range with Custom PID Parameters", func() {
		It("should handle very small rates", func() {
			// TC-OP-029
			start := libdur.Seconds(10)
			end := libdur.Seconds(20)

			rang := start.RangeTo(end, 0.001, 0.0001, 0.0005)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle very large rates", func() {
			// TC-OP-030
			start := libdur.Seconds(10)
			end := libdur.Seconds(20)

			rang := start.RangeTo(end, 0.9, 0.8, 0.7)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})
	})

	Describe("Range Performance", func() {
		It("should complete RangeTo in reasonable time", func() {
			// TC-OP-031
			start := libdur.Seconds(1)
			end := libdur.Minutes(10)

			startTime := time.Now()
			rang := start.RangeTo(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 5*time.Second))
		})

		It("should complete RangeFrom in reasonable time", func() {
			// TC-OP-032
			start := libdur.Minutes(10)
			end := libdur.Seconds(1)

			startTime := time.Now()
			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 5*time.Second))
		})
	})
})
