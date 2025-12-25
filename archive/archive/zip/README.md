# Zip Archive

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-68.9%25-yellow)](TESTING.md)

High-level interface for reading and writing ZIP archives with random access support and standard interface compliance.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Comparison with Standard Library](#comparison-with-standard-library)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Reading a ZIP Archive](#reading-a-zip-archive)
  - [Creating a ZIP Archive](#creating-a-zip-archive)
  - [Path Transformation](#path-transformation)
  - [Walking Files](#walking-files)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Reader Interface](#reader-interface)
  - [Writer Interface](#writer-interface)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **zip** package provides a high-level, user-friendly interface for working with ZIP archives. It wraps Go's standard `archive/zip` library with convenient methods that implement the `github.com/nabbar/golib/archive/archive/types` interfaces, making it consistent with other archive formats (tar, etc.) while maintaining full ZIP compatibility.

### Why Not Just Use archive/zip?

The standard library's `archive/zip` has several limitations that **zip** package addresses:

**Limitations of archive/zip:**
- ❌ **No query methods**: Cannot check if file exists without manual iteration
- ❌ **No path lookup**: Must iterate entire archive to find specific file
- ❌ **Complex directory archiving**: Manual directory walking required
- ❌ **No path transformation**: Must manually modify headers for renaming
- ❌ **Reader interface validation**: No validation of required interfaces

**How zip Package Extends archive/zip:**
- ✅ **High-level methods**: List(), Info(), Get(), Has(), Walk() for easy queries
- ✅ **Direct file access**: Get file by path with random access support
- ✅ **Automatic directory archiving**: FromPath() with glob filtering
- ✅ **Built-in path transformation**: ReplaceName function for path renaming
- ✅ **Interface validation**: Strict validation prevents runtime panics
- ✅ **Interface compliance**: Standard Reader/Writer interfaces for consistency

**Internally**, the package uses `archive/zip.Reader` and `archive/zip.Writer` for all ZIP operations, but adds convenient wrapper methods that handle the random access and interface validation automatically.

### Design Philosophy

1. **Extend Standard Library**: Build on `archive/zip` while addressing its limitations with high-level query methods and automatic interface validation.
2. **Interface Compliance**: Implements `github.com/nabbar/golib/archive/archive/types` interfaces for consistency across different archive formats (zip, tar, etc.).
3. **Simplicity First**: High-level methods (List, Get, Walk) handle common use cases without requiring detailed knowledge of ZIP format internals.
4. **Flexibility**: Supports advanced scenarios (path filtering, renaming) while keeping simple cases simple through intuitive API.
5. **Random Access**: Leverages ZIP's random access capability for efficient file retrieval without full archive scanning.
6. **Safety**: Proper resource cleanup with Close(), comprehensive error handling with fs.ErrNotExist, and path traversal protection with os.OpenRoot (Go 1.24).

### Key Features

- ✅ **Simple API**: Intuitive constructors (NewReader, NewWriter) and methods (List, Get, Add, FromPath)
- ✅ **High-level Operations**: Query methods (List, Info, Has, Walk) and extraction (Get)
- ✅ **Recursive Archiving**: Add entire directory trees with pattern filtering
- ✅ **Path Transformation**: Custom path renaming during archiving
- ✅ **Random Access**: Efficient file retrieval by path without sequential scan
- ✅ **Path Traversal Protection**: Secure file access with os.OpenRoot (Go 1.24)
- ✅ **Standard Interfaces**: Full `io.ReadCloser` and `io.WriteCloser` compatibility
- ✅ **Zero External Dependencies**: Only standard library + golib packages

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────┐
│              zip Package                       │
├────────────────────────────────────────────────┤
│                                                │
│  ┌──────────────┐         ┌──────────────┐     │
│  │   NewReader  │         │   NewWriter  │     │
│  └──────┬───────┘         └──────┬───────┘     │
│         │                        │             │
│         ▼                        ▼             │
│  ┌──────────────┐         ┌──────────────┐     │
│  │     rdr      │         │     wrt      │     │
│  │  (Reader)    │         │  (Writer)    │     │
│  └──────┬───────┘         └──────┬───────┘     │
│         │                        │             │
│         ▼                        ▼             │
│  ┌──────────────────────────────────────┐      │
│  │    archive/zip (std library)         │      │
│  └──────────────────────────────────────┘      │
│                                                │
└────────────────────────────────────────────────┘
```

**Reader Component:**
- Wraps `io.ReadCloser` with `zip.Reader`
- Requires: Size(), ReadAt(), Seeker interfaces
- Query methods: List(), Info(), Has()
- Extraction: Get(), Walk()

**Writer Component:**
- Wraps `io.WriteCloser` with `zip.Writer`
- Add individual files: Add()
- Recursive archiving: FromPath()
- Handles filtering, path transformation

### Data Flow

**Reading:**
1. Open file/stream → NewReader()
2. Validates interfaces (Size, ReadAt, Seeker)
3. Query: List(), Info(), Has()
4. Extract: Get() or Walk()
5. Close reader

**Writing:**
1. Create file/stream → NewWriter()
2. Add files: Add() or FromPath()
3. Close writer (finalizes archive)
4. Close underlying stream

### Comparison with Standard Library

| Feature | zip Package | archive/zip (stdlib) |
|---------|-------------|----------------------|
| **API Complexity** | Simple high-level methods | Low-level file-by-file API |
| **File Lookup** | Info(), Has(), Get() by path | Manual iteration required |
| **Directory Archiving** | FromPath() with filtering | Manual walking required |
| **Path Transformation** | Built-in ReplaceName function | Manual header modification |
| **Interface Validation** | Automatic validation | Runtime panics if missing |
| **Random Access** | Built-in with efficient lookup | Manual random access setup |

---

## Performance

### Benchmarks

Based on test suite measurements (Go 1.24+ with race detector):

| Operation | Median | Mean | Max | Samples |
|-----------|--------|------|-----|---------|
| **Reader List (5 files)** | <100µs | <100µs | <100µs | 1000 |
| **Reader Info (5 files)** | <100µs | <100µs | 200µs | 1000 |
| **Reader Get (5 files)** | <100µs | <100µs | <100µs | 1000 |
| **Reader Has (5 files)** | <100µs | <100µs | 200µs | 1000 |
| **Reader Walk (5 files)** | <100µs | <100µs | 100µs | 500 |
| **Writer Add (100B)** | <100µs | <100µs | 1ms | 1000 |
| **Writer Add (10KB)** | <100µs | <100µs | 200µs | 100 |
| **Writer Add (1MB)** | 400µs | 1.6ms | 5.6ms | 10 |

*Performance measured with race detector enabled. Non-race performance is approximately 2-3x faster.*

### Memory Usage

```
Reader overhead:    ~1KB (struct + zip.Reader)
Writer overhead:    ~1KB (struct + zip.Writer)
Per-file memory:    O(1) - random access
List() memory:      O(n) where n = number of files
```

**Memory Efficiency:**
- Random access allows constant memory per operation
- Files accessed directly without full scan
- List() stores only paths, not file contents

### Scalability

- **Small archives** (<100 files): Excellent performance with random access
- **Medium archives** (100-10K files): Good performance, random access eliminates scan overhead
- **Large archives** (>10K files): Random access remains efficient regardless of size
- **File sizes**: Tested with files up to 1MB, scales linearly

---

## Use Cases

### 1. Configuration Distribution

**Problem**: Package application configurations for deployment.

```go
file, _ := os.Create("config.zip")
defer file.Close()

writer, _ := zip.NewWriter(file)
defer writer.Close()

// Package all configs
writer.FromPath("/etc/myapp", "*.conf", nil)
```

**Real-world**: Application packages, deployment bundles, configuration archives.

### 2. Selective Extraction

**Problem**: Extract specific files from archives without full extraction.

```go
reader, _ := zip.NewReader(file)
defer reader.Close()

if reader.Has("config.json") {
    rc, _ := reader.Get("config.json")
    defer rc.Close()
    data, _ := io.ReadAll(rc)
    processConfig(data)
}
```

**Real-world**: Configuration extraction, partial restore, targeted file retrieval.

### 3. Archive Inspection

**Problem**: List and examine archive contents without extraction.

```go
reader, _ := zip.NewReader(file)
defer reader.Close()

files, _ := reader.List()
for _, path := range files {
    info, _ := reader.Info(path)
    fmt.Printf("%s: %d bytes\n", path, info.Size())
}
```

**Real-world**: Archive validation, content verification, file discovery.

### 4. Backup System

**Problem**: Create incremental backups with pattern filtering.

```go
writer.FromPath("/home/user/docs", "*.pdf", func(path string) string {
    return strings.TrimPrefix(path, "/home/user/")
})
```

**Real-world**: Document backups, selective archiving, data snapshots.

### 5. Data Distribution

**Problem**: Package and distribute data files efficiently.

```go
reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
    if strings.HasSuffix(path, ".dat") {
        data, _ := io.ReadAll(rc)
        processDataFile(path, data)
    }
    return true
})
```

**Real-world**: Software distribution, data packages, content delivery.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/archive/zip
```

### Reading a ZIP Archive

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/archive/zip"
)

func main() {
    file, err := os.Open("archive.zip")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    reader, err := zip.NewReader(file)
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()
    
    // List all files
    files, err := reader.List()
    if err != nil {
        log.Fatal(err)
    }
    
    for _, path := range files {
        fmt.Println(path)
    }
}
```

### Creating a ZIP Archive

```go
package main

import (
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/archive/zip"
)

func main() {
    file, err := os.Create("archive.zip")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    writer, err := zip.NewWriter(file)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()
    
    // Add files from directory
    err = writer.FromPath("/path/to/files", "*", nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Path Transformation

```go
// Strip base directory from archive paths
writer.FromPath("/var/log/app", "*.log", func(path string) string {
    return strings.TrimPrefix(path, "/var/log/app/")
})

// Flatten directory structure
writer.FromPath("/etc/config", "*", func(path string) string {
    return filepath.Base(path)
})
```

### Walking Files

```go
reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
    fmt.Printf("File: %s (%d bytes)\n", path, info.Size())
    
    if strings.HasSuffix(path, ".txt") {
        data, _ := io.ReadAll(rc)
        processTextFile(data)
    }
    
    return true // Continue to next file
})
```

---

## Best Practices

### Testing

The package includes comprehensive tests with **68.9% code coverage**, **33 test specifications**, and **0 race conditions** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All reader operations (List, Info, Get, Has, Walk)
- ✅ All writer operations (Add, FromPath, Close)
- ✅ Integration tests (full write-read cycles)
- ✅ Performance benchmarks (throughput, latency)

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Always close resources:**
```go
// ✅ GOOD: Proper resource management
func processArchive(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    reader, err := zip.NewReader(file)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    return processFiles(reader)
}
```

**Close writer before underlying stream:**
```go
// ✅ GOOD: Writer closed first
func createArchive(path string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    writer, err := zip.NewWriter(file)
    if err != nil {
        return err
    }
    defer writer.Close() // Must close before file.Close()
    
    return addFiles(writer)
}
```

**Use Has before Get:**
```go
// ✅ GOOD: Check existence first
if reader.Has("config.json") {
    rc, _ := reader.Get("config.json")
    defer rc.Close()
    processConfig(rc)
}
```

**Check for fs.ErrNotExist:**
```go
// ✅ GOOD: Proper error handling
info, err := reader.Info("config.json")
if err == fs.ErrNotExist {
    useDefaults()
} else if err != nil {
    return err
}
```

### ❌ DON'T

**Don't forget to close writer:**
```go
// ❌ BAD: Archive will be corrupted
writer, _ := zip.NewWriter(file)
writer.Add(info, rc, "file.txt", "")
// Missing writer.Close()!
```

**Don't use multiple operations simultaneously:**
```go
// ❌ BAD: Not thread-safe
go reader.Get("file1.txt")
go reader.Get("file2.txt")

// ✅ GOOD: Use separate readers
reader1, _ := zip.NewReader(stream1)
reader2, _ := zip.NewReader(stream2)
go reader1.Get("file1.txt")
go reader2.Get("file2.txt")
```

**Don't use streaming sources for reader:**
```go
// ❌ BAD: ZIP requires Size(), ReadAt(), Seeker
resp, _ := http.Get("http://example.com/file.zip")
reader, _ := zip.NewReader(resp.Body) // Will fail!

// ✅ GOOD: Download first or use proper interface
data, _ := io.ReadAll(resp.Body)
reader, _ := zip.NewReader(bytes.NewReader(data))
```

---

## API Reference

### Reader Interface

```go
type Reader interface {
    io.Closer
    
    List() ([]string, error)
    Info(path string) (fs.FileInfo, error)
    Get(path string) (io.ReadCloser, error)
    Has(path string) bool
    Walk(fct FuncExtract)
}
```

**Methods:**

- **`List()`**: Returns all file paths in the archive
- **`Info(path)`**: Gets file metadata for a specific path
- **`Get(path)`**: Extracts a file as io.ReadCloser
- **`Has(path)`**: Checks if a file exists (no error)
- **`Walk(callback)`**: Iterates all files with a callback

**Reader Requirements:**
- `io.ReadCloser` must implement `Size() int64`
- `io.ReadCloser` must implement `io.ReaderAt`
- `io.ReadCloser` must implement `io.Seeker`

### Writer Interface

```go
type Writer interface {
    io.Closer
    
    Add(info fs.FileInfo, reader io.ReadCloser, forcePath, target string) error
    FromPath(source, filter string, replaceName ReplaceName) error
}
```

**Methods:**

- **`Add(info, reader, forcePath, target)`**: Adds a single file
  - `info`: File metadata
  - `reader`: File contents (nil for directories/links)
  - `forcePath`: Custom path in archive (empty uses info.Name())
  - `target`: Link target (empty for regular files)

- **`FromPath(source, filter, replaceName)`**: Adds files recursively
  - `source`: Directory or file path
  - `filter`: Glob pattern (e.g., "*.go", "*" for all)
  - `replaceName`: Function to transform paths (nil for no transformation)

### Error Handling

```go
// Standard errors
fs.ErrNotExist  // File not found in archive (Info, Get)
fs.ErrInvalid   // Invalid file type, operation, or interface
```

**Example:**
```go
info, err := reader.Info("config.json")
if err == fs.ErrNotExist {
    log.Println("Config not found, using defaults")
} else if err != nil {
    log.Fatalf("Error: %v", err)
}
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

- ✅ **68.9% test coverage** (approaching 80% target)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** per-instance (one goroutine per reader/writer)
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard compliance** with archive/zip format

### Security Features

- ✅ **Path Traversal Protection**: Uses Go 1.24's os.OpenRoot for secure file access
- ✅ **Interface Validation**: Strict validation of reader interfaces prevents runtime panics
- ✅ **Error Standardization**: Consistent error handling with fs package errors
- ✅ **Resource Cleanup**: Proper defer patterns ensure resource release

### Known Limitations

**Architectural Constraints:**

1. **Random Access Required**: Reader requires Size(), ReadAt(), and Seeker interfaces
2. **No Streaming Sources**: Cannot read from network streams without buffering
3. **Single-threaded Operations**: Reader/Writer instances not thread-safe
4. **No Compression Control**: Uses default compression level

**Not Suitable For:**
- Streaming archives from network (use tar for sequential access)
- Concurrent access to single reader/writer instance
- Custom compression levels

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Compression Level Control**: Expose compression level configuration
2. **Streaming Support**: Buffered reader for network sources
3. **Concurrent Reading**: Thread-safe reader implementation
4. **Progress Callbacks**: Report progress during FromPath operations
5. **Archive Modification**: Support for adding/removing files from existing archives

These are **optional improvements** and not required for production use. The current implementation is stable and performant for standard ZIP operations.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/zip)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, and comprehensive usage examples. Provides detailed explanations of reader/writer operations and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/archive/archive/types](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/types)** - Common archive interfaces implemented by zip package. Provides Reader and Writer interfaces for uniform archive handling across different formats.

- **[github.com/nabbar/golib/archive/archive/tar](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/tar)** - TAR archive implementation with the same interface. Use tar for sequential streaming, zip for random access.

### Standard Library References

- **[archive/zip](https://pkg.go.dev/archive/zip)** - Standard library ZIP implementation used internally. The zip package provides a higher-level API wrapping this functionality.

- **[io/fs](https://pkg.go.dev/io/fs)** - Standard filesystem interfaces used for file information. The package uses fs.FileInfo and fs.ErrNotExist for consistency with Go's filesystem APIs.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The zip package follows these conventions for idiomatic Go code.

- **[ZIP File Format Specification](https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT)** - Official PKWARE ZIP format specification. Understanding the ZIP format helps explain random access capabilities and limitations.

- **[Go 1.24 os.OpenRoot](https://go.dev/blog/go1.24)** - Introduction to os.OpenRoot for secure file operations. The zip package uses this for path traversal protection in Go 1.24+.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/archive/archive/zip`
