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

package info_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nabbar/golib/monitor/info"
)

// ExampleNew demonstrates creating a new Info instance.
func ExampleNew() {
	i, err := info.New("my-service")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(i.Name())
	// Output: my-service
}

// ExampleInfo_Name demonstrates retrieving the name.
func ExampleInfo_Name() {
	i, _ := info.New("example-service")
	fmt.Println(i.Name())
	// Output: example-service
}

// ExampleInfo_RegisterName demonstrates registering a dynamic name function.
func ExampleInfo_RegisterName() {
	i, _ := info.New("default-name")

	i.RegisterName(func() (string, error) {
		return "dynamic-name", nil
	})

	fmt.Println(i.Name())
	// Output: dynamic-name
}

// ExampleInfo_RegisterInfo demonstrates registering a dynamic info function.
func ExampleInfo_RegisterInfo() {
	i, _ := info.New("service")

	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
			"status":  "running",
		}, nil
	})

	infoMap := i.Info()
	fmt.Printf("Version: %s\n", infoMap["version"])
	fmt.Printf("Status: %s\n", infoMap["status"])
	// Output:
	// Version: 1.0.0
	// Status: running
}

// ExampleInfo_MarshalJSON demonstrates JSON marshaling.
func ExampleInfo_MarshalJSON() {
	i, _ := info.New("api-service")

	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "2.0.0",
		}, nil
	})

	jsonData, err := json.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonData))
	// Output: {"Name":"api-service","Info":{"version":"2.0.0"}}
}

// ExampleInfo_MarshalText demonstrates text marshaling.
func ExampleInfo_MarshalText() {
	i, _ := info.New("web-service")

	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"port": 8080,
		}, nil
	})

	text, err := i.MarshalText()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(text))
	// Output: web-service (port: 8080)
}

// Example_caching demonstrates the caching behavior.
func Example_caching() {
	i, _ := info.New("service")

	callCount := 0
	i.RegisterName(func() (string, error) {
		callCount++
		return fmt.Sprintf("name-%d", callCount), nil
	})

	// First call executes the function
	name1 := i.Name()
	fmt.Printf("First call: %s (callCount: %d)\n", name1, callCount)

	// Second call uses cached value
	name2 := i.Name()
	fmt.Printf("Second call: %s (callCount: %d)\n", name2, callCount)

	// Output:
	// First call: name-1 (callCount: 1)
	// Second call: name-1 (callCount: 1)
}

// Example_errorHandling demonstrates error handling.
func Example_errorHandling() {
	i, _ := info.New("default-service")

	i.RegisterName(func() (string, error) {
		return "", fmt.Errorf("simulated error")
	})

	// Returns default name on error
	name := i.Name()
	fmt.Println(name)
	// Output: default-service
}

// Example_multipleInfo demonstrates handling multiple info registrations.
func Example_multipleInfo() {
	i, _ := info.New("service")

	// First registration
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "1.0.0",
		}, nil
	})

	fmt.Println("Version:", i.Info()["version"])

	// Re-registration clears cache
	i.RegisterInfo(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			"version": "2.0.0",
		}, nil
	})

	fmt.Println("Version:", i.Info()["version"])
	// Output:
	// Version: 1.0.0
	// Version: 2.0.0
}
