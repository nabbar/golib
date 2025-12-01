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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	loglvl "github.com/nabbar/golib/logger/level"
)

var _ = Describe("Level Constants and Values", func() {
	Describe("Level constants", func() {
		Context("with all defined levels", func() {
			It("should have correct uint8 values", func() {
				Expect(loglvl.PanicLevel.Uint8()).To(Equal(uint8(0)))
				Expect(loglvl.FatalLevel.Uint8()).To(Equal(uint8(1)))
				Expect(loglvl.ErrorLevel.Uint8()).To(Equal(uint8(2)))
				Expect(loglvl.WarnLevel.Uint8()).To(Equal(uint8(3)))
				Expect(loglvl.InfoLevel.Uint8()).To(Equal(uint8(4)))
				Expect(loglvl.DebugLevel.Uint8()).To(Equal(uint8(5)))
				Expect(loglvl.NilLevel.Uint8()).To(Equal(uint8(6)))
			})

			It("should be ordered from most to least severe", func() {
				Expect(loglvl.PanicLevel).To(BeNumerically("<", loglvl.FatalLevel))
				Expect(loglvl.FatalLevel).To(BeNumerically("<", loglvl.ErrorLevel))
				Expect(loglvl.ErrorLevel).To(BeNumerically("<", loglvl.WarnLevel))
				Expect(loglvl.WarnLevel).To(BeNumerically("<", loglvl.InfoLevel))
				Expect(loglvl.InfoLevel).To(BeNumerically("<", loglvl.DebugLevel))
				Expect(loglvl.DebugLevel).To(BeNumerically("<", loglvl.NilLevel))
			})
		})
	})

	Describe("Parse", func() {
		Context("with valid level strings", func() {
			It("should parse panic/critical level", func() {
				Expect(loglvl.Parse("critical")).To(Equal(loglvl.PanicLevel))
				Expect(loglvl.Parse("Critical")).To(Equal(loglvl.PanicLevel))
				Expect(loglvl.Parse("CRITICAL")).To(Equal(loglvl.PanicLevel))
			})

			It("should parse fatal level", func() {
				Expect(loglvl.Parse("fatal")).To(Equal(loglvl.FatalLevel))
				Expect(loglvl.Parse("Fatal")).To(Equal(loglvl.FatalLevel))
				Expect(loglvl.Parse("FATAL")).To(Equal(loglvl.FatalLevel))
			})

			It("should parse error level", func() {
				Expect(loglvl.Parse("error")).To(Equal(loglvl.ErrorLevel))
				Expect(loglvl.Parse("Error")).To(Equal(loglvl.ErrorLevel))
				Expect(loglvl.Parse("ERROR")).To(Equal(loglvl.ErrorLevel))
			})

			It("should parse warn/warning level", func() {
				Expect(loglvl.Parse("warning")).To(Equal(loglvl.WarnLevel))
				Expect(loglvl.Parse("Warning")).To(Equal(loglvl.WarnLevel))
				Expect(loglvl.Parse("WARNING")).To(Equal(loglvl.WarnLevel))
			})

			It("should parse info level", func() {
				Expect(loglvl.Parse("info")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("Info")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("INFO")).To(Equal(loglvl.InfoLevel))
			})

			It("should parse debug level", func() {
				Expect(loglvl.Parse("debug")).To(Equal(loglvl.DebugLevel))
				Expect(loglvl.Parse("Debug")).To(Equal(loglvl.DebugLevel))
				Expect(loglvl.Parse("DEBUG")).To(Equal(loglvl.DebugLevel))
			})
		})

		Context("with invalid level strings", func() {
			It("should return InfoLevel for unknown strings", func() {
				Expect(loglvl.Parse("unknown")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("invalid")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("trace")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("nil")).To(Equal(loglvl.InfoLevel))
			})
		})

		Context("with mixed case strings", func() {
			It("should be case-insensitive", func() {
				Expect(loglvl.Parse("DeBuG")).To(Equal(loglvl.DebugLevel))
				Expect(loglvl.Parse("iNfO")).To(Equal(loglvl.InfoLevel))
				Expect(loglvl.Parse("WaRnInG")).To(Equal(loglvl.WarnLevel))
				Expect(loglvl.Parse("ErRoR")).To(Equal(loglvl.ErrorLevel))
			})
		})

		Context("with whitespace", func() {
			It("should not handle leading/trailing whitespace", func() {
				// Note: Parse does not trim whitespace
				Expect(loglvl.Parse(" info ")).To(Equal(loglvl.InfoLevel)) // Falls back to InfoLevel
				Expect(loglvl.Parse(" info")).To(Equal(loglvl.InfoLevel))  // Falls back to InfoLevel
				Expect(loglvl.Parse("info ")).To(Equal(loglvl.InfoLevel))  // Falls back to InfoLevel
			})
		})
	})

	Describe("ListLevels", func() {
		Context("when called", func() {
			It("should return all log levels", func() {
				levels := loglvl.ListLevels()

				Expect(levels).To(HaveLen(6))
				Expect(levels).To(ContainElement("critical"))
				Expect(levels).To(ContainElement("fatal"))
				Expect(levels).To(ContainElement("error"))
				Expect(levels).To(ContainElement("warning"))
				Expect(levels).To(ContainElement("info"))
				Expect(levels).To(ContainElement("debug"))
			})

			It("should return parseable levels", func() {
				levels := loglvl.ListLevels()

				for _, levelStr := range levels {
					parsed := loglvl.Parse(levelStr)
					// All returned levels should parse to a valid level
					Expect(parsed).ToNot(Equal(loglvl.Level(255)))
				}
			})

			It("should not include NilLevel", func() {
				levels := loglvl.ListLevels()

				Expect(levels).ToNot(ContainElement("nil"))
				Expect(levels).ToNot(ContainElement(""))
			})

			It("should return levels in order", func() {
				levels := loglvl.ListLevels()

				Expect(levels[0]).To(Equal("critical"))
				Expect(levels[1]).To(Equal("fatal"))
				Expect(levels[2]).To(Equal("error"))
				Expect(levels[3]).To(Equal("warning"))
				Expect(levels[4]).To(Equal("info"))
				Expect(levels[5]).To(Equal("debug"))
			})
		})
	})
})
