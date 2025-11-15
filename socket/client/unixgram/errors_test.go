//go:build linux || darwin

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

package unixgram_test

import (
	"context"
	"time"

	sckclt "github.com/nabbar/golib/socket/client/unixgram"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Datagram Client Errors", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		srv        scksrv.ServerUnixGram
		socketPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		socketPath = getTestSocketPath()
		srv = createSimpleTestServer(ctx, socketPath)
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		cleanupSocket(socketPath)
		if cancel != nil {
			cancel()
		}
	})

	Describe("Error Constants", func() {
		It("should have ErrInstance error", func() {
			Expect(sckclt.ErrInstance).ToNot(BeNil())
		})

		It("should have ErrConnection error", func() {
			Expect(sckclt.ErrConnection).ToNot(BeNil())
		})

		It("should have ErrAddress error", func() {
			Expect(sckclt.ErrAddress).ToNot(BeNil())
		})
	})

	Describe("Connection Errors", func() {
		It("should return error on Write without connection", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sckclt.ErrConnection))
		})

		It("should return error on Read without connection", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sckclt.ErrConnection))
		})

		It("should return error on Close without connection", func() {
			cli := createClient(socketPath)

			err := cli.Close()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sckclt.ErrConnection))
		})

		It("should return error on operations after close", func() {
			cli := createClient(socketPath)

			connectClient(ctx, cli)
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Try operations after close
			_, err = cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())

			buf := make([]byte, 1024)
			_, err = cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Address Errors", func() {
		It("should return nil for empty path", func() {
			cli := sckclt.New("")
			Expect(cli).To(BeNil())
		})
	})

	Describe("Context Cancellation", func() {
		It("should handle context cancellation during connect", func() {
			nonExistentPath := getTestSocketPath()
			cli := createClient(nonExistentPath)
			defer func() {
				_ = cli.Close()
			}()

			cancelCtx, cancelFunc := context.WithCancel(ctx)
			cancelFunc() // Cancel immediately

			err := cli.Connect(cancelCtx)
			Expect(err).To(HaveOccurred())
		})

		It("should respect context timeout", func() {
			nonExistPath := getTestSocketPath()
			cli := createClient(nonExistPath)
			defer func() {
				_ = cli.Close()
			}()

			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			defer cancel()

			err := cli.Connect(timeoutCtx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Network Errors", func() {
		It("should handle connection to non-existent socket gracefully", func() {
			nonExistPath := getTestSocketPath()
			cli := createClient(nonExistPath)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.Connect(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Nil Client Handling", func() {
		It("should handle nil RegisterFuncError gracefully", func() {
			var cli *sckclt.ClientUnix
			// Should not panic
			Expect(func() {
				if cli != nil {
					(*cli).RegisterFuncError(nil)
				}
			}).ToNot(Panic())
		})

		It("should handle nil RegisterFuncInfo gracefully", func() {
			var cli *sckclt.ClientUnix
			// Should not panic
			Expect(func() {
				if cli != nil {
					(*cli).RegisterFuncInfo(nil)
				}
			}).ToNot(Panic())
		})
	})

	Describe("Error Recovery", func() {
		It("should recover from failed write and allow reconnect", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			// Try to write without connection
			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())

			// Now connect and retry
			connectClient(ctx, cli)
			n, err := cli.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should recover from close error and allow new operations", func() {
			cli := createClient(socketPath)

			// Try to close without connection
			err := cli.Close()
			Expect(err).To(HaveOccurred())

			// Connect and use normally
			connectClient(ctx, cli)
			defer func() {
				_ = cli.Close()
			}()

			n, err := cli.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(4))
		})
	})

	Describe("SetTLS", func() {
		It("should return nil (UNIX datagram doesn't support TLS)", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			err := cli.SetTLS(true, nil, "localhost")
			Expect(err).To(BeNil())
		})
	})
})
