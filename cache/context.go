/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package cache

import "time"

// Deadline returns the time when work done on behalf of this context should be canceled.
// It delegates to the underlying context.
func (o *cc[K, V]) Deadline() (deadline time.Time, ok bool) {
	return o.Context.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
// It delegates to the underlying context.
func (o *cc[K, V]) Done() <-chan struct{} {
	return o.Context.Done()
}

// Err returns a non-nil error value after Done is closed.
// It delegates to the underlying context.
func (o *cc[K, V]) Err() error {
	return o.Context.Err()
}

// Value returns the value associated with this context for key.
// It first checks if the key matches a cache key type, and if so, returns the cached value.
// Otherwise, it delegates to the underlying context.
func (o *cc[K, V]) Value(key any) any {
	if sKey, ok := key.(K); ok {
		if v, _, k := o.Load(sKey); k {
			return v
		}
	}

	return o.Context.Value(key)
}
