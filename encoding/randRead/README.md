# Random Reader Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/encoding/randRead)

**Buffered random data reader from remote sources with automatic reconnection and error handling.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Use Cases](#use-cases)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **randRead** package provides a buffered io.ReadCloser for reading random data from remote sources. It automatically handles connection management, buffering, and reconnection on errors.

### Design Philosophy

- **Automatic Reconnection**: Transparent remote source reconnection on errors
- **Buffering**: Internal buffer for efficient data reads
- **Thread-Safe**: Atomic operations for concurrent access
- **Simple API**: Standard io.ReadCloser interface
- **Flexible**: Works with any remote source (HTTP, gRPC, etc.)

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Remote Source** | Read from any remote source via function |
| **Auto Reconnect** | Automatic reconnection on connection failure |
| **Buffering** | Internal buffer for read optimization |
| **Thread-Safe** | Atomic operations for concurrent use |
| **Error Handling** | Graceful error recovery |
| **Standard Interface** | Implements io.ReadCloser |

---

## Architecture

### Component Architecture

```
┌─────────────────────────────────────────────────────┐
│              RandRead Package                        │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         Public Interface                     │  │
│  │       (io.ReadCloser)                        │  │
│  │  - Read(p []byte) (n int, err error)        │  │
│  │  - Close() error                             │  │
│  └────────────────┬─────────────────────────────┘  │
│                   │                                  │
│                   ▼                                  │
│  ┌──────────────────────────────────────────────┐  │
│  │         Buffered Reader                      │  │
│  │  - Internal buffer (atomic.Value)            │  │
│  │  - Buffer management                         │  │
│  │  - Read caching                              │  │
│  └────────────────┬─────────────────────────────┘  │
│                   │                                  │
│                   ▼                                  │
│  ┌──────────────────────────────────────────────┐  │
│  │         Remote Manager                       │  │
│  │  - Connection handling (atomic.Value)        │  │
│  │  - Auto reconnect on error                   │  │
│  │  - Source function invocation                │  │
│  └────────────────┬─────────────────────────────┘  │
│                   │                                  │
│                   ▼                                  │
│  ┌──────────────────────────────────────────────┐  │
│  │         Remote Source                        │  │
│  │  (User-provided function)                    │  │
│  │  - HTTP request                              │  │
│  │  - gRPC stream                               │  │
│  │  - Any io.ReadCloser source                  │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Data Flow

```
Application Read Request
         │
         ▼
┌────────────────────┐
│  Check Buffer      │
│  (atomic.Value)    │
└────────┬───────────┘
         │
         ├─── Buffer has data ──→ Return buffered data
         │
         └─── Buffer empty
                │
                ▼
      ┌─────────────────────┐
      │  Remote Connection  │
      │  (atomic.Value)     │
      └──────────┬──────────┘
                 │
                 ├─── Connected ──→ Read from source
                 │
                 └─── Not connected / Error
                        │
                        ▼
                 ┌───────────────┐
                 │  Reconnect    │
                 │  (FuncRemote) │
                 └───────┬───────┘
                         │
                         └──→ Read from new source
```

---

## Installation

```bash
go get github.com/nabbar/golib/encoding/randRead
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    
    "github.com/nabbar/golib/encoding/randRead"
)

func main() {
    // Define remote source function
    source := func() (io.ReadCloser, error) {
        resp, err := http.Get("https://www.random.org/cgi-bin/randbyte?nbytes=1024&format=f")
        if err != nil {
            return nil, err
        }
        return resp.Body, nil
    }
    
    // Create random reader
    reader := randRead.New(source)
    defer reader.Close()
    
    // Read random data
    buffer := make([]byte, 100)
    n, err := reader.Read(buffer)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Read %d random bytes\n", n)
}
```

### With Custom Source

```go
// Custom remote source (e.g., gRPC stream)
source := func() (io.ReadCloser, error) {
    stream, err := grpcClient.GetRandomStream(context.Background())
    if err != nil {
        return nil, err
    }
    return streamWrapper(stream), nil
}

reader := randRead.New(source)
defer reader.Close()

// Use reader like any io.Reader
data := make([]byte, 256)
reader.Read(data)
```

---

## Core Concepts

### Remote Source Function

**Type**: `FuncRemote func() (io.ReadCloser, error)`

- **Purpose**: Provides the remote connection
- **Called**: On initialization and reconnection
- **Returns**: io.ReadCloser (data source) and error
- **Examples**: HTTP response, gRPC stream, file, socket

### Automatic Reconnection

When a read error occurs:
1. Current connection is closed
2. Remote source function is called
3. New connection is established
4. Read operation continues

This is **transparent** to the caller.

### Buffering

- Internal buffer stores read data
- Reduces remote source calls
- Improves read performance
- Managed automatically

### Thread Safety

- Uses `atomic.Value` for buffer and connection
- Safe for concurrent reads
- Lock-free design

---

## API Reference

### New

**`New(fct FuncRemote) io.ReadCloser`**

Creates a new random reader from remote source.

```go
reader := randRead.New(sourceFunction)
```

**Parameters:**
- `fct`: Function returning io.ReadCloser and error

**Returns:**
- io.ReadCloser: Random reader instance
- Returns nil if function is nil

### Read

**`Read(p []byte) (n int, err error)`**

Reads random data into p.

```go
n, err := reader.Read(buffer)
```

**Behavior:**
- Reads from internal buffer first
- Fetches from remote if buffer empty
- Auto-reconnects on errors
- Standard io.Reader semantics

**Returns:**
- n: Number of bytes read
- err: Error if any (io.EOF on close)

### Close

**`Close() error`**

Closes the reader and underlying connection.

```go
err := reader.Close()
```

**Returns:**
- error: Error if any

**Note**: Always defer Close() to prevent leaks

---

## Use Cases

### Cryptographic Random Data

```go
import (
    "crypto/rand"
    "io"
    "github.com/nabbar/golib/encoding/randRead"
)

