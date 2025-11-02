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

package ioprogress_test

import (
	"bytes"
	"strings"
	"sync/atomic"

	. "github.com/nabbar/golib/ioutils/ioprogress"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type closeableWriter struct {
	*bytes.Buffer
	closed bool
}

func (c *closeableWriter) Close() error {
	c.closed = true
	return nil
}

func newCloseableWriter() *closeableWriter {
	return &closeableWriter{
		Buffer: &bytes.Buffer{},
		closed: false,
	}
}

var _ = Describe("Writer", func() {
	Context("Creation", func() {
		It("should create writer from io.WriteCloser", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			Expect(writer).ToNot(BeNil())
		})
	})

	Context("Write operations", func() {
		It("should write data", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			n, err := writer.Write([]byte("hello world"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(11))
			Expect(dest.String()).To(Equal("hello world"))
		})

		It("should write multiple times", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			n1, err1 := writer.Write([]byte("hello"))
			n2, err2 := writer.Write([]byte(" "))
			n3, err3 := writer.Write([]byte("world"))

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
			Expect(err3).ToNot(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(n2).To(Equal(1))
			Expect(n3).To(Equal(5))
			Expect(dest.String()).To(Equal("hello world"))
		})

		It("should handle empty write", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			n, err := writer.Write([]byte{})

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(dest.Len()).To(Equal(0))
		})
	})

	Context("Progress tracking with increment callback", func() {
		It("should call increment callback on each write", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var totalIncremented int64
			var callCount atomic.Int32

			writer.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&totalIncremented, size)
				callCount.Add(1)
			})

			// Write in chunks
			writer.Write([]byte("hello")) // 5 bytes
			writer.Write([]byte(" "))     // 1 byte
			writer.Write([]byte("world")) // 5 bytes

			Expect(totalIncremented).To(Equal(int64(11)))
			Expect(callCount.Load()).To(Equal(int32(3)))
		})

		It("should track total bytes written", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var totalBytes int64
			writer.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&totalBytes, size)
			})

			// Write all at once
			data := []byte("1234567890")
			n, _ := writer.Write(data)

			Expect(totalBytes).To(Equal(int64(n)))
			Expect(totalBytes).To(Equal(int64(10)))
		})

		It("should handle nil increment callback", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			// Register nil callback (should not panic)
			writer.RegisterFctIncrement(nil)

			n, err := writer.Write([]byte("test"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should allow changing increment callback", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var count1 int64
			writer.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&count1, size)
			})

			writer.Write([]byte("abcd")) // 4 bytes with first callback

			var count2 int64
			writer.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&count2, size)
			})

			writer.Write([]byte("efgh")) // 4 bytes with second callback

			Expect(count1).To(Equal(int64(4)))
			Expect(count2).To(Equal(int64(4)))
		})

		It("should call increment even for zero-byte writes", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var incrementCalled bool
			writer.RegisterFctIncrement(func(size int64) {
				incrementCalled = true
				Expect(size).To(Equal(int64(0)))
			})

			writer.Write([]byte{})

			Expect(incrementCalled).To(BeTrue())
		})
	})

	Context("EOF callback", func() {
		It("should handle nil EOF callback", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			// Register nil callback (should not panic)
			writer.RegisterFctEOF(nil)

			n, err := writer.Write([]byte("test"))

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should not panic when EOF callback is set", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			eofCalled := false
			writer.RegisterFctEOF(func() {
				eofCalled = true
			})

			writer.Write([]byte("test"))

			// EOF is typically not triggered on writes
			// This tests that the callback can be registered without issues
			Expect(eofCalled).To(BeFalse())
		})
	})

	Context("Reset callback", func() {
		It("should call reset callback with max and current", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var resetMax, resetCurrent int64
			writer.RegisterFctReset(func(max, current int64) {
				resetMax = max
				resetCurrent = current
			})

			// Write some data first
			writer.Write([]byte("abc"))

			// Call reset
			writer.Reset(100)

			Expect(resetMax).To(Equal(int64(100)))
			Expect(resetCurrent).To(Equal(int64(3)))
		})

		It("should handle nil reset callback", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			// Register nil callback (should not panic)
			writer.RegisterFctReset(nil)

			// This should not panic
			writer.Reset(50)
		})

		It("should track current progress correctly", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var progressValues []int64
			writer.RegisterFctReset(func(max, current int64) {
				progressValues = append(progressValues, current)
			})

			// Write incrementally and reset
			writer.Write([]byte("123")) // 3 bytes
			writer.Reset(100)

			writer.Write([]byte("456")) // 6 bytes total
			writer.Reset(100)

			writer.Write([]byte("789")) // 9 bytes total
			writer.Reset(100)

			Expect(progressValues).To(Equal([]int64{3, 6, 9}))
		})

		It("should reset without prior registration", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			writer.Write([]byte("data"))

			// Reset without registering callback (should not panic)
			Expect(func() {
				writer.Reset(100)
			}).ToNot(Panic())
		})
	})

	Context("Close operations", func() {
		It("should close underlying writer", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			err := writer.Close()

			Expect(err).ToNot(HaveOccurred())
			Expect(dest.closed).To(BeTrue())
		})

		It("should be safe to close multiple times", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			err1 := writer.Close()
			err2 := writer.Close()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})

	Context("Combined operations", func() {
		It("should track progress through complete write cycle", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var increments []int64

			writer.RegisterFctIncrement(func(size int64) {
				increments = append(increments, size)
			})

			// Write in chunks
			writer.Write([]byte("hello"))
			writer.Write([]byte(" "))
			writer.Write([]byte("world"))
			writer.Write([]byte(" test"))

			Expect(increments).To(Equal([]int64{5, 1, 5, 5}))
			Expect(dest.String()).To(Equal("hello world test"))
		})

		It("should maintain count across multiple operations", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var counts []int64
			writer.RegisterFctReset(func(max, current int64) {
				counts = append(counts, current)
			})

			writer.Write([]byte("a"))
			writer.Reset(10)
			writer.Write([]byte("b"))
			writer.Reset(10)
			writer.Write([]byte("c"))
			writer.Reset(10)

			Expect(counts).To(Equal([]int64{1, 2, 3}))
		})
	})

	Context("Edge cases", func() {
		It("should handle large data", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var totalBytes int64
			writer.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&totalBytes, size)
			})

			// Write 1 MB
			largeData := []byte(strings.Repeat("x", 1024*1024))
			n, err := writer.Write(largeData)

			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
			Expect(totalBytes).To(Equal(int64(n)))
		})

		It("should handle many small writes", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			var writeCount atomic.Int32
			writer.RegisterFctIncrement(func(size int64) {
				writeCount.Add(1)
			})

			// Many small writes
			for i := 0; i < 1000; i++ {
				writer.Write([]byte("x"))
			}

			Expect(writeCount.Load()).To(Equal(int32(1000)))
			Expect(dest.Len()).To(Equal(1000))
		})
	})

	Context("Concurrent safety", func() {
		It("should handle concurrent callback registration", func() {
			dest := newCloseableWriter()
			writer := NewWriteCloser(dest)

			done := make(chan bool)

			// Register callbacks concurrently
			go func() {
				writer.RegisterFctIncrement(func(size int64) {})
				done <- true
			}()

			go func() {
				writer.RegisterFctReset(func(max, current int64) {})
				done <- true
			}()

			<-done
			<-done

			// Should not panic
			writer.Write([]byte("test"))
		})
	})
})
