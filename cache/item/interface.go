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

// Package item provides the internal cache item implementation for storing values with expiration.
// This package is primarily used internally by the cache package but can be used independently
// for single-item caching scenarios.
package item

import (
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
)

// CacheItem represents a single cached item with automatic expiration support.
// All operations are thread-safe and can be called concurrently from multiple goroutines.
type CacheItem[T any] interface {

	// Check will check if the item has expired or not.
	//
	// Returns:
	//   - bool: true if the item has expired, false otherwise.
	//
	// If the item has expired, it will be removed from the cache automatically.
	Check() bool

	// Clean will remove the item from the cache.
	//
	// It is generally not necessary to call Clean as items are automatically
	// removed from the cache when they expire. However, calling Clean can be
	// useful in certain cases, such as when the item is no longer needed and
	// you want to free up memory.
	//
	Clean()

	// Duration returns the expiration time of the item.
	//
	// Returns:
	//   - time.Duration: the expiration time of the item.
	//
	// The expiration time is the amount of time after which the item will be
	// removed from the cache if not accessed again.
	//
	// The expiration time is set by the `New` function when creating a new
	// item. It can be set to any positive value of time.Duration.
	//
	// If the expiration time is set to 0, the item will never expire.
	Duration() time.Duration

	// Remain returns the remaining time until the item expires and a boolean
	// indicating whether the item has expired or not.
	//
	// Returns:
	//   - time.Duration: the remaining time until the item expires.
	//   - bool: true if the item has expired, false otherwise.
	//
	// If the item has expired, it will be removed from the cache automatically.
	// If the item has not expired, the function returns the remaining time until
	// the item expires.
	Remain() (time.Duration, bool)

	// Store will store the given value into the cache.
	//
	// Parameters:
	// - T: the value to store into the cache.
	Store(val T)

	// Load will return the value stored into the cache for the given key.
	//
	// Returns:
	//   - T: the value stored into the cache.
	//   - bool: true if the value was found and valid, false otherwise.
	Load() (T, bool)

	// LoadRemain loads the value stored into the cache for the given key and returns
	// the remaining time until the item expires. If no valid value is stored for the
	// given key, it returns false.
	//
	// Returns:
	//   - T: the value stored into the cache for the given key.
	//   - time.Duration: the remaining time until the item expires.
	//   - bool: true if the value was found, false otherwise.
	//
	// If the item has expired, the function returns false.
	// If the item does not exist in the cache, the function returns false.
	//
	// It is safe to call LoadRemain while other goroutines are accessing the cache.
	LoadRemain() (T, time.Duration, bool)
}

// New creates a new cache item with the specified expiration duration and initial value.
// The item will expire after the specified duration from the time it was last stored.
// If expire is 0, the item will never expire.
func New[T any](expire time.Duration, val T) CacheItem[T] {
	o := &itm[T]{
		e: expire,
		k: new(atomic.Bool),
		t: libatm.NewValue[time.Time](),
		v: libatm.NewValue[T](),
	}

	o.clean(true)
	o.Store(val)

	return o
}
