# Tar Archive

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-85.6%25-brightgreen)](TESTING.md)

High-level interface for reading and writing tar archives with simplified operations and standard interface compliance.

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
  - [Reading a Tar Archive](#reading-a-tar-archive)
  - [Creating a Tar Archive](#creating-a-tar-archive)
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

The **tar** package provides a high-level, user-friendly interface for working with tar archives. It wraps Go's standard `archive/tar` library with convenient methods that implement the `github.com/nabbar/golib/archive/archive/types` interfaces, making it consistent with other archive formats (zip, etc.) while maintaining full tar compatibility.

### Why Not Just Use archive/tar?

The standard library's `archive/tar` has several limitations that **tar** package addresses:

**Limitations of archive/tar:**
- ❌ **Low-level API**: Requires manual iteration with Next() for reading
- ❌ **No query methods**: Cannot check if file exists without full scan
- ❌ **No path lookup**: Must iterate entire archive to find specific file
- ❌ **Complex recursive archiving**: Manual directory walking required
- ❌ **No path transformation**: Must manually modify headers for renaming
- ❌ **No reset support**: Cannot re-read archive without reopening

**How tar Package Extends archive/tar:**
- ✅ **High-level methods**: List(), Info(), Get(), Has(), Walk() for easy queries
- ✅ **Direct file access**: Get file by path without manual iteration
- ✅ **Automatic directory archiving**: FromPath() with glob filtering
- ✅ **Built-in path transformation**: ReplaceName function for path renaming
- ✅ **Reset capability**: Automatic reset when reader supports seeking
- ✅ **Interface compliance**: Standard Reader/Writer interfaces for consistency

**Internally**, the package uses `archive/tar.Reader` and `archive/tar.Writer` for all tar operations, but adds convenient wrapper methods that handle the sequential iteration and header management automatically.

### Design Philosophy

1. **Extend Standard Library**: Build on `archive/tar` while addressing its limitations with high-level query methods and automatic iteration handling.
2. **Interface Compliance**: Implements `github.com/nabbar/golib/archive/archive/types` interfaces for consistency across different archive formats (tar, zip, etc.).
3. **Simplicity First**: High-level methods (List, Get, Walk) handle common use cases without requiring detailed knowledge of tar format internals or manual iteration.
4. **Flexibility**: Supports advanced scenarios (path filtering, renaming, link preservation) while keeping simple cases simple through intuitive API.
5. **Sequential Efficiency**: Embraces tar's sequential nature with streaming operations and constant memory usage per operation.
6. **Safety**: Proper resource cleanup with Close(), comprehensive error handling with fs.ErrNotExist, and defensive programming against nil values.

### Key Features

- ✅ **Simple API**: Intuitive constructors (NewReader, NewWriter) and methods (List, Get, Add, FromPath)
- ✅ **High-level Operations**: Query methods (List, Info, Has, Walk) and extraction (Get)
- ✅ **Recursive Archiving**: Add entire directory trees with pattern filtering
- ✅ **Link Preservation**: Maintains symbolic and hard links with their targets
- ✅ **Path Transformation**: Custom path renaming during archiving
- ✅ **Reset Support**: Re-read archives when underlying reader supports seeking
- ✅ **Standard Interfaces**: Full `io.ReadCloser` and `io.WriteCloser` compatibility
- ✅ **Zero External Dependencies**: Only standard library + golib packages

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────┐
│              tar Package                       │
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
│  │    archive/tar (std library)         │      │
│  └──────────────────────────────────────┘      │
│                                                │
└────────────────────────────────────────────────┘
```

**Reader Component:**
- Wraps `io.ReadCloser` with `tar.Reader`
- Query methods: List(), Info(), Has()
- Extraction: Get(), Walk()
- Optional Reset() for re-reading

**Writer Component:**
- Wraps `io.WriteCloser` with `tar.Writer`
- Add individual files: Add()
- Recursive archiving: FromPath()
- Handles links, filtering, path transformation

### Data Flow

**Reading:**
1. Open file/stream → NewReader()
2. Query: List(), Info(), Has()
3. Extract: Get() or Walk()
4. Close reader

**Writing:**
1. Create file/stream → NewWriter()
2. Add files: Add() or FromPath()
3. Close writer (finalizes archive)
4. Close underlying stream

### Comparison with Standard Library

| Feature | tar Package | archive/tar (stdlib) |
|---------|-------------|----------------------|
| **API Complexity** | Simple high-level methods | Low-level sequential API |
| **File Lookup** | Info(), Has(), Get() by path | Manual iteration required |
| **Directory Archiving** | FromPath() with filtering | Manual walking required |
| **Path Transformation** | Built-in ReplaceName function | Manual header modification |
| **Reset Support** | Automatic when available | Manual seeking required |
| **Error Handling** | fs.ErrNotExist for missing files | io.EOF for end of archive |

---

## Performance

### Benchmarks

Based on test suite measurements (Go 1.24+ with race detector):

| Operation | Median | Mean | Max | Samples |
|-----------|--------|------|-----|---------|
| **Reader List** | 100µs | 100µs | 800µs | 1000 |
| **Reader Info** | <100µs | 100µs | 600µs | 1000 |
| **Reader Get** | <100µs | 100µs | 600µs | 1000 |
| **Reader Has** | <100µs | 100µs | 1ms | 1000 |
| **Reader Walk** | 100µs | 100µs | 700µs | 1000 |
| **Writer Add (100B)** | <100µs | <100µs | 500µs | 1000 |
| **Writer Add (10KB)** | <100µs | 100µs | 400µs | 100 |
| **Writer Add (1MB)** | 8.3ms | 8.6ms | 9.9ms | 10 |

*Performance measured with race detector enabled. Non-race performance is approximately 2-3x faster.*

### Memory Usage

```
Reader overhead:    ~1KB (struct + tar.Reader)
Writer overhead:    ~1KB (struct + tar.Writer)
Per-file memory:    O(1) - streaming operations
List() memory:      O(n) where n = number of files
```

**Memory Efficiency:**
- Sequential access requires constant memory per operation
- Files are streamed, not fully buffered
- List() stores only paths, not file contents

### Scalability

- **Small archives** (<100 files): Excellent performance
- **Medium archives** (100-10K files): Good performance, consider Walk() over multiple Get()
- **Large archives** (>10K files): Sequential access remains efficient
- **File sizes**: Tested with files up to 1MB, scales linearly

---

## Use Cases

### 1. Backup System

**Problem**: Create backups of directories with pattern filtering.

```go
file, _ := os.Create("backup.tar")
defer file.Close()

writer, _ := tar.NewWriter(file)
defer writer.Close()

// Backup all source files
writer.FromPath("/home/user/project", "*.go", nil)
```

**Real-world**: Configuration backups, code archives, data snapshots.

### 2. Log Aggregation

**Problem**: Collect and archive log files from multiple sources.

```go
writer.FromPath("/var/log/app", "*.log", func(path string) string {
    return strings.TrimPrefix(path, "/var/log/app/")
})
```

**Real-world**: Centralized logging, audit trails, compliance archives.

### 3. Deployment Packages

**Problem**: Package application files for deployment.

```go
// Create deployment archive
writer.FromPath("/build/output", "*", func(path string) string {
    return filepath.Join("app", filepath.Base(path))
})
```

**Real-world**: Docker layers, application packages, software distribution.

### 4. Archive Inspection

**Problem**: Examine archive contents without full extraction.

```go
reader, _ := tar.NewReader(file)
defer reader.Close()

files, _ := reader.List()
for _, path := range files {
    info, _ := reader.Info(path)
    fmt.Printf("%s: %d bytes\n", path, info.Size())
}
```

**Real-world**: Archive validation, content verification, file discovery.

### 5. Selective Extraction

**Problem**: Extract specific files from large archives.

```go
if reader.Has("config.json") {
    rc, _ := reader.Get("config.json")
    defer rc.Close()
    data, _ := io.ReadAll(rc)
    processConfig(data)
}
```

**Real-world**: Configuration extraction, partial restore, targeted file retrieval.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/archive/tar
```

### Reading a Tar Archive

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/archive/tar"
)

func main() {
    file, err := os.Open("archive.tar")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    reader, err := tar.NewReader(file)
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

### Creating a Tar Archive

```go
package main

import (
    "log"
    "os"
    
    "github.com/nabbar/golib/archive/archive/tar"
)

func main() {
    file, err := os.Create("archive.tar")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    writer, err := tar.NewWriter(file)
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

The package includes comprehensive tests with **85.6% code coverage**, **61 test specifications**, and **0 race conditions** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All reader operations (List, Info, Get, Has, Walk)
- ✅ All writer operations (Add, FromPath, Close)
- ✅ Edge cases (empty archives, missing files, large files)
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
    
    reader, err := tar.NewReader(file)
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
    
    writer, err := tar.NewWriter(file)
    if err != nil {
        return err
    }
    defer writer.Close() // Must close before file.Close()
    
    return addFiles(writer)
}
```

**Use Walk for processing all files:**
```go
// ✅ GOOD: Efficient single pass
reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path, link string) bool {
    processFile(path, rc)
    return true
})
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
writer, _ := tar.NewWriter(file)
writer.Add(info, rc, "file.txt", "")
// Missing writer.Close()!
```

**Don't use multiple operations simultaneously:**
```go
// ❌ BAD: Not thread-safe
go reader.Get("file1.txt")
go reader.Get("file2.txt")

// ✅ GOOD: Use separate readers
reader1, _ := tar.NewReader(stream1)
reader2, _ := tar.NewReader(stream2)
go reader1.Get("file1.txt")
go reader2.Get("file2.txt")
```

**Don't assume Reset will work:**
```go
// ❌ BAD: Doesn't check return value
reader.Reset()
files, _ := reader.List()

// ✅ GOOD: Check if reset succeeded
if reader.Reset() {
    files, _ := reader.List()
} else {
    // Reopen archive
}
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
    Reset() bool
}
```

**Methods:**

- **`List()`**: Returns all file paths in the archive
- **`Info(path)`**: Gets file metadata for a specific path
- **`Get(path)`**: Extracts a file as io.ReadCloser
- **`Has(path)`**: Checks if a file exists
- **`Walk(callback)`**: Iterates all files with a callback
- **`Reset()`**: Resets reader to beginning (if supported)

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
  - `reader`: File contents (nil for links)
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
fs.ErrInvalid   // Invalid file type or operation
io.EOF          // End of archive (internal, not returned by high-level methods)
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

- ✅ **85.6% test coverage** (exceeds 80% target)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** per-instance (one goroutine per reader/writer)
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard compliance** with archive/tar format

### Known Limitations

**Architectural Constraints:**

1. **Sequential Access Only**: Tar format requires sequential reading from beginning to end
2. **Reset Requires Seeking**: Reset() only works if underlying reader supports seeking (not network streams)
3. **Hard Links as Symlinks**: Hard links are currently treated as symbolic links
4. **No Special Files**: Device files, pipes, and sockets are not supported
5. **Implicit Directories**: Directory entries are not stored separately (created implicitly from file paths)

**Not Suitable For:**
- Random file access within archives (use zip for random access)
- Streaming archives without seekable storage (reset not available)
- Archives with special file types (devices, pipes)

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Compression Support**: Built-in gzip/bzip2/xz compression wrappers
2. **Streaming Reset**: Buffer mechanism for limited reset on non-seekable streams
3. **Hard Link Support**: Proper hard link preservation (distinct from symlinks)
4. **Progress Callbacks**: Optional callbacks for long-running operations
5. **Parallel Extraction**: Concurrent file extraction for multi-file Get operations

These are **optional improvements** and not required for production use. The current implementation is stable and performant for standard tar operations.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/tar)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, and comprehensive usage examples. Provides detailed explanations of reader/writer operations and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/archive/archive/types](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/types)** - Common archive interfaces implemented by tar package. Provides Reader and Writer interfaces for uniform archive handling across different formats.

- **[github.com/nabbar/golib/archive/archive/zip](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/zip)** - ZIP archive implementation with the same interface. Use zip for random access, tar for sequential streaming.

### Standard Library References

- **[archive/tar](https://pkg.go.dev/archive/tar)** - Standard library tar implementation used internally. The tar package provides a higher-level API wrapping this functionality.

- **[io/fs](https://pkg.go.dev/io/fs)** - Standard filesystem interfaces used for file information. The package uses fs.FileInfo and fs.ErrNotExist for consistency with Go's filesystem APIs.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The tar package follows these conventions for idiomatic Go code.

- **[Tar Format Specification](https://www.gnu.org/software/tar/manual/html_node/Standard.html)** - GNU tar format specification. Understanding the tar format helps explain sequential access requirements and limitations.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/archive/archive/tar`
