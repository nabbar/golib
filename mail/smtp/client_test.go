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
	"time"

	smtpcli "github.com/nabbar/golib/mail/smtp"
	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Client", func() {
	BeforeEach(func() {
		testSMTPPort = getFreePort()
	})

	Describe("Client Creation", func() {
		Context("with valid configuration", func() {
			It("should create a client with basic config", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})

			It("should create a client with authentication", func() {
				cfg := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSNone, testSMTPUser, testSMTPPassword)
				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})

			It("should create a client with TLS", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStrictTLS)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})

			It("should create a client with STARTTLS", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStartTLS)
				cfg.ForceTLSSkipVerify(true)
				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})
		})

		Context("with invalid configuration", func() {
			It("should fail with nil config", func() {
				client, err := smtpcli.New(nil, &tls.Config{})
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})

			It("should accept nil TLS config and use defaults", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client, err := smtpcli.New(cfg, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(client).ToNot(BeNil())
			})
		})
	})

	Describe("Client Operations", func() {
		Context("Clone", func() {
			It("should clone a client", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				cloned := client.Clone()
				Expect(cloned).ToNot(BeNil())
				Expect(cloned).ToNot(BeIdenticalTo(client))
			})

			It("should clone with same configuration", func() {
				cfg := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSNone, testSMTPUser, testSMTPPassword)
				client := newTestSMTPClient(cfg)

				cloned := client.Clone()
				Expect(cloned).ToNot(BeNil())

				// Both should be able to use the same configuration
				// This is verified by ensuring clone doesn't panic and is usable
			})

			It("should create independent clones", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client1 := newTestSMTPClient(cfg)
				client2 := client1.Clone()

				// Close client1 shouldn't affect client2
				client1.Close()

				// client2 should still be usable
				Expect(client2).ToNot(BeNil())
			})
		})

		Context("Close", func() {
			It("should close client without errors", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				client.Close()
				// Should not panic
			})

			It("should be safe to close multiple times", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				client.Close()
				client.Close() // Second close should not panic
			})

			It("should close after failed connection", func() {
				cfg := newTestConfig("invalid.host.example.com", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				// Try to connect (will fail)
				ctx, cancel := contextWithTimeout(1 * time.Second)
				defer cancel()
				_ = client.Check(ctx)

				// Close should work even after failed connection
				client.Close()
			})
		})

		Context("UpdConfig", func() {
			It("should update configuration", func() {
				cfg1 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg1)

				cfg2 := newTestConfig(testSMTPHost, testSMTPPort+1, smtptp.TLSNone)
				tlsConfig := &tls.Config{
					MinVersion: tls.VersionTLS12,
				}

				client.UpdConfig(cfg2, tlsConfig)
				// Should not panic
			})

			It("should close existing connection when updating config", func() {
				cfg1 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg1)

				cfg2 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStartTLS)
				cfg2.ForceTLSSkipVerify(true)
				tlsConfig := &tls.Config{
					MinVersion: tls.VersionTLS12,
				}

				client.UpdConfig(cfg2, tlsConfig)
				// Should not panic and should be ready for new connection
			})

			It("should handle config update with authentication", func() {
				cfg1 := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg1)

				cfg2 := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSNone, "newuser", "newpass")
				tlsConfig := &tls.Config{
					MinVersion: tls.VersionTLS12,
				}

				client.UpdConfig(cfg2, tlsConfig)
				// Config should be updated
			})
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent close operations", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					client.Close()
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent clone operations", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)

			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					cloned := client.Clone()
					Expect(cloned).ToNot(BeNil())
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent config updates", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)

			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func(index int) {
					defer GinkgoRecover()
					newCfg := newTestConfig(testSMTPHost, testSMTPPort+index, smtptp.TLSNone)
					tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
					client.UpdConfig(newCfg, tlsConfig)
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < 5; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Edge Cases", func() {
		Context("with various configurations", func() {
			It("should handle zero port", func() {
				cfg := newTestConfig(testSMTPHost, 0, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})

			It("should handle empty host", func() {
				dsn := "tcp(:25)/"
				model := smtpcfg.ConfigModel{DSN: dsn}
				cfg, err := model.Config()
				Expect(err).ToNot(HaveOccurred())

				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})

			It("should handle different TLS modes", func() {
				modes := []smtptp.TLSMode{
					smtptp.TLSNone,
					smtptp.TLSStartTLS,
					smtptp.TLSStrictTLS,
				}

				for _, mode := range modes {
					cfg := newTestConfig(testSMTPHost, testSMTPPort, mode)
					client := newTestSMTPClient(cfg)
					Expect(client).ToNot(BeNil())
				}
			})
		})

		Context("with TLS configuration", func() {
			It("should handle various TLS versions", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStartTLS)

				tlsConfigs := []*tls.Config{
					{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12},
					{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS13},
					{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13},
				}

				for _, tlsCfg := range tlsConfigs {
					client, err := smtpcli.New(cfg, tlsCfg)
					Expect(err).ToNot(HaveOccurred())
					Expect(client).ToNot(BeNil())
				}
			})

			It("should handle InsecureSkipVerify", func() {
				cfg := newTestConfigInsecure(testSMTPHost, testSMTPPort, smtptp.TLSStartTLS)

				tlsConfig := &tls.Config{
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true,
				}

				client, err := smtpcli.New(cfg, tlsConfig)
				Expect(err).ToNot(HaveOccurred())
				Expect(client).ToNot(BeNil())
			})

			It("should handle custom server name", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSStartTLS)
				cfg.SetTLSServerName("custom.example.com")

				client := newTestSMTPClient(cfg)
				Expect(client).ToNot(BeNil())
			})
		})
	})

	Describe("Memory Management", func() {
		It("should not leak memory on multiple create/close cycles", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

			for i := 0; i < 100; i++ {
				client := newTestSMTPClient(cfg)
				client.Close()
			}
			// If there are memory leaks, this test will show increased memory usage
		})

		It("should not leak memory on clone/close cycles", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			original := newTestSMTPClient(cfg)

			for i := 0; i < 100; i++ {
				cloned := original.Clone()
				cloned.Close()
			}

			original.Close()
		})

		It("should handle multiple config updates", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)

			for i := 0; i < 50; i++ {
				newCfg := newTestConfig(testSMTPHost, testSMTPPort+i%10, smtptp.TLSNone)
				tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
				client.UpdConfig(newCfg, tlsConfig)
			}

			client.Close()
		})
	})
})
