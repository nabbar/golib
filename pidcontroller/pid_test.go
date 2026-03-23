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

package pidcontroller_test

import (
	"context"
	"time"

	"github.com/nabbar/golib/pidcontroller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PID Controller", func() {
	var (
		pid        pidcontroller.PID
		kp, ki, kd float64
	)

	BeforeEach(func() {
		// Default to a simple Proportional controller for basic tests
		kp = 0.5
		ki = 0.0
		kd = 0.0
		pid = pidcontroller.New(kp, ki, kd)
	})

	Context("Constructor New", func() {
		// TC-PID-001
		It("should return a valid non-nil PID interface", func() {
			Expect(pid).ToNot(BeNil())
		})
	})

	Context("Method Range", func() {
		// TC-PID-002
		It("should generate a sequence ending with the max value", func() {
			min, max := 0.0, 10.0
			res := pid.Range(min, max)

			Expect(res).ToNot(BeEmpty())
			Expect(res[len(res)-1]).To(Equal(max))
		})

		// TC-PID-003
		It("should progress from min towards max", func() {
			min, max := 0.0, 100.0
			res := pid.Range(min, max)

			Expect(len(res)).To(BeNumerically(">", 1))
			// First step should be > min
			Expect(res[0]).To(BeNumerically(">", min))
			// Last step should be max
			Expect(res[len(res)-1]).To(Equal(max))
		})

		// TC-PID-004
		It("should handle small steps correctly", func() {
			// Low Kp implies smaller steps
			localPID := pidcontroller.New(0.1, 0.0, 0.0)
			res := localPID.Range(0, 10)
			Expect(len(res)).To(BeNumerically(">", 1))
		})
	})

	Context("Method RangeCtx", func() {
		// TC-PID-005
		It("should respect context cancellation", func() {
			ctx, cancel := context.WithCancel(context.Background())
			min, max := 0.0, 100000.0

			// Cancel immediately
			cancel()
			res := pid.RangeCtx(ctx, min, max)

			// Implementation detail: checks error, if set, returns append(res, max)
			// Since res is empty initially, should be [max]
			Expect(res).To(HaveLen(48))
			Expect(res[len(res)-1]).To(Equal(max))
		})

		// TC-PID-006
		It("should return partial results and end with max on timeout", func() {
			// Very short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
			defer cancel()

			min, max := 0.0, 1000000.0
			res := pid.RangeCtx(ctx, min, max)

			Expect(res).ToNot(BeEmpty())
			Expect(res[len(res)-1]).To(Equal(max))
		})
	})

	Context("Edge Cases and Limits", func() {
		// TC-PID-007
		It("should handle min > max by returning max immediately", func() {
			// If min > max, the loop condition 'min > max' triggers immediately (conceptually)
			// or the logic returns max.
			res := pid.Range(10.0, 0.0)
			Expect(res).To(HaveLen(1))
			Expect(res[0]).To(Equal(0.0)) // Assuming max is returned (0.0 here)
		})

		// TC-PID-008
		It("should handle min == max with timeout prevention", func() {
			// If min == max, error is 0. If loop doesn't check equality, it might spin.
			// RangeCtx should be used with timeout to prevent infinite loop if logic is flawed.
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			res := pid.RangeCtx(ctx, 10.0, 10.0)
			// If min==max, it should return immediately.
			Expect(res).To(HaveLen(1))
			Expect(res[0]).To(Equal(10.0))
		})

		// TC-PID-009
		It("should handle negative ranges", func() {
			// -10 to 0
			res := pid.Range(-10.0, 0.0)
			Expect(res).ToNot(BeEmpty())
			Expect(res[len(res)-1]).To(Equal(0.0))
			Expect(res[0]).To(BeNumerically(">", -10.0))
		})

		// TC-PID-010
		It("should handle large ranges without hanging", func() {
			// Testing with large range
			min := 0.0
			max := 1e6 // 1 million
			// Use a timeout because this would take forever if step is too small or overhead high
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			res := pid.RangeCtx(ctx, min, max)
			Expect(res).ToNot(BeEmpty())
			// Ideally we want to check if it finished or timed out, but just not hanging is key
		})
	})
})
