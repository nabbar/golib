/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package idlemgr_test

import (
	"context"
	"time"

	durbig "github.com/nabbar/golib/duration/big"
	idlemgr "github.com/nabbar/golib/socket/idlemgr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Idle Manager", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		manager idlemgr.Manager
		idle    durbig.Duration
		tick    durbig.Duration
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		idle = durbig.Seconds(2)
		tick = durbig.Seconds(1)

		var err error
		manager, err = idlemgr.New(ctx, idle, tick)
		Expect(err).NotTo(HaveOccurred())
		Expect(manager).NotTo(BeNil())

		err = manager.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if manager != nil {
			_ = manager.Close()
		}
		cancel()
	})

	Describe("Basic Operations", func() {
		It("should return uptime", func() {
			Expect(manager.Uptime()).To(BeNumerically(">=", 0))
		})

		It("should report running state", func() {
			Expect(manager.IsRunning()).To(BeTrue())
		})

		It("should handle restart", func() {
			err := manager.Restart(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager.IsRunning()).To(BeTrue())
		})

		It("should handle Register with nil client", func() {
			err := manager.Register(nil)
			Expect(err).To(MatchError("invalid client"))
		})

		It("should handle Unregister with nil client", func() {
			err := manager.Unregister(nil)
			Expect(err).To(MatchError("invalid client"))
		})
	})

	Describe("Idle Monitoring", func() {
		It("should register and monitor a client", func() {
			client := &mockClient{id: "client-1"}
			err := manager.Register(client)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return client.IsClosed()
			}, 4*time.Second, 500*time.Millisecond).Should(BeTrue())
		})

		It("should not close a client if it resets its counter", func() {
			client := &mockClient{id: "client-2"}
			err := manager.Register(client)
			Expect(err).NotTo(HaveOccurred())

			stopReset := make(chan struct{})
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						client.Reset()
					case <-stopReset:
						return
					}
				}
			}()

			time.Sleep(2 * time.Second)
			Expect(client.IsClosed()).To(BeFalse())

			close(stopReset)

			Eventually(func() bool {
				return client.IsClosed()
			}, 4*time.Second, 500*time.Millisecond).Should(BeTrue())
		})

		It("should handle unregistering a client", func() {
			client := &mockClient{id: "client-3"}
			err := manager.Register(client)
			Expect(err).NotTo(HaveOccurred())

			err = manager.Unregister(client)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return client.IsClosed()
			}, 2*time.Second).Should(BeTrue())
		})

		It("should handle multiple clients with different activity", func() {
			client1 := &mockClient{id: "client-4-1"}
			client2 := &mockClient{id: "client-4-2"}

			err := manager.Register(client1)
			Expect(err).NotTo(HaveOccurred())
			err = manager.Register(client2)
			Expect(err).NotTo(HaveOccurred())

			stopReset := make(chan struct{})
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						client1.Reset()
					case <-stopReset:
						return
					}
				}
			}()

			Eventually(func() bool {
				return client2.IsClosed()
			}, 4*time.Second, 500*time.Millisecond).Should(BeTrue())

			Expect(client1.IsClosed()).To(BeFalse())

			close(stopReset)

			Eventually(func() bool {
				return client1.IsClosed()
			}, 4*time.Second, 500*time.Millisecond).Should(BeTrue())
		})
	})

	Describe("Lifecycle", func() {
		It("should stop when context is cancelled", func() {
			cancel()
			Eventually(func() bool {
				return manager.IsRunning()
			}, 2*time.Second).Should(BeFalse())
		})

		It("should handle Stop manually", func() {
			err := manager.Stop(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager.IsRunning()).To(BeFalse())
		})

		It("should handle Start after Stop", func() {
			_ = manager.Stop(ctx)
			err := manager.Start(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager.IsRunning()).To(BeTrue())
		})
	})
})
