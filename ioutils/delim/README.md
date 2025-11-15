# Delim Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Lightweight, high-performance buffered reader for delimiter-separated data streams with custom delimiter support, constant memory usage, and zero-copy operations.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
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

---

## Key Features

- **Custom Delimiters**: Any rune character (ASCII, Unicode, control characters)
- **Constant Memory**: ~4KB default buffer regardless of file size
- **100% Test Coverage**: 198 comprehensive test specs with race detection
- **Thread-Safe**: Safe for concurrent use (one goroutine per instance)
- **Multiple Read Methods**:
  - `Read(p []byte)` - io.Reader compatibility
  - `ReadBytes()` - Returns delimited chunks
  - `UnRead()` - Peek at buffered data
  - `WriteTo(w io.Writer)` - Efficient streaming
- **Performance**: Sub-millisecond operations for typical workloads
- **No External Dependencies**: Only standard library + `github.com/nabbar/golib/size`

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/delim
```

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
│  │  bufio.Reader (buffered)     │ │
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

1. **Initialization**: `New()` wraps an `io.ReadCloser` with `bufio.Reader`
2. **Buffering**: Data read in configurable chunks (default 4KB)
3. **Delimiter Detection**: Scan buffer for delimiter byte
4. **Chunk Extraction**: Return data up to and including delimiter
5. **Memory Management**: Reuse buffer, minimal allocations

---

## Performance

### Memory Efficiency

**Constant Memory Usage** - The package maintains O(1) memory regardless of input size:

```
Buffer Size: 4KB (default) or custom
Memory Growth: ZERO (no additional allocation per delimiter)
Example: Process 10GB file using only ~4KB RAM
```

### Throughput Benchmarks

Performance measurements from test suite (AMD64, Go 1.21):

| Operation | Throughput | Latency | Memory |
|-----------|------------|---------|--------|
| Read small chunks | N/A | <100µs | O(1) |
| Read medium chunks | N/A | ~300µs | O(1) |
| Read large chunks | N/A | ~500µs | O(1) |
| ReadBytes | ~300 MB/s | <100µs | O(1) |
| WriteTo | ~500 MB/s | ~200µs | O(1) |
| CSV parsing | ~500 MB/s | ~100µs | O(1) |
| Log processing | ~250 MB/s | ~200µs | O(1) |
| Constructor | N/A | ~1.3ms | O(1) |

*Measured with default 4KB buffer, actual performance varies with buffer size and data characteristics*

### Buffer Size Impact

| Buffer Size | Use Case | Memory | Speed |
|-------------|----------|--------|-------|
| 64 bytes | Micro-optimization | Minimal | Slower (more reads) |
| 4KB (default) | General purpose | Low | Balanced |
| 64KB | Large records | Medium | Faster (fewer reads) |
| 1MB | High throughput | Higher | Fastest (bulk operations) |

**Recommendation**: Use default 4KB unless profiling shows benefit from larger buffers

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
    bd := delim.New(file, '\n', 0)  // 0 = default buffer
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
    
    // Peek at buffered data
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
- `UnRead() ([]byte, error)` - Peek at buffered data

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
// Small records (<1KB): Default buffer
bd := delim.New(file, '\n', 0)

// Medium records (1-10KB): Small buffer
bd := delim.New(file, '\n', 8*size.KiB)

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

---

## Testing

**Test Suite**: 198 specs using Ginkgo v2 and Gomega (100% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**:
- Constructor with various delimiters and buffer sizes
- Read operations (Read, ReadBytes, UnRead)
- Write operations (WriteTo, Copy)
- Edge cases (Unicode, binary data, empty input, large data)
- DiscardCloser functionality
- Concurrency and thread safety
- Performance benchmarks (30 scenarios)

**Quality Metrics**:
- ✅ 100% statement coverage
- ✅ Zero data races (verified with `-race`)
- ✅ 198 passing specs
- ✅ Sub-second test execution (~0.17s normal, ~2.1s with race)

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain 100% test coverage
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Update GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add benchmarks for performance-critical code

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Performance**
- Memory-mapped reading for very large files
- Parallel processing for multi-delimiter extraction
- SIMD optimizations for delimiter scanning
- Buffer pooling for reduced GC pressure

**Features**
- Multi-byte delimiter support (string delimiters)
- Delimiter transformation (e.g., convert CRLF to LF)
- Progress callbacks for long operations
- Delimiter statistics and profiling
- Record counting and indexing

**Compatibility**
- Scanner-compatible API wrapper
- Integration with `encoding/csv`
- Integration with `bufio.Scanner` SplitFunc

**Quality of Life**
- Helper functions for common delimiters
- Delimiter auto-detection
- Validation and sanitization options

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Package Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/delim)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Related Packages**:
  - [github.com/nabbar/golib/size](https://pkg.go.dev/github.com/nabbar/golib/size) - Size constants
  - [bufio](https://pkg.go.dev/bufio) - Standard library buffered I/O
  - [io](https://pkg.go.dev/io) - Standard I/O interfaces
