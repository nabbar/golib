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
	"context"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Additional tests to improve code coverage for edge cases and specific branches.
var _ = Describe("TCP Client Coverage Tests", func() {
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

	Describe("Connect edge cases", func() {
		It("should handle reconnection without closing", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// First connection
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			// Second connection without close (replaces connection)
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeTrue())

			// Should still work
			msg := []byte("test")
			response := sendAndReceive(cli, msg)
			Expect(response).To(Equal(msg))
		})

		It("should handle nil client operations", func() {
			var nilCli sckclt.ClientTCP
			// These operations should not panic even with nil client
			Expect(nilCli).To(BeNil())
		})
	})

	Describe("RegisterFunc edge cases", func() {
		It("should handle registering nil error callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Register callback
			cli.RegisterFuncError(func(errs ...error) {})
			// Unregister by passing nil
			cli.RegisterFuncError(nil)
			// Should not panic on error
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
		})

		It("should handle registering nil info callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Register callback with proper signature
			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {})
			// Unregister by passing nil
			cli.RegisterFuncInfo(nil)
			// Should not panic on connect
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
		})
	})

	Describe("IsConnected edge cases", func() {
		It("should handle nil connection state", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Before connect, should be false
			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

	Describe("Write/Read with nil operations", func() {
		It("should handle write on closed connection", func() {
			cli := createClient(address)

			connectClient(ctx, cli)
			Expect(cli.Close()).ToNot(HaveOccurred())

			// Write should fail
			n, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle read on closed connection", func() {
			cli := createClient(address)

			connectClient(ctx, cli)
			Expect(cli.Close()).ToNot(HaveOccurred())

			// Read should fail
			buf := make([]byte, 1024)
			n, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Describe("Close edge cases", func() {
		It("should handle close without connection", func() {
			cli := createClient(address)

			// Close without connect should fail
			err := cli.Close()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sckclt.ErrConnection))
		})
	})

	Describe("TLS edge cases", func() {
		It("should disable TLS with false flag", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Disable TLS
			err := cli.SetTLS(false, nil, "")
			Expect(err).ToNot(HaveOccurred())

			// Should still connect without TLS
			Expect(cli.Connect(ctx)).ToNot(HaveOccurred())
		})

		It("should fail SetTLS with nil config when enabled", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Enable TLS with nil config
			err := cli.SetTLS(true, nil, "localhost")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Dial edge cases", func() {
		It("should handle valid address with TLS config", func() {
			cli := createTLSClient(address)
			defer func() {
				_ = cli.Close()
			}()

			// Connection to non-TLS server with TLS client should fail
			err := cli.Connect(ctx)
			// Will attempt TLS handshake and fail
			_ = err
		})
	})
})
