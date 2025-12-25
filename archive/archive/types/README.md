# Archive Types Package

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-N%2FA-lightgrey)](TESTING.md)

Common interfaces for archive reading and writing operations, providing a unified abstraction for various archive formats (ZIP, TAR, BZIP2, GZIP, etc.).

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Interface Design](#interface-design)
  - [Data Flow](#data-flow)
- [Performance](#performance)
  - [Interface Overhead](#interface-overhead)
  - [Memory Usage](#memory-usage)
  - [Best Practices](#best-practices-for-performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Reading from Archives](#reading-from-archives)
  - [Writing to Archives](#writing-to-archives)
  - [Walking Through Archives](#walking-through-archives)
  - [Archive Conversion](#archive-conversion)
  - [Selective Extraction](#selective-extraction)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Reader Interface](#reader-interface)
  - [Writer Interface](#writer-interface)
  - [Function Types](#function-types)
  - [Error Conventions](#error-conventions)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **types** package defines common interfaces for archive manipulation, enabling format-agnostic code that works with ZIP, TAR, BZIP2, GZIP, and other archive formats. It provides two core interfaces: `Reader` for extraction and `Writer` for creation.

### Design Philosophy

1. **Format Independence**: Single API for all archive formats
2. **Standard Compliance**: Uses `io.Closer`, `fs.FileInfo`, and Go standard error conventions
3. **Simplicity**: Minimal interface surface with essential operations
4. **Extensibility**: Easy to implement for new archive formats
5. **Error Transparency**: Clear error semantics using `fs.ErrNotExist` and standard errors

### Key Features

- ✅ **Reader Interface**: Extract and list files from archives
- ✅ **Writer Interface**: Create archives and add files
- ✅ **Format Agnostic**: Works with any archive format implementation
- ✅ **Standard Interfaces**: Compatible with `io.Closer`, `io.ReadCloser`, `io.Writer`
- ✅ **Callback Support**: `Walk` method for iterating archive contents
- ✅ **Path Transformation**: `ReplaceName` callback for customizing archive structure
- ✅ **Zero Dependencies**: Only standard library types
- ✅ **Well Documented**: Comprehensive GoDoc with examples

---

## Architecture

### Component Diagram

```
┌──────────────────────────────────────────────┐
│          types.Reader Interface              │
│  ┌────────────────────────────────────────┐  │
│  │ Close() error                          │  │
│  │ List() ([]string, error)               │  │
│  │ Info(path) (fs.FileInfo, error)        │  │
│  │ Get(path) (io.ReadCloser, error)       │  │
│  │ Has(path) bool                         │  │
│  │ Walk(FuncExtract)                      │  │
│  └────────────────────────────────────────┘  │
└──────────┬─────────────────────────────────────┘
           │
    ┌──────┴───┬──────────┬──────────┐
    │          │          │          │
  ZIP.rdr   TAR.rdr   BZIP.rdr   Other
 Readers    Readers   Readers   Readers

┌──────────────────────────────────────────────┐
│          types.Writer Interface              │
│  ┌────────────────────────────────────────┐  │
│  │ Close() error                          │  │
│  │ Add(info, reader, path, link) error    │  │
│  │ FromPath(src, filter, fn) error        │  │
│  └────────────────────────────────────────┘  │
└──────────┬─────────────────────────────────────┘
           │
    ┌──────┴───┬──────────┬──────────┐
    │          │          │          │
  ZIP.wrt   TAR.wrt   BZIP.wrt   Other
 Writers    Writers   Writers   Writers
```

### Interface Design

**Reader Interface** - Provides read-only access to archive contents:
- `List()`: Enumerate all files in the archive
- `Info(path)`: Get file metadata (size, permissions, timestamps)
- `Get(path)`: Extract a specific file as `io.ReadCloser`
- `Has(path)`: Check if a file exists (efficient)
- `Walk(fn)`: Iterate through all files with a callback
- `Close()`: Release resources

**Writer Interface** - Provides archive creation capabilities:
- `Add(info, reader, path, link)`: Add a single file with metadata
- `FromPath(src, filter, fn)`: Recursively add files from filesystem
- `Close()`: Finalize and close the archive

**Function Types**:
- `FuncExtract`: Callback for `Walk()` method - `func(fs.FileInfo, io.ReadCloser, string, string) bool`
- `ReplaceName`: Path transformation for `FromPath()` - `func(string) string`

### Data Flow

**Reading from Archive:**
```
Open Archive → List Files → Check Existence → Get File Info → Extract Content → Close
     │              │              │                │                │           │
  NewReader()   List()         Has()           Info()           Get()      Close()
```

**Writing to Archive:**
```
Create Archive → Add Single Files → Add Directory Tree → Finalize Archive
      │                  │                  │                    │
  NewWriter()         Add()            FromPath()            Close()
```

---

## Performance

### Interface Overhead

**Interface Call Cost**: Minimal (single virtual dispatch)
- No heap allocation for interface calls
- Inlined by compiler when possible
- Negligible overhead compared to I/O operations

**Memory Overhead**:
- Interface values: 16 bytes (pointer + type info)
- No additional allocation for method calls
- Same performance as direct struct method calls

### Memory Usage

**Per-Operation Memory**:
```
Reader.List():       O(n) where n = number of files
Reader.Info(path):   O(1) file info struct (~200 bytes)
Reader.Get(path):    O(1) reader handle (~50 bytes)
Writer.Add():        O(1) per call (streaming)
```

**Recommended Patterns**:
- Use `Walk()` instead of `List() + Get()` for full extraction (single pass)
- Cache `List()` results if querying multiple files
- Close readers immediately after use to release file handles

### Best Practices for Performance

1. **Minimize List() Calls**: Cache results when checking multiple files
2. **Use Has() Before Get()**: Avoid opening non-existent files
3. **Stream Large Files**: Don't load entire files into memory
4. **Close Resources**: Prevent file descriptor leaks

---

## Use Cases

### 1. Format-Agnostic Archive Processing

**Problem**: Process archives without knowing their format in advance.

```go
func ProcessArchive(r types.Reader) error {
    files, err := r.List()
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if r.Has(file) {
            processFile(file)
        }
    }
    return nil
}
```

**Real-world**: Backup systems, file managers, archive utilities.

### 2. Archive Format Conversion

**Problem**: Convert between different archive formats.

```go
func ConvertArchive(src types.Reader, dst types.Writer) error {
    return src.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
        if r != nil {
            defer r.Close()
            if err := dst.Add(info, r, path, link); err != nil {
                return false
            }
        }
        return true
    })
}
```

**Real-world**: Migration tools, format standardization.

### 3. Selective Extraction

**Problem**: Extract only specific files from an archive.

```go
func ExtractMatching(r types.Reader, pattern string, outputDir string) error {
    return r.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
        if matched, _ := filepath.Match(pattern, path); matched {
            extractToFile(outputDir, path, rc)
        }
        if rc != nil {
            rc.Close()
        }
        return true
    })
}
```

**Real-world**: Deployment tools, configuration extraction.

### 4. Archive Inspection

**Problem**: Analyze archive contents without extraction.

```go
func InspectArchive(r types.Reader) (int, int64, error) {
    files, err := r.List()
    if err != nil {
        return 0, 0, err
    }
    
    var totalSize int64
    for _, path := range files {
        info, err := r.Info(path)
        if err == nil {
            totalSize += info.Size()
        }
    }
    
    return len(files), totalSize, nil
}
```

**Real-world**: File browsers, disk usage analyzers.

### 5. Batch Archive Creation

**Problem**: Create multiple archives from directory trees.

```go
func CreateBackupArchive(w types.Writer, sourceDir string) error {
    return w.FromPath(sourceDir, "", func(source string) string {
        // Add timestamp prefix
        return time.Now().Format("2006-01-02") + "/" + source
    })
}
```

**Real-world**: Backup tools, deployment systems.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/archive/archive/types
```

### Reading from Archives

```go
package main

import (
    "fmt"
    "io"
    
    "github.com/nabbar/golib/archive/archive/types"
    // Import format-specific implementation:
    // "github.com/nabbar/golib/archive/archive/zip"
)

func main() {
    // Open archive (using format-specific constructor)
    // reader, _ := zip.NewReader(file)
    var reader types.Reader // = actual implementation
    defer reader.Close()
    
    // List all files
    files, err := reader.List()
    if err != nil {
        panic(err)
    }
    
    for _, path := range files {
        fmt.Printf("File: %s\n", path)
        
        // Get file info
        info, _ := reader.Info(path)
        fmt.Printf("  Size: %d bytes\n", info.Size())
        
        // Extract file
        rc, _ := reader.Get(path)
        defer rc.Close()
        
        // Read content
        content, _ := io.ReadAll(rc)
        fmt.Printf("  Content: %s\n", content[:min(50, len(content))])
    }
}
```

### Writing to Archives

```go
package main

import (
    "os"
    
    "github.com/nabbar/golib/archive/archive/types"
    // Import format-specific implementation:
    // "github.com/nabbar/golib/archive/archive/zip"
)

func main() {
    // Create archive (using format-specific constructor)
    // writer, _ := zip.NewWriter(file)
    var writer types.Writer // = actual implementation
    defer writer.Close()
    
    // Add single file
    info, _ := os.Stat("myfile.txt")
    file, _ := os.Open("myfile.txt")
    defer file.Close()
    
    err := writer.Add(info, file, "", "")
    if err != nil {
        panic(err)
    }
    
    // Add directory recursively with filter
    err = writer.FromPath("/path/to/dir", "*.txt", nil)
    if err != nil {
        panic(err)
    }
}
```

### Walking Through Archives

```go
package main

import (
    "fmt"
    "io"
    "io/fs"
    
    "github.com/nabbar/golib/archive/archive/types"
)

func main() {
    var reader types.Reader // = actual implementation
    defer reader.Close()
    
    reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
        if r != nil {
            defer r.Close()
        }
        
        fmt.Printf("File: %s, Size: %d\n", path, info.Size())
        
        if link != "" {
            fmt.Printf("  -> Link target: %s\n", link)
        }
        
        // Return true to continue, false to stop
        return true
    })
}
```

### Archive Conversion

```go
package main

import (
    "io/fs"
    "io"
    
    "github.com/nabbar/golib/archive/archive/types"
)

func ConvertArchive(src types.Reader, dst types.Writer) error {
    defer src.Close()
    defer dst.Close()
    
    return src.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
        if r != nil {
            defer r.Close()
            
            if err := dst.Add(info, r, path, link); err != nil {
                return false // Stop on error
            }
        }
        return true // Continue
    })
}
```

### Selective Extraction

```go
package main

import (
    "io"
    "io/fs"
    "path/filepath"
    
    "github.com/nabbar/golib/archive/archive/types"
)

func ExtractTextFiles(reader types.Reader) error {
    defer reader.Close()
    
    return reader.Walk(func(info fs.FileInfo, rc io.ReadCloser, path string, link string) bool {
        if rc != nil {
            defer rc.Close()
        }
        
        // Extract only .txt files
        if filepath.Ext(path) == ".txt" {
            // Process or save the file...
        }
        
        return true
    })
}
```

---

## Best Practices

### ✅ DO

**Always close readers and writers:**
```go
// ✅ GOOD: Proper resource management
reader, err := format.NewReader(file)
if err != nil {
    return err
}
defer reader.Close()
```

**Close extracted files:**
```go
// ✅ GOOD: Release file handles
rc, err := reader.Get("file.txt")
if err != nil {
    return err
}
defer rc.Close()

content, _ := io.ReadAll(rc)
```

**Check file existence before extraction:**
```go
// ✅ GOOD: Avoid fs.ErrNotExist
if reader.Has("config.json") {
    rc, _ := reader.Get("config.json")
    defer rc.Close()
    // Process file...
}
```

**Handle Walk() errors gracefully:**
```go
// ✅ GOOD: Check for nil reader
reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
    if r == nil {
        log.Printf("Cannot open %s", path)
        return true // Continue despite error
    }
    defer r.Close()
    // Process file...
    return true
})
```

**Use path transformation wisely:**
```go
// ✅ GOOD: Organized archive structure
writer.FromPath(srcDir, "*.txt", func(src string) string {
    return "backup/" + filepath.Base(src)
})
```

### ❌ DON'T

**Don't forget to close resources:**
```go
// ❌ BAD: Resource leak
reader, _ := format.NewReader(file)
// Missing: defer reader.Close()

rc, _ := reader.Get("file.txt")
// Missing: defer rc.Close()
```

**Don't ignore errors:**
```go
// ❌ BAD: Ignoring errors
files, _ := reader.List()  // Might fail silently
rc, _ := reader.Get(path)  // File might not exist

// ✅ GOOD: Check errors
files, err := reader.List()
if err != nil {
    return err
}
```

**Don't load entire archives into memory:**
```go
// ❌ BAD: Memory exhaustion on large files
files, _ := reader.List()
for _, path := range files {
    rc, _ := reader.Get(path)
    data, _ := io.ReadAll(rc) // Entire file in RAM!
    rc.Close()
}

// ✅ GOOD: Stream processing
reader.Walk(func(info fs.FileInfo, r io.ReadCloser, path string, link string) bool {
    if r != nil {
        defer r.Close()
        // Process stream without loading all data
        processStream(r)
    }
    return true
})
```

**Don't assume thread safety:**
```go
// ❌ BAD: Concurrent access without synchronization
go func() { reader.Get("file1.txt") }()
go func() { reader.Get("file2.txt") }() // Race condition!

// ✅ GOOD: Check implementation documentation
// Most implementations are NOT thread-safe
```

### Testing

The package includes a comprehensive test suite with **23 test specifications** using BDD methodology (Ginkgo v2 + Gomega). Tests validate interface contracts, mock implementations, and common usage patterns.

**Key test coverage:**
- ✅ Interface compliance verification
- ✅ Mock reader and writer implementations
- ✅ Error handling (fs.ErrNotExist)
- ✅ Walk callback functionality
- ✅ Path transformation callbacks
- ✅ Integration scenarios

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

---

## API Reference

### Reader Interface

```go
type Reader interface {
    io.Closer
    
    // List returns all file paths in the archive
    List() ([]string, error)
    
    // Info returns file metadata for a specific path
    Info(string) (fs.FileInfo, error)
    
    // Get retrieves a file as io.ReadCloser
    Get(string) (io.ReadCloser, error)
    
    // Has checks if a file exists in the archive
    Has(string) bool
    
    // Walk iterates through all files with a callback
    Walk(FuncExtract)
}
```

**Methods:**

- **`List() ([]string, error)`**: Returns all file paths in the archive. Empty slice if archive is empty. Returns error if archive cannot be read.

- **`Info(path) (fs.FileInfo, error)`**: Returns file metadata (size, permissions, mod time) for the specified path. Returns `fs.ErrNotExist` if file not found.

- **`Get(path) (io.ReadCloser, error)`**: Opens a file for reading. Caller must close the returned reader. Returns `fs.ErrNotExist` if file not found.

- **`Has(path) bool`**: Fast existence check. Should be more efficient than calling `Info()` or `Get()` and checking for errors.

- **`Walk(fn FuncExtract)`**: Iterates through all files, calling the callback for each. Stops if callback returns false. Callback may receive nil reader on errors.

### Writer Interface

```go
type Writer interface {
    io.Closer
    
    // Add adds a single file to the archive
    Add(info fs.FileInfo, reader io.ReadCloser, forcePath string, linkTarget string) error
    
    // FromPath recursively adds files from a directory
    FromPath(source string, filter string, fn ReplaceName) error
}
```

**Methods:**

- **`Add(info, reader, forcePath, linkTarget) error`**: Adds a file to the archive.
  - `info`: File metadata (size, permissions, timestamps)
  - `reader`: File content stream (may be nil for directories)
  - `forcePath`: Override the file path in archive (empty = use info.Name())
  - `linkTarget`: Symlink target (format-dependent, empty for regular files)

- **`FromPath(source, filter, fn) error`**: Recursively adds files from filesystem.
  - `source`: Directory or file path to add
  - `filter`: Glob pattern for file selection (e.g., "*.txt", empty = all files)
  - `fn`: Optional path transformation callback (nil = preserve structure)

### Function Types

#### FuncExtract

```go
type FuncExtract func(fs.FileInfo, io.ReadCloser, string, string) bool
```

Callback for `Walk()` method. Parameters:
- `fs.FileInfo`: File metadata
- `io.ReadCloser`: Content stream (may be nil on error)
- `string`: File path in archive
- `string`: Symlink target (empty for regular files)
- Returns: `true` to continue, `false` to stop

#### ReplaceName

```go
type ReplaceName func(string) string
```

Callback for `FromPath()` method. Parameters:
- `string`: Original filesystem path
- Returns: Transformed path for archive

### Error Conventions

**Standard Errors**:
- `fs.ErrNotExist`: File not found in archive (used by `Info()`, `Get()`)
- `fs.ErrInvalid`: Invalid operation or parameter
- `io.EOF`: End of archive reached (during iteration)

**Best Practices**:
- Use `errors.Is(err, fs.ErrNotExist)` for file existence checks
- `Close()` should be idempotent (safe to call multiple times)
- `Walk()` should handle individual file errors internally
- Implementations should document thread safety

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve test coverage (target: >80%)
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
   - Ensure interface contract compliance
   - Test mock implementations thoroughly

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

- ✅ **100% interface coverage** (all methods tested)
- ✅ **23 test specifications** covering interface compliance
- ✅ **Zero race conditions** (N/A for interface-only package)
- ✅ **Standard compliant** (uses Go standard library types)
- ✅ **Well documented** (comprehensive GoDoc + examples)

### Limitations

**Architectural Constraints:**

1. **No Format Detection**: Applications must know archive format beforehand
2. **No Archive Modification**: No support for updating archives in-place (append/delete)
3. **No Built-in Encryption**: Encryption must be handled by implementations
4. **No Progress Reporting**: Implementations must provide their own progress callbacks
5. **Symlink Support Varies**: Not all formats support symlinks (ZIP, TAR differ)

**Not Covered by Interfaces:**

- Multi-volume archives
- Archive streaming without intermediate files
- Archive integrity verification
- Compression level configuration
- Format-specific options (ZIP extra fields, TAR extensions)

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Extended Interfaces**: Optional interfaces for format-specific features (compression levels, extra metadata)
2. **Progress Callbacks**: Standard interface for progress reporting during operations
3. **Streaming Support**: Interfaces for streaming archives without temporary files
4. **Verification Interface**: Standard methods for integrity checking and validation
5. **Metadata Extensions**: Support for extended attributes, ACLs, and platform-specific metadata

These are **optional improvements** and not required for production use. The current implementation is stable and widely compatible.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/types)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the interfaces and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, implementation guidelines, and best practices. Provides detailed explanations of interface contracts and common patterns.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, interface compliance tests, and mock implementations. Includes guidelines for testing implementations of these interfaces.

### Related golib Packages

- **[github.com/nabbar/golib/archive/archive/zip](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/zip)** - ZIP format implementation of Reader and Writer interfaces. Supports standard ZIP archives with compression.

- **[github.com/nabbar/golib/archive/archive/tar](https://pkg.go.dev/github.com/nabbar/golib/archive/archive/tar)** - TAR format implementation with optional compression (GZIP, BZIP2). Supports POSIX TAR and GNU TAR extensions.

- **[github.com/nabbar/golib/archive/compress](https://pkg.go.dev/github.com/nabbar/golib/archive/compress)** - Compression algorithms (GZIP, BZIP2, XZ) that can be combined with archive formats.

### Standard Library References

- **[io/fs](https://pkg.go.dev/io/fs)** - File system interfaces used by the Reader interface. Provides `fs.FileInfo`, `fs.ErrNotExist`, and other standard types.

- **[io](https://pkg.go.dev/io)** - Standard I/O interfaces implemented by the package. The interfaces extend `io.Closer` and use `io.ReadCloser` for file streams.

- **[path/filepath](https://pkg.go.dev/path/filepath)** - Path manipulation functions useful for archive path handling, pattern matching, and directory traversal.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The types package follows these conventions for idiomatic Go code.

- **[Go Proverbs](https://go-proverbs.github.io/)** - Go programming wisdom including "The bigger the interface, the weaker the abstraction." The types package defines minimal, focused interfaces.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the types package. Check existing issues before creating new ones.

- **[Contributing Guide](../../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/archive/archive/types`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
