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

// Package aggregator_test provides comprehensive BDD-style tests for the aggregator package.
//
// Test Organization:
//   - aggregator_suite_test.go: Test suite setup and helper utilities
//   - new_test.go: Aggregator creation and configuration tests
//   - writer_test.go: Write operations and Close() tests
//   - runner_test.go: Lifecycle management (Start/Stop/Restart) tests
//   - concurrency_test.go: Thread-safety and race condition tests
//   - errors_test.go: Error handling and edge case tests
//   - metrics_test.go: Monitoring metrics (NbWaiting, NbProcessing, etc.) tests
//   - coverage_test.go: Code coverage and atomic testing
//   - benchmark_test.go: Performance benchmarks using gmeasure
//   - example_test.go: Executable examples for GoDoc
//
// The tests use Ginkgo/Gomega for BDD-style testing and achieve >80% code coverage.
package aggregator_test

import (
	"context"
	"testing"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// Global context for all tests
	testCtx    context.Context
	testCancel context.CancelFunc
	globalLog  liblog.Logger
)

// TestAggregator is the entry point for the Ginkgo test suite
func TestAggregator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IOUtils/Aggregator Package Suite")
}

var _ = BeforeSuite(func() {
	testCtx, testCancel = context.WithCancel(context.Background())
	globalLog = liblog.New(context.Background())
	Expect(globalLog.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	if testCancel != nil {
		testCancel()
	}
})
