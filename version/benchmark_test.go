/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package version_test

import (
	"testing"

	"github.com/nabbar/golib/version"
)

// Benchmark for version creation
func BenchmarkNewVersion(b *testing.B) {
	type testStruct struct{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = version.NewVersion(
			version.License_MIT,
			"BenchApp",
			"Benchmark Application",
			"2024-01-15T10:30:00Z",
			"abc123",
			"v1.0.0",
			"Bench Author",
			"BENCH",
			testStruct{},
			0,
		)
	}
}

// Benchmark for GetHeader method
func BenchmarkGetHeader(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetHeader()
	}
}

// Benchmark for GetInfo method
func BenchmarkGetInfo(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetInfo()
	}
}

// Benchmark for GetLicenseLegal method
func BenchmarkGetLicenseLegal(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetLicenseLegal()
	}
}

// Benchmark for GetLicenseBoiler method
func BenchmarkGetLicenseBoiler(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetLicenseBoiler()
	}
}

// Benchmark for GetLicenseFull method
func BenchmarkGetLicenseFull(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetLicenseFull()
	}
}

// Benchmark for CheckGo method
func BenchmarkCheckGo(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.CheckGo("1.18", ">=")
	}
}

// Benchmark for multiple license combination
func BenchmarkGetLicenseLegal_Multiple(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.GetLicenseLegal(
			version.License_Apache_v2,
			version.License_GNU_GPL_v3,
		)
	}
}

// Benchmark for concurrent access
func BenchmarkConcurrentAccess(b *testing.B) {
	type testStruct struct{}
	v := version.NewVersion(
		version.License_MIT,
		"BenchApp",
		"Benchmark Application",
		"2024-01-15T10:30:00Z",
		"abc123",
		"v1.0.0",
		"Bench Author",
		"BENCH",
		testStruct{},
		0,
	)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = v.GetHeader()
			_ = v.GetInfo()
			_ = v.GetLicenseName()
		}
	})
}
