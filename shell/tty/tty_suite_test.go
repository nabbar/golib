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

package tty_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestTTY is the entry point for the Ginkgo test suite
func TestTTY(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TTY Package Suite")
}

// Helper types and functions

// mockTTYSaver is a mock implementation of TTYSaver interface for testing
type mockTTYSaver struct {
	mu            sync.Mutex
	restoreCalled bool
	signalCalled  bool
	restoreError  error
	signalError   error
	isTerminal    bool
}

func (m *mockTTYSaver) IsTerminal() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isTerminal
}

func (m *mockTTYSaver) Restore() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restoreCalled = true
	return m.restoreError
}

func (m *mockTTYSaver) Signal() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signalCalled = true
	return m.signalError
}

func (m *mockTTYSaver) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.restoreCalled
}

func (m *mockTTYSaver) SignalWasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signalCalled
}

func (m *mockTTYSaver) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restoreCalled = false
	m.signalCalled = false
}

func newMockTTYSaver(shouldFail bool) *mockTTYSaver {
	mock := &mockTTYSaver{
		isTerminal: true, // Default to true for testing
	}
	if shouldFail {
		mock.restoreError = ErrorMockRestore
	}
	return mock
}

func newMockTTYSaverWithTerminal(shouldFail, isTerminal bool) *mockTTYSaver {
	mock := &mockTTYSaver{
		isTerminal: isTerminal,
	}
	if shouldFail {
		mock.restoreError = ErrorMockRestore
	}
	return mock
}

var ErrorMockRestore = &mockError{"mock restore failed"}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
