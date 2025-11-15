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

package item

import (
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
)

// itm is the internal implementation of the CacheItem interface.
// It uses atomic operations to ensure thread-safe access to the cached value and expiration time.
type itm[T any] struct {
	e time.Duration           // expiration duration
	k *atomic.Bool            // flag indicating if the item is valid
	t libatm.Value[time.Time] // timestamp when the item was last stored
	v libatm.Value[T]         // the actual cached value
}

// Check verifies if the item is still valid (not expired).
func (o *itm[T]) Check() bool {
	_, _, k := o.LoadRemain()
	return k
}

// Clean resets the item to its zero state, marking it as invalid.
func (o *itm[T]) Clean() {
	o.clean(true)
}

// Duration returns the configured expiration duration for this item.
func (o *itm[T]) Duration() time.Duration {
	return o.e
}

// Remain returns the remaining time until expiration and whether the item is still valid.
func (o *itm[T]) Remain() (time.Duration, bool) {
	_, r, k := o.LoadRemain()
	return r, k
}

// Load retrieves the cached value if it's still valid.
func (o *itm[T]) Load() (T, bool) {
	v, _, k := o.LoadRemain()
	return v, k
}

// LoadRemain retrieves the cached value along with the remaining expiration time.
// It returns the value, remaining duration, and whether the item is still valid.
func (o *itm[T]) LoadRemain() (T, time.Duration, bool) {
	var zero T
	if !o.k.Load() {
		return zero, 0, false
	} else if o.e == 0 {
		return o.v.Load(), 0, true
	} else if t := o.t.Load(); t.IsZero() {
		return zero, 0, o.clean(false)
	} else if r := time.Since(t); r >= o.e {
		return zero, 0, o.clean(false)
	} else {
		return o.v.Load(), r, true
	}
}

// Store saves the given value and resets the expiration timer.
func (o *itm[T]) Store(val T) {
	o.k.Store(true)
	o.t.Store(time.Now())
	o.v.Store(val)
}

// clean is an internal method that resets the item to its zero state.
// The res parameter is returned as-is to allow for convenient chaining.
func (o *itm[T]) clean(res bool) bool {
	var zero T
	o.k.Store(false)
	o.t.Store(time.Time{})
	o.v.Store(zero)
	return res
}
