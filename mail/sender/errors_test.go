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
	"fmt"
	"time"

	smtpsv "github.com/emersion/go-smtp"
	liberr "github.com/nabbar/golib/errors"
	libsnd "github.com/nabbar/golib/mail/sender"
	libsmtp "github.com/nabbar/golib/mail/smtp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Handling", func() {
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

	Describe("Error Code Definitions", func() {
		It("should have ErrorParamEmpty defined", func() {
			err := libsnd.ErrorParamEmpty.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
		})

		It("should have ErrorMailConfigInvalid defined", func() {
			err := libsnd.ErrorMailConfigInvalid.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("config is invalid"))
		})

		It("should have ErrorMailIORead defined", func() {
			err := libsnd.ErrorMailIORead.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot read bytes from io source"))
		})

		It("should have ErrorMailIOWrite defined", func() {
			err := libsnd.ErrorMailIOWrite.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot write given string to IO resource"))
		})

		It("should have ErrorMailDateParsing defined", func() {
			err := libsnd.ErrorMailDateParsing.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error occurs while trying to parse a date string"))
		})

		It("should have ErrorMailSmtpClient defined", func() {
			err := libsnd.ErrorMailSmtpClient.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error occurs while to checking connection with SMTP server"))
		})

		It("should have ErrorMailSenderInit defined", func() {
			err := libsnd.ErrorMailSenderInit.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error occurs while to preparing SMTP Email sender"))
		})

		It("should have ErrorFileOpenCreate defined", func() {
			err := libsnd.ErrorFileOpenCreate.Error(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot open/create file"))
		})
	})

	Describe("Error Wrapping", func() {
		It("should wrap parent error in ErrorParamEmpty", func() {
			parentErr := fmt.Errorf("parent error")
			err := libsnd.ErrorParamEmpty.Error(parentErr)
			Expect(err).To(HaveOccurred())
			// Check that the error is properly created with parent
			if libErr, ok := err.(liberr.Error); ok {
				Expect(libErr.HasParent()).To(BeTrue())
			}
		})

		It("should wrap parent error in ErrorMailConfigInvalid", func() {
			parentErr := fmt.Errorf("validation failed")
			err := libsnd.ErrorMailConfigInvalid.Error(parentErr)
			Expect(err).To(HaveOccurred())
			if libErr, ok := err.(liberr.Error); ok {
				Expect(libErr.HasParent()).To(BeTrue())
			}
		})

		It("should wrap parent error in ErrorMailIORead", func() {
			parentErr := fmt.Errorf("read failed")
			err := libsnd.ErrorMailIORead.Error(parentErr)
			Expect(err).To(HaveOccurred())
			if libErr, ok := err.(liberr.Error); ok {
				Expect(libErr.HasParent()).To(BeTrue())
			}
		})

		It("should wrap parent error in ErrorMailIOWrite", func() {
			parentErr := fmt.Errorf("write failed")
			err := libsnd.ErrorMailIOWrite.Error(parentErr)
			Expect(err).To(HaveOccurred())
			if libErr, ok := err.(liberr.Error); ok {
				Expect(libErr.HasParent()).To(BeTrue())
			}
		})
	})

	Describe("DateTime Parsing Errors", func() {
		It("should return error for invalid date format", func() {
			err := mail.SetDateString(time.RFC1123Z, "not-a-date")
			Expect(err).To(HaveOccurred())

			// Check if it's an error (the specific error code may vary)
			Expect(err.Error()).ToNot(BeEmpty())
		})

		It("should return error for mismatched layout", func() {
			err := mail.SetDateString(time.RFC1123, "Mon, 02 Jan 2006 15:04:05 -0700")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for empty date string", func() {
			err := mail.SetDateString(time.RFC1123Z, "")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for partial date string", func() {
			err := mail.SetDateString(time.RFC1123Z, "Mon, 02 Jan")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Sender Creation Errors", func() {
		It("should handle sender creation with minimal config", func() {
			mail := libsnd.New()
			mail.Email().SetFrom("from@example.com")
			mail.Email().AddRecipients(libsnd.RecipientTo, "to@example.com")

			// Even without body, should create sender
			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			if sender != nil {
				defer sender.Close()
			}
		})

		It("should create sender even without recipients (validation happens at send)", func() {
			mail := libsnd.New()
			mail.Email().SetFrom("from@example.com")

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			Expect(sender).ToNot(BeNil())
			if sender != nil {
				defer sender.Close()
			}
		})
	})

	Describe("Send Operation Errors", func() {
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

		It("should return error when from address is invalid", func() {
			mail.Email().SetFrom("abc") // Invalid email format
			body := newReadCloser("test body")
			mail.SetBody(libsnd.ContentPlainText, body)

			// Creating sender should fail with invalid email
			_, err := mail.Sender()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when from address is empty", func() {
			mail.Email().SetFrom("")
			body := newReadCloser("test body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when no recipients", func() {
			mail.Email().SetRecipients(libsnd.RecipientTo) // Clear all
			body := newReadCloser("test body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when recipient address is invalid", func() {
			mail.Email().SetRecipients(libsnd.RecipientTo, "abc") // Invalid email
			body := newReadCloser("test body")
			mail.SetBody(libsnd.ContentPlainText, body)

			// Creating sender should fail with invalid email
			_, err := mail.Sender()
			Expect(err).To(HaveOccurred())
		})

		It("should return error when SMTP server is unavailable", func() {
			// Close the server to simulate connection failure
			_ = smtpServer.Close()
			smtpServer = nil

			body := newReadCloser("test body")
			mail.SetBody(libsnd.ContentPlainText, body)

			sender, err := mail.Sender()
			Expect(err).ToNot(HaveOccurred())
			defer sender.Close()

			err = sender.Send(ctx, smtpClient)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Config Validation Errors", func() {
		It("should return error for missing charset", func() {
			cfg := libsnd.Config{
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing subject", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing encoding", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Priority: "Normal",
				From:     "from@example.com",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing priority", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				From:     "from@example.com",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing from", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid from email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "not-an-email",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid sender email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				Sender:   "not-an-email",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid replyTo email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				ReplyTo:  "not-an-email",
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid To email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				To:       []string{"valid@example.com", "invalid"},
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid Cc email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				Cc:       []string{"invalid"},
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid Bcc email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				Bcc:      []string{"invalid"},
			}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent attachment file", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				Attach: []libsnd.ConfigFile{
					{
						Name: "file.txt",
						Mime: "text/plain",
						Path: "/non/existent/file.txt",
					},
				},
			}

			// Validation may pass, but NewMailer should fail
			_, err := cfg.NewMailer()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Error Code Registration", func() {
		It("should have all error codes registered", func() {
			codes := []liberr.CodeError{
				libsnd.ErrorParamEmpty,
				libsnd.ErrorMailConfigInvalid,
				libsnd.ErrorMailIORead,
				libsnd.ErrorMailIOWrite,
				libsnd.ErrorMailDateParsing,
				libsnd.ErrorMailSmtpClient,
				libsnd.ErrorMailSenderInit,
				libsnd.ErrorFileOpenCreate,
			}

			for _, code := range codes {
				err := code.Error(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).ToNot(BeEmpty())
			}
		})

		It("should have unique error codes", func() {
			codes := []liberr.CodeError{
				libsnd.ErrorParamEmpty,
				libsnd.ErrorMailConfigInvalid,
				libsnd.ErrorMailIORead,
				libsnd.ErrorMailIOWrite,
				libsnd.ErrorMailDateParsing,
				libsnd.ErrorMailSmtpClient,
				libsnd.ErrorMailSenderInit,
				libsnd.ErrorFileOpenCreate,
			}

			seen := make(map[liberr.CodeError]bool)
			for _, code := range codes {
				Expect(seen[code]).To(BeFalse(), "Duplicate error code: %v", code)
				seen[code] = true
			}
		})
	})
})
