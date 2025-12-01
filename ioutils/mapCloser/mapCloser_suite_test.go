/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package mapCloser_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestMapCloser is the entry point for the Ginkgo test suite.
// It integrates Ginkgo BDD tests with Go's standard testing framework.
//
// The test suite covers:
//   - Basic operations (create, add, get, clean, close)
//   - Clone operations (independent copies)
//   - Error handling (aggregation, fail-safe behavior)
//   - Context cancellation (automatic cleanup)
//   - Concurrency (thread-safety verification)
//   - Edge cases (nil handling, double close, etc.)
//   - Performance (large scale, high concurrency)
//
// Run with: go test -v
// Run with race detector: go test -v -race
// Run with coverage: go test -v -coverprofile=coverage.out
func TestMapCloser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MapCloser Suite")
}

var _ = BeforeSuite(func() {
	// Initialize global test context for the entire test suite
	globalTestCtx, globalTestCancel = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	// Clean up global test context
	if globalTestCancel != nil {
		globalTestCancel()
	}
})
