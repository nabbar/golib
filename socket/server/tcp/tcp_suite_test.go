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

// Package tcp_test contains the BDD (Behavior-Driven Development) test suite for the TCP server.
// It uses the Ginkgo and Gomega frameworks to validate server behavior under various conditions.
//
// # Test Strategy Overview
//
// The test suite is organized into several specialized files to ensure comprehensive coverage:
//   - creation_test.go: Validates server initialization and configuration (address, TLS).
//   - basic_test.go: Standard operational tests (Echo, Read/Write, Shutdown).
//   - tls_test.go: Verifies SSL/TLS handshake, certificate management, and mTLS.
//   - context_test.go: Tests connection-specific context cancellation and deadlines.
//   - concurrency_test.go: Stresses the server with high numbers of simultaneous clients.
//   - robustness_test.go: Tests edge cases (slow clients, large payloads, abrupt closures).
//   - benchmark_test.go: Measures throughput and memory allocation (sync.Pool efficiency).
//
// # Testing Dataflow: Ginkgo Lifecycle
//
//	[BeforeSuite] ───────────┐
//	     │                   │
//	     v                   v
//	[Describe/Context] ───> [BeforeEach (per test)]
//	     │                   │
//	     v                   v
//	[It (Actual Test)] <─── [AfterEach (cleanup)]
//	     │                   │
//	     v                   │
//	[AfterSuite] <───────────┘
//
// # Performance & Race Detection
//
// Tests are designed to be run with the race detector enabled (-race) to ensure 
// that the lock-free state management (atomic.Bool, etc.) is implemented correctly.
package tcp_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// globalCtx is the root context for all tests. It can be used to signal 
	// a full stop of all background test components.
	globalCtx context.Context
	// globalCnl triggers the cancellation of the globalCtx.
	globalCnl context.CancelFunc
)

// TestServerTCP is the entry point for the standard 'go test' command.
// It registers the Ginkgo fail handler and triggers the suite execution.
func TestServerTCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server TCP Suite")
}

// BeforeSuite is executed once before any test file is processed.
// It handles global resource allocation (Context, TLS certs).
var _ = BeforeSuite(func() {
	globalCtx, globalCnl = context.WithCancel(context.Background())

	// Initialize mock TLS configurations (certs, CA) used across multiple tests.
	initTLSConfigs()
})

// AfterSuite is executed once after all tests have finished.
// It ensures that no background goroutines or leaky listeners remain.
var _ = AfterSuite(func() {
	if globalCnl != nil {
		globalCnl()
	}
})
