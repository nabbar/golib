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

// Package info provides a thread-safe, caching implementation for monitor information.
//
// The package offers a flexible way to manage monitor metadata with support for
// dynamic name and info retrieval through registered functions. It implements
// both encoding.TextMarshaler and json.Marshaler interfaces for easy serialization.
//
// # Key Features
//
//   - Thread-safe concurrent access using sync.RWMutex and sync.Map
//   - Lazy evaluation and caching of dynamic data
//   - Support for custom name and info retrieval functions
//   - Built-in text and JSON marshaling
//   - Efficient cache invalidation on re-registration
//
// # Basic Usage
//
// Creating a simple info instance:
//
//	info, err := info.New("my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(info.Name()) // Output: my-service
//
// # Dynamic Name
//
// Register a function to provide dynamic names:
//
//	info.RegisterName(func() (string, error) {
//	    hostname, err := os.Hostname()
//	    if err != nil {
//	        return "", err
//	    }
//	    return fmt.Sprintf("service-%s", hostname), nil
//	})
//	name := info.Name() // Returns "service-<hostname>"
//
// # Dynamic Info
//
// Register a function to provide dynamic metadata:
//
//	info.RegisterInfo(func() (map[string]interface{}, error) {
//	    var m runtime.MemStats
//	    runtime.ReadMemStats(&m)
//	    return map[string]interface{}{
//	        "version":    "1.0.0",
//	        "goroutines": runtime.NumGoroutine(),
//	        "alloc_mb":   m.Alloc / 1024 / 1024,
//	    }, nil
//	})
//	data := info.Info() // Returns current runtime info
//
// # Serialization
//
// The Info type implements standard Go marshaling interfaces:
//
//	// Text marshaling
//	text, err := info.MarshalText()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(text)) // Output: my-service (version: 1.0.0, ...)
//
//	// JSON marshaling
//	jsonData, err := json.Marshal(info)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(jsonData)) // Output: {"Name":"my-service","Info":{...}}
//
// # Caching Behavior
//
// Functions are called only once and results are cached:
//
//	callCount := 0
//	info.RegisterName(func() (string, error) {
//	    callCount++
//	    return fmt.Sprintf("name-%d", callCount), nil
//	})
//	name1 := info.Name() // callCount = 1, returns "name-1"
//	name2 := info.Name() // callCount = 1, returns "name-1" (cached)
//
// Re-registration invalidates the cache:
//
//	info.RegisterName(func() (string, error) {
//	    return "new-name", nil
//	})
//	name3 := info.Name() // Cache cleared, returns "new-name"
//
// # Error Handling
//
// If a registered function returns an error, the default name or nil is returned:
//
//	info.RegisterName(func() (string, error) {
//	    return "", errors.New("failed to get name")
//	})
//	name := info.Name() // Returns default name (not the error value)
//
// # Thread Safety
//
// All methods are thread-safe and can be called concurrently:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 100; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        _ = info.Name()
//	        _ = info.Info()
//	    }()
//	}
//	wg.Wait()
//
// # Performance
//
// The implementation is optimized for read-heavy workloads:
//   - Cached name reads: ~4 ns/op with 0 allocations
//   - Cached info reads: ~140 ns/op with 2 allocations
//   - Name registration: ~37 ns/op with 0 allocations
//   - Info registration: ~32 ns/op with 0 allocations
//
// # Integration
//
// The package integrates seamlessly with the golib monitor system.
// See github.com/nabbar/golib/monitor for complete monitor functionality.
package info
