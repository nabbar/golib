# Delim Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-98.6%25-brightgreen)](TESTING.md)

Lightweight, high-performance buffered reader for delimiter-separated data streams with custom delimiter support, constant memory usage, and zero-copy operations.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The `delim` package provides a buffered reader that efficiently processes data separated by any delimiter character. Unlike `bufio.Scanner` which is limited to newlines, `delim` offers complete flexibility with custom delimiters while maintaining constant memory usage regardless of data size.

### Design Philosophy

1. **Delimiter Flexibility**: Any rune character as delimiter (newlines, commas, pipes, tabs, null bytes, Unicode)
2. **Memory Efficient**: O(1) memory usage with streaming architecture
3. **Standard Interfaces**: Implements `io.ReadCloser` and `io.WriterTo`
4. **Zero-Copy Operations**: Direct passthrough where possible
5. **Simple API**: Minimal learning curve with familiar patterns

### Why Use This Package?

- **Custom Delimiters**: Not limited to newlines like `bufio.Scanner`
- **Delimiter Preservation**: Delimiter included in returned data
- **Configurable Buffers**: Optimize performance for your workload
- **Standard Interfaces**: Drop-in replacement for `io.ReadCloser`
- **Predictable Behavior**: Explicit control over buffer and delimiter handling

### Key Features

- **Custom Delimiters**: Any rune character (ASCII, Unicode, control characters)
- **Constant Memory**: ~32KB default buffer regardless of file size
- **98.6% Test Coverage**: 198 comprehensive test specs with race detection
- **Thread-Safe**: Safe for concurrent use (one goroutine per instance)
- **Multiple Read Methods**:
  - `Read(p []byte)` - io.Reader compatibility
  - `ReadBytes()` - Returns delimited chunks
  - `UnRead()` - Get buffered but unread data (consumes buffer)
  - `WriteTo(w io.Writer)` - Efficient streaming
- **Performance**: Sub-millisecond operations for typical workloads
- **No External Dependencies**: Only standard library + `github.com/nabbar/golib/size`

---

## Architecture

### Package Structure

```
ioutils/delim/
├── interface.go         # BufferDelim interface and New() constructor
├── model.go            # Internal dlm struct implementation
├── io.go               # I/O operations (Read, ReadBytes, WriteTo, etc.)
├── discard.go          # DiscardCloser no-op implementation
├── error.go            # ErrInstance error
└── doc.go              # Package documentation
```

### Component Overview

```
┌────────────────────────────────────────────────┐
│           BufferDelim Interface                │
│  io.ReadCloser + io.WriterTo + Custom Methods  │
└──────────┬─────────────────────────────────────┘
           │
┌──────────▼────────────────────────┐
│        dlm Implementation         │
│                                   │
│  ┌──────────────────────────────┐ │
│  │  io.ReadCloser (source)      │ │
│  └──────────────────────────────┘ │
│              │                    │
│              ▼                    │
│  ┌──────────────────────────────┐ │
│  │  Internal Buffer             │ │
│  └──────────────────────────────┘ │
│              │                    │
│              ▼                    │
│     delimiter detection           │
│              │                    │
│              ▼                    │
│     chunk extraction              │
└───────────────────────────────────┘
```

| Component | Memory | Complexity | Thread-Safe |
|-----------|--------|------------|-------------|
| **BufferDelim** | O(1) | Simple | ✅ per instance |
| **dlm** | O(1) | Internal | ✅ per instance |
| **DiscardCloser** | O(1) | Minimal | ✅ always |

### How It Works

1. **Initialization**: `New()` wraps an `io.ReadCloser` with an internal buffer
2. **Buffering**: Data read in configurable chunks (default 32KB)
3. **Delimiter Detection**: Scan buffer for delimiter byte
4. **Chunk Extraction**: Return data up to and including delimiter
5. **Memory Management**: Reuse buffer, minimal allocations

---

## Performance

### Memory Efficiency

