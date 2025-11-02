# Certificates Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Comprehensive TLS/SSL certificate management for secure communications in Go applications.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Subpackages](#subpackages)
- [Configuration](#configuration)
- [Security Best Practices](#security-best-practices)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Contributing](#contributing)
- [Resources](#resources)
- [License](#license)

---

## Overview

The certificates package provides a complete solution for configuring TLS/SSL connections in Go applications. It offers type-safe configuration for certificates, cipher suites, elliptic curves, TLS versions, and client authentication modes.

### Design Philosophy

1. **Type-Safe**: Leverage Go generics and type wrappers for compile-time safety
2. **Flexible Input**: Support for PEM strings, file paths, and structured configuration
3. **Security-First**: Default to secure configurations with modern TLS standards
4. **Multi-Format**: JSON, YAML, TOML, and CBOR encoding support
5. **Thread-Safe**: All operations are safe for concurrent access

---

## Key Features

- **Certificate Management**: Load and manage certificate pairs (private key + certificate)
- **CA Management**: Support for root CA and client CA certificate pools
- **TLS Version Control**: Configure minimum and maximum TLS versions (1.0-1.3)
- **Cipher Suite Selection**: Modern, secure cipher suites for TLS 1.2 and 1.3
- **Elliptic Curve Configuration**: Support for X25519, P256, P384, and P521
- **Client Authentication**: Five authentication modes from none to strict verification
- **Dynamic Configuration**: Runtime configuration updates and rotation
- **Multiple Encodings**: JSON, YAML, TOML, CBOR support for all types
- **Thread-Safe Operations**: Concurrent access protection throughout

---

## Installation

```bash
go get github.com/nabbar/golib/certificates
```

**Requirements:**
- Go 1.18 or higher (for generics support)
- No external dependencies beyond crypto/tls and encoding libraries

---

## Architecture

### Package Structure

```
certificates/
├── certificates        # Main package
│   ├── interface.go   # TLSConfig interface and types
│   ├── model.go       # Implementation
│   ├── config.go      # Configuration structures
│   └── tools.go       # Helper functions
└── Subpackages/
    ├── auth/          # Client authentication modes
    ├── ca/            # Certificate Authority management
    ├── certs/         # Certificate pair management
    ├── cipher/        # Cipher suite configuration
    ├── curves/        # Elliptic curve configuration
    └── tlsversion/    # TLS version management
```

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│              TLSConfig Interface                │
│   Main configuration for TLS connections        │
└───────────┬─────────────────────────────────────┘
            │
            ├──> Root CA Pool (ca.Cert)
            │    └─ x509.CertPool
            │
            ├──> Client CA Pool (ca.Cert)
            │    └─ x509.CertPool
            │
            ├──> Certificate Pairs (certs.Cert)
            │    └─ tls.Certificate
            │
            ├──> TLS Version (tlsversion.Version)
            │    ├─ Min: TLS 1.2 (recommended)
            │    └─ Max: TLS 1.3 (preferred)
            │
            ├──> Cipher Suites (cipher.Cipher)
            │    ├─ TLS 1.2: ECDHE+AES-GCM
            │    └─ TLS 1.3: AES-GCM, ChaCha20
            │
            ├──> Elliptic Curves (curves.Curves)
            │    ├─ X25519 (preferred)
            │    └─ P256, P384, P521
            │
            └──> Client Auth (auth.ClientAuth)
                 └─ NoClientCert, Request, Require, Verify, Strict
```

### Type System

| Type | Package | Purpose |
|------|---------|---------|
| `TLSConfig` | certificates | Main interface for TLS configuration |
| `ClientAuth` | auth | Client authentication modes |
| `Cert` (CA) | ca | Certificate Authority certificates |
| `Cert` (pairs) | certs | Certificate pairs (key + cert) |
| `Cipher` | cipher | TLS cipher suite identifiers |
| `Curves` | curves | Elliptic curve identifiers |
| `Version` | tlsversion | TLS protocol version |

---

## Quick Start

### Basic Server Configuration

```go
package main

import (
    "crypto/tls"
    "net/http"
    
    "github.com/nabbar/golib/certificates"
    "github.com/nabbar/golib/certificates/tlsversion"
)

func main() {
    // Create TLS configuration
    tlsConfig := certificates.New()
    
    // Set TLS versions
    tlsConfig.SetVersionMin(tlsversion.VersionTLS12)
    tlsConfig.SetVersionMax(tlsversion.VersionTLS13)
    
    // Add server certificate
    err := tlsConfig.AddCertificatePairFile("/path/to/key.pem", "/path/to/cert.pem")
    if err != nil {
        panic(err)
    }
    
    // Create HTTP server with TLS
    server := &http.Server{
        Addr:      ":443",
        TLSConfig: tlsConfig.TLS("example.com"),
    }
    
    server.ListenAndServeTLS("", "")
}
```

### Client Configuration with mTLS

```go
package main

import (
    "crypto/tls"
    "net/http"
    
    "github.com/nabbar/golib/certificates"
    "github.com/nabbar/golib/certificates/auth"
)

func main() {
    // Create client TLS configuration
    tlsConfig := certificates.New()
    
    // Add root CA to verify server
    err := tlsConfig.AddRootCAFile("/path/to/ca.pem")
    if err != nil {
        panic(err)
    }
    
    // Add client certificate for mTLS
    err = tlsConfig.AddCertificatePairFile("/path/to/client-key.pem", "/path/to/client-cert.pem")
    if err != nil {
        panic(err)
    }
    
    // Create HTTP client
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: tlsConfig.TLS("server.example.com"),
        },
    }
    
    resp, err := client.Get("https://server.example.com")
    // ...
}
```

### Configuration from Strings

```go
// PEM-encoded certificate and key
keyPEM := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----`

certPEM := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJ...
-----END CERTIFICATE-----`

tlsConfig := certificates.New()
err := tlsConfig.AddCertificatePairString(keyPEM, certPEM)
if err != nil {
    panic(err)
}
```

---

## Subpackages

###  auth - Client Authentication Modes

Provides client authentication mode types for TLS connections.

**Supported Modes:**
- `NoClientCert`: No client certificate required
- `RequestClientCert`: Request but don't require client certificate
- `RequireAnyClientCert`: Require any client certificate (unverified)
- `VerifyClientCertIfGiven`: Verify client certificate if provided
- `RequireAndVerifyClientCert`: Require and verify client certificate

**Example:**
```go
import "github.com/nabbar/golib/certificates/auth"

authMode := auth.Parse("require")
tlsConfig.SetClientAuth(authMode)
```

[Full auth package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/auth)

---

### ca - Certificate Authority Management

Manages CA certificates for verifying certificate chains.

**Key Features:**
- Parse CA certificates from PEM strings or bytes
- Support for certificate chains
- Convert to x509.CertPool for TLS
- Multiple encoding formats (JSON, YAML, TOML, CBOR)

**Example:**
```go
import "github.com/nabbar/golib/certificates/ca"

caCert, err := ca.Parse(pemString)
if err != nil {
    log.Fatal(err)
}
pool := caCert.GetCertPool()
```

[Full ca package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/ca)

---

### certs - Certificate Pair Management

Manages certificate pairs (private key + certificate) for TLS servers and clients.

**Key Features:**
- Parse certificate pairs from PEM strings or files
- Support for certificate chains
- Multiple configuration formats (ConfigPair, ConfigChain)
- Convert to tls.Certificate

**Example:**
```go
import "github.com/nabbar/golib/certificates/certs"

cert, err := certs.Parse(keyPEM + "\n" + certPEM)
if err != nil {
    log.Fatal(err)
}
tlsCert := cert.GetTLS()
```

[Full certs package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/certs)

---

### cipher - Cipher Suite Selection

Provides TLS cipher suite types and parsing for secure connections.

**Supported Cipher Suites:**

**TLS 1.2:**
- RSA with AES-GCM
- ECDHE-RSA with AES-GCM (forward secrecy)
- ECDHE-ECDSA with AES-GCM (forward secrecy)
- ECDHE with ChaCha20-Poly1305 (forward secrecy, mobile-optimized)

**TLS 1.3:**
- AES-128-GCM-SHA256
- AES-256-GCM-SHA384
- ChaCha20-Poly1305-SHA256

**Example:**
```go
import "github.com/nabbar/golib/certificates/cipher"

cipher := cipher.Parse("ECDHE-RSA-AES128-GCM-SHA256")
if cipher != cipher.Unknown {
    fmt.Println("Supported cipher:", cipher.String())
}
```

[Full cipher package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/cipher)

---

### curves - Elliptic Curve Configuration

Provides elliptic curve types for ECDHE cipher suites.

**Supported Curves:**
- `X25519`: Modern, high-performance (preferred)
- `P256` (secp256r1): NIST curve, widely supported
- `P384` (secp384r1): NIST curve, higher security
- `P521` (secp521r1): NIST curve, maximum security

**Example:**
```go
import "github.com/nabbar/golib/certificates/curves"

curve := curves.Parse("X25519")
tlsConfig.AddCurves(curve)
```

[Full curves package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/curves)

---

### tlsversion - TLS Version Management

Provides TLS protocol version types and management.

**Supported Versions:**
- `VersionTLS10`: TLS 1.0 (deprecated, not recommended)
- `VersionTLS11`: TLS 1.1 (deprecated, not recommended)
- `VersionTLS12`: TLS 1.2 (secure, widely supported)
- `VersionTLS13`: TLS 1.3 (modern, most secure)

**Example:**
```go
import "github.com/nabbar/golib/certificates/tlsversion"

minVer := tlsversion.Parse("1.2")
maxVer := tlsversion.Parse("1.3")
tlsConfig.SetVersionMin(minVer)
tlsConfig.SetVersionMax(maxVer)
```

[Full tlsversion package documentation →](https://pkg.go.dev/github.com/nabbar/golib/certificates/tlsversion)

---

## Configuration

### TLSConfig Interface

The main `TLSConfig` interface provides comprehensive methods for configuring TLS connections:

**Certificate Management:**
- `AddCertificatePairString(key, cert string) error`
- `AddCertificatePairFile(keyFile, certFile string) error`
- `GetCertificatePair() []tls.Certificate`
- `LenCertificatePair() int`
- `CleanCertificatePair()`

**Root CA Management:**
- `AddRootCA(rootCA ca.Cert) bool`
- `AddRootCAString(rootCA string) bool`
- `AddRootCAFile(pemFile string) error`
- `GetRootCA() []ca.Cert`
- `GetRootCAPool() *x509.CertPool`

**Client CA Management:**
- `AddClientCAString(ca string) bool`
- `AddClientCAFile(pemFile string) error`
- `GetClientCA() []ca.Cert`
- `GetClientCAPool() *x509.CertPool`
- `SetClientAuth(auth.ClientAuth)`

**Version Control:**
- `SetVersionMin(tlsversion.Version)`
- `GetVersionMin() tlsversion.Version`
- `SetVersionMax(tlsversion.Version)`
- `GetVersionMax() tlsversion.Version`

**Cipher & Curve Configuration:**
- `SetCipherList([]cipher.Cipher)`
- `AddCiphers(...cipher.Cipher)`
- `GetCiphers() []cipher.Cipher`
- `SetCurveList([]curves.Curves)`
- `AddCurves(...curves.Curves)`
- `GetCurves() []curves.Curves`

**Advanced Options:**
- `RegisterRand(io.Reader)` - Custom randomness source
- `SetDynamicSizingDisabled(bool)` - Control record sizing
- `SetSessionTicketDisabled(bool)` - Control session resumption
- `TLS(serverName string) *tls.Config` - Get final tls.Config

### Configuration Examples

**Minimal Server Configuration:**
```go
cfg := certificates.New()
cfg.AddCertificatePairFile("server-key.pem", "server-cert.pem")
tlsConfig := cfg.TLS("example.com")
```

**Strict Server with mTLS:**
```go
cfg := certificates.New()
cfg.SetVersionMin(tlsversion.VersionTLS12)
cfg.SetVersionMax(tlsversion.VersionTLS13)
cfg.AddCertificatePairFile("server-key.pem", "server-cert.pem")
cfg.AddClientCAFile("client-ca.pem")
cfg.SetClientAuth(auth.RequireAndVerifyClientCert)
tlsConfig := cfg.TLS("example.com")
```

**Client with Custom CA:**
```go
cfg := certificates.New()
cfg.AddRootCAFile("custom-ca.pem")
cfg.AddCertificatePairFile("client-key.pem", "client-cert.pem")
tlsConfig := cfg.TLS("")
```

---

## Security Best Practices

### TLS Version Selection

**✅ Recommended Configuration:**
```go
cfg.SetVersionMin(tlsversion.VersionTLS12)  // Minimum TLS 1.2
cfg.SetVersionMax(tlsversion.VersionTLS13)  // Maximum TLS 1.3
```

**Security Rationale:**
- TLS 1.0 and 1.1 are deprecated (RFC 8996)
- TLS 1.2 provides wide compatibility
- TLS 1.3 offers improved security and performance

### Cipher Suite Selection

**✅ Prefer ECDHE cipher suites for forward secrecy:**
```go
cipherSuites := []cipher.Cipher{
    cipher.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    cipher.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    cipher.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
    cipher.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
}
cfg.SetCipherList(cipherSuites)
```

**❌ Avoid:**
- Non-ECDHE cipher suites (no forward secrecy)
- Legacy cipher suites (RC4, 3DES, MD5)
- Export-grade cryptography

### Elliptic Curve Selection

**✅ Recommended:**
```go
cfg.AddCurves(
    curves.X25519,  // Modern, fast, secure (preferred)
    curves.P256,    // NIST, widely supported
)
```

**Security Notes:**
- X25519 offers best performance and security
- P256 provides broad compatibility
- Avoid P384/P521 unless required by policy

### Certificate Management

**✅ Best Practices:**
- Use strong key sizes (RSA 2048+, ECDSA P-256+)
- Implement certificate rotation
- Monitor certificate expiration
- Use proper file permissions (0600 for private keys)
- Store private keys securely (HSM, vault)

**Example with Rotation:**
```go
func rotateCertificate(cfg certificates.TLSConfig) error {
    // Load new certificate
    err := cfg.AddCertificatePairFile("new-key.pem", "new-cert.pem")
    if err != nil {
        return err
    }
    
    // Remove old certificates
    cfg.CleanCertificatePair()
    
    return nil
}
```

### Client Authentication

**Security Levels:**

| Mode | Use Case | Security |
|------|----------|----------|
| `NoClientCert` | Public services | Low |
| `RequestClientCert` | Optional auth | Medium |
| `RequireAnyClientCert` | Testing | Medium |
| `VerifyClientCertIfGiven` | Flexible auth | Medium-High |
| `RequireAndVerifyClientCert` | mTLS, high security | High |

**✅ For high-security environments:**
```go
cfg.SetClientAuth(auth.RequireAndVerifyClientCert)
cfg.AddClientCAFile("trusted-clients-ca.pem")
```

---

## Use Cases

### HTTPS Web Server

```go
package main

import (
    "net/http"
    "github.com/nabbar/golib/certificates"
    "github.com/nabbar/golib/certificates/tlsversion"
)

func main() {
    // Configure TLS
    tlsCfg := certificates.New()
    tlsCfg.SetVersionMin(tlsversion.VersionTLS12)
    tlsCfg.AddCertificatePairFile("server.key", "server.crt")
    
    // Create HTTPS server
    server := &http.Server{
        Addr:      ":443",
        TLSConfig: tlsCfg.TLS("example.com"),
        Handler:   http.DefaultServeMux,
    }
    
    server.ListenAndServeTLS("", "")
}
```

### Microservice with mTLS

```go
// Server side
serverCfg := certificates.New()
serverCfg.AddCertificatePairFile("service.key", "service.crt")
serverCfg.AddClientCAFile("clients-ca.pem")
serverCfg.SetClientAuth(auth.RequireAndVerifyClientCert)

// Client side
clientCfg := certificates.New()
clientCfg.AddRootCAFile("services-ca.pem")
clientCfg.AddCertificatePairFile("client.key", "client.crt")

client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: clientCfg.TLS("service.example.com"),
    },
}
```

### gRPC Service

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "github.com/nabbar/golib/certificates"
)

tlsCfg := certificates.New()
tlsCfg.AddCertificatePairFile("grpc.key", "grpc.crt")
tlsCfg.AddRootCAFile("ca.pem")

creds := credentials.NewTLS(tlsCfg.TLS("grpc.example.com"))
server := grpc.NewServer(grpc.Creds(creds))
```

