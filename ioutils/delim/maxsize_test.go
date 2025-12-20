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
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"
)

var _ = Describe("MaxPartSize protection", func() {
	Describe("Basic maxPartSize behavior", func() {
		Context("when maxPartSize is not set (0)", func() {
			It("should allow unlimited part size", func() {
				largeData := strings.Repeat("x", 10000) + "\n"
				r := io.NopCloser(strings.NewReader(largeData))
				bd := iotdlm.New(r, '\n', 0, false)
				defer bd.Close()

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(10001))
			})

			It("should work with very large parts", func() {
				veryLargeData := strings.Repeat("a", 1000000) + "\n"
				r := io.NopCloser(strings.NewReader(veryLargeData))
				bd := iotdlm.New(r, '\n', 0, true)
				defer bd.Close()

				data, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(32 * 1024))

				data, err = bd.ReadBytes()
				Expect(err).To(Equal(io.EOF))
				Expect(data).To(BeNil())
			})
		})

		Context("when maxPartSize is set", func() {
			It("should allow parts within limit", func() {
				data := strings.Repeat("x", 50) + "\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(100), false)
				defer bd.Close()

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(result)).To(Equal(51))
			})

			It("should truncate parts exceeding limit to maxPartSize-1 bytes plus delimiter", func() {
				data := strings.Repeat("x", 200) + "\n" + "ok\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
				defer bd.Close()

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(result)).To(Equal(strings.Repeat("x", 99) + "\n"))
			})

			It("should truncate at limit boundary to maxPartSize-1 bytes plus delimiter", func() {
				data := strings.Repeat("x", 101) + "\n" + "valid\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
				defer bd.Close()

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(string(result)).To(Equal(strings.Repeat("x", 99) + "\n"))
			})

			It("should allow part at exact limit", func() {
				data := strings.Repeat("x", 99) + "\n"
				r := io.NopCloser(strings.NewReader(data))
				bd := iotdlm.New(r, '\n', libsiz.Size(100), false)
				defer bd.Close()

				result, err := bd.ReadBytes()
				Expect(err).To(BeNil())
				Expect(len(result)).To(Equal(100))
			})
		})
	})

	Describe("Multiple parts with maxPartSize", func() {
		It("should handle mix of valid and oversized parts, truncating oversized ones", func() {
			data := "small\n" + strings.Repeat("x", 200) + "\n" + "ok\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			result1, err1 := bd.ReadBytes()
			Expect(err1).To(BeNil())
			Expect(string(result1)).To(Equal("small\n"))

			result2, err2 := bd.ReadBytes()
			Expect(err2).To(BeNil())
			Expect(string(result2)).To(Equal(strings.Repeat("x", 99) + "\n"))

			result3, err3 := bd.ReadBytes()
			Expect(err3).To(BeNil())
			Expect(string(result3)).To(Equal("ok\n"))

			result4, err4 := bd.ReadBytes()
			Expect(err4).To(Equal(io.EOF))
			Expect(result4).To(BeNil())
		})

		It("should handle consecutive oversized parts, truncating each to maxPartSize-1 plus delimiter", func() {
			data := strings.Repeat("x", 200) + "\n" + strings.Repeat("y", 150) + "\n" + "ok\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			result1, err1 := bd.ReadBytes()
			Expect(err1).To(BeNil())
			Expect(string(result1)).To(Equal(strings.Repeat("x", 99) + "\n"))

			result2, err2 := bd.ReadBytes()
			Expect(err2).To(BeNil())
			Expect(string(result2)).To(Equal(strings.Repeat("y", 99) + "\n"))

			result3, err3 := bd.ReadBytes()
			Expect(err3).To(BeNil())
			Expect(string(result3)).To(Equal("ok\n"))

			result4, err4 := bd.ReadBytes()
			Expect(err4).To(Equal(io.EOF))
			Expect(result4).To(BeNil())
		})
	})

	Describe("Read method with maxPartSize", func() {
		It("should truncate oversized part and return it with EOF", func() {
			data := strings.Repeat("x", 200) + "\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			buf := make([]byte, 1000)
			n, err := bd.Read(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(201))
		})

		It("should work normally for valid sized parts", func() {
			data := "test\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), false)
			defer bd.Close()

			buf := make([]byte, 10)
			n, err := bd.Read(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal("test\n"))
		})
	})

	Describe("WriteTo with maxPartSize", func() {
		It("should truncate oversized parts during WriteTo", func() {
			data := "ok1\n" + strings.Repeat("x", 200) + "\n" + "ok2\n"
			expt := "ok1\n" + strings.Repeat("x", 99) + "\n" + "ok2\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			var buf strings.Builder
			written, err := bd.WriteTo(&buf)

			Expect(err).To(Equal(io.EOF))
			Expect(written).To(Equal(int64(len(expt))))
			Expect(buf.String()).To(Equal(expt))
		})
	})

	Describe("Edge cases with maxPartSize", func() {
		It("should handle EOF without delimiter when oversized", func() {
			data := strings.Repeat("x", 200)
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(io.EOF))
			Expect(string(result)).To(Equal(strings.Repeat("x", 100)))
		})

		It("should return partial data on EOF when within limit", func() {
			data := "partial data without delimiter"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), true)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(io.EOF))
			Expect(string(result)).To(Equal(data))
		})

		It("should return partial data on EOF even with maxPartSize", func() {
			data := "short"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(10), true)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(io.EOF))
			Expect(string(result)).To(Equal("short"))
		})

		It("should handle empty data with maxPartSize", func() {
			r := io.NopCloser(strings.NewReader(""))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), false)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(io.EOF))
			Expect(result).To(BeNil())
		})

		It("should handle single delimiter with maxPartSize", func() {
			r := io.NopCloser(strings.NewReader("\n"))
			bd := iotdlm.New(r, '\n', libsiz.Size(100), false)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(BeNil())
			Expect(string(result)).To(Equal("\n"))
		})
	})

	Describe("Different delimiters with maxPartSize", func() {
		It("should truncate oversized parts with comma delimiter", func() {
			data := strings.Repeat("x", 200) + "," + "valid,"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, ',', libsiz.Size(100), true)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(BeNil())
			Expect(string(result)).To(Equal(strings.Repeat("x", 99) + ","))
		})

		It("should truncate oversized parts with pipe delimiter", func() {
			data := "ok|" + strings.Repeat("x", 200) + "|end|"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '|', libsiz.Size(100), true)
			defer bd.Close()

			result1, err1 := bd.ReadBytes()
			Expect(err1).To(BeNil())
			Expect(string(result1)).To(Equal("ok|"))

			result2, err2 := bd.ReadBytes()
			Expect(err2).To(BeNil())
			Expect(string(result2)).To(Equal(strings.Repeat("x", 99) + "|"))

			result3, err3 := bd.ReadBytes()
			Expect(err3).To(BeNil())
			Expect(string(result3)).To(Equal("end|"))

			result4, err4 := bd.ReadBytes()
			Expect(err4).To(Equal(io.EOF))
			Expect(result4).To(BeNil())
		})
	})

	Describe("Buffer size vs maxPartSize", func() {
		It("should work when maxPartSize allows all parts", func() {
			data := strings.Repeat("x", 60) + "\n" + "short\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(1024), false)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(61))
		})

		It("should return error when part exceeds buffer without discard", func() {
			data := strings.Repeat("x", 200) + "\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', libsiz.Size(64), false)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(iotdlm.ErrBufferFull))
			Expect(len(result)).To(Equal(64))
		})
	})

	Describe("Security scenarios", func() {
		It("should protect against memory exhaustion from missing delimiter", func() {
			data := strings.Repeat("x", 1000000)
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', 2*libsiz.SizeMega, false)
			defer bd.Close()

			result, err := bd.ReadBytes()
			Expect(err).To(Equal(io.EOF))
			Expect(len(result)).To(Equal(1000000))
		})

		It("should protect against DoS with realistic limits and discard", func() {
			data := strings.Repeat("a", 500000) + "\n" + "ok\n"
			r := io.NopCloser(strings.NewReader(data))
			bd := iotdlm.New(r, '\n', 100*libsiz.SizeKilo, true)
			defer bd.Close()

			result0, err0 := bd.ReadBytes()
			Expect(err0).To(BeNil())
			Expect(string(result0)).To(Equal(strings.Repeat("a", (100*1024)-1) + "\n"))

			result1, err1 := bd.ReadBytes()
			Expect(err1).To(BeNil())
			Expect(string(result1)).To(Equal("ok\n"))

			result2, err2 := bd.ReadBytes()
			Expect(err2).To(Equal(io.EOF))
			Expect(result2).To(BeNil())
		})
	})

	Context("ReadBytes with discard enabled and full buffer", func() {
		It("should replace last byte with delimiter when buffer is full and delimiter found after discard", func() {
			// Scenario:
			// 1. Buffer size 4.
			// 2. Reader returns "abcd" (4 bytes). Buffer full. No delimiter.
			// 3. Discard logic triggers.
			// 4. Reader returns "\n". Discard finds it.
			// 5. ReadBytes logic should see buffer full (nbr=4, len(p)=4).
			// 6. Should replace last byte 'd' with '\n' -> "abc\n".

			tr := &transientReader{
				data: []byte("abcd"), // fills buffer
				// second call returns '\n' (handled by transientReader logic if we set it up right)
			}

			// New with size 4.
			bd := iotdlm.New(tr, '\n', libsiz.Size(4), true)
			defer bd.Close()

			res, err := bd.ReadBytes()

			Expect(err).To(BeNil())
			// "abcd" -> last char replaced by '\n' -> "abc\n"
			Expect(string(res)).To(Equal("abc\n"))
		})
	})
})
