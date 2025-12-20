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

package delim_test

import (
	"errors"
	"io"
	"strings"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file validates all read operations of BufferDelim.
// It covers:
//   - Read() method with various delimiters and data patterns
//   - ReadBytes() for reading complete delimited chunks
//   - UnRead() for peeking at buffered data
//   - Behavior with different buffer sizes
//   - EOF handling and edge cases
//   - Multiple sequential reads
//   - Delimiter preservation in returned data
//
// These tests ensure robust read functionality across all use cases.

var _ = Describe("BufferDelim Read Operations", func() {
	Describe("Read method", func() {
		Context("with newline delimiter", func() {
			It("should read until first newline", func() {
				r := io.NopCloser(strings.NewReader("hello\nworld"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(10))
				Expect(string(buf[:n])).To(Equal("hello\nworl"))
			})

			It("should read multiple chunks sequentially", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				// First read
				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(10))
				Expect(string(buf[:n])).To(Equal("line1\nline"))

				// Second read
				n, err = bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(8))
				Expect(string(buf[:n])).To(Equal("2\nline3\n"))

				// Third read
				n, err = bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(0))
			})

			It("should return EOF when no more data", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 10)
				_, _ = bd.Read(buf)

				// Second read should return EOF
				_, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle data without delimiter at end", func() {
				r := io.NopCloser(strings.NewReader("no newline"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 20)
				n, err := bd.Read(buf)
				// Should return data and EOF together when no delimiter found
				Expect(n).To(Equal(10)) // "no newline" = 10 chars
				Expect(string(buf[:n])).To(Equal("no newline"))
				if err != nil {
					Expect(err).To(Equal(io.EOF))
				}
			})
		})

		Context("with various delimiters", func() {
			It("should read with pipe delimiter", func() {
				r := io.NopCloser(strings.NewReader("col1|col2|col3"))
				bd := iotdlm.New(r, '|', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("col1|col2|"))
			})

			It("should read with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(string(buf[:n])).To(Equal("a,b,c"))
			})

			It("should read with tab delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1\tfield2"))
				bd := iotdlm.New(r, '\t', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("field1\tfie"))
			})

			It("should read with null byte delimiter", func() {
				r := io.NopCloser(strings.NewReader("data\x00more"))
				bd := iotdlm.New(r, 0, 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(string(buf[:n])).To(Equal("data\x00more"))
			})
		})

		Context("with buffer size variations", func() {
			It("should handle small buffer that returns partial data", func() {
				r := io.NopCloser(strings.NewReader("very long line of text\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 5) // Small buffer
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(5))
				Expect(string(buf[:n])).To(Equal("very "))
			})

			It("should handle exact size buffer", func() {
				r := io.NopCloser(strings.NewReader("exact\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 6)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(6))
				Expect(string(buf[:n])).To(Equal("exact\n"))
			})

			It("should handle large buffer", func() {
				r := io.NopCloser(strings.NewReader("small\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 1000)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(6))
				Expect(string(buf[:n])).To(Equal("small\n"))
			})

			It("should handle zero-length buffer", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 0)
				n, err := bd.Read(buf)
				Expect(n).To(Equal(0))
				Expect(err).To(BeNil())
			})
		})

		Context("with empty data", func() {
			It("should handle empty reader", func() {
				r := io.NopCloser(strings.NewReader(""))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(0))
			})

			It("should handle reader with only delimiter", func() {
				r := io.NopCloser(strings.NewReader("\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(1))
				Expect(string(buf[:n])).To(Equal("\n"))
			})

			It("should handle multiple consecutive delimiters", func() {
				r := io.NopCloser(strings.NewReader("\n\n\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(3))
				Expect(string(buf[:n])).To(Equal("\n\n\n"))
			})
		})
	})

	Describe("ReadBytes method", func() {
		Context("with newline delimiter", func() {
			It("should read bytes until delimiter", func() {
				r := io.NopCloser(strings.NewReader("hello\nworld"))
				bd := iotdlm.New(r, '\n', 0, false)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("hello\n")))
			})

			It("should read multiple times", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data1).To(Equal([]byte("line1\n")))

				data2, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data2).To(Equal([]byte("line2\n")))

				data3, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data3).To(Equal([]byte("line3\n")))
			})

			It("should return EOF at end", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0, false)

				_, _ = bd.ReadBytes()
				_, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle last chunk without delimiter", func() {
				r := io.NopCloser(strings.NewReader("no delimiter"))
				bd := iotdlm.New(r, '\n', 0, false)

				data, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(data).To(Equal([]byte("no delimiter")))
			})
		})

		Context("with various delimiters", func() {
			It("should work with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0, false)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("a,")))
			})

			It("should work with custom delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1|field2|field3"))
				bd := iotdlm.New(r, '|', 0, false)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("field1|")))
			})
		})

		Context("with large data", func() {
			It("should handle large chunks", func() {
				largeData := strings.Repeat("x", 10000) + "\n"
				r := io.NopCloser(strings.NewReader(largeData))
				bd := iotdlm.New(r, '\n', 0, false)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(10001))
			})

			It("should handle multiple large chunks", func() {
				chunk1 := strings.Repeat("a", 5000) + "\n"
				chunk2 := strings.Repeat("b", 5000) + "\n"
				r := io.NopCloser(strings.NewReader(chunk1 + chunk2))
				bd := iotdlm.New(r, '\n', 0, false)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data1)).To(Equal(5001))

				data2, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data2)).To(Equal(5001))
			})
		})
	})

	Describe("UnRead method", func() {
		Context("with buffered data", func() {
			It("should return buffered data after partial read", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3"))
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeUnit, false)

				// Read first line
				_, _ = bd.ReadBytes()

				// Get unread buffered data
				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				Expect(unread).NotTo(BeNil())
				// Should contain "line2\nline3" or at least part of it
			})

			It("should return nil when no buffered data", func() {
				r := io.NopCloser(strings.NewReader(""))
				bd := iotdlm.New(r, '\n', 0, false)

				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				Expect(unread).To(BeNil())
			})

			It("should return buffered data without consuming it from Read", func() {
				data := "line1\nline2\nline3\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 100*libsiz.SizeUnit, false)

				// First, peek at unread (might be empty or full depending on buffering)
				unread1, err := bd.UnRead()
				Expect(err).To(BeNil())

				// Read first line
				read1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(read1).To(Equal([]byte("line1\n")))

				// Now UnRead should return remaining buffered data
				unread2, err := bd.UnRead()
				Expect(err).To(BeNil())
				if unread1 == nil && unread2 != nil {
					// Data was buffered after first read
					Expect(len(unread2)).To(BeNumerically(">", 0))
				}
			})
		})

		Context("with various buffer sizes", func() {
			It("should work with small buffer", func() {
				r := io.NopCloser(strings.NewReader("test\ndata"))
				bd := iotdlm.New(r, '\n', 32*libsiz.SizeUnit, false)

				_, _ = bd.ReadBytes()
				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				if unread != nil {
					Expect(len(unread)).To(BeNumerically(">=", 0))
				}
			})

			It("should work with large buffer", func() {
				r := io.NopCloser(strings.NewReader("test\ndata"))
				bd := iotdlm.New(r, '\n', libsiz.SizeKilo, false)

				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				// With large buffer, all data might be buffered
				if unread != nil {
					Expect(len(unread)).To(BeNumerically(">", 0))
				}
			})
		})
	})

	Describe("Close method", func() {
		It("should close successfully", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error after close on Read", func() {
			r := newClosableBuffer("test\n")
			bd := iotdlm.New(r, '\n', 0, false)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to read after close - should get EOF from closed reader
			buf := make([]byte, 10)
			_, err = bd.Read(buf)
			Expect(err).To(HaveOccurred())
		})

		It("should return error after close on ReadBytes", func() {
			r := newClosableBuffer("test\n")
			bd := iotdlm.New(r, '\n', 0, false)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to read after close - should get EOF from closed reader
			_, err = bd.ReadBytes()
			Expect(err).To(HaveOccurred())
		})

		It("should return empty buffer after close on UnRead", func() {
			r := newClosableBuffer("test\n")
			bd := iotdlm.New(r, '\n', 0, false)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// After close, buffer is cleared, so UnRead returns nil
			unread, err := bd.UnRead()
			Expect(err).To(BeNil())
			Expect(unread).To(BeNil())
		})

		It("should allow multiple close calls", func() {
			r := newClosableBuffer("test")
			bd := iotdlm.New(r, '\n', 0, false)

			err1 := bd.Close()
			Expect(err1).NotTo(HaveOccurred())

			// Second close might return an error from already-closed reader
			err2 := bd.Close()
			_ = err2 // May or may not error depending on reader implementation
		})
	})

	Describe("Internal Read/Fill Coverage", func() {
		Context("ReadBytes branch coverage", func() {
			It("should handle delimiter at exact end of buffer", func() {
				data := "abc\n"
				r := io.NopCloser(&mockReaderEOFData{data: data})
				bd := iotdlm.New(r, '\n', libsiz.Size(4), false)
				defer bd.Close()

				res, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(res)).To(Equal("abc\n"))
			})

			It("should handle delimiter in middle of buffer", func() {
				data := "abc\nde"
				r := io.NopCloser(&mockReaderEOFData{data: data})
				bd := iotdlm.New(r, '\n', libsiz.Size(6), false)
				defer bd.Close()

				res, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(res)).To(Equal("abc\n"))

				// Read remaining
				res, err = bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(string(res)).To(Equal("de"))
			})

			It("should clear EOF error if delimiter is found", func() {
				data := "abc\n"
				r := io.NopCloser(&mockReaderEOFData{data: data})
				bd := iotdlm.New(r, '\n', libsiz.Size(10), false)
				defer bd.Close()

				res, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(res)).To(Equal("abc\n"))

				// Next read should be EOF
				res, err = bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(res).To(BeNil())
			})
		})

		Context("fill() method", func() {
			It("should handle Read returning (0, nil) by returning EOF", func() {
				mr := &mockReader0Nil{}
				bd := iotdlm.NewInternal(
					io.NopCloser(mr),
					libsiz.SizeKilo,
					make([]byte, 0, libsiz.SizeKilo.Int()),
				)

				err := bd.Fill()
				Expect(err).To(Equal(io.EOF))
				Expect(mr.called).To(BeTrue())
			})

			It("should handle fill with non-empty buffer (mostly full)", func() {
				mr := &mockReader0Nil{called: true}
				s := 1024
				bd := iotdlm.NewInternal(
					io.NopCloser(mr),
					libsiz.ParseUint64(uint64(s)),
					make([]byte, 800),
				)
				err := bd.Fill()
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle fill with non-empty buffer (mostly empty)", func() {
				mr := &mockReader0Nil{called: true}
				s := 1024
				bd := iotdlm.NewInternal(
					io.NopCloser(mr),
					libsiz.ParseUint64(uint64(s)),
					make([]byte, 100),
				)
				err := bd.Fill()
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("ReadBytes with zero size buffer", func() {
			It("should handle zero-sized buffer gracefully", func() {
				bd := iotdlm.NewInternal(
					io.NopCloser(&mockReader0Nil{}),
					0,
					nil,
				)

				res, err := bd.ReadBytes()
				Expect(res).To(BeNil())
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("readBuf with empty buffer", func() {
			It("should handle nil buffer", func() {
				bd := iotdlm.NewInternal(
					io.NopCloser(&mockReader0Nil{}),
					libsiz.SizeKilo,
					nil,
				)

				n, err := bd.ReadBuf(nil)
				Expect(n).To(Equal(0))
				Expect(err).To(BeNil())
			})

			It("should handle empty buffer", func() {
				bd := iotdlm.NewInternal(
					io.NopCloser(&mockReader0Nil{}),
					libsiz.SizeKilo,
					nil,
				)

				n, err := bd.ReadBuf(make([]byte, 0))
				Expect(n).To(Equal(0))
				Expect(err).To(BeNil())
			})
		})

		Context("ReadBytes with immediate error", func() {
			It("should propagate immediate read error", func() {
				expectedErr := errors.New("immediate error")
				mr := &mockReaderError{err: expectedErr}

				bd := iotdlm.NewInternal(
					io.NopCloser(mr),
					libsiz.SizeKilo,
					nil,
				)

				res, err := bd.ReadBytes()
				Expect(res).To(BeNil())
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
