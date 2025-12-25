# Archive Helper

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-84.5%25-brightgreen)](TESTING.md)

Streaming compression and decompression helpers providing unified io.ReadWriteCloser interfaces for transparent data transformation.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Implementation Types](#implementation-types)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Compress While Reading](#compress-while-reading)
  - [Compress While Writing](#compress-while-writing)
  - [Decompress While Reading](#decompress-while-reading)
  - [Decompress While Writing](#decompress-while-writing)
  - [Automatic Type Detection](#automatic-type-detection)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Helper Interface](#helper-interface)
  - [Constructor Functions](#constructor-functions)
  - [Operation Types](#operation-types)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **helper** package provides a simplified, unified interface for compression and decompression operations on streaming data. It acts as an adapter layer between standard Go io interfaces and compression algorithms from `github.com/nabbar/golib/archive/compress`.

### Design Philosophy

1. **Interface Simplicity**: Single `Helper` interface (io.ReadWriteCloser) for all operations
2. **Algorithm Agnostic**: Works with any compression algorithm without algorithm-specific code
3. **Streaming First**: Process data in chunks without loading entire streams into memory
4. **Resource Safety**: Proper cleanup through Close() and automatic wrapper conversions
5. **Type Safety**: Compile-time guarantees through interface-based design

### Key Features

- ✅ **Unified Interface**: Single Helper interface for compression and decompression
- ✅ **Bidirectional Operations**: Compress/decompress on read or write
- ✅ **Algorithm Support**: Works with GZIP, ZLIB, DEFLATE, BZIP2, LZ4, ZSTD, Snappy
- ✅ **Automatic Wrappers**: Auto-converts io.Reader/Writer to io.ReadCloser/WriteCloser
- ✅ **Streaming Architecture**: Minimal memory overhead with chunk-based processing
- ✅ **Thread-Safe**: Safe for single-instance use (one Helper per goroutine)
- ✅ **Zero Dependencies**: Only standard library and golib packages

---

## Architecture

### Component Diagram

```
┌──────────────────────────────────────────────┐
│               Helper Interface               │
│             (io.ReadWriteCloser)             │
└──────────────────────┬───────────────────────┘
                       │
          ┌────────────┴────────────┐
          │                         │
   ┌──────▼──────┐           ┌──────▼──────┐
   │  Compress   │           │  Decompress │
   └──────┬──────┘           └──────┬──────┘
     ┌────┴────┐               ┌────┴────┐
     ▼         ▼               ▼         ▼
 ┌────────┬────────┐       ┌────────┬────────┐
 │ Reader │ Writer │       │ Reader │ Writer │
 └────────┴────────┘       └────────┴────────┘
  compressReader         deCompressReader
  compressWriter         deCompressWriter
```

### Data Flow

**Compression Read Flow** (compressReader):
1. Client calls Read(p) on Helper
2. Helper reads raw data from source reader
3. Data is compressed using algorithm's Writer
4. Compressed data is buffered internally
5. Compressed chunks are returned to client

**Compression Write Flow** (compressWriter):
1. Client calls Write(p) on Helper
2. Data is passed to algorithm's Writer
3. Compressed data is written to destination
4. Close() finalizes compression stream

**Decompression Read Flow** (deCompressReader):
1. Client calls Read(p) on Helper
2. Compressed data is read from source
3. Algorithm's Reader decompresses data
4. Decompressed data is returned to client

**Decompression Write Flow** (deCompressWriter):
1. Client calls Write(p) on Helper
2. Compressed data is buffered
3. Background goroutine reads from buffer
4. Algorithm's Reader decompresses data
5. Decompressed data is written to destination

### Implementation Types

| Type | Direction | Operation | Use Case |
|------|-----------|-----------|----------|
| **compressReader** | Read | Compress | Read file → compress → output |
| **compressWriter** | Write | Compress | Input → compress → write file |
| **deCompressReader** | Read | Decompress | Read compressed → decompress → output |
| **deCompressWriter** | Write | Decompress | Input compressed → decompress → write file |

---

## Performance

### Benchmarks

Based on actual test results from the comprehensive test suite (Go 1.25, AMD64):

| Operation | Median | Mean | Max | Throughput |
|-----------|--------|------|-----|------------|
| **Compress Read (Small)** | 100µs | 300µs | 7.3ms | ~100 ops/sec |
| **Compress Read (Medium)** | 200µs | 300µs | 800µs | ~500 ops/sec |
| **Compress Read (Large)** | 500µs | 600µs | 900µs | ~200 ops/sec |
| **Compress Write (Small)** | 100µs | 200µs | 1ms | ~500 ops/sec |
| **Compress Write (Medium)** | 100µs | 200µs | 500µs | ~500 ops/sec |
| **Compress Write (Large)** | 500µs | 600µs | 1ms | ~200 ops/sec |
| **Decompress Read** | <100µs | <100µs | 100µs | ~1000 ops/sec |
| **Round-trip** | 200µs | 300µs | 1.3ms | ~500 ops/sec |

**Notes:**
- Small: ~100 bytes
- Medium: ~1KB
- Large: ~10KB
- Actual performance depends on data characteristics and chosen algorithm

### Memory Usage

```
Base overhead:        ~100 bytes (struct)
compressReader:       512 bytes (chunkSize buffer)
deCompressWriter:     + 1 goroutine (~2KB stack)
Per operation:        Minimal allocations
```

**Memory Efficiency:**
- O(1) memory usage regardless of input size
- No buffer growth or dynamic allocation
- Streaming operations with constant memory footprint

### Scalability

**Concurrent Operations:**
- Create separate Helper instances per goroutine
- No shared state between instances
- Thread-safe for single-instance usage

**Data Size:**
- Tested with files up to 10MB
- Constant memory regardless of data size
- Streaming architecture enables unlimited data processing

---

## Use Cases

### 1. Compressed File I/O

**Problem**: Read and write compressed files transparently.

```go
h, _ := helper.NewWriter(compress.GZIP, helper.Compress, file)
defer h.Close()
h.Write(data) // Automatically compressed
```

**Real-world**: Log file compression, archive creation, data storage.

### 2. Network Data Compression

**Problem**: Compress data before sending over network.

```go
h, _ := helper.NewWriter(compress.LZ4, helper.Compress, conn)
defer h.Close()
h.Write(payload) // Compressed before transmission
```

**Real-world**: API data transfer, socket communication, bandwidth optimization.

### 3. Transparent Decompression

**Problem**: Read compressed data as if uncompressed.

```go
h, _ := helper.NewReader(compress.GZIP, helper.Decompress, compressedStream)
defer h.Close()
scanner := bufio.NewScanner(h) // Works with any io.Reader consumer
```

**Real-world**: Processing compressed logs, reading archives, data ingestion.

### 4. Format Conversion

**Problem**: Convert between compression formats.

```go
// Read GZIP, write LZ4
src, _ := helper.NewReader(compress.GZIP, helper.Decompress, gzipFile)
dst, _ := helper.NewWriter(compress.LZ4, helper.Compress, lz4File)
defer src.Close()
defer dst.Close()
io.Copy(dst, src)
```

**Real-world**: Migration between formats, compatibility adaptation.

### 5. Streaming Processing

**Problem**: Process compressed data incrementally.

```go
h, _ := helper.NewReader(compress.ZSTD, helper.Decompress, compressedData)
defer h.Close()
processInChunks(h) // Read and process incrementally
```

**Real-world**: Large file processing, ETL pipelines, real-time data transformation.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/helper
```

### Compress While Reading

Read from a source and get compressed data:

```go
package main

import (
    "io"
    "os"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    input, _ := os.Open("input.txt")
    defer input.Close()
    
    output, _ := os.Create("output.txt.gz")
    defer output.Close()
    
    // Create compression reader
    h, _ := helper.NewReader(compress.Gzip, helper.Compress, input)
    defer h.Close()
    
    // Copy compressed data to output
    io.Copy(output, h)
}
```

### Compress While Writing

Write data and have it compressed automatically:

```go
package main

import (
    "os"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    output, _ := os.Create("output.txt.gz")
    defer output.Close()
    
    // Create compression writer
    h, _ := helper.NewWriter(compress.Gzip, helper.Compress, output)
    defer h.Close()
    
    // Write data - it will be compressed automatically
    data := []byte("Hello, World!")
    h.Write(data)
}
```

### Decompress While Reading

Read compressed data and get decompressed output:

```go
package main

import (
    "io"
    "os"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    input, _ := os.Open("compressed.gz")
    defer input.Close()
    
    // Create decompression reader
    h, _ := helper.NewReader(compress.Gzip, helper.Decompress, input)
    defer h.Close()
    
    // Read decompressed data
    data, _ := io.ReadAll(h)
    os.Stdout.Write(data)
}
```

### Decompress While Writing

Write compressed data and have it decompressed:

```go
package main

import (
    "io"
    "os"
    "strings"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    output, _ := os.Create("decompressed.txt")
    defer output.Close()
    
    // First, get some compressed data
    var buf strings.Builder
    cw, _ := helper.NewWriter(compress.Gzip, helper.Compress, &buf)
    cw.Write([]byte("Hello, World!"))
    cw.Close()
    
    // Now decompress while writing
    h, _ := helper.NewWriter(compress.Gzip, helper.Decompress, output)
    defer h.Close()
    
    // Write compressed data - it will be decompressed
    h.Write([]byte(buf.String()))
}
```

### Automatic Type Detection

The `New()` function auto-detects reader vs writer:

```go
// Source is io.Reader - creates a reader helper
input, _ := os.Open("file.txt")
h1, _ := helper.New(compress.Gzip, helper.Compress, input)
data, _ := io.ReadAll(h1)

// Source is io.Writer - creates a writer helper
output, _ := os.Create("file.gz")
h2, _ := helper.New(compress.Gzip, helper.Compress, output)
h2.Write(data)
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **84.5% code coverage** and **76 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and constructor functions
- ✅ Concurrent operations with race detector (zero races detected)
- ✅ Performance benchmarks (compression/decompression throughput)
- ✅ Error handling and edge cases
- ✅ Multiple compression algorithms

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Always Close Resources:**
```go
h, err := helper.NewWriter(algo, helper.Compress, dst)
if err != nil {
    return err
}
defer h.Close() // Ensures compression finalization
```

**Check Error Returns:**
```go
n, err := h.Write(data)
if err != nil {
    return fmt.Errorf("compression failed: %w", err)
}
if n != len(data) {
    return fmt.Errorf("incomplete write: %d of %d bytes", n, len(data))
}
```

**Choose Appropriate Algorithms:**
```go
// Fast compression for temporary data
helper.NewWriter(compress.LZ4, helper.Compress, dst)

// Maximum compression for archival
helper.NewWriter(compress.Zstd, helper.Compress, dst)

// Compatibility with external tools
helper.NewWriter(compress.Gzip, helper.Compress, dst)
```

**Stream Large Data:**
```go
// ✅ GOOD: Streaming, constant memory
h, _ := helper.NewReader(algo, helper.Compress, largeFile)
defer h.Close()
io.Copy(destination, h)
```

### ❌ DON'T

**Don't forget to close:**
```go
// ❌ BAD: Resource leak
h, _ := helper.NewWriter(compress.Gzip, helper.Compress, file)
h.Write(data) // Missing defer h.Close()

// ✅ GOOD: Proper cleanup
h, _ := helper.NewWriter(compress.Gzip, helper.Compress, file)
defer h.Close()
h.Write(data)
```

**Don't share instances across goroutines:**
```go
// ❌ BAD: Race condition
go func() { h.Write(data1) }()
go func() { h.Write(data2) }() // Concurrent writes!

// ✅ GOOD: Separate instances
go func() {
    h1, _ := helper.NewWriter(algo, helper.Compress, dst1)
    defer h1.Close()
    h1.Write(data1)
}()
go func() {
    h2, _ := helper.NewWriter(algo, helper.Compress, dst2)
    defer h2.Close()
    h2.Write(data2)
}()
```

**Don't use wrong operation:**
```go
// ❌ BAD: Read() on writer returns error
h, _ := helper.NewWriter(compress.Gzip, helper.Compress, dst)
h.Read(buf) // Returns ErrInvalidSource

// ✅ GOOD: Use correct operation
h, _ := helper.NewReader(compress.Gzip, helper.Compress, src)
h.Read(buf)
```

---

## API Reference

### Helper Interface

```go
type Helper interface {
    io.ReadWriteCloser
}
```

The Helper interface combines:
- `io.Reader`: Read compressed/decompressed data
- `io.Writer`: Write data for compression/decompression
- `io.Closer`: Release resources and finalize streams

**Note**: Read() and Write() are mutually exclusive based on Helper type. Compression/decompression readers don't support Write(), and writers don't support Read().

### Constructor Functions

#### New

```go
func New(algo arccmp.Algorithm, ope Operation, src any) (Helper, error)
```

Creates a Helper with automatic type detection based on source.

**Parameters:**
- `algo`: Compression algorithm (from github.com/nabbar/golib/archive/compress)
- `ope`: Operation type (Compress or Decompress)
- `src`: Data source (io.Reader or io.Writer)

**Returns:**
- `Helper`: New Helper instance
- `error`: ErrInvalidSource if src is neither io.Reader nor io.Writer

#### NewReader

```go
func NewReader(algo arccmp.Algorithm, ope Operation, src io.Reader) (Helper, error)
```

Creates a Helper for reading data from an io.Reader.

**Parameters:**
- `algo`: Compression algorithm
- `ope`: Compress to compress while reading, Decompress to decompress while reading
- `src`: Source reader

**Returns:**
- `Helper`: Reader Helper instance
- `error`: ErrInvalidOperation for invalid operation

#### NewWriter

```go
func NewWriter(algo arccmp.Algorithm, ope Operation, dst io.Writer) (Helper, error)
```

Creates a Helper for writing data to an io.Writer.

**Parameters:**
- `algo`: Compression algorithm
- `ope`: Compress to compress while writing, Decompress to decompress while writing
- `dst`: Destination writer

**Returns:**
- `Helper`: Writer Helper instance
- `error`: ErrInvalidOperation for invalid operation

### Operation Types

```go
type Operation uint8

const (
    Compress   Operation = iota  // Compress data
    Decompress                   // Decompress data
)
```

### Error Codes

```go
var (
    ErrInvalidSource    = errors.New("invalid source")
    ErrClosedResource   = errors.New("closed resource")
    ErrInvalidOperation = errors.New("invalid operation")
)
```

**Error Conditions:**
- `ErrInvalidSource`: Returned when source is neither io.Reader nor io.Writer, or when calling Read() on a writer/Write() on a reader
- `ErrClosedResource`: Returned when writing to a closed deCompressWriter
- `ErrInvalidOperation`: Returned for unsupported Operation values (not Compress or Decompress)

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

- ✅ **84.5% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** per instance (one Helper per goroutine)
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Configurable Buffer Sizes**: Allow users to specify custom chunk sizes for performance tuning
2. **Progress Callbacks**: Optional progress reporting for long-running operations
3. **Compression Level Control**: Expose algorithm-specific compression levels
4. **Metrics Integration**: Built-in performance metrics collection

These are **optional improvements** and not required for production use. The current implementation is stable and performant for its intended use cases.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/helper)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, data flow, implementation details, and best practices. Provides comprehensive explanations of internal mechanisms and compression algorithm integration.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 84.5% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/archive/compress](https://pkg.go.dev/github.com/nabbar/golib/archive/compress)** - Compression algorithm implementations (GZIP, ZLIB, DEFLATE, BZIP2, LZ4, ZSTD, Snappy). The helper package wraps these algorithms with unified streaming interfaces.

- **[github.com/nabbar/golib/ioutils/nopwritecloser](https://pkg.go.dev/github.com/nabbar/golib/ioutils/nopwritecloser)** - io.WriteCloser wrapper utilities used internally for automatic interface conversions. Provides no-op Close() implementations for writers that don't natively support closing.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The helper package follows these conventions for idiomatic Go code.

- **[Go I/O Patterns](https://go.dev/blog/pipelines)** - Official Go blog article explaining pipeline patterns and streaming I/O. Relevant for understanding how helper fits into larger data processing pipelines with compression/decompression.

- **[compress Package](https://pkg.go.dev/compress)** - Standard library compression packages. The helper package provides a unified interface that works with both standard library and custom compression implementations.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/archive/helper`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
