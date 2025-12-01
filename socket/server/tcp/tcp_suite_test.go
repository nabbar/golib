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

// tcp_suite_test.go initializes the Ginkgo test suite for the TCP server package.
// Sets up the BDD test framework and integrates with Go's testing infrastructure.
package tcp_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// ctx is the global test context shared across all test specs
	globalCtx context.Context
	// globalCnl is the cancel function for the global context
	globalCnl context.CancelFunc
)

func TestServerTCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server TCP Suite")
}

var _ = BeforeSuite(func() {
	globalCtx, globalCnl = context.WithCancel(context.Background())

	// Initialize TLS configurations in the test suite
	initTLSConfigs()
})

var _ = AfterSuite(func() {
	if globalCnl != nil {
		globalCnl()
	}
})
