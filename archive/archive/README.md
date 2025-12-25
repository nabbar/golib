# Archive Package

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-75.2%25-yellow)](TESTING.md)

Unified archive handling for TAR and ZIP formats with automatic format detection, providing a consistent interface for working with different archive types.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Sub-Packages](#sub-packages)
- [Performance](#performance)
  - [Detection Performance](#detection-performance)
  - [Format Comparison](#format-comparison)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Archive Creation](#archive-creation)
  - [Archive Extraction](#archive-extraction)
  - [Format Detection](#format-detection)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Algorithm Type](#algorithm-type)
  - [Detection Functions](#detection-functions)
  - [Reader/Writer Factories](#readerwriter-factories)
  - [Encoding Support](#encoding-support)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **archive** package provides unified handling of TAR and ZIP archive formats with automatic format detection. It abstracts away the fundamental differences between sequential (TAR) and random-access (ZIP) archives, providing a consistent interface for both reading and writing operations.

### Why Use This Package?

Standard library `archive/tar` and `archive/zip` packages have different APIs and operational models:

**Limitations of Using Standard Library Directly:**
- ❌ **Different APIs**: TAR and ZIP have completely different interfaces
- ❌ **No format detection**: Must know format before opening
- ❌ **Sequential vs Random**: TAR requires sequential access, ZIP allows random
- ❌ **Manual format handling**: Application must handle format differences
- ❌ **No unified iteration**: Different methods for walking through files
- ❌ **Static format selection**: Cannot switch between formats easily

**How archive Package Helps:**
- ✅ **Unified interface**: Single Reader/Writer interface for both formats
- ✅ **Automatic detection**: Detect format from magic numbers in header
- ✅ **Format abstraction**: Walk() method works for both TAR and ZIP
- ✅ **Easy switching**: Parse() and Algorithm enum for format selection
- ✅ **JSON/Text marshaling**: Serialize format configuration
- ✅ **Type safety**: Enum-based algorithm prevents invalid format strings

**Internally**, the package delegates to standard library `archive/tar` and `archive/zip` implementations while providing a consistent facade that respects the fundamental differences between formats.

### Design Philosophy

1. **Format Abstraction**: Single interface for multiple archive formats without hiding their differences
2. **Auto-Detection**: Automatic format identification from magic numbers
3. **Standard Compliance**: Implements encoding.TextMarshaler and encoding.JSONMarshaler
4. **Zero-Copy Wrapping**: Efficient delegation to standard library implementations
5. **Type Safety**: Enum-based algorithm selection prevents invalid format strings
6. **Structural Awareness**: Exposes and respects TAR (sequential) vs ZIP (random-access) differences

### Key Features

- ✅ **Algorithm Enumeration**: Type-safe format selection (None, Tar, Zip)
- ✅ **Automatic Detection**: Magic number-based format detection
- ✅ **Unified Interfaces**: types.Reader and types.Writer for both formats
- ✅ **Format-Independent Iteration**: Walk() method for both TAR and ZIP
- ✅ **JSON/Text Marshaling**: Configuration serialization support
- ✅ **File Extensions**: Automatic extension generation (.tar, .zip)
- ✅ **Header Validation**: DetectHeader() for format verification
- ✅ **75.2% Coverage**: Comprehensive testing with 115 test specifications

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────┐
│                   archive (main package)                   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────────┐      ┌────────────────────────┐      │
│  │ Algorithm (enum) │      │  Detection & Parsing   │      │
│  │                  │      │                        │      │
│  │ • None           │      │ • Parse(string)        │      │
│  │ • Tar            │      │ • Detect(io.Reader)    │      │
│  │ • Zip            │      │ • DetectHeader([]byte) │      │
│  │                  │      │                        │      │
│  │ • String()       │      │                        │      │
│  │ • Extension()    │      │                        │      │
│  │ • IsNone()       │      │                        │      │
│  └──────────────────┘      └────────────────────────┘      │
│                                                            │
│  ┌──────────────────────────────────────────────────┐      │
│  │         Reader/Writer Factory                    │      │
│  │                                                  │      │
│  │ • Algorithm.Reader(io.ReadCloser)                │      │
│  │     → types.Reader interface                     │      │
│  │                                                  │      │
│  │ • Algorithm.Writer(io.WriteCloser)               │      │
│  │     → types.Writer interface                     │      │
│  └──────────────────────────────────────────────────┘      │
│                                                            │
│  ┌──────────────────────────────────────────────────┐      │
│  │         Encoding/Marshaling                      │      │
│  │                                                  │      │
│  │ • MarshalText() / UnmarshalText()                │      │
│  │ • MarshalJSON() / UnmarshalJSON()                │      │
│  └──────────────────────────────────────────────────┘      │
│                                                            │
└──────────────────────┬─────────────────────────────────────┘
                       │
         ┌─────────────┴─────────────┐
         ▼                           ▼
┌─────────────────┐         ┌─────────────────┐
│   archive/tar   │         │   archive/zip   │
│  (sequential)   │         │ (random access) │
└─────────────────┘         └─────────────────┘
         │                           │
         ▼                           ▼
┌─────────────────────────────────────────────┐
│     Standard Library (archive/tar, /zip)    │
└─────────────────────────────────────────────┘
```

### Data Flow

1. **Detection Flow**:
   - Input stream arrives at `Detect()`
   - Peek first 265 bytes using bufio.Reader
   - Check magic numbers: TAR ("ustar\x00" at 257) or ZIP (0x504B0304 at 0)
   - Return Algorithm, types.Reader, and buffered stream
   - Reader preserves peeked data via internal adapter

2. **Read Operation**:
   - Algorithm.Reader() creates format-specific reader
   - TAR: Sequential access only, wraps archive/tar
   - ZIP: Random access, requires ReaderAt/Seeker
   - Walk() provides unified iteration
   - Get() returns specific file content

3. **Write Operation**:
   - Algorithm.Writer() creates format-specific writer
   - Add files via Add() method or FromPath() helper
   - TAR: Streaming writes, immediate flush
   - ZIP: Central directory at end
   - Close() finalizes archive structure

### Sub-Packages

This package consists of three sub-packages:

**archive/archive/types**:
- Defines Reader and Writer interfaces
- Provides FuncExtract callback type for Walk()
- Provides ReplaceName callback type for path transformation
- See [types/README.md](types/README.md)

**archive/archive/tar**:
- TAR format implementation (sequential access)
- Wraps standard library archive/tar
- Supports hard links and symbolic links
- See [tar/README.md](tar/README.md)

**archive/archive/zip**:
- ZIP format implementation (random access via central directory)
- Wraps standard library archive/zip
- Does not preserve hard links or symbolic links
- See [zip/README.md](zip/README.md)

---

## Performance

### Detection Performance

Format detection is extremely fast, requiring only a 265-byte header peek:

| Operation | Complexity | Typical Latency |
|-----------|------------|-----------------|
| **Peek Header** | O(1) | ~1-2µs |
| **Format Match** | O(1) | <1µs |
| **Total Detection** | O(1) | ~2-3µs |

*Performance measured on AMD64, Go 1.25*

### Format Comparison

Understanding the performance characteristics of each format:

**TAR (Sequential)**:
- **List()**: O(n) - must scan entire archive
- **Get(file)**: O(n) - must scan until found
- **Has(file)**: O(n) - must scan until found
- **Walk()**: O(n) - single sequential pass
- **Memory**: O(1) - constant, streaming-friendly
- **Best for**: Backups, streaming, network transfers

**ZIP (Random Access)**:
- **List()**: O(1) - reads central directory only
- **Get(file)**: O(1) - direct seek via directory
- **Has(file)**: O(1) - lookup in directory
- **Walk()**: O(n) - iterates directory entries
- **Memory**: O(n) - central directory in memory
- **Best for**: Random file access, GUI tools, distribution

### Scalability

- **Writers**: Both formats scale well with file count
- **Concurrency**: Package is not thread-safe per instance (design choice for performance)
- **Throughput**: Limited by underlying I/O, minimal overhead from abstraction layer
- **Memory**: TAR constant, ZIP proportional to file count

---

## Use Cases

### 1. Automatic Backup Restoration

Restore backups without knowing the archive format:

```go
func RestoreBackup(backupFile, targetDir string) error {
    file, _ := os.Open(backupFile)
    defer file.Close()
    
    alg, reader, stream, err := archive.Detect(file)
    if err != nil || reader == nil {
        return fmt.Errorf("invalid backup archive")
    }
    defer stream.Close()
    defer reader.Close()
    
    log.Printf("Restoring %s backup...", alg.String())
    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
        // Extract files to targetDir
        return true
    })
    return nil
}
```

### 2. Format Conversion

Convert between TAR and ZIP formats:

```go
func ConvertArchive(src, dst string, dstFormat archive.Algorithm) error {
    srcFile, _ := os.Open(src)
    defer srcFile.Close()
    _, srcReader, srcStream, _ := archive.Detect(srcFile)
    defer srcStream.Close()
    defer srcReader.Close()
    
    dstFile, _ := os.Create(dst)
    defer dstFile.Close()
    dstWriter, _ := dstFormat.Writer(dstFile)
    defer dstWriter.Close()
    
    srcReader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
        dstWriter.Add(info, r, path, link)
        return true
    })
    return nil
}
```

### 3. Configuration-Driven Archiving

Use JSON/YAML configuration to select format:

```go
type Config struct {
    ArchiveFormat archive.Algorithm `json:"format"`
}

func CreateArchive(cfg Config, files []string, output string) error {
    outFile, _ := os.Create(output)
    defer outFile.Close()
    
    writer, _ := cfg.ArchiveFormat.Writer(outFile)
    defer writer.Close()
    
    for _, file := range files {
        // Add files to archive
    }
    return nil
}
```

### 4. Selective Extraction

Extract only files matching a pattern:

```go
func ExtractMatching(archivePath, pattern, destDir string) error {
    file, _ := os.Open(archivePath)
    defer file.Close()
    
    _, reader, stream, _ := archive.Detect(file)
    defer stream.Close()
    defer reader.Close()
    
    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
        matched, _ := filepath.Match(pattern, filepath.Base(path))
        if matched {
            // Extract matching file
        }
        return true
    })
    return nil
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/archive
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/archive/archive"
)

func main() {
    // Algorithm selection
    alg := archive.Tar
    fmt.Println(alg.String())     // "tar"
    fmt.Println(alg.Extension())  // ".tar"
    
    // Parse from string
    alg = archive.Parse("zip")
    fmt.Println(alg == archive.Zip) // true
}
```

### Archive Creation

Creating a TAR archive:

```go
package main

import (
    "os"
    "github.com/nabbar/golib/archive/archive"
)

func main() {
    outFile, _ := os.Create("output.tar")
    defer outFile.Close()
    
    writer, _ := archive.Tar.Writer(outFile)
    defer writer.Close()
    
    // Add entire directory recursively
    writer.FromPath("/path/to/source", "*", nil)
}
```

### Archive Extraction

Reading with known format:

```go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/nabbar/golib/archive/archive"
)

func main() {
    inFile, _ := os.Open("input.tar")
    defer inFile.Close()
    
    reader, _ := archive.Tar.Reader(inFile)
    defer reader.Close()
    
    // List all files
    files, _ := reader.List()
    fmt.Println(files)
    
    // Extract specific file
    fileReader, _ := reader.Get("path/in/archive.txt")
    defer fileReader.Close()
    data, _ := io.ReadAll(fileReader)
}
```

### Format Detection

Automatic format detection:

```go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/nabbar/golib/archive/archive"
)

func main() {
    file, _ := os.Open("unknown-format.archive")
    defer file.Close()
    
    // Detect format and get appropriate reader
    alg, reader, stream, err := archive.Detect(file)
    if err != nil {
        panic(err)
    }
    defer stream.Close()
    
    if reader == nil {
        fmt.Println("Not a recognized archive format")
        return
    }
    defer reader.Close()
    
    fmt.Printf("Detected format: %s\n", alg.String())
    
    // Walk through all files (works for both TAR and ZIP)
    reader.Walk(func(info os.FileInfo, r io.ReadCloser, path, link string) bool {
        fmt.Printf("File: %s (%d bytes)\n", path, info.Size())
        return true // continue walking
    })
}
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **75.2% code coverage** and **115 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All algorithm operations (String, Extension, IsNone, Parse)
- ✅ Format detection with TAR and ZIP
- ✅ Reader/Writer factory methods
- ✅ JSON and Text marshaling/unmarshaling
- ✅ Error handling and edge cases
- ✅ Zero race conditions detected

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Use Detection for Unknown Formats:**
```go
// ✅ GOOD: Let the package detect the format
alg, reader, stream, err := archive.Detect(file)
if err != nil {
    return err
}
defer stream.Close()
if reader != nil {
    defer reader.Close()
}
```

**Parse User Input:**
```go
// ✅ GOOD: Parse user-provided format strings
alg := archive.Parse(userInput)
if alg == archive.None {
    return fmt.Errorf("unsupported format: %s", userInput)
}
```

**Use Walk() for Unified Processing:**
```go
// ✅ GOOD: Works for both TAR and ZIP
reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
    // Process each file
    return true
})
```

**Close in Proper Order:**
```go
// ✅ GOOD: Close reader before stream
defer stream.Close()
defer reader.Close() // Closes first (defer is LIFO)
```

### ❌ DON'T

**Don't assume TAR supports random access:**
```go
// ❌ BAD: TAR readers must be read sequentially
files, _ := reader.List()
for _, file := range files {
    reader.Get(file) // Requires re-reading for TAR!
}

// ✅ GOOD: Use Walk() for sequential access
reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path, link string) bool {
    // Process file here
    return true
})
```

**Don't use ZIP with non-seekable streams:**
```go
// ❌ BAD: ZIP requires seekable stream
pipeReader, pipeWriter := io.Pipe()
reader, _ := archive.Zip.Reader(pipeReader) // Will error!

// ✅ GOOD: Use TAR for pipes and streams
reader, _ := archive.Tar.Reader(pipeReader)
```

**Don't forget to check for nil reader:**
```go
// ❌ BAD: Detect() returns nil reader for unknown formats
_, reader, stream, _ := archive.Detect(file)
reader.List() // Panic if format unknown!

// ✅ GOOD: Always check for nil
if reader == nil {
    return fmt.Errorf("unknown archive format")
}
defer reader.Close()
```

**Don't modify Algorithm values:**
```go
// ❌ BAD: Never modify enum values
var myAlg archive.Algorithm = 99 // Undefined behavior!

// ✅ GOOD: Use predefined constants only
alg := archive.Tar
```

---

## API Reference

### Algorithm Type

```go
type Algorithm uint8

const (
    None Algorithm = iota // No archive format
    Tar                   // TAR format (sequential)
    Zip                   // ZIP format (random access)
)
```

**Methods:**
- `String() string` - Returns "tar", "zip", or "none"
- `Extension() string` - Returns ".tar", ".zip", or ""
- `IsNone() bool` - Returns true if None
- `DetectHeader([]byte) bool` - Validates header magic numbers
- `Reader(io.ReadCloser) (types.Reader, error)` - Creates reader
- `Writer(io.WriteCloser) (types.Writer, error)` - Creates writer
- `MarshalText() ([]byte, error)` - Text encoding
- `UnmarshalText([]byte) error` - Text decoding
- `MarshalJSON() ([]byte, error)` - JSON encoding
- `UnmarshalJSON([]byte) error` - JSON decoding

### Detection Functions

```go
func Parse(s string) Algorithm
```

Converts string to Algorithm (case-insensitive). Returns None for unknown.

```go
func Detect(r io.ReadCloser) (Algorithm, types.Reader, io.ReadCloser, error)
```

Automatically detects archive format and returns appropriate reader.

**Returns:**
- Algorithm: detected format (Tar, Zip, or None)
- types.Reader: format-specific reader (nil if None or error)
- io.ReadCloser: buffered stream preserving peeked data
- error: peek error or reader creation error

### Reader/Writer Factories

See [types/README.md](types/README.md) for detailed interface documentation.

**types.Reader Interface:**
- `Close() error`
- `List() ([]string, error)`
- `Info(path string) (fs.FileInfo, error)`
- `Get(path string) (io.ReadCloser, error)`
- `Has(path string) bool`
- `Walk(fn types.FuncExtract) error`

**types.Writer Interface:**
- `Close() error`
- `Add(info fs.FileInfo, r io.ReadCloser, path, link string) error`
- `FromPath(path, pattern string, replaceName types.ReplaceName) error`

### Encoding Support

The Algorithm type implements standard encoding interfaces:

```go
// Text encoding (for YAML, TOML, etc.)
text, _ := archive.Tar.MarshalText()        // []byte("tar")
var alg archive.Algorithm
alg.UnmarshalText([]byte("zip"))            // alg == archive.Zip

// JSON encoding
json, _ := archive.Tar.MarshalJSON()        // "tar"
archive.None.MarshalJSON()                   // null
var alg archive.Algorithm
alg.UnmarshalJSON([]byte(`"zip"`))          // alg == archive.Zip
```

### Error Codes

```go
var ErrInvalidAlgorithm = errors.New("invalid algorithm")
```

Returned when attempting to create Reader/Writer for None algorithm.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >75%)
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

