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

package tcp_test

import (
	"bytes"
	"context"
	"io"
	"time"

	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Advanced tests targeting specific uncovered code paths to reach 80% coverage.
var _ = Describe("TCP Client Advanced Coverage", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrt.ServerTcp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		address = getTestAddress()
		srv = createSimpleTestServer(ctx, address)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Once with various scenarios", func() {
		It("should handle Once with nil response callback", func() {
			cli := createClient(address)

			msg := []byte("test message\n")
			request := bytes.NewReader(msg)

			// Once with nil callback should work
			err := cli.Once(ctx, request, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle Once with empty request", func() {
			cli := createClient(address)

			request := bytes.NewReader([]byte{})

			err := cli.Once(ctx, request, func(r io.Reader) {
				// Do nothing
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle Once with large request", func() {
			cli := createClient(address)

			// Create large request
			largeData := bytes.Repeat([]byte("x"), 10*1024)
			request := bytes.NewReader(largeData)

			received := make([]byte, 0)
			err := cli.Once(ctx, request, func(r io.Reader) {
				buf := make([]byte, 1024)
				for {
					n, err := r.Read(buf)
					if n > 0 {
						received = append(received, buf[:n]...)
					}
					if err != nil {
						break
					}
				}
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(received)).To(BeNumerically(">=", len(largeData)))
		})

		It("should handle Once with context timeout", func() {
			cli := createClient(address)

			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Millisecond)
			defer timeoutCancel()

			msg := []byte("test")
			request := bytes.NewReader(msg)

			// Very short timeout might cause error
			err := cli.Once(timeoutCtx, request, nil)
			// May or may not error depending on timing
			_ = err
		})
	})

	Describe("Multiple rapid connections", func() {
		It("should handle rapid reconnection cycles", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Rapid connect/disconnect cycles
			for i := 0; i < 5; i++ {
				err := cli.Connect(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Small delay
				time.Sleep(10 * time.Millisecond)

				// Send data to verify connection
				msg := []byte("test\n")
				n, err := cli.Write(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(msg)))

				// Close
				err = cli.Close()
				Expect(err).ToNot(HaveOccurred())

				// Small delay before next cycle
				time.Sleep(10 * time.Millisecond)
			}
		})

		It("should handle connection replacement by reconnecting", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// First connection
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Get connection state
			Expect(cli.IsConnected()).To(BeTrue())

			// Reconnect without closing (replaces connection)
			err = cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			// Third reconnection
			err = cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			// Verify still works
			msg := []byte("test\n")
			response := sendAndReceive(cli, msg)
			Expect(response).To(Equal(msg))
		})
	})

	Describe("Write edge cases", func() {
		It("should handle large writes", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Large write
			largeData := bytes.Repeat([]byte("x"), 100*1024)
			n, err := cli.Write(largeData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(largeData)))
		})

		It("should handle write after server closes", func() {
			// Server that closes immediately
			closeAddr := getTestAddress()
			closeSrv := createAndRegisterServer(closeAddr, closingHandler)
			startServer(ctx, closeSrv)
			defer func() {
				_ = closeSrv.Shutdown(ctx)
			}()
			waitForServerRunning(closeAddr, 2*time.Second)

			cli := createClient(closeAddr)
			defer func() {
				_ = cli.Close()
			}()

			// Connect - server will close immediately
			_ = cli.Connect(ctx)
			time.Sleep(200 * time.Millisecond)

			// Try to write - should fail
			msg := []byte("test\n")
			_, err := cli.Write(msg)
			// May succeed or fail depending on timing
			_ = err
		})
	})

	Describe("Read edge cases", func() {
		It("should handle partial reads", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Send data
			msg := []byte("Hello, World!\n")
			_, err := cli.Write(msg)
			Expect(err).ToNot(HaveOccurred())

			// Read in small chunks
			buf := make([]byte, 5)
			var received []byte
			for len(received) < len(msg) {
				n, err := cli.Read(buf)
				if n > 0 {
					received = append(received, buf[:n]...)
				}
				if err != nil {
					break
				}
			}
			Expect(received).To(Equal(msg))
		})

		It("should handle concurrent reads and writes", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			done := make(chan bool, 2)

			// Writer goroutine
			go func() {
				defer GinkgoRecover()
				for i := 0; i < 10; i++ {
					msg := []byte("message\n")
					_, err := cli.Write(msg)
					if err != nil {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				done <- true
			}()

			// Reader goroutine
			go func() {
				defer GinkgoRecover()
				buf := make([]byte, 1024)
				for i := 0; i < 10; i++ {
					_, err := cli.Read(buf)
					if err != nil {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				done <- true
			}()

			// Wait for both with timeout
			timeout := time.After(5 * time.Second)
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					// Success
				case <-timeout:
					Fail("Timeout waiting for concurrent operations")
				}
			}
		})
	})

	Describe("TLS with actual server", func() {
		It("should connect to TLS server with valid config", func() {
			// Create TLS server
			tlsAddr := getTestAddress()
			tlsSrv := createTLSServer(tlsAddr, echoHandler)
			startServer(ctx, tlsSrv)
			defer func() {
				_ = tlsSrv.Shutdown(ctx)
			}()
			waitForServerRunning(tlsAddr, 2*time.Second)

			// Create TLS client
			cli := createTLSClient(tlsAddr)
			defer func() {
				_ = cli.Close()
			}()

			// Connect with TLS
			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Send encrypted data
			msg := []byte("encrypted message\n")
			response := sendAndReceive(cli, msg)
			Expect(response).To(Equal(msg))
		})

		It("should disable TLS after it was enabled", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Enable TLS
			_, clientConfig := createTLSConfig()
			err := cli.SetTLS(true, clientConfig, "localhost")
			Expect(err).ToNot(HaveOccurred())

			// Disable TLS
			err = cli.SetTLS(false, nil, "")
			Expect(err).ToNot(HaveOccurred())

			// Should connect without TLS
			err = cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Verify works
			msg := []byte("test\n")
			response := sendAndReceive(cli, msg)
			Expect(response).To(Equal(msg))
		})
	})
})
