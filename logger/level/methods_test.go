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
	"github.com/sirupsen/logrus"

	loglvl "github.com/nabbar/golib/logger/level"
)

var _ = Describe("Level Methods", func() {
	Describe("String", func() {
		Context("with all defined levels", func() {
			It("should return correct string for PanicLevel", func() {
				Expect(loglvl.PanicLevel.String()).To(Equal("Critical"))
			})

			It("should return correct string for FatalLevel", func() {
				Expect(loglvl.FatalLevel.String()).To(Equal("Fatal"))
			})

			It("should return correct string for ErrorLevel", func() {
				Expect(loglvl.ErrorLevel.String()).To(Equal("Error"))
			})

			It("should return correct string for WarnLevel", func() {
				Expect(loglvl.WarnLevel.String()).To(Equal("Warning"))
			})

			It("should return correct string for InfoLevel", func() {
				Expect(loglvl.InfoLevel.String()).To(Equal("Info"))
			})

			It("should return correct string for DebugLevel", func() {
				Expect(loglvl.DebugLevel.String()).To(Equal("Debug"))
			})

			It("should return empty string for NilLevel", func() {
				Expect(loglvl.NilLevel.String()).To(Equal(""))
			})
		})

		Context("with unknown level", func() {
			It("should return unknown", func() {
				unknownLevel := loglvl.Level(99)
				Expect(unknownLevel.String()).To(Equal("unknown"))
			})
		})
	})

	Describe("Uint8", func() {
		Context("with all defined levels", func() {
			It("should convert PanicLevel to 0", func() {
				Expect(loglvl.PanicLevel.Uint8()).To(Equal(uint8(0)))
			})

			It("should convert FatalLevel to 1", func() {
				Expect(loglvl.FatalLevel.Uint8()).To(Equal(uint8(1)))
			})

			It("should convert ErrorLevel to 2", func() {
				Expect(loglvl.ErrorLevel.Uint8()).To(Equal(uint8(2)))
			})

			It("should convert WarnLevel to 3", func() {
				Expect(loglvl.WarnLevel.Uint8()).To(Equal(uint8(3)))
			})

			It("should convert InfoLevel to 4", func() {
				Expect(loglvl.InfoLevel.Uint8()).To(Equal(uint8(4)))
			})

			It("should convert DebugLevel to 5", func() {
				Expect(loglvl.DebugLevel.Uint8()).To(Equal(uint8(5)))
			})

			It("should convert NilLevel to 6", func() {
				Expect(loglvl.NilLevel.Uint8()).To(Equal(uint8(6)))
			})
		})

		Context("with roundtrip conversion", func() {
			It("should maintain value integrity", func() {
				for i := uint8(0); i <= 6; i++ {
					level := loglvl.Level(i)
					Expect(level.Uint8()).To(Equal(i))
				}
			})
		})
	})

	Describe("Logrus", func() {
		Context("with all defined levels", func() {
			It("should convert PanicLevel to logrus.PanicLevel", func() {
				Expect(loglvl.PanicLevel.Logrus()).To(Equal(logrus.PanicLevel))
			})

			It("should convert FatalLevel to logrus.FatalLevel", func() {
				Expect(loglvl.FatalLevel.Logrus()).To(Equal(logrus.FatalLevel))
			})

			It("should convert ErrorLevel to logrus.ErrorLevel", func() {
				Expect(loglvl.ErrorLevel.Logrus()).To(Equal(logrus.ErrorLevel))
			})

			It("should convert WarnLevel to logrus.WarnLevel", func() {
				Expect(loglvl.WarnLevel.Logrus()).To(Equal(logrus.WarnLevel))
			})

			It("should convert InfoLevel to logrus.InfoLevel", func() {
				Expect(loglvl.InfoLevel.Logrus()).To(Equal(logrus.InfoLevel))
			})

			It("should convert DebugLevel to logrus.DebugLevel", func() {
				Expect(loglvl.DebugLevel.Logrus()).To(Equal(logrus.DebugLevel))
			})
		})

		Context("with NilLevel", func() {
			It("should return MaxInt32", func() {
				Expect(loglvl.NilLevel.Logrus()).To(Equal(logrus.Level(math.MaxInt32)))
			})
		})

		Context("with unknown level", func() {
			It("should return MaxInt32", func() {
				unknownLevel := loglvl.Level(99)
				Expect(unknownLevel.Logrus()).To(Equal(logrus.Level(math.MaxInt32)))
			})
		})

		Context("with logrus compatibility", func() {
			It("should be compatible with logrus levels", func() {
				// Ensure our levels map correctly to logrus
				Expect(int(loglvl.PanicLevel.Logrus())).To(Equal(int(logrus.PanicLevel)))
				Expect(int(loglvl.FatalLevel.Logrus())).To(Equal(int(logrus.FatalLevel)))
				Expect(int(loglvl.ErrorLevel.Logrus())).To(Equal(int(logrus.ErrorLevel)))
				Expect(int(loglvl.WarnLevel.Logrus())).To(Equal(int(logrus.WarnLevel)))
				Expect(int(loglvl.InfoLevel.Logrus())).To(Equal(int(logrus.InfoLevel)))
				Expect(int(loglvl.DebugLevel.Logrus())).To(Equal(int(logrus.DebugLevel)))
			})
		})
	})

	Describe("Integration tests", func() {
		Context("with Parse and String roundtrip", func() {
			It("should maintain level identity", func() {
				levels := []loglvl.Level{
					loglvl.PanicLevel,
					loglvl.FatalLevel,
					loglvl.ErrorLevel,
					loglvl.WarnLevel,
					loglvl.InfoLevel,
					loglvl.DebugLevel,
				}

				for _, level := range levels {
					str := level.String()
					parsed := loglvl.Parse(str)
					Expect(parsed).To(Equal(level))
				}
			})
		})

		Context("with all conversion methods", func() {
			It("should provide consistent conversions", func() {
				level := loglvl.InfoLevel

				// String representation
				Expect(level.String()).To(Equal("Info"))

				// Uint8 representation
				Expect(level.Uint8()).To(Equal(uint8(4)))

				// Logrus representation
				Expect(level.Logrus()).To(Equal(logrus.InfoLevel))

				// Parse back
				Expect(loglvl.Parse("info")).To(Equal(level))
			})
		})
	})
})
