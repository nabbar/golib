/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package duration_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

var _ = Describe("Duration Interface", func() {
	Describe("Constructors", func() {
		It("should create duration from nanoseconds", func() {
			// TC-MD-014
			d := libdur.Nanoseconds(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Nanosecond))
		})

		It("should create duration from microseconds", func() {
			// TC-MD-015
			d := libdur.Microseconds(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Microsecond))
		})

		It("should create duration from milliseconds", func() {
			// TC-MD-016
			d := libdur.Milliseconds(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Millisecond))
		})

		It("should create duration from seconds", func() {
			// TC-MD-017
			d := libdur.Seconds(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Second))
		})

		It("should create duration from minutes", func() {
			// TC-MD-018
			d := libdur.Minutes(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Minute))
		})

		It("should create duration from hours", func() {
			// TC-MD-019
			d := libdur.Hours(100)
			Expect(time.Duration(d)).To(Equal(100 * time.Hour))
		})

		It("should create duration from days", func() {
			// TC-MD-020
			d := libdur.Days(100)
			Expect(time.Duration(d)).To(Equal(100 * 24 * time.Hour))
		})
	})

	Describe("ParseDuration", func() {
		It("should parse time.Duration", func() {
			// TC-MD-021
			td := 5 * time.Hour
			d := libdur.ParseDuration(td)
			Expect(time.Duration(d)).To(Equal(td))
		})
	})

	Describe("ParseFloat64", func() {
		It("should parse normal float64", func() {
			// TC-MD-022
			f := 100.5
			d := libdur.ParseFloat64(f)
			Expect(time.Duration(d)).To(Equal(time.Duration(101)))
		})

		It("should parse max float64", func() {
			// TC-MD-023
			f := math.MaxFloat64
			d := libdur.ParseFloat64(f)
			Expect(time.Duration(d)).To(Equal(time.Duration(math.MaxInt64)))
		})

		It("should parse min float64", func() {
			// TC-MD-024
			f := -math.MaxFloat64
			d := libdur.ParseFloat64(f)
			Expect(time.Duration(d)).To(Equal(time.Duration(-math.MaxInt64)))
		})
	})

	Describe("ParseUint32", func() {
		It("should parse uint32", func() {
			// TC-MD-025
			u := uint32(100)
			d := libdur.ParseUint32(u)
			Expect(time.Duration(d)).To(Equal(time.Duration(100)))
		})
		It("should parse large uint32 (simulated)", func() {
			// TC-MD-026
			// Since uint32 can't really overflow math.MaxInt64, this branch is technically unreachable
			// with actual uint32 values, but we test the logic anyway.
			// The only way to reach the branch is if math.MaxInt64 was smaller than math.MaxUint32, which is not true on 64-bit systems.
			// On 32-bit systems, int is 32-bit, so MaxInt64 is still 64-bit, so it's fine.
			// However, the code handles it.
			u := uint32(math.MaxUint32)
			d := libdur.ParseUint32(u)
			Expect(time.Duration(d)).To(Equal(time.Duration(math.MaxUint32)))
		})
	})

	Describe("ParseByte", func() {
		It("should parse valid byte slice", func() {
			// TC-MD-027
			b := []byte("1h30m")
			d, err := libdur.ParseByte(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(time.Duration(d)).To(Equal(90 * time.Minute))
		})

		It("should return error for invalid byte slice", func() {
			// TC-MD-028
			b := []byte("invalid")
			_, err := libdur.ParseByte(b)
			Expect(err).To(HaveOccurred())
		})
	})
})
