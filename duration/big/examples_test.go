/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package big_test

import (
	"fmt"
	"time"

	durbig "github.com/nabbar/golib/duration/big"
)

func ExampleParse() {
	// Parse a duration string
	d, err := durbig.Parse("10d5h30m")
	if err != nil {
		panic(err)
	}

	fmt.Println(d.String())
	// Output: 10d5h30m
}

func ExampleDays() {
	// Create a duration of 500 days
	d := durbig.Days(500)
	fmt.Println(d.String())
	// Output: 500d
}

func ExampleHours() {
	// Create a duration of 24 hours
	d := durbig.Hours(24)
	fmt.Println(d.String())
	// Output: 1d
}

func ExampleMinutes() {
	// Create a duration of 60 minutes
	d := durbig.Minutes(60)
	fmt.Println(d.String())
	// Output: 1h
}

func ExampleSeconds() {
	// Create a duration of 3600 seconds
	d := durbig.Seconds(3600)
	fmt.Println(d.String())
	// Output: 1h
}

func ExampleDuration_String() {
	d := durbig.Days(1) + durbig.Hours(2) + durbig.Minutes(30) + durbig.Seconds(45)
	fmt.Println(d.String())
	// Output: 1d2h30m45s
}

func ExampleDuration_Time() {
	// Convert big.Duration to time.Duration
	d := durbig.Hours(1)
	td, err := d.Time()
	if err != nil {
		panic(err)
	}
	fmt.Println(td)
	// Output: 1h0m0s
}

func ExampleDuration_Truncate() {
	d := durbig.ParseFloat64(123.456)
	// Truncate to seconds (implied, as base unit is seconds)
	// Truncating to a larger unit like minutes
	d = durbig.Minutes(1) + durbig.Seconds(30)
	t := d.Truncate(durbig.Minute)
	fmt.Println(t.String())
	// Output: 1m
}

func ExampleDuration_Round() {
	d := durbig.Minutes(1) + durbig.Seconds(30)
	// Round to the nearest minute (half up)
	r := d.Round(durbig.Minute)
	fmt.Println(r.String())
	// Output: 2m
}

func ExampleDuration_IsDays() {
	d := durbig.Days(2)
	fmt.Println(d.IsDays())
	// Output: true
}

func ExampleDuration_IsHours() {
	d := durbig.Hours(2)
	fmt.Println(d.IsHours())
	// Output: true
}

func ExampleDuration_IsMinutes() {
	d := durbig.Minutes(2)
	fmt.Println(d.IsMinutes())
	// Output: true
}

func ExampleDuration_IsSeconds() {
	d := durbig.Seconds(2)
	fmt.Println(d.IsSeconds())
	// Output: true
}

func ExampleParseFloat64() {
	// Parse a float64 as seconds
	d := durbig.ParseFloat64(3600.5)
	fmt.Println(d.String())
	// Output: 1h1s
}

func ExampleParseDuration() {
	// Parse a time.Duration
	td := time.Hour + 30*time.Minute
	d := durbig.ParseDuration(td)
	fmt.Println(d.String())
	// Output: 1h30m
}
