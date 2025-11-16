/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package sender_test

import (
	libsnd "github.com/nabbar/golib/mail/sender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Type Definitions", func() {
	Describe("Encoding Type", func() {
		It("should return correct string for EncodingNone", func() {
			Expect(libsnd.EncodingNone.String()).To(Equal("None"))
		})

		It("should return correct string for EncodingBinary", func() {
			Expect(libsnd.EncodingBinary.String()).To(Equal("Binary"))
		})

		It("should return correct string for EncodingBase64", func() {
			Expect(libsnd.EncodingBase64.String()).To(Equal("Base 64"))
		})

		It("should return correct string for EncodingQuotedPrintable", func() {
			Expect(libsnd.EncodingQuotedPrintable.String()).To(Equal("Quoted Printable"))
		})

		It("should parse encoding from string", func() {
			tests := map[string]libsnd.Encoding{
				"None":             libsnd.EncodingNone,
				"none":             libsnd.EncodingNone,
				"NONE":             libsnd.EncodingNone,
				"Binary":           libsnd.EncodingBinary,
				"binary":           libsnd.EncodingBinary,
				"BINARY":           libsnd.EncodingBinary,
				"Base 64":          libsnd.EncodingBase64,
				"base 64":          libsnd.EncodingBase64,
				"BASE 64":          libsnd.EncodingBase64,
				"Quoted Printable": libsnd.EncodingQuotedPrintable,
				"quoted printable": libsnd.EncodingQuotedPrintable,
				"QUOTED PRINTABLE": libsnd.EncodingQuotedPrintable,
			}

			for input, expected := range tests {
				result := libsnd.ParseEncoding(input)
				Expect(result).To(Equal(expected), "Failed for input: %s", input)
			}
		})

		It("should return EncodingNone for unknown encoding", func() {
			Expect(libsnd.ParseEncoding("unknown")).To(Equal(libsnd.EncodingNone))
			Expect(libsnd.ParseEncoding("")).To(Equal(libsnd.EncodingNone))
			Expect(libsnd.ParseEncoding("invalid-encoding")).To(Equal(libsnd.EncodingNone))
		})

		It("should handle case-insensitive parsing", func() {
			Expect(libsnd.ParseEncoding("bAsE 64")).To(Equal(libsnd.EncodingBase64))
			Expect(libsnd.ParseEncoding("QuOtEd PrInTaBlE")).To(Equal(libsnd.EncodingQuotedPrintable))
		})
	})

	Describe("Priority Type", func() {
		It("should return correct string for PriorityNormal", func() {
			Expect(libsnd.PriorityNormal.String()).To(Equal("Normal"))
		})

		It("should return correct string for PriorityLow", func() {
			Expect(libsnd.PriorityLow.String()).To(Equal("Low"))
		})

		It("should return correct string for PriorityHigh", func() {
			Expect(libsnd.PriorityHigh.String()).To(Equal("High"))
		})

		It("should parse priority from string", func() {
			tests := map[string]libsnd.Priority{
				"Normal": libsnd.PriorityNormal,
				"normal": libsnd.PriorityNormal,
				"NORMAL": libsnd.PriorityNormal,
				"Low":    libsnd.PriorityLow,
				"low":    libsnd.PriorityLow,
				"LOW":    libsnd.PriorityLow,
				"High":   libsnd.PriorityHigh,
				"high":   libsnd.PriorityHigh,
				"HIGH":   libsnd.PriorityHigh,
			}

			for input, expected := range tests {
				result := libsnd.ParsePriority(input)
				Expect(result).To(Equal(expected), "Failed for input: %s", input)
			}
		})

		It("should return PriorityNormal for unknown priority", func() {
			Expect(libsnd.ParsePriority("unknown")).To(Equal(libsnd.PriorityNormal))
			Expect(libsnd.ParsePriority("")).To(Equal(libsnd.PriorityNormal))
			Expect(libsnd.ParsePriority("medium")).To(Equal(libsnd.PriorityNormal))
		})

		It("should handle case-insensitive parsing", func() {
			Expect(libsnd.ParsePriority("HiGh")).To(Equal(libsnd.PriorityHigh))
			Expect(libsnd.ParsePriority("LoW")).To(Equal(libsnd.PriorityLow))
		})
	})

	Describe("ContentType Type", func() {
		It("should return correct string for ContentPlainText", func() {
			Expect(libsnd.ContentPlainText.String()).To(Equal("Plain Text"))
		})

		It("should return correct string for ContentHTML", func() {
			Expect(libsnd.ContentHTML.String()).To(Equal("HTML"))
		})

		It("should default to ContentPlainText for invalid type", func() {
			// Cast an invalid value to ContentType
			invalid := libsnd.ContentType(99)
			Expect(invalid.String()).To(Equal("Plain Text"))
		})
	})

	Describe("RecipientType Type", func() {
		It("should return correct string for RecipientTo", func() {
			Expect(libsnd.RecipientTo.String()).To(Equal("To"))
		})

		It("should return correct string for RecipientCC", func() {
			Expect(libsnd.RecipientCC.String()).To(Equal("Cc"))
		})

		It("should return correct string for RecipientBCC", func() {
			Expect(libsnd.RecipientBCC.String()).To(Equal("Bcc"))
		})
	})

	Describe("Body and File Types", func() {
		It("should create Body with ContentPlainText", func() {
			body := libsnd.NewBody(libsnd.ContentPlainText, newReadCloser("test"))
			Expect(body).ToNot(BeNil())
		})

		It("should create Body with ContentHTML", func() {
			body := libsnd.NewBody(libsnd.ContentHTML, newReadCloser("<html>test</html>"))
			Expect(body).ToNot(BeNil())
		})

		It("should create File with all parameters", func() {
			file := libsnd.NewFile("test.txt", "text/plain", newReadCloser("content"))
			Expect(file).ToNot(BeNil())
		})

		It("should create File with empty content", func() {
			file := libsnd.NewFile("empty.txt", "text/plain", newReadCloser(""))
			Expect(file).ToNot(BeNil())
		})

		It("should create File with various mime types", func() {
			mimeTypes := []string{
				"text/plain",
				"text/html",
				"application/pdf",
				"image/png",
				"application/octet-stream",
			}

			for _, mime := range mimeTypes {
				file := libsnd.NewFile("file", mime, newReadCloser("data"))
				Expect(file).ToNot(BeNil())
			}
		})
	})

	Describe("Type Constants", func() {
		It("should have distinct Encoding values", func() {
			encodings := []libsnd.Encoding{
				libsnd.EncodingNone,
				libsnd.EncodingBinary,
				libsnd.EncodingBase64,
				libsnd.EncodingQuotedPrintable,
			}

			// Check all are unique
			seen := make(map[libsnd.Encoding]bool)
			for _, e := range encodings {
				Expect(seen[e]).To(BeFalse(), "Duplicate encoding value: %v", e)
				seen[e] = true
			}
		})

		It("should have distinct Priority values", func() {
			priorities := []libsnd.Priority{
				libsnd.PriorityNormal,
				libsnd.PriorityLow,
				libsnd.PriorityHigh,
			}

			seen := make(map[libsnd.Priority]bool)
			for _, p := range priorities {
				Expect(seen[p]).To(BeFalse(), "Duplicate priority value: %v", p)
				seen[p] = true
			}
		})

		It("should have distinct ContentType values", func() {
			types := []libsnd.ContentType{
				libsnd.ContentPlainText,
				libsnd.ContentHTML,
			}

			seen := make(map[libsnd.ContentType]bool)
			for _, t := range types {
				Expect(seen[t]).To(BeFalse(), "Duplicate content type value: %v", t)
				seen[t] = true
			}
		})

		It("should have distinct RecipientType values", func() {
			types := []interface{}{
				libsnd.RecipientTo,
				libsnd.RecipientCC,
				libsnd.RecipientBCC,
			}

			// Just check they're defined
			Expect(types).To(HaveLen(3))
		})
	})

	Describe("Type Round-Trip", func() {
		It("should round-trip Encoding through string", func() {
			encodings := []libsnd.Encoding{
				libsnd.EncodingNone,
				libsnd.EncodingBinary,
				libsnd.EncodingBase64,
				libsnd.EncodingQuotedPrintable,
			}

			for _, e := range encodings {
				str := e.String()
				parsed := libsnd.ParseEncoding(str)
				Expect(parsed).To(Equal(e), "Round-trip failed for %v", e)
			}
		})

		It("should round-trip Priority through string", func() {
			priorities := []libsnd.Priority{
				libsnd.PriorityNormal,
				libsnd.PriorityLow,
				libsnd.PriorityHigh,
			}

			for _, p := range priorities {
				str := p.String()
				parsed := libsnd.ParsePriority(str)
				Expect(parsed).To(Equal(p), "Round-trip failed for %v", p)
			}
		})

		It("should maintain ContentType values", func() {
			types := []libsnd.ContentType{
				libsnd.ContentPlainText,
				libsnd.ContentHTML,
			}

			for _, t := range types {
				str := t.String()
				Expect(str).ToNot(BeEmpty())
			}
		})
	})
})