### Database Connection

```go
import (
    "database/sql"
    "crypto/tls"
    "github.com/go-sql-driver/mysql"
    "github.com/nabbar/golib/certificates"
)

tlsCfg := certificates.New()
tlsCfg.AddRootCAFile("db-ca.pem")
tlsCfg.AddCertificatePairFile("client.key", "client.crt")

mysql.RegisterTLSConfig("custom", tlsCfg.TLS(""))
db, err := sql.Open("mysql", "user:pass@tcp(host:3306)/db?tls=custom")
```

---

## API Reference

### Main Types

**TLSConfig** - Main interface for TLS configuration
```go
type TLSConfig interface {
    // Certificate management
    AddCertificatePairString(key, crt string) error
    AddCertificatePairFile(keyFile, crtFile string) error
    GetCertificatePair() []tls.Certificate
    
    // CA management  
    AddRootCAString(rootCA string) bool
    AddRootCAFile(pemFile string) error
    GetRootCAPool() *x509.CertPool
    
    // Version control
    SetVersionMin(v tlsversion.Version)
    SetVersionMax(v tlsversion.Version)
    
    // Generate final config
    TLS(serverName string) *tls.Config
}
```

### Factory Functions

**New()** - Create new TLSConfig
```go
func New() TLSConfig
```

