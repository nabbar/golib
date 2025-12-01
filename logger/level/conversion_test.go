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

package level_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	loglvl "github.com/nabbar/golib/logger/level"
)

var _ = Describe("Level Conversion Methods", func() {
	Describe("Int", func() {
		Context("with all defined levels", func() {
			It("should convert PanicLevel to 0", func() {
				Expect(loglvl.PanicLevel.Int()).To(Equal(0))
			})

			It("should convert FatalLevel to 1", func() {
				Expect(loglvl.FatalLevel.Int()).To(Equal(1))
			})

			It("should convert ErrorLevel to 2", func() {
				Expect(loglvl.ErrorLevel.Int()).To(Equal(2))
			})

			It("should convert WarnLevel to 3", func() {
				Expect(loglvl.WarnLevel.Int()).To(Equal(3))
			})

			It("should convert InfoLevel to 4", func() {
				Expect(loglvl.InfoLevel.Int()).To(Equal(4))
			})

			It("should convert DebugLevel to 5", func() {
				Expect(loglvl.DebugLevel.Int()).To(Equal(5))
			})

			It("should convert NilLevel to 6", func() {
				Expect(loglvl.NilLevel.Int()).To(Equal(6))
			})
		})

		Context("with roundtrip conversion", func() {
			It("should maintain value integrity", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
					loglvl.NilLevel,
				}

				for _, level := range levels {
					Expect(level.Int()).To(Equal(int(level)))
				}
			})
		})
	})

	Describe("Uint32", func() {
		Context("with all defined levels", func() {
			It("should convert PanicLevel to 0", func() {
				Expect(loglvl.PanicLevel.Uint32()).To(Equal(uint32(0)))
			})

			It("should convert FatalLevel to 1", func() {
				Expect(loglvl.FatalLevel.Uint32()).To(Equal(uint32(1)))
			})

			It("should convert ErrorLevel to 2", func() {
				Expect(loglvl.ErrorLevel.Uint32()).To(Equal(uint32(2)))
			})

			It("should convert WarnLevel to 3", func() {
				Expect(loglvl.WarnLevel.Uint32()).To(Equal(uint32(3)))
			})

			It("should convert InfoLevel to 4", func() {
				Expect(loglvl.InfoLevel.Uint32()).To(Equal(uint32(4)))
			})

			It("should convert DebugLevel to 5", func() {
				Expect(loglvl.DebugLevel.Uint32()).To(Equal(uint32(5)))
			})

			It("should convert NilLevel to 6", func() {
				Expect(loglvl.NilLevel.Uint32()).To(Equal(uint32(6)))
			})
		})

		Context("with roundtrip conversion", func() {
			It("should maintain value integrity", func() {
				for i := uint32(0); i <= 6; i++ {
					level := loglvl.Level(i)
					Expect(level.Uint32()).To(Equal(i))
				}
			})
		})
	})

	Describe("Code", func() {
		Context("with all defined levels", func() {
			It("should return correct code for PanicLevel", func() {
				Expect(loglvl.PanicLevel.Code()).To(Equal("Crit"))
			})

			It("should return correct code for FatalLevel", func() {
				Expect(loglvl.FatalLevel.Code()).To(Equal("Fatal"))
			})

			It("should return correct code for ErrorLevel", func() {
				Expect(loglvl.ErrorLevel.Code()).To(Equal("Err"))
			})

			It("should return correct code for WarnLevel", func() {
				Expect(loglvl.WarnLevel.Code()).To(Equal("Warn"))
			})

			It("should return correct code for InfoLevel", func() {
				Expect(loglvl.InfoLevel.Code()).To(Equal("Info"))
			})

			It("should return correct code for DebugLevel", func() {
				Expect(loglvl.DebugLevel.Code()).To(Equal("Debug"))
			})

			It("should return empty string for NilLevel", func() {
				Expect(loglvl.NilLevel.Code()).To(Equal(""))
			})
		})

		Context("with unknown level", func() {
			It("should return unknown", func() {
				unknownLevel := loglvl.Level(99)
				Expect(unknownLevel.Code()).To(Equal("unknown"))
			})
		})

		Context("with code parsing", func() {
			It("should parse codes back to levels", func() {
				Expect(loglvl.Parse("Crit")).To(Equal(loglvl.PanicLevel))
				Expect(loglvl.Parse("Fatal")).To(Equal(loglvl.FatalLevel))
				Expect(loglvl.Parse("Err")).To(Equal(loglvl.ErrorLevel))
				Expect(loglvl.Parse("Warn")).To(Equal(loglvl.WarnLevel))
				Expect(loglvl.Parse("Info")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("Debug")).To(Equal(loglvl.DebugLevel))
			})
		})
	})

	Describe("ParseFromInt", func() {
		Context("with valid integer values", func() {
			It("should parse 0 to PanicLevel", func() {
				Expect(loglvl.ParseFromInt(0)).To(Equal(loglvl.PanicLevel))
			})

			It("should parse 1 to FatalLevel", func() {
				Expect(loglvl.ParseFromInt(1)).To(Equal(loglvl.FatalLevel))
			})

			It("should parse 2 to ErrorLevel", func() {
				Expect(loglvl.ParseFromInt(2)).To(Equal(loglvl.ErrorLevel))
			})

			It("should parse 3 to WarnLevel", func() {
				Expect(loglvl.ParseFromInt(3)).To(Equal(loglvl.WarnLevel))
			})

			It("should parse 4 to InfoLevel", func() {
				Expect(loglvl.ParseFromInt(4)).To(Equal(loglvl.InfoLevel))
			})

			It("should parse 5 to DebugLevel", func() {
				Expect(loglvl.ParseFromInt(5)).To(Equal(loglvl.DebugLevel))
			})

			It("should parse 6 to NilLevel", func() {
				Expect(loglvl.ParseFromInt(6)).To(Equal(loglvl.NilLevel))
			})
		})

		Context("with invalid integer values", func() {
			It("should return InfoLevel for negative values", func() {
				Expect(loglvl.ParseFromInt(-1)).To(Equal(loglvl.InfoLevel))
			})

			It("should return InfoLevel for values > 6", func() {
				Expect(loglvl.ParseFromInt(7)).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.ParseFromInt(10)).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.ParseFromInt(99)).To(Equal(loglvl.InfoLevel))
			})

			It("should return InfoLevel for very large values", func() {
				Expect(loglvl.ParseFromInt(math.MaxInt)).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with roundtrip conversion", func() {
			It("should maintain level identity", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
					loglvl.NilLevel,
				}

				for _, level := range levels {
					converted := loglvl.ParseFromInt(level.Int())
					Expect(converted).To(Equal(level))
				}
			})
		})
	})

	Describe("ParseFromUint32", func() {
		Context("with valid uint32 values", func() {
			It("should parse 0 to PanicLevel", func() {
				Expect(loglvl.ParseFromUint32(0)).To(Equal(loglvl.PanicLevel))
			})

			It("should parse 1 to FatalLevel", func() {
				Expect(loglvl.ParseFromUint32(1)).To(Equal(loglvl.FatalLevel))
			})

			It("should parse 2 to ErrorLevel", func() {
				Expect(loglvl.ParseFromUint32(2)).To(Equal(loglvl.ErrorLevel))
			})

			It("should parse 3 to WarnLevel", func() {
				Expect(loglvl.ParseFromUint32(3)).To(Equal(loglvl.WarnLevel))
			})

			It("should parse 4 to InfoLevel", func() {
				Expect(loglvl.ParseFromUint32(4)).To(Equal(loglvl.InfoLevel))
			})

			It("should parse 5 to DebugLevel", func() {
				Expect(loglvl.ParseFromUint32(5)).To(Equal(loglvl.DebugLevel))
			})

			It("should parse 6 to NilLevel", func() {
				Expect(loglvl.ParseFromUint32(6)).To(Equal(loglvl.NilLevel))
			})
		})

		Context("with invalid uint32 values", func() {
			It("should return InfoLevel for values > 6", func() {
				Expect(loglvl.ParseFromUint32(7)).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.ParseFromUint32(10)).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.ParseFromUint32(99)).To(Equal(loglvl.InfoLevel))
			})

			It("should handle large values", func() {
				// Large values should return InfoLevel
				Expect(loglvl.ParseFromUint32(1000)).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.ParseFromUint32(1000000)).To(Equal(loglvl.InfoLevel))
			})

			It("should handle maximum uint32 value", func() {
				// math.MaxUint32 should be clamped and return InfoLevel
				Expect(loglvl.ParseFromUint32(math.MaxUint32)).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with roundtrip conversion", func() {
			It("should maintain level identity for valid values", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
					loglvl.NilLevel,
				}

				for _, level := range levels {
					converted := loglvl.ParseFromUint32(level.Uint32())
					Expect(converted).To(Equal(level))
				}
			})
		})

		Context("with boundary values", func() {
			It("should handle zero", func() {
				Expect(loglvl.ParseFromUint32(0)).To(Equal(loglvl.PanicLevel))
			})

			It("should handle max valid level value", func() {
				Expect(loglvl.ParseFromUint32(6)).To(Equal(loglvl.NilLevel))
			})

			It("should handle first invalid value", func() {
				Expect(loglvl.ParseFromUint32(7)).To(Equal(loglvl.InfoLevel))
			})
		})
	})

	Describe("Integration tests for all conversion methods", func() {
		Context("with complete conversion chain", func() {
			It("should convert through all methods consistently", func() {
				original := loglvl.WarnLevel

				// Test all conversion methods
				u8 := original.Uint8()
				u32 := original.Uint32()
				i := original.Int()
				str := original.String()
				code := original.Code()

				Expect(u8).To(Equal(uint8(3)))
				Expect(u32).To(Equal(uint32(3)))
				Expect(i).To(Equal(3))
				Expect(str).To(Equal("Warning"))
				Expect(code).To(Equal("Warn"))

				// Verify roundtrips
				Expect(loglvl.ParseFromInt(i)).To(Equal(original))
				Expect(loglvl.ParseFromUint32(u32)).To(Equal(original))
				Expect(loglvl.Parse(str)).To(Equal(original))
				Expect(loglvl.Parse(code)).To(Equal(original))
			})
		})

		Context("with all levels", func() {
			It("should convert consistently across all methods", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
					loglvl.NilLevel,
				}

				for _, level := range levels {
					// Verify all numeric conversions match
					Expect(int(level.Uint8())).To(Equal(level.Int()))
					Expect(int(level.Uint32())).To(Equal(level.Int()))

					// Verify parsing consistency
					Expect(loglvl.ParseFromInt(level.Int())).To(Equal(level))
					Expect(loglvl.ParseFromUint32(level.Uint32())).To(Equal(level))

					// Verify string parsing (except NilLevel)
					if level != loglvl.NilLevel {
						Expect(loglvl.Parse(level.String())).To(Equal(level))
						Expect(loglvl.Parse(level.Code())).To(Equal(level))
					}
				}
			})
		})
	})
})
