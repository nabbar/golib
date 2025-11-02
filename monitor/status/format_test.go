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

var _ = Describe("Status Format Methods", func() {
	Describe("Code() method", func() {
		It("should return 'OK' for OK status", func() {
			Expect(status.OK.Code()).To(Equal("OK"))
		})

		It("should return 'WARN' for Warn status", func() {
			Expect(status.Warn.Code()).To(Equal("WARN"))
		})

		It("should return 'KO' for KO status", func() {
			Expect(status.KO.Code()).To(Equal("KO"))
		})

		It("should return uppercase for all statuses", func() {
			for _, s := range []status.Status{status.KO, status.Warn, status.OK} {
				code := s.Code()
				Expect(code).To(Equal(code))
				// Verify it's uppercase by comparing with String() result
				Expect(code).To(MatchRegexp("^[A-Z]+$"))
			}
		})
	})

	Describe("Int64() method", func() {
		It("should return 0 for KO status", func() {
			Expect(status.KO.Int64()).To(Equal(int64(0)))
		})

		It("should return 1 for Warn status", func() {
			Expect(status.Warn.Int64()).To(Equal(int64(1)))
		})

		It("should return 2 for OK status", func() {
			Expect(status.OK.Int64()).To(Equal(int64(2)))
		})
	})

	Describe("Uint() method", func() {
		It("should return 0 for KO status", func() {
			Expect(status.KO.Uint()).To(Equal(uint(0)))
		})

		It("should return 1 for Warn status", func() {
			Expect(status.Warn.Uint()).To(Equal(uint(1)))
		})

		It("should return 2 for OK status", func() {
			Expect(status.OK.Uint()).To(Equal(uint(2)))
		})
	})

	Describe("Uint64() method", func() {
		It("should return 0 for KO status", func() {
			Expect(status.KO.Uint64()).To(Equal(uint64(0)))
		})

		It("should return 1 for Warn status", func() {
			Expect(status.Warn.Uint64()).To(Equal(uint64(1)))
		})

		It("should return 2 for OK status", func() {
			Expect(status.OK.Uint64()).To(Equal(uint64(2)))
		})
	})

	Describe("Type conversions consistency", func() {
		It("should have consistent values across all numeric types for KO", func() {
			s := status.KO
			Expect(int64(s.Int())).To(Equal(s.Int64()))
			Expect(uint64(s.Int())).To(Equal(s.Uint64()))
			Expect(uint(s.Int())).To(Equal(s.Uint()))
			Expect(int64(s.Float())).To(Equal(s.Int64()))
		})

		It("should have consistent values across all numeric types for Warn", func() {
			s := status.Warn
			Expect(int64(s.Int())).To(Equal(s.Int64()))
			Expect(uint64(s.Int())).To(Equal(s.Uint64()))
			Expect(uint(s.Int())).To(Equal(s.Uint()))
			Expect(int64(s.Float())).To(Equal(s.Int64()))
		})

		It("should have consistent values across all numeric types for OK", func() {
			s := status.OK
			Expect(int64(s.Int())).To(Equal(s.Int64()))
			Expect(uint64(s.Int())).To(Equal(s.Uint64()))
			Expect(uint(s.Int())).To(Equal(s.Uint()))
			Expect(int64(s.Float())).To(Equal(s.Int64()))
		})
	})

	Describe("String representations", func() {
		It("should have String() return mixed case", func() {
			Expect(status.OK.String()).To(Equal("OK"))
			Expect(status.Warn.String()).To(Equal("Warn"))
			Expect(status.KO.String()).To(Equal("KO"))
		})

		It("should have Code() return uppercase", func() {
			Expect(status.OK.Code()).To(Equal("OK"))
			Expect(status.Warn.Code()).To(Equal("WARN"))
			Expect(status.KO.Code()).To(Equal("KO"))
		})

		It("should differentiate between String() and Code() for Warn", func() {
			s := status.Warn
			Expect(s.String()).To(Equal("Warn"))
			Expect(s.Code()).To(Equal("WARN"))
			Expect(s.String()).NotTo(Equal(s.Code()))
		})
	})
})
