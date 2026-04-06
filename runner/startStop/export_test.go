/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package startStop

import (
	"context"
	"sync/atomic"
	"time"
)

// This file (export_test.go) is part of the startStop package but is only included
// during testing. It follows the Go idiom of exporting internal/unexported
// members and methods to the external test package (startStop_test).

// ExportIsRunning provides external test access to the internal IsRunning method.
func (o *run) ExportIsRunning() bool {
	return o.IsRunning()
}

// ExportCancel provides external test access to the internal cancel method,
// allowing tests to simulate manual cancellation.
func (o *run) ExportCancel() {
	o.cancel()
}

// ExportNewCancel provides external test access to the internal newCancel method.
func (o *run) ExportNewCancel() context.Context {
	return o.newCancel()
}

// ExportUptime provides external test access to the internal Uptime method.
func (o *run) ExportUptime() time.Duration {
	return o.Uptime()
}

// NewRunNil returns a nil pointer to a run struct, used for testing
// nil receiver behavior and error handling.
func NewRunNil() *run {
	return nil
}

// NewRunNilAtomic returns a run struct with a nil atomic boolean,
// used to test edge cases in initialization and robustness.
func NewRunNilAtomic() *run {
	return &run{
		r: nil,
	}
}

// ExportNewRunNoStartTime returns a run struct where the atomic boolean is
// initialized but no start time has been set, useful for testing initial state.
func ExportNewRunNoStartTime() *run {
	return &run{
		r: &atomic.Bool{},
	}
}

// SetRunning allows tests to directly manipulate the internal running state
// without going through the Start/Stop cycle.
func (o *run) SetRunning(val bool) {
	o.r.Store(val)
}
