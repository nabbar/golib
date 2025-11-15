# Multiplexer/DeMultiplexer Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/encoding/mux)

**Thread-safe multiplexing and demultiplexing for routing multiple logical channels over a single I/O stream.**

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

The **mux** package provides multiplexing and demultiplexing capabilities to route multiple logical channels over a single physical stream. It uses CBOR encoding for message structure and hexadecimal encoding for payload safety.

### Design Philosophy

- **Channel-Based**: Route data using rune (Unicode) channel keys
- **Thread-Safe**: Full concurrency support with minimal contention
- **Stream-Based**: Standard io.Writer/io.Reader interfaces
- **Reliable**: Delimiter-based framing prevents data corruption
- **Efficient**: Lock-free demultiplexing with sync.Map

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Multiple Channels** | Route data from multiple sources over single stream |
| **Rune Keys** | Use Unicode characters for channel identification |
| **Bidirectional** | Full multiplexing (write) and demultiplexing (read) |
| **Delimiter Framing** | Configurable message boundaries |
| **CBOR Encoding** | Compact binary encoding for metadata |
| **Hex Payload** | Safe encoding preventing delimiter collision |
| **Thread-Safe** | Concurrent goroutine access |
| **Lock-Free Reads** | sync.Map for high-performance routing |

---

## Architecture

### Component Architecture

```
┌─────────────────────────────────────────────────────┐
│              MULTIPLEXER SIDE                        │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │Channel 'a'│  │Channel 'b'│  │Channel 'c'│         │
│  │io.Writer │  │io.Writer │  │io.Writer │         │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘         │
│       │             │             │                 │
│       └─────────────┴─────────────┘                 │
│                     │                               │
│         ┌───────────▼──────────┐                    │
│         │    Multiplexer       │                    │
│         │  - CBOR encode       │                    │
│         │  - Hex payload       │                    │
│         │  - Mutex lock        │                    │
│         └───────────┬──────────┘                    │
└─────────────────────┼──────────────────────────────┘
                      │
         ╔════════════▼═══════════╗
         ║    Physical Stream     ║
         ║   (Single io.Writer)   ║
         ╚════════════╤═══════════╝
                      │
┌─────────────────────┼──────────────────────────────┐
│         ┌───────────▼──────────┐                    │
│         │   DeMultiplexer      │                    │
│         │  - CBOR decode       │                    │
│         │  - Hex decode        │                    │
│         │  - sync.Map lookup   │                    │
│         └───────────┬──────────┘                    │
│                     │                               │
│       ┌─────────────┴─────────────┐                 │
│       │             │             │                 │
│  ┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐         │
│  │Channel 'a'│  │Channel 'b'│  │Channel 'c'│         │
│  │io.Writer │  │io.Writer │  │io.Writer │         │
│  └──────────┘  └──────────┘  └──────────┘         │
│                                                      │
│             DEMULTIPLEXER SIDE                       │
└─────────────────────────────────────────────────────┘
```

### Message Format

```
┌──────────────────────────────────────────────┐
│           Single Message Structure            │
├──────────────────────────────────────────────┤
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │        CBOR Encoded Header           │   │
│  │  {                                   │   │
│  │    "K": rune,      // Channel key    │   │
│  │    "D": string     // Hex payload    │   │
│  │  }                                   │   │
│  └──────────────────────────────────────┘   │
│                     │                        │
│                     ▼                        │
│  ┌──────────────────────────────────────┐   │
│  │         Delimiter Byte               │   │
│  │           (e.g., '\n')               │   │
│  └──────────────────────────────────────┘   │
│                                              │
└──────────────────────────────────────────────┘

Example on wire:
  A26144486568446B61\n
  │                │
  CBOR+Hex       Delim
```

---

## Installation

```bash
go get github.com/nabbar/golib/encoding/mux
```

**Dependencies:**
- `github.com/fxamacker/cbor/v2` - CBOR encoding
- `github.com/nabbar/golib/encoding/hexa` - Hex encoding

---

## Quick Start

### Multiplexer Example

```go
package main

import (
    "os"
    encmux "github.com/nabbar/golib/encoding/mux"
)

func main() {
    // Create multiplexer writing to stdout
    mux := encmux.NewMultiplexer(os.Stdout, '\n')
    
    // Create channels
    channelA := mux.NewChannel('a')
    channelB := mux.NewChannel('b')
    
    // Write to channels
    channelA.Write([]byte("Message on channel A"))
    channelB.Write([]byte("Message on channel B"))
}
```

### DeMultiplexer Example

