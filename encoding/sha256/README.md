# SHA-256 Encoding Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/encoding/sha256)

**SHA-256 hashing with streaming I/O support implementing the encoding.Coder interface.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Use Cases](#use-cases)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **sha256** package provides SHA-256 cryptographic hashing functionality implementing the `encoding.Coder` interface for consistent hash operations across the golib ecosystem.

### Design Philosophy

- **Standard Algorithm**: FIPS 180-4 compliant SHA-256
- **Stream Support**: io.Reader and io.Writer interfaces
- **Simple API**: Clean encode/decode pattern
- **Stateless**: Thread-safe operations
- **Integration**: Compatible with golib encoding interface

---

## Key Features

| Feature | Description |
|---------|-------------|
| **SHA-256 Hashing** | Cryptographic-strength 256-bit hashes |
| **Streaming Support** | Hash via io.Reader/io.Writer |
| **Hex Encoding** | Output as hexadecimal string |
| **Coder Interface** | Standard encoding.Coder implementation |
| **Thread-Safe** | Stateless operations |
| **Zero Config** | No setup required |

---

## Architecture

### Component Architecture

```
┌─────────────────────────────────────────────────────┐
│              SHA256 Package                          │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         Coder Interface                      │  │
│  │  - Encode(data) → hash                       │  │
│  │  - Decode(hash) → not applicable             │  │
│  │  - EncodeReader(io.Reader) → io.Reader      │  │
│  │  - DecodeReader(io.Reader) → io.Reader      │  │
│  │  - Reset()                                   │  │
│  └────────────────┬─────────────────────────────┘  │
│                   │                                  │
│                   ▼                                  │
│  ┌──────────────────────────────────────────────┐  │
│  │      Go crypto/sha256                        │  │
│  │  - sha256.New()                              │  │
│  │  - Write(data)                               │  │
│  │  - Sum(nil) → 32 bytes                       │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Hashing Flow

```
Input Data (any size)
         │
         ▼
┌────────────────────┐
│  SHA-256 Algorithm │
│  (FIPS 180-4)      │
└────────┬───────────┘
         │
         ▼
   32 bytes hash
         │
         ▼
┌────────────────────┐
│  Hex Encoding      │
└────────┬───────────┘
         │
         ▼
  64 hex characters
```

---

## Installation

```bash
go get github.com/nabbar/golib/encoding/sha256
```

---

## Quick Start

### Basic Hashing

```go
package main

import (
    "fmt"
    encsha "github.com/nabbar/golib/encoding/sha256"
)

func main() {
    // Create hasher
    hasher := encsha.New()
    
    // Hash data
    data := []byte("Hello, World!")
    hash := hasher.Encode(data)
    
    fmt.Printf("SHA-256: %s\n", hash)
    // Output: SHA-256: dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
}
```

### Streaming Hash

```go
import (
    "os"
    encsha "github.com/nabbar/golib/encoding/sha256"
)

func hashFile(filename string) ([]byte, error) {
    file, _ := os.Open(filename)
    defer file.Close()
    
    hasher := encsha.New()
    hashReader := hasher.EncodeReader(file)
    
    // Read to compute hash
    _, _ = io.Copy(io.Discard, hashReader)
    
    // Get hash from coder after reading
    return hasher.Encode(nil), nil
}
```

---

## Core Concepts

### SHA-256 Algorithm

**Properties:**
- **Output Size**: 256 bits (32 bytes)
- **Hex Output**: 64 characters
- **Security**: Cryptographic-strength
- **Collision Resistance**: 2^128 operations
- **Standard**: FIPS 180-4

### One-Way Function

SHA-256 is a **one-way hash function**:
- Easy to compute hash from data
- Computationally infeasible to reverse
- Small change in input → completely different hash
- Deterministic (same input = same output)

### Hex Encoding

Hash output is hex-encoded:
- 32 bytes → 64 hex characters
- Character set: 0-9, a-f
- Lowercase by default

---

## API Reference

### New

**`New() encoding.Coder`**

Creates a new SHA-256 hasher.

```go
hasher := encsha.New()
```

**Returns:** Coder instance

### Encode

**`Encode(p []byte) []byte`**

Computes SHA-256 hash of data.

```go
data := []byte("Hello")
hash := hasher.Encode(data)
// hash is 64 hex characters
```

**Parameters:**
- `p`: Data to hash

**Returns:**
- SHA-256 hash (hex-encoded, 64 bytes)

### Decode

**`Decode(p []byte) ([]byte, error)`**

Not applicable for hashing (returns error).

```go
_, err := hasher.Decode(hash)
// err != nil (hashing is one-way)
```

**Note:** Hash functions are one-way; cannot decode.

### EncodeReader

**`EncodeReader(r io.Reader) io.Reader`**

Creates a reader that hashes data as it's read.

```go
file, _ := os.Open("data.bin")
hashReader := hasher.EncodeReader(file)

// Reading from hashReader computes hash
io.Copy(output, hashReader)
```

### Reset

**`Reset()`**

Resets the hasher state.

```go
hasher.Reset()
```

---

## Use Cases

### File Integrity Verification

```go
func verifyFile(filename string, expectedHash string) bool {
    hasher := encsha.New()
    
    file, _ := os.Open(filename)
    defer file.Close()
    
    data, _ := io.ReadAll(file)
    actualHash := hasher.Encode(data)
    
    return string(actualHash) == expectedHash
}
```

### Password Hashing (Basic)

```go
// Note: For passwords, use bcrypt or argon2 instead
func hashPassword(password string) string {
    hasher := encsha.New()
    hash := hasher.Encode([]byte(password))
    return string(hash)
}
```

### Data Deduplication

```go
func getDataID(data []byte) string {
    hasher := encsha.New()
    hash := hasher.Encode(data)
    return string(hash)
}

// Use hash as unique identifier
dataID := getDataID(fileContents)
if !storage.Exists(dataID) {
    storage.Store(dataID, fileContents)
}
```

### Checksum Generation

```go
func generateChecksum(files []string) map[string]string {
    hasher := encsha.New()
    checksums := make(map[string]string)
    
    for _, file := range files {
        data, _ := os.ReadFile(file)
        hash := hasher.Encode(data)
        checksums[file] = string(hash)
        hasher.Reset()
    }
    
    return checksums
}
```

### Content-Addressed Storage

```go
type ContentStore struct {
    hasher encoding.Coder
}

func (cs *ContentStore) Store(data []byte) string {
    hash := cs.hasher.Encode(data)
    key := string(hash)
    
    // Store with hash as key
    storage.Put(key, data)
    return key
}

func (cs *ContentStore) Retrieve(hash string) []byte {
    return storage.Get(hash)
}
```

---

## Performance

### Characteristics

| Operation | Throughput | Notes |
|-----------|------------|-------|
| **Hash (1KB)** | ~400 MB/s | Modern CPU |
| **Hash (1MB)** | ~450 MB/s | Larger blocks |
| **Streaming** | ~400 MB/s | Buffered I/O |

*Benchmarks on AMD64, Go 1.21*

### Memory

| Aspect | Size | Notes |
|--------|------|-------|
| **Coder Instance** | ~100 bytes | Includes hash state |
| **Hash Output** | 64 bytes | Hex-encoded |
| **Raw Hash** | 32 bytes | Binary form |

### Performance Tips

**1. Reuse Hasher**
```go
// ✅ Good: Reuse with Reset
hasher := encsha.New()
for _, data := range dataList {
    hash := hasher.Encode(data)
    process(hash)
    hasher.Reset()
}
```

**2. Stream Large Files**
```go
// ✅ Good: Stream large files
file, _ := os.Open("large.bin")
hashReader := hasher.EncodeReader(file)
io.Copy(io.Discard, hashReader)
```

**3. Batch Operations**
```go
// ✅ Good: Hash multiple items
hasher := encsha.New()
for _, item := range items {
    hasher.Encode(item)
    hasher.Reset()
}
```

---

## Best Practices

### 1. Use for Integrity, Not Security Alone

```go
// ✅ Good: Integrity verification
fileHash := hasher.Encode(fileData)
verifyIntegrity(fileHash)

// ⚠️ Caution: For passwords, use bcrypt/argon2
// SHA-256 alone is too fast for passwords
```

### 2. Always Reset After Use

```go
// ✅ Good
hasher := encsha.New()
hash1 := hasher.Encode(data1)
hasher.Reset()
hash2 := hasher.Encode(data2)

// ❌ Bad: State carries over
hash1 := hasher.Encode(data1)
hash2 := hasher.Encode(data2)  // Includes data1!
```

### 3. Compare Hashes Correctly

```go
// ✅ Good: Constant-time comparison for security
import "crypto/subtle"

if subtle.ConstantTimeCompare(hash1, hash2) == 1 {
    // Hashes match
}

// ⚠️ Less secure: Timing attack vulnerable
if string(hash1) == string(hash2) {
    // Hashes match
}
```

### 4. Handle Large Files with Streaming

```go
// ✅ Good: Stream large files
file, _ := os.Open("large.iso")
defer file.Close()

hasher := encsha.New()
hashReader := hasher.EncodeReader(file)
io.Copy(io.Discard, hashReader)

// ❌ Bad: Load entire file
data, _ := os.ReadFile("large.iso")  // Memory intensive
hash := hasher.Encode(data)
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd encoding/sha256
go test -v -cover
```

**Test Metrics:**
- Comprehensive test coverage
- Known hash verification
- Streaming tests
- Ginkgo v2 + Gomega framework

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Follow existing code style

**Testing**
- Write tests for all new features
- Verify against known test vectors
- Test streaming operations
- Include benchmarks

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Document all public APIs with GoDoc

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

**Features**
- SHA-512 variant support
- HMAC-SHA256 support
- Parallel hashing for multiple files
- Progress callbacks for large files

**Performance**
- Hardware acceleration (SHA-NI)
- SIMD optimizations
- Zero-copy operations
- Memory pooling

**Utilities**
- Hash verification utilities
- Merkle tree support
- Hash chain validation

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Cryptography
- **[FIPS 180-4](https://csrc.nist.gov/publications/detail/fips/180/4/final)** - SHA-256 specification
- **[SHA-2](https://en.wikipedia.org/wiki/SHA-2)** - Wikipedia article

### Go Standard Library
- **[crypto/sha256](https://pkg.go.dev/crypto/sha256)** - Go SHA-256 implementation
- **[hash](https://pkg.go.dev/hash)** - Hash interface

### Related Golib Packages
- **[encoding](../README.md)** - Encoding interfaces
- **[encoding/hexa](../hexa/README.md)** - Hex encoding
- **[encoding/aes](../aes/README.md)** - AES encryption

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2023 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/sha256)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
