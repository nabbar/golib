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

package startStop_test

import (
	"context"
	"time"

	. "github.com/nabbar/golib/runner/startStop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Uptime tests verify that the runner correctly tracks the duration for which
// the service has been running and resets it properly on stop/restart.
var _ = Describe("Uptime", func() {
	Context("Before start", func() {
		// Verify that uptime is zero before the runner is started
		It("should return zero uptime when not started", func() {
			start := func(ctx context.Context) error { return nil }
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)

			Expect(runner.Uptime()).To(BeZero())
		})
	})

	Context("After start", func() {
		// Verify that uptime increases while the runner is running
		It("should track uptime correctly", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Wait a bit and verify uptime is tracking
			time.Sleep(100 * time.Millisecond)
			uptime1 := runner.Uptime()
			Expect(uptime1).To(BeNumerically(">", 0))
			Expect(uptime1).To(BeNumerically(">=", 100*time.Millisecond))

			// Wait more and verify uptime increased
			time.Sleep(100 * time.Millisecond)
			uptime2 := runner.Uptime()
			Expect(uptime2).To(BeNumerically(">", uptime1))

			// Cleanup
			_ = runner.Stop(x)
		})

		// Verify that uptime continues to increase over multiple measurements
		It("should maintain uptime while running", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Check uptime multiple times
			measurements := make([]time.Duration, 5)
			for i := 0; i < 5; i++ {
				time.Sleep(50 * time.Millisecond)
				measurements[i] = runner.Uptime()
			}

			// Each measurement should be greater than the previous
			for i := 1; i < len(measurements); i++ {
				Expect(measurements[i]).To(BeNumerically(">", measurements[i-1]))
			}

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("After stop", func() {
		// Verify that uptime returns to zero after stopping
		It("should reset uptime to zero after stop", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Verify uptime is non-zero
			time.Sleep(100 * time.Millisecond)
			Expect(runner.Uptime()).To(BeNumerically(">", 0))

			// Stop the runner
			err = runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, time.Second).Should(BeFalse())

			// Uptime should be zero
			Eventually(runner.Uptime, time.Second).Should(BeZero())
		})

		// Verify that uptime resets and starts fresh on restart
		It("should track uptime correctly after restart", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			// First run - measure uptime
			runner1 := New(start, stop)
			err := runner1.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner1.IsRunning, 100*time.Millisecond).Should(BeTrue())
			time.Sleep(100 * time.Millisecond)
			uptime1 := runner1.Uptime()
			Expect(uptime1).To(BeNumerically(">", 0))

			// Stop
			err = runner1.Stop(x)
			Expect(err).ToNot(HaveOccurred())
			Eventually(runner1.IsRunning, time.Second).Should(BeFalse())

			// Restart with new runner - uptime should reset
			runner2 := New(start, stop)
			err = runner2.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner2.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// New uptime should start from zero again
			time.Sleep(50 * time.Millisecond)
			uptime2 := runner2.Uptime()
			Expect(uptime2).To(BeNumerically(">", 0))
			Expect(uptime2).To(BeNumerically("<", uptime1))

			// Cleanup
			_ = runner2.Stop(x)
		})
	})

	Context("Quick exit scenarios", func() {
		// Verify uptime handling when the start function completes quickly
		It("should handle uptime when start function exits immediately", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				return nil // Exit immediately
			}
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			// Runner should briefly show as running then stop
			time.Sleep(200 * time.Millisecond)

			// Uptime should eventually be zero
			Eventually(runner.Uptime, time.Second).Should(BeZero())
			Expect(runner.IsRunning()).To(BeFalse())
		})

		// Verify that uptime can measure small time intervals accurately
		It("should handle uptime precision", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Very short sleep
			time.Sleep(10 * time.Millisecond)
			uptime := runner.Uptime()

			// Should be able to measure small durations
			Expect(uptime).To(BeNumerically(">", 0))
			Expect(uptime).To(BeNumerically(">=", 10*time.Millisecond))

			// Cleanup
			_ = runner.Stop(x)
		})
	})
})
