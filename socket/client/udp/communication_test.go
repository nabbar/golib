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

package udp_test

import (
	"bytes"
	"context"
	"time"

	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// *****
// * By default UDP socket are datagram
// * send and forget
// * so reading must not be tested
// *****

var _ = Describe("UDP Client Communication", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrv.ServerUdp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 30*time.Second)
		address = getTestAddress()
		srv = createSimpleTestServer(ctx, address)
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Write", func() {
		It("should write data successfully", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			data := []byte("Hello, UDP!")
			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should write empty data", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			n, err := cli.Write([]byte{})
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should write large data", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// UDP has packet size limits, use reasonable size
			data := make([]byte, 1024)
			for i := range data {
				data[i] = byte(i % 256)
			}

			n, err := cli.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))
		})

		It("should fail when not connected", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			_, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Read", func() {
		It("should fail when not connected", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			buf := make([]byte, 1024)
			_, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Once", func() {
		It("should work without response callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			request := bytes.NewBufferString("Fire and forget")
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should handle empty request", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			request := bytes.NewBuffer(nil)
			err := cli.Once(ctx, request, nil)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should close connection even on error", func() {
			nonExistentAddr := getTestAddress()
			cli := createClient(nonExistentAddr)
			defer func() {
				_ = cli.Close()
			}()

			request := bytes.NewBufferString("test")
			_ = cli.Once(ctx, request, nil)

			// Should still be closed even if operation failed
			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

})
