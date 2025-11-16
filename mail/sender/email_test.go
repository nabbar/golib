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

var _ = Describe("Email Address Operations", func() {
	var (
		mail  libsnd.Mail
		email libsnd.Email
	)

	BeforeEach(func() {
		mail = newMail()
		email = mail.Email()
	})

	Describe("From Address", func() {
		It("should set and get from address", func() {
			email.SetFrom("sender@example.com")
			Expect(email.GetFrom()).To(Equal("sender@example.com"))
		})

		It("should handle empty from address", func() {
			email.SetFrom("")
			Expect(email.GetFrom()).To(Equal(""))
		})

		It("should override previous from address", func() {
			email.SetFrom("first@example.com")
			email.SetFrom("second@example.com")
			Expect(email.GetFrom()).To(Equal("second@example.com"))
		})
	})

	Describe("Sender Address", func() {
		It("should set and get sender address", func() {
			email.SetSender("sender@example.com")
			Expect(email.GetSender()).To(Equal("sender@example.com"))
		})

		It("should fallback to replyTo when sender not set", func() {
			email.SetReplyTo("reply@example.com")
			Expect(email.GetSender()).To(Equal("reply@example.com"))
		})

		It("should fallback to returnPath when sender and replyTo not set", func() {
			email.SetReturnPath("return@example.com")
			Expect(email.GetSender()).To(Equal("return@example.com"))
		})

		It("should prefer sender over other addresses", func() {
			email.SetSender("sender@example.com")
			email.SetReplyTo("reply@example.com")
			email.SetReturnPath("return@example.com")
			Expect(email.GetSender()).To(Equal("sender@example.com"))
		})
	})

	Describe("ReplyTo Address", func() {
		It("should set and get replyTo address", func() {
			email.SetReplyTo("reply@example.com")
			Expect(email.GetReplyTo()).To(Equal("reply@example.com"))
		})

		It("should fallback to sender when replyTo not set", func() {
			email.SetSender("sender@example.com")
			Expect(email.GetReplyTo()).To(Equal("sender@example.com"))
		})

		It("should fallback to returnPath when replyTo and sender not set", func() {
			email.SetReturnPath("return@example.com")
			Expect(email.GetReplyTo()).To(Equal("return@example.com"))
		})

		It("should prefer replyTo over other addresses", func() {
			email.SetReplyTo("reply@example.com")
			email.SetSender("sender@example.com")
			email.SetReturnPath("return@example.com")
			Expect(email.GetReplyTo()).To(Equal("reply@example.com"))
		})
	})

	Describe("ReturnPath Address", func() {
		It("should set and get returnPath address", func() {
			email.SetReturnPath("return@example.com")
			Expect(email.GetReturnPath()).To(Equal("return@example.com"))
		})

		It("should fallback to sender when returnPath not set", func() {
			email.SetSender("sender@example.com")
			Expect(email.GetReturnPath()).To(Equal("sender@example.com"))
		})

		It("should fallback to replyTo when returnPath and sender not set", func() {
			email.SetReplyTo("reply@example.com")
			Expect(email.GetReturnPath()).To(Equal("reply@example.com"))
		})

		It("should prefer returnPath over other addresses", func() {
			email.SetReturnPath("return@example.com")
			email.SetSender("sender@example.com")
			email.SetReplyTo("reply@example.com")
			Expect(email.GetReturnPath()).To(Equal("return@example.com"))
		})
	})

	Describe("To Recipients", func() {
		It("should add To recipients", func() {
			email.AddRecipients(libsnd.RecipientTo, "to1@example.com", "to2@example.com")
			recipients := email.GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(2))
			Expect(recipients).To(ContainElements("to1@example.com", "to2@example.com"))
		})

		It("should set To recipients (replace all)", func() {
			email.AddRecipients(libsnd.RecipientTo, "old1@example.com", "old2@example.com")
			email.SetRecipients(libsnd.RecipientTo, "new1@example.com")
			recipients := email.GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(1))
			Expect(recipients).To(ContainElement("new1@example.com"))
		})

		It("should not add duplicate To recipients", func() {
			email.AddRecipients(libsnd.RecipientTo, "to@example.com")
			email.AddRecipients(libsnd.RecipientTo, "to@example.com")
			recipients := email.GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(1))
		})

		It("should handle empty To recipients", func() {
			recipients := email.GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(BeEmpty())
		})
	})

	Describe("CC Recipients", func() {
		It("should add CC recipients", func() {
			email.AddRecipients(libsnd.RecipientCC, "cc1@example.com", "cc2@example.com")
			recipients := email.GetRecipients(libsnd.RecipientCC)
			Expect(recipients).To(HaveLen(2))
			Expect(recipients).To(ContainElements("cc1@example.com", "cc2@example.com"))
		})

		It("should set CC recipients (replace all)", func() {
			email.AddRecipients(libsnd.RecipientCC, "old@example.com")
			email.SetRecipients(libsnd.RecipientCC, "new@example.com")
			recipients := email.GetRecipients(libsnd.RecipientCC)
			Expect(recipients).To(HaveLen(1))
			Expect(recipients).To(ContainElement("new@example.com"))
		})

		It("should not add duplicate CC recipients", func() {
			email.AddRecipients(libsnd.RecipientCC, "cc@example.com")
			email.AddRecipients(libsnd.RecipientCC, "cc@example.com")
			recipients := email.GetRecipients(libsnd.RecipientCC)
			Expect(recipients).To(HaveLen(1))
		})
	})

	Describe("BCC Recipients", func() {
		It("should add BCC recipients", func() {
			email.AddRecipients(libsnd.RecipientBCC, "bcc1@example.com", "bcc2@example.com")
			recipients := email.GetRecipients(libsnd.RecipientBCC)
			Expect(recipients).To(HaveLen(2))
			Expect(recipients).To(ContainElements("bcc1@example.com", "bcc2@example.com"))
		})

		It("should set BCC recipients (replace all)", func() {
			email.AddRecipients(libsnd.RecipientBCC, "old@example.com")
			email.SetRecipients(libsnd.RecipientBCC, "new@example.com")
			recipients := email.GetRecipients(libsnd.RecipientBCC)
			Expect(recipients).To(HaveLen(1))
			Expect(recipients).To(ContainElement("new@example.com"))
		})

		It("should not add duplicate BCC recipients", func() {
			email.AddRecipients(libsnd.RecipientBCC, "bcc@example.com")
			email.AddRecipients(libsnd.RecipientBCC, "bcc@example.com")
			recipients := email.GetRecipients(libsnd.RecipientBCC)
			Expect(recipients).To(HaveLen(1))
		})
	})

	Describe("Mixed Recipients", func() {
		It("should handle all recipient types independently", func() {
			email.AddRecipients(libsnd.RecipientTo, "to@example.com")
			email.AddRecipients(libsnd.RecipientCC, "cc@example.com")
			email.AddRecipients(libsnd.RecipientBCC, "bcc@example.com")

			Expect(email.GetRecipients(libsnd.RecipientTo)).To(HaveLen(1))
			Expect(email.GetRecipients(libsnd.RecipientCC)).To(HaveLen(1))
			Expect(email.GetRecipients(libsnd.RecipientBCC)).To(HaveLen(1))
		})

		It("should allow same email in different recipient types", func() {
			email.AddRecipients(libsnd.RecipientTo, "same@example.com")
			email.AddRecipients(libsnd.RecipientCC, "same@example.com")
			email.AddRecipients(libsnd.RecipientBCC, "same@example.com")

			Expect(email.GetRecipients(libsnd.RecipientTo)).To(ContainElement("same@example.com"))
			Expect(email.GetRecipients(libsnd.RecipientCC)).To(ContainElement("same@example.com"))
			Expect(email.GetRecipients(libsnd.RecipientBCC)).To(ContainElement("same@example.com"))
		})

		It("should add multiple recipients in one call", func() {
			email.AddRecipients(libsnd.RecipientTo, "to1@example.com", "to2@example.com", "to3@example.com")
			recipients := email.GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(3))
		})
	})

	Describe("Integration with Mail Headers", func() {
		It("should reflect from address in mail headers", func() {
			email.SetFrom("from@example.com")
			headers := mail.GetHeaders()
			Expect(headers.Get("From")).To(Equal("from@example.com"))
		})

		It("should reflect recipients in mail headers", func() {
			email.AddRecipients(libsnd.RecipientTo, "to@example.com")
			email.AddRecipients(libsnd.RecipientCC, "cc@example.com")
			headers := mail.GetHeaders()
			Expect(headers["To"]).To(ContainElement("to@example.com"))
			Expect(headers["Cc"]).To(ContainElement("cc@example.com"))
		})
	})
})
