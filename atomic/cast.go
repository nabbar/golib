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

// Cast attempts to safely convert any value to the target type M while performing a zero-value validation.
//
// This function is a core component of the atomic package's safety mechanism. It returns the converted
// value and true if successful, or the zero value of M and false if:
//  1. The source interface is nil.
//  2. The source value cannot be asserted to type M.
//  3. The source value is the "zero value" or "empty value" for its type.
//
// # PERFORMANCE DESIGN
//
// To avoid the massive overhead of reflect.DeepEqual (which was identified as a major bottleneck),
// Cast uses a tiered evaluation strategy designed for nanosecond-level latency:
//  - Immediate exit for nil interfaces to avoid any further processing cost.
//  - Native Go type assertion (src.(M)). This is the fastest way to check type compatibility.
//  - Specialized Type Switching: For all Go scalar types (integers, strings, floats, bools),
//    it performs direct comparisons against zero constants. This bypasses reflection entirely.
//  - Reflection Fallback: For complex types (structs, slices, maps, channels, functions, pointers),
//    it uses reflect.Value.IsZero(), which is the modern, recommended way to check for zero-state
//    without the recursive cost of DeepEqual.
//
// Parameters:
//   - src: The source value of any type to be converted and validated.
//
// Returns:
//   - model: The converted value of type M, or the zero value of M on failure.
//   - casted: A boolean indicating if the cast was successful AND the value is not empty.
func Cast[M any](src any) (model M, casted bool) {
	// Immediate exit for nil interfaces to avoid any further processing cost.
	if src == nil {
		return model, false
	}

	// Native Go type assertion. This is the fastest way to check type compatibility.
	// This ensures that we are working with the expected type before doing zero-value checks.
	v, ok := src.(M)
	if !ok {
		return model, false
	}

	// Zero-value detection optimized by type.
	// We convert 'v' back to 'any' to perform a type switch on the concrete value.
	// This switch covers all primitive Go types to ensure maximum performance.
	switch va := any(v).(type) {
	case bool:
		if !va {
			return model, false
		}
	case int:
		if va == 0 {
			return model, false
		}
	case int8:
		if va == 0 {
			return model, false
		}
	case int16:
		if va == 0 {
			return model, false
		}
	case int32:
		if va == 0 {
			return model, false
		}
	case int64:
		if va == 0 {
			return model, false
		}
	case uint:
		if va == 0 {
			return model, false
		}
	case uint8:
		if va == 0 {
			return model, false
		}
	case uint16:
		if va == 0 {
			return model, false
		}
	case uint32:
		if va == 0 {
			return model, false
		}
	case uint64:
		if va == 0 {
			return model, false
		}
	case uintptr:
		if va == 0 {
			return model, false
		}
	case float32:
		if va == 0 {
			return model, false
		}
	case float64:
		if va == 0 {
			return model, false
		}
	case complex64:
		if va == 0 {
			return model, false
		}
	case complex128:
		if va == 0 {
			return model, false
		}
	case string:
		if va == "" {
			return model, false
		}
	default:
		// Fallback for complex types (structs, slices, maps, channels, functions, pointers).
		// We use reflect.ValueOf(v) to get a reflection-enabled view of the value.
		rv := reflect.ValueOf(v)

		// IsValid() returns false if the value is zero-reflect-value (like nil).
		// IsZero() returns true if the value is the zero value for its type.
		if !rv.IsValid() || rv.IsZero() {
			return model, false
		}
	}

	return v, true
}

// IsEmpty checks if the source value is nil, zero, or cannot be cast to type M.
//
// This is a convenience wrapper around Cast[M] used primarily for input validation
// in Store operations. It returns true if the value is considered "vacant" or
// "invalid" for the given type M.
//
// # OPTIMIZATION
//
// It performs a direct nil check before invoking the more complex Cast logic
// to maximize performance for nil inputs, which are common in concurrent state management.
func IsEmpty[M any](src any) bool {
	if src == nil {
		return true
	}

	// Delegate to Cast for full type-aware zero-value detection.
	if _, k := Cast[M](src); !k {
		return true
	}

	return false
}
