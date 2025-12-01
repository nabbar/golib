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

package fields

import (
	"time"
)

// Deadline returns the time when work done on behalf of this context should be canceled.
//
// This method implements context.Context. It delegates to the underlying context provided
// to New(). If no deadline was set on the context, ok will be false.
//
// This is useful for implementing timeout-aware operations with Fields.
//
// Example:
//
//	if deadline, ok := flds.Deadline(); ok {
//	    timeLeft := time.Until(deadline)
//	    // Adjust behavior based on time remaining
//	}
func (o *fldModel) Deadline() (deadline time.Time, ok bool) {
	return o.c.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
//
// This method implements context.Context. It delegates to the underlying context provided
// to New(). The channel is closed when the context is canceled, either explicitly via a
// cancel function, by deadline expiration, or by parent context cancellation.
//
// Example:
//
//	select {
//	case <-flds.Done():
//	    // Context was canceled, cleanup and exit
//	case result := <-workChannel:
//	    // Normal completion
//	}
func (o *fldModel) Done() <-chan struct{} {
	return o.c.Done()
}

// Err returns nil if Done is not yet closed, or a non-nil error explaining why Done was closed.
//
// This method implements context.Context. Common errors are:
//   - context.Canceled: Context was explicitly canceled
//   - context.DeadlineExceeded: Context deadline was exceeded
//
// Example:
//
//	if err := flds.Err(); err != nil {
//	    if err == context.Canceled {
//	        // Operation was canceled
//	    } else if err == context.DeadlineExceeded {
//	        // Operation timed out
//	    }
//	}
func (o *fldModel) Err() error {
	return o.c.Err()
}

// Value returns the value associated with this context for key.
//
// This method implements context.Context. It delegates to the underlying context provided
// to New(). This is distinct from Get() which accesses the Fields key-value store.
//
// Use this for context values (request-scoped data), not for logging fields.
//
// Example:
//
//	if requestID := flds.Value("request_id"); requestID != nil {
//	    // Use context value
//	}
func (o *fldModel) Value(key any) any {
	return o.c.Value(key)
}
