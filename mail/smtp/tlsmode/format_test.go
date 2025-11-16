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

package tlsmode_test

import (
	. "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS Mode Format and Conversion", func() {

	Describe("String Conversion", func() {
		It("should convert TLSNone to empty string", func() {
			Expect(TLSNone.String()).To(Equal(""))
		})

		It("should convert TLSStartTLS to 'starttls'", func() {
			Expect(TLSStartTLS.String()).To(Equal("starttls"))
		})

		It("should convert TLSStrictTLS to 'tls'", func() {
			Expect(TLSStrictTLS.String()).To(Equal("tls"))
		})

		It("should be consistent with Parse (roundtrip)", func() {
			modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
			for _, mode := range modes {
				str := mode.String()
				parsed := Parse(str)
				Expect(parsed).To(Equal(mode), "Mode %v should roundtrip through string '%s'", mode, str)
			}
		})
	})

	Describe("Uint Conversions", func() {
		Context("Uint method", func() {
			It("should convert TLSNone to 0", func() {
				Expect(TLSNone.Uint()).To(Equal(uint8(0)))
			})

			It("should convert TLSStartTLS to 1", func() {
				Expect(TLSStartTLS.Uint()).To(Equal(uint8(1)))
			})

			It("should convert TLSStrictTLS to 2", func() {
				Expect(TLSStrictTLS.Uint()).To(Equal(uint8(2)))
			})
		})

		Context("Uint32 method", func() {
			It("should convert all modes correctly", func() {
				Expect(TLSNone.Uint32()).To(Equal(uint32(0)))
				Expect(TLSStartTLS.Uint32()).To(Equal(uint32(1)))
				Expect(TLSStrictTLS.Uint32()).To(Equal(uint32(2)))
			})
		})

		Context("Uint64 method", func() {
			It("should convert all modes correctly", func() {
				Expect(TLSNone.Uint64()).To(Equal(uint64(0)))
				Expect(TLSStartTLS.Uint64()).To(Equal(uint64(1)))
				Expect(TLSStrictTLS.Uint64()).To(Equal(uint64(2)))
			})

			It("should be consistent with ParseUint64 (roundtrip)", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, mode := range modes {
					val := mode.Uint64()
					parsed := ParseUint64(val)
					Expect(parsed).To(Equal(mode))
				}
			})
		})
	})

	Describe("Int Conversions", func() {
		Context("Int method", func() {
			It("should convert TLSNone to 0", func() {
				Expect(TLSNone.Int()).To(Equal(int(0)))
			})

			It("should convert TLSStartTLS to 1", func() {
				Expect(TLSStartTLS.Int()).To(Equal(int(1)))
			})

			It("should convert TLSStrictTLS to 2", func() {
				Expect(TLSStrictTLS.Int()).To(Equal(int(2)))
			})
		})

		Context("Int32 method", func() {
			It("should convert all modes correctly", func() {
				Expect(TLSNone.Int32()).To(Equal(int32(0)))
				Expect(TLSStartTLS.Int32()).To(Equal(int32(1)))
				Expect(TLSStrictTLS.Int32()).To(Equal(int32(2)))
			})
		})

		Context("Int64 method", func() {
			It("should convert all modes correctly", func() {
				Expect(TLSNone.Int64()).To(Equal(int64(0)))
				Expect(TLSStartTLS.Int64()).To(Equal(int64(1)))
				Expect(TLSStrictTLS.Int64()).To(Equal(int64(2)))
			})

			It("should be consistent with ParseInt64 (roundtrip)", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, mode := range modes {
					val := mode.Int64()
					parsed := ParseInt64(val)
					Expect(parsed).To(Equal(mode))
				}
			})
		})
	})

	Describe("Float Conversions", func() {
		Context("Float32 method", func() {
			It("should convert TLSNone to 0.0", func() {
				Expect(TLSNone.Float32()).To(Equal(float32(0.0)))
			})

			It("should convert TLSStartTLS to 1.0", func() {
				Expect(TLSStartTLS.Float32()).To(Equal(float32(1.0)))
			})

			It("should convert TLSStrictTLS to 2.0", func() {
				Expect(TLSStrictTLS.Float32()).To(Equal(float32(2.0)))
			})
		})

		Context("Float64 method", func() {
			It("should convert all modes correctly", func() {
				Expect(TLSNone.Float64()).To(Equal(float64(0.0)))
				Expect(TLSStartTLS.Float64()).To(Equal(float64(1.0)))
				Expect(TLSStrictTLS.Float64()).To(Equal(float64(2.0)))
			})

			It("should be consistent with ParseFloat64 (roundtrip)", func() {
				modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
				for _, mode := range modes {
					val := mode.Float64()
					parsed := ParseFloat64(val)
					Expect(parsed).To(Equal(mode))
				}
			})
		})
	})

	Describe("Type Consistency", func() {
		It("should have consistent values across all numeric types", func() {
			testMode := TLSStartTLS

			Expect(int(testMode.Uint())).To(Equal(testMode.Int()))
			Expect(int64(testMode.Uint32())).To(Equal(testMode.Int64()))
			Expect(uint64(testMode.Int64())).To(Equal(testMode.Uint64()))
			Expect(int(testMode.Float32())).To(Equal(testMode.Int()))
			Expect(int(testMode.Float64())).To(Equal(testMode.Int()))
		})

		It("should maintain value consistency for all modes", func() {
			modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}
			for _, mode := range modes {
				intVal := mode.Int()
				Expect(mode.Uint()).To(Equal(uint8(intVal)))
				Expect(mode.Int32()).To(Equal(int32(intVal)))
				Expect(mode.Int64()).To(Equal(int64(intVal)))
				Expect(mode.Uint32()).To(Equal(uint32(intVal)))
				Expect(mode.Uint64()).To(Equal(uint64(intVal)))
				Expect(mode.Float32()).To(Equal(float32(intVal)))
				Expect(mode.Float64()).To(Equal(float64(intVal)))
			}
		})
	})
})
