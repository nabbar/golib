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

// Package cache provides a high-performance, thread-safe caching solution with automatic expiration management.
// Built with Go generics, it supports any comparable key type and any value type, making it flexible for various use cases.
//
// Key Features:
//   - Generic Support: Works with any comparable key type and any value type
//   - Thread-Safe: All operations are safe for concurrent access
//   - Automatic Expiration: Items expire automatically after a configurable duration
//   - Context Integration: Built on context.Context for lifecycle management
//   - High Performance: Optimized for concurrent access with minimal lock contention
//   - Memory Efficient: Automatic cleanup of expired items
package cache

import (
	"context"
	"io"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	cchitm "github.com/nabbar/golib/cache/item"
)

// FuncCache is a function type that returns a Cache instance.
// It is useful for lazy initialization or factory patterns.
type FuncCache[K comparable, V any] func() Cache[K, V]

// Generic defines the base interface for cache operations that are independent of the key-value types.
// It combines context.Context and io.Closer interfaces with cache-specific cleanup methods.
type Generic interface {
	context.Context
	io.Closer

	// Clean removes all expired items from the cache.
	// It is safe to call Clean while other goroutines are accessing the cache.
	Clean()

	// Expire is used to check all stored items and clean expired items.
	// It is safe to call Expire while other goroutines are accessing the cache.
	Expire()
}

// Cache is the main interface for interacting with a typed cache.
// It provides type-safe storage and retrieval of key-value pairs with automatic expiration.
//
// The cache is generic and works with any comparable key type K and any value type V.
// All operations are thread-safe and can be called concurrently from multiple goroutines.
//
// Example:
//
//	cache := cache.New[string, int](ctx, 5*time.Minute)
//	defer cache.Close()
//	cache.Store("key", 42)
//	value, remaining, ok := cache.Load("key")
type Cache[K comparable, V any] interface {
	Generic

	// Clone returns a new Cache with the same content as the current one.
	//
	// Parameters:
	// - ctx: the context to use for the new cache.
	//
	// Returns:
	//   - Cache[K, V]: the new cache.
	//   - error: an error if something went wrong.
	//
	// The new cache is a deep copy of the current one, which means that it has
	// its own copy of the data and is not affected by changes made to the current
	// cache.
	Clone(context.Context) (Cache[K, V], error)

	// Merge merges the given cache into the current one.
	//
	// Parameters:
	// - c: the cache to merge into the current one.
	//
	// The items from the given cache are added to the current one.
	// If an item with the same key already exists in the current cache, it is
	// not overwritten by the item from the given cache.
	//
	// It is safe to call Merge while other goroutines are accessing the cache.
	//
	// Returns:
	//   - None
	Merge(Cache[K, V])

	// Walk walks over all stored items in the cache and calls the given function
	// for each item.
	//
	// Parameters:
	// - fct: the function to call for each item.
	//
	// The function is called with the key, value and remaining time of each item.
	// If the function returns false, Walk stops walking over the items and returns.
	//
	// It is safe to call Walk while other goroutines are accessing the cache.
	//
	// Returns:
	//   - None
	Walk(func(K, V, time.Duration) bool)

	// Load returns the value stored into the cache for the given key.
	//
	// Parameters:
	// - K: the key to use to retrieve the value from the cache.
	//
	// Returns:
	//   - V: the value stored into the cache for the given key.
	//   - time.Duration: the remaining time until the item expires.
	//   - bool: true if the value was found and valid, false otherwise.
	//
	// If the item has expired, it will be removed from the cache automatically.
	// If the item does not exist in the cache, the function returns false.
	//
	// It is safe to call Load while other goroutines are accessing the cache.
	Load(K) (V, time.Duration, bool)

	// Store stores the given value into the cache for the given key.
	//
	// Parameters:
	// - K: the key to use to store the value into the cache.
	// - V: the value to store into the cache.
	//
	// The stored item expires after the expiration time set by the New
	// function when creating the cache. If the expiration time is set to 0,
	// the item will never expire.
	//
	// It is safe to call Store while other goroutines are accessing the cache.
	//
	// Returns:
	//   - None
	Store(K, V)

	// Delete removes the item from the cache for the given key.
	//
	// Parameters:
	// - K: the key to use to remove the item from the cache.
	//
	// It is safe to call Delete while other goroutines are accessing the cache.
	//
	// Returns:
	//   - None
	Delete(K)

	// LoadOrStore loads the value stored into the cache for the given key.
	// If no value is stored for the given key, it stores the given value into the cache
	// and returns the value and the remaining time until the item expires.
	//
	// Parameters:
	// - K: the key to use to retrieve the value from the cache.
	// - V: the value to store into the cache if no value is stored for the given key.
	//
	// Returns:
	//   - V: the value stored into the cache for the given key.
	//   - time.Duration: the remaining time until the item expires.
	//   - bool: true if the value was found and valid, false otherwise.
	//
	// If the old item has expired, the function return false.
	// If the old item does not exist in the cache, the function returns false.
	//
	// It is safe to call LoadOrStore while other goroutines are accessing the cache.
	LoadOrStore(K, V) (V, time.Duration, bool)

	// LoadAndDelete loads the value stored into the cache for the given key and
	// deletes the item from the cache. If no value is stored for the given key,
	// it returns false.
	//
	// Parameters:
	// - K: the key to use to load the value from the cache.
	//
	// Returns:
	//   - V: the value stored into the cache for the given key.
	//   - bool: true if the value was found and valid, false otherwise.
	//
	// It is safe to call LoadAndDelete while other goroutines are accessing the cache.
	LoadAndDelete(K) (V, bool)

	// Swap loads the value stored into the cache for the given key, stores the given
	// value into the cache for the given key and returns the value and the remaining
	// time until the item expires. If no valid value is stored for the given key,
	// it returns false.
	//
	// Parameters:
	// - K: the key to use to load the value from the cache.
	// - V: the value to store into the cache for the given key.
	//
	// Returns:
	//   - V: the value stored into the cache for the given key.
	//   - time.Duration: the remaining time until the item expires.
	//   - bool: true if the value was found and valid, false otherwise.
	//
	// If the old item has expired, the function returns false.
	// If the old item does not exist in the cache, the function returns false.
	//
	// It is safe to call Swap while other goroutines are accessing the cache.
	Swap(key K, val V) (V, time.Duration, bool)
}

// New returns a new Cache with the given expiration time.
//
// Parameters:
// - ctx: the context to use for the new cache. If nil, context.Background() is used.
// - exp: the expiration time for the cache. If 0, items will never expire.
//
// Returns:
//   - Cache[K, V]: the new cache.
//
// The new cache is a deep copy of the current one, which means that it has
// its own copy of the data and is not affected by changes made to the current
// cache.
func New[K comparable, V any](ctx context.Context, exp time.Duration) Cache[K, V] {
	if ctx == nil {
		ctx = context.Background()
	}

	var cnl context.CancelFunc
	ctx, cnl = context.WithCancel(ctx)

	n := &cc[K, V]{
		Context: ctx,

		n: cnl,
		v: libatm.NewMapTyped[K, cchitm.CacheItem[V]](),
		e: exp,
	}

	return n
}
