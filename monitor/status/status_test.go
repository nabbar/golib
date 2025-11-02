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

package status_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/status"
)

var _ = Describe("Status Type and Constants", func() {
	Describe("Status constants", func() {
		It("should have correct values", func() {
			Expect(status.KO).To(BeNumerically("==", 0))
			Expect(status.Warn).To(BeNumerically("==", 1))
			Expect(status.OK).To(BeNumerically("==", 2))
		})

		It("should be ordered correctly (KO < Warn < OK)", func() {
			Expect(status.KO).To(BeNumerically("<", status.Warn))
			Expect(status.Warn).To(BeNumerically("<", status.OK))
			Expect(status.KO).To(BeNumerically("<", status.OK))
		})
	})

	Describe("String() method", func() {
		It("should return 'KO' for KO status", func() {
			Expect(status.KO.String()).To(Equal("KO"))
		})

		It("should return 'Warn' for Warn status", func() {
			Expect(status.Warn.String()).To(Equal("Warn"))
		})

		It("should return 'OK' for OK status", func() {
			Expect(status.OK.String()).To(Equal("OK"))
		})

		It("should return 'KO' for unknown status values", func() {
			unknownStatus := status.Status(99)
			Expect(unknownStatus.String()).To(Equal("KO"))
		})

		It("should return 'KO' for status value 3", func() {
			status3 := status.Status(3)
			Expect(status3.String()).To(Equal("KO"))
		})
	})

	Describe("Int() method", func() {
		It("should return 0 for KO status", func() {
			Expect(status.KO.Int()).To(Equal(0))
		})

		It("should return 1 for Warn status", func() {
			Expect(status.Warn.Int()).To(Equal(1))
		})

		It("should return 2 for OK status", func() {
			Expect(status.OK.Int()).To(Equal(2))
		})

		It("should convert back to Status correctly", func() {
			for _, s := range []status.Status{status.KO, status.Warn, status.OK} {
				i := s.Int64()
				converted := status.NewFromInt(i)
				Expect(converted).To(Equal(s))
			}
		})
	})

	Describe("Float() method", func() {
		It("should return 0.0 for KO status", func() {
			Expect(status.KO.Float()).To(Equal(float64(0)))
		})

		It("should return 1.0 for Warn status", func() {
			Expect(status.Warn.Float()).To(Equal(float64(1)))
		})

		It("should return 2.0 for OK status", func() {
			Expect(status.OK.Float()).To(Equal(float64(2)))
		})

		It("should maintain precision", func() {
			Expect(status.KO.Float()).To(BeNumerically("==", 0.0))
			Expect(status.Warn.Float()).To(BeNumerically("==", 1.0))
			Expect(status.OK.Float()).To(BeNumerically("==", 2.0))
		})
	})

	Describe("Type safety", func() {
		It("should be a uint8 type", func() {
			var s status.Status
			Expect(s).To(BeAssignableToTypeOf(status.Status(0)))
		})

		It("should allow comparison", func() {
			Expect(status.OK == status.OK).To(BeTrue())
			Expect(status.KO == status.Warn).To(BeFalse())
			Expect(status.Warn != status.OK).To(BeTrue())
		})
	})
})

