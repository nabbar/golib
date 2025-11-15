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

package atomic

import (
	"sync"
	"sync/atomic"
)

type Value[T any] interface {
	// SetDefaultLoad sets the default load value for this Value.
	// The default value is returned when Load is called and the value is not present in the underlying store.
	//
	// Note: SetDefaultLoad should be called before first use of Load.
	SetDefaultLoad(def T)
	// SetDefaultStore sets the default store value for this Value.
	// The default value is used when Store is called with a value of zero.
	// Note: SetDefaultStore should be called before first use of Store.
	SetDefaultStore(def T)

	// Load returns the value stored in the underlying store for this Value.
	// If no value is present, the default load value (set by SetDefaultLoad) is returned.
	// Note: Load will return the default load value until the first successful call to Store.
	Load() (val T)
	// Store sets the value for the given key in the underlying store for this Value.
	// Note: Store will use the default store value (set by SetDefaultStore) if the value passed is zero.
	//
	// If the value is not zero, the underlying store will be updated with the new value.
	//
	// Example:
	//  v := NewValue[int]()
	//  v.SetDefaultStore(42)
	//  v.Store(0) // sets 42 as the value in the underlying store
	//  v.Store(99) // sets 99 as the value in the underlying store
	Store(val T)
	// Swap atomically swaps the value of the underlying store for this Value with the given new value.
	// It returns the previous value stored in the underlying store.
	// If the previous value is zero, the default store value (set by SetDefaultStore) is returned.
	//
	// Example:
	//  v := NewValue[int]()
	//  v.SetDefaultStore(42)
	//  old, _ := v.Swap(0) // old is 42
	Swap(new T) (old T)
	// CompareAndSwap atomically compares the value stored in the underlying store for this Value
	// with the given old value. If they are equal, it atomically swaps the value with the given new value.
	// It returns true if the swap was successful, or false otherwise.
	//
	// Note: If the old value is zero, the default store value (set by SetDefaultStore) is used for comparison.
	// If the new value is zero, the default store value (set by SetDefaultStore) is used for swapping.
	//
	// Example:
	//  v := NewValue[int]()
	//  v.SetDefaultStore(42)
	//  swapped := v.CompareAndSwap(0, 99) // swapped is true, and the value in the underlying store is 99
	CompareAndSwap(old, new T) (swapped bool)
}

type Map[K comparable] interface {
	// Load returns the value stored in the underlying store for this Map.
	// If no value is present, ok is false.
	// Otherwise, ok is true and the value is returned.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, ok := m.Load("key")
	//  fmt.Println(val, ok) // prints "value true"
	Load(key K) (value any, ok bool)
	// Store atomically stores the given value in the underlying store for this Map.
	// It overwrites any existing value stored in the underlying store for the given key.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, ok := m.Load("key")
	//  fmt.Println(val, ok) // prints "value true"
	Store(key K, value any)

	// LoadOrStore atomically loads the value stored in the underlying store for this Map with the given key.
	// If no value is present, it atomically stores the given value in the underlying store for this Map,
	// and returns the given value along with true for the loaded flag.
	// If a value is present, it returns the present value along with true for the loaded flag.
	// Note: If the given value is zero, the default store value (set by SetDefaultStore) is used for storing.
	//
	// Example:
	//  m := NewMap[string]()
	//  val, loaded := m.LoadOrStore("key", "value")
	//  fmt.Println(val, loaded) // prints "value true"
	LoadOrStore(key K, value any) (actual any, loaded bool)
	// LoadAndDelete atomically loads and deletes the value stored in the underlying store for this Map.
	// It returns the value stored in the underlying store for the given key, along with true if the value was present.
	// If the value was not present, the default load value (set by SetDefaultLoad) is returned, and the loaded flag is false.
	//
	// Note: If the value is not zero, the underlying store will be updated with the new value.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, loaded := m.LoadAndDelete("key")
	//  fmt.Println(val, loaded) // prints "value true"
	LoadAndDelete(key K) (value any, loaded bool)

	// Delete atomically deletes the value stored in the underlying store for this Map with the given key.
	// It returns true if the value was present, and false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  deleted := m.Delete("key")
	//  fmt.Println(deleted) // prints "true"
	Delete(key K)
	// Swap atomically swaps the value stored in the underlying store for this Map with the given key
	// with the given value. It returns the previous value stored in the underlying store for the given
	// key, and a boolean indicating whether the swap was successful.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "old")
	//  prev, swapped := m.Swap("key", "new")
	//  fmt.Println(prev, swapped) // prints "old true"
	Swap(key K, value any) (previous any, loaded bool)

	// CompareAndSwap atomically compares the value stored in the underlying store for this Map with the given key
	// with the given old value. If they are equal, it atomically swaps the value with the given new value.
	// It returns true if the swap was successful, or false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "old")
	//  swapped := m.CompareAndSwap("key", "old", "new")
	//  fmt.Println(swapped) // prints "true"
	CompareAndSwap(key K, old, new any) bool
	// CompareAndDelete atomically compares the value stored in the underlying store for this Map with the given key
	// with the given old value. If they are equal, it atomically deletes the value from the underlying store.
	// It returns true if the value was present and deleted, and false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  deleted := m.CompareAndDelete("key", "value")
	//  fmt.Println(deleted) // prints "true"
	CompareAndDelete(key K, old any) (deleted bool)

	// Range calls the given function for each key in the underlying store.
	// It calls the function in an unspecified order, and will stop calling the function
	// if the function returns false.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  m.Range(func(k string, v any) bool {
	//      fmt.Println(k, v)
	//      return true
	//  })
	//  // prints "key value"
	Range(f func(key K, value any) bool)
}

