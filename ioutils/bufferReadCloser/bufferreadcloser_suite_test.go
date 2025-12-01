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

// Package bufferReadCloser_test provides comprehensive BDD-style tests for the
// bufferReadCloser package using Ginkgo v2 and Gomega.
//
// Test Coverage:
//   - Buffer wrapper: Creation, read/write operations, close behavior, nil handling
//   - Reader wrapper: Creation, read operations, close behavior, nil handling
//   - Writer wrapper: Creation, write operations, flush/close behavior, nil handling
//   - ReadWriter wrapper: Creation, bidirectional I/O, close behavior, nil handling
//   - Custom close functions: Execution, error propagation
//   - Edge cases: Empty buffers, large data, multiple close calls
//
// The test suite achieves 100% code coverage with 62 specs covering all
// functionality, error paths, and boundary conditions.
package bufferReadCloser_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestBufferReadCloser is the entry point for the Ginkgo test suite.
// It registers the Gomega fail handler and runs all specs defined in the package.
func TestBufferReadCloser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BufferReadCloser Suite")
}
