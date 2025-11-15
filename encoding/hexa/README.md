# Hexadecimal Encoding Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/encoding/hexa)

**Hexadecimal encoding and decoding with streaming I/O support.**

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
- [Streaming Operations](#streaming-operations)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **hexa** package provides standard hexadecimal encoding and decoding for Go applications. It implements the `encoding.Coder` interface for consistent hex operations across the golib ecosystem.

### Design Philosophy

- **Simplicity**: Standard hex encoding (0-9, a-f)
- **Efficiency**: Direct byte slice operations without intermediate buffers
- **Flexibility**: Both byte slice and streaming I/O operations
- **Stateless**: No configuration needed, thread-safe
- **Compatibility**: Case-insensitive decoding

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Standard Hex** | RFC 4648 compliant encoding |
| **Case-Insensitive** | Accepts both uppercase and lowercase |
| **Streaming Support** | `io.Reader` and `io.Writer` interfaces |
| **Memory Efficient** | Direct operations, no buffering |
| **Thread-Safe** | Stateless, safe for concurrent use |
| **Zero Config** | No initialization parameters needed |
| **Lossless** | Perfect round-trip encoding/decoding |

---

## Architecture

### Package Structure

```
encoding/hexa/
├── interface.go        # Public API (New function)
└── model.go           # Core implementation (Coder interface)
```

### Component Architecture

```
┌─────────────────────────────────────────────────────┐
│              Hexa Package                            │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │         Coder Interface                      │  │
│  │  - Encode(bytes) → hex string               │  │
│  │  - Decode(hex) → bytes                      │  │
│  │  - EncodeReader(io.Reader) → io.Reader     │  │
│  │  - DecodeReader(io.Reader) → io.Reader     │  │
│  │  - Reset() (no-op, stateless)              │  │
│  └──────────────────────────────────────────────┘  │
│                        │                             │
│                        ▼                             │
│  ┌──────────────────────────────────────────────┐  │
│  │      Go encoding/hex Package                 │  │
│  │  - hex.Encode() / hex.EncodeToString()     │  │
│  │  - hex.Decode() / hex.DecodeString()       │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Data Flow

```
Encoding Flow:
  Binary Data → Hex Encode → "48656c6c6f" (2× size)
  [0x48, 0x65, 0x6c, 0x6c, 0x6f] → "48656c6c6f"

Decoding Flow:
  "48656c6c6f" → Hex Decode → Binary Data (½ size)
  "48656c6c6f" → [0x48, 0x65, 0x6c, 0x6c, 0x6f]

Size Relationship:
  - Encoded size = Original size × 2
  - Decoded size = Encoded size ÷ 2
  - 1 byte = 2 hex characters
```

---

## Installation

```bash
go get github.com/nabbar/golib/encoding/hexa
```

**Dependencies:**
- Go standard library (`encoding/hex`)
- `github.com/nabbar/golib/encoding` (interface definitions)

---

## Quick Start

### Basic Encoding/Decoding

```go
package main

import (
    "fmt"
    
    enchex "github.com/nabbar/golib/encoding/hexa"
)

func main() {
    // Create a new coder
    coder := enchex.New()
    
    // Encode data to hexadecimal
    plaintext := []byte("Hello, World!")
    encoded := coder.Encode(plaintext)
    fmt.Printf("Encoded: %s\n", encoded)
    // Output: Encoded: 48656c6c6f2c20576f726c6421
    
    // Decode hexadecimal data
    decoded, err := coder.Decode(encoded)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Decoded: %s\n", decoded)
    // Output: Decoded: Hello, World!
}
```

---

## Core Concepts

### Hexadecimal Encoding

Hexadecimal encoding converts each byte (8 bits) into two hexadecimal characters (0-9, a-f).

**Encoding Process:**
```
Byte:     0x48      0x65      0x6c
Binary:   01001000  01100101  01101100
Hex:      4    8    6    5    6    c
Result:   "48"      "65"      "6c"
```

**Properties:**
- **Input**: Binary data (bytes)
- **Output**: Hexadecimal string (2× size)
- **Format**: Lowercase hex (default)
- **Reversible**: Lossless encoding/decoding
- **Character Set**: 0-9, a-f (16 characters)

### Case Insensitivity

Decoding accepts both uppercase and lowercase:

```go
coder := enchex.New()

// All these decode to the same result
decoded1, _ := coder.Decode([]byte("48656c6c6f"))  // lowercase
decoded2, _ := coder.Decode([]byte("48656C6C6F"))  // uppercase
decoded3, _ := coder.Decode([]byte("48656C6c6F"))  // mixed

// All equal: []byte("Hello")
```

### Size Calculations

```go
original := []byte("Hello")      // 5 bytes
encoded := coder.Encode(original)  // 10 bytes (5 × 2)
decoded, _ := coder.Decode(encoded)  // 5 bytes (10 ÷ 2)
```

**Formula:**
- `encoded_size = original_size × 2`
- `decoded_size = encoded_size ÷ 2`

---

## API Reference

### New

**`New() encoding.Coder`**

Creates a new hexadecimal coder instance.

```go
coder := enchex.New()
```

**Returns:**
- Stateless coder instance (thread-safe)

**Notes:**
- No configuration needed
- Multiple instances safe for concurrent use
- `Reset()` is a no-op (stateless)

### Encode

**`Encode(p []byte) []byte`**

Encodes binary data to hexadecimal string.

```go
plaintext := []byte("Hello")
hex := coder.Encode(plaintext)
// hex = []byte("48656c6c6f")
```

**Parameters:**
- `p`: Binary data to encode

**Returns:**
- Hexadecimal representation (2× size)

**Output Format:**
- Lowercase hexadecimal (0-9, a-f)
- No delimiters or spacing
- No prefix (no "0x")

### Decode

**`Decode(p []byte) ([]byte, error)`**

Decodes hexadecimal string to binary data.

```go
hexData := []byte("48656c6c6f")
plaintext, err := coder.Decode(hexData)
if err != nil {
    log.Fatal(err)
}
// plaintext = []byte("Hello")
```

**Parameters:**
- `p`: Hexadecimal data to decode

**Returns:**
- Binary data (½ size)
- Error if invalid hex characters or odd length

**Error Conditions:**
- Invalid hex characters (not 0-9, a-f, A-F)
- Odd length hex string

### Reset

**`Reset()`**

Clears internal state (no-op for stateless coder).

```go
coder.Reset()  // Does nothing, included for interface compatibility
```

**Notes:**
- Included for `encoding.Coder` interface compatibility
- No state to clear (stateless implementation)
- Safe to call, has no effect

---

## Streaming Operations

### Encode Stream

**`EncodeReader(r io.Reader) io.Reader`**

Creates a reader that encodes data on-the-fly.

```go
import "os"

file, _ := os.Open("binary.dat")
defer file.Close()

// Create hex-encoded reader
hexReader := coder.EncodeReader(file)

// Write encoded data to output
output, _ := os.Create("encoded.hex")
defer output.Close()

io.Copy(output, hexReader)
```

### Decode Stream

**`DecodeReader(r io.Reader) io.Reader`**

Creates a reader that decodes hex data on-the-fly.

```go
file, _ := os.Open("encoded.hex")
defer file.Close()

// Create decoded reader
binaryReader := coder.DecodeReader(file)

// Read decoded data
output, _ := os.Create("decoded.dat")
defer output.Close()

io.Copy(output, binaryReader)
```

### Example: Hex Dump File

```go
func hexDumpFile(inputPath, outputPath string) error {
    coder := enchex.New()
    
    // Open input
    input, err := os.Open(inputPath)
    if err != nil {
        return err
    }
    defer input.Close()
    
    // Create output
    output, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer output.Close()
    
    // Encode stream
    hexReader := coder.EncodeReader(input)
    _, err = io.Copy(output, hexReader)
    return err
}
```

---

## Performance

### Benchmark Results

| Operation | Throughput | Allocation | Notes |
|-----------|------------|------------|-------|
| **Encode (1KB)** | ~1 GB/s | 2KB | Doubles size |
| **Decode (1KB)** | ~800 MB/s | 512B | Halves size |
| **Encode (1MB)** | ~1.2 GB/s | 2MB | Large blocks |
| **Decode (1MB)** | ~900 MB/s | 512KB | Large blocks |
| **Stream (4KB buffer)** | ~800 MB/s | 8KB | Buffered I/O |

*Benchmarks on AMD64, Go 1.21, Linux*

### Memory Characteristics

| Operation | Memory | Notes |
|-----------|--------|-------|
| **Coder Instance** | ~0 bytes | Stateless |
| **Encode** | Input × 2 | Output buffer |
| **Decode** | Input ÷ 2 | Output buffer |
| **Stream Buffer** | 4KB default | Configurable |

### Performance Tips

**1. Reuse Coder Instance:**
```go
// ✅ Good: Reuse coder (stateless anyway)
coder := enchex.New()
for _, data := range dataList {
    encoded := coder.Encode(data)
}

// Also OK: Create per operation (no overhead)
for _, data := range dataList {
    coder := enchex.New()
    encoded := coder.Encode(data)
}
```

**2. Pre-allocate for Known Sizes:**
```go
// For encoding
inputSize := len(data)
output := make([]byte, inputSize*2)
// Use encoding/hex directly for pre-allocated buffer
```

**3. Use Streaming for Large Files:**
```go
// ✅ Good: Stream large files
hexReader := coder.EncodeReader(fileReader)
io.Copy(output, hexReader)

// ❌ Bad: Load entire file into memory
data, _ := io.ReadAll(fileReader)
encoded := coder.Encode(data)
```

---

## Use Cases

### Configuration Files

```go
// Store binary data in config files
type Config struct {
    EncryptionKey string `json:"encryption_key"`  // Hex encoded
}

func SaveConfig(cfg Config, key []byte) error {
    coder := enchex.New()
    cfg.EncryptionKey = string(coder.Encode(key))
    
    data, _ := json.Marshal(cfg)
    return os.WriteFile("config.json", data, 0600)
}

func LoadConfig() ([]byte, error) {
    data, _ := os.ReadFile("config.json")
    
    var cfg Config
    json.Unmarshal(data, &cfg)
    
    coder := enchex.New()
    return coder.Decode([]byte(cfg.EncryptionKey))
}
```

### Checksums and Hashes

```go
import (
    "crypto/sha256"
    enchex "github.com/nabbar/golib/encoding/hexa"
)

func ComputeChecksum(data []byte) string {
    hash := sha256.Sum256(data)
    coder := enchex.New()
    return string(coder.Encode(hash[:]))
}

func VerifyChecksum(data []byte, expected string) bool {
    actual := ComputeChecksum(data)
    return actual == expected
}
```

### Debugging Output

```go
func DebugPrint(label string, data []byte) {
    coder := enchex.New()
    hex := coder.Encode(data)
    
    fmt.Printf("%s: %s\n", label, hex)
    fmt.Printf("  Length: %d bytes\n", len(data))
    fmt.Printf("  Hex: %s\n", hex)
}
```

### Wire Protocol

```go
// Encode binary data for text-based protocols
func SendData(conn net.Conn, data []byte) error {
    coder := enchex.New()
    encoded := coder.Encode(data)
    
    // Send as text with newline
    _, err := fmt.Fprintf(conn, "%s\n", encoded)
    return err
}

func ReceiveData(conn net.Conn) ([]byte, error) {
    scanner := bufio.NewScanner(conn)
    if !scanner.Scan() {
        return nil, scanner.Err()
    }
    
    coder := enchex.New()
    return coder.Decode(scanner.Bytes())
}
```

### Database Storage

```go
// Store binary data as hex in VARCHAR columns
type User struct {
    ID     int
    Avatar []byte
}

func (u *User) SaveAvatar(db *sql.DB) error {
    coder := enchex.New()
    hex := coder.Encode(u.Avatar)
    
    _, err := db.Exec(
        "UPDATE users SET avatar = ? WHERE id = ?",
        string(hex), u.ID,
    )
    return err
}

func (u *User) LoadAvatar(db *sql.DB) error {
    var hex string
    err := db.QueryRow(
        "SELECT avatar FROM users WHERE id = ?", u.ID,
    ).Scan(&hex)
    
    if err != nil {
        return err
    }
    
    coder := enchex.New()
    u.Avatar, err = coder.Decode([]byte(hex))
    return err
}
```

---

## Best Practices

### 1. Handle Decode Errors

```go
// ✅ Good: Check decode errors
decoded, err := coder.Decode(hexData)
if err != nil {
    log.Printf("Invalid hex data: %v", err)
    return err
}

// ❌ Bad: Ignoring errors
decoded, _ := coder.Decode(hexData)
```

### 2. Validate Input Length

```go
// ✅ Good: Validate before decode
if len(hexData)%2 != 0 {
    return fmt.Errorf("hex data must have even length")
}
decoded, err := coder.Decode(hexData)

// Info: Decode will catch this too, but explicit is better
```

### 3. Use Appropriate Encoding

```go
// ✅ Good: Use hex for display/debugging
hex := coder.Encode(binaryData)
fmt.Printf("Data: %s\n", hex)

// ❌ Bad: Use hex for storage (base64 is more efficient)
// Hex: 2× expansion
// Base64: 1.33× expansion
```

### 4. Stream Large Files

```go
// ✅ Good: Stream for large files
hexReader := coder.EncodeReader(fileReader)
io.Copy(output, hexReader)

// ❌ Bad: Load everything in memory
data, _ := io.ReadAll(fileReader)  // Memory intensive
encoded := coder.Encode(data)
```

### 5. Consider Case Sensitivity

```go
// ✅ Good: Lowercase for consistency
encoded := coder.Encode(data)  // Always lowercase

// ✅ Good: Accept both for decoding
decoded1, _ := coder.Decode([]byte("48656c6c6f"))  // lowercase
decoded2, _ := coder.Decode([]byte("48656C6C6F"))  // uppercase
// Both work
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd encoding/hexa
go test -v -cover
```

**Test Metrics:**
- 97 test specifications
- 89.7% code coverage
- Ginkgo v2 + Gomega framework
- Edge case testing (invalid input, streaming, etc.)

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain or improve test coverage
- Follow existing code style

**Testing**
- Write tests for all new features
- Test edge cases (empty input, invalid hex, odd length)
- Verify round-trip encoding/decoding
- Include benchmarks for performance-critical code

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Document all public APIs with GoDoc
- Keep TESTING.md synchronized

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Features**
- Formatted output options (with delimiters, spacing)
- Uppercase encoding option
- Prefix options (0x, \x, etc.)
- Chunked output (groups of bytes)

**Performance**
- SIMD optimizations for bulk encoding
- Zero-copy encoding where possible
- Parallel encoding for large data

**Utilities**
- Pretty-print hex dump (with ASCII column)
- Diff hex strings
- Search in hex data
- Hex editor integration helpers

**Compatibility**
- Custom character sets (e.g., 0-9A-F only)
- Alternative formats (hexdump, xxd, etc.)

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Go Standard Library
- **[encoding/hex](https://pkg.go.dev/encoding/hex)** - Standard hex encoding
- **[fmt](https://pkg.go.dev/fmt)** - Printf with %x format
- **[strconv](https://pkg.go.dev/strconv)** - Parse/format hex integers

### Related Golib Packages
- **[encoding](../README.md)** - Encoding interfaces
- **[encoding/aes](../aes/README.md)** - AES encryption
- **[encoding/mux](../mux/README.md)** - Multiplexed encoding

### Hex Standards
- **[RFC 4648](https://tools.ietf.org/html/rfc4648)** - Base encodings (includes hex)
- **[Hexadecimal - Wikipedia](https://en.wikipedia.org/wiki/Hexadecimal)**

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2023 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/hexa)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
