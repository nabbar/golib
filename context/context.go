/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"context"
	"time"
)

// GetContext returns the underlying context.Context instance.
// If the context function is nil or returns nil, it returns context.Background().
// Thread-safe: Uses read lock for concurrent access.
func (c *ccx[T]) GetContext() context.Context {
	if c.x != nil {
		return c.x
	} else {
		return context.Background()
	}
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. Delegates to the underlying context.Context.
func (c *ccx[T]) Deadline() (deadline time.Time, ok bool) {
	return c.x.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Delegates to the underlying context.Context.
func (c *ccx[T]) Done() <-chan struct{} {
	return c.x.Done()
}

// Err returns a non-nil error value after Done is closed.
// Delegates to the underlying context.Context.
func (c *ccx[T]) Err() error {
	return c.x.Err()
}

// Value returns the value associated with this context for key.
// First attempts to load from the Config storage using type assertion to T.
// Falls back to the underlying context.Context.Value if not found.
func (c *ccx[T]) Value(key any) any {
	if i, k := key.(T); !k {
		return c.x.Value(key)
	} else if v, ok := c.Load(i); ok {
		return v
	} else {
		return c.x.Value(key)
	}
}
