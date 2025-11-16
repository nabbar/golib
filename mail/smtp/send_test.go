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

package smtp_test

import (
	"time"

	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Send Operations", func() {

	Describe("Error Scenarios", func() {
		Context("with connection failures", func() {
			It("should fail gracefully when server is unreachable", func() {
				cfg := newTestConfig("localhost", getUnusedPort(), smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				email := newTestEmail("sender@example.com", "recipient@example.com", "Test", "Body")
				err := client.Send(ctx, "sender@example.com", []string{"recipient@example.com"}, email)

				Expect(err).To(HaveOccurred())
			})

			It("should respect context timeout during send", func() {
				cfg := newTestConfig("localhost", getFreePort(), smtptp.TLSNone) // Non-routable
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(500 * time.Millisecond)
				defer cancel()

				email := newTestEmail("sender@example.com", "recipient@example.com", "Test", "Body")
				start := time.Now()
				err := client.Send(ctx, "sender@example.com", []string{"recipient@example.com"}, email)
				elapsed := time.Since(start)

				Expect(err).To(HaveOccurred())
				Expect(elapsed).To(BeNumerically("<", 2*time.Second))
			})

			It("should handle cancelled context", func() {
				cfg := newTestConfig("localhost", getFreePort(), smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				cancel() // Cancel immediately

				email := newTestEmail("sender@example.com", "recipient@example.com", "Test", "Body")
				err := client.Send(ctx, "sender@example.com", []string{"recipient@example.com"}, email)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
