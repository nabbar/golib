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
	"fmt"
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

// BenchmarkName measures the performance of Name() method returning default name
func BenchmarkName(b *testing.B) {
	i, _ := info.New("benchmark-service")
	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Name()
	}
}

// BenchmarkNameWithFunction measures Name() with registered function (dynamic execution)
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

// BenchmarkSetName measures the performance of manually setting the name
func BenchmarkSetName(b *testing.B) {
	i, _ := info.New("benchmark-service")
	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.SetName("manual-name")
	}
}

// BenchmarkInfo measures the performance of Info() method returning empty map
func BenchmarkInfo(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Data()
	}
}

// BenchmarkInfoWithFunction measures Info() with registered function (dynamic execution)
func BenchmarkInfoWithFunction(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
			"count":   42,
		}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Data()
	}
}

// BenchmarkSetData measures performance of SetData with a small map
func BenchmarkSetData(b *testing.B) {
	i, _ := info.New("benchmark-service")
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.SetData(data)
	}
}

// BenchmarkAddData measures performance of adding a single key
func BenchmarkAddData(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.AddData("newKey", n)
	}
}

// BenchmarkDelData measures performance of deleting a single key
func BenchmarkDelData(b *testing.B) {
	i, _ := info.New("benchmark-service")
	// Pre-populate to ensure something is deleted (mostly) or just measure overhead
	i.AddData("key", "val")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		// We alternate add and del to actually measure deletion work?
		// If we only delete, subsequent deletes are no-ops.
		// Measuring no-op delete is valid too.
		i.DelData("key")
	}
}

// BenchmarkInfoLargeData measures performance with large info data
func BenchmarkInfoLargeData(b *testing.B) {
	i, _ := info.New("benchmark-service")

	largeData := make(map[string]interface{})
	for j := 0; j < 100; j++ {
		largeData[fmt.Sprintf("key-%d", j)] = j
	}

	i.RegisterData(func() (map[string]interface{}, error) {
		return largeData, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = i.Data()
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

// BenchmarkRegisterData measures RegisterData() performance
func BenchmarkRegisterData(b *testing.B) {
	i, _ := info.New("benchmark-service")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		i.RegisterData(func() (map[string]interface{}, error) {
			return map[string]interface{}{"key": "value"}, nil
		})
	}
}

// BenchmarkMarshalText measures text marshaling performance
func BenchmarkMarshalText(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = i.MarshalText()
	}
}

// BenchmarkMarshalJSON measures JSON marshaling performance
func BenchmarkMarshalJSON(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = i.MarshalJSON()
	}
}

// BenchmarkJSONMarshal measures standard json.Marshal performance
func BenchmarkJSONMarshal(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

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
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = i.Data()
		}
	})
}

// BenchmarkConcurrentMixedOperations measures mixed concurrent operations
func BenchmarkConcurrentMixedOperations(b *testing.B) {
	i, _ := info.New("benchmark-service")
	i.RegisterName(func() (string, error) {
		return "concurrent-name", nil
	})
	i.RegisterData(func() (map[string]interface{}, error) {
		return map[string]interface{}{"key": "value"}, nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = i.Name()
			_ = i.Data()
		}
	})
}
