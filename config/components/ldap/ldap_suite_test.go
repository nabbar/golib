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

package ldap_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestLDAP runs the Ginkgo test suite for the LDAP component package.
// This suite tests LDAP client configuration management and lifecycle operations.
//
// Test coverage includes:
//   - Component lifecycle (Init, Start, Reload, Stop)
//   - Configuration management and validation
//   - LDAP helper integration
//   - Attributes management
//   - Registration and loading mechanisms
//   - Error conditions and edge cases
//   - Dependencies management
//   - Callback execution
//
// The tests use standalone implementations without external LDAP servers
// to avoid dependencies on external services.
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
func TestLDAP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LDAP Component Suite")
}
