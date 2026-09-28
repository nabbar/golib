# OpenPGP Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)

Post-Quantum OpenPGP key management and streaming encryption/decryption library. Provides a uniform `OpenGPG` interface across hybrid quantum-safe algorithms (ML-DSA/EdDSA), with two concrete implementations: **ML-DSA65 + Ed25519** and **ML-DSA87 + Ed448**. Built on top of [ProtonMail/go-crypto/openpgp](https://github.com/ProtonMail/go-crypto), it adds deterministic key lifetime management, resource cleanup via `mapCloser`, and memory-efficient streaming through `io.Pipe` with `sync.Pool` buffer reuse.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Subpackages](#subpackages)
   - [ML-DSA65 + Ed25519](#mldsa65ed)
   - [ML-DSA87 + Ed448](#mldsa87ed)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

The `encoding/opengpg` package defines a contract — the `OpenGPG` interface — and supplies the error code registry and the `Identity` data structure shared by all OpenPGP backends. Each backend sub-package implements `OpenGPG` with a specific Post-Quantum hybrid KEM algorithm combined with an EdDSA subkey:

| Sub-package | KEM          | Signature | Hash       | Symmetric Cipher |
|-------------|--------------|-----------|------------|------------------|
| `mldsa65ed` | ML-DSA-65    | Ed25519   | SHA-3/512  | AES-256          |
| `mldsa87ed` | ML-DSA-87    | Ed448     | SHA-3/512  | AES-256          |

Both backends operate on OpenPGP V6 keys and support streaming encryption/decryption via `io.Reader` and `io.Writer` adapters, making them suitable for large payloads without loading everything into memory.

### Design Philosophy

1. **Interface-driven design**: A single `OpenGPG` interface decouples consumers from the chosen algorithm. Swapping between ML-DSA65 and ML-DSA87 requires only changing the `New()` call — no consumer code changes.
2. **Deterministic resource management**: All pipe ends, file descriptors, and temporary buffers created during streaming are tracked by `mapCloser.Closer` and released atomically when `Close()` is called, preventing resource leaks even under panic.
3. **Memory-efficient streaming**: Streaming paths use `io.Pipe` and `sync.Pool`-backed buffers (32 KB default) so data never accumulates entirely in memory. Buffer contents are explicitly cleared (via `clear()`) after use to prevent memory-resident sensitive data.
4. **Typed error registry**: All errors flow through `golib/errors` with unique per-package codes, enabling callers to handle errors by code rather than string matching.

### Key Features

- Post-Quantum hybrid key generation (KEM + EdDSA per FIPS 203 / FIPS 204)
- Streaming encryption and decryption (ReadCloser and WriteCloser adapters)
- Automatic key expiration via `KeyLifetimeSecs` when `KeyTime` is in the future
- ASCII-armored and raw binary key import/export
- Built-in self-test (`Check()`) that verifies encrypt-then-decrypt round-trip
- Deterministic cleanup of all ephemeral resources on `Close()`
- OpenPGP V6 key format support

### Key Benefits

- **Post-Quantum security**: Hybrid KEM (ML-DSA + EdDSA) provides quantum resilience per FIPS 203/204 standards while maintaining compatibility with classical EdDSA for signature operations.
- **Zero memory footprint for large payloads**: Streaming via `io.Pipe` means a multi-gigabyte file is never fully loaded — data flows through a 32 KB buffer pool.
- **Predictable resource lifecycle**: `mapCloser.Closer` tracks every pipe end created during streaming, ensuring no goroutine hangs or leaked descriptors on errors.
- **Uniform API across backends**: The `OpenGPG` interface is identical for both ML-DSA65 and ML-DSA87 implementations, simplifying code reuse and testing.

---

## Architecture

### Package Structure

```
encoding/opengpg/                           # root package — interface, errors, Identity
├── errors.go                               # Error code constants and registry with golib/errors
├── interface.go                            # OpenGPG interface and Identity struct definitions
│
├── mldsa65ed/                              # ML-DSA65 + Ed25519 backend
│   ├── interface.go                        # Options struct and New() constructor
│   ├── model.go                            # Concrete implementation of OpenGPG
│   ├── tools.go                            # parseKeyRing helper
│   ├── benchmark_test.go                   # Benchmark tests
│   ├── helper_test.go                      # Test helpers
│   ├── mldsa65ed_suite_test.go            # Ginkgo test suite
│   └── model_test.go                       # Unit tests for model methods
│
└── mldsa87ed/                              # ML-DSA87 + Ed448 backend
    ├── interface.go                        # Options struct and New() constructor
    ├── model.go                            # Concrete implementation of OpenGPG
    ├── tools.go                            # parseKeyRing helper
    ├── benchmark_test.go                   # Benchmark tests
    ├── helper_test.go                      # Test helpers
    ├── mldsa87ed_suite_test.go            # Ginkgo test suite
    └── model_test.go                       # Unit tests for model methods
```

### Package Architecture

```
                ┌──────────────────────┐
                │   Consumer Code      │
                │ (any Go application) │
                └──────────┬───────────┘
                           │ uses
                           ▼
                ┌──────────────────────┐
                │  opengpg.OpenGPG     │◄──── contract (interface)
                │  (interface)         │
                └────┬────────┬────────┘
                     │        │ implements
                ┌────▼────┐ ┌─▼───────┐
                │mldsa65ed│ │mldsa87ed│  ── two Post-Quantum backends
                │ backend │ │ backend │
                └───┬─────┘ └───┬─────┘
                    │           │ depends on
                    ▼           ▼
┌─────────────────────────────────────────────────────────────┐
│                 ProtonMail/go-crypto                        │
│             openpgp + openpgp/pkt (V6 layer)                │
└─────────────────────────────────────────────────────────────┘
           │                              │
           ▼ depends on                   ▼ depends on
┌──────────────────────┐    ┌─────────────────────────────────┐
│golib/errors          │    │golib/ioutils/mapCloser          │
│(error code registry) │    │(resource tracking)              │
└──────────────────────┘    └─────────────────────────────────┘
```

### Dataflow

```
┌─────────────────────────────────────────────────────────────┐
│                   Key Management Flow                       │
├─────────────────────────────────────────────────────────────┤
│  1. Create Key Pair:                                        │
│     Identity{Name, Comment, Email}                          │
│          │                                                  │
│          ├──► NewEntity (go-crypto)                         │
│          ├──► Serialize ──────────► PublicKey               │
│          └──► SerializePrivate ───► PrivateKey              │
│                                                             │
│  2. Load Existing Key Pair:                                 │
│     Identity{PublicKey or PrivateKey}                       │
│          │                                                  │
│          └──► parseKeyRing ───────► Stored EntityList       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Streaming I/O Flow                        │
├─────────────────────────────────────────────────────────────┤
│  Encrypt Path:                                              │
│     [Reader] ──(pipe)──► [Encrypt Process] ──► [Ciphertext] │
│                                                             │
│  Decrypt Path:                                              │
│     [Reader] ──(pipe)──► [Decrypt Process] ──► [Plaintext]  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Resource Cleanup                         │
├─────────────────────────────────────────────────────────────┤
│  Close():                                                   │
│   ├── c.Store(true) ──► Stop new operations                 │
│   ├── l.Close()     ──► Release pipes, writers, descriptors │
│   ├── o.i = nil     ──► Clear entity list (GC eligible)     │
│   └── o.o = nil     ──► Clear config (GC eligible)          │
└─────────────────────────────────────────────────────────────┘
```

### Data Fields Reference

| Field                | Type               | Description                                                                     |
|----------------------|--------------------|---------------------------------------------------------------------------------|
| `Identity.Name`      | `string`           | User name portion of the OpenPGP user ID                              |
| `Identity.Comment`   | `string`           | Optional comment portion of the OpenPGP user ID                       |
| `Identity.Email`     | `string`           | Email address portion of the OpenPGP user ID                          |
| `Identity.PublicKey` | `[]byte`           | ASCII-armored public key material (populated by `Create` or provided for `Load`) |
| `Identity.PrivateKey`| `[]byte`           | ASCII-armored private key material (populated by `Create` or provided for `Load`) |
| `Options.Rand`       | `io.Reader`        | Cryptographic random source; defaults to `crypto/rand.Reader` when nil |
| `Options.Time`       | `func() time.Time` | Clock override for deterministic timestamps; nil for system clock     |
| `Options.KeyTime`    | `uint32`           | Target key creation time; sets `KeyLifetimeSecs` for automatic expiration |

| Error Code       | Constant                    | Message                            |
|------------------|-----------------------------|------------------------------------|
| Param empty      | `ErrorParamEmpty`           | "given parameters is empty" |
| Identity invalid | `ErrorIdentityInvalid`      | "identity is empty or invalid" |
| Public key error | `ErrorPublicKeyInvalid`     | "public key is empty or invalid" |
| Private key error| `ErrorPrivateKeyInvalid`    | "private key is empty or invalid" |

---

## Performance

Both backends use streaming I/O with `io.Pipe` and a `sync.Pool` of 32 KB buffers, achieving O(1) memory usage relative to payload size. The following highlights are worth noting:

- **Memory**: Encryption/decryption never materializes the full payload in memory. A 1 GB file is processed with ~32 KB of buffer space per streaming goroutine.
- **Buffer reuse**: `sync.Pool` eliminates per-operation allocations for the streaming buffer. `clear()` is called before return to prevent sensitive data from persisting.
- **Hash/Cipher overhead**: Both backends use SHA-3/512 (default hash) and AES-256 (default cipher). SHA-3 is software-only and slower than SHA-2, but provides the required hash for ML-DSA FIPS compliance. AES-256 benefits from hardware acceleration (AES-NI) where available.
- **KEM key size trade-off**: ML-DSA-87 keys are larger (~5.6 KB public, ~2.5 KB secret) than ML-DSA-65 (~4.1 KB public, ~1.6 KB secret), trading key material size for higher security level (FIPS 203 Level 5 vs. Level 3).

---

## Subpackages

### mldsa65ed

Post-Quantum OpenPGP backend using the **ML-DSA65 + Ed25519** hybrid KEM. ML-DSA65 corresponds to FIPS 203 Level 3 security; Ed25519 provides the classical signature subkey.

**Key Features**:
- V6 key generation with automatic expiration when `Options.KeyTime` is set
- Streaming encrypt/decrypt via `io.Pipe` with 32 KB buffer pool
- ASCII-armored and binary key import via `parseKeyRing`
- Self-test round-trip verification with `Check()`
- Deterministic resource cleanup with `mapCloser.Closer`
- OpenPGP V6 key format with SHA-3/512 hash and AES-256 cipher

---

### mldsa87ed

Post-Quantum OpenPGP backend using the **ML-DSA87 + Ed448** hybrid KEM. ML-DSA87 corresponds to FIPS 203 Level 5 security (highest level); Ed448 provides the classical signature subkey.

**Key Features**:
- V6 key generation with automatic expiration when `Options.KeyTime` is set
- Streaming encrypt/decrypt via `io.Pipe` with 32 KB buffer pool
- ASCII-armored and binary key import via `parseKeyRing`
- Self-test round-trip verification with `Check()`
- Deterministic resource cleanup with `mapCloser.Closer`
- OpenPGP V6 key format with SHA-3/512 hash and AES-256 cipher
- Higher security level at the cost of larger key material

---

## Use Cases

### 1. Post-Quantum Key Management

Generate and manage Post-Quantum OpenPGP key pairs with automatic key expiration, suitable for environments requiring FIPS 203/204 compliance.

**References**: [FIPS 203](https://csrc.nist.gov/pubs/fips/203/final), [FIPS 204](https://csrc.nist.gov/pubs/fips/204/final), [RFC 9580 — The OpenPGP Message Format](https://datatracker.ietf.org/doc/html/rfc9580)

```go
package main

import (
    "context"

    "github.com/nabbar/golib/encoding/opengpg"
    "github.com/nabbar/golib/encoding/opengpg/mldsa65ed"
)

func main() {
    ctx := context.Background()

	gpg := mldsa65ed.New(ctx, mldsa65ed.Options{
		KeyTime: 0, // no automatic expiration
	})
    defer func() {
        _ = gpg.Close()
	}()

	// Create a new Post-Quantum key pair
	identity := &opengpg.Identity{
		Name:    "Alice",
		Comment: "Post-Quantum Key",
		Email:   "alice@example.com",
	}

	if err := gpg.Create(identity); err != nil {
		panic(err)
	}

	// identity.PublicKey and identity.PrivateKey are now populated
	// with ASCII-armored key material
}
```

### 2. Streaming Encryption for Large Files

Encrypt or decrypt large data streams (files, network connections) without loading the full payload into memory. The `io.Pipe` + `sync.Pool` design ensures O(1) memory use.

```go
package main

import (
	"context"
	"io"
	"os"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/mldsa87ed"
)

func main() {
    ctx := context.Background()

	// Load an existing key from stored armored material
	identity := &opengpg.Identity{
		Name:      "Bob",
		Comment:   "Post-Quantum Key",
		Email:     "bob@example.com",
		PublicKey: []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n..."),
	}

	gpg := mldsa87ed.New(ctx, mldsa87ed.Options{})
    defer func() {
       _ = gpg.Close()
    }()


   if err := gpg.Load(identity); err != nil {
		panic(err)
	}

	// Self-test the key pair
	if err := gpg.Check(); err != nil {
		panic(err)
	}

	// Stream encrypt a file
	in, err := os.Open("large_file.bin")
	if err != nil {
		panic(err)
	}
    defer func() {
	    _ = in.Close()
    }()
	
	out, err := os.Create("large_file.bin.gpg")
	if err != nil {
		panic(err)
	}
    defer func() {
        _ = out.Close()
    }()

	encReader, err := gpg.EncryptReader(in)
	if err != nil {
		panic(err)
	}
	
	_, err = io.Copy(out, encReader)
	if err != nil {
		panic(err)
	}
	
	_ = encReader.Close()
}
```

### 3. Key Expiration for Temporary Key Rotation

Generate keys with automatic expiration for rotation policies in HSM or certificate management scenarios.

```go
package main

import (
    "context"
    "time"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/mldsa65ed"
)

func main() {
    ctx := context.Background()

	// Expire key after 90 days from now
	expireIn := time.Now().Add(90 * 24 * time.Hour)
	keyTime := uint32(expireIn.Unix())

	gpg := mldsa65ed.New(ctx, mldsa65ed.Options{
		KeyTime: keyTime,
	})
    defer func() {
        _ = gpg.Close()
    }()

	identity := &opengpg.Identity{
		Name:    "TemporaryKey",
		Comment: "90-day rotation",
		Email:   "temp@example.com",
	}

	err := gpg.Create(identity)
	// Key will expire after 90 days as recorded in KeyLifetimeSecs
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/encoding/opengpg
go get github.com/nabbar/golib/encoding/opengpg/mldsa65ed
# or
go get github.com/nabbar/golib/encoding/opengpg/mldsa87ed
```

### Basic Implementation

Create a key pair, encrypt a string, and decrypt it back using the ML-DSA65 + Ed25519 backend:

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "io"

	"github.com/nabbar/golib/encoding/opengpg"
	"github.com/nabbar/golib/encoding/opengpg/mldsa65ed"
)

func main() {
ctx := context.Background()

	// Create OpenGPG instance with ML-DSA65/Ed25519
	gpg := mldsa65ed.New(ctx, mldsa65ed.Options{})
    defer func() {
        _ = gpg.Close()
    }()

	// Generate key pair
	identity := &opengpg.Identity{
		Name:    "Alice",
		Comment: "Test Key",
		Email:   "alice@example.com",
	}

	if err := gpg.Create(identity); err != nil {
		panic(err)
	}

	// Verify key pair is functional
	if err := gpg.Check(); err != nil {
		panic(err)
	}

	// Encrypt data
	plaintext := []byte("Hello, Post-Quantum World!")
	buf := bytes.NewBuffer(nil)

	encWriter, err := gpg.EncryptWriter(buf)
	if err != nil {
		panic(err)
	}

	_, err = encWriter.Write(plaintext)
	if err != nil {
		panic(err)
	}

	err = encWriter.Close()
	if err != nil {
		panic(err)
	}

	// Decrypt data
	decReader, err := gpg.DecryptReader(buf)
	if err != nil {
		panic(err)
	}

	defer func() {
        _ = decReader.Close()
    }()

    decrypted, err := io.ReadAll(decReader)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Decrypted: %s\n", string(decrypted))
	// → Decrypted: Hello, Post-Quantum World!
}
```

### Streaming Encryption from Reader

Stream-encrypt data from a reader without buffering the full payload:

```go
// EncryptReader — data flows through io.Pipe with 32 KB buffer pool
plaintext := bytes.NewReader([]byte("streaming data"))

encReader, err := gpg.EncryptReader(plaintext)

if err != nil {
    panic(err)
}

defer encReader.Close()

// Read encrypted bytes directly — no full materialization
out, _ := io.ReadAll(encReader)
```

---

## Best Practices

### ✅ DO

- **Always call `Close()` on the OpenGPG instance** — use `defer gpg.Close()`. This releases all tracked pipe ends and nils out the entity list to allow garbage collection of key material.
- **Use `Check()` after `Create` or `Load`** — validates the round-trip encrypt/decrypt before relying on the key.
```go
    if err := gpg.Check(); err != nil {
        return err // key is not functional
    }
```
- **Provide `Options.Time` in tests** — use a fixed clock to make key generation deterministic across test runs.
```go
    mldsa65ed.New(ctx, mldsa65ed.Options{
        Time: func() time.Time { 
			return time.Unix(1700000000, 0) 
		},
    })
  ```
- **Use `EncryptWriter` / `DecryptWriter` for synchronous writes** — simpler than the reader-based path when the consumer writes sequentially.
- **Set `KeyTime` for automatic key expiration** — ensures keys self-expire after a defined lifetime.

### ❌ DON'T

- **Don't keep an OpenGPG instance alive indefinitely** — if `Close()` is never called, pipe ends tracked by `mapCloser.Closer` remain allocated, and key material persists in memory.
- **Don't reuse an instance after `Close()`** — all operations return `os.ErrClosed`. Create a new instance with `New()`.
- **Don't pass nil `Identity` to `Create` or `Load`** — returns `ErrorIdentityInvalid`.
- **Don't load an instance with both empty `PublicKey` and `PrivateKey`** — `Load` returns `ErrorIdentityInvalid` if both are empty.
- **Don't use `DecryptReader` for large streams** — it reads the full message synchronously into memory via `ReadMessage`; prefer `DecryptWriter` for large data.

---

## API Reference

### OpenGPG Interface

| Method | Parameters | Result | Description |
|--------|------------|--------|-------------|
| `Create` | `*Identity` | `error` | Generates a new key pair and writes ASCII-armored keys into `Identity` |
| `Load` | `*Identity` | `error` | Imports existing key from `Identity.PublicKey` or `Identity.PrivateKey` |
| `Check` | *(none)* | `error` | Validates key pair via encrypt-then-decrypt round-trip |
| `EncryptReader` | `io.Reader` | `(io.ReadCloser, error)` | Streaming encryption from reader (uses `io.Pipe`) |
| `EncryptWriter` | `io.Writer` | `(io.WriteCloser, error)` | Synchronous encryption to writer |
| `DecryptReader` | `io.Reader` | `(io.ReadCloser, error)` | Synchronous decryption from reader (returns `UnverifiedBody`) |
| `DecryptWriter` | `io.Writer` | `(io.WriteCloser, error)` | Streaming decryption to writer (uses `io.Pipe` + `waitCloser`) |
| `Close` | *(none)* | `error` | Releases all managed resources; subsequent calls return `os.ErrClosed` |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >80%)
   - Pass all tests including race detector
   - Use `gofmt`, `golangci-lint` and `gosec`

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
   - Ensure zero race conditions
   - Maintain coverage above 80%

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

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/encoding/opengpg)** — OpenGPG interface, error codes, and Identity struct

### Related golib Packages

- **[github.com/nabbar/golib/errors](https://pkg.go.dev/github.com/nabbar/golib/errors)** — Standardized error code registry used by the error constants in this package
- **[github.com/nabbar/golib/ioutils/mapCloser](https://pkg.go.dev/github.com/nabbar/golib/ioutils/mapCloser)** — Deterministic resource tracker used for pipe end cleanup

### External References

- **[ProtonMail/go-crypto/openpgp](https://pkg.go.dev/github.com/ProtonMail/go-crypto/openpgp)** — Underlying OpenPGP V6 implementation providing key generation, packet serialization, and armor encoding
- **[FIPS 203 — ML-DSA](https://csrc.nist.gov/pubs/fips/203/final)** — NIST Federal Standard defining the ML-DSA signatures used in the KEM hybrid with EdDSA
- **[FIPS 204 — ML-DSA Key Management](https://csrc.nist.gov/pubs/fips/204/final)** — Key management and storage requirements for ML-DSA
- **[RFC 9580 — The OpenPGP Message Format](https://datatracker.ietf.org/doc/html/rfc9580)** — Message format specification for OpenPGP

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020-2026 Nicolas JUHEL