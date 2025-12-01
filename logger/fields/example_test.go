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

package fields_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nabbar/golib/logger/fields"
)

// ExampleNew demonstrates basic Fields creation.
// This is the simplest use case for creating a new Fields instance.
func ExampleNew() {
	// Create a new Fields instance with background context
	flds := fields.New(context.Background())

	// Add a simple field
	flds.Add("message", "hello")

	// Get logrus compatible fields
	fmt.Println(len(flds.Logrus()))

	// Output:
	// 1
}

// ExampleFields_Add demonstrates adding fields to a Fields instance.
// This shows the basic field addition with method chaining.
func ExampleFields_Add() {
	flds := fields.New(nil)

	// Add fields using method chaining
	flds.Add("service", "api").
		Add("version", "1.0.0").
		Add("port", 8080)

	// Retrieve fields
	fmt.Println(len(flds.Logrus()))

	// Output:
	// 3
}

// ExampleFields_Add_overwrite demonstrates overwriting existing fields.
func ExampleFields_Add_overwrite() {
	flds := fields.New(nil)

	// Add initial value
	flds.Add("status", "pending")

	// Overwrite with new value
	flds.Add("status", "completed")

	// Retrieve the updated value
	if val, ok := flds.Get("status"); ok {
		fmt.Println(val)
	}

	// Output:
	// completed
}

// ExampleFields_Store demonstrates direct storage without method chaining.
// Store is useful when you don't need the returned Fields instance.
func ExampleFields_Store() {
	flds := fields.New(nil)

	// Store fields directly (no return value)
	flds.Store("config_key", "config_value")
	flds.Store("setting", 42)
	flds.Store("enabled", true)

	// Verify stored values
	fmt.Println(len(flds.Logrus()))

	// Output:
	// 3
}

// ExampleFields_Get demonstrates retrieving field values.
func ExampleFields_Get() {
	flds := fields.New(nil)
	flds.Add("user_id", 12345)
	flds.Add("username", "john_doe")

	// Get existing field
	if val, ok := flds.Get("user_id"); ok {
		fmt.Printf("User ID: %v\n", val)
	}

	// Get non-existent field
	if _, ok := flds.Get("non_existent"); !ok {
		fmt.Println("Field not found")
	}

	// Output:
	// User ID: 12345
	// Field not found
}

// ExampleFields_Delete demonstrates deleting fields.
func ExampleFields_Delete() {
	flds := fields.New(nil)
	flds.Add("temp_field", "temporary")
	flds.Add("keep_field", "permanent")

	// Delete a field
	flds.Delete("temp_field")

	fmt.Println(len(flds.Logrus()))

	// Output:
	// 1
}

// ExampleFields_Clone demonstrates creating independent copies.
// This shows how to create derived field sets without affecting the original.
func ExampleFields_Clone() {
	// Original fields
	original := fields.New(nil)
	original.Add("base", "value")

	// Clone creates independent copy
	clone := original.Clone()
	clone.Add("extra", "data")

	// Original remains unchanged
	fmt.Printf("Original: %d fields\n", len(original.Logrus()))
	fmt.Printf("Clone: %d fields\n", len(clone.Logrus()))

	// Output:
	// Original: 1 fields
	// Clone: 2 fields
}

// ExampleFields_Merge demonstrates merging multiple Fields instances.
func ExampleFields_Merge() {
	base := fields.New(nil)
	base.Add("service", "api")
	base.Add("env", "production")

	extra := fields.New(nil)
	extra.Add("version", "2.0")
	extra.Add("region", "eu-west")

	// Merge extra into base
	base.Merge(extra)

	fmt.Println(len(base.Logrus()))

	// Output:
	// 4
}

// ExampleFields_Map demonstrates transforming field values.
// This shows how to apply transformations to all fields.
func ExampleFields_Map() {
	flds := fields.New(nil)
	flds.Add("name", "john")
	flds.Add("city", "paris")

	// Transform all values to uppercase
	flds.Map(func(key string, val interface{}) interface{} {
		if str, ok := val.(string); ok {
			return fmt.Sprintf("%s_TRANSFORMED", str)
		}
		return val
	})

	if val, ok := flds.Get("name"); ok {
		fmt.Println(val)
	}

	// Output:
	// john_TRANSFORMED
}

