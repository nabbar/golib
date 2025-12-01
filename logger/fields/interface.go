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

import (
	"context"
	"encoding/json"

	libctx "github.com/nabbar/golib/context"
	"github.com/sirupsen/logrus"
)

// Fields provides a thread-safe, context-aware structured logging fields management interface.
//
// This interface combines three key capabilities:
//  1. context.Context implementation for lifecycle management and cancellation propagation
//  2. json.Marshaler/Unmarshaler for serialization and persistence
//  3. Key-value storage with various access patterns
//
// Thread Safety:
// - Read operations (Get, Logrus, Walk) are thread-safe for concurrent access
// - Single write operations (Add, Store, Delete, LoadOrStore, LoadAndDelete) are thread-safe
//   thanks to the underlying sync.Map implementation
// - Composite operations (Map, Merge, Clean) require external synchronization if used concurrently
// - For concurrent composite operations, use Clone() to create independent instances per goroutine
//
// Context Integration:
// Fields fully implements context.Context, allowing it to participate in Go's cancellation
// and deadline mechanisms. The context provided to New() determines the lifecycle of the Fields
// instance.
//
// See example_test.go for comprehensive usage examples.
type Fields interface {
	context.Context
	json.Marshaler
	json.Unmarshaler

	// Clone creates a deep copy of the Fields instance.
	//
	// The returned Fields instance is completely independent and modifications to either
	// the original or the clone will not affect the other. This is essential for creating
	// derived field sets without side effects.
	//
	// Note: While the internal map is deep copied, the values themselves are not. If values
	// are pointers or references, modifications to the underlying data will affect all clones.
	//
	// Returns nil if the receiver is nil.
	//
	// Example:
	//   base := fields.New(ctx).Add("service", "api")
	//   request := base.Clone().Add("request_id", "123")
	//   // base still has only "service" field
	Clone() Fields

	// Clean removes all key-value pairs from the Fields instance.
	//
	// This is useful for resetting a Fields instance to empty state while preserving
	// the underlying context. After calling Clean(), the Fields instance can be reused.
	//
	// This is a composite operation that requires external synchronization if used
	// concurrently with other operations.
	Clean()

	// Add inserts or updates a key-value pair in the Fields instance.
	//
	// If the key already exists, its value is overwritten with the new value.
	// If the key does not exist, a new key-value pair is added.
	//
	// The method returns the same Fields instance to enable method chaining.
	// This operation is thread-safe and can be called concurrently from multiple goroutines.
	//
	// Any type can be stored as a value via interface{}, but consider JSON serialization
	// compatibility if persistence is needed.
	//
	// Example:
	//   flds.Add("key1", "value1").Add("key2", 42).Add("key3", true)
	Add(key string, val interface{}) Fields

	// Delete removes the key-value pair associated with the given key.
	//
	// If the key does not exist, this is a no-op (no error is returned).
	// The method returns the same Fields instance to enable method chaining.
	// This operation is thread-safe.
	//
	// Example:
	//   flds.Delete("temp_key").Delete("another_key")
	Delete(key string) Fields

	// Merge combines all key-value pairs from the source Fields into the receiver.
	//
	// For keys that exist in both Fields instances, the source value overwrites the
	// receiver's value. The source Fields instance is not modified.
	//
	// This is a composite operation that requires external synchronization if used
	// concurrently with other operations.
	//
	// Returns the receiver to enable method chaining.
	// Returns the receiver unchanged if source is nil.
	//
	// Example:
	//   base.Merge(extra)  // Adds all fields from extra to base
	Merge(f Fields) Fields

	// Walk iterates over all key-value pairs, calling the provided function for each pair.
	//
	// The iteration continues until either:
	//   - All pairs have been visited
	//   - The callback function returns false
	//
	// The iteration order is not guaranteed due to the underlying map implementation.
	//
	// Returns the receiver to enable method chaining.
	//
	// Example:
	//   flds.Walk(func(key string, val interface{}) bool {
	//       fmt.Printf("%s: %v\n", key, val)
	//       return true  // Continue iteration
	//   })
	Walk(fct libctx.FuncWalk[string]) Fields

	// WalkLimit iterates only over the specified keys, calling the provided function for each.
	//
	// Only the keys listed in validKeys will be visited. If a listed key does not exist,
	// it is silently skipped. This is more efficient than Walk when only specific fields
	// are needed.
	//
	// Returns the receiver to enable method chaining.
	//
	// Example:
	//   flds.WalkLimit(func(key string, val interface{}) bool {
	//       // Only processes "request_id" and "trace_id"
	//       return true
	//   }, "request_id", "trace_id")
	WalkLimit(fct libctx.FuncWalk[string], validKeys ...string) Fields

	// Get retrieves the value associated with the given key.
	//
	// Returns the value and true if the key exists, or nil and false if it does not.
	// This follows the standard Go map access pattern.
	//
	// Example:
	//   if val, ok := flds.Get("key"); ok {
	//       // Use val safely
	//   }
	Get(key string) (val interface{}, ok bool)

	// Store inserts or updates a key-value pair without returning the Fields instance.
	//
	// This method is similar to Add() but doesn't return the Fields instance, making it
	// suitable for use when method chaining is not needed. It's a direct storage operation.
	//
	// This operation is thread-safe and can be called concurrently from multiple goroutines
	// thanks to the underlying sync.Map implementation.
	//
	// Example:
	//   flds.Store("config_key", configValue)
	//   flds.Store("timestamp", time.Now())
	Store(key string, cfg interface{})

	// LoadOrStore atomically retrieves or stores a value for the given key.
	//
	// If the key exists, returns the existing value and true.
	// If the key does not exist, stores the provided value and returns it with false.
	//
	// This is useful for lazy initialization patterns where a default value should be
	// set only if the key doesn't already exist.
	//
	// Returns:
	//   - val: The existing value if loaded=true, or the stored value if loaded=false
	//   - loaded: true if the key existed, false if the value was stored
	//
	// Example:
	//   val, loaded := flds.LoadOrStore("counter", 0)
	//   if !loaded {
	//       // First access, counter was initialized to 0
	//   }
	LoadOrStore(key string, cfg interface{}) (val interface{}, loaded bool)

	// LoadAndDelete atomically retrieves and removes a value for the given key.
	//
	// If the key exists, returns the value and true, and the key is deleted.
	// If the key does not exist, returns nil and false.
	//
	// This is useful for one-time operations or cleanup scenarios.
	//
	// Returns:
	//   - val: The value if it existed, nil otherwise
	//   - loaded: true if the key existed and was deleted, false otherwise
	//
	// Example:
	//   if val, existed := flds.LoadAndDelete("temp"); existed {
	//       // Process val, field is now deleted
	//   }
	LoadAndDelete(key string) (val interface{}, loaded bool)

	// Logrus returns the logrus.Fields instance associated with the current Fields instance.
	//
	// This method is useful when you want to directly access the logrus.Fields instance
	// associated with the current Fields instance.
	//
	// The returned logrus.Fields instance is a reference to the same instance as the one
	// associated with the current Fields instance. Any modification to the returned logrus.Fields
	// instance will affect the Fields instance.
	//
	// The returned logrus.Fields instance is valid until the Fields instance is modified or
	// until the Fields instance is garbage collected.
	//
	// If the Fields instance is nil, this method will return nil.
	Logrus() logrus.Fields
	// Map applies a transformation function to all key-value pairs in the Fields instance.
	//
	// The transformation function is called for each key-value pair. It takes the key
	// and value as arguments, and returns the new value to store.
	//
	// This is a composite operation that requires external synchronization if used
	// concurrently with other operations. The transformation is applied in-place.
	//
	// Example:
	//   flds.Map(func(key string, val interface{}) interface{} {
	//       if key == "password" {
	//           return "[REDACTED]"
	//       }
	//       return val
	//   })
	Map(fct func(key string, val interface{}) interface{}) Fields
}

// New creates a new Fields instance from the given context.Context.
//
// It returns a new Fields instance which is associated with the given context.Context.
// The returned Fields instance can be used to add, remove, or modify key/val pairs.
//
// If the given context.Context is nil, this method will return nil.
//
// Example usage:
//
//	 flds := New(context.Background)
//	 flds.Add("key", "value")
//	 flds.Map(func(key string, val interface{}) interface{} {
//		return fmt.Sprintf("%s-%s", key, val)
//	 })
//
// The above example shows how to create a new Fields instance from a context.Context,
// and how to use the returned Fields instance to add, remove, or modify key/val pairs.
func New(ctx context.Context) Fields {
	return &fldModel{
		c: libctx.New[string](ctx),
	}
}
