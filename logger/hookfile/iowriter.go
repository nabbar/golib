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

// Package hookfile provides file-based logging hooks for logrus.
// This file implements the io.Writer interface and related methods for the file hook.
package hookfile

import (
	"context"
	"errors"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"
)

// Write writes the given byte slice to the underlying file writer.
// Implements the io.Writer interface.
//
// Parameters:
//   - p: The byte slice to write
//
// Returns:
//   - int: The number of bytes written
//   - error: Any error that occurred during writing
func (o *hkf) Write(p []byte) (n int, err error) {
	n, err = o.w.Write(p)

	if err == nil {
		return n, err
	} else if !errors.Is(err, iotagg.ErrClosedResources) {
		return n, err
	}

	// prevent only one run to update writer instance
	o.m.Lock()
	defer o.m.Unlock()

	// ok so having lock, but maybe the writer has just been update during waiting lock
	// so need to check if the error still here
	n, err = o.w.Write(p)
	if err == nil {
		return n, err
	} else if !errors.Is(err, iotagg.ErrClosedResources) {
		return n, err
	}

	a, e := setAgg(o.o.filepath, o.o.filemode, o.o.filecreate)
	if e != nil {
		return n, e
	}

	o.w = a
	if n, err = o.w.Write(p); err != nil {
		return n, err
	}

	// adding message on output file to inform the recovering process
	_, _ = o.w.Write([]byte(time.Now().Format(time.RFC3339) + " recovered closed resources, maybe some implementation error\n"))

	return n, err
}

// Close stops the hook and releases associated resources.
// It marks the hook as not running and removes it from the aggregator.
//
// Returns:
//   - error: Always returns nil, included for interface compatibility
func (o *hkf) Close() error {
	o.r.Store(false)
	delAgg(o.o.filepath)
	return nil
}

// IsRunning checks if the hook is currently active and accepting log entries.
//
// Returns:
//   - bool: True if the hook is running, false otherwise
func (o *hkf) IsRunning() bool {
	return o.r.Load()
}

// Run starts the hook's main processing loop.
// This method blocks until the provided context is canceled or the hook is closed.
// It's typically run in a separate goroutine.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// The method will automatically clean up resources when the context is done.
func (o *hkf) Run(ctx context.Context) {
	<-ctx.Done()
	o.r.Store(false)
}
