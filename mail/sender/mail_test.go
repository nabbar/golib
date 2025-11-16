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
	"time"

	libsnd "github.com/nabbar/golib/mail/sender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mail Operations", func() {
	var mail libsnd.Mail

	BeforeEach(func() {
		mail = newMail()
	})

	Describe("Mail Creation", func() {
		It("should create a new mail instance", func() {
			Expect(mail).ToNot(BeNil())
		})

		It("should have default charset UTF-8", func() {
			Expect(mail.GetCharset()).To(Equal("UTF-8"))
		})

		It("should have default encoding EncodingNone", func() {
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingNone))
		})

		It("should have an Email interface", func() {
			Expect(mail.Email()).ToNot(BeNil())
		})
	})

	Describe("Charset Operations", func() {
		It("should set and get charset", func() {
			mail.SetCharset("ISO-8859-1")
			Expect(mail.GetCharset()).To(Equal("ISO-8859-1"))
		})

		It("should handle empty charset", func() {
			mail.SetCharset("")
			Expect(mail.GetCharset()).To(Equal(""))
		})

		It("should handle unicode charset", func() {
			mail.SetCharset("UTF-16")
			Expect(mail.GetCharset()).To(Equal("UTF-16"))
		})
	})

	Describe("Subject Operations", func() {
		It("should set and get subject", func() {
			mail.SetSubject("Test Subject")
			Expect(mail.GetSubject()).To(Equal("Test Subject"))
		})

		It("should handle empty subject", func() {
			mail.SetSubject("")
			Expect(mail.GetSubject()).To(Equal(""))
		})

		It("should handle subject with special characters", func() {
			subject := "Tëst Sübject with Émojis 🎉"
			mail.SetSubject(subject)
			Expect(mail.GetSubject()).To(Equal(subject))
		})

		It("should handle long subject", func() {
			subject := string(make([]byte, 1000))
			mail.SetSubject(subject)
			Expect(mail.GetSubject()).To(Equal(subject))
		})
	})

	Describe("Priority Operations", func() {
		It("should set and get normal priority", func() {
			mail.SetPriority(libsnd.PriorityNormal)
			Expect(mail.GetPriority()).To(Equal(libsnd.PriorityNormal))
		})

		It("should set and get low priority", func() {
			mail.SetPriority(libsnd.PriorityLow)
			Expect(mail.GetPriority()).To(Equal(libsnd.PriorityLow))
		})

		It("should set and get high priority", func() {
			mail.SetPriority(libsnd.PriorityHigh)
			Expect(mail.GetPriority()).To(Equal(libsnd.PriorityHigh))
		})
	})

	Describe("Encoding Operations", func() {
		It("should set and get encoding none", func() {
			mail.SetEncoding(libsnd.EncodingNone)
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingNone))
		})

		It("should set and get encoding binary", func() {
			mail.SetEncoding(libsnd.EncodingBinary)
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingBinary))
		})

		It("should set and get encoding base64", func() {
			mail.SetEncoding(libsnd.EncodingBase64)
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingBase64))
		})

		It("should set and get encoding quoted-printable", func() {
			mail.SetEncoding(libsnd.EncodingQuotedPrintable)
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingQuotedPrintable))
		})
	})

	Describe("DateTime Operations", func() {
		It("should set and get datetime", func() {
			now := time.Now()
			mail.SetDateTime(now)
			Expect(mail.GetDateTime()).To(BeTemporally("~", now, time.Second))
		})

		It("should format datetime string", func() {
			now := time.Now()
			mail.SetDateTime(now)
			dateStr := mail.GetDateString()
			Expect(dateStr).ToNot(BeEmpty())
		})

		It("should parse datetime from string", func() {
			dateStr := "Mon, 02 Jan 2006 15:04:05 -0700"
			err := mail.SetDateString(time.RFC1123Z, dateStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.GetDateString()).To(Equal(dateStr))
		})

		It("should return error for invalid datetime string", func() {
			err := mail.SetDateString(time.RFC1123Z, "invalid-date")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Header Operations", func() {
		It("should add custom header", func() {
			mail.AddHeader("X-Custom-Header", "custom-value")
			headers := mail.GetHeaders()
			Expect(headers.Get("X-Custom-Header")).To(Equal("custom-value"))
		})

		It("should add multiple values to same header", func() {
			mail.AddHeader("X-Custom", "value1")
			mail.AddHeader("X-Custom", "value2")
			values := mail.GetHeader("X-Custom")
			Expect(values).To(HaveLen(2))
			Expect(values).To(ContainElements("value1", "value2"))
		})

		It("should skip empty header values", func() {
			mail.AddHeader("X-Test", "")
			values := mail.GetHeader("X-Test")
			Expect(values).To(BeEmpty())
		})

		It("should get all headers", func() {
			mail.SetSubject("Test")
			mail.SetDateTime(time.Now())
			headers := mail.GetHeaders()
			Expect(headers).ToNot(BeNil())
			Expect(headers.Get("Subject")).To(Equal("Test"))
			Expect(headers.Get("MIME-Version")).To(Equal("1.0"))
		})
	})

	Describe("Body Operations", func() {
		It("should set plain text body", func() {
			body := newReadCloser("Test plain text body")
			mail.SetBody(libsnd.ContentPlainText, body)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should set HTML body", func() {
			body := newReadCloser("<html><body>Test HTML</body></html>")
			mail.SetBody(libsnd.ContentHTML, body)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should replace body with SetBody", func() {
			body1 := newReadCloser("First body")
			body2 := newReadCloser("Second body")
			mail.SetBody(libsnd.ContentPlainText, body1)
			mail.SetBody(libsnd.ContentPlainText, body2)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should add alternative body", func() {
			plainBody := newReadCloser("Plain text")
			htmlBody := newReadCloser("<html>HTML</html>")
			mail.SetBody(libsnd.ContentPlainText, plainBody)
			mail.AddBody(libsnd.ContentHTML, htmlBody)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(2))
		})

		It("should replace existing body type with AddBody", func() {
			body1 := newReadCloser("First plain body")
			body2 := newReadCloser("Second plain body")
			mail.AddBody(libsnd.ContentPlainText, body1)
			mail.AddBody(libsnd.ContentPlainText, body2)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})
	})

	Describe("Attachment Operations", func() {
		It("should add regular attachment", func() {
			data := newReadCloser("attachment data")
			mail.AddAttachment("test.txt", "text/plain", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should add inline attachment", func() {
			data := newReadCloser("inline data")
			mail.AddAttachment("image.png", "image/png", data, true)
			inlines := mail.GetAttachment(true)
			Expect(inlines).To(HaveLen(1))
		})

		It("should set attachment (replace all)", func() {
			data1 := newReadCloser("data1")
			data2 := newReadCloser("data2")
			mail.AddAttachment("file1.txt", "text/plain", data1, false)
			mail.SetAttachment("file2.txt", "text/plain", data2, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should replace attachment with same name", func() {
			data1 := newReadCloser("old data")
			data2 := newReadCloser("new data")
			mail.AddAttachment("test.txt", "text/plain", data1, false)
			mail.AddAttachment("test.txt", "text/plain", data2, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should attach file by path", func() {
			data := newReadCloser("file content")
			mail.AttachFile("/path/to/document.pdf", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should detect mime type from file extension", func() {
			data := newReadCloser("image data")
			mail.AttachFile("/path/to/image.jpg", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should separate inline and regular attachments", func() {
			regularData := newReadCloser("regular")
			inlineData := newReadCloser("inline")
			mail.AddAttachment("file.txt", "text/plain", regularData, false)
			mail.AddAttachment("image.png", "image/png", inlineData, true)
			Expect(mail.GetAttachment(false)).To(HaveLen(1))
			Expect(mail.GetAttachment(true)).To(HaveLen(1))
		})
	})

	Describe("Clone Operations", func() {
		It("should clone mail instance", func() {
			mail.SetSubject("Original")
			mail.SetCharset("UTF-8")
			mail.Email().SetFrom("original@example.com")

			cloned := mail.Clone()
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.GetSubject()).To(Equal("Original"))
			Expect(cloned.GetCharset()).To(Equal("UTF-8"))
			Expect(cloned.Email().GetFrom()).To(Equal("original@example.com"))
		})

		It("should create independent clone", func() {
			mail.SetSubject("Original")
			cloned := mail.Clone()

			// Modify original
			mail.SetSubject("Modified")

			// Clone should not be affected
			Expect(cloned.GetSubject()).To(Equal("Original"))
		})

		It("should clone all mail properties", func() {
			now := time.Now()
			mail.SetDateTime(now)
			mail.SetSubject("Test")
			mail.SetCharset("ISO-8859-1")
			mail.SetEncoding(libsnd.EncodingBase64)
			mail.SetPriority(libsnd.PriorityHigh)
			mail.AddHeader("X-Custom", "value")

			cloned := mail.Clone()
			Expect(cloned.GetDateTime()).To(BeTemporally("~", now, time.Second))
			Expect(cloned.GetSubject()).To(Equal("Test"))
			Expect(cloned.GetCharset()).To(Equal("ISO-8859-1"))
			Expect(cloned.GetEncoding()).To(Equal(libsnd.EncodingBase64))
			Expect(cloned.GetPriority()).To(Equal(libsnd.PriorityHigh))
		})
	})
})
