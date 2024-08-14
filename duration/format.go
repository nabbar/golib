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

package duration

import (
	"fmt"
	"math"
	"time"
)

func (d Duration) Time() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	var (
		s string
		n = d.Days()
		i = d.Time()
	)

	if n > 0 {
		i = i - (time.Duration(n) * 24 * time.Hour)
		s = fmt.Sprintf("%dd", n)
	}

	if n < 1 || i > 0 {
		s = s + i.String()
	}

	return s
}

func (d Duration) Days() int64 {
	t := math.Floor(d.Time().Hours() / 24)

	if t > math.MaxInt64 {
		return math.MaxInt64
	}

	return int64(t)
}