- ✅ **75.2% test coverage** (target: >75%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Type-safe** enum-based algorithm selection
- ✅ **Standard compliant** implements encoding interfaces
- ✅ **Well documented** comprehensive GoDoc and examples

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Additional Formats**: Support for 7z, RAR (read-only), or other archive formats
2. **Compression Integration**: Built-in compression handling (gzip, bzip2, xz)
3. **Streaming ZIP**: Support for streaming ZIP creation without central directory
4. **Parallel Extraction**: Concurrent file extraction for ZIP archives
5. **Progress Callbacks**: Optional progress reporting for long operations

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/archive)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, format comparison, auto-detection mechanism, and implementation details. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 75.2% coverage analysis, and guidelines for writing new tests. Includes troubleshooting and bug reporting templates.

### Sub-Package Documentation

- **[types/README.md](types/README.md)** - Reader and Writer interface definitions with detailed method documentation and usage patterns.

- **[tar/README.md](tar/README.md)** - TAR format implementation details, sequential access patterns, and performance characteristics.

- **[zip/README.md](zip/README.md)** - ZIP format implementation details, random access capabilities, and central directory management.

### Related golib Packages

- **[github.com/nabbar/golib/archive/compress](../compress)** - Compression wrapper (gzip, bzip2, lz4, xz) for use with archive formats. Enables creation and extraction of compressed archives like .tar.gz and .tar.xz.

- **[github.com/nabbar/golib/archive/helper](../helper)** - Helper utilities for common archive operations and patterns.

### Standard Library References

- **[archive/tar](https://pkg.go.dev/archive/tar)** - Standard library TAR format support. The archive package wraps and extends this with unified interfaces.

- **[archive/zip](https://pkg.go.dev/archive/zip)** - Standard library ZIP format support. The archive package wraps and extends this with unified interfaces.

- **[io/fs](https://pkg.go.dev/io/fs)** - File system interfaces used by Reader.Info() and Writer.Add() methods.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and type assertions. The archive package follows these conventions for idiomatic Go code.

- **[TAR Format Specification](https://www.gnu.org/software/tar/manual/html_node/Standard.html)** - POSIX ustar format specification. Understanding TAR structure helps when working with sequential archives.

- **[ZIP Format Specification](https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT)** - PKWARE's APPNOTE.TXT specification. Understanding ZIP structure helps when working with random-access archives.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)
**Package**: `github.com/nabbar/golib/archive/archive`
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
