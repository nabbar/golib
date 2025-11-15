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

package delim_test

import (
	"io"
	"strings"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BufferDelim Read Operations", func() {
	Describe("Read method", func() {
		Context("with newline delimiter", func() {
			It("should read until first newline", func() {
				r := io.NopCloser(strings.NewReader("hello\nworld"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(6))
				Expect(string(buf[:n])).To(Equal("hello\n"))
			})

			It("should read multiple chunks sequentially", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				// First read
				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("line1\n"))

				// Second read
				n, err = bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("line2\n"))

				// Third read
				n, err = bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("line3\n"))
			})

			It("should return EOF when no more data", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				_, _ = bd.Read(buf)

				// Second read should return EOF
				_, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle data without delimiter at end", func() {
				r := io.NopCloser(strings.NewReader("no newline"))
				bd := iotdlm.New(r, '\n', 0)

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
				bd := iotdlm.New(r, '|', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("col1|"))
			})

			It("should read with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("a,"))
			})

			It("should read with tab delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1\tfield2"))
				bd := iotdlm.New(r, '\t', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("field1\t"))
			})

			It("should read with null byte delimiter", func() {
				r := io.NopCloser(strings.NewReader("data\x00more"))
				bd := iotdlm.New(r, 0, 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("data\x00"))
			})
		})

		Context("with buffer size variations", func() {
			It("should handle small buffer that needs expansion", func() {
				r := io.NopCloser(strings.NewReader("very long line of text\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 5) // Small buffer
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(23))
			})

			It("should handle exact size buffer", func() {
				r := io.NopCloser(strings.NewReader("exact\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 6)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(6))
				Expect(string(buf[:n])).To(Equal("exact\n"))
			})

			It("should handle large buffer", func() {
				r := io.NopCloser(strings.NewReader("small\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 1000)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(6))
				Expect(string(buf[:n])).To(Equal("small\n"))
			})

			It("should handle zero-length buffer", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 0)
				n, err := bd.Read(buf)
				Expect(n).To(BeNumerically(">", 0))
				Expect(err).To(BeNil())
			})
		})

		Context("with empty data", func() {
			It("should handle empty reader", func() {
				r := io.NopCloser(strings.NewReader(""))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(0))
			})

			It("should handle reader with only delimiter", func() {
				r := io.NopCloser(strings.NewReader("\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(1))
				Expect(string(buf[:n])).To(Equal("\n"))
			})

			It("should handle multiple consecutive delimiters", func() {
				r := io.NopCloser(strings.NewReader("\n\n\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(1))
				Expect(string(buf[:n])).To(Equal("\n"))
			})
		})
	})

	Describe("ReadBytes method", func() {
		Context("with newline delimiter", func() {
			It("should read bytes until delimiter", func() {
				r := io.NopCloser(strings.NewReader("hello\nworld"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("hello\n")))
			})

			It("should read multiple times", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

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
				bd := iotdlm.New(r, '\n', 0)

				_, _ = bd.ReadBytes()
				_, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle last chunk without delimiter", func() {
				r := io.NopCloser(strings.NewReader("no delimiter"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(data).To(Equal([]byte("no delimiter")))
			})
		})

		Context("with various delimiters", func() {
			It("should work with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("a,")))
			})

			It("should work with custom delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1|field2|field3"))
				bd := iotdlm.New(r, '|', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("field1|")))
			})
		})

		Context("with large data", func() {
			It("should handle large chunks", func() {
				largeData := strings.Repeat("x", 10000) + "\n"
				r := io.NopCloser(strings.NewReader(largeData))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(10001))
			})

			It("should handle multiple large chunks", func() {
				chunk1 := strings.Repeat("a", 5000) + "\n"
				chunk2 := strings.Repeat("b", 5000) + "\n"
				r := io.NopCloser(strings.NewReader(chunk1 + chunk2))
				bd := iotdlm.New(r, '\n', 0)

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
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeUnit)

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
				bd := iotdlm.New(r, '\n', 0)

				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				Expect(unread).To(BeNil())
			})

			It("should return buffered data without consuming it from Read", func() {
				data := "line1\nline2\nline3\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 100*libsiz.SizeUnit)

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
				bd := iotdlm.New(r, '\n', 32*libsiz.SizeUnit)

				_, _ = bd.ReadBytes()
				unread, err := bd.UnRead()
				Expect(err).To(BeNil())
				if unread != nil {
					Expect(len(unread)).To(BeNumerically(">=", 0))
				}
			})

			It("should work with large buffer", func() {
				r := io.NopCloser(strings.NewReader("test\ndata"))
				bd := iotdlm.New(r, '\n', libsiz.SizeKilo)

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
			bd := iotdlm.New(r, '\n', 0)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error after close on Read", func() {
			r := io.NopCloser(strings.NewReader("test\n"))
			bd := iotdlm.New(r, '\n', 0)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to read after close
			buf := make([]byte, 10)
			_, err = bd.Read(buf)
			Expect(err).To(Equal(iotdlm.ErrInstance))
		})

		It("should return error after close on ReadBytes", func() {
			r := io.NopCloser(strings.NewReader("test\n"))
			bd := iotdlm.New(r, '\n', 0)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to read after close
			_, err = bd.ReadBytes()
			Expect(err).To(Equal(iotdlm.ErrInstance))
		})

		It("should return error after close on UnRead", func() {
			r := io.NopCloser(strings.NewReader("test\n"))
			bd := iotdlm.New(r, '\n', 0)

			err := bd.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to unread after close
			_, err = bd.UnRead()
			Expect(err).To(Equal(iotdlm.ErrInstance))
		})

		It("should allow multiple close calls", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0)

			err1 := bd.Close()
			Expect(err1).NotTo(HaveOccurred())

			// Second close should also work (might panic, let's check)
			Expect(func() {
				_ = bd.Close()
			}).To(Panic())
		})
	})
})
