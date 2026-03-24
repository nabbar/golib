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

package duration_test

import (
	"testing"
	"time"

	libdur "github.com/nabbar/golib/duration"
)

// BenchmarkDuration_String benchmarks the String method of the Duration type.
// It measures the performance of converting a Duration to its string representation.
func BenchmarkDuration_String(b *testing.B) {
	d := libdur.Days(5) + libdur.Hours(23) + libdur.Minutes(15) + libdur.Seconds(13)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.String()
	}
}

// BenchmarkParse benchmarks the Parse function.
// It measures the performance of parsing a duration string into a Duration object.
func BenchmarkParse(b *testing.B) {
	s := "5d23h15m13s"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = libdur.Parse(s)
	}
}

// BenchmarkDuration_Days benchmarks the Days method.
// It measures the performance of calculating the number of days in a duration.
func BenchmarkDuration_Days(b *testing.B) {
	d := libdur.Days(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Days()
	}
}

// BenchmarkDuration_Hours benchmarks the Hours method.
// It measures the performance of calculating the number of hours in a duration.
func BenchmarkDuration_Hours(b *testing.B) {
	d := libdur.Hours(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Hours()
	}
}

// BenchmarkDuration_Minutes benchmarks the Minutes method.
// It measures the performance of calculating the number of minutes in a duration.
func BenchmarkDuration_Minutes(b *testing.B) {
	d := libdur.Minutes(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Minutes()
	}
}

// BenchmarkDuration_Seconds benchmarks the Seconds method.
// It measures the performance of calculating the number of seconds in a duration.
func BenchmarkDuration_Seconds(b *testing.B) {
	d := libdur.Seconds(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Seconds()
	}
}

// BenchmarkDuration_Truncate benchmarks the Truncate methods.
// It measures the performance of truncating a duration to different units.
func BenchmarkDuration_Truncate(b *testing.B) {
	d := libdur.ParseDuration(5*time.Hour + 30*time.Minute + 15*time.Second)
	b.Run("Days", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.TruncateDays()
		}
	})
	b.Run("Hours", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.TruncateHours()
		}
	})
	b.Run("Minutes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.TruncateMinutes()
		}
	})
	b.Run("Seconds", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = d.TruncateSeconds()
		}
	})
}
