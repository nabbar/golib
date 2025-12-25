# Archive Compress

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-97.7%25-brightgreen)](TESTING.md)

Unified compression and decompression utilities for multiple algorithms with automatic format detection, encoding/decoding support, and transparent Reader/Writer wrapping.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Supported Algorithms](#supported-algorithms)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Compression](#basic-compression)
  - [Basic Decompression](#basic-decompression)
  - [Automatic Detection](#automatic-detection)
  - [Format Parsing](#format-parsing)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Algorithm Type](#algorithm-type)
  - [Core Functions](#core-functions)
  - [Encoding Support](#encoding-support)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **compress** package provides a simple, consistent interface for working with various compression formats including Gzip, Bzip2, LZ4, and XZ. It offers automatic format detection, encoding/decoding support (JSON, text marshaling), and transparent Reader/Writer wrapping for seamless integration with Go's standard io interfaces.

### Design Philosophy

1. **Algorithm Agnostic**: Single interface for multiple compression formats (Gzip, Bzip2, LZ4, XZ)
2. **Auto-Detection**: Automatic compression format detection from data headers
3. **Standard Compliance**: Implements encoding.TextMarshaler/Unmarshaler and json.Marshaler/Unmarshaler
4. **Zero-Copy Wrapping**: Efficient Reader/Writer wrapping without data buffering
5. **Type Safety**: Enum-based algorithm selection prevents invalid format strings

### Key Features

- ✅ **Unified Algorithm Enumeration**: 5 supported formats (None, Gzip, Bzip2, LZ4, XZ)
- ✅ **Automatic Detection**: Magic number analysis for format identification
- ✅ **Reader/Writer Factories**: Transparent compression/decompression wrappers
- ✅ **JSON/Text Marshaling**: Configuration serialization support
- ✅ **Header Validation**: Format verification via magic numbers
- ✅ **97.7% Test Coverage**: 165 comprehensive test specs with race detection

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────┐
│              Algorithm (enum type)                  │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────┐    ┌──────────────────────────┐   │
│  │   Format     │    │   Detection & Parsing    │   │
│  │              │    │                          │   │
│  │ • String()   │    │ • Parse(string)          │   │
│  │ • Extension()│    │ • Detect(io.Reader)      │   │
│  │ • IsNone()   │    │ • DetectOnly(io.Reader)  │   │
│  └──────────────┘    │ • DetectHeader([]byte)   │   │
│                      └──────────────────────────┘   │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │         I/O Wrapping                         │   │
│  │                                              │   │
│  │ • Reader(io.Reader) → io.ReadCloser          │   │
│  │ • Writer(io.WriteCloser) → io.WriteCloser    │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │         Encoding/Marshaling                  │   │
│  │                                              │   │
│  │ • MarshalText() / UnmarshalText()            │   │
│  │ • MarshalJSON() / UnmarshalJSON()            │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
└─────────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│          Standard Library & External                │
│                                                     │
│  compress/gzip  compress/bzip2  lz4  xz             │
└─────────────────────────────────────────────────────┘
```

| Component | Memory | Complexity | Thread-Safe |
|-----------|--------|------------|-------------|
| **Algorithm** | O(1) | Simple | ✅ Stateless |
| **Parse/Detect** | O(1) | Header scan | ✅ Stateless |
| **Reader/Writer** | O(1) | Delegation | ✅ Per instance |

### Data Flow

```
User Input → Parse/Detect → Algorithm Selection
                │
                ▼
        Reader/Writer Wrapping
                │
                ▼
    Stdlib/External Compression Library
                │
                ▼
        Compressed/Decompressed Output
```

### Supported Algorithms

The package supports five compression algorithms:

| Algorithm | Extension | Compression | Speed | Use Case |
|-----------|-----------|-------------|-------|----------|
| **None** | (none) | 0% | Instant | Pass-through, testing |
| **LZ4** | .lz4 | 20-30% | ~500 MB/s | Real-time, logging |
| **Gzip** | .gz | 30-50% | ~100 MB/s | Web content, general |
| **Bzip2** | .bz2 | 40-60% | ~10 MB/s | Archival, cold storage |
| **XZ** | .xz | 50-70% | ~5 MB/s | Distribution, max compression |

**Magic Numbers (Header Detection):**

```
Gzip:   0x1F 0x8B
Bzip2:  'B' 'Z' 'h' [0-9]
LZ4:    0x04 0x22 0x4D 0x18
XZ:     0xFD 0x37 0x7A 0x58 0x5A 0x00
```

---

## Performance

### Benchmarks

Based on actual benchmark results (AMD64, Go 1.25):

| Operation | Data Size | Median | Mean | Max |
|-----------|-----------|--------|------|-----|
| **Gzip Compress (1KB)** | 1KB | <1µs | <1µs | 300µs |
| **Gzip Decompress (1KB)** | 1KB | <1µs | <1µs | 300µs |
| **Bzip2 Compress (1KB)** | 1KB | <1µs | <1µs | 300µs |
| **LZ4 Compress (1KB)** | 1KB | <1µs | <1µs | 300µs |
| **XZ Compress (1KB)** | 1KB | 300µs | 500µs | 700µs |
| **Detection** | 6 bytes | <1µs | <1µs | 100µs |
| **Parse** | String | <1µs | <1µs | 100µs |

**Compression Ratios (1KB test data):**

```
gzip:  94.2%
bzip2: 90.4%
lz4:   93.1%
xz:    89.8%
```

### Memory Usage

```
Base overhead:        Minimal (enum operations)
Detection:            6-byte peek buffer
Reader wrapping:      Depends on algorithm
  - Gzip: ~256KB internal buffer
  - Bzip2: ~64KB internal buffer
  - LZ4: ~64KB internal buffer
  - XZ: Variable (algorithm-dependent)
Writer wrapping:      Depends on algorithm and settings
```

### Scalability

- **Stateless Operations**: All format functions (String, Extension, IsNone) are O(1) and thread-safe
- **Detection**: O(1) header scan, requires only 6 bytes
- **Concurrent Use**: Multiple goroutines can use separate Algorithm instances safely
- **Zero Allocations**: Parse and detection operations have minimal allocation overhead

---

## Use Cases

### 1. File Archiving with Auto-Detection

Extract files regardless of compression format:

```go
func ExtractFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    alg, reader, err := compress.Detect(in)
    if err != nil {
        return err
    }
    defer reader.Close()

    log.Printf("Detected: %s", alg.String())

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, reader)
    return err
}
```

### 2. HTTP Response Compression

Compress HTTP responses based on client capabilities:

```go
func CompressResponse(w http.ResponseWriter, data []byte, format string) error {
    alg := compress.Parse(format)
    if alg == compress.None {
        w.Write(data)
        return nil
    }

    w.Header().Set("Content-Encoding", alg.String())
    
    writer, err := alg.Writer(struct {
        io.Writer
        io.Closer
    }{w, io.NopCloser(nil)})
    if err != nil {
        return err
    }
    defer writer.Close()

    _, err = writer.Write(data)
    return err
}
```

### 3. Log File Rotation with Compression

Compress rotated log files:

```go
func RotateLog(path string, compression compress.Algorithm) error {
    src, err := os.Open(path)
    if err != nil {
        return err
    }
    defer src.Close()

    dstPath := path + compression.Extension()
    dst, err := os.Create(dstPath)
    if err != nil {
        return err
    }
    defer dst.Close()

    writer, err := compression.Writer(dst)
    if err != nil {
        return err
    }
    defer writer.Close()

    _, err = io.Copy(writer, src)
    return err
}
```

### 4. Configuration with Compression Settings

Store compression preferences in config files:

```go
type AppConfig struct {
    DataCompression compress.Algorithm `json:"data_compression"`
    LogCompression  compress.Algorithm `json:"log_compression"`
}

// Save config
cfg := AppConfig{
    DataCompression: compress.LZ4,
    LogCompression:  compress.Gzip,
}
data, _ := json.Marshal(cfg)
os.WriteFile("config.json", data, 0644)

// Load config
data, _ = os.ReadFile("config.json")
var loaded AppConfig
json.Unmarshal(data, &loaded)
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/compress
```

### Basic Compression

```go
package main

import (
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/compress"
)

func main() {
    file, err := os.Create("output.txt.gz")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Create gzip writer
    writer, err := compress.Gzip.Writer(file)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()

    // Write compressed data
    writer.Write([]byte("This data will be compressed"))
}
```

### Basic Decompression

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/compress"
)

func main() {
    file, err := os.Open("input.txt.gz")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Create gzip reader
    reader, err := compress.Gzip.Reader(file)
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // Read decompressed data
    data, _ := io.ReadAll(reader)
    fmt.Println(string(data))
}
```

### Automatic Detection

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/compress"
)

func main() {
    file, err := os.Open("unknown.dat")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Detect and decompress automatically
    alg, reader, err := compress.Detect(file)
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    fmt.Printf("Detected: %s\n", alg.String())
    data, _ := io.ReadAll(reader)
    fmt.Println(string(data))
}
```

### Format Parsing

```go
package main

import (
    "fmt"
    
    "github.com/nabbar/golib/archive/compress"
)

func main() {
    // Parse from string
    alg := compress.Parse("gzip")
    fmt.Println(alg.String())     // "gzip"
    fmt.Println(alg.Extension())  // ".gz"

    // List all algorithms
    algorithms := compress.List()
    for _, alg := range algorithms {
        fmt.Printf("%s (%s)\n", alg.String(), alg.Extension())
    }
}
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **97.7% code coverage** and **165 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All algorithm operations (String, Extension, IsNone, DetectHeader)
- ✅ Format detection and parsing
- ✅ JSON and text marshaling/unmarshaling
- ✅ Reader/Writer wrapping for all algorithms
- ✅ Round-trip compression/decompression
- ✅ Edge cases and error handling
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Resource Management:**
```go
// ✅ GOOD: Always close resources
writer, err := alg.Writer(file)
if err != nil {
    file.Close()  // Close file if writer creation failed
    return err
}
defer writer.Close()  // Writer.Close() also flushes buffers
```

**Error Handling:**
```go
// ✅ GOOD: Check errors and validate
alg, reader, err := compress.Detect(input)
if err != nil {
    // Fallback to uncompressed
    reader = io.NopCloser(input)
    alg = compress.None
}
defer reader.Close()
```

**Format Validation:**
```go
// ✅ GOOD: Validate parsed format
alg := compress.Parse(userInput)
if alg == compress.None && userInput != "none" {
    return fmt.Errorf("unsupported compression: %s", userInput)
}
```

**Algorithm Selection:**
```go
// ✅ GOOD: Check if compression is needed
if !alg.IsNone() {
    writer, _ := alg.Writer(file)
    defer writer.Close()
    // ... write compressed data
}
```

### ❌ DON'T

**Don't forget to close:**
```go
// ❌ BAD: Writer not closed (data loss)
writer, _ := alg.Writer(file)
writer.Write(data)  // Data may be buffered

// ✅ GOOD: Always close
writer, _ := alg.Writer(file)
defer writer.Close()
writer.Write(data)
```

**Don't assume format:**
```go
// ❌ BAD: Assuming format without detection
reader, _ := compress.Gzip.Reader(file)

// ✅ GOOD: Use automatic detection
alg, reader, _ := compress.Detect(file)
```

**Don't use DetectHeader with truncated data:**
```go
// ❌ BAD: Truncated header (returns false, not error)
data := []byte{0x1F}
if compress.Gzip.DetectHeader(data) {  // Always false
    // Never executed
}

// ✅ GOOD: Ensure sufficient data
if len(data) >= 6 && compress.Gzip.DetectHeader(data) {
    // Now safe
}
```

**Don't parse untrusted input without validation:**
```go
// ❌ BAD: No validation
alg := compress.Parse(untrustedInput)
writer, _ := alg.Writer(file)

// ✅ GOOD: Validate result
alg := compress.Parse(untrustedInput)
if alg == compress.None && untrustedInput != "none" {
    return errors.New("invalid compression format")
}
```

---

## API Reference

### Algorithm Type

```go
type Algorithm uint8

const (
    None  Algorithm = iota  // No compression
    Bzip2                   // Bzip2 compression
    Gzip                    // Gzip compression
    LZ4                     // LZ4 compression
    XZ                      // XZ compression
)
```

**Methods:**
- `String() string` - Get lowercase string representation
- `Extension() string` - Get file extension (e.g., ".gz")
- `IsNone() bool` - Check if algorithm is None
- `DetectHeader([]byte) bool` - Validate magic number
- `Reader(io.Reader) (io.ReadCloser, error)` - Create decompression reader
- `Writer(io.WriteCloser) (io.WriteCloser, error)` - Create compression writer
- `MarshalText() ([]byte, error)` - Text marshaling
- `UnmarshalText([]byte) error` - Text unmarshaling
- `MarshalJSON() ([]byte, error)` - JSON marshaling
- `UnmarshalJSON([]byte) error` - JSON unmarshaling

### Core Functions

**Parse:**
```go
func Parse(s string) Algorithm
```
Parse string to Algorithm (case-insensitive). Returns None if unknown.

**Detect:**
```go
func Detect(r io.Reader) (Algorithm, io.ReadCloser, error)
```
Auto-detect format and return decompression reader.

**DetectOnly:**
```go
func DetectOnly(r io.Reader) (Algorithm, io.ReadCloser, error)
```
Detect format without creating decompression reader.

**List:**
```go
func List() []Algorithm
```
Return all supported algorithms.

**ListString:**
```go
func ListString() []string
```
Return string names of all algorithms.

### Encoding Support

The package implements standard encoding interfaces:

- `encoding.TextMarshaler` / `encoding.TextUnmarshaler`
- `json.Marshaler` / `json.Unmarshaler`

**JSON Marshaling:**
```go
cfg := Config{Compression: compress.Gzip}
json, _ := json.Marshal(cfg)  // {"compression":"gzip"}

// None is marshaled as null
cfg := Config{Compression: compress.None}
json, _ := json.Marshal(cfg)  // {"compression":null}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >80%)
   - Pass all tests including race detector
   - Use `gofmt` and `golint`

2. **AI Usage Policy**
   - ❌ **AI must NEVER be used** to generate package code or core functionality
   - ✅ **AI assistance is limited to**:
     - Testing (writing and improving tests)
     - Debugging (troubleshooting and bug resolution)
     - Documentation (comments, README, TESTING.md)
   - All AI-assisted work must be reviewed and validated by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega for test framework
   - Use `gmeasure` for performance benchmarks
   - Ensure zero race conditions with `go test -race`

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md and TESTING.md if needed

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description of changes

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **97.7% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** stateless operations
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Compression Levels**: Support for configurable compression levels (currently uses defaults)
2. **Custom Parameters**: Support for algorithm-specific compression parameters
3. **Streaming API**: Additional helpers for streaming large files
4. **Multi-Format Archives**: Integration with tar/zip for complete archive solutions
5. **Progress Callbacks**: Optional progress reporting for long operations

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/compress)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, supported algorithms, magic numbers, performance characteristics, use cases, and implementation details. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 97.7% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and test execution examples.

### Related golib Packages

- **[github.com/nabbar/golib/archive/tar](https://pkg.go.dev/github.com/nabbar/golib/archive/tar)** - Tar archive format support that works with compress for compressed tar files (.tar.gz, .tar.bz2, etc.).

- **[github.com/nabbar/golib/archive/zip](https://pkg.go.dev/github.com/nabbar/golib/archive/zip)** - Zip archive format support with optional compression integration.

### External Dependencies

- **[compress/gzip](https://pkg.go.dev/compress/gzip)** - Standard library Gzip support. Used for Gzip compression and decompression.

- **[compress/bzip2](https://pkg.go.dev/compress/bzip2)** - Standard library Bzip2 decompression support (read-only).

- **[github.com/dsnet/compress](https://pkg.go.dev/github.com/dsnet/compress)** - Third-party Bzip2 compression support (write operations).

- **[github.com/pierrec/lz4](https://pkg.go.dev/github.com/pierrec/lz4/v4)** - Pure Go LZ4 compression and decompression implementation.

- **[github.com/ulikunitz/xz](https://pkg.go.dev/github.com/ulikunitz/xz)** - Pure Go XZ compression and decompression implementation.

### Standard Library References

- **[io](https://pkg.go.dev/io)** - Standard I/O interfaces implemented by compress. The package fully implements io.ReadCloser and io.WriteCloser for seamless integration with Go's I/O ecosystem.

- **[encoding](https://pkg.go.dev/encoding)** - Standard encoding interfaces. The package implements TextMarshaler/Unmarshaler for text-based serialization.

- **[encoding/json](https://pkg.go.dev/encoding/json)** - JSON marshaling support. The package implements json.Marshaler/Unmarshaler for JSON serialization.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The compress package follows these conventions for idiomatic Go code.

- **[RFC 1952 (Gzip)](https://www.rfc-editor.org/rfc/rfc1952)** - Gzip file format specification. Understanding this helps with debugging Gzip-specific issues.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/archive/compress`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
