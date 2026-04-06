/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

// Package atomic provides type-safe wrappers around sync/atomic operations.
//
// This file defines the public interfaces and factory functions for the package.
// The design focuses on providing the convenience of Go generics while maintaining
// the strict performance requirements of low-level atomic primitives.
package atomic

import (
	"sync"
	"sync/atomic"
)

// Value provides a type-safe interface for atomic operations on a single value of type T.
// It acts as a generic-aware container that eliminates the need for manual type assertions
// when working with sync/atomic.Value.
//
// Features:
//   - Full type safety for Load, Store, Swap, and CAS operations.
//   - Optional default values for Load() when the container is empty.
//   - Optional default values for Store() to prevent storing zero-values.
//   - Tiered performance model that bypasses validation logic when defaults are not set.
//
// Thread Safety:
// All methods are safe for concurrent use by multiple goroutines without additional locking.
type Value[T any] interface {
	// SetDefaultLoad configures the value to be returned by Load() if the underlying
	// storage is uninitialized or contains a zero-value (as defined by the Cast utility).
	// Calling this method enables the validation path for Load() and Swap() operations.
	SetDefaultLoad(def T)

	// SetDefaultStore configures a value that will be automatically stored by Store(),
	// Swap(), and CompareAndSwap() if the user attempts to store a zero-value of type T.
	// This ensures that the container never transitions to an "empty" state.
	SetDefaultStore(def T)

	// Load atomically retrieves the current value.
	// If a default load value is set and the current value is empty, the default is returned.
	Load() (val T)

	// Store atomically updates the current value.
	// If a default store value is set and 'val' is empty, the default is stored instead.
	Store(val T)

	// Swap atomically stores a new value and returns the previous one.
	// It respects both SetDefaultLoad and SetDefaultStore configurations.
	Swap(new T) (old T)

	// CompareAndSwap performs an atomic compare-and-swap operation.
	// It only succeeds if the current value matches 'old'. If a default store value
	// is configured, it is applied to both 'old' and 'new' before comparison.
	CompareAndSwap(old, new T) (swapped bool)
}

// Map provides a thread-safe map with a fixed key type and arbitrary value types (any).
// It is a type-safe wrapper around sync.Map that adds a constant-time length counter.
//
// Implementation Note:
// The keys must satisfy the 'comparable' constraint as required by Go maps.
// While values are 'any', the Map implementation ensures that keys are correctly
// typed during iteration and internal operations.
type Map[K comparable] interface {
	// Len returns the number of entries currently stored in the map.
	// This is an O(1) operation maintained via atomic increments/decrements.
	Len() uint64

	// Load retrieves the value for a key. Returns (nil, false) if the key is not found.
	Load(key K) (value any, ok bool)

	// Store sets the value for a key. If the key already exists, the value is updated.
	Store(key K, value any)

	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The 'loaded' result is true if the value was loaded, false if stored.
	LoadOrStore(key K, value any) (actual any, loaded bool)

	// LoadAndDelete deletes the value for a key, returning the previous value if any.
	// The 'loaded' result is true if the key was present in the map.
	LoadAndDelete(key K) (value any, loaded bool)

	// Delete removes the value for a key from the map.
	Delete(key K)

	// Swap exchanges the value for a key and returns the previous value.
	// The 'loaded' result is true if the key was previously present.
	Swap(key K, value any) (previous any, loaded bool)

	// CompareAndSwap performs a CAS operation on a map entry.
	// It only updates the value if the current value matches 'old'.
	CompareAndSwap(key K, old, new any) bool

	// CompareAndDelete deletes the entry if the current value matches 'old'.
	CompareAndDelete(key K, old any) (deleted bool)

	// Range iterates over the map. The function 'f' is called for each key/value pair.
	// If 'f' returns false, iteration stops.
	// Range also performs a "self-healing" check: any key found with an invalid type
	// is automatically evicted from the map.
	Range(f func(key K, value any) bool)
}

// MapTyped provides a fully type-safe thread-safe map for both keys and values.
// It extends the Map[K] interface by enforcing a specific type V for all values.
type MapTyped[K comparable, V any] interface {
	// Len returns the number of entries currently stored in the map.
	Len() uint64

	// Load retrieves the typed value for a key.
	Load(key K) (value V, ok bool)

	// Store sets the typed value for a key.
	Store(key K, value V)

	// LoadOrStore returns the existing typed value or stores the new one.
	LoadOrStore(key K, value V) (actual V, loaded bool)

	// LoadAndDelete deletes the typed value for a key.
	LoadAndDelete(key K) (value V, loaded bool)

	// Delete removes the typed value for a key.
	Delete(key K)

	// Swap exchanges the typed value for a key and returns the previous value.
	Swap(key K, value V) (previous V, loaded bool)

	// CompareAndSwap performs a typed CAS operation.
	CompareAndSwap(key K, old, new V) bool

	// CompareAndDelete performs a typed conditional delete.
	CompareAndDelete(key K, old V) (deleted bool)

	// Range iterates over typed entries.
	// Range performs a "self-healing" check: any value found that cannot be cast
	// to type V is automatically evicted from the map.
	Range(f func(key K, value V) bool)
}

// NewValue initializes and returns a new high-performance atomic container for type T.
//
// Performance Note:
// This constructor creates a Value instance optimized for speed. Until SetDefaultLoad
// or SetDefaultStore are called, the implementation uses a "fast path" that
// skips all validation logic, matching the performance of sync/atomic.Value.
func NewValue[T any]() Value[T] {
	return &val[T]{
		av: new(atomic.Value),
		dl: new(atomic.Value),
		ds: new(atomic.Value),
		bl: new(atomic.Bool),
		bs: new(atomic.Bool),
	}
}

// NewValueDefault creates an atomic box with pre-configured default values for both
// Load and Store operations.
//
// Note: Using this constructor immediately enables the validation logic path,
// which involves type casting and zero-value detection.
func NewValueDefault[T any](load, store T) Value[T] {
	o := NewValue[T]()
	o.SetDefaultLoad(load)
	o.SetDefaultStore(store)
	return o
}

// NewMapAny creates a thread-safe map with a specific key type K and arbitrary values.
// It initializes the atomic length counter and the underlying sync.Map.
func NewMapAny[K comparable]() Map[K] {
	return &ma[K]{
		l: new(atomic.Int64),
		m: sync.Map{},
	}
}

// NewMapTyped creates a fully type-safe thread-safe map for keys of type K and
// values of type V.
func NewMapTyped[K comparable, V any]() MapTyped[K, V] {
	return &mt[K, V]{
		m: NewMapAny[K](),
	}
}