**Constant Memory Usage** - The package maintains O(1) memory regardless of input size:

```
Buffer Size: 32KB (default) or custom
Memory Growth: ZERO (no additional allocation per delimiter)
Example: Process 10GB file using only ~32KB RAM
```

### Throughput Benchmarks

Performance measurements from test suite (AMD64, Go 1.21):

| Operation | Throughput | Latency (Median) | Memory |
|-----------|------------|------------------|--------|
| Read (Buffered) | ~6.2 GB/s | 3.2ms (20MB) | O(1) |
| ReadBytes (Small) | ~3.5 GB/s | 700µs (2.5MB) | O(1) |
| ReadBytes (Large) | ~5.8 GB/s | 13.6ms (80MB) | O(1) |
| WriteTo | ~1.5 GB/s | 51.6ms (80MB) | O(1) |
| CSV parsing | ~890 MB/s | 1.4ms (1.2MB) | O(1) |
| Log processing | ~630 MB/s | 4.7ms (3MB) | O(1) |

*Measured with default 32KB buffer, actual performance varies with buffer size and data characteristics*

### Buffer Size Impact

| Buffer Size | Use Case | Memory | Speed |
|-------------|----------|--------|-------|
| 64 bytes | Micro-optimization | Minimal | Slower (more reads) |
| 4KB | Legacy default | Low | Balanced |
| 32KB (default) | General purpose | Low | High performance |
| 64KB | Large records | Medium | Faster (fewer reads) |
| 1MB | High throughput | Higher | Fastest (bulk operations) |

**Recommendation**: Use default 32KB unless profiling shows benefit from larger buffers

### Comparison with Alternatives

```
delim Package:
├─ Custom delimiters: ✅ Any rune
├─ Delimiter in result: ✅ Included
├─ Memory: O(1) constant
├─ Buffer control: ✅ Configurable
└─ Interfaces: io.ReadCloser + io.WriterTo

bufio.Scanner:
├─ Custom delimiters: ⚠️  Via SplitFunc (complex)
├─ Delimiter in result: ❌ Removed
├─ Memory: O(1) but limited to MaxScanTokenSize
├─ Buffer control: ❌ Fixed
└─ Interfaces: Scanner only (custom)

strings.Split:
├─ Custom delimiters: ✅ String-based
├─ Delimiter in result: ❌ Removed  
├─ Memory: ❌ O(n) loads full data
├─ Buffer control: ❌ None
└─ Interfaces: Returns []string
```

---

## Use Cases

This package excels in scenarios requiring delimiter-based data processing:

**CSV/TSV Processing**
- Read CSV fields with custom separators (`,`, `;`, `|`)
- Process TSV data with tab delimiters
- Handle quoted fields that contain delimiters

**Log File Processing**
- Parse log entries separated by newlines
- Extract log records with custom markers
- Process structured logs with field delimiters

**Data Format Parsing**
- Null-terminated strings (C-style `\0`)
- Record-oriented data with custom separators
- Binary protocols with delimiter bytes

**Stream Processing**
- Real-time data ingestion from network streams
- Event processing with delimiter-separated events
- ETL pipelines with delimited records

**Text File Analysis**
- Line-by-line processing with custom line endings
- Extract sections separated by specific markers
- Parse configuration files with custom delimiters

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/delim
```

### Basic Line Reading

Read lines from a file:

```go
package main

import (
    "fmt"
    "io"
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    file, _ := os.Open("data.txt")
    defer file.Close()
    
    // Create BufferDelim with newline delimiter
    bd := delim.New(file, '\n', 0)  // 0 = default buffer (32KB)
    defer bd.Close()
    
    // Read line by line
    for {
        line, err := bd.ReadBytes()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        fmt.Printf("Line: %s", line)  // line includes '\n'
    }
}
```

### CSV Field Reading

Process CSV data with comma delimiter:

```go
package main

