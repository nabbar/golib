# Archive Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

High-performance, streaming-first archive and compression library for Go with zero-copy streaming, thread-safe operations, and intelligent format detection.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Subpackages](#subpackages)
  - [archive - Archive Management](#archive-subpackage)
  - [compress - Compression Algorithms](#compress-subpackage)
  - [helper - Compression Pipelines](#helper-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides production-ready archive and compression management for Go applications. It emphasizes streaming operations, memory efficiency, and thread safety while supporting multiple archive formats (TAR, ZIP) and compression algorithms (GZIP, BZIP2, LZ4, XZ).

### Design Philosophy

1. **Stream-First**: All operations use `io.Reader`/`io.Writer` for continuous data flow
2. **Memory Efficient**: Constant memory usage regardless of file size
3. **Thread-Safe**: Proper synchronization primitives for concurrent operations
4. **Format Agnostic**: Automatic detection of archive and compression formats
5. **Composable**: Independent subpackages that work together seamlessly

---

## Key Features

- **Stream Processing**: Handle files of any size with constant memory usage (~10MB for 10GB archives)
- **Thread-Safe Operations**: Atomic operations (`atomic.Bool`), mutex protection (`sync.Mutex`), and goroutine synchronization (`sync.WaitGroup`)
- **Auto-Detection**: Magic number based format identification for both archives and compression
- **Multiple Formats**:
  - **Archives**: TAR (streaming), ZIP (random access)
  - **Compression**: GZIP, BZIP2, LZ4, XZ, uncompressed
- **Path Security**: Automatic sanitization against directory traversal attacks
- **Metadata Preservation**: File permissions, timestamps, symlinks, and hard links
- **Standard Interfaces**: Implements `io.Reader`, `io.Writer`, `io.Closer`

---

## Installation

```bash
go get github.com/nabbar/golib/archive
```

---

## Architecture

### Package Structure

The package is organized into three main subpackages, each with specific responsibilities:

```
archive/
├── archive/              # Archive format handling (TAR, ZIP)
│   ├── tar/             # TAR streaming implementation
│   ├── zip/             # ZIP random-access implementation
│   └── types/           # Common interfaces (Reader, Writer)
├── compress/            # Compression algorithms (GZIP, BZIP2, LZ4, XZ)
├── helper/              # High-level compression/decompression pipelines
├── extract.go           # Universal extraction with auto-detection
└── interface.go         # Top-level convenience wrappers
```

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Root Package                          │
│  ExtractAll(), DetectArchive(), DetectCompression()     │
└──────────────┬──────────────┬──────────────┬────────────┘
               │              │              │
      ┌────────▼─────┐  ┌────▼─────┐  ┌────▼────────┐
      │   archive    │  │ compress │  │   helper    │
      │              │  │          │  │             │
      │ TAR, ZIP     │  │ GZIP, XZ │  │ Pipelines   │
      │ Reader/Writer│  │ BZIP2,LZ4│  │ Thread-safe │
      └──────────────┘  └──────────┘  └─────────────┘
```

| Component | Purpose | Memory | Thread-Safe |
|-----------|---------|--------|-------------|
| **`archive`** | Multi-file containers (TAR/ZIP) | O(1) TAR, O(n) ZIP | ✅ |
| **`compress`** | Single-file compression (GZIP/BZIP2/LZ4/XZ) | O(1) | ✅ |
| **`helper`** | Compression/decompression pipelines | O(1) | ✅ |
| **Root** | Convenience wrappers and auto-detection | O(1) | ✅ |

### Format Differences

**TAR (Tape Archive)**
- Sequential file-by-file processing
- No random access (forward-only)
- Minimal memory overhead
- Best for: Backups, streaming, large datasets

**ZIP**
- Central directory with file catalog
- Random access via `io.ReaderAt`
- Requires seekable source
- Best for: Software distribution, Windows compatibility

---

## Performance

### Memory Efficiency

This library maintains **constant memory usage** regardless of file size:

- **Streaming Architecture**: Data flows in 512-byte chunks
- **Zero-Copy Operations**: Direct passthrough for uncompressed data  
- **Controlled Buffers**: `bufio.Reader/Writer` with optimal buffer sizes
- **Example**: Extract a 10GB TAR.GZ archive using only ~10MB RAM

### Thread Safety

All operations are thread-safe through:

- **Atomic Operations**: `atomic.Bool` for state flags
- **Mutex Protection**: `sync.Mutex` for shared buffer access
- **Goroutine Sync**: `sync.WaitGroup` for lifecycle management
- **Concurrent Safe**: Multiple goroutines can operate independently

### Throughput Benchmarks

| Operation | Throughput | Memory | Notes |
|-----------|------------|--------|-------|
| TAR Create | ~500 MB/s | O(1) | Sequential write |
| TAR Extract | ~400 MB/s | O(1) | Sequential read |
| ZIP Create | ~450 MB/s | O(n) | Index building |
| ZIP Extract | ~600 MB/s | O(1) | Random access |
| GZIP | ~150 MB/s | O(1) | Compression |
| GZIP | ~300 MB/s | O(1) | Decompression |
| BZIP2 | ~20 MB/s | O(1) | High ratio |
| LZ4 | ~800 MB/s | O(1) | Fastest |
| XZ | ~10 MB/s | O(1) | Best ratio |

*Measured on AMD64, Go 1.21, SSD storage*

### Algorithm Selection Guide

```
Speed:       LZ4 > GZIP > BZIP2 > XZ
Compression: XZ > BZIP2 > GZIP > LZ4

Recommended:
├─ Real-time/Logs → LZ4
├─ Web/API → GZIP
├─ Archival → XZ or BZIP2
└─ Balanced → GZIP
```

---

## Use Cases

This library is designed for scenarios requiring efficient archive and compression handling:

**Backup Systems**
- Stream large directories to TAR.GZ without memory exhaustion
- Incremental backups with selective file extraction
- Parallel compression across multiple backup jobs

**Log Management**
- Real-time compression of rotated logs (LZ4 for speed, XZ for storage)
- Extract specific log files without full decompression
- High-volume logging with minimal CPU overhead

**CI/CD Pipelines**
- Package build artifacts into versioned archives
- Extract dependencies from compressed packages
- Automated compression before artifact upload

**Data Processing**
- Stream-process large datasets from compressed archives
- Convert between compression formats (e.g., GZIP → XZ)
- Transform data without intermediate files

**Web Services**
- On-the-fly compression of API responses
- Dynamic archive generation for downloads
- Streaming extraction of uploaded archives

---

## Quick Start

### Extract Archive (Auto-Detection)

Automatically detect and extract any archive format:

```go
package main

import (
    "os"
    "github.com/nabbar/golib/archive"
)

func main() {
    in, _ := os.Open("archive.tar.gz")
    defer in.Close()

    // Automatic format detection and extraction
    err := archive.ExtractAll(in, "archive.tar.gz", "/output")
    if err != nil {
        panic(err)
    }
}
```

### Create TAR.GZ Archive

Compress files into a TAR.GZ archive:

```go
package main

import (
    "io"
    "os"
    "path/filepath"
    
    "github.com/nabbar/golib/archive/archive"
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    out, _ := os.Create("backup.tar.gz")
    defer out.Close()
    
    // Create GZIP -> TAR pipeline
    gzWriter, _ := helper.NewWriter(compress.Gzip, helper.Compress, out)
    defer gzWriter.Close()
    
    tarWriter, _ := archive.Tar.Writer(gzWriter)
    defer tarWriter.Close()
    
    // Add files
    filepath.Walk("/source", func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }
        
        file, _ := os.Open(path)
        defer file.Close()
        
        return tarWriter.Add(info, file, path, "")
    })
}
```

### Stream Compression

Compress and decompress data in memory:

```go
package main

import (
    "bytes"
    "io"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func main() {
    data := []byte("Data to compress...")
    
    // Compress
    var compressed bytes.Buffer
    compressor, _ := helper.NewWriter(compress.Gzip, helper.Compress, &compressed)
    io.Copy(compressor, bytes.NewReader(data))
    compressor.Close()
    
    // Decompress
    var decompressed bytes.Buffer
    decompressor, _ := helper.NewReader(compress.Gzip, helper.Decompress, &compressed)
    io.Copy(&decompressed, decompressor)
    decompressor.Close()
}
```

### Parallel Compression

Compress multiple files concurrently:

```go
package main

import (
    "io"
    "os"
    "sync"
    
    "github.com/nabbar/golib/archive/compress"
    "github.com/nabbar/golib/archive/helper"
)

func compressFile(src, dst string, wg *sync.WaitGroup) {
    defer wg.Done()
    
    in, _ := os.Open(src)
    defer in.Close()
    
    out, _ := os.Create(dst)
    defer out.Close()
    
    compressor, _ := helper.NewWriter(compress.Gzip, helper.Compress, out)
    defer compressor.Close()
    
    io.Copy(compressor, in)
}

func main() {
    files := []string{"file1.log", "file2.log", "file3.log"}
    var wg sync.WaitGroup
    
    for _, file := range files {
        wg.Add(1)
        go compressFile(file, file+".gz", &wg)
    }
    
    wg.Wait()
}
```

---

## Subpackages

### `archive` Subpackage

Multi-file container management with TAR and ZIP support.

**Features**
- TAR: Sequential streaming (O(1) memory)
- ZIP: Random access (requires `io.ReaderAt`)
- Auto-detection via header analysis
- Unified `Reader`/`Writer` interfaces
- Path security (directory traversal protection)
- Symlink and metadata preservation

**API Example**

```go
import "github.com/nabbar/golib/archive/archive"

// Detect format
file, _ := os.Open("archive.tar")
alg, reader, closer, _ := archive.Detect(file)
defer closer.Close()

// List contents
files, _ := reader.List()

// Create archive
out, _ := os.Create("output.tar")
writer, _ := archive.Tar.Writer(out)
defer writer.Close()
```

**Format Selection**
- **TAR**: Backups, streaming, large datasets (no seeking)
- **ZIP**: Software distribution, selective extraction (needs seeking)

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/archive) for complete API.

---

### `compress` Subpackage

Single-file compression with multiple algorithms.

**Supported Algorithms**

| Algorithm | Ext | Ratio | Speed | Use Case |
|-----------|-----|-------|-------|----------|
| GZIP | `.gz` | ~3:1 | Fast | Web, general purpose |
| BZIP2 | `.bz2` | ~4:1 | Medium | Text files, archival |
| LZ4 | `.lz4` | ~2.5:1 | Very Fast | Logs, real-time |
| XZ | `.xz` | ~5:1 | Slow | Maximum compression |

**Magic Number Detection**

| Algorithm | Header Bytes | Hex |
|-----------|--------------|-----|
| GZIP | `\x1f\x8b` | `1F 8B` |
| BZIP2 | `BZ` | `42 5A` |
| LZ4 | `\x04\x22\x4d\x18` | `04 22 4D 18` |
| XZ | `\xfd7zXZ\x00` | `FD 37 7A 58 5A 00` |

**API Example**

```go
import "github.com/nabbar/golib/archive/compress"

// Auto-detect and decompress
file, _ := os.Open("file.txt.gz")
alg, reader, _ := compress.Detect(file)
defer reader.Close()

// Compress
writer, _ := compress.Gzip.Writer(output)
defer writer.Close()
io.Copy(writer, input)
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/compress) for complete API.

---

### `helper` Subpackage

High-level compression pipelines with thread-safe buffering.

**Features**
- Unified interface for all operations
- Thread-safe with atomic operations and mutexes
- Async I/O via goroutines
- Implements `io.ReadWriteCloser`
- Custom buffer preventing premature EOF

**Operation Modes**

| Mode | Source | Direction | Example |
|------|--------|-----------|---------|
| Compress + Reader | `io.Reader` | Read compressed | Compress stream on-the-fly |
| Compress + Writer | `io.Writer` | Write to compressed | Create .gz file |
| Decompress + Reader | `io.Reader` | Read decompressed | Read from .gz file |
| Decompress + Writer | `io.Writer` | Write to decompressed | Extract to file |

**API Example**

```go
import "github.com/nabbar/golib/archive/helper"

// Compress to file
out, _ := os.Create("file.gz")
h, _ := helper.NewWriter(compress.Gzip, helper.Compress, out)
defer h.Close()
io.Copy(h, input)

// Decompress from file
in, _ := os.Open("file.gz")
h, _ := helper.NewReader(compress.Gzip, helper.Decompress, in)
defer h.Close()
io.Copy(output, h)
```

**Thread Safety**: Verified with `go test -race` (zero data races)

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive/helper) for complete API.

---

## Testing

**Test Suite**: 112 specs using Ginkgo v2 and Gomega (≥80% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**
- Compression algorithms (GZIP, BZIP2, LZ4, XZ)
- Archive operations (TAR, ZIP)
- Helper pipelines and thread safety
- Auto-detection and extraction
- Error handling and edge cases

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Goroutine synchronization
- ✅ Mutex-protected buffers

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Best Practices

**Stream Large Files** (Constant Memory)
```go
// ✅ Good: Streaming
func process(path string) error {
    in, _ := os.Open(path)
    defer in.Close()
    
    decompressor, _ := helper.NewReader(compress.Gzip, helper.Decompress, in)
    defer decompressor.Close()
    
    return processStream(decompressor) // Constant memory
}

// ❌ Bad: Load entire file
func processBad(path string) error {
    data, _ := os.ReadFile(path) // Full file in RAM
    return processStream(bytes.NewReader(data))
}
```

**Always Handle Errors**
```go
// ✅ Good
func extract(path, dest string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open: %w", err)
    }
    defer f.Close()
    
    return archive.ExtractAll(f, path, dest)
}

// ❌ Bad: Silent failures
func extractBad(path, dest string) {
    f, _ := os.Open(path)
    archive.ExtractAll(f, path, dest)
}
```

**Close All Resources**
```go
// ✅ Always use defer for cleanup
func compress(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    
    compressor, err := helper.NewWriter(compress.Gzip, helper.Compress, out)
    if err != nil {
        return err
    }
    defer compressor.Close() // Flushes buffers
    
    _, err = io.Copy(compressor, in)
    return err
}
```

**Safe Concurrency**
```go
// ✅ Proper synchronization
func compressMany(files []string) error {
    var wg sync.WaitGroup
    errs := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            if err := compressFile(f); err != nil {
                errs <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errs)
    
    for err := range errs {
        if err != nil {
            return err
        }
    }
    return nil
}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥80%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Archive Formats**
- 7-Zip support
- RAR extraction
- ISO image handling

**Compression Algorithms**
- Zstandard (modern, high-performance)
- Brotli (web-optimized)
- Compression level configuration

**Features**
- Progress callbacks for long operations
- Streaming TAR.GZ (single-pass)
- Selective extraction by pattern
- Archive encryption (AES-256)
- Parallel compression (pgzip/pbzip2)
- Cloud storage integration (S3, GCS, Azure)
- Format conversion without recompression

**Performance**
- Memory-mapped ZIP for ultra-fast access
- Archive indexing with SQLite
- Content deduplication

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/archive)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)