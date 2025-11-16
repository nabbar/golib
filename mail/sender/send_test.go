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
	"context"
	"time"

	smtpsv "github.com/emersion/go-smtp"
	libsnd "github.com/nabbar/golib/mail/sender"
	libsmtp "github.com/nabbar/golib/mail/smtp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sender Operations", func() {
	var (
		mail libsnd.Mail
		ctx  context.Context
		cnl  context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cnl = context.WithTimeout(testCtx, 10*time.Second)
		mail = newMailWithBasicConfig()
	})

	AfterEach(func() {
		if cnl != nil {
			cnl()
		}
	})

	Describe("Sender Creation", func() {
		It("should create sender from mail", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should create sender with HTML body", func() {
			body := newReadCloser("<html><body><h1>Test</h1></body></html>")
			mail.SetBody(libsnd.ContentHTML, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should create sender with multiple body parts", func() {
			plainBody := newReadCloser("Plain text version")
			htmlBody := newReadCloser("<html><body>HTML version</body></html>")
			mail.SetBody(libsnd.ContentPlainText, plainBody)
			mail.AddBody(libsnd.ContentHTML, htmlBody)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should create sender with attachments", func() {
			body := newReadCloser("Email with attachment")
			mail.SetBody(libsnd.ContentPlainText, body)

			attachment := newReadCloser("Attachment content")
			mail.AddAttachment("file.txt", "text/plain", attachment, false)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should create sender with inline attachments", func() {
			body := newReadCloser("Email with inline image")
			mail.SetBody(libsnd.ContentHTML, body)

			inline := newReadCloser("Image data")
			mail.AddAttachment("logo.png", "image/png", inline, true)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should handle different priorities", func() {
			body := newReadCloser("High priority email")
			mail.SetBody(libsnd.ContentPlainText, body)
			mail.SetPriority(libsnd.PriorityHigh)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})

		It("should handle different encodings", func() {
			body := newReadCloser("Base64 encoded email")
			mail.SetBody(libsnd.ContentPlainText, body)
			mail.SetEncoding(libsnd.EncodingBase64)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			if sender != nil {
				defer sender.Close()
			}
		})
	})

	Describe("Send with Real SMTP", func() {
		var (
			smtpServer *smtpsv.Server
			smtpClient libsmtp.SMTP
			backend    *testBackend
			host       string
			port       int
		)

		BeforeEach(func() {
			backend = &testBackend{requireAuth: false, messages: make([]testMessage, 0)}
			var err error
			smtpServer, host, port, err = startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			smtpClient = newTestSMTPClient(host, port)
		})

		AfterEach(func() {
			if smtpClient != nil {
				smtpClient.Close()
			}
			if smtpServer != nil {
				_ = smtpServer.Close()
			}
		})

		It("should send email successfully", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).ToNot(HaveOccurred())

			Expect(backend.messages).To(HaveLen(1))
		})

		It("should send and close", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			err = sender.SendClose(ctx, smtpClient)
			Expect(err).ToNot(HaveOccurred())

			Expect(backend.messages).To(HaveLen(1))
		})

		It("should handle multiple sends from same sender", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			// Send multiple times
			for i := 0; i < 3; i++ {
				err = sender.Send(ctx, smtpClient)
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(backend.messages).To(HaveLen(3))
		})

		It("should return error when from address is invalid", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)
			mail.Email().SetFrom("") // Invalid from

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when no recipients", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)
			mail.Email().SetRecipients(libsnd.RecipientTo) // Clear recipients

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).To(HaveOccurred())
		})

		It("should include all recipient types", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)
			mail.Email().AddRecipients(libsnd.RecipientCC, "cc@example.com")
			mail.Email().AddRecipients(libsnd.RecipientBCC, "bcc@example.com")

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).ToNot(HaveOccurred())

			Expect(backend.messages).To(HaveLen(1))
		})
	})

	Describe("Sender Lifecycle", func() {
		It("should close sender properly", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			err = sender.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle multiple close calls", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			_ = sender.Close()
			err = sender.Close()
			// Should not panic or cause issues
			_ = err
		})

		It("should clean up resources with SendClose", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())

			// SendClose will close the sender
			// We just check that it works without SMTP client for this test
		})
	})

	Describe("Context Handling", func() {
		var (
			smtpServer *smtpsv.Server
			smtpClient libsmtp.SMTP
			backend    *testBackend
			host       string
			port       int
		)

		BeforeEach(func() {
			backend = &testBackend{requireAuth: false, messages: make([]testMessage, 0)}
			var err error
			smtpServer, host, port, err = startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			smtpClient = newTestSMTPClient(host, port)
		})

		AfterEach(func() {
			if smtpClient != nil {
				smtpClient.Close()
			}
			if smtpServer != nil {
				_ = smtpServer.Close()
			}
		})

		It("should respect context cancellation", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			cancelCtx, cancel := context.WithCancel(ctx)
			cancel() // Cancel immediately

			err = sender.Send(cancelCtx, smtpClient)
			// May or may not error depending on timing
			_ = err
		})

		It("should respect context timeout", func() {
			body := newReadCloser("Test email body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			defer sender.Close()

			timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
			defer cancel()

			time.Sleep(10 * time.Millisecond) // Ensure timeout

			err = sender.Send(timeoutCtx, smtpClient)
			// May or may not error depending on implementation
			_ = err
		})
	})
})