import (
    "fmt"
    "io"
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    file, _ := os.Open("data.csv")
    defer file.Close()
    
    bd := delim.New(file, ',', 0)
    defer bd.Close()
    
    for {
        field, err := bd.ReadBytes()
        if err == io.EOF {
            break
        }
        // field includes the comma
        processField(field)
    }
}

func processField(field []byte) {
    // Remove trailing comma if needed
    if len(field) > 0 && field[len(field)-1] == ',' {
        field = field[:len(field)-1]
    }
    fmt.Printf("Field: %s\n", field)
}
```

### Streaming Copy with Delimiters

Copy data while preserving delimiter structure:

```go
package main

import (
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    input, _ := os.Open("input.txt")
    defer input.Close()
    
    output, _ := os.Create("output.txt")
    defer output.Close()
    
    bd := delim.New(input, '\n', 0)
    defer bd.Close()
    
    // Stream copy with delimiter awareness
    _, err := bd.WriteTo(output)
    if err != nil {
        panic(err)
    }
}
```

### Large Buffer for High Throughput

Optimize for large records:

```go
package main

import (
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
    "github.com/nabbar/golib/size"
)

func main() {
    file, _ := os.Open("large-records.txt")
    defer file.Close()
    
    // Use 64KB buffer for better performance
    bd := delim.New(file, '\n', 64*size.KiB)
    defer bd.Close()
    
    // Process large records efficiently
    for {
        record, err := bd.ReadBytes()
        if err != nil {
            break
        }
        processLargeRecord(record)
    }
}

func processLargeRecord(record []byte) {
    // Handle large records
}
```

### Using Read() Method

Standard io.Reader interface:

```go
package main

import (
    "fmt"
    "io"
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    file, _ := os.Open("data.txt")
    defer file.Close()
    
    bd := delim.New(file, '\n', 0)
    defer bd.Close()
    
    // Use Read() for io.Reader compatibility
    buf := make([]byte, 1024)
    for {
        n, err := bd.Read(buf)
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        
        chunk := buf[:n]  // includes delimiter
        fmt.Printf("Chunk: %s", chunk)
    }
}
```

### Peeking at Buffered Data

Look ahead without consuming:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    file, _ := os.Open("data.txt")
    defer file.Close()
    
    bd := delim.New(file, '\n', 1024)  // 1KB buffer
    defer bd.Close()
    
    // Get buffered data
    buffered, _ := bd.UnRead()
    fmt.Printf("Next %d bytes: %s\n", len(buffered), buffered)
    
    // Data is now consumed, next Read() will get different data
}
```

### Null-Terminated Strings

Process C-style strings:

```go
package main

import (
    "fmt"
    "io"
    "os"
    
    "github.com/nabbar/golib/ioutils/delim"
)

func main() {
    file, _ := os.Open("binary-data.bin")
    defer file.Close()
    
    bd := delim.New(file, 0, 0)  // 0 = null byte delimiter
    defer bd.Close()
    
    for {
        str, err := bd.ReadBytes()
        if err == io.EOF {
            break
        }
        // str is null-terminated
        fmt.Printf("String: %q\n", str)
    }
}
```

---

## API Reference

### Types

#### BufferDelim Interface

```go
type BufferDelim interface {
    io.ReadCloser
    io.WriterTo
    
    Delim() rune
    Reader() io.ReadCloser
    Copy(w io.Writer) (n int64, err error)
    ReadBytes() ([]byte, error)
    UnRead() ([]byte, error)
}
```

Primary interface for delimiter-based reading.

**Methods**:
- `Read(p []byte) (n int, err error)` - Read one delimited chunk (from io.Reader)
- `Close() error` - Close and release resources (from io.Closer)
- `WriteTo(w io.Writer) (n int64, err error)` - Stream all data (from io.WriterTo)
- `Delim() rune` - Get the configured delimiter
- `Reader() io.ReadCloser` - Get self as io.ReadCloser
- `Copy(w io.Writer) (n int64, err error)` - Alias for WriteTo
- `ReadBytes() ([]byte, error)` - Read next delimited chunk
- `UnRead() ([]byte, error)` - Get buffered data (consumes buffer)

