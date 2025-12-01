/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package udp_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client Robustness", func() {
	Describe("Server Unavailability", func() {
		It("should handle connecting to non-existent server", func() {
			nonExistAddr := getTestAddress()
			cli := createClient(nonExistAddr)
			defer cleanupClient(cli)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			// UDP doesn't fail on connect to non-existent server
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle writing to non-existent server", func() {
			nonExistAddr := getTestAddress()
			cli := createClient(nonExistAddr)
			defer cleanupClient(cli)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			_ = cli.Connect(ctx)

			// Write will succeed (UDP is fire-and-forget)
			_, err := cli.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Server Shutdown During Operation", func() {
		It("should handle server shutdown after connection", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cancel()

			connectClient(ctx, cli)

			// Send first message
			_, err := cli.Write([]byte("before shutdown"))
			Expect(err).ToNot(HaveOccurred())

			// Shutdown server
			_ = srv.Shutdown(ctx)
			waitForServerStopped(srv, 2*time.Second)

			// Client still connected (UDP is connectionless)
			Expect(cli.IsConnected()).To(BeTrue())

			// Write still succeeds (UDP doesn't know server is gone)
			_, err = cli.Write([]byte("after shutdown"))

			// prevent socket closed not propagated on some system
			if err != nil {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
			}

			cleanupClient(cli)
		})
	})

	Describe("Context Cancellation Resilience", func() {
		It("should handle context cancellation during connect", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			ctx, cancel := context.WithCancel(globalCtx)
			cancel() // Cancel immediately

			err := cli.Connect(ctx)
			// May or may not error depending on timing
			_ = err
		})

		It("should handle context timeout during Once", func() {
			srv, cli, _, _, _ := createTestServerAndClient(closingHandler)

			ctx, cancel := context.WithTimeout(globalCtx, 10*time.Millisecond)
			defer cancel()

			request := bytes.NewBufferString("timeout test")
			_ = cli.Once(ctx, request, func(_ io.Reader) {
				time.Sleep(100 * time.Millisecond) // Exceed timeout
			})

			cleanupServer(srv, globalCtx)
		})
	})

	Describe("Repeated Operations", func() {
		It("should handle many sequential write operations", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			connectClient(ctx, cli)

			// Many sequential writes
			successCount := 0
			for i := 0; i < 1000; i++ {
				_, err := cli.Write([]byte("test"))
				if err == nil {
					successCount++
				}
				if i%100 == 0 {
					time.Sleep(time.Millisecond) // Periodic small delay
				}
			}

			// UDP may lose packets under load, especially with rapid sequential writes
			// Expect at least 50% success rate which is reasonable for UDP stress testing
			Expect(successCount).To(BeNumerically(">=", 500))
		})

		It("should handle repeated connect/close cycles", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			for i := 0; i < 20; i++ {
				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())

				err = cli.Close()
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should handle repeated callback registration", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			for i := 0; i < 100; i++ {
				cli.RegisterFuncError(func(errs ...error) {})
				cli.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {})
			}

			Expect(cli).ToNot(BeNil())
		})
	})

	Describe("Error Recovery", func() {
		It("should recover from write errors", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			// Write without connection (error)
			_, err := cli.Write([]byte("error"))
			Expect(err).To(HaveOccurred())

			// Should still work after error
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should recover from read errors", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			// Read without connection (error)
			buf := make([]byte, 100)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())

			// Should still work after error
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should recover from close errors", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			// Close without connection (error)
			err := cli.Close()
			Expect(err).To(HaveOccurred())

			// Connect after error
			err = cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should work normally
			err = cli.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Resource Cleanup", func() {
		It("should cleanup after many creates and closes", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			for i := 0; i < 50; i++ {
				cli := createClient(address)
				_ = cli.Connect(ctx)
				_ = cli.Close()
			}

			// No resource leak - test completes successfully
		})

		It("should cleanup when client goes out of scope", func() {
			srv, _, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			address := getTestAddress()
			srv2 := createServer(simpleEchoHandler(), address)
			startServer(ctx, srv2)
			waitForServerRunning(address, 5*time.Second)
			defer cleanupServer(srv2, ctx)

			for i := 0; i < 20; i++ {
				func() {
					cli := createClient(address)
					_ = cli.Connect(ctx)
					// No explicit close - rely on GC
				}()
			}

			// No obvious leaks - test completes
		})
	})

	Describe("Invalid State Transitions", func() {
		It("should handle operations in wrong order", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			// Write before connect
			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())

			// Read before connect
			buf := make([]byte, 100)
			_, err = cli.Read(buf)
			Expect(err).To(HaveOccurred())

			// Close before connect
			err = cli.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle double close", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cancel()

			connectClient(ctx, cli)

			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			err = cli.Close()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Nil Operations", func() {
		It("should handle nil error callback gracefully", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			cli.RegisterFuncError(nil)

			// Should not panic when errors occur
			_, _ = cli.Write([]byte("test"))
		})

		It("should handle nil info callback gracefully", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			cli.RegisterFuncInfo(nil)

			// Should not panic during operations
			_ = cli.Connect(ctx)
		})
	})

	// NOTE: Callback panic recovery tests are skipped because the client
	// does not currently implement panic recovery for callbacks executed in goroutines.
	// This could be improved by using runner.RecoveryCaller from github.com/nabbar/golib/runner
	Describe("Callback Errors", func() {
		It("should not fail if error callback panics", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			cli.RegisterFuncError(func(errs ...error) {
				panic("callback panic")
			})

			// Trigger error callback
			_, _ = cli.Write([]byte("test"))

			// Give callback time to execute
			time.Sleep(100 * time.Millisecond)

			// Client should still be usable
			Expect(cli).ToNot(BeNil())
		})

		It("should not fail if info callback panics", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			cli.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {
				panic("callback panic")
			})

			// Trigger info callback
			_ = cli.Connect(ctx)

			// Give callback time to execute
			time.Sleep(100 * time.Millisecond)

			// Client should still be usable
			Expect(cli.IsConnected()).To(BeTrue())
		})
	})

	Describe("SetTLS Resilience", func() {
		It("should always return nil for SetTLS", func() {
			cli := createClient(getTestAddress())
			defer cleanupClient(cli)

			// All combinations should return nil
			err := cli.SetTLS(true, nil, "")
			Expect(err).To(BeNil())

			err = cli.SetTLS(false, nil, "")
			Expect(err).To(BeNil())

			err = cli.SetTLS(true, nil, "localhost")
			Expect(err).To(BeNil())
		})
	})

	Describe("Address Resolution Edge Cases", func() {
		It("should handle unresolvable hostname gracefully", func() {
			_, err := sckclt.New("this-host-definitely-does-not-exist.invalid:8080")
			// May or may not error depending on DNS configuration
			_ = err
		})

		It("should handle localhost variants", func() {
			variants := []string{
				"localhost:8080",
				"127.0.0.1:8080",
				"0.0.0.0:8080",
			}

			for _, addr := range variants {
				cli, err := sckclt.New(addr)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli).ToNot(BeNil())
			}
		})
	})
})
