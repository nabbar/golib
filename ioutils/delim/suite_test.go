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

package delim_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestDelim is the entry point for the Ginkgo BDD test suite.
// This suite provides comprehensive testing for the delim package including:
//   - Basic constructor and interface tests
//   - Read/Write operation tests
//   - Concurrency and race condition tests
//   - Edge cases and error handling tests
//   - Performance benchmarks
//   - Robustness and boundary tests
//
// The test suite uses Ginkgo v2 for BDD-style testing and Gomega for assertions.
// Tests are organized into logical groups for maintainability and clarity.
//
// A global test context is initialized at the beginning of the suite and
// cleaned up at the end to ensure proper resource management.
//
// Run with: go test -v
// Run with race detection: CGO_ENABLED=1 go test -race
// Run with coverage: go test -coverprofile=coverage.out
func TestDelim(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOUtils/Delim Package Suite")
}

var _ = BeforeSuite(func() {
	// Initialize global test context for all tests in the suite
	initTestContext()
})

var _ = AfterSuite(func() {
	// Cleanup global test context and release resources
	cleanupTestContext()
})