#### DiscardCloser Struct

```go
type DiscardCloser struct{}
```

No-op implementation of `io.ReadWriteCloser` for testing and benchmarks.

**Methods**:
- `Read(p []byte) (n int, err error)` - Always returns (0, nil)
- `Write(p []byte) (n int, err error)` - Always returns (len(p), nil)
- `Close() error` - Always returns nil

### Functions

#### New

```go
func New(r io.ReadCloser, delim rune, sizeBufferRead libsiz.Size) BufferDelim
```

Creates a new BufferDelim instance.

**Parameters**:
- `r` - Source reader to read from
- `delim` - Delimiter character (any rune)
- `sizeBufferRead` - Buffer size (0 for default 4KB)

**Returns**: BufferDelim instance

**Example**:
```go
bd := delim.New(file, '\n', 0)  // Default buffer
bd := delim.New(file, ',', 64*size.KiB)  // 64KB buffer
```

### Errors

#### ErrInstance

```go
var ErrInstance = fmt.Errorf("invalid buffer delim instance")
```

Returned when operations are attempted on a closed or invalid BufferDelim.

**When it occurs**:
- After `Close()` is called
- On nil instance
- When internal reader is invalidated

### Read Behavior

| Method | Returns Delimiter | Buffer Allocation | Error on EOF |
|--------|-------------------|-------------------|--------------|
| `Read()` | ✅ Yes | May expand | io.EOF |
| `ReadBytes()` | ✅ Yes | Slice to internal | io.EOF |
| `UnRead()` | N/A | New slice | ErrInstance |
| `WriteTo()` | ✅ Yes (writes) | None | io.EOF |

---

## Best Practices

### Resource Management

**Always close resources**:
```go
// ✅ Good
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()  // Ensure file is closed
    
    bd := delim.New(file, '\n', 0)
    defer bd.Close()  // Ensure BufferDelim is closed
    
    return processData(bd)
}

// ❌ Bad
func processBad(path string) {
    file, _ := os.Open(path)  // Never closed!
    bd := delim.New(file, '\n', 0)  // Never closed!
    processData(bd)
}
```

### Error Handling

**Check all errors**:
```go
// ✅ Good
data, err := bd.ReadBytes()
if err == io.EOF {
    return nil  // Normal end
}
if err != nil {
    return fmt.Errorf("read failed: %w", err)
}

// ❌ Bad
data, _ := bd.ReadBytes()  // Ignoring errors!
```

### Memory Efficiency

**Stream large files**:
```go
// ✅ Good: Constant memory
func process(path string) error {
    file, _ := os.Open(path)
    defer file.Close()
    
    bd := delim.New(file, '\n', 0)
    defer bd.Close()
    
    for {
        line, err := bd.ReadBytes()
        if err == io.EOF {
            break
        }
        processLine(line)  // Process incrementally
    }
    return nil
}

// ❌ Bad: Loads entire file
func processBad(path string) error {
    data, _ := os.ReadFile(path)  // Full file in RAM!
    lines := bytes.Split(data, []byte{'\n'})
    for _, line := range lines {
        processLine(line)
    }
    return nil
}
```

### Buffer Size Selection

**Choose appropriate buffer size**:
```go
// Small records (<1KB): Default buffer (32KB)
bd := delim.New(file, '\n', 0)

// Specific size (4KB)
bd := delim.New(file, '\n', 4*size.KiB)

// Large records (10KB-1MB): Large buffer
bd := delim.New(file, '\n', 64*size.KiB)

// Huge records (>1MB): Very large buffer
bd := delim.New(file, '\n', size.MiB)
```

### Delimiter Selection

