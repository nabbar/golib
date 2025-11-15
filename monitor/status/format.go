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

package status

import "strings"

// String returns a string representation of the Status.
//
// The string representation is as follows:
//
// - OK: "OK"
// - Warn: "Warn"
// - Any other value: "KO"
//
// This is useful for logging or displaying the status.
func (o Status) String() string {
	switch o {
	case OK:
		return "OK"
	case Warn:
		return "Warn"
	default:
		return "KO"
	}
}

// Code returns the uppercase string representation of the Status.
// Unlike String(), this always returns uppercase values.
//
// Returns:
//   - "OK" for OK status
//   - "WARN" for Warn status
//   - "KO" for KO status
func (o Status) Code() string {
	return strings.ToUpper(o.String())
}

// Int returns an int representation of the Status.
//
// Returns:
//   - 0 for KO status
//   - 1 for Warn status
//   - 2 for OK status
//
// This is useful for storing the status in a database or using it in calculations.
func (o Status) Int() int {
	return int(o)
}

// Int64 returns an int64 representation of the Status.
// See Int() for value mappings.
func (o Status) Int64() int64 {
	return int64(o)
}

// Uint returns an unsigned int representation of the Status.
// See Int() for value mappings.
func (o Status) Uint() uint {
	return uint(o)
}

// Uint64 returns an uint64 representation of the Status.
// See Int() for value mappings.
func (o Status) Uint64() uint64 {
	return uint64(o)
}

// Float returns a float64 representation of the Status.
//
// Returns:
//   - 0.0 for KO status
//   - 1.0 for Warn status
//   - 2.0 for OK status
//
// This is useful for storing the status in a database or using it in calculations.
func (o Status) Float() float64 {
	return float64(o)
}
