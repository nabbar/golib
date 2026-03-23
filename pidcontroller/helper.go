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

package pidcontroller

import "math"

// Int64ToFloat64 converts an int64 value to a float64 representation.
// This is a simple type conversion helper.
func Int64ToFloat64(i int64) float64 {
	return float64(i)
}

// Float64ToInt64 converts a float64 value to an int64 representation.
// It handles potential overflow/underflow by clamping the value to the range of int64.
func Float64ToInt64(i float64) int64 {
	// Check for positive overflow
	// The maximum value an int64 can hold is 9223372036854775807.
	// When converted to float64, 9223372036854775807 becomes 9.223372036854776e+18.
	// This value is actually larger than math.MaxInt64.
	// If `i` is greater than or equal to this threshold, casting it to int64 results in undefined behavior (often wraps to minInt64).
	// We use 9.223372036854776e+18 as the upper bound check.
	if i >= 9.223372036854776e+18 {
		return math.MaxInt64
	}

	// Check for negative overflow
	// The minimum value an int64 can hold is -9223372036854775808.
	// This value is exactly representable as a float64 because it's a power of 2 (-2^63).
	// If `i` is less than this value, we return math.MinInt64.
	if i < -9.223372036854775808e+18 {
		return math.MinInt64
	}

	// Safe to convert
	return int64(i)
}
