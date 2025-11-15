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

	. "github.com/nabbar/golib/runner/startStop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Construction tests verify that the StartStop runner can be properly instantiated
// with various combinations of start/stop functions, including nil values.
var _ = Describe("Construction", func() {
	Context("Creating new runner", func() {
		// Verify that a runner can be created with valid functions
		It("should create runner with valid start and stop functions", func() {
			start := func(ctx context.Context) error { return nil }
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)

			Expect(runner).ToNot(BeNil())
			Expect(runner.IsRunning()).To(BeFalse())
			Expect(runner.Uptime()).To(BeZero())
		})

		// Nil start function should be handled gracefully (error generated at runtime)
		It("should create runner with nil start function", func() {
			stop := func(ctx context.Context) error { return nil }

			runner := New(nil, stop)

			Expect(runner).ToNot(BeNil())
			Expect(runner.IsRunning()).To(BeFalse())
		})

		// Nil stop function should be handled gracefully (error generated at runtime)
		It("should create runner with nil stop function", func() {
			start := func(ctx context.Context) error { return nil }

			runner := New(start, nil)

			Expect(runner).ToNot(BeNil())
			Expect(runner.IsRunning()).To(BeFalse())
		})

		// Both nil functions should still create a valid runner instance
		It("should create runner with both nil functions", func() {
			runner := New(nil, nil)

			Expect(runner).ToNot(BeNil())
			Expect(runner.IsRunning()).To(BeFalse())
		})
	})

	// Verify the initial state of a newly created runner
	Context("Initial state", func() {
		var runner StartStop

		BeforeEach(func() {
			// Create a fresh runner for each test
			start := func(ctx context.Context) error { return nil }
			stop := func(ctx context.Context) error { return nil }
			runner = New(start, stop)
		})

		It("should not be running initially", func() {
			Expect(runner.IsRunning()).To(BeFalse())
		})

		It("should have zero uptime initially", func() {
			Expect(runner.Uptime()).To(BeZero())
		})

		It("should have no errors initially", func() {
			Expect(runner.ErrorsLast()).To(BeNil())
			Expect(runner.ErrorsList()).To(BeEmpty())
		})
	})
})
