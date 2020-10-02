/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package network

import (
	"fmt"
	"math"
	"strings"
)

const (
	_PowBytesNumber_ float64 = 2
	_PowBytesPower_  float64 = 10
	_PowBytesFactor_ int     = 3
)

type Bytes uint64

func (n Bytes) String() string {
	return fmt.Sprintf("%d", n)
}

func (n Bytes) FormatUnitFloat(precision int) string {
	if precision < 1 {
		return n.FormatUnitInt()
	}

	m := float64(n)

	f := fmt.Sprintf("%%.%df", precision)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		r := m / math.Pow(math.Pow(_PowBytesNumber_, _PowBytesPower_), float64(p/_PowBytesFactor_))
		q := strings.SplitN(fmt.Sprintf(f, r), ".", 2)

		if len(q) > 0 {
			if len(q[0]) < _MaxSizeOfPad_ {
				return strings.Repeat(" ", _MaxSizeOfPad_-len(q[0])) + fmt.Sprintf(f+" %s", r, power2Unit(p)+"B")
			}
			return fmt.Sprintf(f+" %s", r, power2Unit(p)+"B")
		}
	}

	return strings.Repeat(" ", _MaxSizeOfPad_) + fmt.Sprintf(f+" %s", m, " ")
}

func (n Bytes) FormatUnitInt() string {
	m := float64(n)

	for _, p := range powerList() {
		if m < math.Pow10(p+1) {
			continue
		}

		return fmt.Sprintf(_PadIntPattern_+" %s", int(math.Round(m/math.Pow(math.Pow(_PowBytesNumber_, _PowBytesPower_), float64(p/_PowBytesFactor_)))), power2Unit(p)+"B")
	}

	return fmt.Sprintf(_PadIntPattern_+" %s", n, " ")
}

func (n Bytes) AsNumber() Number {
	return Number(n)
}

func (n Bytes) AsUint64() uint64 {
	return uint64(n)
}

func (n Bytes) AsFloat64() float64 {
	return float64(n)
}
