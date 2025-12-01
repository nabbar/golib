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
	"bytes"
	"io"
	"strings"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file validates all write and copy operations of BufferDelim.
// It covers:
//   - WriteTo() method for efficient data copying
//   - Copy() method (alias for WriteTo)
//   - Error handling during write operations
//   - Large data transfers with various buffer sizes
//   - Integration with different io.Writer implementations
//   - Delimiter-aware writing behavior
//
// Helper types like errorWriter (defined in helper_test.go) are used to
// simulate error conditions and validate proper error propagation.

var _ = Describe("BufferDelim Write Operations", func() {
	Describe("WriteTo method", func() {
		Context("with basic data", func() {
			It("should write all data until EOF", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(18)))
				Expect(buf.String()).To(Equal("line1\nline2\nline3\n"))
			})

			It("should write empty data", func() {
				r := io.NopCloser(strings.NewReader(""))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(0)))
				Expect(buf.String()).To(Equal(""))
			})

			It("should write single line with delimiter", func() {
				r := io.NopCloser(strings.NewReader("single\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(7)))
				Expect(buf.String()).To(Equal("single\n"))
			})

			It("should write single line without delimiter", func() {
				r := io.NopCloser(strings.NewReader("single"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(6)))
				Expect(buf.String()).To(Equal("single"))
			})
		})

		Context("with various delimiters", func() {
			It("should write with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(5)))
				Expect(buf.String()).To(Equal("a,b,c"))
			})

			It("should write with pipe delimiter", func() {
				r := io.NopCloser(strings.NewReader("col1|col2|col3"))
				bd := iotdlm.New(r, '|', 0)

				buf := &bytes.Buffer{}
				_, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("col1|col2|col3"))
			})

			It("should write with tab delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1\tfield2\tfield3"))
				bd := iotdlm.New(r, '\t', 0)

				buf := &bytes.Buffer{}
				_, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("field1\tfield2\tfield3"))
			})

			It("should write with null byte delimiter", func() {
				r := io.NopCloser(strings.NewReader("data\x00more\x00"))
				bd := iotdlm.New(r, 0, 0)

				buf := &bytes.Buffer{}
				_, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("data\x00more\x00"))
			})
		})

		Context("with large data", func() {
			It("should write large content", func() {
				largeData := strings.Repeat("x", 100000) + "\n"
				r := io.NopCloser(strings.NewReader(largeData))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(100001)))
				Expect(buf.Len()).To(Equal(100001))
			})

			It("should write multiple large chunks", func() {
				chunk1 := strings.Repeat("a", 50000) + "\n"
				chunk2 := strings.Repeat("b", 50000) + "\n"
				chunk3 := strings.Repeat("c", 50000) + "\n"
				r := io.NopCloser(strings.NewReader(chunk1 + chunk2 + chunk3))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(150003)))
			})

			It("should write with small internal buffer", func() {
				data := strings.Repeat("test\n", 1000)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeUnit)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(len(data))))
				Expect(buf.String()).To(Equal(data))
			})

			It("should write with large internal buffer", func() {
				data := strings.Repeat("test\n", 1000)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeKilo)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(len(data))))
				Expect(buf.String()).To(Equal(data))
			})
		})

		Context("with writer errors", func() {
			It("should handle write error", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				// Writer that fails after 2 writes
				ew := newErrorWriter(2)
				_, err := bd.WriteTo(ew)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("write error"))
			})

			It("should return proper byte count before error", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				ew := newErrorWriter(1)
				n, err := bd.WriteTo(ew)
				Expect(err).To(HaveOccurred())
				// Should have written at least the first line
				Expect(n).To(BeNumerically(">=", 6))
			})
		})

		Context("with consecutive delimiters", func() {
			It("should write multiple consecutive newlines", func() {
				r := io.NopCloser(strings.NewReader("\n\n\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(3)))
				Expect(buf.String()).To(Equal("\n\n\n"))
			})

			It("should write data with empty chunks", func() {
				r := io.NopCloser(strings.NewReader("a\n\nb\n\nc"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				_, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("a\n\nb\n\nc"))
			})
		})

		Context("after close", func() {
			It("should return error when closed", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 0)

				err := bd.Close()
				Expect(err).NotTo(HaveOccurred())

				buf := &bytes.Buffer{}
				_, err = bd.WriteTo(buf)
				Expect(err).To(Equal(iotdlm.ErrInstance))
			})
		})
	})

	Describe("Copy method", func() {
		Context("with basic data", func() {
			It("should copy all data", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.Copy(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(12)))
				Expect(buf.String()).To(Equal("line1\nline2\n"))
			})

			It("should copy empty data", func() {
				r := io.NopCloser(strings.NewReader(""))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.Copy(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(0)))
			})

			It("should behave identically to WriteTo", func() {
				data := "test1\ntest2\ntest3\n"

				r1 := io.NopCloser(strings.NewReader(data))
				bd1 := iotdlm.New(r1, '\n', 0)
				buf1 := &bytes.Buffer{}
				n1, err1 := bd1.WriteTo(buf1)

				r2 := io.NopCloser(strings.NewReader(data))
				bd2 := iotdlm.New(r2, '\n', 0)
				buf2 := &bytes.Buffer{}
				n2, err2 := bd2.Copy(buf2)

				Expect(n1).To(Equal(n2))
				Expect(err1).To(Equal(err2))
				Expect(buf1.String()).To(Equal(buf2.String()))
			})
		})

		Context("with various delimiters", func() {
			It("should copy with custom delimiter", func() {
				r := io.NopCloser(strings.NewReader("a|b|c"))
				bd := iotdlm.New(r, '|', 0)

				buf := &bytes.Buffer{}
				_, err := bd.Copy(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("a|b|c"))
			})
		})

		Context("with large data", func() {
			It("should copy large content efficiently", func() {
				largeData := strings.Repeat("line\n", 10000)
				r := io.NopCloser(strings.NewReader(largeData))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				n, err := bd.Copy(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(n).To(Equal(int64(len(largeData))))
				Expect(buf.Len()).To(Equal(len(largeData)))
			})
		})

		Context("with writer errors", func() {
			It("should handle copy errors", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				ew := newErrorWriter(1)
				_, err := bd.Copy(ew)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Integration between Read and Write operations", func() {
		It("should not interfere when mixing Read and WriteTo", func() {
			r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
			bd := iotdlm.New(r, '\n', 0)

			// First read one line
			buf := make([]byte, 10)
			n, err := bd.Read(buf)
			Expect(err).To(BeNil())
			Expect(string(buf[:n])).To(Equal("line1\n"))

			// Then write the rest
			writeBuf := &bytes.Buffer{}
			wn, werr := bd.WriteTo(writeBuf)
			Expect(werr).To(Equal(io.EOF))
			Expect(wn).To(BeNumerically(">", 0))
			Expect(writeBuf.String()).To(Equal("line2\nline3\n"))
		})

		It("should handle ReadBytes followed by WriteTo", func() {
			r := io.NopCloser(strings.NewReader("first\nsecond\nthird\n"))
			bd := iotdlm.New(r, '\n', 0)

			// Read first chunk
			data, err := bd.ReadBytes()
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte("first\n")))

			// Write the rest
			buf := &bytes.Buffer{}
			_, err = bd.WriteTo(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(buf.String()).To(Equal("second\nthird\n"))
		})
	})
})
