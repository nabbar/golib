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

/*
Package fields provides a thread-safe, context-aware structured logging fields management system
that seamlessly integrates with logrus and Go's standard context package.

# Design Philosophy

The fields package is designed around three core principles:

1. Context Integration: Full implementation of context.Context interface for seamless integration
with Go's cancellation and deadline propagation mechanisms.

2. Type Safety: Generic-based implementation ensuring type-safe operations while maintaining
flexibility for any value type through interface{}.

3. Immutability Options: Support for both mutable operations (Add, Delete, Map) and immutable
patterns (Clone) to suit different use cases.

# Architecture

The package consists of three main components:

1. Fields Interface: Public API providing logging field operations, context methods, and JSON
serialization capabilities.

2. fldModel Implementation: Internal structure wrapping github.com/nabbar/golib/context.Config[string]
for thread-safe storage and retrieval of key-value pairs.

3. Integration Layer: Bidirectional conversion between Fields and logrus.Fields for compatibility
with existing logging infrastructure.

# Key Features

Thread Safety: Read operations (Get, Logrus, Walk) are thread-safe for concurrent access.
Single write operations (Add, Store, Delete, LoadOrStore, LoadAndDelete) are also thread-safe
thanks to the underlying sync.Map implementation. Composite operations (Map, Merge, Clean)
require external synchronization if used concurrently from multiple goroutines.

Context Propagation: Full context.Context implementation allows Fields to participate in Go's
cancellation and deadline mechanisms, enabling proper cleanup in long-running operations.

JSON Serialization: Built-in JSON marshaling and unmarshaling support for persistence, network
transmission, or configuration storage.

Flexible Operations:
  - Add/Store: Insert or update key-value pairs
  - Get/LoadOrStore: Retrieve values with optional defaults
  - Delete/LoadAndDelete: Remove entries with optional retrieval
  - Walk/WalkLimit: Iterate over entries with filtering support
  - Map: Transform all values using a custom function
  - Merge: Combine multiple Fields instances

Logrus Integration: Direct conversion to/from logrus.Fields for zero-friction integration with
existing logrus-based logging systems.

# Performance Characteristics

Memory Usage: O(n) where n is the number of fields stored. Each field requires storage for both
key (string) and value (interface{}). The underlying sync.Map implementation adds minimal overhead.

Time Complexity:
  - Add, Get, Delete: O(1) average case (map operations)
  - Walk, WalkLimit: O(n) where n is number of fields
  - Map: O(n) with additional transformation cost
  - Clone: O(n) deep copy operation
  - Logrus: O(n) conversion to logrus.Fields
  - Merge: O(m) where m is number of fields in source

Thread Safety: Lock-free reads with atomic operations. Writes use appropriate synchronization
mechanisms provided by the underlying context.Config implementation.

# Use Cases

Structured Logging: Create field sets for structured log entries with consistent formatting and
easy transformation.

	flds := fields.New(ctx)
	flds.Add("request_id", "abc123")
	flds.Add("user_id", 42)
	flds.Add("action", "login")
	logger.WithFields(flds.Logrus()).Info("User action")

Request Context Enrichment: Attach metadata to request contexts for distributed tracing and
debugging.

	flds := fields.New(r.Context())
	flds.Add("trace_id", traceID)
	flds.Add("span_id", spanID)
	newCtx := context.WithValue(r.Context(), "fields", flds)

Field Transformation: Apply transformations to all field values for sanitization, formatting,
or encoding.

	flds.Map(func(key string, val interface{}) interface{} {
		if key == "password" {
			return "[REDACTED]"
		}
		return val
	})

Multi-Source Aggregation: Combine fields from multiple sources while maintaining immutability
of originals.

	base := fields.New(ctx).Add("service", "api")
	request := base.Clone().Add("method", "POST")
	response := request.Clone().Add("status", 200)

Configuration Serialization: Store and retrieve field configurations using JSON.

	data, _ := json.Marshal(flds)
	restored := fields.New(ctx)
	json.Unmarshal(data, restored)

# Limitations and Considerations

Context Lifetime: Fields instances hold a reference to their context. Ensure proper context
cancellation to avoid resource leaks in long-running applications.

Value Types: While any type can be stored via interface{}, complex types (structs, pointers)
should be used carefully. JSON serialization may not preserve all type information.

Nil Handling: The package handles nil Fields instances gracefully, returning nil or empty
results. However, this behavior should not be relied upon in production code. Always check
for nil before operations.

Deep Copy Behavior: Clone() performs a deep copy of the internal map but does not deep copy
the values themselves. If values are pointers or references, modifications to the underlying
data will affect all clones.

logrus.Fields Conversion: The Logrus() method creates a new map on each call. For
performance-critical code, cache the result if multiple accesses are needed.

Walk Function Behavior: Walk and WalkLimit continue iteration until the callback returns false
or all entries are processed. The iteration order is not guaranteed due to the underlying map
implementation.

# Best Practices

Prefer Clone for Immutability: When creating derived field sets, use Clone() to avoid
unintended modifications to the original.

	derived := original.Clone().Add("extra", "field")

Use WalkLimit for Selective Operations: When only specific fields are needed, use WalkLimit
to improve performance.

	flds.WalkLimit(func(key string, val interface{}) bool {
		// Process only specified keys
		return true
	}, "request_id", "trace_id")

Context-Aware Cleanup: Always pass an appropriate context to New() for proper lifecycle
management.

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	flds := fields.New(ctx)

Avoid Storing Large Values: Fields are designed for metadata, not data storage. Keep values
small and reference larger data structures by ID or key.

Type Assertions with Care: When retrieving values, always check type assertions to avoid panics.

	if val, ok := flds.Get("key"); ok {
		if str, ok := val.(string); ok {
			// Use str safely
		}
	}

# Thread Safety Model

The Fields implementation provides strong thread-safety guarantees:

Thread-Safe Operations (no synchronization needed):
  - Read operations: Get, Logrus, Walk, WalkLimit
  - Single write operations: Add, Store, Delete, LoadOrStore, LoadAndDelete
  - Clone() creates independent instances safe for parallel modification

Composite Operations (require external synchronization if concurrent):
  - Map: iterates and modifies all values
  - Merge: iterates source and stores in receiver
  - Clean: removes all entries

Safe Concurrent Patterns:

	// Multiple goroutines can safely call Add/Store/Delete
	go func() { flds.Add("key1", "value1") }()
	go func() { flds.Add("key2", "value2") }()

	// For composite operations, use Clone per goroutine
	go func() {
		local := flds.Clone()
		local.Map(transformFunc)
	}()

# Examples

See the example_test.go file for comprehensive examples demonstrating:
  - Basic field creation and manipulation
  - Integration with logrus logger
  - Context propagation and cancellation
  - JSON serialization and deserialization
  - Field transformation with Map
  - Merging multiple field sources
  - Walking and filtering operations

# Related Packages

  - github.com/nabbar/golib/context: Underlying context management
  - github.com/sirupsen/logrus: Logging framework integration
  - encoding/json: Standard JSON serialization

# Versioning and Compatibility

This package follows semantic versioning. The API is stable and backwards compatible within
major versions. Breaking changes will only occur in major version increments.

Minimum Go version: 1.18 (requires generics support)

# License

MIT License - See LICENSE file for details.

Copyright (c) 2025 Nicolas JUHEL
*/
package fields
