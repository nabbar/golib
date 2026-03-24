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
 */

// Package duration provides an extended time.Duration type that supports parsing and formatting with "days" (d) as a unit.
// It also integrates seamlessly with various encoding formats (JSON, YAML, TOML, CBOR) and offers advanced features like PID-controlled range generation and precise truncation.
//
// # Overview
//
// The standard Go time.Duration is limited to hours as its largest parsed unit. This package solves that by introducing a `Duration` type that behaves like time.Duration but adds support for days in string representations (e.g., "5d12h").
// It also provides a robust set of helpers for conversion, comparison, and arithmetic operations.
//
// # Features
//
// - **Extended Parsing:** Support for "d" (days) in duration strings (e.g., "1d2h30m").
// - **Serialization:** Built-in support for Marshaling/Unmarshaling to/from JSON, YAML, TOML, CBOR, and Text.
// - **Helper Functions:** Easy constructors like Days(n), Hours(n), Minutes(n), etc.
// - **Truncation:** methods to truncate durations to days, hours, minutes, seconds, etc.
// - **Range Generation:** Generate sequences of durations using a PID controller for smooth transitions (useful for backoff strategies).
// - **Configuration Integration:** Includes a hook for Viper/Mapstructure decoding.
// - **Big Duration Support:** A sub-package `big` handles durations exceeding the standard int64 limit (±290 years).
//
// Architecture & Data Flow
//
// The core of the package is the `Duration` type, which is a simple type alias for `int64` (same as `time.Duration`).
// This ensures zero-cost conversion to standard `time.Duration` while allowing method attachment.
//
//	+----------------+       +---------------------+      +-------------------+
//	| Input String   | ----> | Parse() / Unmarshal | ---> | Duration (int64)  |
//	| "1d2h"         |       +---------------------+      |                   |
//	+----------------+                                    +---------+---------+
//	                                                                |
//	                                                                v
//	+----------------+       +---------------------+      +---------+---------+
//	| Output String  | <---- | String() / Marshal  | <----| Operations        |
//	| "1d2h0m0s"     |       +---------------------+      | - Truncate        |
//	+----------------+                                    | - Range           |
//	                                                      | - Convert         |
//	                                                      +-------------------+
//
// Quick Start
//
//	package main
//
//	import (
//	    "fmt"
//	    "time"
//	    "github.com/nabbar/golib/duration"
//	)
//
//	func main() {
//	    // 1. Parsing a duration string with days
//	    d, err := duration.Parse("2d4h")
//	    if err != nil {
//	        panic(err)
//	    }
//	    fmt.Println("Parsed:", d) // Output: 2d4h0m0s
//
//	    // 2. Creating a duration from integers
//	    d2 := duration.Days(1) + duration.Hours(12)
//	    fmt.Println("Constructed:", d2) // Output: 1d12h0m0s
//
//	    // 3. Converting to standard time.Duration
//	    stdDuration := d.Time()
//	    fmt.Printf("Standard: %v\n", stdDuration) // Output: 52h0m0s
//
//	    // 4. Truncating
//	    fmt.Println("Truncated to days:", d.TruncateDays()) // Output: 2d
//
//	    // 5. Checking units
//	    if d.IsDays() {
//	        fmt.Println("Duration is at least one day")
//	    }
//	}
//
// # Use Cases
//
//  1. Configuration Files:
//     Allow users to specify long timeouts or intervals in a human-readable format in config files (JSON, YAML, etc.).
//     Example: `timeout: "7d"` is clearer than `timeout: "168h"`.
//
//  2. Exponential Backoff / Retries:
//     Use `RangeTo` or `RangeCtxTo` to generate a sequence of durations for retry attempts that increase over time using a PID controller logic.
//
//  3. API Responses:
//     Return duration fields in JSON APIs that are easy for humans to read and for clients to parse.
//
//  4. Scheduling:
//     Calculate precise schedules involving days, where standard `time.Duration` logic might be cumbersome due to the lack of a 'day' unit.
package duration
