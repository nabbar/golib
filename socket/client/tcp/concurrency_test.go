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
	"io"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Concurrency", func() {
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
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("Concurrent Connections", func() {
		It("should handle multiple clients concurrently", func() {
			var (
				numClients = 10
				clients    = libatm.NewMapTyped[int, io.Closer]()
				wg         sync.WaitGroup
				msg        = []byte("concurrent test\n")
			)

			for i := 0; i < numClients; i++ {
				wg.Go(func() {
					defer GinkgoRecover()

					cli := createClient(address)
					connectClient(ctx, cli)
					Expect(cli.IsConnected()).To(BeTrue())

					// Send test message
					Expect(sendAndReceive(cli, msg)).To(Equal(msg))
					clients.Store(i, cli)
				})
			}
			wg.Wait()
			time.Sleep(50 * time.Millisecond)

			// Cleanup
			clients.Range(func(_ int, cli io.Closer) bool {
				Expect(cli.Close()).ToNot(HaveOccurred())
				return true
			})
		})

		It("should maintain connection integrity under concurrent load", func() {
			var (
				nbClient = 5
				msgByCli = 20
				errCount = new(atomic.Int32)

				msg = []byte("test\n")
				wg  sync.WaitGroup
			)

			for i := 0; i < nbClient; i++ {
				wg.Go(func() {
					defer GinkgoRecover()

					cli := createClient(address)
					defer func() {
						_ = cli.Close()
					}()

					connectClient(ctx, cli)

					for j := 0; j < msgByCli; j++ {
						if string(sendAndReceive(cli, msg)) != string(msg) {
							errCount.Add(1)
						}
					}
				})
			}

			wg.Wait()
			time.Sleep(50 * time.Millisecond)

			Expect(errCount.Load()).To(Equal(int32(0)))
		})
		It("should handle rapid connection creation and destruction", func() {
			var (
				nbIte = 20
				msg   = []byte("quick test\n")
				wg    sync.WaitGroup
			)

			for i := 0; i < nbIte; i++ {
				wg.Go(func() {
					defer GinkgoRecover()

					cli := createClient(address)
					defer func() {
						_ = cli.Close()
					}()

					err := cli.Connect(ctx)
					Expect(err).ToNot(HaveOccurred())
					_, e := cli.Write(msg)
					Expect(e).ToNot(HaveOccurred())
				})
			}

			wg.Wait()
			time.Sleep(50 * time.Millisecond)
		})
	})
})

// Helper type for Once tests
type fixedReader struct {
	data []byte
	pos  int
}

func (f *fixedReader) Read(p []byte) (n int, err error) {
	if f.pos >= len(f.data) {
		return 0, nil
	}
	n = copy(p, f.data[f.pos:])
	f.pos += n
	return n, nil
}
