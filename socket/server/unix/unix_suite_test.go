//go:build linux || darwin

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

// Package unix_test provides a Behavior-Driven Development (BDD) test suite for the Unix domain socket server.
//
// # unix_suite_test.go: Test Suite Bootstrapping and Global Lifecycle
//
// This file initializes the Ginkgo testing framework and defines the global setup and teardown
// logic shared by all test files within the `unix_test` package. It ensures that every test run
// starts with a clean environment and that global resources are correctly released.
//
// # Responsibilities:
//
// ## 1. Framework Integration
//   - Ginkgo/Gomega Setup: Hooks into Go's `testing` package via the `TestServerUnix` function.
//   - Fail Handler: Registers the Gomega fail handler to provide descriptive error messages when
//     expectations are not met.
//
// ## 2. Global Context Management
//   - `globalCtx`: A top-level context shared by all test specs. This allows for a unified way
//     to cancel all background processes (like servers) at the end of the suite.
//   - `globalCnl`: The cancel function for the global context, called in the `AfterSuite` block.
//
// ## 3. Lifecycle Hooks
//   - `BeforeSuite`: Executed once before any tests run. Initializes the global context.
//   - `AfterSuite`: Executed once after all tests have finished. Cancels the global context to
//     ensure no goroutines or resources are leaked.
//
// # Testing Philosophy:
// The suite is designed to be fully automated and self-contained. Each test file focuses on
// a specific aspect (Creation, Basic Ops, Concurrency, Robustness, Benchmarks), but they
// all share the same bootstrapping logic defined here.
//
// # How to Run the Suite:
//
//   go test -v .                   # Run all tests
//   ginkgo -v -r                   # Run with Ginkgo's enhanced CLI
//   go test -v -race .             # Run with race detector enabled (highly recommended)
package unix_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// globalCtx is the global test context shared across all test specs
	globalCtx context.Context
	// globalCnl is the cancel function for the global context
	globalCnl context.CancelFunc
)

// TestServerUnix is the entry point for the Go test runner.
// It delegates the execution to the Ginkgo BDD framework.
func TestServerUnix(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server Unix Suite")
}

var _ = BeforeSuite(func() {
	// Initialize the shared context for all tests.
	globalCtx, globalCnl = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	// Ensure all background tasks are stopped after the suite completes.
	if globalCnl != nil {
		globalCnl()
	}
})
