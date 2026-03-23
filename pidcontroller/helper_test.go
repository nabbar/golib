/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package pidcontroller_test

import (
	"math"

	"github.com/nabbar/golib/pidcontroller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PID Helper Functions", func() {
	Context("Int64ToFloat64", func() {
		// TC-HLP-001
		It("should correctly convert int64 to float64", func() {
			val := int64(12345)
			res := pidcontroller.Int64ToFloat64(val)
			Expect(res).To(Equal(12345.0))
		})

		// TC-HLP-001
		It("should handle negative values correctly", func() {
			val := int64(-9876)
			res := pidcontroller.Int64ToFloat64(val)
			Expect(res).To(Equal(-9876.0))
		})

		// TC-HLP-001
		It("should handle zero correctly", func() {
			val := int64(0)
			res := pidcontroller.Int64ToFloat64(val)
			Expect(res).To(Equal(0.0))
		})
	})

	Context("Float64ToInt64", func() {
		// TC-HLP-002
		It("should correctly convert float64 to int64 within range", func() {
			val := 12345.678
			res := pidcontroller.Float64ToInt64(val)
			Expect(res).To(Equal(int64(12345)))
		})

		// TC-HLP-002
		It("should handle negative float64 values within range", func() {
			val := -5432.10
			res := pidcontroller.Float64ToInt64(val)
			Expect(res).To(Equal(int64(-5432)))
		})

		// TC-HLP-002
		It("should clamp positive overflow to MaxInt64", func() {
			val := float64(math.MaxInt64) + 1000.0 // Value larger than MaxInt64
			res := pidcontroller.Float64ToInt64(val)
			Expect(res).To(Equal(int64(math.MaxInt64)))
		})

		// TC-HLP-002
		It("should clamp negative overflow to MinInt64", func() {
			// math.MinInt64 is -9223372036854775808
			// A float64 has enough range but less precision.
			// However, if we exceed the float representation of MinInt64 significantly:
			val := float64(math.MinInt64) * 2.0 // Very large negative number
			res := pidcontroller.Float64ToInt64(val)
			Expect(res).To(Equal(int64(math.MinInt64)))
		})

		// TC-HLP-002
		It("should handle zero correctly", func() {
			val := 0.0
			res := pidcontroller.Float64ToInt64(val)
			Expect(res).To(Equal(int64(0)))
		})
	})
})
