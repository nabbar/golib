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

package ticker_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestTicker runs the ginkgo test suite for the ticker package.
// This test suite validates the ticker-based runner functionality,
// including lifecycle management, error handling, concurrency, and edge cases.
//
// Test Structure:
//   - lifecycle_test.go: Basic lifecycle operations (Start, Stop, Restart, Uptime, IsRunning)
//   - concurrency_test.go: Concurrent operations and race condition detection
//   - errors_test.go: Error collection and handling
//   - edge_cases_test.go: Edge cases, boundary conditions, and unusual scenarios
//
// Running Tests:
// To run with race detector (recommended):
//
//	CGO_ENABLED=1 go test -race ./...
//
// To run with ginkgo and repeat for stability testing:
//
//	CGO_ENABLED=1 ginkgo -v --race --repeat=10 .
//
// For more information on the ticker package, see github.com/nabbar/golib/runner/ticker.
func TestTicker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ticker Suite")
}
