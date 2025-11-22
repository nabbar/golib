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
 */

package aggregator

import (
	"context"
	"time"

	"github.com/nabbar/golib/runner"
)

// Deadline returns the time when work done on behalf of this context should be cancelled.
//
// This implements context.Context interface and delegates to the aggregator's internal context.
// The deadline is inherited from the parent context provided to New() or Start().
//
// Returns:
//   - deadline: Time when the context will be cancelled
//   - ok: true if a deadline is set, false otherwise
func (o *agg) Deadline() (deadline time.Time, ok bool) {
	if x := o.x.Load(); x != nil {
		return x.Deadline()
	}
	return time.Time{}, false
}

// Done returns a channel that's closed when work done on behalf of this context should be cancelled.
//
// This implements context.Context interface. The channel is closed when:
//   - The parent context is cancelled
//   - Stop() is called
//   - Close() is called
//   - The context deadline is exceeded
//
// Returns:
//   - <-chan struct{}: A channel that's closed when the context is done
func (o *agg) Done() <-chan struct{} {
	if x := o.x.Load(); x != nil {
		return x.Done()
	}

	c := make(chan struct{})
	close(c)
	return c
}

// Err returns nil if Done is not yet closed.
// If Done is closed, Err returns a non-nil error explaining why:
//   - Canceled: if the context was cancelled
//   - DeadlineExceeded: if the context's deadline passed
//
// This implements context.Context interface.
//
// Returns:
//   - error: nil if context is active, otherwise the cancellation error
func (o *agg) Err() error {
	if x := o.x.Load(); x != nil {
		return x.Err()
	}
	return nil
}

// Value returns the value associated with this context for key.
//
// This implements context.Context interface and delegates to the parent context.
// Use context values only for request-scoped data that transits processes and APIs,
// not for passing optional parameters to functions.
//
// Parameters:
//   - key: The key to lookup
//
// Returns:
//   - any: The value associated with key, or nil if no value is associated
func (o *agg) Value(key any) any {
	if x := o.x.Load(); x != nil {
		return x.Value(key)
	}
	return nil
}

// ctxNew creates a new internal context derived from the provided parent context.
// This is called during initialization and when the aggregator starts.
//
// The method creates a cancellable context and stores both the context and its
// cancel function for later use. If there was a previous context, its cancel
// function is called to prevent resource leaks.
func (o *agg) ctxNew(ctx context.Context) {
	defer runner.RecoveryCaller("golib/ioutils/aggregator/ctxnew", recover())

	if ctx == nil || ctx.Err() != nil {
		ctx = context.Background()
	}

	x, n := context.WithCancel(ctx)
	o.x.Store(x)

	old := o.n.Swap(n)
	if old != nil {
		old()
	}
}

// ctxClose cancels the internal context and replaces it with a pre-cancelled context.
// This ensures that any ongoing operations respecting the context will terminate,
// and future Done() calls will receive a closed channel.
//
// The method is safe to call multiple times.
func (o *agg) ctxClose() {
	defer runner.RecoveryCaller("golib/ioutils/aggregator/ctxclose", recover())

	// Cancel old context first and clear it atomically
	old := o.n.Swap(func() {})
	if old != nil {
		old()
	}

	// Create a new cancelled context for future Done() calls
	x, n := context.WithCancel(context.Background())
	n() // Cancel immediately - don't store this cancel func

	o.x.Store(x)
}
