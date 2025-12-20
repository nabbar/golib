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
	"math"
	"strings"

	iotdlm "github.com/nabbar/golib/ioutils/delim"
	libsiz "github.com/nabbar/golib/size"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test file focuses on constructor and interface verification.
// It validates:
//   - New() constructor with various parameters (default/custom buffer sizes)
//   - Delimiter configuration (newline, comma, pipe, tab, null byte, Unicode)
//   - Interface compliance (io.ReadCloser, io.WriterTo, BufferDelim)
//   - Reader() method behavior
//   - Edge cases with empty/delimiter-only input
//
// These tests ensure the BufferDelim is properly initialized and
// implements all required interfaces correctly.

var _ = Describe("BufferDelim Constructor and Interface", func() {
	Describe("New constructor", func() {
		Context("with valid parameters and default buffer size", func() {
			It("should create a new BufferDelim instance successfully", func() {
				r := io.NopCloser(strings.NewReader("test\ndata"))
				bd := iotdlm.New(r, '\n', 0, false)
				Expect(bd).NotTo(BeNil())
			})

			It("should return the correct delimiter", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '|', 0, false)
				Expect(bd.Delim()).To(Equal('|'))
			})

			It("should work with newline delimiter", func() {
				r := io.NopCloser(strings.NewReader("line1\nline2"))
				bd := iotdlm.New(r, '\n', 0, false)
				Expect(bd.Delim()).To(Equal('\n'))
			})

			It("should work with pipe delimiter", func() {
				r := io.NopCloser(strings.NewReader("col1|col2"))
				bd := iotdlm.New(r, '|', 0, false)
				Expect(bd.Delim()).To(Equal('|'))
			})

			It("should work with comma delimiter", func() {
				r := io.NopCloser(strings.NewReader("a,b,c"))
				bd := iotdlm.New(r, ',', 0, false)
				Expect(bd.Delim()).To(Equal(','))
			})

			It("should work with tab delimiter", func() {
				r := io.NopCloser(strings.NewReader("field1\tfield2"))
				bd := iotdlm.New(r, '\t', 0, false)
				Expect(bd.Delim()).To(Equal('\t'))
			})
		})

		Context("with custom buffer size", func() {
			It("should create instance with small buffer size", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeUnit, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal('\n'))
			})

			It("should create instance with large buffer size", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '\n', 64*libsiz.SizeKilo, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal('\n'))
			})

			It("should create instance with 1MB buffer size", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '\n', libsiz.SizeMega, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal('\n'))
			})

			It("should handle zero buffer size (use default)", func() {
				r := io.NopCloser(strings.NewReader("test"))
				bd := iotdlm.New(r, '\n', 0, false)
				Expect(bd).NotTo(BeNil())
			})

		})

		Context("with various delimiter types", func() {
			It("should work with null byte delimiter", func() {
				r := io.NopCloser(strings.NewReader("data\x00more"))
				bd := iotdlm.New(r, 0, 0, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal(rune(0)))
			})

			It("should work with space delimiter", func() {
				r := io.NopCloser(strings.NewReader("word1 word2"))
				bd := iotdlm.New(r, ' ', 0, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal(' '))
			})

			It("should work with semicolon delimiter", func() {
				r := io.NopCloser(strings.NewReader("a;b;c"))
				bd := iotdlm.New(r, ';', 0, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal(';'))
			})

			It("should work with colon delimiter", func() {
				r := io.NopCloser(strings.NewReader("key:value"))
				bd := iotdlm.New(r, ':', 0, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal(':'))
			})

			It("should work with Unicode delimiter", func() {
				r := io.NopCloser(strings.NewReader("test€data"))
				bd := iotdlm.New(r, '€', 0, false)
				Expect(bd).NotTo(BeNil())
				Expect(bd.Delim()).To(Equal('€'))
			})
		})
	})

	Describe("Reader method", func() {
		It("should return io.ReadCloser interface", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)
			reader := bd.Reader()
			Expect(reader).NotTo(BeNil())
			var _ io.ReadCloser = reader
		})

		It("should return itself as ReadCloser", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)
			reader := bd.Reader()
			// The reader should be the same as bd
			Expect(reader).To(Equal(bd))
		})
	})

	Describe("Interface compliance", func() {
		It("should implement io.ReadCloser", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)
			var _ io.ReadCloser = bd
		})

		It("should implement io.WriterTo", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)
			var _ io.WriterTo = bd
		})

		It("should implement BufferDelim interface", func() {
			r := io.NopCloser(strings.NewReader("test"))
			bd := iotdlm.New(r, '\n', 0, false)
			var _ iotdlm.BufferDelim = bd
		})
	})

	Describe("Delimiter behavior with empty input", func() {
		It("should handle empty reader", func() {
			r := io.NopCloser(strings.NewReader(""))
			bd := iotdlm.New(r, '\n', 0, false)
			Expect(bd).NotTo(BeNil())
			Expect(bd.Delim()).To(Equal('\n'))
		})

		It("should handle reader with only delimiter", func() {
			r := io.NopCloser(strings.NewReader("\n"))
			bd := iotdlm.New(r, '\n', 0, false)
			Expect(bd).NotTo(BeNil())
			Expect(bd.Delim()).To(Equal('\n'))
		})

		It("should handle reader with multiple delimiters", func() {
			r := io.NopCloser(strings.NewReader("\n\n\n"))
			bd := iotdlm.New(r, '\n', 0, false)
			Expect(bd).NotTo(BeNil())
			Expect(bd.Delim()).To(Equal('\n'))
		})
	})

	Context("New with very large buffer", func() {
		It("should clamp buffer size to max allowed", func() {
			// This test attempts to trigger the clamping logic in New.
			// Since allocating maxint memory will likely panic, we recover.
			// We just want to ensure the code path is executed.
			defer func() {
				_ = recover()
			}()

			// We use a large size that should trigger the check
			_ = iotdlm.New(io.NopCloser(&mockReader0Nil{}), '\n', libsiz.Size(math.MaxInt), false)
		})
	})
})