// Fallback to crypto/rand on HTTP failure
source := func() (io.ReadCloser, error) {
    resp, err := http.Get("https://random-service.example.com/bytes")
    if err != nil {
        // Fallback to local crypto/rand
        return io.NopCloser(rand.Reader), nil
    }
    return resp.Body, nil
}

reader := randRead.New(source)
defer reader.Close()

// Generate random key
key := make([]byte, 32)
io.ReadFull(reader, key)
```

### Random Testing Data

```go
// Generate random test data from remote service
source := func() (io.ReadCloser, error) {
    resp, err := http.Get("https://test-data-gen.example.com/random")
    return resp.Body, err
}

reader := randRead.New(source)
defer reader.Close()

// Generate test data
testData := make([]byte, 1024*1024) // 1MB
io.ReadFull(reader, testData)
```

### Load Testing

```go
// Random data for load testing
source := func() (io.ReadCloser, error) {
    return http.Get("https://random.org/bytes?num=10000")
}

reader := randRead.New(source)
defer reader.Close()

// Send random payloads
for i := 0; i < 1000; i++ {
    payload := make([]byte, 1024)
    reader.Read(payload)
    sendToServer(payload)
}
```

### Streaming Random Data

```go
// Stream random data to output
source := func() (io.ReadCloser, error) {
    conn, err := net.Dial("tcp", "random-server:9000")
    return conn, err
}

reader := randRead.New(source)
defer reader.Close()

// Copy to destination
io.Copy(outputWriter, reader)
```

---

## Performance

### Characteristics

| Aspect | Performance | Notes |
|--------|-------------|-------|
| **Buffering** | ~90% hit rate | Depends on read patterns |
| **Reconnection** | ~100ms overhead | Depends on remote source |
| **Memory** | ~10KB | Buffer + connection state |
| **Concurrency** | Lock-free reads | Atomic operations |

### Performance Tips

**1. Buffer Size**
```go
// Read in chunks matching buffer size
buffer := make([]byte, 4096)  // Typical buffer size
for {
    n, err := reader.Read(buffer)
    // Process data
}
```

**2. Reuse Reader**
```go
// ✅ Good: Reuse reader
reader := randRead.New(source)
defer reader.Close()
for i := 0; i < 1000; i++ {
    reader.Read(data)
}

// ❌ Bad: Create new reader each time
for i := 0; i < 1000; i++ {
    reader := randRead.New(source)
    reader.Read(data)
    reader.Close()
}
```

**3. Handle Errors**
```go
// ✅ Good: Check errors
n, err := reader.Read(data)
if err != nil && err != io.EOF {
    log.Printf("Read error: %v", err)
}
```

---

## Best Practices

### 1. Always Close

```go
// ✅ Good
reader := randRead.New(source)
defer reader.Close()

// ❌ Bad: Resource leak
reader := randRead.New(source)
reader.Read(data)
// Never closed!
```

### 2. Handle Nil Function

```go
// ✅ Good: Check nil
if source == nil {
    return errors.New("source required")
}
reader := randRead.New(source)

// ✅ Also handled: New returns nil
reader := randRead.New(nil)
if reader == nil {
    // Handle nil reader
}
```

### 3. Implement Proper Source Function

```go
// ✅ Good: Handle errors in source
source := func() (io.ReadCloser, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("connection failed: %w", err)
    }
    if resp.StatusCode != 200 {
        resp.Body.Close()
        return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
    }
    return resp.Body, nil
}

// ❌ Bad: No error handling
source := func() (io.ReadCloser, error) {
    resp, _ := http.Get(url)  // Ignoring errors
    return resp.Body, nil
}
```

### 4. Use Context for Cancellation

```go
// ✅ Good: Use context in source function
source := func() (io.ReadCloser, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    return resp.Body, err
}
```

### 5. Verify Data Quality

```go
// ✅ Good: Verify random data quality
reader := randRead.New(source)
defer reader.Close()

data := make([]byte, 1000)
io.ReadFull(reader, data)

// Check for patterns (all zeros, repeating, etc.)
if isLowEntropy(data) {
    log.Warn("Low quality random data detected")
}
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd encoding/randRead
go test -v -cover
```

**Test Metrics:**
- Comprehensive test coverage
- Error scenario testing
- Reconnection testing
- Ginkgo v2 + Gomega framework

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain thread safety guarantees
- Follow existing code style

**Testing**
- Write tests for all new features
- Test error scenarios
- Test reconnection behavior
- Include edge cases

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Document all public APIs with GoDoc
- Keep TESTING.md synchronized

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

**Features**
- Configurable buffer size
- Retry policies (exponential backoff)
- Connection pooling
- Metrics/statistics (bytes read, reconnections)
- Circuit breaker pattern

**Performance**
- Zero-copy operations
- Parallel source fetching
- Adaptive buffering
- Connection keepalive

**Reliability**
- Health checks
- Fallback sources
- Error rate limiting
- Graceful degradation

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Go Standard Library
- **[io](https://pkg.go.dev/io)** - Reader/Writer interfaces
- **[crypto/rand](https://pkg.go.dev/crypto/rand)** - Cryptographic random
- **[net/http](https://pkg.go.dev/net/http)** - HTTP client

### Related Golib Packages
- **[encoding](../README.md)** - Encoding interfaces

### Random Sources
- **[Random.org](https://www.random.org/)** - True random numbers
- **[CloudFlare randomness](https://www.cloudflare.com/learning/ssl/lava-lamp-encryption/)** - Lava lamp randomness

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2024 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/randRead)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
