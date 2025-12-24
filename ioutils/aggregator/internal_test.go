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
 */

package aggregator_test

import (
	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Internal Improvement", func() {
	var (
		cfg = iotagg.Config{
			FctWriter: func(p []byte) (int, error) { return len(p), nil },
		}
	)

	Describe("TC-IN-001: Context methods with nil internal context", func() {
		It("TC-IN-002: should handle nil internal context gracefully for Deadline", func() {
			agg, err := iotagg.InternalCtx(testCtx, nil, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			dl, ok := agg.Deadline()
			Expect(ok).To(BeFalse())
			Expect(dl).To(BeZero())
		})

		It("TC-IN-003: should handle nil internal context gracefully for Done", func() {
			agg, err := iotagg.InternalCtx(testCtx, nil, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			done := agg.Done()
			Expect(done).ToNot(BeNil())
			// Should be closed immediately
			Eventually(done).Should(BeClosed())
		})

		It("TC-IN-004: should handle nil internal context gracefully for Err", func() {
			agg, err := iotagg.InternalCtx(testCtx, nil, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			err = agg.Err()
			Expect(err).To(BeNil())
		})

		It("TC-IN-005: should handle nil internal context gracefully for Value", func() {
			agg, err := iotagg.InternalCtx(testCtx, nil, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			val := agg.Value("key")
			Expect(val).To(BeNil())
		})
	})

	Describe("TC-IN-006: IsRunning state synchronization", func() {
		It("TC-IN-007: should detect and fix inconsistent state: Runner Running but Op False", func() {
			agg, err := iotagg.New(testCtx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			// Start naturally first to get a valid runner
			Expect(agg.Start(testCtx)).To(Succeed())

			// Simulate inconsistent state: Runner is running (from Start), but we set Op to false
			iotagg.InternalOp(agg, false)

			// IsRunning should detect this and stop the runner to match Op=false
			Expect(agg.IsRunning()).To(BeFalse())

			// Verify runner is stopped
			// We can't easily verify runner is stopped without access to it,
			// but IsRunning returning false suggests it handled it.
			// Let's verify by checking if we can Start again without error
			Expect(agg.Start(testCtx)).To(Succeed())
		})

		It("TC-IN-008: should detect and fix inconsistent state: Runner Stopped but Op True", func() {
			agg, err := iotagg.New(testCtx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			// Start naturally
			Expect(agg.Start(testCtx)).To(Succeed())

			// Stop naturally
			Expect(agg.Stop(testCtx)).To(Succeed())

			// Simulate inconsistent state: Runner is stopped, but we set Op to true
			iotagg.InternalOp(agg, true)

			// IsRunning should detect this and close Op to match Runner=Stopped
			Expect(agg.IsRunning()).To(BeFalse())

			// Verify Op is now false
			Expect(iotagg.InternalGetOp(agg)).To(BeFalse())
		})

		It("TC-IN-009: should handle nil runner with Op True", func() {
			agg, err := iotagg.New(testCtx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			defer func() {
				_ = agg.Close()
			}()

			// Ensure runner is nil (default after New, but let's be explicit)
			iotagg.InternalRunner(agg, nil)
			iotagg.InternalOp(agg, true)

			// IsRunning should detect nil runner but Op=true, and cleanup (Op=false)
			Expect(agg.IsRunning()).To(BeFalse())
			Expect(iotagg.InternalGetOp(agg)).To(BeFalse())
		})
	})
})
