/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package version_test

import (
	"io"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// originalStdout and originalStderr are saved to restore after tests
	originalStdout *os.File
	originalStderr *os.File
)

// TestVersion is the entry point for Ginkgo test suite.
// It integrates Ginkgo with Go's testing framework and runs all version package tests.
func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Suite")
}

// testStruct is a helper struct used for testing reflection-based package path extraction.
// It's used to verify that NewVersion correctly extracts package information from reflection.
type testStruct struct{}

var (
	// testTime is a fixed RFC3339 time string used across tests for consistency.
	testTime = "2024-01-15T10:30:00Z"
	// testTimeParsed is the parsed version of testTime.
	testTimeParsed time.Time
)

var _ = BeforeSuite(func() {
	var err error
	testTimeParsed, err = time.Parse(time.RFC3339, testTime)
	Expect(err).ToNot(HaveOccurred())

	// Save original stdout and stderr
	originalStdout = os.Stdout
	originalStderr = os.Stderr

	// Redirect stdout and stderr to discard output during tests
	// This prevents Print methods from polluting test output
	os.Stdout = nil
	os.Stderr = nil
	_, _ = io.Discard.Write([]byte{})
})

var _ = AfterSuite(func() {
	// Restore original stdout and stderr
	if originalStdout != nil {
		os.Stdout = originalStdout
	}
	if originalStderr != nil {
		os.Stderr = originalStderr
	}
})
