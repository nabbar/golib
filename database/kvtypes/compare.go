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
	IsEqual(ref, part K) bool
	IsContains(ref, part K) bool
	IsEmpty(part K) bool
}

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
