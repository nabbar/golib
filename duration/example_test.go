/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package duration_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nabbar/golib/duration"
)

// ExampleParse demonstrates how to parse a duration string, including days.
func ExampleParse() {
	// Parsing a string with days, hours, minutes, and seconds.
	d, err := duration.Parse("3d6h15m30s")
	if err != nil {
		log.Fatalf("Failed to parse duration: %v", err)
	}
	fmt.Println(d)
	// Output: 3d6h15m30s
}

// ExampleDays demonstrates creating a duration using the Days helper function.
func ExampleDays() {
	// Create a duration of 2 days.
	d := duration.Days(2)
	fmt.Println(d)
	// Output: 2d
}

// ExampleHours demonstrates creating a duration using the Hours helper function.
func ExampleHours() {
	// Create a duration of 36 hours.
	d := duration.Hours(36)
	fmt.Println(d)
	// Output: 1d12h0m0s
}

// ExampleDuration_Time shows how to convert a custom Duration back to a standard time.Duration.
func ExampleDuration_Time() {
	d := duration.Days(1) + duration.Hours(12)
	stdDur := d.Time()
	fmt.Printf("Standard time.Duration: %v", stdDur)
	// Output: Standard time.Duration: 36h0m0s
}

// ExampleDuration_TruncateDays demonstrates truncating a duration to the nearest whole day.
func ExampleDuration_TruncateDays() {
	d, _ := duration.Parse("3d18h")
	trunc := d.TruncateDays()
	fmt.Println(trunc)
	// Output: 3d
}

// ExampleDuration_IsMinutes checks if a duration is at least one minute long.
func ExampleDuration_IsMinutes() {
	d1 := duration.Seconds(90)
	d2 := duration.Seconds(30)
	fmt.Printf("90s is at least a minute: %t\n", d1.IsMinutes())
	fmt.Printf("30s is at least a minute: %t\n", d2.IsMinutes())
	// Output:
	// 90s is at least a minute: true
	// 30s is at least a minute: false
}

// Example_configuration demonstrates how to use duration in a configuration struct with JSON.
func Example_configuration() {
	// Imagine this JSON is from a config file.
	configJSON := `{"timeout": "2d12h"}`

	var config struct {
		Timeout duration.Duration `json:"timeout"`
	}

	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	fmt.Printf("Timeout is: %s\n", config.Timeout)
	fmt.Printf("In hours: %d hours", config.Timeout.Hours())
	// Output:
	// Timeout is: 2d12h0m0s
	// In hours: 60 hours
}

// Example_retryStrategy demonstrates generating a series of durations for a retry mechanism.
func Example_retryStrategy() {
	start := duration.Seconds(1)
	end := duration.Minutes(1)

	// Generate a backoff sequence.
	// Note: The PID controller parameters here are for demonstration.
	backoffSequence := start.RangeTo(end, 0.5, 0.1, 0.1)

	fmt.Println("Retry intervals:")
	for i, d := range backoffSequence {
		fmt.Printf("Attempt %d: wait %s\n", i+1, d.String())
	}
}