### Subpackage Types

See individual subpackage documentation for detailed type information:
- [auth.ClientAuth](https://pkg.go.dev/github.com/nabbar/golib/certificates/auth)
- [ca.Cert](https://pkg.go.dev/github.com/nabbar/golib/certificates/ca)
- [certs.Cert](https://pkg.go.dev/github.com/nabbar/golib/certificates/certs)
- [cipher.Cipher](https://pkg.go.dev/github.com/nabbar/golib/certificates/cipher)
- [curves.Curves](https://pkg.go.dev/github.com/nabbar/golib/certificates/curves)
- [tlsversion.Version](https://pkg.go.dev/github.com/nabbar/golib/certificates/tlsversion)

---

## Testing

**Test Suite**: Ginkgo v2 + Gomega with comprehensive coverage

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -r
```

**Coverage by Package:**

| Package | Coverage | Specs |
|---------|----------|-------|
| certificates | ~70% | 15 |
| auth | 73.0% | 12 |
| ca | 68.5% | 18 |
| certs | 47.8% | 9 |
| cipher | 50.6% | 12 |
| curves | 50.5% | 9 |
| tlsversion | 54.5% | 9 |

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions:**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Follow existing code style and patterns
- Add tests for new features

**Documentation:**
- Update README.md for new features
- Add examples for common use cases
- Keep subpackage documentation synchronized

**Security:**
- Report security issues privately
- Follow responsible disclosure practices
- Use secure defaults in new features

**Pull Requests:**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Documentation:**
- [Go crypto/tls Package](https://pkg.go.dev/crypto/tls)
- [Go crypto/x509 Package](https://pkg.go.dev/crypto/x509)
- [RFC 5246 - TLS 1.2](https://tools.ietf.org/html/rfc5246)
- [RFC 8446 - TLS 1.3](https://tools.ietf.org/html/rfc8446)
- [RFC 8996 - Deprecating TLS 1.0 and 1.1](https://tools.ietf.org/html/rfc8996)

**Tools:**
- [SSL Labs Server Test](https://www.ssllabs.com/ssltest/)
- [testssl.sh](https://testssl.sh/)
- [OpenSSL](https://www.openssl.org/)

**Package Links:**
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/certificates)
- [GitHub Repository](https://github.com/nabbar/golib)
- [Testing Documentation](TESTING.md)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2020 Nicolas JUHEL

---

**Last Updated**: 2025-11-07
