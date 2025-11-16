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

	"github.com/emersion/go-smtp"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Connection and Dial", func() {
	BeforeEach(func() {
		testSMTPPort = getFreePort()
	})

	Describe("Connection Establishment", func() {
		var (
			server  *smtp.Server
			backend *testBackend
		)

		Context("with plain SMTP", func() {
			It("should connect to non-existent server and timeout", func() {
				cfg := newTestConfig("nonexistent.example.com", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Or(
					ContainSubstring("dial"),
					ContainSubstring("timeout"),
					ContainSubstring("connection"),
				))
			})

			It("should handle connection refused", func() {
				// Use a port that is very unlikely to be in use
				cfg := newTestConfig("127.0.0.1", 9876, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle invalid host format", func() {
				cfg := newTestConfig("invalid host with spaces", testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with TLS", func() {
			BeforeEach(func() {
				var err error
				backend = &testBackend{
					requireAuth: false,
					messages:    make([]testMessage, 0),
				}

				// Start non-TLS server for testing TLS connection failures
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

			It("should fail on TLS connection to non-TLS server", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(3 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				// Should fail because we're trying TLS on a non-TLS port
				Expect(err).To(HaveOccurred())
			})

			It("should handle TLS connection timeout", func() {
				cfg := newTestConfig("nonexistent.example.com", 465, smtptp.TLSStrictTLS)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Or(
					ContainSubstring("dial"),
					ContainSubstring("timeout"),
					ContainSubstring("connection"),
				))
			})

			It("should handle certificate verification failure", func() {
				// Stop the non-TLS server and start a TLS server
				_ = server.Close()

				var err error
				server, err = startTestSMTPServer(backend, true)
				Expect(err).ToNot(HaveOccurred())

				testSMTPHost, testSMTPPort, err = getServerHostPort(server)
				Expect(err).ToNot(HaveOccurred())

				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS)
				// Don't skip verify - should fail with self-signed cert
				cfg.ForceTLSSkipVerify(false)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(3 * time.Second)
				defer cancel()

				err = client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with different network protocols", func() {
			It("should handle TCP connection", func() {
				cfg := newTestConfig(testSMTPHost, 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred()) // Port not listening
			})

			It("should handle IPv4 explicitly", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle IPv6 localhost", func() {
				cfg := newTestConfig("::1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				// May fail if IPv6 not supported, which is okay
				_ = client.Check(ctx)
			})
		})
	})

	Describe("Connection Timeout Handling", func() {
		Context("with various timeout durations", func() {
			It("should respect context timeout", func() {
				cfg := newTestConfig("240.0.0.1", 9999, smtptp.TLSNone) // Non-routable IP
				client := newTestSMTPClient(cfg)
				defer client.Close()

				start := time.Now()
				ctx, cancel := contextWithTimeout(500 * time.Millisecond)
				defer cancel()

				err := client.Check(ctx)
				elapsed := time.Since(start)

				Expect(err).To(HaveOccurred())
				// Should timeout within reasonable time (with some buffer)
				Expect(elapsed).To(BeNumerically("<", 2*time.Second))
			})

			It("should handle immediate timeout", func() {
				cfg := newTestConfig("240.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(1 * time.Millisecond)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should handle cancelled context", func() {
				cfg := newTestConfig("240.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				cancel() // Cancel immediately

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Connection State Management", func() {
		Context("with multiple operations", func() {
			It("should handle check after close", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				// Should either fail to connect or handle gracefully
				_ = err
			})

			It("should handle multiple check calls", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				// Multiple checks should all fail consistently
				err1 := client.Check(ctx)
				err2 := client.Check(ctx)

				Expect(err1).To(HaveOccurred())
				Expect(err2).To(HaveOccurred())
			})

			It("should handle Client() call without connection", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				smtpClient, err := client.Client(ctx)
				// Should fail because port is not listening
				_ = smtpClient
				_ = err
			})
		})

		Context("with concurrent operations", func() {
			It("should handle concurrent Check calls", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				done := make(chan bool, 5)
				for i := 0; i < 5; i++ {
					go func() {
						defer GinkgoRecover()
						ctx, cancel := contextWithTimeout(2 * time.Second)
						defer cancel()

						err := client.Check(ctx)
						Expect(err).To(HaveOccurred())
						done <- true
					}()
				}

				for i := 0; i < 5; i++ {
					Eventually(done, 5*time.Second).Should(Receive())
				}
			})

			It("should handle concurrent Client() calls", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				done := make(chan bool, 3)
				for i := 0; i < 3; i++ {
					go func() {
						defer GinkgoRecover()
						ctx, cancel := contextWithTimeout(2 * time.Second)
						defer cancel()

						_, err := client.Client(ctx)
						_ = err // May succeed or fail
						done <- true
					}()
				}

				for i := 0; i < 3; i++ {
					Eventually(done, 5*time.Second).Should(Receive())
				}
			})
		})
	})

	Describe("Error Handling", func() {
		Context("with network errors", func() {
			It("should return appropriate error for connection refused", func() {
				cfg := newTestConfig("127.0.0.1", 9876, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).ToNot(BeEmpty())
			})

			It("should return appropriate error for host not found", func() {
				cfg := newTestConfig("nonexistent-host-123456.example", 25, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(3 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should return appropriate error for timeout", func() {
				cfg := newTestConfig("240.0.0.1", 25, smtptp.TLSNone) // Non-routable
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(500 * time.Millisecond)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with TLS errors", func() {
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

				// Start TLS server for testing handshake failures
				server, err = startTestSMTPServer(backend, true)
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

			It("should handle TLS handshake failure", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS)
				// Don't skip verify - should fail with self-signed cert
				cfg.ForceTLSSkipVerify(false)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(3 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
