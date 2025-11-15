# MaxStdio Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue)]()
[![CGO](https://img.shields.io/badge/CGO-Required-orange)]()

Windows-specific package for managing the maximum number of simultaneously open file handles (stdio streams) in a process.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Platform Requirements](#platform-requirements)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Build Instructions](#build-instructions)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This package provides Go bindings to the Windows C Runtime (CRT) functions `_getmaxstdio()` and `_setmaxstdio()`, allowing Go applications to query and modify the maximum number of simultaneously open file descriptors.

### Design Philosophy

1. **Platform-Specific**: Leverages native Windows CRT for optimal integration
2. **Minimal Overhead**: Thin CGO wrapper with negligible performance cost
3. **Type-Safe**: Go-friendly API wrapping C functions
4. **No External Dependencies**: Only requires Windows CRT
5. **Build-Time Safety**: Conditional compilation prevents cross-platform issues

---

## Key Features

- **Get Maximum**: Query the current process limit for open files
- **Set Maximum**: Dynamically adjust the limit at runtime
- **Windows Native**: Direct access to Windows CRT functions
- **Type-Safe**: Integer-based API with clear semantics
- **CGO Integration**: Properly configured CGO directives

---

## Platform Requirements

### Mandatory Requirements

- **Operating System**: Windows (any version with CRT support)
- **CGO**: Must be enabled (`CGO_ENABLED=1`)
- **C Compiler**: One of:
  - MinGW-w64 (recommended)
  - Microsoft Visual C++ (MSVC)
  - TDM-GCC

### Build Constraints

The package uses conditional compilation:

```go
//go:build windows && cgo
// +build windows,cgo
```

**This ensures**:
- Package only compiles on Windows
- CGO must be enabled
- Non-Windows platforms are safely excluded

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/maxstdio
```

**Note**: Installation alone is not sufficient. You must have CGO and a C compiler configured (see [Build Instructions](#build-instructions)).

---

## Architecture

### Package Structure

```
maxstdio/
├── maxstdio.go     # Go API and CGO interface
├── maxstdio.c      # C wrapper functions
├── maxstdio.h      # C header declarations
└── maxstdio.o      # Compiled C object (generated)
```

### Component Diagram

```
┌───────────────────────────────────────┐
│     Go Application                    │
└──────────────┬────────────────────────┘
               │
    ┌──────────▼──────────┐
    │  maxstdio.go        │
    │  Go Functions:      │
    │  • GetMaxStdio()    │
    │  • SetMaxStdio()    │
    └──────────┬──────────┘
               │ CGO
    ┌──────────▼──────────┐
    │  maxstdio.c         │
    │  C Wrappers:        │
    │  • CGetMaxSTDIO()   │
    │  • CSetMaxSTDIO()   │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Windows CRT        │
    │  • _getmaxstdio()   │
    │  • _setmaxstdio()   │
    └─────────────────────┘
```

### Call Flow

```
Go: maxstdio.GetMaxStdio()
       ↓
1. Go function call
       ↓
2. CGO transition (Go → C)
       ↓
3. C wrapper: CGetMaxSTDIO()
       ↓
4. Windows CRT: _getmaxstdio()
       ↓
5. Return value (int)
       ↓
6. CGO transition (C → Go)
       ↓
7. Return to caller
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/ioutils/maxstdio"
)

func main() {
    // Get current limit
    current := maxstdio.GetMaxStdio()
    fmt.Printf("Current limit: %d files\n", current)
    
    // Increase limit to 2048
    old := maxstdio.SetMaxStdio(2048)
    fmt.Printf("Changed from %d to %d\n", old, 2048)
    
    // Verify new limit
    actual := maxstdio.GetMaxStdio()
    fmt.Printf("Actual limit: %d\n", actual)
}
```

### With Error Checking

```go
func configureFileLimit() error {
    required := 2048
    current := maxstdio.GetMaxStdio()
    
    if current >= required {
        fmt.Printf("Limit already sufficient: %d\n", current)
        return nil
    }
    
    old := maxstdio.SetMaxStdio(required)
    actual := maxstdio.GetMaxStdio()
    
    if actual < required {
        return fmt.Errorf("could not set limit to %d, actual: %d", 
            required, actual)
    }
    
    fmt.Printf("Increased limit from %d to %d\n", old, actual)
    return nil
}
```

---

## Use Cases

### High-Concurrency Server

Increase file handle limit for servers handling many connections:

```go
package main

import (
    "log"
    "net/http"
    "github.com/nabbar/golib/ioutils/maxstdio"
)

func main() {
    // Increase limit for high-concurrency server
    old := maxstdio.SetMaxStdio(4096)
    log.Printf("File handle limit: %d → %d", 
        old, maxstdio.GetMaxStdio())
    
    // Start server
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Each connection may use multiple file handles
    w.Write([]byte("OK"))
}
```

**Why**: Web servers may exhaust the default 512 file limit with many concurrent connections.

### Build System / File Processor

Process many files concurrently:

```go
func processManyFiles(files []string) error {
    // Increase limit for batch processing
    maxstdio.SetMaxStdio(len(files) + 100) // Extra buffer
    
    // Open all files
    handles := make([]*os.File, 0, len(files))
    defer closeAll(handles)
    
    for _, path := range files {
        f, err := os.Open(path)
        if err != nil {
            return fmt.Errorf("failed to open %s: %w", path, err)
        }
        handles = append(handles, f)
    }
    
    // Process all files...
    return nil
}
```

**Why**: Build systems often need to read from hundreds of files simultaneously.

### Database Connection Pool

Manage large connection pools:

```go
func initDB() *sql.DB {
    // Increase limit for database connection pool
    maxstdio.SetMaxStdio(1024)
    
    db, _ := sql.Open("postgres", connString)
    db.SetMaxOpenConns(500)  // Can now support more connections
    db.SetMaxIdleConns(100)
    
    return db
}
```

**Why**: Database connections consume file handles; large pools need higher limits.

### Application Initialization

Set appropriate limits during startup:

```go
func init() {
    // Query current limit
    current := maxstdio.GetMaxStdio()
    log.Printf("Default file limit: %d", current)
    
    // Set based on application needs
    required := determineRequiredLimit()
    if current < required {
        old := maxstdio.SetMaxStdio(required)
        log.Printf("Adjusted file limit: %d → %d", old, required)
    }
}

func determineRequiredLimit() int {
    // Calculate based on:
    // - Expected concurrent connections
    // - Number of log files
    // - Database pool size
    // - Temporary file usage
    return 2048 // Example
}
```

**Why**: Proactive configuration prevents runtime errors.

---

## API Reference

### GetMaxStdio

```go
func GetMaxStdio() int
```

Returns the current maximum number of simultaneously open file handles.

**Returns:**
- `int`: Current process limit for open files

**Default Value:**
- Typically 512 on Windows

**Example:**
```go
limit := maxstdio.GetMaxStdio()
fmt.Printf("Current limit: %d\n", limit)
```

### SetMaxStdio

```go
func SetMaxStdio(newMax int) int
```

Sets a new maximum for simultaneously open file handles.

**Parameters:**
- `newMax int`: Desired maximum (typically 512-8192)

**Returns:**
- `int`: Previous maximum value before the change

**Behavior:**
- Returns old limit even if new limit wasn't fully applied
- Actual limit may be constrained by system resources
- Verify with `GetMaxStdio()` after calling

**Typical Values:**
- Default applications: 512
- Web servers: 2048-4096
- Build systems: 2048-8192
- Maximum practical: ~8192

**Example:**
```go
old := maxstdio.SetMaxStdio(2048)
actual := maxstdio.GetMaxStdio()
fmt.Printf("Changed from %d to %d\n", old, actual)
```

---

## Build Instructions

### Prerequisites

**1. Enable CGO**

PowerShell:
```powershell
$env:CGO_ENABLED=1
```

Command Prompt:
```cmd
set CGO_ENABLED=1
```

Permanent (Environment Variables):
- Open System Properties → Advanced → Environment Variables
- Add `CGO_ENABLED` = `1`

**2. Install C Compiler**

#### Option A: MinGW-w64 (Recommended)

Using Chocolatey:
```powershell
choco install mingw
```

Using MSYS2:
```bash
pacman -S mingw-w64-x86_64-gcc
```

Verify:
```powershell
gcc --version
```

#### Option B: Microsoft Visual C++

1. Install Visual Studio Build Tools
2. Select "Desktop development with C++"
3. Verify: `cl` in terminal

### Building Your Application

```bash
# Basic build
go build

# With specific compiler
set CC=gcc
go build

# With verbose output
go build -v
```

### Cross-Compilation

From Linux:
```bash
sudo apt-get install gcc-mingw-w64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc go build
```

From macOS:
```bash
brew install mingw-w64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc go build
```

---

## Best Practices

### 1. Set During Initialization

```go
// ✅ Good - Set once at startup
func init() {
    maxstdio.SetMaxStdio(2048)
}

// ❌ Bad - Repeated calls
func processFile() {
    maxstdio.SetMaxStdio(2048) // Don't do this repeatedly
}
```

### 2. Verify the Change

```go
// ✅ Good - Verify actual limit
old := maxstdio.SetMaxStdio(4096)
actual := maxstdio.GetMaxStdio()
if actual < 4096 {
    log.Printf("Warning: Requested 4096, got %d", actual)
}

// ❌ Bad - Assume it worked
maxstdio.SetMaxStdio(4096) // Might not actually be 4096
```

### 3. Use Reasonable Values

```go
// ✅ Good - Reasonable limits
maxstdio.SetMaxStdio(2048)  // For most servers
maxstdio.SetMaxStdio(4096)  // For high-load servers

// ❌ Bad - Excessive limits
maxstdio.SetMaxStdio(100000) // Wastes resources
```

### 4. Monitor Usage

```go
func monitorFileHandles() {
    limit := maxstdio.GetMaxStdio()
    // Track actual usage vs limit
    // Alert if approaching limit
}
```

---

## Troubleshooting

### Build Error: "C compiler not found"

**Error:**
```
cgo: C compiler "gcc" not found
```

**Solution:**
1. Install MinGW or MSVC (see [Build Instructions](#build-instructions))
2. Add compiler to PATH
3. Restart terminal/IDE
4. Verify: `gcc --version` or `cl`

### Build Error: "CGO not enabled"

**Error:**
```
build constraints exclude all Go files
```

**Solution:**
```powershell
$env:CGO_ENABLED=1
go build
```

### Runtime: Limit Not Changing

**Problem:** `SetMaxStdio()` doesn't increase the limit

**Possible Causes:**
- System-wide hard limit
- Insufficient permissions
- Windows version constraints

**Solution:**
- Run as Administrator
- Verify requested value is reasonable (< 8192)
- Check Windows Event Log for details

### Cross-Compilation Fails

**Problem:** Building from Linux/macOS fails

**Solution:**
```bash
# Install cross-compiler
# Linux: sudo apt-get install gcc-mingw-w64
# macOS: brew install mingw-w64

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
go build
```

---

## Performance Considerations

### CGO Overhead

- **GetMaxStdio()**: ~100-200ns (single CGO call)
- **SetMaxStdio()**: ~100-200ns (single CGO call + system update)

### When to Use

**Beneficial For:**
- Servers with >100 concurrent connections
- Batch processors handling >100 files
- Applications with large connection pools
- Build systems processing many files

**Not Needed For:**
- Simple CLI tools
- Applications with sequential file access
- Low-concurrency services

### Recommendations

1. **Set once** during initialization
2. **Don't exceed** 8192 unless necessary
3. **Monitor** actual file handle usage
4. **Log** limit changes for debugging

---

## Technical Details

### Windows CRT Functions

**_getmaxstdio()**
- Returns current process limit
- Thread-safe
- No side effects
- [Microsoft Docs](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/getmaxstdio)

**_setmaxstdio(int newmax)**
- Sets new process limit
- Returns previous limit
- May be constrained by system
- [Microsoft Docs](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio)

### Default Limits

- **Windows Default**: 512
- **Typical Maximum**: 2048-8192
- **Hard Maximum**: System-dependent

### CGO Configuration

```go
// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "maxstdio.h"
import "C"
```

**CFLAGS:**
- `-g`: Debugging information
- `-Wall`: All warnings

---

## Future Enhancements

Potential improvements for future versions:

**Features**
- Automatic limit detection: Query system maximum
- Validation: Warn if requested limit is too high
- Cross-platform abstraction: Unified API for Windows/Linux/macOS
- Metrics: Track file handle usage

**Developer Experience**
- Build-time validation: Better error messages
- Example applications: More usage patterns
- Benchmarking: Performance impact measurement

Suggestions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for documentation and bug fixing under human supervision.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with documentation and bug fixing
- Maintain Windows-only compatibility
- Preserve CGO integration
- Test on multiple Windows versions when possible

**Documentation**
- Update README.md for new features
- Document C code changes
- Use clear, concise English

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include build verification results

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for project-wide guidelines.

---

## License

MIT License © Nicolas JUHEL

All source files in this package are licensed under the MIT License. See individual files for the full license header.

---

## Resources

**Windows Documentation**
- [C Runtime Library Reference](https://learn.microsoft.com/en-us/cpp/c-runtime-library/)
- [_getmaxstdio](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/getmaxstdio)
- [_setmaxstdio](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio)

**CGO Documentation**
- [CGO Command](https://pkg.go.dev/cmd/cgo)
- [CGO Wiki](https://github.com/golang/go/wiki/cgo)

**C Compilers**
- [MinGW-w64](https://www.mingw-w64.org/)
- [Visual Studio Build Tools](https://visualstudio.microsoft.com/downloads/)

**Related Packages**
- [fileDescriptor](../fileDescriptor) - Linux file descriptor management
- [ioutils](../) - Parent package with additional I/O utilities

**Community**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [Contributing Guide](../../CONTRIBUTING.md)

---

**Platform**: Windows only (CGO required)  
**Maintained By**: maxstdio Package Contributors
