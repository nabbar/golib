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

var (
	// closeStruct is a pre-closed channel used as a sentinel value
	// to indicate that the hook has been closed.
	closeStruct = make(chan struct{})

	// closeByte is a pre-closed channel used as a sentinel value
	// to indicate that the data channel has been closed.
	closeByte = make(chan []data, 250)
)

// init closes the sentinel channels at package initialization.
func init() {
	close(closeStruct)
	close(closeByte)
}

// prepareChan initializes the internal channels for the hook.
// Called by Run() before starting the background writer goroutine.
//
// Creates:
//   - Data channel: Buffered channel for log entries (capacity: 250)
//   - Done channel: Unbuffered signal channel for shutdown
func (o *hks) prepareChan() {
	o.d.Store(make(chan []data, 250))
	o.s.Store(make(chan struct{}))
}

func (o *hks) Done() <-chan struct{} {
	c := o.s.Load()

	if c != nil {
		return c.(chan struct{})
	}

	return closeStruct
}

// Data returns the receive-only data channel for log entries.
// This is an internal method used by Run() to receive buffered entries.
//
// Returns:
//   - Active channel: If the hook is running
//   - Closed channel: If Close() was called (sentinel value)
func (o *hks) Data() <-chan []data {
	c := o.d.Load()

	if c != nil {
		return c.(chan []data)
	}

	return closeByte
}
