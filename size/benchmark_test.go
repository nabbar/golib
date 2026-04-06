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

package size_test

import (
	"math"
	"testing"

	libsiz "github.com/nabbar/golib/size"
)

// BenchmarkParse benchmarks the Parse function with a composite string.
func BenchmarkParse(b *testing.B) {
	s := "5EB23PB15TB13KB"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = libsiz.Parse(s)
	}
}

// BenchmarkParse_Cases benchmarks the Parse function with various single unit cases.
func BenchmarkParse_Cases(b *testing.B) {
	testCases := map[string]string{
		"B":   "100B",
		"KB":  "100KB",
		"MB":  "100MB",
		"GB":  "100GB",
		"TB":  "100TB",
		"PB":  "100PB",
		"EB":  "10EB",
		"Dec": "1.5GB",
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = libsiz.Parse(tc)
			}
		})
	}
}

// BenchmarkParse_Cases benchmarks the Parse function with various single unit cases.
func BenchmarkParse_CasesSpace(b *testing.B) {
	testCases := map[string]string{
		"B":   "100 B",
		"KB":  "100 KB",
		"MB":  "100 MB",
		"GB":  "100 GB",
		"TB":  "100 TB",
		"PB":  "100 PB",
		"EB":  "10 EB",
		"Dec": "1.5 GB",
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = libsiz.Parse(tc)
			}
		})
	}
}

// BenchmarkString benchmarks the String method with a large size.
func BenchmarkString(b *testing.B) {
	s := libsiz.SizeExa + libsiz.SizePeta*23 + libsiz.SizeTera*15 + libsiz.SizeKilo*13
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.String()
	}
}

// BenchmarkString_Cases benchmarks the String method with various unit thresholds.
func BenchmarkString_Cases(b *testing.B) {
	testCases := map[string]libsiz.Size{
		"B":   libsiz.ParseUint64(100),
		"KB":  libsiz.SizeKilo * 100,
		"MB":  libsiz.SizeMega * 100,
		"GB":  libsiz.SizeGiga * 100,
		"TB":  libsiz.SizeTera * 100,
		"PB":  libsiz.SizePeta * 100,
		"EB":  libsiz.SizeExa * 10,
		"Max": libsiz.Size(math.MaxUint64),
	}

	for name, tc := range testCases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.String()
			}
		})
	}
}

// BenchmarkAdd benchmarks the Add method.
func BenchmarkAdd(b *testing.B) {
	s := libsiz.ParseUint64(1024)
	val := uint64(512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(val)
	}
}

// BenchmarkSub benchmarks the Sub method.
func BenchmarkSub(b *testing.B) {
	s := libsiz.ParseUint64(1024)
	val := uint64(512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Sub(val)
	}
}

// BenchmarkMul benchmarks the Mul method.
func BenchmarkMul(b *testing.B) {
	s := libsiz.ParseUint64(1024)
	val := 2.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Mul(val)
	}
}

// BenchmarkDiv benchmarks the Div method.
func BenchmarkDiv(b *testing.B) {
	s := libsiz.ParseUint64(1024)
	val := 2.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Div(val)
	}
}
