//go:build linux || darwin

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

package unix_test

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"time"

	scksrv "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Client Communication", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		srv        scksrv.ServerUnix
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

	Describe("Write", func() {
		It("should write data successfully", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			data := []byte("Hello, UNIX!")
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should write empty data", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			n, err := cli.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should fail when not connected", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Read", func() {
		It("should read echoed data", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			data := []byte("Echo test")
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			// Read response
			buf := make([]byte, 1024)
			n, err = cli.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
			Expect(buf[:n]).To(Equal(data))
		})

		It("should handle multiple read/write cycles", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			for i := 0; i < 5; i++ {
				data := []byte("Message " + string(rune('0'+i)))

				n, err := cli.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))

				buf := make([]byte, 1024)
				n, err = cli.Read(buf)
				Expect(err).ToNot(HaveOccurred())
				Expect(buf[:n]).To(Equal(data))
			}
		})

		It("should fail when not connected", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Once", func() {
		It("should perform request/response operation", func() {
			cli := createClient(socketPath)

			request := bytes.NewBufferString("Once test")
			var response []byte

			err := cli.Once(ctx, request, func(reader io.Reader) {
				response, _ = io.ReadAll(reader)
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(Equal([]byte("Once test")))

			// Connection should be closed after Once
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should work without response callback", func() {
			cli := createClient(socketPath)

			request := bytes.NewBufferString("Fire and forget")
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should handle empty request", func() {
			cli := createClient(socketPath)

			request := bytes.NewBuffer(nil)
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should close connection even on error", func() {
			nonExistentPath := getTestSocketPath()
			cli := createClient(nonExistentPath)

			request := bytes.NewBufferString("test")
			_ = cli.Once(ctx, request, nil)

			// Should still be closed even if operation failed
			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle multiple writes from same client", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func(id int) {
					defer GinkgoRecover()
					data := []byte("Concurrent " + string(rune('0'+id)))
					_, err := cli.Write(data)
					Expect(err).ToNot(HaveOccurred())
					done <- true
				}(i)
			}

			// Wait for all writes
			for i := 0; i < 5; i++ {
				Eventually(done, 2*time.Second).Should(Receive())
			}
		})

		It("should handle multiple concurrent clients", func() {
			counterPath := getTestSocketPath()
			cleanupSocket(counterPath)
			defer cleanupSocket(counterPath)

			counter := new(atomic.Int32)
			counterSrv := createAndRegisterServer(counterPath, countingHandler(counter))
			startServer(ctx, counterSrv)
			defer counterSrv.Shutdown(ctx)
			waitForServerRunning(counterPath, 5*time.Second)

			numClients := 10
			done := make(chan bool, numClients)

			for i := 0; i < numClients; i++ {
				go func() {
					defer GinkgoRecover()
					cli := createClient(counterPath)
					defer func() {
						_ = cli.Close()
					}()

					connectClient(ctx, cli)

					data := []byte("test")
					_, err := cli.Write(data)
					Expect(err).ToNot(HaveOccurred())

					// Read response
					buf := make([]byte, 1024)
					_, _ = cli.Read(buf)

					done <- true
				}()
			}

			// Wait for all clients
			for i := 0; i < numClients; i++ {
				Eventually(done, 5*time.Second).Should(Receive())
			}

			// Allow time for server to process
			time.Sleep(100 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">=", int32(numClients)))
		})
	})

	Describe("Binary Data", func() {
		It("should handle binary data correctly", func() {
			cli := createClient(socketPath)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Binary data with all byte values
			data := make([]byte, 256)
			for i := 0; i < 256; i++ {
				data[i] = byte(i)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			buf := make([]byte, 512)
			n, err = cli.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(buf[:n]).To(Equal(data))
		})
	})
})