var _ = Describe("NewFromString Constructor", func() {
	Context("with valid status strings", func() {
		It("should create OK status from 'ok' (lowercase)", func() {
			s := status.NewFromString("ok")
			Expect(s).To(Equal(status.OK))
		})

		It("should create OK status from 'OK' (uppercase)", func() {
			s := status.NewFromString("OK")
			Expect(s).To(Equal(status.OK))
		})

		It("should create OK status from 'Ok' (mixed case)", func() {
			s := status.NewFromString("Ok")
			Expect(s).To(Equal(status.OK))
		})

		It("should create Warn status from 'warn' (lowercase)", func() {
			s := status.NewFromString("warn")
			Expect(s).To(Equal(status.Warn))
		})

		It("should create Warn status from 'WARN' (uppercase)", func() {
			s := status.NewFromString("WARN")
			Expect(s).To(Equal(status.Warn))
		})

		It("should create Warn status from 'Warn' (mixed case)", func() {
			s := status.NewFromString("Warn")
			Expect(s).To(Equal(status.Warn))
		})

		It("should create KO status from 'ko' (lowercase)", func() {
			s := status.NewFromString("ko")
			Expect(s).To(Equal(status.KO))
		})

		It("should create KO status from 'KO' (uppercase)", func() {
			s := status.NewFromString("KO")
			Expect(s).To(Equal(status.KO))
		})
	})

	Context("with invalid or unknown strings", func() {
		It("should default to KO for empty string", func() {
			s := status.NewFromString("")
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for 'error'", func() {
			s := status.NewFromString("error")
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for 'fail'", func() {
			s := status.NewFromString("fail")
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for 'unknown'", func() {
			s := status.NewFromString("unknown")
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for numeric strings", func() {
			s := status.NewFromString("123")
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for special characters", func() {
			s := status.NewFromString("!@#$%")
			Expect(s).To(Equal(status.KO))
		})
	})

	Context("with whitespace", func() {
		It("should handle strings with leading spaces correctly", func() {
			s := status.NewFromString(" ok")
			Expect(s).To(Equal(status.OK))
		})

		It("should handle strings with trailing spaces correctly", func() {
			s := status.NewFromString("ok ")
			Expect(s).To(Equal(status.OK))
		})

		It("should handle strings with both leading and trailing spaces correctly", func() {
			s := status.NewFromString(" warn ")
			Expect(s).To(Equal(status.Warn))
		})
	})
})

var _ = Describe("NewFromInt Constructor", func() {
	Context("with valid status integers", func() {
		It("should create KO status from 0", func() {
			s := status.NewFromInt(0)
			Expect(s).To(Equal(status.KO))
		})

		It("should create Warn status from 1", func() {
			s := status.NewFromInt(1)
			Expect(s).To(Equal(status.Warn))
		})

		It("should create OK status from 2", func() {
			s := status.NewFromInt(2)
			Expect(s).To(Equal(status.OK))
		})
	})

	Context("with invalid status integers", func() {
		It("should default to KO for 3", func() {
			s := status.NewFromInt(3)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for negative values", func() {
			s := status.NewFromInt(-1)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for large negative values", func() {
			s := status.NewFromInt(-999)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for 255", func() {
			s := status.NewFromInt(255)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for values greater than MaxUint8", func() {
			s := status.NewFromInt(256)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for very large values", func() {
			s := status.NewFromInt(999999)
			Expect(s).To(Equal(status.KO))
		})

		It("should default to KO for int64 max value", func() {
			s := status.NewFromInt(9223372036854775807) // math.MaxInt64
			Expect(s).To(Equal(status.KO))
		})
	})

	Context("with boundary values", func() {
		It("should handle 254 correctly as KO", func() {
			s := status.NewFromInt(254)
			Expect(s).To(Equal(status.KO))
		})

		It("should handle MaxUint8 (255) correctly as KO", func() {
			s := status.NewFromInt(255)
			Expect(s).To(Equal(status.KO))
		})

		It("should handle MaxUint8+1 (256) correctly as KO", func() {
			s := status.NewFromInt(256)
			Expect(s).To(Equal(status.KO))
		})
	})
})

var _ = Describe("Round-trip conversions", func() {
	It("should convert Status -> String -> Status correctly", func() {
		for _, original := range []status.Status{status.KO, status.Warn, status.OK} {
			str := original.String()
			converted := status.NewFromString(str)
			Expect(converted).To(Equal(original))
		}
	})

	It("should convert Status -> Int -> Status correctly", func() {
		for _, original := range []status.Status{status.KO, status.Warn, status.OK} {
			i := original.Int64()
			converted := status.NewFromInt(i)
			Expect(converted).To(Equal(original))
		}
	})

	It("should handle multiple conversions", func() {
		s := status.OK
		Expect(s.String()).To(Equal("OK"))
		Expect(s.Int()).To(Equal(2))
		Expect(s.Float()).To(Equal(float64(2)))

		// Convert back from string
		s2 := status.NewFromString(s.String())
		Expect(s2).To(Equal(s))

		// Convert back from int
		s3 := status.NewFromInt(s.Int64())
		Expect(s3).To(Equal(s))
	})
})
