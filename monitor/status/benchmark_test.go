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

package status_test

import (
	"encoding/json"
	"testing"

	"github.com/nabbar/golib/monitor/status"
)

// BenchmarkString measures String() performance
func BenchmarkString(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.String()
	}
}

// BenchmarkInt measures Int() performance
func BenchmarkInt(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.Int()
	}
}

// BenchmarkFloat measures Float() performance
func BenchmarkFloat(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.Float()
	}
}

// BenchmarkNewFromString measures NewFromString performance
func BenchmarkNewFromString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.NewFromString("OK")
	}
}

// BenchmarkNewFromStringLowercase measures NewFromString with lowercase
func BenchmarkNewFromStringLowercase(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.NewFromString("ok")
	}
}

// BenchmarkNewFromInt measures NewFromInt performance
func BenchmarkNewFromInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.NewFromInt(2)
	}
}

// BenchmarkNewFromIntInvalid measures NewFromInt with invalid value
func BenchmarkNewFromIntInvalid(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.NewFromInt(999)
	}
}

// BenchmarkMarshalJSON measures MarshalJSON performance
func BenchmarkMarshalJSON(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.MarshalJSON()
	}
}

// BenchmarkUnmarshalJSONString measures UnmarshalJSON with string
func BenchmarkUnmarshalJSONString(b *testing.B) {
	data := []byte(`"OK"`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s status.Status
		_ = s.UnmarshalJSON(data)
	}
}

// BenchmarkUnmarshalJSONInt measures UnmarshalJSON with integer
func BenchmarkUnmarshalJSONInt(b *testing.B) {
	data := []byte(`2`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s status.Status
		_ = s.UnmarshalJSON(data)
	}
}

// BenchmarkUnmarshalJSONNull measures UnmarshalJSON with null
func BenchmarkUnmarshalJSONNull(b *testing.B) {
	data := []byte(`null`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s status.Status
		_ = s.UnmarshalJSON(data)
	}
}

// BenchmarkJSONMarshal measures standard json.Marshal
func BenchmarkJSONMarshal(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(s)
	}
}

// BenchmarkJSONUnmarshal measures standard json.Unmarshal
func BenchmarkJSONUnmarshal(b *testing.B) {
	data := []byte(`"OK"`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s status.Status
		_ = json.Unmarshal(data, &s)
	}
}

// BenchmarkRoundTripStringConversion measures full round-trip
func BenchmarkRoundTripStringConversion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := status.OK
		str := s.String()
		_ = status.NewFromString(str)
	}
}

// BenchmarkRoundTripIntConversion measures full round-trip via int
func BenchmarkRoundTripIntConversion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := status.OK
		val := s.Int64()
		_ = status.NewFromInt(val)
	}
}

// BenchmarkRoundTripJSON measures full JSON round-trip
func BenchmarkRoundTripJSON(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(s)
		var unmarshaled status.Status
		_ = json.Unmarshal(data, &unmarshaled)
	}
}

// BenchmarkComparison measures status comparison
func BenchmarkComparison(b *testing.B) {
	s1 := status.OK
	s2 := status.Warn
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s1 == s2
	}
}

// BenchmarkAllStatuses measures operations on all status values
func BenchmarkAllStatuses(b *testing.B) {
	statuses := []status.Status{status.KO, status.Warn, status.OK}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range statuses {
			_ = s.String()
			_ = s.Int()
			_ = s.Float()
		}
	}
}

// BenchmarkConcurrentString measures concurrent String() calls
func BenchmarkConcurrentString(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.String()
		}
	})
}

// BenchmarkConcurrentNewFromString measures concurrent constructor calls
func BenchmarkConcurrentNewFromString(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = status.NewFromString("OK")
		}
	})
}

// BenchmarkConcurrentMarshalJSON measures concurrent marshaling
func BenchmarkConcurrentMarshalJSON(b *testing.B) {
	s := status.OK
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.MarshalJSON()
		}
	})
}
