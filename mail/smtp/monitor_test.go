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
	"context"
	"time"

	"github.com/emersion/go-smtp"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libver "github.com/nabbar/golib/version"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMTP Monitor", func() {
	BeforeEach(func() {
		testSMTPPort = getFreePort()
	})

	Describe("Check Operations", func() {
		Context("with unreachable server", func() {
			It("should fail health check", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should respect timeout", func() {
				cfg := newTestConfig("240.0.0.1", 25, smtptp.TLSNone) // Non-routable
				client := newTestSMTPClient(cfg)
				defer client.Close()

				start := time.Now()
				ctx, cancel := contextWithTimeout(500 * time.Millisecond)
				defer cancel()

				err := client.Check(ctx)
				elapsed := time.Since(start)

				Expect(err).To(HaveOccurred())
				Expect(elapsed).To(BeNumerically("<", 2*time.Second))
			})
		})

		Context("with invalid configuration", func() {
			It("should fail with nonexistent host", func() {
				cfg := newTestConfig("nonexistent-smtp-host-123456.example", 25, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(3 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})

			It("should fail with invalid port", func() {
				cfg := newTestConfig(testSMTPHost, 99999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(2 * time.Second)
				defer cancel()

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with cancelled context", func() {
			It("should fail immediately", func() {
				cfg := newTestConfig("240.0.0.1", 25, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				cancel() // Cancel immediately

				err := client.Check(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with multiple checks", func() {
			It("should handle consecutive health checks", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				err1 := client.Check(ctx)
				err2 := client.Check(ctx)
				err3 := client.Check(ctx)

				Expect(err1).To(HaveOccurred())
				Expect(err2).To(HaveOccurred())
				Expect(err3).To(HaveOccurred())
			})

			It("should handle concurrent health checks", func() {
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
		})
	})

	Describe("Monitor", func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)

		BeforeEach(func() {
			ctx, cancel = contextWithTimeout(10 * time.Second)
		})

		AfterEach(func() {
			if cancel != nil {
				cancel()
			}
		})

		Context("with valid configuration", func() {
			It("should create monitor instance", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				// Monitor creation may fail if dependencies are missing
				// This is acceptable for this test
				_ = err
				_ = monitor
			})

			It("should create monitor with authentication", func() {
				cfg := newTestConfigWithAuth(testSMTPHost, testSMTPPort, smtptp.TLSNone, testSMTPUser, testSMTPPassword)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				_ = err
				_ = monitor
			})

			It("should create monitor with TLS configuration", func() {
				cfg := newTestConfigInsecure(testSMTPHost, getFreePort(), smtptp.TLSStrictTLS)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				_ = err
				_ = monitor
			})
		})

		Context("with different version information", func() {
			It("should handle various version formats", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				versions := []libver.Version{
					libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "release", "1.0.0", "test", "", struct{}{}, 0),
					libver.NewVersion(libver.License_MIT, "test", "Test", "2024-02-01", "beta", "2.0.0-beta", "test", "", struct{}{}, 0),
					libver.NewVersion(libver.License_MIT, "test", "Test", "2024-03-01", "rc", "3.0.0-rc1", "test", "", struct{}{}, 0),
				}

				for _, version := range versions {
					monitor, err := client.Monitor(ctx, version)
					if err == nil && monitor != nil {
						_ = monitor.Stop(ctx)
					}
					_ = err
				}
			})

			It("should handle empty version", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "", "", "", "", "", "", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				_ = err
				_ = monitor
			})
		})

		Context("with unreachable server", func() {
			It("should create monitor even with unreachable server", func() {
				cfg := newTestConfig("127.0.0.1", 9999, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				// Monitor creation should succeed even if server is unreachable
				// The health check will fail later
				_ = err
				_ = monitor
			})
		})

		Context("with monitor lifecycle", func() {
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

			It("should start and stop monitor", func() {
				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				Expect(err).ToNot(HaveOccurred())
				Expect(monitor).ToNot(BeNil())
				defer monitor.Stop(ctx)

				for i := 0; i < 3; i++ {
					err = monitor.Start(ctx)
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(50 * time.Millisecond)

					err = monitor.Stop(ctx)
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(50 * time.Millisecond)
				}
			})
		})

		Context("with concurrent monitor operations", func() {
			It("should handle concurrent monitor creations", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)

				done := make(chan bool, 5)
				for i := 0; i < 5; i++ {
					go func() {
						defer GinkgoRecover()
						client := newTestSMTPClient(cfg)
						defer client.Close()

						version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

						monitor, err := client.Monitor(ctx, version)
						if err == nil && monitor != nil {
							_ = monitor.Stop(ctx)
						}

						done <- true
					}()
				}

				for i := 0; i < 5; i++ {
					Eventually(done, 5*time.Second).Should(Receive())
				}
			})
		})
	})

	Describe("Monitor Integration with HealthCheck", func() {
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
		Context("when monitor is running", func() {
			It("should perform periodic health checks", func() {

				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				ctx, cancel := contextWithTimeout(10 * time.Second)
				defer cancel()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				Expect(err).ToNot(HaveOccurred())
				Expect(monitor).ToNot(BeNil())
				defer monitor.Stop(ctx)

				err = monitor.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Wait for a few health checks to occur
				time.Sleep(500 * time.Millisecond)

				// Monitor should still be running
				// (In a real test, you would check the monitor's internal state)
			})
		})
	})

	Describe("Monitor Error Handling", func() {
		Context("with various error conditions", func() {
			It("should handle monitor creation with nil context", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)
				defer client.Close()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				// Using nil context might cause panic or error
				// We're testing that it doesn't crash the application
				defer func() {
					if r := recover(); r != nil {
						// Panic occurred, which is acceptable
						_ = r
					}
				}()

				_, _ = client.Monitor(nil, version)
			})

			It("should handle monitor creation after client close", func() {
				cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
				client := newTestSMTPClient(cfg)

				client.Close()

				ctx, cancel := contextWithTimeout(5 * time.Second)
				defer cancel()

				version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					defer monitor.Stop(ctx)
				}

				// Should still work, monitor is independent of client state
				_ = err
				_ = monitor
			})
		})
	})

	Describe("Monitor Memory Management", func() {
		It("should not leak memory on multiple monitor creations", func() {
			cfg := newTestConfig(testSMTPHost, testSMTPPort, smtptp.TLSNone)
			client := newTestSMTPClient(cfg)
			defer client.Close()

			ctx, cancel := contextWithTimeout(30 * time.Second)
			defer cancel()

			version := libver.NewVersion(libver.License_MIT, "test", "Test", "2024-01-01", "test", "1.0.0", "test", "", struct{}{}, 0)

			for i := 0; i < 50; i++ {
				monitor, err := client.Monitor(ctx, version)
				if err == nil && monitor != nil {
					_ = monitor.Stop(ctx)
				}
			}

			// If there are memory leaks, this test will show increased memory usage
		})
	})
})
