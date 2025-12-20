/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

// Package ioprogress_test provides comprehensive BDD-style tests for the ioprogress package.
//
// Test Framework: Uses Ginkgo v2 for BDD-style test organization and Gomega for expressive
// assertions. This combination provides clear, readable test specifications that document
// package behavior.
//
// Test Coverage:
//   - Reader operations: 22 specs covering Read(), callbacks, EOF, Reset(), Close()
//   - Writer operations: 20 specs covering Write(), callbacks, EOF, Reset(), Close()
//   - Edge cases: nil callbacks, empty data, large data, zero-byte operations
//   - Concurrency: thread-safe callback registration during I/O operations
//   - Total: 42 specs with 84.7% code coverage
//
// Test Organization: Tests follow BDD hierarchical structure (Describe → Context → It)
// for clear documentation of expected behavior. Each test is independent and follows the
// AAA pattern (Arrange, Act, Assert).
//
// Running Tests:
//
//	go test ./...                    # Basic test run
//	go test -cover ./...             # With coverage report
//	CGO_ENABLED=1 go test -race ./... # With race detector
//	ginkgo -v                        # Using Ginkgo CLI (verbose)
//
// Race Detection: All tests pass with -race flag, validating lock-free thread safety
// using atomic operations throughout the implementation.
package ioprogress_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestIoProgress is the entry point for the Ginkgo test suite.
//
// This function registers the Ginkgo fail handler and runs all specs defined in
// reader_test.go and writer_test.go. The suite name "IOProgress Suite" appears
// in test output for identification.
func TestIoProgress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOProgress Suite")
}
