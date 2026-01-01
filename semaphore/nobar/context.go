/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package nobar

import "time"

// Deadline implements context.Context interface.
// Returns the deadline from the underlying semaphore context.
func (o *bar) Deadline() (deadline time.Time, ok bool) {
	return o.s.Deadline()
}

// Done implements context.Context interface.
// Returns a channel that's closed when the underlying semaphore context is cancelled.
func (o *bar) Done() <-chan struct{} {
	return o.s.Done()
}

// Err implements context.Context interface.
// Returns the error from the underlying semaphore context.
func (o *bar) Err() error {
	return o.s.Err()
}

// Value implements context.Context interface.
// Returns the value associated with key from the underlying semaphore context.
func (o *bar) Value(key any) any {
	return o.s.Value(key)
}
