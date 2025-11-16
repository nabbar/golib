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
	"fmt"
	"strings"
	"time"

	libsnd "github.com/nabbar/golib/mail/sender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Performance Benchmarks", func() {

	Describe("Mail Creation Performance", func() {
		It("should measure mail creation time", func() {
			experiment := gmeasure.NewExperiment("Mail Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("creation", func() {
					mail := newMail()
					Expect(mail).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("creation")
			AddReportEntry("Creation Stats", fmt.Sprintf("Mean: %v, StdDev: %v", stats.DurationFor(gmeasure.StatMean), stats.DurationFor(gmeasure.StatStdDev)))
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure configured mail creation time", func() {
			experiment := gmeasure.NewExperiment("Configured Mail Creation")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("configured_creation", func() {
					mail := newMailWithBasicConfig()
					Expect(mail).ToNot(BeNil())
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("configured_creation")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 500*time.Microsecond))
		})
	})

	Describe("Mail Property Operations Performance", func() {
		var mail libsnd.Mail

		BeforeEach(func() {
			mail = newMail()
		})

		It("should measure subject set/get performance", func() {
			experiment := gmeasure.NewExperiment("Subject Operations")
			AddReportEntry(experiment.Name, experiment)

			subject := "Test Subject"
			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_subject", func() {
					mail.SetSubject(subject)
				})
				experiment.MeasureDuration("get_subject", func() {
					_ = mail.GetSubject()
				})
			}, gmeasure.SamplingConfig{N: 10000})

			setStats := experiment.GetStats("set_subject")
			getStats := experiment.GetStats("get_subject")

			Expect(setStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
			Expect(getStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure encoding set/get performance", func() {
			experiment := gmeasure.NewExperiment("Encoding Operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_encoding", func() {
					mail.SetEncoding(libsnd.EncodingBase64)
				})
				experiment.MeasureDuration("get_encoding", func() {
					_ = mail.GetEncoding()
				})
			}, gmeasure.SamplingConfig{N: 10000})

			setStats := experiment.GetStats("set_encoding")
			getStats := experiment.GetStats("get_encoding")

			Expect(setStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
			Expect(getStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure priority set/get performance", func() {
			experiment := gmeasure.NewExperiment("Priority Operations")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_priority", func() {
					mail.SetPriority(libsnd.PriorityHigh)
				})
				experiment.MeasureDuration("get_priority", func() {
					_ = mail.GetPriority()
				})
			}, gmeasure.SamplingConfig{N: 10000})

			setStats := experiment.GetStats("set_priority")
			getStats := experiment.GetStats("get_priority")

			Expect(setStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
			Expect(getStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})
	})

	Describe("Email Address Operations Performance", func() {
		var mail libsnd.Mail

		BeforeEach(func() {
			mail = newMail()
		})

		It("should measure from address operations", func() {
			experiment := gmeasure.NewExperiment("From Address Operations")
			AddReportEntry(experiment.Name, experiment)

			email := mail.Email()
			addr := "sender@example.com"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_from", func() {
					email.SetFrom(addr)
				})
				experiment.MeasureDuration("get_from", func() {
					_ = email.GetFrom()
				})
			}, gmeasure.SamplingConfig{N: 10000})

			setStats := experiment.GetStats("set_from")
			getStats := experiment.GetStats("get_from")

			Expect(setStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
			Expect(getStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure recipient operations", func() {
			experiment := gmeasure.NewExperiment("Recipient Operations")
			AddReportEntry(experiment.Name, experiment)

			email := mail.Email()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add_recipient", func() {
					email.AddRecipients(libsnd.RecipientTo, fmt.Sprintf("recipient%d@example.com", idx))
				})
			}, gmeasure.SamplingConfig{N: 100})

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("get_recipients", func() {
					_ = email.GetRecipients(libsnd.RecipientTo)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			addStats := experiment.GetStats("add_recipient")
			getStats := experiment.GetStats("get_recipients")

			Expect(addStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 50*time.Microsecond))
			Expect(getStats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 20*time.Microsecond))
		})
	})

	Describe("Header Operations Performance", func() {
		var mail libsnd.Mail

		BeforeEach(func() {
			mail = newMail()
		})

		It("should measure header addition", func() {
			experiment := gmeasure.NewExperiment("Header Addition")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add_header", func() {
					mail.AddHeader(fmt.Sprintf("X-Custom-%d", idx), "value")
				})
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("add_header")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 50*time.Microsecond))
		})

		It("should measure get all headers", func() {
			// Pre-populate headers
			for i := 0; i < 50; i++ {
				mail.AddHeader(fmt.Sprintf("X-Header-%d", i), "value")
			}

			experiment := gmeasure.NewExperiment("Get All Headers")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("get_headers", func() {
					_ = mail.GetHeaders()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("get_headers")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 200*time.Microsecond))
		})
	})

	Describe("Body Operations Performance", func() {
		var mail libsnd.Mail

		BeforeEach(func() {
			mail = newMail()
		})

		It("should measure small body operations", func() {
			experiment := gmeasure.NewExperiment("Small Body Operations")
			AddReportEntry(experiment.Name, experiment)

			smallContent := "Small body content"

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_body", func() {
					body := newReadCloser(smallContent)
					mail.SetBody(libsnd.ContentPlainText, body)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("set_body")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure large body operations", func() {
			experiment := gmeasure.NewExperiment("Large Body Operations")
			AddReportEntry(experiment.Name, experiment)

			largeContent := strings.Repeat("Content ", 10000)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("set_large_body", func() {
					body := newReadCloser(largeContent)
					mail.SetBody(libsnd.ContentPlainText, body)
				})
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("set_large_body")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 500*time.Microsecond))
		})

		It("should measure add alternative body", func() {
			experiment := gmeasure.NewExperiment("Add Alternative Body")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				plainBody := newReadCloser("Plain text")
				mail.SetBody(libsnd.ContentPlainText, plainBody)

				experiment.MeasureDuration("add_html_body", func() {
					htmlBody := newReadCloser("<html>HTML</html>")
					mail.AddBody(libsnd.ContentHTML, htmlBody)
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("add_html_body")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})
	})

	Describe("Attachment Operations Performance", func() {
		var mail libsnd.Mail

		BeforeEach(func() {
			mail = newMail()
		})

		It("should measure attachment addition", func() {
			experiment := gmeasure.NewExperiment("Attachment Addition")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add_attachment", func() {
					data := newReadCloser("attachment data")
					mail.AddAttachment(fmt.Sprintf("file%d.txt", idx), "text/plain", data, false)
				})
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("add_attachment")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure inline attachment addition", func() {
			experiment := gmeasure.NewExperiment("Inline Attachment Addition")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("add_inline", func() {
					data := newReadCloser("inline data")
					mail.AddAttachment(fmt.Sprintf("inline%d.png", idx), "image/png", data, true)
				})
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("add_inline")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure getting attachments", func() {
			// Pre-populate attachments
			for i := 0; i < 20; i++ {
				data := newReadCloser("data")
				mail.AddAttachment(fmt.Sprintf("file%d.txt", i), "text/plain", data, false)
			}

			experiment := gmeasure.NewExperiment("Get Attachments")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("get_attachments", func() {
					_ = mail.GetAttachment(false)
				})
			}, gmeasure.SamplingConfig{N: 10000})

			stats := experiment.GetStats("get_attachments")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 20*time.Microsecond))
		})
	})

	Describe("Clone Performance", func() {
		It("should measure clone operation with minimal data", func() {
			experiment := gmeasure.NewExperiment("Clone Minimal Mail")
			AddReportEntry(experiment.Name, experiment)

			mail := newMail()

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone", func() {
					_ = mail.Clone()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("clone")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Microsecond))
		})

		It("should measure clone operation with full data", func() {
			experiment := gmeasure.NewExperiment("Clone Full Mail")
			AddReportEntry(experiment.Name, experiment)

			mail := newMailWithBasicConfig()
			mail.AddHeader("X-Custom-1", "value1")
			mail.AddHeader("X-Custom-2", "value2")
			mail.SetBody(libsnd.ContentPlainText, newReadCloser("body"))
			mail.AddAttachment("file.txt", "text/plain", newReadCloser("data"), false)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("clone_full", func() {
					_ = mail.Clone()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("clone_full")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 200*time.Microsecond))
		})
	})

	Describe("Sender Creation Performance", func() {
		It("should measure sender creation", func() {
			experiment := gmeasure.NewExperiment("Sender Creation")
			AddReportEntry(experiment.Name, experiment)

			mail := newMailWithBasicConfig()
			mail.SetBody(libsnd.ContentPlainText, newReadCloser("Test body"))

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create_sender", func() {
					sender, err := mail.Sender()
					Expect(err).ToNot(HaveOccurred())
					if sender != nil {
						_ = sender.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 100})

			stats := experiment.GetStats("create_sender")
			AddReportEntry("Sender Creation Stats", fmt.Sprintf("Mean: %v, StdDev: %v", stats.DurationFor(gmeasure.StatMean), stats.DurationFor(gmeasure.StatStdDev)))
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Millisecond))
		})

		It("should measure sender creation with attachments", func() {
			experiment := gmeasure.NewExperiment("Sender with Attachments")
			AddReportEntry(experiment.Name, experiment)

			mail := newMailWithBasicConfig()
			mail.SetBody(libsnd.ContentPlainText, newReadCloser("Body"))
			mail.AddAttachment("file.txt", "text/plain", newReadCloser("data"), false)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("create_sender_with_attachment", func() {
					sender, err := mail.Sender()
					Expect(err).ToNot(HaveOccurred())
					if sender != nil {
						_ = sender.Close()
					}
				})
			}, gmeasure.SamplingConfig{N: 50})

			stats := experiment.GetStats("create_sender_with_attachment")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 20*time.Millisecond))
		})
	})

	Describe("Config Operations Performance", func() {
		It("should measure config validation", func() {
			experiment := gmeasure.NewExperiment("Config Validation")
			AddReportEntry(experiment.Name, experiment)

			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "Base 64",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("validate", func() {
					_ = cfg.Validate()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("validate")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 500*time.Microsecond))
		})

		It("should measure new mailer from config", func() {
			experiment := gmeasure.NewExperiment("NewMailer from Config")
			AddReportEntry(experiment.Name, experiment)

			cfg := libsnd.Config{
				Charset:  "UTF-8",
				Subject:  "Test",
				Encoding: "Base 64",
				Priority: "Normal",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
			}

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("new_mailer", func() {
					_, _ = cfg.NewMailer()
				})
			}, gmeasure.SamplingConfig{N: 1000})

			stats := experiment.GetStats("new_mailer")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 500*time.Microsecond))
		})
	})

	Describe("Type Parsing Performance", func() {
		It("should measure encoding parsing", func() {
			experiment := gmeasure.NewExperiment("Encoding Parsing")
			AddReportEntry(experiment.Name, experiment)

			encodings := []string{"None", "Binary", "Base 64", "Quoted Printable"}

			experiment.Sample(func(idx int) {
				enc := encodings[idx%len(encodings)]
				experiment.MeasureDuration("parse_encoding", func() {
					_ = libsnd.ParseEncoding(enc)
				})
			}, gmeasure.SamplingConfig{N: 10000})

			stats := experiment.GetStats("parse_encoding")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})

		It("should measure priority parsing", func() {
			experiment := gmeasure.NewExperiment("Priority Parsing")
			AddReportEntry(experiment.Name, experiment)

			priorities := []string{"Normal", "Low", "High"}

			experiment.Sample(func(idx int) {
				pri := priorities[idx%len(priorities)]
				experiment.MeasureDuration("parse_priority", func() {
					_ = libsnd.ParsePriority(pri)
				})
			}, gmeasure.SamplingConfig{N: 10000})

			stats := experiment.GetStats("parse_priority")
			Expect(stats.DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Microsecond))
		})
	})
})
