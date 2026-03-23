# Info Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-91.0%25-brightgreen)](TESTING.md)

The `info` package provides a flexible, thread-safe mechanism for managing and exposing component metadata. It supports a hybrid approach of static configuration, manual updates, and dynamic on-demand generation of information.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

This package allows applications to expose self-describing information (such as version, status, runtime metrics) in a structured and concurrent-safe manner. It is designed to be the foundational layer for monitoring and service discovery metadata.

### Design Philosophy

1. **Flexibility**: Support both static data (set once) and dynamic data (computed on access).
2. **Concurrency**: Ensure all operations are thread-safe without heavy locking overhead (using atomic maps).
3. **Simplicity**: Provide a clean interface for registering data providers and retrieving results.

### Key Features

- ✅ **Dynamic Retrieval**: Register functions to generate names and info data on the fly.
- ✅ **Manual Control**: Manually set, update, or delete specific data points.
- ✅ **Thread Safety**: Safe for concurrent reads and writes using lock-free atomic structures.
- ✅ **Serialization**: Built-in support for `json.Marshaler` and `encoding.TextMarshaler`.
- ✅ **Fallback Logic**: Robust name resolution strategy (Function > Manual > Default).

### Key Benefits

- **Real-time Accuracy**: Dynamic functions are executed on every call, ensuring data is always fresh.
- **Low Overhead**: Atomic operations minimize contention compared to standard mutexes.
- **Ease of Use**: Simple API for both simple static cases and complex dynamic scenarios.

---

## Architecture

### Package Structure

```
info/                        # Main package
├── doc.go                   # Package documentation and design overview
├── encode.go                # JSON and Text marshaling logic
├── info.go                  # Public API implementation (Name/Info wrappers)
├── info_test.go             # Core unit tests
├── interface.go             # Interface definitions and constructor
├── manual_test.go           # Tests for manual manipulation methods
└── model.go                 # Internal struct and logic implementation
```

### Dataflow

When `Info()` is called, the component aggregates data from multiple sources.

```
+----------------+       +-------------------------+
|  Client Call   | ----> |      Info() Method      |
+----------------+       +------------+------------+
                                      |
                                      v
                         +-------------------------+
                         | 1. Execute Registered   |
                         |    Info Function (if any)|
                         +------------+------------+
                                      |
                                      v
                         +-------------------------+
                         | 2. Merge with Manual    |
                         |    Data (SetData/AddData)|
                         +------------+------------+
                                      |
                                      v
                         +-------------------------+
                         | 3. Return Combined Map  |
                         +-------------------------+
```

### Name Resolution Strategy

When `Name()` is called, the following priority is applied:

1.  **Dynamic Function**: If registered, execute it. If it returns a valid string, use it.
2.  **Manual Override**: If `SetName()` was called, use that value.
3.  **Default Name**: Fallback to the name provided during `New()`.

---

## Performance

The implementation uses `github.com/nabbar/golib/atomic` (based on `sync.Map`) for internal state management.

### Benchmarks (Intel Core i7-4700HQ)

- **Concurrent Reads**: ~19 ns/op (Zero allocations) for Name reads.
- **Dynamic Execution**: ~39 ns/op for dynamic name functions.
- **Data Deletion**: ~27 ns/op for removing keys.
- **Info Generation**: ~500-1000 ns/op depending on complexity and allocations.

### Characteristics

- **Reads**: Highly efficient, especially for manually set data.
- **Dynamic Generation**: The performance depends on the registered function. Since **no caching** is performed by the library, expensive operations in registered functions will impact every read call. Users are encouraged to implement their own caching inside the function if needed.
- **Allocations**: Optimized to minimize allocations during data merging and encoding.

---

## Use Cases

### 1. Service Metadata

Expose static build information alongside dynamic uptime metrics.

```go
inf, _ := info.New("my-service")
inf.SetData(map[string]interface{}{
    "version": "1.0.0",
    "commit":  "a1b2c3d",
})
inf.RegisterInfo(func() (map[string]interface{}, error) {
    return map[string]interface{}{
        "uptime": time.Since(startTime).String(),
    }, nil
})
```

### 2. Health Check Response

Use `Info()` to generate a health status payload.

```go
inf.RegisterInfo(func() (map[string]interface{}, error) {
    dbStatus := checkDB()
    redisStatus := checkRedis()
    return map[string]interface{}{
        "status": "up",
        "components": map[string]string{
            "db":    dbStatus,
            "redis": redisStatus,
        },
    }, nil
})
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/monitor/info
```

### Basic Implementation

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/monitor/info"
)

func main() {
    // Initialize
    i, err := info.New("my-app")
    if err != nil {
        panic(err)
    }

    // Set some static data
    i.AddData("environment", "production")

    // Output name
    fmt.Println(i.Name()) // Output: my-app
}
```

### Dynamic Data

```go
i.RegisterInfo(func() (map[string]interface{}, error) {
    return map[string]interface{}{
        "goroutines": runtime.NumGoroutine(),
    }, nil
})

// Call Info() to execute function and get data
data := i.Info()
```

---

## Best Practices

### ✅ DO

- **Do** use `SetName` to update the component name if the default one changes during lifecycle.
- **Do** handle `nil` returns from `Info()` if no data has been populated.
- **Do** keep registered functions lightweight to ensure fast `Info()` responses.

### ❌ DON'T

- **Don't** perform heavy blocking operations (like network calls) inside `RegisterInfo` functions without your own caching layer, as this will slow down every call to `Info()`.
- **Don't** assume data persistence if you rely solely on `RegisterInfo`; the function result is transiently merged with stored data.

---

## API Reference

### Interface `Info`

The main interface combining read and write operations.

```go
type Info interface {
    montps.Info
    montps.InfoSet
}
```

#### Read Methods (`montps.Info`)

- `Name() string`: Returns the component name.
- `Info() map[string]interface{}`: Returns the aggregated info map.
- `MarshalJSON() ([]byte, error)`: Serializes to JSON.
- `MarshalText() ([]byte, error)`: Serializes to Text.

#### Write Methods (`montps.InfoSet`)

- `SetName(string)`: Sets/overrides the name.
- `SetData(map[string]interface{})`: Replaces all info data.
- `AddData(string, interface{})`: Adds or updates a specific key.
- `DelData(string)`: Removes a specific key.
- `RegisterName(func() (string, error))`: Registers a dynamic name generator.
- `RegisterInfo(func() (map[string]interface{}, error))`: Registers a dynamic info generator.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
    - Follow Go best practices.
    - Maintain code coverage > 90%.
    - Ensure all tests pass, including race detection.

2. **AI Usage Policy**
    - ❌ **AI must NEVER be used** to generate package code or core functionality.
    - ✅ **AI assistance is limited to**: Testing, Debugging, and Documentation.
    - All AI-assisted work must be reviewed by humans.

3. **Testing**
    - Use Ginkgo v2 / Gomega.
    - Add tests for any new features or bug fixes.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/monitor/info)** - Full API documentation.

### Related golib Packages

- **[github.com/nabbar/golib/monitor](https://github.com/nabbar/golib/tree/master/monitor)** - Parent monitor package.
- **[github.com/nabbar/golib/atomic](https://github.com/nabbar/golib/tree/master/atomic)** - Underlying atomic map implementation.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL
