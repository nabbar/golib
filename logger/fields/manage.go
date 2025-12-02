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
 *
 */

package fields

import libctx "github.com/nabbar/golib/context"

// Clean removes all key-value pairs from the Fields instance.
//
// This operation resets the Fields to an empty state while preserving the underlying
// context. It's a composite operation that requires external synchronization if used
// concurrently with other operations.
func (o *fldModel) Clean() {
	o.c.Clean()
}

// Get retrieves the value associated with the given key.
//
// Returns the value and true if the key exists, nil and false otherwise.
// This operation is thread-safe for concurrent access.
func (o *fldModel) Get(key string) (val interface{}, ok bool) {
	return o.c.Load(key)
}

// Store inserts or updates a key-value pair without returning Fields.
//
// This is a thread-safe operation suitable for direct storage when method
// chaining is not needed. It's functionally equivalent to Add() but without
// the return value.
func (o *fldModel) Store(key string, cfg interface{}) {
	o.c.Store(key, cfg)
}

// Delete removes the key-value pair for the given key and returns Fields for chaining.
//
// If the key doesn't exist, this is a no-op. This operation is thread-safe.
func (o *fldModel) Delete(key string) Fields {
	o.c.Delete(key)
	return o
}

// Merge combines all fields from the source Fields into the receiver.
//
// For duplicate keys, the source value overwrites the receiver's value.
// This is a composite operation requiring external synchronization if used
// concurrently. Returns the receiver for method chaining.
func (o *fldModel) Merge(f Fields) Fields {
	if f == nil || o == nil {
		return o
	}

	f.Walk(func(key string, val interface{}) bool {
		o.c.Store(key, val)
		return true
	})

	return o
}

// Walk iterates over all key-value pairs, calling the callback for each.
//
// Iteration continues until all pairs are visited or the callback returns false.
// The iteration order is not guaranteed. Thread-safe for concurrent reads.
func (o *fldModel) Walk(fct libctx.FuncWalk[string]) Fields {
	o.c.Walk(fct)
	return o
}

// WalkLimit iterates only over specified keys, calling the callback for each found key.
//
// Non-existent keys are silently skipped. More efficient than Walk when only specific
// fields are needed. Thread-safe for concurrent reads.
func (o *fldModel) WalkLimit(fct libctx.FuncWalk[string], validKeys ...string) Fields {
	o.c.WalkLimit(fct, validKeys...)
	return o
}

// LoadOrStore atomically loads an existing value or stores a new one.
//
// Returns the existing value and true if the key existed, or the stored value
// and false if the key was newly created. This operation is thread-safe and atomic.
func (o *fldModel) LoadOrStore(key string, cfg interface{}) (val interface{}, loaded bool) {
	return o.c.LoadOrStore(key, cfg)
}

// LoadAndDelete atomically loads and deletes a value.
//
// Returns the value and true if the key existed (and was deleted), or nil and false
// if the key didn't exist. This operation is thread-safe and atomic.
func (o *fldModel) LoadAndDelete(key string) (val interface{}, loaded bool) {
	return o.c.LoadAndDelete(key)
}