```go
package main

import (
    "bytes"
    "os"
    encmux "github.com/nabbar/golib/encoding/mux"
)

func main() {
    // Create demultiplexer reading from stdin
    demux := encmux.NewDeMultiplexer(os.Stdin, '\n', 4096)
    
    // Create output buffers
    bufA := &bytes.Buffer{}
    bufB := &bytes.Buffer{}
    
    // Register channels
    demux.NewChannel('a', bufA)
    demux.NewChannel('b', bufB)
    
    // Start demultiplexing
    err := demux.Copy()
    if err != nil {
        panic(err)
    }
    
    // Read results
    println("Channel A:", bufA.String())
    println("Channel B:", bufB.String())
}
```

---

## Core Concepts

### Multiplexing

**Definition**: Combining multiple logical channels into a single physical stream.

**Process:**
1. Data written to channel writer
2. CBOR encodes: `{K: rune, D: hex_data}`
3. Appends delimiter byte
4. Writes to underlying stream

**Thread Safety**: Mutex-protected writes ensure atomic messages.

### DeMultiplexing

**Definition**: Splitting a single physical stream into multiple logical channels.

**Process:**
1. Read until delimiter
2. CBOR decode message
3. Hex decode payload
4. Route to channel writer via key

**Thread Safety**: Lock-free reads using sync.Map for channel lookup.

### Channel Keys

- **Type**: `rune` (Unicode code point)
- **Examples**: `'a'`, `'b'`, `'1'`, `'@'`, `'α'`, `'中'`
- **Flexibility**: Any Unicode character
- **Collision**: Different keys = different channels

### Delimiter

- **Purpose**: Message boundary marker
- **Type**: Single byte
- **Common**: `'\n'`, `'\r'`, `'|'`
- **Requirement**: Must not appear in CBOR+Hex output
- **Safety**: Hex encoding prevents collision

---

## API Reference

### Multiplexer Interface

**`NewMultiplexer(w io.Writer, delim byte) Multiplexer`**

Creates a new multiplexer.

```go
mux := encmux.NewMultiplexer(writer, '\n')
```

**Parameters:**
- `w`: Underlying writer
- `delim`: Message delimiter byte

**Returns:** Multiplexer instance

**`NewChannel(key rune) io.Writer`**

Creates a channel writer.

```go
channel := mux.NewChannel('a')
channel.Write([]byte("data"))
```

**Parameters:**
- `key`: Channel identifier (rune)

**Returns:** io.Writer for the channel

### DeMultiplexer Interface

**`NewDeMultiplexer(r io.Reader, delim byte, size int) DeMultiplexer`**

Creates a new demultiplexer.

```go
demux := encmux.NewDeMultiplexer(reader, '\n', 4096)
```

**Parameters:**
- `r`: Underlying reader
- `delim`: Message delimiter byte
- `size`: Buffer size (0 for default)

**Returns:** DeMultiplexer instance

**`NewChannel(key rune, w io.Writer)`**

Registers a channel writer.

```go
demux.NewChannel('a', outputWriter)
```

**Parameters:**
- `key`: Channel identifier
- `w`: Destination writer

**`Copy() error`**

Starts demultiplexing (blocks until EOF).

```go
err := demux.Copy()
```

**Returns:** Error if any (nil on EOF)

**`Read(p []byte) (n int, err error)`**

Implements io.Reader (reads next message).

```go
n, err := demux.Read(buffer)
```

---

## Use Cases

### Network Multiplexing

```go
// Server side - multiplex multiple connections
conn, _ := net.Dial("tcp", "server:8080")
mux := encmux.NewMultiplexer(conn, '\n')

// Separate channels for control and data
control := mux.NewChannel('c')
data := mux.NewChannel('d')

control.Write([]byte("READY"))
data.Write(fileData)
```

### Log Aggregation

```go
// Aggregate logs from multiple sources
logFile, _ := os.Create("combined.log")
mux := encmux.NewMultiplexer(logFile, '\n')

appLog := mux.NewChannel('a')
errorLog := mux.NewChannel('e')
accessLog := mux.NewChannel('s')

// Each subsystem writes to its channel
go app.Run(appLog)
go errorHandler.Run(errorLog)
go webServer.Run(accessLog)
```

### Test Output Routing

```go
// Route test output by category
testOutput := &bytes.Buffer{}
mux := encmux.NewMultiplexer(testOutput, '\n')

stdout := mux.NewChannel('o')
stderr := mux.NewChannel('e')
results := mux.NewChannel('r')

// Run tests with separate output channels
cmd.Stdout = stdout
cmd.Stderr = stderr
cmd.Run()
results.Write(testResults)
```

