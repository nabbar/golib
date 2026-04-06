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

package udp_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// # Ginkgo Test Suite Entry Point
//
// This file orchestrates the execution of all tests in the 'udp_test' package.
//
// # Responsibilities:
//  - Registration of the fail handler for Gomega.
//  - Initialization of the global test context used by all specs.
//  - Orchestration of BeforeSuite and AfterSuite logic for global teardown.

// testCtx is a global context used by all tests in the suite.
var testCtx context.Context

// testCancel is a global cancellation function to trigger cleanup after the entire suite.
var testCancel context.CancelFunc

// TestUdpServer is the main entry point for the "go test" command.
func TestUdpServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket/Server/UDP Package Suite")
}

// BeforeSuite runs once before any other test file.
var _ = BeforeSuite(func() {
	// Global context to prevent test hangs and ensure clean resource release.
	testCtx, testCancel = context.WithCancel(context.Background())
})

// AfterSuite runs once after all test files have completed.
var _ = AfterSuite(func() {
	if testCancel != nil {
		testCancel()
	}
})
