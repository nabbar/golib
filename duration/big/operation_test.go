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

package big_test

import (
	"context"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Operation Functions", func() {
	Describe("Abs", func() {
		It("should return positive duration unchanged", func() {
			d := durbig.Seconds(100)
			Expect(d.Abs()).To(Equal(d))
		})

		It("should return zero duration unchanged", func() {
			d := durbig.Seconds(0)
			Expect(d.Abs()).To(Equal(d))
		})

		It("should convert negative to positive", func() {
			d := durbig.Seconds(-100)
			Expect(d.Abs()).To(Equal(durbig.Seconds(100)))
		})

		It("should handle negative days", func() {
			d := durbig.Days(-5)
			Expect(d.Abs()).To(Equal(durbig.Days(5)))
		})

		It("should handle minimum duration", func() {
			d := durbig.Duration(math.MinInt64)
			abs := d.Abs()
			Expect(abs).To(Equal(durbig.Duration(math.MaxInt64)))
		})

		It("should handle large negative values", func() {
			d := durbig.Seconds(-1000000)
			Expect(d.Abs()).To(Equal(durbig.Seconds(1000000)))
		})
	})

	Describe("RangeTo", func() {
		It("should generate range from smaller to larger duration", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should include start and end points", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should generate monotonically increasing values", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			for i := 1; i < len(rang); i++ {
				Expect(rang[i].Int64()).To(BeNumerically(">=", rang[i-1].Int64()))
			}
		})

		It("should handle zero start", func() {
			start := durbig.Seconds(0)
			end := durbig.Seconds(10)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
		})

		It("should complete within timeout", func() {
			start := durbig.Seconds(1)
			end := durbig.Minutes(10)

			startTime := time.Now()
			rang := start.RangeTo(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 6*time.Second))
		})
	})

	Describe("RangeDefTo", func() {
		It("should use default PID parameters", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeDefTo(end)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should generate reasonable step count", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeDefTo(end)

			Expect(len(rang)).To(BeNumerically(">=", 2))
			Expect(len(rang)).To(BeNumerically("<", 1000))
		})
	})

	Describe("RangeFrom", func() {
		It("should generate range from larger to smaller duration", func() {
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should include start and end points", func() {
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should generate monotonically increasing values", func() {
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)

			for i := 1; i < len(rang); i++ {
				Expect(rang[i].Int64()).To(BeNumerically(">=", rang[i-1].Int64()))
			}
		})

		It("should complete within timeout", func() {
			start := durbig.Minutes(10)
			end := durbig.Seconds(1)

			startTime := time.Now()
			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 6*time.Second))
		})
	})

	Describe("RangeDefFrom", func() {
		It("should use default PID parameters", func() {
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeDefFrom(end)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should generate reasonable step count", func() {
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeDefFrom(end)

			Expect(len(rang)).To(BeNumerically(">=", 2))
			Expect(len(rang)).To(BeNumerically("<", 1000))
		})
	})

	Describe("RangeCtxTo with Context", func() {
		It("should respect context timeout", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := durbig.Seconds(1)
			end := durbig.Hours(1000) // Very large range

			rang := start.RangeCtxTo(ctx, end, 0.001, 0.0001, 0.0005)

			// Should complete even with timeout (may return partial or fallback results)
			Expect(rang).ToNot(BeNil())
		})

		It("should work with valid context", func() {
			ctx := context.Background()
			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle cancelled context gracefully", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			start := durbig.Seconds(10)
			end := durbig.Seconds(100)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			// Should still return valid range (fallback behavior)
			Expect(rang).ToNot(BeNil())
		})

		It("should ensure minimum 2 elements even with short timeout", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeCtxTo(ctx, end, 0.1, 0.01, 0.05)

			// Fallback ensures at least start and end
			Expect(len(rang)).To(BeNumerically(">=", 2))
		})
	})

	Describe("RangeCtxFrom with Context", func() {
		It("should respect context timeout", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			start := durbig.Hours(1000) // Very large value
			end := durbig.Seconds(1)

			rang := start.RangeCtxFrom(ctx, end, 0.001, 0.0001, 0.0005)

			// Should complete even with timeout
			Expect(rang).ToNot(BeNil())
		})

		It("should work with valid context", func() {
			ctx := context.Background()
			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeCtxFrom(ctx, end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(end))
			Expect(rang[len(rang)-1]).To(Equal(start))
		})

		It("should handle cancelled context gracefully", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			start := durbig.Seconds(100)
			end := durbig.Seconds(10)

			rang := start.RangeCtxFrom(ctx, end, 0.1, 0.01, 0.05)

			// Should still return valid range (fallback behavior)
			Expect(rang).ToNot(BeNil())
		})
	})

	Describe("Range with Custom PID Parameters", func() {
		It("should handle very small rates", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeTo(end, 0.001, 0.0001, 0.0005)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle very large rates", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeTo(end, 0.9, 0.8, 0.7)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle zero rates", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeTo(end, 0, 0, 0)

			// Should still work with fallback
			Expect(len(rang)).To(BeNumerically(">=", 2))
		})

		It("should handle negative rates", func() {
			start := durbig.Seconds(10)
			end := durbig.Seconds(20)

			rang := start.RangeTo(end, -0.1, -0.01, -0.05)

			// Should still work with fallback
			Expect(len(rang)).To(BeNumerically(">=", 2))
		})
	})

	Describe("Default Rate Constants", func() {
		It("should have valid default rates", func() {
			Expect(durbig.DefaultRateProportional).To(Equal(0.1))
			Expect(durbig.DefaultRateIntegral).To(Equal(0.01))
			Expect(durbig.DefaultRateDerivative).To(Equal(0.05))
		})
	})

	Describe("Range Performance", func() {
		It("should complete RangeTo in reasonable time", func() {
			start := durbig.Seconds(1)
			end := durbig.Minutes(10)

			startTime := time.Now()
			rang := start.RangeTo(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 6*time.Second))
		})

		It("should complete RangeFrom in reasonable time", func() {
			start := durbig.Minutes(10)
			end := durbig.Seconds(1)

			startTime := time.Now()
			rang := start.RangeFrom(end, 0.1, 0.01, 0.05)
			duration := time.Since(startTime)

			Expect(rang).ToNot(BeEmpty())
			Expect(duration).To(BeNumerically("<", 6*time.Second))
		})
	})

	Describe("Range with Days", func() {
		It("should handle range with days", func() {
			start := durbig.Days(1)
			end := durbig.Days(10)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})

		It("should handle mixed units", func() {
			start := durbig.Hours(12)
			end := durbig.Days(2)

			rang := start.RangeTo(end, 0.1, 0.01, 0.05)

			Expect(rang).ToNot(BeEmpty())
			Expect(rang[0]).To(Equal(start))
			Expect(rang[len(rang)-1]).To(Equal(end))
		})
	})
})
