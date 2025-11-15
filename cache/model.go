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

import (
	"context"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	cchitm "github.com/nabbar/golib/cache/item"
)

// cc is the internal implementation of the Cache interface.
// It uses a thread-safe map to store cache items with automatic expiration.
type cc[K comparable, V any] struct {
	context.Context

	n context.CancelFunc                      // cancel function for the cache context
	v libatm.MapTyped[K, cchitm.CacheItem[V]] // thread-safe map storing cache items
	e time.Duration                           // expiration duration for cache items
}

// Clone creates a new independent copy of the cache with the same items.
// The new cache uses the provided context and is not affected by changes to the original cache.
func (o *cc[K, V]) Clone(ctx context.Context) (Cache[K, V], error) {
	if e := o.Err(); e != nil {
		return nil, e
	}

	var cnl context.CancelFunc
	if ctx == nil {
		ctx = context.WithoutCancel(o.Context)
	}
	ctx, cnl = context.WithCancel(ctx)

	n := &cc[K, V]{
		Context: ctx,

		n: cnl,
		v: libatm.NewMapTyped[K, cchitm.CacheItem[V]](),
		e: o.e,
	}

	o.Walk(func(k K, v V, _ time.Duration) bool {
		n.Store(k, v)
		return true
	})

	return n, nil
}

// Merge adds all items from the provided cache into the current cache.
// Items are copied with their current expiration times reset.
func (o *cc[K, V]) Merge(c Cache[K, V]) {
	c.Walk(func(key K, val V, exp time.Duration) bool {
		if o.Err() != nil {
			return false
		} else if i, l := o.v.Load(key); l {
			i.Clean()
		}

		o.v.Store(key, cchitm.New[V](o.e, val))
		return true
	})

	if o.Err() != nil {
		o.Clean()
	}
}

// Walk iterates over all valid (non-expired) items in the cache.
// The provided function is called for each item with its key, value, and remaining time.
// If the function returns false, iteration stops early.
func (o *cc[K, V]) Walk(fct func(K, V, time.Duration) bool) {
	o.v.Range(func(key K, val cchitm.CacheItem[V]) bool {
		if o.Err() != nil {
			return false
		} else if v, r, k := val.LoadRemain(); !k {
			o.v.Delete(key)
			return true
		} else {
			return fct(key, v, r)
		}
	})

	if o.Err() != nil {
		o.Clean()
	}
}

// Load retrieves the value for the given key from the cache.
// Returns the value, remaining expiration time, and a boolean indicating if the value was found and valid.
// Expired items are automatically removed during this operation.
func (o *cc[K, V]) Load(key K) (V, time.Duration, bool) {
	var zero V

	if o.Err() != nil {
		o.Clean()
	} else if i, l := o.v.Load(key); !l {
		o.v.Delete(key)
	} else if v, r, k := i.LoadRemain(); !k {
		o.v.Delete(key)
	} else {
		return v, r, true
	}

	return zero, 0, false
}

// Store saves the given value in the cache with the specified key.
// The item will expire after the duration set when the cache was created.
func (o *cc[K, V]) Store(key K, val V) {
	o.v.Store(key, cchitm.New[V](o.e, val))
}

// Delete removes the item with the given key from the cache.
func (o *cc[K, V]) Delete(key K) {
	o.v.Delete(key)
}

// LoadOrStore loads the value for the given key if it exists and is valid.
// If the key doesn't exist or has expired, it stores the provided value.
// Returns the actual value (either existing or newly stored), remaining time, and whether the value was loaded (true) or stored (false).
func (o *cc[K, V]) LoadOrStore(key K, val V) (V, time.Duration, bool) {
	var zero V

	if o.Err() != nil {
		o.Clean()
	} else if i, l := o.v.LoadOrStore(key, cchitm.New[V](o.e, val)); !l {
	} else if v, r, k := i.LoadRemain(); !k {
		o.v.Store(key, cchitm.New[V](o.e, val))
	} else {
		return v, r, true
	}

	return zero, 0, false
}

// LoadAndDelete retrieves the value for the given key and removes it from the cache atomically.
// Returns the value and a boolean indicating if the value was found and valid.
func (o *cc[K, V]) LoadAndDelete(key K) (V, bool) {
	var zero V

	if o.Err() != nil {
		o.Clean()
	} else if i, l := o.v.LoadAndDelete(key); !l {
	} else if v, k := i.Load(); !k {
		o.v.Delete(key)
	} else {
		return v, true
	}

	return zero, false
}

// Swap replaces the value for the given key with the new value and returns the old value.
// Returns the old value, its remaining expiration time, and whether the old value was found and valid.
func (o *cc[K, V]) Swap(key K, val V) (V, time.Duration, bool) {
	var zero V

	if o.Err() != nil {
		o.Clean()
	}

	i, l := o.v.Load(key)
	o.v.Store(key, cchitm.New[V](o.e, val))

	if l {
		if v, r, k := i.LoadRemain(); k {
			o.v.Store(key, cchitm.New[V](o.e, val))
			return v, r, true
		}
	}

	return zero, 0, false
}
