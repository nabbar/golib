/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package iowrapper_test contains comprehensive BDD tests for the iowrapper package.
//
// Test Organization:
//   - basic_test.go: Basic I/O operations and wrapper creation (20 specs)
//   - custom_test.go: Custom function registration and behavior (24 specs)
//   - edge_cases_test.go: Edge cases and boundary conditions (18 specs)
//   - errors_test.go: Error handling and propagation (19 specs)
//   - concurrency_test.go: Thread safety and concurrent operations (17 specs)
//   - integration_test.go: Real-world integration scenarios (8 specs)
//   - benchmark_test.go: Performance and memory benchmarks (8 specs)
//   - example_test.go: Executable examples demonstrating usage patterns
//
// Test Coverage: 100.0% (114 passing specs)
// Framework: Ginkgo v2 (BDD) + Gomega (matchers)
//
// Run tests with:
//
//	go test -v -cover .
//	go test -race .  # With race detector
//	ginkgo -v -cover # With Ginkgo CLI
package iowrapper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestIOWrapper is the entry point for the Ginkgo test suite.
// It registers the Gomega failure handler and runs all specs in this package.
func TestIOWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOWrapper Suite")
}
