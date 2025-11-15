/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package kvtypes

type CompareEqual[K comparable] func(ref, part K) bool
type CompareContains[K comparable] func(ref, part K) bool
type CompareEmpty[K comparable] func(part K) bool

type Compare[K comparable] interface {
	// IsEqual checks if two strings are equal.
	//
	// It takes two strings as parameters and returns true if the strings are equal,
	// false otherwise.
	//
	// If either of the strings passed as parameters are empty or nil, the method
	// will return false.
	IsEqual(ref, part K) bool
	// IsContains checks if a given string is contained in another.
	//
	// It takes two strings as parameters and returns true if the first string
	// is contained in the second, false otherwise.
	//
	// If the first string passed as parameter is empty or nil, the method will
	// return false.
	IsContains(ref, part K) bool
	// IsEmpty checks if a given string is empty.
	//
	// It takes a string as parameter and returns true if the string is empty,
	// false otherwise.
	//
	// If the string passed as parameter is nil, the method will return false.
	IsEmpty(part K) bool
}

// NewCompare creates a new Compare object.
//
// It takes three functions as parameters:
// - eq, which checks if two strings are equal.
// - cn, which checks if a string contains another.
// - em, which checks if a string is empty.
//
// It returns a Compare object which can be used to compare strings.
//
// If any of the functions passed as parameters are nil, the corresponding
// Compare method will return false.
func NewCompare[K comparable](eq CompareEqual[K], cn CompareContains[K], em CompareEmpty[K]) Compare[K] {
	return &cmp[K]{
		feq: eq,
		fcn: cn,
		fem: em,
	}
}

type cmp[K comparable] struct {
	feq CompareEqual[K]
	fcn CompareContains[K]
	fem CompareEmpty[K]
}

func (o *cmp[K]) IsEqual(ref, part K) bool {
	if o == nil || o.feq == nil {
		return false
	}

	return o.feq(ref, part)
}

func (o *cmp[K]) IsContains(ref, part K) bool {
	if o == nil || o.fcn == nil {
		return false
	}

	return o.fcn(ref, part)
}

func (o *cmp[K]) IsEmpty(part K) bool {
	if o == nil || o.fem == nil {
		return false
	}

	return o.fem(part)
}