### Protocol Bridging

```go
// Bridge two protocols over single connection
mux := encmux.NewMultiplexer(networkConn, '\n')
http := mux.NewChannel('h')
websocket := mux.NewChannel('w')

go httpProxy(http, backend1)
go wsProxy(websocket, backend2)
```

---

## Performance

### Benchmark Results

| Operation | Throughput | Notes |
|-----------|------------|-------|
| **Mux Write (1KB)** | ~50 MB/s | CBOR + Hex overhead |
| **DeMux Read (1KB)** | ~40 MB/s | Parse + decode |
| **Concurrent Mux** | ~150 MB/s | 3 goroutines |
| **Concurrent DeMux** | ~120 MB/s | 3 goroutines |

*Benchmarks on AMD64, Go 1.21, Linux*

### Memory Characteristics

| Component | Memory | Notes |
|-----------|--------|-------|
| **Multiplexer** | ~100 bytes | + mutex |
| **DeMultiplexer** | ~200 bytes | + sync.Map |
| **Channel** | ~50 bytes | Per channel |
| **Message Overhead** | ~30 bytes | CBOR header |

### Performance Tips

**1. Buffer Size:**
```go
// ✅ Good: Appropriate buffer for message size
demux := NewDeMultiplexer(r, '\n', 8192)

// ❌ Bad: Too small buffer for large messages
demux := NewDeMultiplexer(r, '\n', 64)
```

**2. Delimiter Choice:**
```go
// ✅ Good: Use '\n' for line-based protocols
mux := NewMultiplexer(w, '\n')

// ✅ Also good: Use rare byte for binary data
mux := NewMultiplexer(w, 0xFF)
```

**3. Channel Reuse:**
```go
// ✅ Good: Reuse channel writers
ch := mux.NewChannel('a')
for _, data := range dataList {
    ch.Write(data)
}

// ❌ Bad: Creating new channel each time
for _, data := range dataList {
    mux.NewChannel('a').Write(data)
}
```

---

## Best Practices

### 1. Always Check Errors

```go
// ✅ Good
n, err := channel.Write(data)
if err != nil {
    log.Printf("Write error: %v", err)
}

// ❌ Bad
channel.Write(data)  // Ignoring errors
```

### 2. Choose Delimiter Carefully

```go
// ✅ Good: Common text delimiter
mux := NewMultiplexer(w, '\n')

// ✅ Good: For binary protocols
mux := NewMultiplexer(w, 0x00)

// ❌ Bad: Common byte in data
mux := NewMultiplexer(w, 'a')  // May appear in hex!
```

### 3. Register Channels Before Copy

```go
// ✅ Good
demux.NewChannel('a', writerA)
demux.NewChannel('b', writerB)
go demux.Copy()

// ❌ Bad: Race condition
go demux.Copy()
time.Sleep(100 * time.Millisecond)
demux.NewChannel('a', writerA)
```

### 4. Handle EOF Gracefully

```go
// ✅ Good
err := demux.Copy()
if err != nil && err != io.EOF {
    log.Printf("Error: %v", err)
}

// Copy() returns nil on EOF by design
```

### 5. Use Appropriate Buffer Size

```go
// ✅ Good: Match expected message size
demux := NewDeMultiplexer(r, '\n', 4096)

// ✅ Good: No buffer for small messages
demux := NewDeMultiplexer(r, '\n', 0)
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd encoding/mux
go test -v -cover
```

**Test Metrics:**
- 59 test specifications
- Comprehensive coverage
- Race condition tested
- Concurrent scenarios

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain thread safety guarantees
- Follow existing code style

**Testing**
- Write tests for all new features
- Test concurrent scenarios
- Verify thread safety with `-race`
- Include benchmarks for performance changes

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Document all public APIs with GoDoc
- Keep TESTING.md synchronized

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

**Protocol Features**
- Compression support (gzip, zstd)
- Message priority/ordering
- Flow control mechanisms
- Heartbeat/keepalive

**Performance**
- Zero-copy operations where possible
- Batch message encoding
- Parallel channel processing
- Memory pooling

**Features**
- Channel statistics/metrics
- Dynamic channel creation
- Broadcast channels
- Error recovery strategies

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Go Standard Library
- **[io](https://pkg.go.dev/io)** - Reader/Writer interfaces
- **[bufio](https://pkg.go.dev/bufio)** - Buffered I/O
- **[sync](https://pkg.go.dev/sync)** - Synchronization primitives

### External Libraries
- **[CBOR](https://github.com/fxamacker/cbor)** - CBOR encoding library

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
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/mux)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
