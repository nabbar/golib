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

package aws_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestAws runs the Ginkgo test suite for the AWS component package.
// This suite tests AWS S3 client configuration, lifecycle management,
// monitoring integration, and error handling.
//
// Test coverage includes:
//   - Component lifecycle (Init, Start, Reload, Stop)
//   - Configuration drivers (Standard, StandardStatus, Custom, CustomStatus)
//   - AWS client initialization with mock endpoints
//   - Monitor registration and health checks
//   - Error conditions and edge cases
//
// Run tests with:
//
//	go test -v
//	go test -v -cover
//	go test -v -race (requires CGO_ENABLED=1)
func TestAws(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Component Suite")
}
