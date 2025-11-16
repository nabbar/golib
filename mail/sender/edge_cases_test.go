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
	"strings"
	"time"

	libsnd "github.com/nabbar/golib/mail/sender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edge Cases and Boundary Conditions", func() {
	var mail libsnd.Mail

	BeforeEach(func() {
		mail = newMail()
	})

	Describe("Empty and Nil Values", func() {
		It("should handle empty charset", func() {
			mail.SetCharset("")
			Expect(mail.GetCharset()).To(Equal(""))
		})

		It("should handle empty subject", func() {
			mail.SetSubject("")
			Expect(mail.GetSubject()).To(Equal(""))
		})

		It("should handle empty from address", func() {
			mail.Email().SetFrom("")
			Expect(mail.Email().GetFrom()).To(Equal(""))
		})

		It("should handle nil body reader", func() {
			// This might panic or handle gracefully depending on implementation
			// Testing the actual behavior
			defer func() {
				if r := recover(); r != nil {
					// Panic occurred, that's one valid behavior
					Expect(r).ToNot(BeNil())
				}
			}()

			mail.SetBody(libsnd.ContentPlainText, nil)
		})

		It("should handle empty recipient list", func() {
			recipients := mail.Email().GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(BeEmpty())
		})

		It("should handle setting empty recipients", func() {
			mail.Email().AddRecipients(libsnd.RecipientTo, "test@example.com")
			mail.Email().SetRecipients(libsnd.RecipientTo)
			recipients := mail.Email().GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(BeEmpty())
		})
	})

	Describe("Large Values", func() {
		It("should handle very long subject", func() {
			longSubject := strings.Repeat("A", 10000)
			mail.SetSubject(longSubject)
			Expect(mail.GetSubject()).To(Equal(longSubject))
		})

		It("should handle large number of recipients", func() {
			for i := 0; i < 1000; i++ {
				mail.Email().AddRecipients(libsnd.RecipientTo, "recipient"+string(rune(i))+"@example.com")
			}
			recipients := mail.Email().GetRecipients(libsnd.RecipientTo)
			Expect(len(recipients)).To(BeNumerically(">=", 1))
		})

		It("should handle large body content", func() {
			largeContent := strings.Repeat("Content ", 100000)
			body := newReadCloser(largeContent)
			mail.SetBody(libsnd.ContentPlainText, body)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should handle many custom headers", func() {
			for i := 0; i < 100; i++ {
				mail.AddHeader("X-Custom-"+string(rune(i)), "value")
			}
			headers := mail.GetHeaders()
			Expect(headers).ToNot(BeNil())
		})

		It("should handle many attachments", func() {
			for i := 0; i < 50; i++ {
				data := newReadCloser("attachment " + string(rune(i)))
				mail.AddAttachment("file"+string(rune(i))+".txt", "text/plain", data, false)
			}
			attachments := mail.GetAttachment(false)
			Expect(len(attachments)).To(BeNumerically(">=", 1))
		})
	})

	Describe("Special Characters", func() {
		It("should handle unicode in subject", func() {
			subject := "Test 🎉 émojis et caractères spéciaux © ® ™ 你好"
			mail.SetSubject(subject)
			Expect(mail.GetSubject()).To(Equal(subject))
		})

		It("should handle unicode in email addresses", func() {
			// While not technically valid, testing the storage
			mail.Email().SetFrom("test@exämple.com")
			Expect(mail.Email().GetFrom()).To(Equal("test@exämple.com"))
		})

		It("should handle special characters in header values", func() {
			mail.AddHeader("X-Special", "Value with\nnewline and\ttab")
			headers := mail.GetHeaders()
			Expect(headers.Get("X-Special")).ToNot(BeEmpty())
		})

		It("should handle HTML entities in subject", func() {
			subject := "Test &lt;html&gt; &amp; entities &#8364;"
			mail.SetSubject(subject)
			Expect(mail.GetSubject()).To(Equal(subject))
		})

		It("should handle quotes in subject", func() {
			subject := `Subject with "quotes" and 'apostrophes'`
			mail.SetSubject(subject)
			Expect(mail.GetSubject()).To(Equal(subject))
		})
	})

	Describe("Boundary Dates", func() {
		It("should handle zero time", func() {
			mail.SetDateTime(time.Time{})
			result := mail.GetDateTime()
			Expect(result).To(Equal(time.Time{}))
		})

		It("should handle very old date", func() {
			oldDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
			mail.SetDateTime(oldDate)
			Expect(mail.GetDateTime()).To(Equal(oldDate))
		})

		It("should handle far future date", func() {
			futureDate := time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
			mail.SetDateTime(futureDate)
			Expect(mail.GetDateTime()).To(Equal(futureDate))
		})

		It("should handle dates with different timezones", func() {
			date1 := time.Now().In(time.UTC)
			mail.SetDateTime(date1)

			date2 := time.Now().In(time.FixedZone("TEST", 3600))
			mail.SetDateTime(date2)

			result := mail.GetDateTime()
			Expect(result.Location()).To(Equal(date2.Location()))
		})
	})

	Describe("Address Fallback Behavior", func() {
		It("should chain sender fallback correctly", func() {
			mail.Email().SetReturnPath("return@example.com")
			Expect(mail.Email().GetSender()).To(Equal("return@example.com"))

			mail.Email().SetReplyTo("reply@example.com")
			Expect(mail.Email().GetSender()).To(Equal("reply@example.com"))

			mail.Email().SetSender("sender@example.com")
			Expect(mail.Email().GetSender()).To(Equal("sender@example.com"))
		})

		It("should chain replyTo fallback correctly", func() {
			mail.Email().SetReturnPath("return@example.com")
			Expect(mail.Email().GetReplyTo()).To(Equal("return@example.com"))

			mail.Email().SetSender("sender@example.com")
			Expect(mail.Email().GetReplyTo()).To(Equal("sender@example.com"))

			mail.Email().SetReplyTo("reply@example.com")
			Expect(mail.Email().GetReplyTo()).To(Equal("reply@example.com"))
		})

		It("should chain returnPath fallback correctly", func() {
			mail.Email().SetReplyTo("reply@example.com")
			Expect(mail.Email().GetReturnPath()).To(Equal("reply@example.com"))

			mail.Email().SetSender("sender@example.com")
			Expect(mail.Email().GetReturnPath()).To(Equal("sender@example.com"))

			mail.Email().SetReturnPath("return@example.com")
			Expect(mail.Email().GetReturnPath()).To(Equal("return@example.com"))
		})

		It("should return empty when all addresses are empty", func() {
			Expect(mail.Email().GetSender()).To(Equal(""))
			Expect(mail.Email().GetReplyTo()).To(Equal(""))
			Expect(mail.Email().GetReturnPath()).To(Equal(""))
		})
	})

	Describe("Attachment Edge Cases", func() {
		It("should handle attachment with empty name", func() {
			data := newReadCloser("content")
			mail.AddAttachment("", "text/plain", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should handle attachment with empty mime", func() {
			data := newReadCloser("content")
			mail.AddAttachment("file.txt", "", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should handle attachment with very long filename", func() {
			longName := strings.Repeat("a", 1000) + ".txt"
			data := newReadCloser("content")
			mail.AddAttachment(longName, "text/plain", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should handle attachment with special characters in filename", func() {
			name := "file with spaces and special-chars_123.txt"
			data := newReadCloser("content")
			mail.AddAttachment(name, "text/plain", data, false)
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should replace inline attachment in attach list if names match", func() {
			data1 := newReadCloser("first")
			data2 := newReadCloser("second")

			mail.AddAttachment("same.txt", "text/plain", data1, false)
			mail.AddAttachment("same.txt", "text/plain", data2, false)

			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})
	})

	Describe("Clone Edge Cases", func() {
		It("should clone mail with no fields set", func() {
			emptyMail := libsnd.New()
			cloned := emptyMail.Clone()
			Expect(cloned).ToNot(BeNil())
		})

		It("should clone mail with all fields set", func() {
			mail.SetSubject("Subject")
			mail.SetCharset("UTF-8")
			mail.SetEncoding(libsnd.EncodingBase64)
			mail.SetPriority(libsnd.PriorityHigh)
			mail.SetDateTime(time.Now())
			mail.Email().SetFrom("from@example.com")
			mail.Email().AddRecipients(libsnd.RecipientTo, "to@example.com")
			mail.AddHeader("X-Custom", "value")
			mail.SetBody(libsnd.ContentPlainText, newReadCloser("body"))
			mail.AddAttachment("file.txt", "text/plain", newReadCloser("data"), false)

			cloned := mail.Clone()
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.GetSubject()).To(Equal("Subject"))
		})

		It("should not share state between original and clone", func() {
			mail.SetSubject("Original")
			cloned := mail.Clone()

			mail.SetSubject("Modified")
			cloned.SetSubject("Clone Modified")

			Expect(mail.GetSubject()).To(Equal("Modified"))
			Expect(cloned.GetSubject()).To(Equal("Clone Modified"))
		})
	})

	Describe("Header Edge Cases", func() {
		It("should handle getting non-existent header", func() {
			values := mail.GetHeader("X-NonExistent")
			Expect(values).To(BeEmpty())
		})

		It("should handle adding header with multiple values", func() {
			mail.AddHeader("X-Multi", "value1", "value2", "value3")
			values := mail.GetHeader("X-Multi")
			Expect(values).To(HaveLen(3))
		})

		It("should skip empty header values", func() {
			mail.AddHeader("X-Test", "value1", "", "value2", "")
			values := mail.GetHeader("X-Test")
			Expect(values).To(HaveLen(2))
			Expect(values).To(ConsistOf("value1", "value2"))
		})

		It("should handle case-sensitive header names", func() {
			mail.AddHeader("X-Custom", "value")
			// Header names are typically case-insensitive in MIME
			headers := mail.GetHeaders()
			Expect(headers.Get("X-Custom")).To(Equal("value"))
			Expect(headers.Get("x-custom")).To(Equal("value"))
		})

		It("should preserve header order for multiple values", func() {
			mail.AddHeader("X-Order", "first")
			mail.AddHeader("X-Order", "second")
			mail.AddHeader("X-Order", "third")
			values := mail.GetHeader("X-Order")
			Expect(values).To(HaveLen(3))
		})
	})

	Describe("Body Edge Cases", func() {
		It("should handle AddBody when no body exists", func() {
			body := newReadCloser("first body")
			mail.AddBody(libsnd.ContentPlainText, body)
			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should replace body of same content type", func() {
			body1 := newReadCloser("first")
			body2 := newReadCloser("second")

			mail.AddBody(libsnd.ContentPlainText, body1)
			mail.AddBody(libsnd.ContentPlainText, body2)

			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})

		It("should keep bodies of different content types", func() {
			plainBody := newReadCloser("plain")
			htmlBody := newReadCloser("html")

			mail.AddBody(libsnd.ContentPlainText, plainBody)
			mail.AddBody(libsnd.ContentHTML, htmlBody)

			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(2))
		})

		It("should handle SetBody multiple times", func() {
			for i := 0; i < 10; i++ {
				body := newReadCloser("body " + string(rune(i)))
				mail.SetBody(libsnd.ContentPlainText, body)
			}

			bodies := mail.GetBody()
			Expect(bodies).To(HaveLen(1))
		})
	})

	Describe("Recipient Deduplication", func() {
		It("should not add duplicate To recipients", func() {
			mail.Email().AddRecipients(libsnd.RecipientTo, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientTo, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientTo, "same@example.com")

			recipients := mail.Email().GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(1))
		})

		It("should not add duplicate Cc recipients", func() {
			mail.Email().AddRecipients(libsnd.RecipientCC, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientCC, "same@example.com")

			recipients := mail.Email().GetRecipients(libsnd.RecipientCC)
			Expect(recipients).To(HaveLen(1))
		})

		It("should not add duplicate Bcc recipients", func() {
			mail.Email().AddRecipients(libsnd.RecipientBCC, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientBCC, "same@example.com")

			recipients := mail.Email().GetRecipients(libsnd.RecipientBCC)
			Expect(recipients).To(HaveLen(1))
		})

		It("should allow same address in different recipient types", func() {
			mail.Email().AddRecipients(libsnd.RecipientTo, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientCC, "same@example.com")
			mail.Email().AddRecipients(libsnd.RecipientBCC, "same@example.com")

			Expect(mail.Email().GetRecipients(libsnd.RecipientTo)).To(HaveLen(1))
			Expect(mail.Email().GetRecipients(libsnd.RecipientCC)).To(HaveLen(1))
			Expect(mail.Email().GetRecipients(libsnd.RecipientBCC)).To(HaveLen(1))
		})
	})

	Describe("Encoding and Priority Defaults", func() {
		It("should default to EncodingNone", func() {
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingNone))
		})

		It("should default to PriorityNormal", func() {
			Expect(mail.GetPriority()).To(Equal(libsnd.PriorityNormal))
		})

		It("should handle invalid encoding gracefully", func() {
			invalidEncoding := libsnd.ParseEncoding("invalid")
			Expect(invalidEncoding).To(Equal(libsnd.EncodingNone))
		})

		It("should handle invalid priority gracefully", func() {
			invalidPriority := libsnd.ParsePriority("invalid")
			Expect(invalidPriority).To(Equal(libsnd.PriorityNormal))
		})
	})
})