// ExampleFields_Walk demonstrates iterating over fields.
func ExampleFields_Walk() {
	flds := fields.New(nil)
	flds.Add("field1", "value1")
	flds.Add("field2", "value2")
	flds.Add("field3", "value3")

	count := 0
	flds.Walk(func(key string, val interface{}) bool {
		count++
		return true // Continue iteration
	})

	fmt.Printf("Total fields: %d\n", count)

	// Output:
	// Total fields: 3
}

// ExampleFields_WalkLimit demonstrates filtered iteration.
func ExampleFields_WalkLimit() {
	flds := fields.New(nil)
	flds.Add("request_id", "abc123")
	flds.Add("user_id", 42)
	flds.Add("action", "login")
	flds.Add("timestamp", "2024-01-01")

	// Walk only specific fields
	count := 0
	flds.WalkLimit(func(key string, val interface{}) bool {
		count++
		return true
	}, "request_id", "user_id")

	fmt.Printf("Filtered fields: %d\n", count)

	// Output:
	// Filtered fields: 2
}

// ExampleFields_LoadOrStore demonstrates atomic load-or-store operations.
func ExampleFields_LoadOrStore() {
	flds := fields.New(nil)
	flds.Add("counter", 1)

	// Load existing value
	val, loaded := flds.LoadOrStore("counter", 10)
	fmt.Printf("Loaded: %v, Value: %v\n", loaded, val)

	// Store new value (key doesn't exist)
	val, loaded = flds.LoadOrStore("new_counter", 5)
	fmt.Printf("Loaded: %v, Value: %v\n", loaded, val)

	// Output:
	// Loaded: true, Value: 1
	// Loaded: false, Value: 5
}

// ExampleFields_LoadAndDelete demonstrates atomic load-and-delete operations.
func ExampleFields_LoadAndDelete() {
	flds := fields.New(nil)
	flds.Add("temp", "temporary_value")

	// Load and delete existing field
	val, deleted := flds.LoadAndDelete("temp")
	fmt.Printf("Deleted: %v, Value: %v\n", deleted, val)

	// Try to delete non-existent field
	_, deleted = flds.LoadAndDelete("temp")
	fmt.Printf("Second delete: %v\n", deleted)

	// Output:
	// Deleted: true, Value: temporary_value
	// Second delete: false
}

// ExampleFields_Logrus demonstrates integration with logrus logger.
func ExampleFields_Logrus() {
	// Create fields
	flds := fields.New(nil)
	flds.Add("request_id", "req-123")
	flds.Add("user_id", 456)

	// Convert to logrus.Fields
	logrusFields := flds.Logrus()

	// Use with logrus (simulated output)
	fmt.Printf("Fields count: %d\n", len(logrusFields))
	fmt.Printf("Type: %T\n", logrusFields)

	// Output:
	// Fields count: 2
	// Type: logrus.Fields
}

// ExampleFields_MarshalJSON demonstrates JSON serialization.
func ExampleFields_MarshalJSON() {
	flds := fields.New(nil)
	flds.Add("name", "service-api")
	flds.Add("version", "1.2.3")
	flds.Add("active", true)

	// Marshal to JSON
	data, err := json.Marshal(flds)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(data))

	// Output:
	// {"active":true,"name":"service-api","version":"1.2.3"}
}

// ExampleFields_UnmarshalJSON demonstrates JSON deserialization.
func ExampleFields_UnmarshalJSON() {
	jsonData := `{"service":"api","port":8080,"enabled":true}`

	flds := fields.New(nil)
	err := json.Unmarshal([]byte(jsonData), flds)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(len(flds.Logrus()))

	// Output:
	// 3
}

