# Encoding Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/encoding)

**Unified encoding/decoding interface with implementations for encryption, hashing, hex, multiplexing, and remote reading.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Sub-Packages](#sub-packages)
- [Coder Interface](#coder-interface)
- [Quick Start](#quick-start)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

The **encoding** package provides a unified `Coder` interface for encoding and decoding operations across multiple implementations including encryption, hashing, hex encoding, multiplexing, and remote data reading.

### Design Philosophy

- **Unified Interface**: Single `Coder` interface for all encoding operations
- **Pluggable Implementations**: Easy to swap between different encoders
- **Stream Support**: Built-in io.Reader/io.Writer wrappers
- **Consistent API**: Same methods across all implementations
- **Extensible**: Easy to add new encoding implementations

---

## Architecture

### Package Structure

```
encoding/
├── interface.go          # Coder interface definition
│
├── aes/                  # AES-256-GCM authenticated encryption
│   ├── interface.go
│   ├── model.go
│   ├── README.md
│   └── TESTING.md
│
├── hexa/                 # Hexadecimal encoding/decoding
│   ├── interface.go
│   ├── model.go
│   ├── README.md
│   └── TESTING.md
│
├── mux/                  # Multiplexer/DeMultiplexer
│   ├── interface.go
│   ├── mux.go
│   ├── demux.go
│   ├── README.md
│   └── TESTING.md
│
├── randRead/             # Remote random data reader
│   ├── interface.go
│   ├── model.go
│   ├── remote.go
│   ├── README.md
│   └── TESTING.md
│
└── sha256/               # SHA-256 cryptographic hashing
    ├── interface.go
    ├── model.go
    ├── README.md
    └── TESTING.md
```

### Coder Interface Flow

```
┌──────────────────────────────────────────────────┐
│              Coder Interface                      │
│                                                   │
│  ┌────────────────────────────────────────────┐ │
│  │  Core Methods:                             │ │
│  │  - Encode([]byte) → []byte                │ │
│  │  - Decode([]byte) → ([]byte, error)       │ │
│  │  - Reset()                                 │ │
│  └────────────────────────────────────────────┘ │
│                                                   │
│  ┌────────────────────────────────────────────┐ │
│  │  Streaming Methods:                        │ │
│  │  - EncodeReader(io.Reader) → io.Reader    │ │
│  │  - DecodeReader(io.Reader) → io.Reader    │ │
│  │  - EncodeWriter(io.Writer) → io.Writer    │ │
│  │  - DecodeWriter(io.Writer) → io.Writer    │ │
│  └────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┬──────────────┬──────────────┐
        ▼              ▼              ▼              ▼              ▼
   ┌─────────┐   ┌──────────┐   ┌───────┐   ┌──────────┐   ┌─────────┐
   │   AES   │   │   Hexa   │   │  Mux  │   │ RandRead │   │ SHA256  │
   │ Encrypt │   │Hex Encode│   │Channel│   │  Remote  │   │  Hash   │
   └─────────┘   └──────────┘   └───────┘   └──────────┘   └─────────┘
```

---

## Sub-Packages

### AES Package

**Purpose**: AES-256-GCM authenticated encryption

**Features:**
- Industry-standard authenticated encryption
- 256-bit key, 96-bit nonce
- Built-in integrity verification
- Hardware acceleration (AES-NI)

**Use Cases:**
- Secure configuration files
- Database field encryption
- Secure file storage
- API response encryption

**Documentation**: [aes/README.md](aes/README.md)

**Quick Example:**
```go
import encaes "github.com/nabbar/golib/encoding/aes"

key, _ := encaes.GenKey()
nonce, _ := encaes.GenNonce()
coder, _ := encaes.New(key, nonce)

encrypted := coder.Encode(plaintext)
decrypted, _ := coder.Decode(encrypted)
```

---

### Hexa Package

**Purpose**: Hexadecimal encoding and decoding

**Features:**
- Standard hex format (0-9, a-f)
- Case-insensitive decoding
- Streaming support
- Stateless operations

**Use Cases:**
- Display binary data
- Configuration files
- Debugging output
- Checksums/hashes

**Documentation**: [hexa/README.md](hexa/README.md)

**Quick Example:**
```go
import enchex "github.com/nabbar/golib/encoding/hexa"

coder := enchex.New()
hex := coder.Encode([]byte("Hello"))
// Output: 48656c6c6f

decoded, _ := coder.Decode(hex)
// Output: Hello
```

---

### Mux Package

**Purpose**: Multiplexing/demultiplexing multiple channels over single stream

**Features:**
- Multiple logical channels
- Rune-based channel keys
- Thread-safe operations
- CBOR+Hex encoding

**Use Cases:**
- Network protocol multiplexing
- Log aggregation
- Test output routing
- Protocol bridging

**Documentation**: [mux/README.md](mux/README.md)

**Quick Example:**
```go
import encmux "github.com/nabbar/golib/encoding/mux"

// Multiplexer
mux := encmux.NewMultiplexer(output, '\n')
channelA := mux.NewChannel('a')
channelA.Write([]byte("data"))

// DeMultiplexer
demux := encmux.NewDeMultiplexer(input, '\n', 4096)
demux.NewChannel('a', outputA)
demux.Copy()
```

---

### RandRead Package

**Purpose**: Buffered random data reader from remote sources

**Features:**
- Automatic reconnection
- Internal buffering
- Thread-safe operations
- Flexible remote sources

**Use Cases:**
- Cryptographic random data
- Random testing data
- Load testing
- Streaming random data

**Documentation**: [randRead/README.md](randRead/README.md)

**Quick Example:**
```go
import "github.com/nabbar/golib/encoding/randRead"

source := func() (io.ReadCloser, error) {
    return http.Get("https://random-service.example.com/bytes")
}

reader := randRead.New(source)
defer reader.Close()

buffer := make([]byte, 100)
reader.Read(buffer)
```

---

### SHA256 Package

**Purpose**: SHA-256 cryptographic hashing

**Features:**
- FIPS 180-4 compliant
- 256-bit output
- Streaming support
- Hex-encoded output

**Use Cases:**
- File integrity verification
- Data deduplication
- Checksum generation
- Content-addressed storage

**Documentation**: [sha256/README.md](sha256/README.md)

**Quick Example:**
```go
import encsha "github.com/nabbar/golib/encoding/sha256"

hasher := encsha.New()
hash := hasher.Encode([]byte("data"))
// Output: 64 hex characters
```

---

## Coder Interface

The `Coder` interface provides a unified API for all encoding operations.

### Interface Definition

```go
type Coder interface {
    // Core operations
    Encode(p []byte) []byte
    Decode(p []byte) ([]byte, error)
    Reset()
    
    // Streaming operations
    EncodeReader(r io.Reader) io.ReadCloser
    DecodeReader(r io.Reader) io.ReadCloser
    EncodeWriter(w io.Writer) io.WriteCloser
    DecodeWriter(w io.Writer) io.WriteCloser
}
```

### Method Overview

| Method | Purpose | Returns |
|--------|---------|---------|
| **Encode** | Encode byte slice | Encoded bytes |
| **Decode** | Decode byte slice | Decoded bytes + error |
| **EncodeReader** | Wrap reader for encoding | io.ReadCloser |
| **DecodeReader** | Wrap reader for decoding | io.ReadCloser |
| **EncodeWriter** | Wrap writer for encoding | io.WriteCloser |
| **DecodeWriter** | Wrap writer for decoding | io.WriteCloser |
| **Reset** | Clear internal state | - |

### Usage Pattern

```go
// Create coder (implementation-specific)
coder := NewCoder(params...)

// Direct encoding/decoding
encoded := coder.Encode(data)
decoded, err := coder.Decode(encoded)

// Streaming encoding
encReader := coder.EncodeReader(inputReader)
io.Copy(output, encReader)

// Streaming decoding
decReader := coder.DecodeReader(inputReader)
io.Copy(output, decReader)

// Clean up
coder.Reset()
```

---

## Quick Start

### Installation

```bash
# Install specific sub-packages
go get github.com/nabbar/golib/encoding/aes
go get github.com/nabbar/golib/encoding/hexa
go get github.com/nabbar/golib/encoding/mux
go get github.com/nabbar/golib/encoding/randRead
go get github.com/nabbar/golib/encoding/sha256
```

### Basic Example

```go
package main

import (
    "fmt"
    enchex "github.com/nabbar/golib/encoding/hexa"
    encsha "github.com/nabbar/golib/encoding/sha256"
)

func main() {
    // Hex encoding
    hexCoder := enchex.New()
    hexEncoded := hexCoder.Encode([]byte("Hello"))
    fmt.Printf("Hex: %s\n", hexEncoded)
    
    // SHA-256 hashing
    hasher := encsha.New()
    hash := hasher.Encode([]byte("Hello"))
    fmt.Printf("SHA-256: %s\n", hash)
}
```

---

## Use Cases

### Secure Data Pipeline

```go
// Encrypt → Hex encode → Store
func secureStore(data []byte, key, nonce) error {
    // Encrypt
    aes, _ := encaes.New(key, nonce)
    encrypted := aes.Encode(data)
    
    // Hex encode for storage
    hex := enchex.New()
    hexEncoded := hex.Encode(encrypted)
    
    // Store
    return storage.Save(hexEncoded)
}

// Load → Hex decode → Decrypt
func secureLoad(key, nonce) ([]byte, error) {
    // Load
    hexEncoded, _ := storage.Load()
    
    // Hex decode
    hex := enchex.New()
    encrypted, _ := hex.Decode(hexEncoded)
    
    // Decrypt
    aes, _ := encaes.New(key, nonce)
    return aes.Decode(encrypted)
}
```

### Multi-Channel Logging

```go
// Aggregate logs from multiple sources
func aggregateLogs() {
    mux := encmux.NewMultiplexer(logFile, '\n')
    
    // Different log channels
    appLog := mux.NewChannel('a')
    errorLog := mux.NewChannel('e')
    accessLog := mux.NewChannel('s')
    
    // Each subsystem writes independently
    go application.Run(appLog)
    go errorHandler.Run(errorLog)
    go webServer.Run(accessLog)
}
```

### File Integrity System

```go
// Generate and verify checksums
type IntegrityChecker struct {
    hasher encoding.Coder
}

func (ic *IntegrityChecker) GenerateChecksum(file string) string {
    data, _ := os.ReadFile(file)
    hash := ic.hasher.Encode(data)
    return string(hash)
}

func (ic *IntegrityChecker) VerifyChecksum(file, expected string) bool {
    actual := ic.GenerateChecksum(file)
    return actual == expected
}
```

---

## Best Practices

### 1. Choose the Right Implementation

```go
// ✅ Good: Use appropriate encoding for use case
aes := encaes.New(key, nonce)       // Encryption
hex := enchex.New()                  // Display/storage
sha := encsha.New()                  // Integrity
```

### 2. Handle Errors

```go
// ✅ Good: Check decode errors
decoded, err := coder.Decode(data)
if err != nil {
    log.Printf("Decode error: %v", err)
}

// ❌ Bad: Ignoring errors
decoded, _ := coder.Decode(data)
```

### 3. Use Streaming for Large Data

```go
// ✅ Good: Stream large files
file, _ := os.Open("large.bin")
encReader := coder.EncodeReader(file)
io.Copy(output, encReader)

// ❌ Bad: Load entire file
data, _ := os.ReadFile("large.bin")
encoded := coder.Encode(data)
```

### 4. Reset State Between Operations

```go
// ✅ Good: Reset between operations
for _, data := range dataList {
    encoded := coder.Encode(data)
    process(encoded)
    coder.Reset()
}
```

### 5. Close Readers/Writers

```go
// ✅ Good: Always close
reader := coder.EncodeReader(input)
defer reader.Close()
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd encoding
go test -v ./...
```

**Sub-Package Tests:**
```bash
cd encoding/aes && go test -v
cd encoding/hexa && go test -v
cd encoding/mux && go test -v
cd encoding/randRead && go test -v
cd encoding/sha256 && go test -v
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain interface compatibility
- Follow existing code style

**New Implementations**
- Implement full `Coder` interface
- Include comprehensive tests
- Document use cases
- Provide examples

**Documentation**
- Update README.md for new features
- Add sub-package documentation
- Keep TESTING.md synchronized
- Document all public APIs with GoDoc

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

Copyright (c) 2023 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
