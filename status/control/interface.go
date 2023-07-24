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

package control

import (
	"math"
	"strings"
)

type Mode uint8

const (
	Ignore Mode = iota
	Should
	Must
	One
)

func Parse(s string) Mode {
	switch {
	case strings.EqualFold(Must.Code(), s):
		return Must
	case strings.EqualFold(One.Code(), s):
		return One
	case strings.EqualFold(Should.Code(), s):
		return Should
	}

	return Ignore
}

func ParseBytes(p []byte) Mode {
	return Parse(string(p))
}

func ParseInt64(val int64) Mode {
	if val > int64(math.MaxUint8) {
		return Mode(math.MaxUint8)
	}

	return Mode(uint8(val))
}
