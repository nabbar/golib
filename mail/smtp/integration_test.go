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
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
	liberr "github.com/nabbar/golib/errors"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Integration Tests", Label("integration"), func() {
	BeforeEach(func() {
		testSMTPPort = getFreePort()
	})

	Describe("Full Email Flow", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		BeforeEach(func() {
			var err error
			backend = &testBackend{
				requireAuth: false,
				messages:    make([]testMessage, 0),
			}

			server, err = startTestSMTPServer(backend, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(server).ToNot(BeNil())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if server != nil {
				_ = server.Close()
			}
		})

		Context("without authentication", func() {
			It("should send a simple email", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				from := "sender@example.com"
				to := []string{"recipient@example.com"}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err := client.Send(ctx, from, to, email)
				Expect(err).ToNot(HaveOccurred())

				// Verify message was received
				Eventually(func() int {
					return len(backend.messages)
				}, 2*time.Second).Should(Equal(1))

				Expect(backend.messages[0].From).To(Equal(from))
				Expect(backend.messages[0].To).To(ConsistOf(to))
			})

			It("should send email to multiple recipients", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				from := "sender@example.com"
				to := []string{
					"recipient1@example.com",
					"recipient2@example.com",
					"recipient3@example.com",
				}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err := client.Send(ctx, from, to, email)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() int {
					return len(backend.messages)
				}, 2*time.Second).Should(Equal(1))

				Expect(backend.messages[0].From).To(Equal(from))
				Expect(backend.messages[0].To).To(ConsistOf(to))
			})

			It("should handle check before send", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				// Check connection
				err := client.Check(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Send email
				from := "sender@example.com"
				to := []string{"recipient@example.com"}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err = client.Send(ctx, from, to, email)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() int {
					return len(backend.messages)
				}, 2*time.Second).Should(Equal(1))
			})

			It("should handle multiple sends", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(10 * time.Second)
				defer cancel()

				for i := 0; i < 5; i++ {
					from := fmt.Sprintf("sender%d@example.com", i)
					to := []string{fmt.Sprintf("recipient%d@example.com", i)}
					email := newTestEmail(from, to[0], "Test Subject", "Test Body")

					err := client.Send(ctx, from, to, email)
					Expect(err).ToNot(HaveOccurred())
				}

				Eventually(func() int {
					return len(backend.messages)
				}, 3*time.Second).Should(Equal(5))
			})
		})

		Context("with authentication", func() {
			BeforeEach(func() {
				backend.requireAuth = true

				// Restart server with TLS for auth tests (SMTP clients require TLS for PLAIN auth)
				_ = server.Close()

				var err error
				server, err = startTestSMTPServer(backend, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(server).ToNot(BeNil())

				testSMTPHost, testSMTPPort, err = getServerHostPort(server)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should send with valid credentials", func() {
				cfg := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS, testSMTPUser, testSMTPPassword)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				from := "sender@example.com"
				to := []string{"recipient@example.com"}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err := client.Send(ctx, from, to, email)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() int {
					return len(backend.messages)
				}, 2*time.Second).Should(Equal(1))
			})

			It("should fail with invalid credentials", func() {
				cfg := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS, "wronguser", "wrongpass")
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				from := "sender@example.com"
				to := []string{"recipient@example.com"}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err := client.Send(ctx, from, to, email)
				Expect(err).To(HaveOccurred())
				Expect(liberr.ContainsString(err, "auth")).To(BeTrue())
			})

			It("should fail with missing credentials", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				from := "sender@example.com"
				to := []string{"recipient@example.com"}
				email := newTestEmail(from, to[0], "Test Subject", "Test Body")

				err := client.Send(ctx, from, to, email)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Concurrent Operations", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		BeforeEach(func() {
			var err error
			backend = &testBackend{
				requireAuth: false,
				messages:    make([]testMessage, 0),
			}

			server, err = startTestSMTPServer(backend, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(server).ToNot(BeNil())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if server != nil {
				_ = server.Close()
			}
		})

		It("should handle concurrent sends from single client", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			var wg sync.WaitGroup
			concurrency := 10

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, cancel := contextWithTimeout(5 * time.Second)
					defer cancel()

					from := fmt.Sprintf("sender%d@example.com", index)
					to := []string{fmt.Sprintf("recipient%d@example.com", index)}
					email := newTestEmail(from, to[0], "Test Subject", "Test Body")

					err := client.Send(ctx, from, to, email)
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()

			Eventually(func() int {
				return len(backend.messages)
			}, 5*time.Second).Should(Equal(concurrency))
		})

		It("should handle concurrent sends from multiple clients", func() {
			var wg sync.WaitGroup
			concurrency := 5

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
					client := newTestSMTPClient(cfg)
					defer client.Close()

					ctx, cancel := contextWithTimeout(5 * time.Second)
					defer cancel()

					from := fmt.Sprintf("sender%d@example.com", index)
					to := []string{fmt.Sprintf("recipient%d@example.com", index)}
					email := newTestEmail(from, to[0], "Test Subject", "Test Body")

					err := client.Send(ctx, from, to, email)
					Expect(err).ToNot(HaveOccurred())
				}(i)
			}

			wg.Wait()

			Eventually(func() int {
				return len(backend.messages)
			}, 5*time.Second).Should(Equal(concurrency))
		})

		It("should handle mixed operations concurrently", func() {
			var wg sync.WaitGroup
			concurrency := 10

			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, cancel := contextWithTimeout(5 * time.Second)
					defer cancel()

					if index%2 == 0 {
						// Check operation
						_ = client.Check(ctx)
					} else {
						// Send operation
						from := fmt.Sprintf("sender%d@example.com", index)
						to := []string{fmt.Sprintf("recipient%d@example.com", index)}
						email := newTestEmail(from, to[0], "Test Subject", "Test Body")
						_ = client.Send(ctx, from, to, email)
					}
				}(i)
			}

			wg.Wait()

			// At least some sends should have succeeded
			Eventually(func() int {
				return len(backend.messages)
			}, 3*time.Second).Should(BeNumerically(">", 0))
		})
	})

	Describe("Client Lifecycle", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		BeforeEach(func() {
			var err error
			backend = &testBackend{
				requireAuth: false,
				messages:    make([]testMessage, 0),
			}

			server, err = startTestSMTPServer(backend, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(server).ToNot(BeNil())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if server != nil {
				_ = server.Close()
			}
		})

		It("should handle clone and independent operations", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client1 := newTestSMTPClient(cfg)
			defer client1.Close()

			client2 := client1.Clone()
			defer client2.Close()

			ctx, cancel := contextWithTimeout(5 * time.Second)
			defer cancel()

			// Send from client1
			from1 := "sender1@example.com"
			to1 := []string{"recipient1@example.com"}
			email1 := newTestEmail(from1, to1[0], "Test 1", "Body 1")
			err := client1.Send(ctx, from1, to1, email1)
			Expect(err).ToNot(HaveOccurred())

			// Send from client2
			from2 := "sender2@example.com"
			to2 := []string{"recipient2@example.com"}
			email2 := newTestEmail(from2, to2[0], "Test 2", "Body 2")
			err = client2.Send(ctx, from2, to2, email2)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				return len(backend.messages)
			}, 3*time.Second).Should(Equal(2))
		})

		It("should handle config update and continue operations", func() {
			cfg1 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg1)
			defer client.Close()

			ctx, cancel := contextWithTimeout(5 * time.Second)
			defer cancel()

			// Send with original config
			from1 := "sender1@example.com"
			to1 := []string{"recipient1@example.com"}
			email1 := newTestEmail(from1, to1[0], "Test 1", "Body 1")
			err := client.Send(ctx, from1, to1, email1)
			Expect(err).ToNot(HaveOccurred())

			// Update config (same server, different settings)
			cfg2 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
			client.UpdConfig(cfg2, tlsConfig)

			// Send with updated config
			from2 := "sender2@example.com"
			to2 := []string{"recipient2@example.com"}
			email2 := newTestEmail(from2, to2[0], "Test 2", "Body 2")
			err = client.Send(ctx, from2, to2, email2)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				return len(backend.messages)
			}, 3*time.Second).Should(Equal(2))
		})

		It("should handle rapid create and destroy cycles", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

			for i := 0; i < 20; i++ {
				client := newTestSMTPClient(cfg)

				ctx, cancel := contextWithTimeout(2 * time.Second)

				from := fmt.Sprintf("sender%d@example.com", i)
				to := []string{fmt.Sprintf("recipient%d@example.com", i)}
				email := newTestEmail(from, to[0], "Test", "Body")

				_ = client.Send(ctx, from, to, email)

				cancel()
				client.Close()
			}

			// At least some should have succeeded
			Eventually(func() int {
				return len(backend.messages)
			}, 5*time.Second).Should(BeNumerically(">", 0))
		})
	})

	Describe("Error Recovery", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		BeforeEach(func() {
			var err error
			backend = &testBackend{
				requireAuth: false,
				messages:    make([]testMessage, 0),
			}

			server, err = startTestSMTPServer(backend, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(server).ToNot(BeNil())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if server != nil {
				_ = server.Close()
			}
		})

		It("should recover after server restart", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			ctx, cancel := contextWithTimeout(30 * time.Second)
			defer cancel()

			// Send before restart
			from1 := "sender1@example.com"
			to1 := []string{"recipient1@example.com"}
			email1 := newTestEmail(from1, to1[0], "Test 1", "Body 1")
			err := client.Send(ctx, from1, to1, email1)
			Expect(err).ToNot(HaveOccurred())

			// Stop server
			_ = server.Close()
			time.Sleep(200 * time.Millisecond)

			// Restart server
			backend.messages = make([]testMessage, 0)
			server, err = startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
			// Update client config with new port
			newCfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client.UpdConfig(newCfg, cliTLS.TlsConfig(""))

			// Send after restart
			from2 := "sender2@example.com"
			to2 := []string{"recipient2@example.com"}
			email2 := newTestEmail(from2, to2[0], "Test 2", "Body 2")
			err = client.Send(ctx, from2, to2, email2)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				return len(backend.messages)
			}, 2*time.Second).Should(Equal(1))
		})
	})

	Describe("Real-World Scenarios", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		BeforeEach(func() {
			var err error
			backend = &testBackend{
				requireAuth: false,
				messages:    make([]testMessage, 0),
			}

			server, err = startTestSMTPServer(backend, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(server).ToNot(BeNil())

			testSMTPHost, testSMTPPort, err = getServerHostPort(server)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if server != nil {
				_ = server.Close()
			}
		})

		It("should handle email with attachments (simulated)", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			ctx, cancel := contextWithTimeout(30 * time.Second)
			defer cancel()

			// Simulate email with attachment (smaller size)

			from := "sender@example.com"
			to := []string{"recipient@example.com"}

			largeBody := "From: " + from + "\n"
			largeBody += "To: " + to[0] + "\n"
			largeBody += "\n"
			largeBody += "Body:\n"

			for i := 0; i < 50; i++ {
				largeBody += "This is line " + fmt.Sprintf("%d", i) + " of the email content."
			}

			email := &testWriter{data: largeBody}

			err := client.Send(ctx, from, to, email)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				return len(backend.messages)
			}, 5*time.Second).Should(Equal(1))
		})

		It("should handle batch email sending", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			ctx, cancel := contextWithTimeout(30 * time.Second)
			defer cancel()

			batchSize := 25
			for i := 0; i < batchSize; i++ {
				from := "newsletter@example.com"
				to := []string{fmt.Sprintf("subscriber%d@example.com", i)}
				email := newTestEmail(from, to[0], "Newsletter", "Content")

				err := client.Send(ctx, from, to, email)
				Expect(err).ToNot(HaveOccurred())
			}

			Eventually(func() int {
				return len(backend.messages)
			}, 5*time.Second).Should(Equal(batchSize))
		})
	})
})
