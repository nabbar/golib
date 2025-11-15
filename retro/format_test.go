/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package retro_test

import (
	. "github.com/nabbar/golib/retro"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Format", func() {
	Describe("Format constants", func() {
		Context("when checking format values", func() {
			It("should have correct JSON format value", func() {
				Expect(FormatJSON).To(Equal(Format("json")))
			})

			It("should have correct YAML format value", func() {
				Expect(FormatYAML).To(Equal(Format("yaml")))
			})

			It("should have correct TOML format value", func() {
				Expect(FormatTOML).To(Equal(Format("toml")))
			})
		})

		Context("when checking supported formats", func() {
			It("should contain all three supported formats", func() {
				Expect(SupportedFormats).To(HaveLen(3))
				Expect(SupportedFormats).To(ContainElements(FormatJSON, FormatYAML, FormatTOML))
			})
		})
	})

	Describe("Format.Valid", func() {
		Context("when validating supported formats", func() {
			It("should validate JSON format", func() {
				Expect(FormatJSON.Valid()).To(BeTrue())
			})

			It("should validate YAML format", func() {
				Expect(FormatYAML.Valid()).To(BeTrue())
			})

			It("should validate TOML format", func() {
				Expect(FormatTOML.Valid()).To(BeTrue())
			})
		})

		Context("when validating unsupported formats", func() {
			It("should reject XML format", func() {
				format := Format("xml")
				Expect(format.Valid()).To(BeFalse())
			})

			It("should reject empty format", func() {
				format := Format("")
				Expect(format.Valid()).To(BeFalse())
			})

			It("should reject invalid format", func() {
				format := Format("invalid")
				Expect(format.Valid()).To(BeFalse())
			})

			It("should reject format with wrong case", func() {
				format := Format("JSON")
				Expect(format.Valid()).To(BeFalse())
			})
		})
	})
})
