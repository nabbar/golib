/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package bufferReadCloser_test

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/nabbar/golib/ioutils/bufferReadCloser"
)

// Benchmark tests verify that the wrappers add minimal overhead compared to
// the underlying stdlib types. These benchmarks measure:
//   - Read/Write operations performance
//   - Close operation overhead
//   - Memory allocations
//   - Comparison with stdlib baseline

// BenchmarkBufferRead measures read performance compared to bytes.Buffer.
func BenchmarkBufferRead(b *testing.B) {
	data := generateTestData(1024)
	readBuf := make([]byte, 128)
	
	b.Run("stdlib_bytes.Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(data)
			for {
				_, err := buf.Read(readBuf)
				if err == io.EOF {
					break
				}
			}
		}
	})
	
	b.Run("wrapped_Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(data)
			wrapped := bufferReadCloser.NewBuffer(buf, nil)
			for {
				_, err := wrapped.Read(readBuf)
				if err == io.EOF {
					break
				}
			}
			wrapped.Close()
		}
	})
}

// BenchmarkBufferWrite measures write performance compared to bytes.Buffer.
func BenchmarkBufferWrite(b *testing.B) {
	data := generateTestData(1024)
	
	b.Run("stdlib_bytes.Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, 2048))
			buf.Write(data)
		}
	})
	
	b.Run("wrapped_Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, 2048))
			wrapped := bufferReadCloser.NewBuffer(buf, nil)
			wrapped.Write(data)
			wrapped.Close()
		}
	})
}

// BenchmarkBufferClose measures the overhead of Close operation.
func BenchmarkBufferClose(b *testing.B) {
	b.Run("without_custom_close", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			wrapped := bufferReadCloser.NewBuffer(buf, nil)
			wrapped.Close()
		}
	})
	
	b.Run("with_custom_close", func(b *testing.B) {
		closeFn := func() error { return nil }
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			wrapped := bufferReadCloser.NewBuffer(buf, closeFn)
			wrapped.Close()
		}
	})
}

// BenchmarkReaderRead measures read performance compared to bufio.Reader.
func BenchmarkReaderRead(b *testing.B) {
	data := generateTestData(1024)
	readBuf := make([]byte, 128)
	
	b.Run("stdlib_bufio.Reader", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			source := bytes.NewReader(data)
			reader := bufio.NewReader(source)
			for {
				_, err := reader.Read(readBuf)
				if err == io.EOF {
					break
				}
			}
		}
	})
	
	b.Run("wrapped_Reader", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			source := bytes.NewReader(data)
			br := bufio.NewReader(source)
			wrapped := bufferReadCloser.NewReader(br, nil)
			for {
				_, err := wrapped.Read(readBuf)
				if err == io.EOF {
					break
				}
			}
			wrapped.Close()
		}
	})
}

// BenchmarkWriterWrite measures write performance compared to bufio.Writer.
func BenchmarkWriterWrite(b *testing.B) {
	data := generateTestData(1024)
	
	b.Run("stdlib_bufio.Writer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dest := &bytes.Buffer{}
			writer := bufio.NewWriter(dest)
			writer.Write(data)
			writer.Flush()
		}
	})
	
	b.Run("wrapped_Writer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			wrapped := bufferReadCloser.NewWriter(bw, nil)
			wrapped.Write(data)
			wrapped.Close()
		}
	})
}

// BenchmarkReadWriterBidirectional measures bidirectional I/O performance.
func BenchmarkReadWriterBidirectional(b *testing.B) {
	data := generateTestData(512)
	readBuf := make([]byte, 128)
	
	b.Run("stdlib_bufio.ReadWriter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(data)
			rw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			rw.Read(readBuf)
			rw.WriteString("response")
			rw.Flush()
		}
	})
	
	b.Run("wrapped_ReadWriter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(data)
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			wrapped := bufferReadCloser.NewReadWriter(brw, nil)
			wrapped.Read(readBuf)
			wrapped.WriteString("response")
			wrapped.Close()
		}
	})
}

// BenchmarkAllocation measures memory allocation overhead.
func BenchmarkAllocation(b *testing.B) {
	b.Run("NewBuffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			_ = bufferReadCloser.NewBuffer(buf, nil)
		}
	})
	
	b.Run("NewReader", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			source := bytes.NewReader([]byte("test"))
			br := bufio.NewReader(source)
			_ = bufferReadCloser.NewReader(br, nil)
		}
	})
	
	b.Run("NewWriter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dest := &bytes.Buffer{}
			bw := bufio.NewWriter(dest)
			_ = bufferReadCloser.NewWriter(bw, nil)
		}
	})
	
	b.Run("NewReadWriter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			brw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
			_ = bufferReadCloser.NewReadWriter(brw, nil)
		}
	})
}

// BenchmarkLargeData measures performance with large data transfers.
func BenchmarkLargeData(b *testing.B) {
	sizes := []int{1024, 10240, 102400, 1024000} // 1KB, 10KB, 100KB, 1MB
	
	for _, size := range sizes {
		data := generateTestData(size)
		
		b.Run("Buffer_"+byteSizeString(size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				buf := bytes.NewBuffer(make([]byte, 0, size*2))
				wrapped := bufferReadCloser.NewBuffer(buf, nil)
				wrapped.Write(data)
				wrapped.Close()
			}
		})
	}
}

// byteSizeString converts byte size to human-readable string.
func byteSizeString(size int) string {
	switch {
	case size < 1024:
		return "1KB"
	case size < 10240:
		return "10KB"
	case size < 102400:
		return "100KB"
	default:
		return "1MB"
	}
}

// BenchmarkNilHandling measures the overhead of nil parameter handling.
func BenchmarkNilHandling(b *testing.B) {
	b.Run("Buffer_nil", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			wrapped := bufferReadCloser.NewBuffer(nil, nil)
			wrapped.WriteString("test")
			wrapped.Close()
		}
	})
	
	b.Run("Reader_nil", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			wrapped := bufferReadCloser.NewReader(nil, nil)
			buf := make([]byte, 10)
			wrapped.Read(buf)
			wrapped.Close()
		}
	})
	
	b.Run("Writer_nil", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			wrapped := bufferReadCloser.NewWriter(nil, nil)
			wrapped.WriteString("test")
			wrapped.Close()
		}
	})
}