**Common delimiters**:
```go
// Newlines (Unix)
bd := delim.New(file, '\n', 0)

// Carriage return (Mac classic)
bd := delim.New(file, '\r', 0)

// CSV comma
bd := delim.New(file, ',', 0)

// TSV tab
bd := delim.New(file, '\t', 0)

// Null-terminated (C strings)
bd := delim.New(file, 0, 0)

// Pipe-separated
bd := delim.New(file, '|', 0)

// Custom Unicode
bd := delim.New(file, '€', 0)
```

### Concurrency

**One instance per goroutine**:
```go
// ✅ Good: Independent instances
func processFiles(files []string) error {
    var wg sync.WaitGroup
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            processFile(f)  // Each goroutine has own BufferDelim
        }(file)
    }
    
    wg.Wait()
    return nil
}

// ❌ Bad: Shared instance
var sharedBD BufferDelim  // Race condition!

func processParallel() {
    go readData(sharedBD)  // UNSAFE!
    go readData(sharedBD)  // UNSAFE!
}
```

### Testing

The package includes a comprehensive test suite with **98.6% code coverage** and **198 test specifications** using Ginkgo v2 and Gomega. All tests pass with race detection enabled, ensuring thread safety.

**Quick test commands:**
```bash
go test ./...                          # Run all tests
go test -cover ./...                   # With coverage
CGO_ENABLED=1 go test -race ./...      # With race detection
```

See **[TESTING.md](TESTING.md)** for comprehensive testing documentation, including test architecture, performance benchmarks, and troubleshooting guides.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >98%)
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

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **98.6% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** per instance (one goroutine per BufferDelim)
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Performance Optimizations:**
1. Memory-mapped reading for very large files
2. Parallel processing for multi-delimiter extraction
3. SIMD optimizations for delimiter scanning
4. Buffer pooling for reduced GC pressure

**Feature Additions:**
1. Multi-byte delimiter support (string delimiters instead of rune)
2. Delimiter transformation (e.g., convert CRLF to LF)
3. Progress callbacks for long operations
4. Delimiter statistics and profiling
5. Record counting and indexing

**API Extensions:**
1. Scanner-compatible API wrapper for easier migration
2. Integration helpers with `encoding/csv`
3. Custom `bufio.Scanner` SplitFunc generator

**Quality of Life:**
1. Helper functions for common delimiters (CSV, TSV, etc.)
2. Delimiter auto-detection
3. Validation and sanitization options

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/delim)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, delimiter handling, buffer management, performance considerations, and comparison with `bufio.Scanner`. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 98.6% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/size](https://pkg.go.dev/github.com/nabbar/golib/size)** - Size constants and utilities (KiB, MiB, GiB, etc.) used for configurable buffer sizing. Provides type-safe size constants to avoid magic numbers and improve code readability when specifying buffer sizes.

- **[github.com/nabbar/golib/ioutils/aggregator](https://pkg.go.dev/github.com/nabbar/golib/ioutils/aggregator)** - Thread-safe write aggregator that can work with `delim` for concurrent log processing. Useful when multiple goroutines need to write delimiter-separated data to a single output stream.

### Standard Library References

- **[bufio](https://pkg.go.dev/bufio)** - Standard library buffered I/O package. The `delim` package provides similar delimiter-aware reading functionality but with additional control, custom delimiters, and constant memory usage. Understanding `bufio` helps in choosing the right tool for the task.

- **[io](https://pkg.go.dev/io)** - Standard I/O interfaces implemented by `delim`. The package fully implements `io.ReadCloser` and `io.WriterTo` for seamless integration with Go's I/O ecosystem and compatibility with existing tools and libraries.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The `delim` package follows these conventions for idiomatic Go code.

- **[Go I/O Patterns](https://go.dev/blog/pipelines)** - Official Go blog article explaining pipeline patterns and streaming I/O. Relevant for understanding how `delim` fits into larger data processing pipelines with delimiter-based segmentation.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the `delim` package. Check existing issues before creating new ones.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/ioutils/delim`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
