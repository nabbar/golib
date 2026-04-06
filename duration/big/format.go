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
 */

package big

import (
	"fmt"
	"math"
	"strconv"
	"time"

	libpid "github.com/nabbar/golib/pidcontroller"
)

// Time returns a time.Duration representation of the duration.
// If the duration is larger than the maximum value of time.Duration,
// Time returns an error with the message "overflow max time.Duration value".
// Otherwise, Time returns the duration as a time.Duration.
//
// Example:
//
//	d, _ := durbig.Parse("1h30m")
//	td, err := d.Time()
//	if err != nil {
//	    panic(err)
//	}
//	fmt.Println(td) // Output: 1h30m0s
func (d Duration) Time() (time.Duration, error) {
	mxt := float64(math.MaxInt64) / float64(time.Second)
	if d.Float64() > mxt {
		return 0, fmt.Errorf("overflow max time.Duration value")
	}

	return time.Duration(d) * time.Second, nil
}

// String returns a string representation of the duration.
// The string is in the format "NdNhNmNs" where N is a number.
// The days are omitted if n is 0 or negative. The hours, minutes, and seconds
// are omitted if they are 0.
//
// Example:
//
//	d, _ := durbig.Parse("1d2h3m4s")
//	fmt.Println(d.String()) // Output: 1d2h3m4s
func (d Duration) String() string {
	var (
		p = make([]byte, 25)
		r uint64
	)

	if d < 0 {
		copy(p[0:], []byte{'-'})
	} else if d == 0 {
		return "0s"
	}

	// Hours
	r = stringUnit(d.Abs().Uint64(), Day.Uint64(), 'd', p)

	// Hours
	r = stringUnit(r, Hour.Uint64(), 'h', p)

	// Minutes
	r = stringUnit(r, Minute.Uint64(), 'm', p)

	// Seconds
	_ = stringUnit(r, Second.Uint64(), 's', p)

	return string(p[0:lastPos(p)])
}

// Days returns the total number of full days in the duration.
// The value is truncated towards zero.
//
// Example:
//
//	d := durbig.Days(5) + durbig.Hours(12)
//	fmt.Println(d.Days()) // Output: 5
func (d Duration) Days() int64 {
	t := math.Floor(d.Float64() / Day.Float64())
	return libpid.Float64ToInt64(t)
}

// Hours returns the total number of full hours in the duration.
// The value is truncated towards zero.
//
// Example:
//
//	d := durbig.Days(1) + durbig.Hours(2)
//	fmt.Println(d.Hours()) // Output: 26
func (d Duration) Hours() int64 {
	t := math.Floor(d.Float64() / Hour.Float64())
	return libpid.Float64ToInt64(t)
}

// Minutes returns the total number of full minutes in the duration.
// The value is truncated towards zero.
//
// Example:
//
//	d := durbig.Hours(1) + durbig.Minutes(30)
//	fmt.Println(d.Minutes()) // Output: 90
func (d Duration) Minutes() int64 {
	t := math.Floor(d.Float64() / Minute.Float64())
	return libpid.Float64ToInt64(t)
}

// Seconds returns the total number of full seconds in the duration.
// The value is truncated towards zero.
//
// Example:
//
//	d := durbig.Minutes(1) + durbig.Seconds(30)
//	fmt.Println(d.Seconds()) // Output: 90
func (d Duration) Seconds() int64 {
	t := math.Floor(d.Float64() / Second.Float64())
	return libpid.Float64ToInt64(t)
}

// Int64 returns the duration as a total number of seconds in an int64.
//
// Example:
//
//	d, _ := durbig.Parse("1h30m")
//	i := d.Int64()
//	fmt.Println(i) // Output: 5400
func (d Duration) Int64() int64 {
	return int64(d)
}

// Uint64 returns the duration as a total number of seconds in a uint64.
// If the duration is negative, it returns 0.
//
// Example:
//
//	d, _ := durbig.Parse("-1h30m")
//	i := d.Uint64()
//	fmt.Println(i) // Output: 0
func (d Duration) Uint64() uint64 {
	if i := int64(d); i < 0 {
		return uint64(0)
	} else {
		return uint64(i)
	}
}

// Uint32 returns the duration as a total number of seconds in a uint32.
// If the duration is negative, it returns 0.
//
// Example:
//
//	d, _ := durbig.Parse("-1h30m")
//	i := d.Uint32()
//	fmt.Println(i) // Output: 0
func (d Duration) Uint32() uint32 {
	if i := int64(d); i < 0 {
		return uint32(0)
	} else if i > math.MaxUint32 {
		return math.MaxUint32
	} else {
		return uint32(i)
	}
}

// Float64 returns the duration as a total number of seconds in a float64.
//
// Example:
//
//	d, _ := durbig.Parse("1h30m")
//	f := d.Float64()
//	fmt.Println(f) // Output: 5400
func (d Duration) Float64() float64 {
	return float64(d)
}

func stringUnit(val, div uint64, unit byte, buf []byte) uint64 {
	if val < div {
		// Value is smaller than the unit, so skip.
		return val
	}

	rst := val % div
	val = (val - rst) / div

	if val > 0 {
		idx := lastPos(buf)
		copy(buf[idx:], strconv.FormatUint(val, 10))
		idx = lastPos(buf)
		copy(buf[idx:], []byte{unit})
	}

	return rst
}

func lastPos(p []byte) int {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] != 0 {
			return i + 1
		}
	}
	return 0
}
