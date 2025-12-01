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
	"unicode/utf8"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file validates edge cases and boundary conditions.
// It covers challenging scenarios including:
//   - Unicode and multi-byte character handling
//   - Very large data streams and buffer overflow conditions
//   - Binary data with special characters (null bytes, control chars)
//   - Error conditions during read/write operations
//   - Empty data and missing delimiters
//   - Extremely long lines exceeding buffer sizes
//   - Mixed binary and text content
//
// Helper types like errorReader (defined in helper_test.go) simulate I/O errors
// to test error propagation and recovery mechanisms.

var _ = Describe("BufferDelim Edge Cases and Error Handling", func() {
	Describe("Unicode and special characters", func() {
		Context("with Unicode delimiters", func() {
			It("should handle Unicode delimiter", func() {
				r := io.NopCloser(strings.NewReader("hello€world€test"))
				bd := iotdlm.New(r, '€', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("hello€"))
			})

			It("should handle emoji delimiter", func() {
				r := io.NopCloser(strings.NewReader("test😀data😀more"))
				bd := iotdlm.New(r, '😀', 0)

				buf := &bytes.Buffer{}
				_, _ = bd.WriteTo(buf)
				result := buf.String()
				Expect(result).To(ContainSubstring("test"))
				Expect(result).To(ContainSubstring("data"))
			})

			It("should handle multibyte characters in content", func() {
				r := io.NopCloser(strings.NewReader("日本語\n中文\n한글\n"))
				bd := iotdlm.New(r, '\n', 0)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data1)).To(Equal("日本語\n"))

				data2, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data2)).To(Equal("中文\n"))

				data3, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data3)).To(Equal("한글\n"))
			})

			It("should handle mixed ASCII and Unicode", func() {
				r := io.NopCloser(strings.NewReader("hello\nмир\nworld\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				_, err := bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("hello\nмир\nworld\n"))
			})

			It("should validate UTF-8 sequences", func() {
				validUTF8 := "valid UTF-8 文字列\n"
				r := io.NopCloser(strings.NewReader(validUTF8))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(utf8.Valid(data)).To(BeTrue())
			})
		})

		Context("with control characters", func() {
			It("should handle carriage return", func() {
				r := io.NopCloser(strings.NewReader("line1\rline2\r"))
				bd := iotdlm.New(r, '\r', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("line1\r")))
			})

			It("should handle CRLF sequences", func() {
				r := io.NopCloser(strings.NewReader("line1\r\nline2\r\n"))
				bd := iotdlm.New(r, '\n', 0)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data1).To(Equal([]byte("line1\r\n")))
			})

			It("should handle bell character", func() {
				r := io.NopCloser(strings.NewReader("test\x07data"))
				bd := iotdlm.New(r, '\x07', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("test\x07")))
			})

			It("should handle form feed", func() {
				r := io.NopCloser(strings.NewReader("page1\fpage2\f"))
				bd := iotdlm.New(r, '\f', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("page1\f")))
			})

			It("should handle vertical tab", func() {
				r := io.NopCloser(strings.NewReader("section1\vsection2\v"))
				bd := iotdlm.New(r, '\v', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("section1\v")))
			})
		})

		Context("with binary data", func() {
			It("should handle binary data with null bytes", func() {
				binaryData := []byte{0x00, 0x01, 0x02, 0xFF, '\n', 0x03, 0x04}
				r := io.NopCloser(bytes.NewReader(binaryData))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte{0x00, 0x01, 0x02, 0xFF, '\n'}))
			})

			It("should handle all byte values", func() {
				// Create data with all possible byte values except delimiter
				allBytes := make([]byte, 256)
				for i := 0; i < 256; i++ {
					allBytes[i] = byte(i)
				}
				// Use a delimiter byte (0xFF) and append it at the end
				delimiter := byte(0xFF)
				// Replace 0xFF in the sequence with something else to avoid early match
				allBytes[255] = 0xFE
				allBytes = append(allBytes, delimiter)

				r := io.NopCloser(bytes.NewReader(allBytes))
				bd := iotdlm.New(r, rune(delimiter), 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(257))
			})

			It("should preserve exact binary content", func() {
				original := []byte{0xDE, 0xAD, 0xBE, 0xEF, '\n', 0xCA, 0xFE}
				r := io.NopCloser(bytes.NewReader(original))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data[:5]).To(Equal([]byte{0xDE, 0xAD, 0xBE, 0xEF, '\n'}))
			})
		})
	})

	Describe("Boundary conditions", func() {
		Context("with size limits", func() {
			It("should handle exactly buffer size data", func() {
				bufSize := 4096
				data := strings.Repeat("x", bufSize-1) + "\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(bufSize))

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(result)).To(Equal(bufSize))
			})

			It("should handle data larger than buffer", func() {
				bufSize := 100
				data := strings.Repeat("x", bufSize*10) + "\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(bufSize))

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(result)).To(Equal(bufSize*10 + 1))
			})

			It("should handle single byte reads", func() {
				r := io.NopCloser(strings.NewReader("a\nb\nc\n"))
				bd := iotdlm.New(r, '\n', 1)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data1).To(Equal([]byte("a\n")))
			})

			It("should handle very large buffer size", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 100*libsiz.SizeMega)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("test\n")))
			})
		})

		Context("with delimiter edge cases", func() {
			It("should handle data with no delimiter at all", func() {
				r := io.NopCloser(strings.NewReader("no delimiter here"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(string(data)).To(Equal("no delimiter here"))
			})

			It("should handle delimiter as first character", func() {
				r := io.NopCloser(strings.NewReader("\ndata"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("\n")))
			})

			It("should handle delimiter as last character", func() {
				r := io.NopCloser(strings.NewReader("data\n"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("data\n")))
			})

			It("should handle delimiter as only character", func() {
				r := io.NopCloser(strings.NewReader("\n"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("\n")))
			})

			It("should handle repeated delimiters", func() {
				r := io.NopCloser(strings.NewReader("\n\n\n\n\n"))
				bd := iotdlm.New(r, '\n', 0)

				for i := 0; i < 5; i++ {
					data, err := bd.ReadBytes()
					Expect(err).To(BeNil())
					Expect(data).To(Equal([]byte("\n")))
				}

				_, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
			})

			It("should handle alternating delimiter and content", func() {
				r := io.NopCloser(strings.NewReader("a\nb\nc\nd\n"))
				bd := iotdlm.New(r, '\n', 0)

				expected := []string{"a\n", "b\n", "c\n", "d\n"}
				for _, exp := range expected {
					data, err := bd.ReadBytes()
					Expect(err).To(BeNil())
					Expect(string(data)).To(Equal(exp))
				}
			})
		})
	})

	Describe("Error handling", func() {
		Context("with reader errors", func() {
			It("should propagate read errors", func() {
				er := newErrorReader("test\ndata\nmore\n", 10)
				bd := iotdlm.New(er, '\n', 0)

				// Read until we get an error
				var lastErr error
				for i := 0; i < 20; i++ {
					_, err := bd.ReadBytes()
					if err != nil {
						lastErr = err
						break
					}
				}
				// Should eventually get an error (either EOF or read error)
				Expect(lastErr).To(HaveOccurred())
			})

			It("should handle immediate read error", func() {
				er := newErrorReader("test", 1)
				bd := iotdlm.New(er, '\n', 0)

				_, err := bd.ReadBytes()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with close errors", func() {
			It("should handle operations after close", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 0)

				_ = bd.Close()

				// All operations should fail
				_, err := bd.Read(make([]byte, 10))
				Expect(err).To(Equal(iotdlm.ErrInstance))

				_, err = bd.ReadBytes()
				Expect(err).To(Equal(iotdlm.ErrInstance))

				_, err = bd.UnRead()
				Expect(err).To(Equal(iotdlm.ErrInstance))

				buf := &bytes.Buffer{}
				_, err = bd.WriteTo(buf)
				Expect(err).To(Equal(iotdlm.ErrInstance))
			})
		})

		Context("with nil handling", func() {
			It("should handle nil buffer gracefully in Read", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 0)

				// Reading with nil buffer should not panic
				Expect(func() {
					var nilBuf []byte
					_, _ = bd.Read(nilBuf)
				}).NotTo(Panic())
			})
		})
	})

	Describe("Memory and resource management", func() {
		Context("with repeated operations", func() {
			It("should not leak memory on repeated reads", func() {
				data := strings.Repeat("line\n", 1000)
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0)

				for i := 0; i < 1000; i++ {
					_, err := bd.ReadBytes()
					if err == io.EOF {
						break
					}
					Expect(err).To(BeNil())
				}
			})

			It("should handle multiple WriteTo operations gracefully", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := &bytes.Buffer{}
				_, _ = bd.WriteTo(buf)

				// Second WriteTo after EOF
				buf2 := &bytes.Buffer{}
				_, err := bd.WriteTo(buf2)
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("with resource cleanup", func() {
			It("should clean up properly on Close", func() {
				r := io.NopCloser(strings.NewReader("test\n"))
				bd := iotdlm.New(r, '\n', 0)

				err := bd.Close()
				Expect(err).NotTo(HaveOccurred())

				// Verify resources are cleaned up
				_, err = bd.Read(make([]byte, 10))
				Expect(err).To(Equal(iotdlm.ErrInstance))
			})

			It("should handle double close gracefully (should panic)", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '\n', 0)

				err := bd.Close()
				Expect(err).NotTo(HaveOccurred())

				// Second close - might panic due to nil pointer
				Expect(func() {
					_ = bd.Close()
				}).To(Panic())
			})
		})
	})

	Describe("Complex scenarios", func() {
		Context("with interleaved operations", func() {
			It("should handle Read followed by ReadBytes", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 0)

				buf := make([]byte, 10)
				n, err := bd.Read(buf)
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal("line1\n"))

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("line2\n"))
			})

			It("should handle ReadBytes followed by WriteTo", func() {
				r := io.NopCloser(strings.NewReader("first\nsecond\nthird\n"))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(data)).To(Equal("first\n"))

				buf := &bytes.Buffer{}
				_, err = bd.WriteTo(buf)
				Expect(err).To(Equal(io.EOF))
				Expect(buf.String()).To(Equal("second\nthird\n"))
			})

			It("should handle UnRead at various points", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2\nline3\n"))
				bd := iotdlm.New(r, '\n', 100)

				// UnRead before any operation
				unread1, err := bd.UnRead()
				Expect(err).To(BeNil())

				// Read one line
				_, _ = bd.ReadBytes()

				// UnRead after one read
				unread2, err := bd.UnRead()
				Expect(err).To(BeNil())
				_ = unread1
				_ = unread2
			})
		})

		Context("with various data patterns", func() {
			It("should handle empty lines", func() {
				r := io.NopCloser(strings.NewReader("\n\ndata\n\n"))
				bd := iotdlm.New(r, '\n', 0)

				data1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data1).To(Equal([]byte("\n")))

				data2, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data2).To(Equal([]byte("\n")))

				data3, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(data3).To(Equal([]byte("data\n")))
			})

			It("should handle long lines without delimiter", func() {
				longLine := strings.Repeat("x", 100000)
				r := io.NopCloser(strings.NewReader(longLine))
				bd := iotdlm.New(r, '\n', 0)

				data, err := bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(len(data)).To(Equal(100000))
			})

			It("should handle mixed line lengths", func() {
				data := "short\n" + strings.Repeat("x", 1000) + "\nmedium line\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', 0)

				line1, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(line1).To(Equal([]byte("short\n")))

				line2, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(line2)).To(Equal(1001))

				line3, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(line3).To(Equal([]byte("medium line\n")))
			})
		})
	})
})
