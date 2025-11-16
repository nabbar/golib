/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
 */

package queuer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nabbar/golib/mail/queuer"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pooler", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Construction", func() {
		Context("with nil SMTP client", func() {
			It("should create pooler successfully", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should return error when trying to send", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				err := pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
			})

			It("should return error when trying to check", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				err := pooler.Check(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
			})

			It("should return error when trying to get client", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				cli, err := pooler.Client(ctx)
				Expect(err).To(HaveOccurred())
				Expect(cli).To(BeNil())
			})

			It("should not error when trying to reset", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				// Reset should work even without SMTP client as it resets the counter
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with valid SMTP client", func() {
			It("should create pooler with client", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)
				Expect(pooler).ToNot(BeNil())
			})

			It("should clone the SMTP client", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)
				Expect(pooler).ToNot(BeNil())

				// The pooler should work independently
				err = pooler.Check(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Send Operations", func() {
		Context("with throttling enabled", func() {
			It("should send messages respecting throttle limits", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  3,
					Wait: 200 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				// Send up to max messages quickly
				start := time.Now()
				for i := 0; i < 3; i++ {
					err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg%d", i)))
					Expect(err).ToNot(HaveOccurred())
				}
				elapsed := time.Since(start)

				// Should complete quickly
				Expect(elapsed).To(BeNumerically("<", 150*time.Millisecond))
				Expect(backend.msgCount.Load()).To(Equal(int32(3)))
			})

			It("should throttle after exceeding max", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  2,
					Wait: 300 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				// Send max messages
				for i := 0; i < 2; i++ {
					err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg%d", i)))
					Expect(err).ToNot(HaveOccurred())
				}

				// Next message should wait
				start := time.Now()
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("msg3"))
				elapsed := time.Since(start)
				Expect(err).ToNot(HaveOccurred())

				// Should have waited for the duration
				Expect(elapsed).To(BeNumerically(">=", 250*time.Millisecond))
				Expect(backend.msgCount.Load()).To(Equal(int32(3)))
			})
		})

		Context("with throttling disabled", func() {
			It("should send messages without delay when max is 0", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  0,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)

				start := time.Now()
				for i := 0; i < 10; i++ {
					err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg%d", i)))
					Expect(err).ToNot(HaveOccurred())
				}
				elapsed := time.Since(start)

				// Should complete quickly without throttling
				Expect(elapsed).To(BeNumerically("<", 200*time.Millisecond))
				Expect(backend.msgCount.Load()).To(Equal(int32(10)))
			})

			It("should send messages without delay when duration is 0", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 0,
				}

				pooler := queuer.New(cfg, cli)

				start := time.Now()
				for i := 0; i < 10; i++ {
					err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg%d", i)))
					Expect(err).ToNot(HaveOccurred())
				}
				elapsed := time.Since(start)

				// Should complete quickly without throttling
				Expect(elapsed).To(BeNumerically("<", 200*time.Millisecond))
				Expect(backend.msgCount.Load()).To(Equal(int32(10)))
			})
		})

		Context("with various message types", func() {
			It("should handle empty message", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, cli)
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(""))
				Expect(err).ToNot(HaveOccurred())
				Expect(backend.msgCount.Load()).To(Equal(int32(1)))
			})

			It("should handle large message", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  0, // Disable throttling for this test
					Wait: 0,
				}

				// Create large message - using 5KB for faster test execution
				largeContent := make([]byte, 5*1024) // 5KB
				for i := range largeContent {
					largeContent[i] = byte('A' + (i % 26))
				}

				pooler := queuer.New(cfg, cli)
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(string(largeContent)))
				// The main goal is to verify it doesn't error with larger payloads
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle multiple recipients", func() {
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				recipients := []string{
					"to1@test.com",
					"to2@test.com",
					"to3@test.com",
				}

				pooler := queuer.New(cfg, cli)
				err = pooler.Send(ctx, "from@test.com", recipients, newSimpleMessage("test"))
				Expect(err).ToNot(HaveOccurred())
				Expect(backend.msgCount.Load()).To(Equal(int32(1)))
				backend.mu.Lock()
				toLen := len(backend.messages[0].To)
				backend.mu.Unlock()
				Expect(toLen).To(Equal(3))
			})
		})
	})

	Describe("Check Operations", func() {
		It("should verify SMTP server connection", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)
			err = pooler.Check(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when server is down", func() {
			Skip("Test requires server setup refactoring")
		})
	})

	Describe("Client Operations", func() {
		It("should return SMTP client", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			smtpCli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, smtpCli)
			cli, err := pooler.Client(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli).ToNot(BeNil())
		})
	})

	Describe("Clone Operations", func() {
		It("should create independent pooler copy", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  3,
				Wait: 100 * time.Millisecond,
			}

			pooler1 := queuer.New(cfg, cli)
			pooler2 := pooler1.NewPooler()

			Expect(pooler2).ToNot(BeNil())
			Expect(pooler2).ToNot(BeIdenticalTo(pooler1))

			// Both should work independently
			err = pooler1.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("msg1"))
			Expect(err).ToNot(HaveOccurred())

			err = pooler2.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("msg2"))
			Expect(err).ToNot(HaveOccurred())

			Expect(backend.msgCount.Load()).To(Equal(int32(2)))
		})

		It("should preserve throttle configuration", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  2,
				Wait: 200 * time.Millisecond,
			}

			pooler1 := queuer.New(cfg, cli)
			pooler2 := pooler1.NewPooler()

			// Use up pooler1's quota
			for i := 0; i < 2; i++ {
				err = pooler1.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("p1msg%d", i)))
				Expect(err).ToNot(HaveOccurred())
			}

			// pooler2 should still have full quota
			start := time.Now()
			for i := 0; i < 2; i++ {
				err = pooler2.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("p2msg%d", i)))
				Expect(err).ToNot(HaveOccurred())
			}
			elapsed := time.Since(start)

			// pooler2 should work quickly
			Expect(elapsed).To(BeNumerically("<", 100*time.Millisecond))
			Expect(backend.msgCount.Load()).To(Equal(int32(4)))
		})

		It("should use Clone() method from interface", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler1 := queuer.New(cfg, cli)
			pooler2 := pooler1.Clone()

			Expect(pooler2).ToNot(BeNil())
			Expect(pooler2).ToNot(BeIdenticalTo(pooler1))
		})
	})

	Describe("Reset Operations", func() {
		It("should reset counter state", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  2,
				Wait: 300 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			// Use up quota
			for i := 0; i < 2; i++ {
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage(fmt.Sprintf("msg%d", i)))
				Expect(err).ToNot(HaveOccurred())
			}

			// Reset
			err = pooler.Reset()
			Expect(err).ToNot(HaveOccurred())

			// Should work immediately
			start := time.Now()
			err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("msg3"))
			elapsed := time.Since(start)
			Expect(err).ToNot(HaveOccurred())
			Expect(elapsed).To(BeNumerically("<", 50*time.Millisecond))
		})
	})

	Describe("Close Operations", func() {
		It("should close pooler cleanly", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)
			Expect(func() {
				pooler.Close()
			}).ToNot(Panic())
		})

		It("should handle multiple close calls", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)
			Expect(func() {
				pooler.Close()
				pooler.Close()
				pooler.Close()
			}).ToNot(Panic())
		})

		It("should handle close with nil client", func() {
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(func() {
				pooler.Close()
			}).ToNot(Panic())
		})
	})

	Describe("UpdConfig Operations", func() {
		It("should update configuration on existing client", func() {
			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			cli := newTestSMTPClient(host, port)
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, cli)

			// Update config
			newCfg := newTestConfig(host, port, smtptp.TLSNone)
			pooler.UpdConfig(newCfg, nil)

			// Should still work
			err = pooler.Check(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create client when nil", func() {
			cfg := &queuer.Config{
				Max:  10,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)

			backend := &testBackend{}
			srv, host, port, err := startTestSMTPServer(backend, false)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = srv.Close() }()

			// Update config - should create client
			newCfg := newTestConfig(host, port, smtptp.TLSNone)
			pooler.UpdConfig(newCfg, cliTLS.TlsConfig(""))

			// Should now work
			err = pooler.Check(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
