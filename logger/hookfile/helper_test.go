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

// Package hookfile_test provides test utilities and helpers for the hookfile package.
// This file contains shared test setup, teardown, and helper functions used across
// the test suite.
package hookfile_test

import (
	"os"
	"path/filepath"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Global test variables shared across test suites
var (
	// tempDir is the temporary directory created for test files
	tempDir string
	// testLogFile is the path to the main test log file
	testLogFile string
)

var _ = BeforeSuite(func() {
	// Create a temporary directory for test files
	var err error
	tempDir, err = os.MkdirTemp("", "hookfile-test-*")
	Expect(err).NotTo(HaveOccurred(), "Failed to create temp directory")

	// Set up test log file path
	testLogFile = filepath.Join(tempDir, "test.log")
})

var _ = AfterSuite(func() {
	// Clean up test files
	// Note: ResetOpenFiles is called in each test's AfterEach
	if tempDir != "" {
		// Delay to ensure all goroutines have stopped (longer with race detector)
		time.Sleep(500 * time.Millisecond)
		_ = os.RemoveAll(tempDir)
	}
})

// createTestHook creates a new HookFile instance for testing with default options.
func createTestHook() (logfil.HookFile, error) {
	// Use default options with test log file path
	opts := logcfg.OptionsFile{
		Filepath:   testLogFile,
		FileMode:   0600,
		PathMode:   0700,
		CreatePath: true,
		LogLevel:   []string{"debug", "info", "warn", "error"},
	}

	// Create a simple text formatter for testing
	formatter := &logrus.TextFormatter{
		DisableTimestamp: true,
	}

	return logfil.New(opts, formatter)
}
