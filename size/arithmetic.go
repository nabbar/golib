/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package size

import (
	"fmt"
	"math"
)

// Mul multiplies a Size by a float64 value.
//
// This is a convenience wrapper for MulErr which discards any potential errors.
// See the documentation for MulErr for more information.
func (s *Size) Mul(v float64) {
	_ = s.MulErr(v)
}

// MulErr multiplies a Size by a float64 value.
//
// If the result overflowed (i.e. be greater than the maximum
// allowable size), the function will return an error and set the
// Size to the maximum allowable size.
//
// Otherwise, it will set the Size to the result of the multiplication,
// rounded up to the nearest whole number.
func (s *Size) MulErr(v float64) error {
	if m := float64(math.MaxUint64) / v; m < s.Float64() {
		*s = math.MaxUint64
		return fmt.Errorf("overflow detected, force value to max allowable size")
	} else {
		*s = Size(math.Ceil(s.Float64() * v))
		return nil
	}
}

// Div divides a Size by a float64 value.
//
// This is a convenience wrapper for DivErr which discards any potential errors.
// See the documentation for DivErr for more information.
func (s *Size) Div(v float64) {
	_ = s.DivErr(v)
}

// DivErr divides a Size by a float64 value.
//
// If the divisor is 0 or negative, the function will return an error.
//
// Otherwise, it will set the Size to the result of the division,
// rounded up to the nearest whole number.
func (s *Size) DivErr(v float64) error {
	if v <= 0 {
		return fmt.Errorf("invalid diviser, only accepts positive float")
	}

	*s = Size(math.Ceil(s.Float64() / v))
	return nil
}

// Add adds a uint64 value to a Size.
//
// This is a convenience wrapper for AddErr which discards any potential errors.
// See the documentation for AddErr for more information.
func (s *Size) Add(v uint64) {
	_ = s.AddErr(v)
}

// AddErr adds a uint64 value to a Size.
//
// If the result overflowed (i.e. be greater than the maximum
// allowable size), the function will return an error and set the
// Size to the maximum allowable size.
//
// Otherwise, it will set the Size to the result of the addition,
// rounded up to the nearest whole number.
func (s *Size) AddErr(v uint64) error {
	if m := uint64(math.MaxUint64) - v; s.Uint64() > m {
		*s = math.MaxUint64
		return fmt.Errorf("overflow detected, force value to max allowable size")
	} else {
		*s = Size(s.Uint64() + v)
		return nil
	}
}

// Sub subtracts a uint64 value from a Size.
//
// This is a convenience wrapper for SubErr which discards any potential errors.
// See the documentation for SubErr for more information.
func (s *Size) Sub(v uint64) {
	_ = s.SubErr(v)
}

// SubErr subtracts a uint64 value from a Size.
//
// If the result underflow (i.e. be less than the minimum
// allowable size), the function will return an error and set the
// Size to the minimum allowable size.
//
// Otherwise, it will set the Size to the result of the subtraction.
func (s *Size) SubErr(v uint64) error {
	if v > s.Uint64() {
		*s = Size(0)
		return fmt.Errorf("invalid substractor, force value to min allowable size")
	}

	*s = Size(s.Uint64() - v)
	return nil
}
