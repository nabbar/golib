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

package hooksyslog

import (
	"fmt"
)

// Write implements io.Writer by writing data to syslog with INFO severity.
// This is a convenience method that delegates to WriteSev with severity 0
// (which defaults to INFO level in the wrapper).
//
// Parameters:
//   - p: Data to write to syslog
//
// Returns:
//   - n: Number of bytes queued (len(p)) if successful
//   - err: Error if channel is closed
//
// Note: This method is non-blocking as long as the channel buffer has space.
func (o *hks) Write(p []byte) (n int, err error) {
	return o.WriteSev(0, p)
}

func (o *hks) WriteSev(s SyslogSeverity, p []byte) (n int, err error) {
	c := o.d.Load()

	if c != nil {
		if c.(chan []data) != closeByte {
			// Send a slice containing a single data element
			c.(chan []data) <- []data{newData(s, p)}
			return len(p), nil
		}
	}

	return 0, fmt.Errorf("%v, path: %s", errStreamClosed, o.getSyslogInfo())

}

// Close terminates the hook by closing the internal channels.
// After calling Close, no new log entries can be written.
//
// This method should be called during application shutdown, after
// cancelling the context passed to Run().
//
// Typical shutdown sequence:
//
//	cancel() // Stop the Run goroutine
//	hook.Close() // Close the channels
//	<-hook.Done() // Wait for Run to complete
//
// Returns:
//   - Always returns nil (implements io.Closer)
func (o *hks) Close() error {
	o.d.Store(closeByte)
	o.s.Store(closeStruct)
	return nil
}
