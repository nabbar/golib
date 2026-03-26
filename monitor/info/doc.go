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

// Package info provides a flexible and thread-safe way to manage and expose
// identifying information for a component or service.
//
// It allows for static data to be set manually and dynamic data to be provided
// via registered functions that are executed on-demand. This package is designed
// for scenarios like service discovery, health checks, and exposing runtime
// metadata where information needs to be retrieved dynamically.
//
// # Key Features
//
//   - Thread-safe operations using a lock-free atomic map.
//   - Dynamic name and info retrieval via registered functions.
//   - Manual override and modification of info data.
//   - Built-in text and JSON marshaling for easy serialization.
//
// # Architecture & Dataflow
//
// The Info component follows a layered approach to data retrieval. When a method
// like Name() or Info() is called, it checks for different sources of data in a
// specific order of priority.
//
// ## Data Retrieval Flow
//
// The data retrieval logic can be visualized as follows:
//
//	+------------------+
//	|   Call Name()    |
//	+------------------+
//	         |
//	         v
//	+------------------+      YES      +----------------------+
//	| Has Registered   |------------->|  Execute Function &  |
//	| Name Function?   |              |    Return Result     |
//	+------------------+              +----------------------+
//	         | NO
//	         v
//	+------------------+      YES      +----------------------+
//	| Has Manually Set |------------->|   Return Set Name    |
//	|      Name?       |              |                      |
//	+------------------+              +----------------------+
//	         | NO
//	         v
//	+------------------+
//	| Return Default   |
//	|      Name        |
//	+------------------+
//
// The Info() method follows a similar pattern but merges data from the
// registered function with any manually set data.
//
// # Quick Start
//
// Here is a simple example of how to create and use an Info instance.
//
//	// 1. Create a new Info instance with a default name.
//	inf, err := info.New("my-awesome-service")
//	if err != nil {
//	    log.Fatalf("Failed to create info instance: %v", err)
//	}
//
//	// 2. Register a function to provide dynamic runtime information.
//	inf.RegisterData(func() (map[string]interface{}, error) {
//	    return map[string]interface{}{
//	        "version":    "1.2.3",
//	        "go_version": runtime.Version(),
//	        "goroutines": runtime.NumGoroutine(),
//	    }, nil
//	})
//
//	// 3. Manually add a static piece of information.
//	inf.AddData("region", "us-east-1")
//
//	// 4. Marshal the Info object to JSON for an API response.
//	jsonData, err := json.Marshal(inf)
//	if err != nil {
//	    log.Fatalf("Failed to marshal info to JSON: %v", err)
//	}
//
//	// The output will be a JSON object containing the name and merged info data.
//	fmt.Println(string(jsonData))
//	// Output: {"name":"my-awesome-service","data":{"go_version":"go1.19.5","goroutines":1,"region":"us-east-1","version":"1.2.3"}}
//
// # Usage Patterns
//
// ## Dynamic vs. Static Data
//
// You can combine dynamic and static data sources. The Info() method merges
// them, with the dynamic function's data taking precedence in case of key collisions.
//
//	// Set a static version.
//	inf.SetData(map[string]interface{}{"version": "1.0.0-static"})
//
//	// Register a dynamic function that also provides a version.
//	inf.RegisterData(func() (map[string]interface{}, error) {
//	    return map[string]interface{}{"version": "2.0.0-dynamic"}, nil
//	})
//
//	// The dynamic version will overwrite the static one.
//	data := inf.Data()
//	fmt.Println(data["version"]) // Output: 2.0.0-dynamic
//
// ## Unregistering Functions
//
// To stop a dynamic function from being called, simply register `nil`.
//
//	// Unregister the info function.
//	inf.RegisterData(nil)
//
//	// Now, Data() will only return manually set data.
//	data = inf.Data()
//	fmt.Println(data["version"]) // Output: 1.0.0-static
//
// # Thread Safety
//
// All methods on the Info type are designed to be safe for concurrent use
// from multiple goroutines. The internal state is managed by a `libatm.Map`,
// which is a wrapper around Go's native `sync.Map`.
//
//	var wg sync.WaitGroup
//	for i := 0; i < 100; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        _ = inf.Name()
//	        _ = inf.Data()
//	    }()
//	}
//	wg.Wait() // This will complete without race conditions.
//
// # Important Note on Caching
//
// The current implementation does **not** cache the results of registered functions.
// The registered `FuncInfoName` and `FuncInfoData` functions are executed
// **every time** `Name()` or `Data()` is called, respectively. This ensures that
// the returned information is always up-to-date. If the data retrieval process
// is expensive, the registered function should implement its own caching mechanism.
package info
