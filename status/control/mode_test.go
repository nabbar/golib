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

package control_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/status/control"
)

var _ = Describe("Mode", func() {
	Describe("Constants", func() {
		It("should have correct values", func() {
			Expect(control.Ignore).To(Equal(control.Mode(0)))
			Expect(control.Should).To(Equal(control.Mode(1)))
			Expect(control.Must).To(Equal(control.Mode(2)))
			Expect(control.AnyOf).To(Equal(control.Mode(3)))
			Expect(control.Quorum).To(Equal(control.Mode(4)))
		})
	})

	Describe("String", func() {
		It("should return correct string for Should", func() {
			Expect(control.Should.String()).To(Equal("Should"))
		})

		It("should return correct string for Must", func() {
			Expect(control.Must.String()).To(Equal("Must"))
		})

		It("should return correct string for AnyOf", func() {
			Expect(control.AnyOf.String()).To(Equal("AnyOf"))
		})

		It("should return correct string for Quorum", func() {
			Expect(control.Quorum.String()).To(Equal("Quorum"))
		})

		It("should return empty string for Ignore", func() {
			Expect(control.Ignore.String()).To(Equal(""))
		})

		It("should return empty string for unknown mode", func() {
			unknown := control.Mode(99)
			Expect(unknown.String()).To(Equal(""))
		})
	})

	Describe("Code", func() {
		It("should return lowercase code for Should", func() {
			Expect(control.Should.Code()).To(Equal("should"))
		})

		It("should return lowercase code for Must", func() {
			Expect(control.Must.Code()).To(Equal("must"))
		})

		It("should return lowercase code for AnyOf", func() {
			Expect(control.AnyOf.Code()).To(Equal("anyof"))
		})

		It("should return lowercase code for Quorum", func() {
			Expect(control.Quorum.Code()).To(Equal("quorum"))
		})

		It("should return empty string for Ignore", func() {
			Expect(control.Ignore.Code()).To(Equal(""))
		})
	})

	Describe("Parse", func() {
		Context("with valid strings", func() {
			It("should parse 'should'", func() {
				Expect(control.Parse("should")).To(Equal(control.Should))
			})

			It("should parse 'Should'", func() {
				Expect(control.Parse("Should")).To(Equal(control.Should))
			})

			It("should parse 'SHOULD'", func() {
				Expect(control.Parse("SHOULD")).To(Equal(control.Should))
			})

			It("should parse 'must'", func() {
				Expect(control.Parse("must")).To(Equal(control.Must))
			})

			It("should parse 'Must'", func() {
				Expect(control.Parse("Must")).To(Equal(control.Must))
			})

			It("should parse 'anyof'", func() {
				Expect(control.Parse("anyof")).To(Equal(control.AnyOf))
			})

			It("should parse 'AnyOf'", func() {
				Expect(control.Parse("AnyOf")).To(Equal(control.AnyOf))
			})

			It("should parse 'quorum'", func() {
				Expect(control.Parse("quorum")).To(Equal(control.Quorum))
			})

			It("should parse 'Quorum'", func() {
				Expect(control.Parse("Quorum")).To(Equal(control.Quorum))
			})
		})

		Context("with invalid strings", func() {
			It("should return Ignore for empty string", func() {
				Expect(control.Parse("")).To(Equal(control.Ignore))
			})

			It("should return Ignore for unknown string", func() {
				Expect(control.Parse("unknown")).To(Equal(control.Ignore))
			})

			It("should return Ignore for 'ignore'", func() {
				Expect(control.Parse("ignore")).To(Equal(control.Ignore))
			})
		})
	})

	Describe("ParseBytes", func() {
		It("should parse byte slice to Should", func() {
			Expect(control.ParseBytes([]byte("should"))).To(Equal(control.Should))
		})

		It("should parse byte slice to Must", func() {
			Expect(control.ParseBytes([]byte("must"))).To(Equal(control.Must))
		})

		It("should parse byte slice to AnyOf", func() {
			Expect(control.ParseBytes([]byte("anyof"))).To(Equal(control.AnyOf))
		})

		It("should parse byte slice to Quorum", func() {
			Expect(control.ParseBytes([]byte("quorum"))).To(Equal(control.Quorum))
		})

		It("should return Ignore for invalid bytes", func() {
			Expect(control.ParseBytes([]byte("invalid"))).To(Equal(control.Ignore))
		})
	})

	Describe("ParseUint64", func() {
		It("should parse 1 to Should", func() {
			Expect(control.ParseUint64(1)).To(Equal(control.Should))
		})

		It("should parse 2 to Must", func() {
			Expect(control.ParseUint64(2)).To(Equal(control.Must))
		})

		It("should parse 3 to AnyOf", func() {
			Expect(control.ParseUint64(3)).To(Equal(control.AnyOf))
		})

		It("should parse 4 to Quorum", func() {
			Expect(control.ParseUint64(4)).To(Equal(control.Quorum))
		})

		It("should parse 0 to Ignore", func() {
			Expect(control.ParseUint64(0)).To(Equal(control.Ignore))
		})

		It("should parse invalid value to Ignore", func() {
			Expect(control.ParseUint64(99)).To(Equal(control.Ignore))
		})

		It("should handle large values", func() {
			Expect(control.ParseUint64(1000000)).To(Equal(control.Ignore))
		})
	})

	Describe("ParseInt64", func() {
		It("should parse 1 to Should", func() {
			Expect(control.ParseInt64(1)).To(Equal(control.Should))
		})

		It("should parse 2 to Must", func() {
			Expect(control.ParseInt64(2)).To(Equal(control.Must))
		})

		It("should parse 3 to AnyOf", func() {
			Expect(control.ParseInt64(3)).To(Equal(control.AnyOf))
		})

		It("should parse 4 to Quorum", func() {
			Expect(control.ParseInt64(4)).To(Equal(control.Quorum))
		})

		It("should parse 0 to Ignore", func() {
			Expect(control.ParseInt64(0)).To(Equal(control.Ignore))
		})

		It("should parse negative value to Ignore", func() {
			Expect(control.ParseInt64(-1)).To(Equal(control.Ignore))
		})

		It("should parse negative large value to Ignore", func() {
			Expect(control.ParseInt64(-100)).To(Equal(control.Ignore))
		})
	})
})
