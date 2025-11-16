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
	"os"
	"path/filepath"

	libsnd "github.com/nabbar/golib/mail/sender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config Operations", func() {
	Describe("Config Validation", func() {
		It("should validate valid minimal config", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test Subject",
				Encoding: "Base 64",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject config with missing required fields", func() {
			cfg := libsnd.Config{}
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should reject config with invalid email format", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "invalid-email",
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should validate config with optional fields", func() {
			cfg := libsnd.Config{
				Charset:    "UTF-8",
				Subject:    "Test",
				Encoding:   "Base 64",
				Priority:   "High",
				From:       "sender@example.com",
				Sender:     "actual-sender@example.com",
				ReplyTo:    "reply@example.com",
				ReturnPath: "return@example.com",
				To:         []string{"to@example.com"},
				Cc:         []string{"cc@example.com"},
				Bcc:        []string{"bcc@example.com"},
			}

			err := cfg.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject config with invalid To email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"invalid-email"},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should reject config with invalid Cc email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Cc:       []string{"invalid-cc"},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should reject config with invalid Bcc email", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Bcc:      []string{"invalid-bcc"},
			}

			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewMailer from Config", func() {
		It("should create mail from valid config", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test Subject",
				Encoding: "Base 64",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail).ToNot(BeNil())
			Expect(mail.GetSubject()).To(Equal("Test Subject"))
			Expect(mail.GetCharset()).To(Equal("UTF-8"))
			Expect(mail.GetEncoding()).To(Equal(libsnd.EncodingBase64))
			Expect(mail.GetPriority()).To(Equal(libsnd.PriorityNormal))
		})

		It("should set from address", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.Email().GetFrom()).To(Equal("sender@example.com"))
		})

		It("should set sender address when provided", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				Sender:   "sender@example.com",
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.Email().GetSender()).To(Equal("sender@example.com"))
		})

		It("should set replyTo address when provided", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "from@example.com",
				ReplyTo:  "reply@example.com",
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.Email().GetReplyTo()).To(Equal("reply@example.com"))
		})

		It("should set returnPath when provided", func() {
			cfg := libsnd.Config{
				Charset:    "UTF-8",
				Subject:    "Test",
				Encoding:   "None",
				Priority:   "Normal",
				From:       "from@example.com",
				ReturnPath: "return@example.com",
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.Email().GetReturnPath()).To(Equal("return@example.com"))
		})

		It("should set To recipients", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"to1@example.com", "to2@example.com"},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			recipients := mail.Email().GetRecipients(libsnd.RecipientTo)
			Expect(recipients).To(HaveLen(2))
			Expect(recipients).To(ContainElements("to1@example.com", "to2@example.com"))
		})

		It("should set Cc recipients", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Cc:       []string{"cc@example.com"},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			recipients := mail.Email().GetRecipients(libsnd.RecipientCC)
			Expect(recipients).To(ContainElement("cc@example.com"))
		})

		It("should set Bcc recipients", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Bcc:      []string{"bcc@example.com"},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			recipients := mail.Email().GetRecipients(libsnd.RecipientBCC)
			Expect(recipients).To(ContainElement("bcc@example.com"))
		})

		It("should set custom headers", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
					"X-Another":       "another-value",
				},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			headers := mail.GetHeaders()
			Expect(headers.Get("X-Custom-Header")).To(Equal("custom-value"))
			Expect(headers.Get("X-Another")).To(Equal("another-value"))
		})

		It("should handle different encodings", func() {
			encodings := map[string]libsnd.Encoding{
				"None":             libsnd.EncodingNone,
				"Binary":           libsnd.EncodingBinary,
				"Base 64":          libsnd.EncodingBase64,
				"Quoted Printable": libsnd.EncodingQuotedPrintable,
			}

			for encodingStr, expectedEncoding := range encodings {
				cfg := libsnd.Config{
					Charset:  "UTF-8",
					Subject:  "Test",
					Encoding: encodingStr,
					Priority: "Normal",
					From:     "sender@example.com",
				}

				mail, err := cfg.NewMailer()
				Expect(err).ToNot(HaveOccurred())
				Expect(mail.GetEncoding()).To(Equal(expectedEncoding))
			}
		})

		It("should handle different priorities", func() {
			priorities := map[string]libsnd.Priority{
				"Normal": libsnd.PriorityNormal,
				"Low":    libsnd.PriorityLow,
				"High":   libsnd.PriorityHigh,
			}

			for priorityStr, expectedPriority := range priorities {
				cfg := libsnd.Config{
					Charset:  "UTF-8",
					Subject:  "Test",
					Encoding: "None",
					Priority: priorityStr,
					From:     "sender@example.com",
				}

				mail, err := cfg.NewMailer()
				Expect(err).ToNot(HaveOccurred())
				Expect(mail.GetPriority()).To(Equal(expectedPriority))
			}
		})
	})

	Describe("ConfigFile with Attachments", func() {
		var tempDir string
		var testFile string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "mail-sender-test-*")
			Expect(err).ToNot(HaveOccurred())

			testFile = filepath.Join(tempDir, "test-attachment.txt")
			err = os.WriteFile(testFile, []byte("Test attachment content"), 0644)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if tempDir != "" {
				_ = os.RemoveAll(tempDir)
			}
		})

		It("should attach file from config", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Attach: []libsnd.ConfigFile{
					{
						Name: "attachment.txt",
						Mime: "text/plain",
						Path: testFile,
					},
				},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(1))
		})

		It("should inline file from config", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Inline: []libsnd.ConfigFile{
					{
						Name: "inline.txt",
						Mime: "text/plain",
						Path: testFile,
					},
				},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			inlines := mail.GetAttachment(true)
			Expect(inlines).To(HaveLen(1))
		})

		It("should return error for non-existent file", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Attach: []libsnd.ConfigFile{
					{
						Name: "missing.txt",
						Mime: "text/plain",
						Path: "/non/existent/file.txt",
					},
				},
			}

			_, err := cfg.NewMailer()
			Expect(err).To(HaveOccurred())
		})

		It("should handle multiple attachments", func() {
			file2 := filepath.Join(tempDir, "test2.txt")
			err := os.WriteFile(file2, []byte("Second file"), 0644)
			Expect(err).ToNot(HaveOccurred())

			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Attach: []libsnd.ConfigFile{
					{Name: "file1.txt", Mime: "text/plain", Path: testFile},
					{Name: "file2.txt", Mime: "text/plain", Path: file2},
				},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			attachments := mail.GetAttachment(false)
			Expect(attachments).To(HaveLen(2))
		})

		It("should handle both attachments and inlines", func() {
			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "None",
				Priority: "Normal",
				From:     "sender@example.com",
				Attach: []libsnd.ConfigFile{
					{Name: "attach.txt", Mime: "text/plain", Path: testFile},
				},
				Inline: []libsnd.ConfigFile{
					{Name: "inline.txt", Mime: "text/plain", Path: testFile},
				},
			}

			mail, err := cfg.NewMailer()
			Expect(err).ToNot(HaveOccurred())
			Expect(mail.GetAttachment(false)).To(HaveLen(1))
			Expect(mail.GetAttachment(true)).To(HaveLen(1))
		})
	})
})
