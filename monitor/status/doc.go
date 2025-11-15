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

// Package status provides a robust enumeration type for representing monitor health status.
//
// The Status type is a uint8-based enumeration with three possible values:
//   - KO (0): Represents a failed or critical status
//   - Warn (1): Represents a warning or degraded status
//   - OK (2): Represents a healthy or successful status
//
// # Key Features
//
//   - Type-safe enumeration with compile-time guarantees
//   - Multiple encoding format support (JSON, YAML, TOML, CBOR, Text)
//   - Flexible parsing from various input types (string, int, float)
//   - Zero dependencies on external packages for core functionality
//   - Efficient uint8 storage (1 byte per value)
//
// # Basic Usage
//
// Creating and using status values:
//
//	s := status.OK
//	fmt.Println(s) // Output: OK
//	fmt.Println(s.Int()) // Output: 2
//	fmt.Println(s.Code()) // Output: OK
//
// # Parsing from Strings
//
// The Parse function is case-insensitive and handles various formats:
//
//	status.Parse("OK")       // returns status.OK
//	status.Parse("ok")       // returns status.OK
//	status.Parse(" warn ")   // returns status.Warn
//	status.Parse("'ko'")     // returns status.KO
//	status.Parse("unknown")  // returns status.KO (default)
//
// # Parsing from Numbers
//
// Convert numeric values to status:
//
//	status.ParseInt(2)        // returns status.OK
//	status.ParseFloat64(1.5)  // returns status.Warn (floored to 1)
//	status.ParseUint(0)       // returns status.KO
//
// Invalid numeric values default to KO:
//
//	status.ParseInt(-1)   // returns status.KO
//	status.ParseInt(999)  // returns status.KO
//
// # JSON Encoding
//
// Status implements json.Marshaler and json.Unmarshaler:
//
//	s := status.OK
//	data, _ := json.Marshal(s)
//	fmt.Println(string(data)) // Output: "OK"
//
//	var s2 status.Status
//	json.Unmarshal([]byte(`"Warn"`), &s2)
//	fmt.Println(s2) // Output: Warn
//
// Works seamlessly in structs:
//
//	type HealthCheck struct {
//	    Status status.Status `json:"status"`
//	    Message string `json:"message"`
//	}
//
//	hc := HealthCheck{Status: status.OK, Message: "All systems operational"}
//	data, _ := json.Marshal(hc)
//	// {"status":"OK","message":"All systems operational"}
//
// # Other Encoding Formats
//
// YAML encoding:
//
//	data, _ := yaml.Marshal(status.Warn)
//	// Output: Warn
//
// Text encoding:
//
//	text, _ := status.OK.MarshalText()
//	// Output: []byte("OK")
//
// CBOR encoding:
//
//	data, _ := status.OK.MarshalCBOR()
//	// Binary CBOR representation of "OK"
//
// # Type Conversions
//
// Convert to various numeric types:
//
//	s := status.Warn
//	s.Int()     // returns 1
//	s.Int64()   // returns int64(1)
//	s.Uint()    // returns uint(1)
//	s.Uint64()  // returns uint64(1)
//	s.Float()   // returns float64(1.0)
//
// String representations:
//
//	s := status.Warn
//	s.String()  // returns "Warn"
//	s.Code()    // returns "WARN" (uppercase)
//
// # Comparison and Ordering
//
// Status values can be compared directly:
//
//	if status.OK > status.Warn {
//	    fmt.Println("OK is better than Warn")
//	}
//
// The ordering is: KO (0) < Warn (1) < OK (2)
//
// # Round-trip Conversions
//
// Status values maintain their identity through encoding round-trips:
//
//	original := status.OK
//
//	// String round-trip
//	str := original.String()
//	parsed := status.Parse(str)
//	// parsed == original
//
//	// JSON round-trip
//	jsonData, _ := json.Marshal(original)
//	var decoded status.Status
//	json.Unmarshal(jsonData, &decoded)
//	// decoded == original
//
// # Error Handling
//
// Parse functions never return errors. Instead, they default to KO for invalid inputs:
//
//	status.Parse("")           // returns KO
//	status.ParseInt(-1)        // returns KO
//	status.ParseFloat64(999.0) // returns KO
//
// This design ensures that status values are always valid, preventing nil pointer or
// invalid state errors.
//
// # Use Cases
//
// Health monitoring:
//
//	func checkDatabase() status.Status {
//	    if err := db.Ping(); err != nil {
//	        return status.KO
//	    }
//	    if latency > threshold {
//	        return status.Warn
//	    }
//	    return status.OK
//	}
//
// Aggregation:
//
//	func aggregateStatus(statuses []status.Status) status.Status {
//	    worst := status.OK
//	    for _, s := range statuses {
//	        if s < worst {
//	            worst = s
//	        }
//	    }
//	    return worst
//	}
//
// # Integration
//
// This package is part of the golib monitor system. For complete monitoring functionality:
//   - github.com/nabbar/golib/monitor
//   - github.com/nabbar/golib/monitor/types
//   - github.com/nabbar/golib/monitor/info
//
// # Performance
//
// The Status type is highly optimized:
//   - 1 byte storage (uint8)
//   - Zero-allocation comparisons
//   - Minimal allocation conversions
//   - Fast string parsing (~35 ns/op)
//   - Fast JSON marshaling (~40 ns/op)
package status