type MapTyped[K comparable, V any] interface {
	// Load returns the value stored in the underlying store for this Map with the given key.
	// If no value is present, ok is false.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, ok := m.Load("key")
	//  fmt.Println(val, ok) // prints "value true"
	Load(key K) (value V, ok bool)
	// Store atomically stores the given value in the underlying store for this Map with the given key.
	// It overwrites any existing value stored in the underlying store for the given key.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, ok := m.Load("key")
	//  fmt.Println(val, ok) // prints "value true"
	Store(key K, value V)

	// LoadOrStore atomically loads the value stored in the underlying store for this Map with the given key.
	// If no value is present, it atomically stores the given value in the underlying store for this Map,
	// and returns the given value along with true for the loaded flag.
	// If a value is present, it returns the present value along with true for the loaded flag.
	//
	// Note: If the given value is zero, the default store value (set by SetDefaultStore) is used for storing.
	//
	// Example:
	//  m := NewMap[string]()
	//  val, loaded := m.LoadOrStore("key", "value")
	//  fmt.Println(val, loaded) // prints "value true"
	LoadOrStore(key K, value V) (actual V, loaded bool)
	// LoadAndDelete atomically loads and deletes the value stored in the underlying store for this Map with the given key.
	// It returns the value stored in the underlying store for the given key, along with true if the value was present.
	// If the value was not present, the default load value (set by SetDefaultLoad) is returned, and the loaded flag is false.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  val, loaded := m.LoadAndDelete("key")
	//  fmt.Println(val, loaded) // prints "value true"
	LoadAndDelete(key K) (value V, loaded bool)

	// Delete atomically deletes the value stored in the underlying store for this Map with the given key.
	// It returns true if the value was present, and false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  deleted := m.Delete("key")
	//  fmt.Println(deleted) // prints "true"
	Delete(key K)
	// Swap atomically swaps the value stored in the underlying store for this Map with the given key
	// with the given new value. It returns the previous value stored in the underlying store,
	// along with true if the swap was successful, or false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "old")
	//  val, loaded := m.Swap("key", "new")
	//  fmt.Println(val, loaded) // prints "old true"
	Swap(key K, value V) (previous V, loaded bool)

	// CompareAndSwap atomically compares the value stored in the underlying store for this Map with the given key
	// with the given old value. If they are equal, it atomically swaps the value with the given new value.
	// It returns true if the swap was successful, or false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "old")
	//  swapped := m.CompareAndSwap("key", "old", "new")
	//  fmt.Println(swapped) // prints "true"
	CompareAndSwap(key K, old, new V) bool
	// CompareAndDelete atomically compares the value stored in the underlying store for this Map with the given key
	// with the given old value. If they are equal, it atomically deletes the value from the underlying store.
	// It returns true if the value was present and deleted, and false otherwise.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "old")
	//  deleted := m.CompareAndDelete("key", "old")
	//  fmt.Println(deleted) // prints "true"
	CompareAndDelete(key K, old V) (deleted bool)

	// Range calls the given function for each key in the underlying store.
	// It calls the function in an unspecified order, and will stop calling the function
	// if the function returns false.
	//
	// Example:
	//  m := NewMap[string]()
	//  m.Store("key", "value")
	//  m.Range(func(k string, v any) bool {
	//      fmt.Println(k, v)
	//      return true
	//  })
	//  // prints "key value"
	Range(f func(key K, value V) bool)
}

// NewValue returns a new Value with the given type. The default load value is the zero value
// of the given type, and the default store value is the zero value of the given type.
//
// Example:
//
//	v := NewValue[int]()
//	// v is a Value with default load value 0 and default store value 0.
func NewValue[T any]() Value[T] {
	var (
		tmp1 T
		tmp2 T
	)

	return NewValueDefault[T](tmp1, tmp2)
}

// NewValueDefault returns a new Value with the given type, default load value, and default store value.
// The default load value is the value passed to the load parameter, and the default store value is the value
// passed to the store parameter.
//
// Example:
//
//	v := NewValueDefault[int](0, 42)
//	// v is a Value with default load value 0 and default store value 42.
func NewValueDefault[T any](load, store T) Value[T] {
	o := &val[T]{
		av: new(atomic.Value),
		dl: new(atomic.Value),
		ds: new(atomic.Value),
	}

	o.SetDefaultLoad(load)
	o.SetDefaultStore(store)

	return o
}

// NewMapAny returns a new Map with the given key type. It uses a sync.Map as the underlying store.
//
// Example:
//
//	m := NewMapAny[int]()
//	// m is a Map with key type int and underlying store sync.Map{}.
func NewMapAny[K comparable]() Map[K] {
	return &ma[K]{
		m: sync.Map{},
	}
}

// NewMapTyped returns a new Map with the given key type and value type.
// It uses a sync.Map as the underlying store.
//
// Example:
//
//	m := NewMapTyped[int, string]()
//	// m is a Map with key type int and value type string, and underlying store sync.Map{}.
func NewMapTyped[K comparable, V any]() MapTyped[K, V] {
	return &mt[K, V]{
		m: NewMapAny[K](),
	}
}
