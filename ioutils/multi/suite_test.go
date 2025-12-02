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
 *
 */

// Package multi_test provides comprehensive test coverage for the multi package.
//
// The test suite is organized into the following files:
//   - suite_test.go: Test suite entry point and configuration
//   - constructor_test.go: Tests for constructor and interface compliance
//   - reader_test.go: Tests for read operations and input management
//   - writer_test.go: Tests for write operations and output management
//   - copy_test.go: Tests for copy operations and integration scenarios
//   - concurrent_test.go: Tests for concurrent safety and thread-safe operations
//   - edge_cases_test.go: Tests for error handling and boundary conditions
//   - benchmark_test.go: Performance benchmarks and allocation measurements
//   - helper_test.go: Shared test helpers and utilities
//   - example_test.go: Runnable examples demonstrating package usage
//
// The test suite uses Ginkgo/Gomega for BDD-style testing and gmeasure
// for performance benchmarking. Tests are designed to validate:
//   - Thread-safe concurrent operations
//   - Error propagation and handling
//   - Interface compliance (io.ReadWriteCloser, io.StringWriter)
//   - Memory allocation efficiency
//   - Edge cases and boundary conditions
package multi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestMulti is the entry point for the Ginkgo test suite.
// It registers the Gomega fail handler and runs all specs in the multi package.
func TestMulti(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOUtils/Multi Package Suite")
}
