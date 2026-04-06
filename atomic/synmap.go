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

// mt is the internal implementation of the MapTyped interface.
//
// # IMPLEMENTATION DETAILS
//
// This structure acts as a typed decorator around a Map[K]. It leverages generics
// [K, V] to ensure that both keys and values are strictly typed at the API level.
// Internal conversions are handled by the castBool and Cast utilities.
//
// Internal State:
//   - m (Map[K]): The underlying Map[K] which handles keys of type K and values of type 'any'.
//
// The MapTyped implementation provides the highest level of type safety by ensuring
// that all values are of type V before they are returned to the user or stored
// in the underlying map.
type mt[K comparable, V any] struct {
	m Map[K] // Underlying map with typed keys but arbitrary values.
}

// castBool is an internal helper to handle the common pattern of casting a value
// and returning it along with a success flag. It uses Cast[V] to ensure that
// any value retrieved from the underlying Map[K] matches the expected type V.
func (o *mt[K, V]) castBool(in any, chk bool) (value V, ok bool) {
	// If the underlying operation returned true (chk), we perform a type-safe cast.
	if v, k := Cast[V](in); k {
		return v, chk
	}

	// If the cast fails (e.g., if a value of an incorrect type was stored),
	// we return the zero value of type V and false.
	return value, false
}

// Len returns the current number of entries stored in the map.
// This is a direct delegation to the underlying Map[K].
func (o *mt[K, V]) Len() uint64 {
	return o.m.Len()
}

// Load retrieves the typed value for a key. Returns (zero-value, false) if not found.
func (o *mt[K, V]) Load(key K) (value V, ok bool) {
	return o.castBool(o.m.Load(key))
}

// Store sets the typed value for a key.
func (o *mt[K, V]) Store(key K, value V) {
	o.m.Store(key, value)
}

// LoadOrStore returns the existing typed value or stores the new one.
// The returned 'loaded' flag indicates if the value was found (true) or newly stored (false).
func (o *mt[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	return o.castBool(o.m.LoadOrStore(key, value))
}

// LoadAndDelete deletes the typed value for a key, returning the previous value.
// The 'loaded' flag indicates if an entry was found and removed.
func (o *mt[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	return o.castBool(o.m.LoadAndDelete(key))
}

// Delete removes the typed value for a key.
func (o *mt[K, V]) Delete(key K) {
	o.m.Delete(key)
}

// Swap exchanges the typed value for a key and returns the previous value.
// The 'loaded' flag indicates if the key was previously present.
func (o *mt[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	return o.castBool(o.m.Swap(key, value))
}

// CompareAndSwap performs a typed CAS operation.
// It updates the value for 'key' only if the current value matches 'old'.
func (o *mt[K, V]) CompareAndSwap(key K, old, new V) bool {
	return o.m.CompareAndSwap(key, old, new)
}

// CompareAndDelete performs a typed conditional delete.
// It removes the entry for 'key' only if the current value matches 'old'.
func (o *mt[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return o.m.CompareAndDelete(key, old)
}

// Range iterates over the map, providing typed keys and values to the function 'f'.
//
// # SELF-HEALING & INTEGRITY GUARD
//
// Like the MapAny implementation, this Range method includes an integrity check.
// If an entry with a value that cannot be cast to type V is encountered (e.g., if
// an incorrect value was somehow injected into the map), it is automatically
// removed from the map.
//
// This ensures that the user-provided function 'f' only receives valid, strictly-typed data.
func (o *mt[K, V]) Range(f func(key K, value V) bool) {
	o.m.Range(func(key K, value any) bool {
		var (
			l bool
			v V
		)

		// Validate value type before calling the user function.
		if v, l = Cast[V](value); !l {
			// Evict invalid values to ensure map consistency.
			o.m.Delete(key)
			return true
		}

		// Execute user-defined function for the typed entry.
		return f(key, v)
	})
}
