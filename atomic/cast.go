/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package atomic

import "reflect"

// Cast attempts to safely convert any value to the target type M.
// Returns the converted value and true if successful, or the zero value of M and false if conversion fails.
//
// The function performs two checks:
//  1. Deep equality check to detect if src is already the zero value of M
//  2. Type assertion to convert src to M
//
// This is particularly useful when working with atomic.Value which stores values as interface{}.
//
// Example:
//
//	var v atomic.Value
//	v.Store(42)
//	num, ok := Cast[int](v.Load())  // num = 42, ok = true
//	str, ok := Cast[string](v.Load())  // str = "", ok = false
func Cast[M any](src any) (model M, casted bool) {
	if reflect.DeepEqual(src, model) {
		return model, false
	} else if v, k := src.(M); !k {
		return model, false
	} else {
		return v, true
	}
}

// IsEmpty checks if the source value is nil, zero, or cannot be cast to type M.
// Returns true if the value is empty or if Cast[M] fails, false otherwise.
//
// This is useful for validating values before storing them in atomic containers.
//
// Example:
//
//	empty := IsEmpty[string](nil)           // true
//	empty := IsEmpty[string]("")            // false (empty string is a valid string)
//	empty := IsEmpty[string](42)            // true (cannot cast int to string)
//	empty := IsEmpty[int](0)                // false (0 is a valid int)
func IsEmpty[M any](src any) bool {
	if _, k := Cast[M](src); !k {
		return true
	}

	return false
}
