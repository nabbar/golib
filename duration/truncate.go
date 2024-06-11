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
	"math"
	"time"
)

func (d Duration) TruncateMicroseconds() Duration {
	return Duration(time.Duration(d.Time().Microseconds()) * time.Microsecond)
}

func (d Duration) TruncateMilliseconds() Duration {
	return Duration(time.Duration(d.Time().Milliseconds()) * time.Millisecond)
}

func (d Duration) TruncateSeconds() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Seconds())) * time.Second)
}

func (d Duration) TruncateMinutes() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Minutes())) * time.Minute)
}

func (d Duration) TruncateHours() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Hours())) * time.Hour)
}

func (d Duration) TruncateDays() Duration {
	return Duration(time.Duration(d.Days()) * 24 * time.Hour)
}
