/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package nopwritecloser_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nabbar/golib/ioutils/nopwritecloser"
)

// BenchmarkNew measures the cost of creating a new nopwritecloser wrapper.
// This benchmark establishes the baseline overhead for wrapper creation.
func BenchmarkNew(b *testing.B) {
	buf := &bytes.Buffer{}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = nopwritecloser.New(buf)
	}
}

// BenchmarkWrite_Small benchmarks writing small data (16 bytes).
// This represents typical short message writes.
func BenchmarkWrite_Small(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)
	data := []byte("small test data!")

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		wc.Write(data)
	}
}

// BenchmarkWrite_Medium benchmarks writing medium data (1KB).
// This represents typical packet or message sizes.
func BenchmarkWrite_Medium(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		wc.Write(data)
	}
}

// BenchmarkWrite_Large benchmarks writing large data (1MB).
// This represents file-sized chunks or large messages.
func BenchmarkWrite_Large(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		wc.Write(data)
	}
}

// BenchmarkWrite_Discard benchmarks writing to io.Discard.
// This isolates the wrapper overhead from buffer operations.
func BenchmarkWrite_Discard(b *testing.B) {
	wc := nopwritecloser.New(io.Discard)
	data := make([]byte, 1024)

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wc.Write(data)
	}
}

// BenchmarkClose measures the cost of calling Close().
// This should be near-zero as it's a no-op.
func BenchmarkClose(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wc.Close()
	}
}

// BenchmarkWriteClose benchmarks the combined write-close pattern.
// This represents a typical usage pattern with defer.
func BenchmarkWriteClose(b *testing.B) {
	data := []byte("test data")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		wc := nopwritecloser.New(buf)
		wc.Write(data)
		wc.Close()
	}
}

// BenchmarkMultipleWrites benchmarks multiple sequential writes.
// This represents streaming or chunked write patterns.
func BenchmarkMultipleWrites(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)
	data := []byte("chunk")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		for j := 0; j < 10; j++ {
			wc.Write(data)
		}
	}
}

// BenchmarkComparison_DirectWrite benchmarks direct buffer writes for comparison.
// This establishes the baseline performance without wrapper.
func BenchmarkComparison_DirectWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	data := []byte("test data")

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(data)
	}
}

// BenchmarkComparison_WrappedWrite benchmarks wrapped buffer writes.
// This shows the overhead introduced by the wrapper.
func BenchmarkComparison_WrappedWrite(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)
	data := []byte("test data")

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		wc.Write(data)
	}
}

// BenchmarkAllocation benchmarks memory allocation patterns.
// This measures the allocation overhead of the wrapper.
func BenchmarkAllocation(b *testing.B) {
	buf := &bytes.Buffer{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wc := nopwritecloser.New(buf)
		wc.Write([]byte("test"))
		wc.Close()
	}
}

// BenchmarkInterfaceConversion benchmarks type conversions.
// This measures the cost of interface assertions.
func BenchmarkInterfaceConversion(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var w io.Writer = wc
		var c io.Closer = wc
		var wc2 io.WriteCloser = wc
		_, _, _ = w, c, wc2
	}
}

// BenchmarkConcurrentWrites benchmarks concurrent write operations.
// This tests the wrapper's behavior under concurrent load.
func BenchmarkConcurrentWrites(b *testing.B) {
	buf := &safeBuffer{}
	wc := nopwritecloser.New(buf)
	data := []byte("concurrent data")

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wc.Write(data)
		}
	})
}

// BenchmarkConcurrentClose benchmarks concurrent close operations.
// This tests the safety of concurrent Close() calls.
func BenchmarkConcurrentClose(b *testing.B) {
	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wc.Close()
		}
	})
}

// BenchmarkStreamingPattern benchmarks a streaming write pattern.
// This simulates real-world streaming scenarios.
func BenchmarkStreamingPattern(b *testing.B) {
	chunkSize := 4096
	chunk := make([]byte, chunkSize)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		wc := nopwritecloser.New(buf)

		for j := 0; j < 100; j++ {
			wc.Write(chunk)
		}
		wc.Close()
	}
}

// BenchmarkVariableSizes benchmarks writes of varying sizes.
// This simulates variable-length message patterns.
func BenchmarkVariableSizes(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096}
	buffers := make([][]byte, len(sizes))
	for i, size := range sizes {
		buffers[i] = make([]byte, size)
	}

	buf := &bytes.Buffer{}
	wc := nopwritecloser.New(buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, data := range buffers {
			wc.Write(data)
		}
	}
}
