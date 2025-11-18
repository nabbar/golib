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
 */

package command_test

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestCommand is the entry point for the Ginkgo test suite
func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shell/Command Package Suite")
}

// Helper types and functions

// safeBuffer is a thread-safe buffer for concurrent writes
type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (sb *safeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func (sb *safeBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.buf.Reset()
}

func (sb *safeBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Len()
}

// newSafeBuffer creates a new thread-safe buffer
func newSafeBuffer() *safeBuffer {
	return &safeBuffer{}
}

// captureOutput captures output from a function that writes to io.Writer
func captureOutput(fn func(out, err io.Writer)) (stdout, stderr string) {
	outBuf := newSafeBuffer()
	errBuf := newSafeBuffer()
	fn(outBuf, errBuf)
	return outBuf.String(), errBuf.String()
}

// testWriter is a writer that can be configured to fail
type testWriter struct {
	mu         sync.Mutex
	shouldFail bool
	written    []byte
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.shouldFail {
		return 0, fmt.Errorf("intentional write error")
	}

	tw.written = append(tw.written, p...)
	return len(p), nil
}

func (tw *testWriter) SetShouldFail(fail bool) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.shouldFail = fail
}

func (tw *testWriter) GetWritten() []byte {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	result := make([]byte, len(tw.written))
	copy(result, tw.written)
	return result
}

func (tw *testWriter) Reset() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.written = nil
}

// newTestWriter creates a new test writer
func newTestWriter() *testWriter {
	return &testWriter{
		written: make([]byte, 0),
	}
}
