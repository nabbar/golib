/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package fileDescriptor_test

import (
	"context"

	. "github.com/nabbar/golib/ioutils/fileDescriptor"
	. "github.com/onsi/ginkgo/v2"
)

// Global test context for coordinating test lifecycle.
// Initialized before all tests and cancelled after completion.
var (
	testCtx    context.Context
	testCancel context.CancelFunc
)

// initTestContext initializes the global test context.
// Called once at the beginning of the test suite.
func initTestContext() {
	testCtx, testCancel = context.WithCancel(context.Background())
}

// cleanupTestContext cancels the global test context.
// Called once at the end of the test suite.
func cleanupTestContext() {
	if testCancel != nil {
		testCancel()
	}
}

// Helper functions for common test operations

// queryLimits is a helper to query current file descriptor limits.
// Returns current, max, and error. Useful for avoiding repetition in tests.
func queryLimits() (current int, max int, err error) {
	return SystemFileDescriptor(0)
}

// canIncreaseTo checks if the system can potentially increase to the target value.
// Returns true if target is within the hard limit, false otherwise.
// Note: This doesn't guarantee the increase will succeed (may need privileges).
func canIncreaseTo(target int) bool {
	_, max, err := queryLimits()
	if err != nil {
		return false
	}
	return target <= max
}

// attemptIncrease tries to increase the limit to target.
// Returns true if successful, false otherwise (permissions, already at max, etc.).
// Does not fail the test on error - use for optional operations.
func attemptIncrease(target int) bool {
	_, _, err := SystemFileDescriptor(target)
	return err == nil
}

// calculateSafeTarget calculates a safe target for testing increases.
// Returns a value that is:
//   - Higher than current (to actually test an increase)
//   - Within hard limit (so it's theoretically possible)
//   - Returns 0 if no safe target can be calculated
func calculateSafeTarget(increment int) int {
	current, max, err := queryLimits()
	if err != nil {
		return 0
	}

	target := current + increment
	if target > max {
		return 0 // Cannot increase beyond max
	}

	return target
}

// isLimitIncreasePossible checks if any limit increase is possible.
// Returns true if current < max, indicating room for increase.
func isLimitIncreasePossible() bool {
	current, max, err := queryLimits()
	if err != nil {
		return false
	}
	return current < max
}

// skipIfCannotIncrease skips the test if limit increases are not possible.
// Useful for tests that specifically require the ability to increase limits.
func skipIfCannotIncrease() {
	if !isLimitIncreasePossible() {
		Skip("Cannot test: already at maximum limit")
	}
}

// skipIfTargetExceedsMax skips the test if target exceeds system maximum.
func skipIfTargetExceedsMax(target int) {
	if !canIncreaseTo(target) {
		Skip("Cannot test: target exceeds system maximum")
	}
}

// verifyInvariants checks that returned values satisfy basic invariants.
// Useful for validating results in a consistent way across tests.
func verifyInvariants(current, max int, err error) {
	if err != nil {
		// If there's an error, we can't verify invariants
		return
	}

	// Basic sanity checks
	if current <= 0 {
		GinkgoWriter.Printf("WARNING: current limit is %d (expected > 0)\n", current)
	}
	if max < current {
		GinkgoWriter.Printf("WARNING: max (%d) < current (%d)\n", max, current)
	}
}

// getTestContext returns the global test context.
// Useful for operations that need context support.
func getTestContext() context.Context {
	return testCtx
}

// isContextCancelled checks if the test context has been cancelled.
// Useful for long-running tests that should respect cancellation.
func isContextCancelled() bool {
	select {
	case <-testCtx.Done():
		return true
	default:
		return false
	}
}
