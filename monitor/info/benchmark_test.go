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

package info_test

import (
	"encoding/json"
	"testing"

	"github.com/nabbar/golib/monitor/info"
)

// BenchmarkNew measures the performance of creating new Info instances
func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = info.New("benchmark-service")
	}
}

// BenchmarkName measures the performance of Name() method
func BenchmarkName(b *testing.B) {
	i, _ := info.New("benchmark-service")
	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Name()
	}
}

// BenchmarkNameWithFunction measures Name() with registered function
func BenchmarkNameWithFunction(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterName(func() (string, error) {
		return "dynamic-name", nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Name()
	}
}

// BenchmarkNameCached measures cached Name() performance
func BenchmarkNameCached(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterName(func() (string, error) {
		return "cached-name", nil
	})

	// Prime the cache
	_ = i.Name()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Name()
	}
}

// BenchmarkInfo measures the performance of Info() method
func BenchmarkInfo(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Info()
	}
}

// BenchmarkInfoWithFunction measures Info() with registered function
func BenchmarkInfoWithFunction(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
			"count":   42,
		}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Info()
	}
}

// BenchmarkInfoCached measures cached Info() performance
func BenchmarkInfoCached(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	// Prime the cache
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Info()
	}
}

// BenchmarkInfoLargeData measures performance with large info data
func BenchmarkInfoLargeData(b *testing.B) {
	i, _ := info.New("benchmark-service")

	largeData := make(map[string]interface{})
	for j := 0; j < 100; j++ {
		largeData[string(rune('a'+j%26))+string(rune('0'+j%10))] = j
	}

	i.RegisterInfo(func() (map[string]interface{}, error) {
		return largeData, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Info()
	}
}

// BenchmarkRegisterName measures RegisterName() performance
func BenchmarkRegisterName(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.RegisterName(func() (string, error) {
			return "new-name", nil
		})
	}
}

// BenchmarkRegisterInfo measures RegisterInfo() performance
func BenchmarkRegisterInfo(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.RegisterInfo(func() (map[string]interface{}, error) {
			return map[string]interface{}{"key": "value"}, nil
		})
	}
}

// BenchmarkMarshalText measures text marshaling performance
func BenchmarkMarshalText(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	// Prime the cache
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = i.MarshalText()
	}
}

// BenchmarkMarshalJSON measures JSON marshaling performance
func BenchmarkMarshalJSON(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	// Prime the cache
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = i.MarshalJSON()
	}
}

// BenchmarkJSONMarshal measures standard json.Marshal performance
func BenchmarkJSONMarshal(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	// Prime the cache
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = json.Marshal(i)
	}
}

// BenchmarkConcurrentNameReads measures concurrent Name() reads
func BenchmarkConcurrentNameReads(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterName(func() (string, error) {
		return "concurrent-name", nil
	})

	// Prime the cache
	_ = i.Name()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = i.Name()
		}
	})
}

// BenchmarkConcurrentInfoReads measures concurrent Info() reads
func BenchmarkConcurrentInfoReads(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	// Prime the cache
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = i.Info()
		}
	})
}

// BenchmarkConcurrentMixedOperations measures mixed concurrent operations
func BenchmarkConcurrentMixedOperations(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterName(func() (string, error) {
		return "concurrent-name", nil
	})
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{"key": "value"}, nil
	})

	// Prime the cache
	_ = i.Name()
	_ = i.Info()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = i.Name()
			_ = i.Info()
		}
	})
}
