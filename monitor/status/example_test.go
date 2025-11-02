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

package status_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nabbar/golib/monitor/status"
)

// ExampleStatus demonstrates basic Status usage.
func ExampleStatus() {
	s := status.OK
	fmt.Println(s.String())
	fmt.Println(s.Int())
	// Output:
	// OK
	// 2
}

// ExampleParse demonstrates string parsing.
func ExampleParse() {
	s1 := status.Parse("OK")
	s2 := status.Parse("warn")
	s3 := status.Parse("unknown")

	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(s3)
	// Output:
	// OK
	// Warn
	// KO
}

// ExampleParse_withWhitespace demonstrates parsing with whitespace.
func ExampleParse_withWhitespace() {
	s := status.Parse(" OK ")
	fmt.Println(s)
	// Output: OK
}

// ExampleParseInt demonstrates integer parsing.
func ExampleParseInt() {
	s0 := status.ParseInt(0)
	s1 := status.ParseInt(1)
	s2 := status.ParseInt(2)

	fmt.Println(s0)
	fmt.Println(s1)
	fmt.Println(s2)
	// Output:
	// KO
	// Warn
	// OK
}

// ExampleParseFloat64 demonstrates float parsing.
func ExampleParseFloat64() {
	s1 := status.ParseFloat64(1.9)
	s2 := status.ParseFloat64(2.1)

	fmt.Println(s1)
	fmt.Println(s2)
	// Output:
	// Warn
	// OK
}

// ExampleStatus_String demonstrates String method.
func ExampleStatus_String() {
	fmt.Println(status.OK.String())
	fmt.Println(status.Warn.String())
	fmt.Println(status.KO.String())
	// Output:
	// OK
	// Warn
	// KO
}

// ExampleStatus_Code demonstrates Code method.
func ExampleStatus_Code() {
	fmt.Println(status.OK.Code())
	fmt.Println(status.Warn.Code())
	fmt.Println(status.KO.Code())
	// Output:
	// OK
	// WARN
	// KO
}

// ExampleStatus_Int demonstrates numeric conversion.
func ExampleStatus_Int() {
	fmt.Println(status.KO.Int())
	fmt.Println(status.Warn.Int())
	fmt.Println(status.OK.Int())
	// Output:
	// 0
	// 1
	// 2
}

// ExampleStatus_MarshalJSON demonstrates JSON marshaling.
func ExampleStatus_MarshalJSON() {
	s := status.OK
	data, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
	// Output: "OK"
}

// ExampleStatus_UnmarshalJSON demonstrates JSON unmarshaling.
func ExampleStatus_UnmarshalJSON() {
	var s status.Status
	err := json.Unmarshal([]byte(`"Warn"`), &s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output: Warn
}

// Example_jsonStruct demonstrates using Status in a struct.
func Example_jsonStruct() {
	type HealthCheck struct {
		Status  status.Status `json:"status"`
		Message string        `json:"message"`
	}

	hc := HealthCheck{
		Status:  status.OK,
		Message: "All systems operational",
	}

	data, _ := json.Marshal(hc)
	fmt.Println(string(data))
	// Output: {"status":"OK","message":"All systems operational"}
}

// Example_comparison demonstrates status comparison.
func Example_comparison() {
	if status.OK > status.Warn {
		fmt.Println("OK is better than Warn")
	}
	if status.Warn > status.KO {
		fmt.Println("Warn is better than KO")
	}
	// Output:
	// OK is better than Warn
	// Warn is better than KO
}

// Example_roundTrip demonstrates round-trip conversion.
func Example_roundTrip() {
	original := status.Warn

	// String round-trip
	str := original.String()
	parsed := status.Parse(str)
	fmt.Println(parsed == original)

	// JSON round-trip
	jsonData, _ := json.Marshal(original)
	var decoded status.Status
	json.Unmarshal(jsonData, &decoded)
	fmt.Println(decoded == original)

	// Output:
	// true
	// true
}

// Example_aggregation demonstrates status aggregation.
func Example_aggregation() {
	statuses := []status.Status{
		status.OK,
		status.Warn,
		status.OK,
		status.OK,
	}

	// Find worst status
	worst := status.OK
	for _, s := range statuses {
		if s < worst {
			worst = s
		}
	}

	fmt.Println(worst)
	// Output: Warn
}

// Example_defaulting demonstrates default behavior.
func Example_defaulting() {
	// Invalid inputs default to KO
	fmt.Println(status.Parse(""))
	fmt.Println(status.Parse("invalid"))
	fmt.Println(status.ParseInt(-1))
	fmt.Println(status.ParseInt(999))
	// Output:
	// KO
	// KO
	// KO
	// KO
}
