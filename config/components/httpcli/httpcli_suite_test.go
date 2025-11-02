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

package httpcli_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestHTTPCli runs the Ginkgo test suite for the HTTPCli component package.
// This suite tests HTTP client configuration management, DNS mapping,
// and integration with the certificate system.
//
// Test coverage includes:
//   - Component lifecycle (Init, Start, Reload, Stop)
//   - Configuration management and validation
//   - DNS mapper creation and manipulation
//   - Root CA certificate handling
//   - Default HTTP client behavior
//   - Message function callbacks
//   - Error conditions and edge cases
//   - Concurrent access scenarios
//   - Integration with DNS and TLS systems
//   - Flag registration for CLI
//
// The tests use standalone implementations without external dependencies
// to avoid billing or security issues. All tests are designed to be
// human-readable and maintainable with separate files per scope.
//
// Run tests with:
//
//	go test -v
//	go test -v -cover
//	CGO_ENABLED=1 go test -v -race
//
// For detailed coverage:
//
//	go test -v -coverprofile=coverage.out
//	go tool cover -html=coverage.out
func TestHTTPCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTPCli Component Suite")
}
