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
 */

package console_test

import (
	"io"
	"os"
	"testing"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConsole(t *testing.T) {
	// Disable color output to prevent stdout pollution in tests
	color.NoColor = true

	RegisterFailHandler(Fail)
	RunSpecs(t, "console Suite")
}

// captureStdout captures stdout during the execution of f and discards the output
func captureStdout(f func()) {
	// Save original values
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldColorOutput := color.Output

	// Create pipe to discard output
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	color.Output = w

	// Discard output in background
	done := make(chan struct{})
	go func() {
		io.Copy(io.Discard, r)
		close(done)
	}()

	// Execute function
	f()

	// Restore original values
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	color.Output = oldColorOutput
	<-done
}
