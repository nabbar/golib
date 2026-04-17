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

package ticker_test

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsrv "github.com/nabbar/golib/runner"
	. "github.com/nabbar/golib/runner/ticker"
)

var _ = Describe("Lifecycle Operations", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Start", func() {
		It("should start a stopped ticker", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			Expect(tick.IsRunning()).To(BeFalse())

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for the state to transition to running
			Eventually(tick.IsRunning, 100*time.Millisecond, 10*time.Millisecond).Should(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if context is already done", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			cancelCtx, cancelFunc := context.WithCancel(ctx)
			cancelFunc()

			err := tick.Start(cancelCtx)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(context.Canceled))
		})

		It("should report uptime when running", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			time.Sleep(20 * time.Millisecond)
			uptime := tick.Uptime()
			// Uptime should be at least a few milliseconds but less than a large margin
			Expect(uptime).To(BeNumerically(">=", 1*time.Millisecond))
			Expect(uptime).To(BeNumerically("<", 200*time.Millisecond))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop existing instance before starting new one", func() {
			counter := int32(0)
			tick := New(100*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			// Start first time
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(150 * time.Millisecond)

			firstCount := atomic.LoadInt32(&counter)
			Expect(firstCount).To(BeNumerically(">=", int32(1)))

			// Start again - should restart
			err = tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should execute ticker function multiple times", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			// Should have executed about 4 times
			currentCount := counter.Load()
			Expect(currentCount).To(BeNumerically(">=", uint32(2)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Stop", func() {
		It("should stop a running ticker", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeFalse())
			Expect(tick.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should return no error if already stopped", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should stop ticker function execution", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(25 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			countAtStop := atomic.LoadInt32(&counter)
			time.Sleep(20 * time.Millisecond)
			countAfterStop := atomic.LoadInt32(&counter)

			// Counter should not increase after stop
			Expect(countAfterStop).To(Equal(countAtStop))
		})
	})

	Describe("Restart", func() {
		It("should restart a running ticker", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				counter.Add(1)
				return nil
			})

			// Start first time
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(30 * time.Millisecond)

			firstCount := counter.Load()
			firstUptime := tick.Uptime()

			// Restart
			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			// Uptime should reset
			time.Sleep(5 * time.Millisecond)
			newUptime := tick.Uptime()
			Expect(newUptime).To(BeNumerically("<", firstUptime))

			// Should continue ticking
			time.Sleep(100 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">", firstCount))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should start ticker if not running", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				counter.Add(1)
				return nil
			})

			err := tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset ticker uptime on restart", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			oldUptime := tick.Uptime()

			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(15 * time.Millisecond)
			newUptime := tick.Uptime()
			Expect(newUptime).To(BeNumerically("<", oldUptime))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Context Cancellation", func() {
		It("should stop when ticker internal context is cancelled via Stop", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(20 * time.Millisecond)
			Expect(tick.IsRunning()).To(BeTrue())

			// Stop will cancel the internal context
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should stop eventually
			Eventually(tick.IsRunning, 50*time.Millisecond, 3*time.Millisecond).Should(BeFalse())
		})

		// This test is kept for compatibility, but the context of Start
		// is only used for the duration of the Start call itself in the current implementation.
		// If we want the ticker to stop when the Start's context is cancelled,
		// we need to change deMuxStart.
	})
})