// ExampleFields_context demonstrates context integration.
// This shows how Fields implements context.Context interface.
func ExampleFields_context() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flds := fields.New(ctx)
	flds.Add("trace_id", "xyz789")

	// Fields implements context.Context
	select {
	case <-flds.Done():
		fmt.Println("Context cancelled")
	default:
		fmt.Println("Context active")
	}

	// Output:
	// Context active
}

// ExampleFields_structuredLogging demonstrates a complete structured logging workflow.
// This is a comprehensive example combining multiple features.
func ExampleFields_structuredLogging() {
	// Initialize base fields for the service
	baseFields := fields.New(context.Background())
	baseFields.Add("service", "user-api")
	baseFields.Add("version", "2.1.0")
	baseFields.Add("environment", "production")

	// Create request-specific fields by cloning
	requestFields := baseFields.Clone()
	requestFields.Add("request_id", "req-abc-123")
	requestFields.Add("method", "POST")
	requestFields.Add("path", "/api/users")

	// Transform sensitive data
	requestFields.Map(func(key string, val interface{}) interface{} {
		if key == "password" {
			return "[REDACTED]"
		}
		return val
	})

	// Convert to logrus format for logging
	logrusFields := requestFields.Logrus()

	fmt.Printf("Total fields: %d\n", len(logrusFields))
	fmt.Printf("Has service: %v\n", logrusFields["service"] != nil)
	fmt.Printf("Has request_id: %v\n", logrusFields["request_id"] != nil)

	// Output:
	// Total fields: 6
	// Has service: true
	// Has request_id: true
}

// ExampleFields_multiSourceAggregation demonstrates combining fields from multiple sources.
// This shows a real-world scenario of aggregating metadata from different components.
func ExampleFields_multiSourceAggregation() {
	// System-level fields
	sysFields := fields.New(nil)
	sysFields.Add("hostname", "server-01")
	sysFields.Add("pid", 12345)

	// Application-level fields
	appFields := fields.New(nil)
	appFields.Add("app_name", "auth-service")
	appFields.Add("app_version", "3.0.0")

	// Request-level fields
	reqFields := fields.New(nil)
	reqFields.Add("request_id", "req-xyz")
	reqFields.Add("user_agent", "Mozilla/5.0")

	// Merge all sources
	combined := sysFields.Clone()
	combined.Merge(appFields)
	combined.Merge(reqFields)

	fmt.Printf("Combined fields: %d\n", len(combined.Logrus()))

	// Output:
	// Combined fields: 6
}

// ExampleFields_clean demonstrates clearing all fields.
func ExampleFields_Clean() {
	flds := fields.New(nil)
	flds.Add("field1", "value1")
	flds.Add("field2", "value2")
	flds.Add("field3", "value3")

	fmt.Printf("Before clean: %d\n", len(flds.Logrus()))

	// Clear all fields
	flds.Clean()

	fmt.Printf("After clean: %d\n", len(flds.Logrus()))

	// Output:
	// Before clean: 3
	// After clean: 0
}

// ExampleFields_complexTypes demonstrates handling complex data types.
func ExampleFields_complexTypes() {
	flds := fields.New(nil)

	// Add various types
	flds.Add("string", "text")
	flds.Add("int", 42)
	flds.Add("float", 3.14)
	flds.Add("bool", true)
	flds.Add("slice", []int{1, 2, 3})
	flds.Add("map", map[string]string{"key": "value"})

	fmt.Printf("Total fields: %d\n", len(flds.Logrus()))

	// Output:
	// Total fields: 6
}

// Example demonstrates a typical usage pattern combining multiple operations.
func Example() {
	// Create base fields
	flds := fields.New(context.Background())

	// Add fields with chaining
	flds.Add("service", "api").
		Add("version", "1.0").
		Add("env", "prod")

	// Clone for specific request
	reqFields := flds.Clone()
	reqFields.Add("request_id", "12345")

	// Get logrus-compatible fields
	logFields := reqFields.Logrus()

	fmt.Println(len(logFields))

	// Output:
	// 4
}
