# golib

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue)](https://golang.org/)
[![Go](https://github.com/nabbar/golib/workflows/Go/badge.svg)](https://github.com/nabbar/golib/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/nabbar/golib)](https://pkg.go.dev/github.com/nabbar/golib)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabbar/golib)](https://goreportcard.com/report/github.com/nabbar/golib)
[![Known Vulnerabilities](https://snyk.io/test/github/nabbar/golib/badge.svg)](https://snyk.io/test/github/nabbar/golib)

Comprehensive Go library collection providing production-ready packages for common development needs including AWS integration, HTTP servers/clients, logging, monitoring, archive management, and more.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Package Overview](#package-overview)
  - [Cloud & Infrastructure](#cloud--infrastructure)
  - [Web & Networking](#web--networking)
  - [Data Management](#data-management)
  - [Utilities & Helpers](#utilities--helpers)
  - [Monitoring & Logging](#monitoring--logging)
  - [Development Tools](#development-tools)
- [Quick Start](#quick-start)
- [Testing](#testing)
- [Build Configuration](#build-configuration)
- [Contributing](#contributing)
- [Resources](#resources)
- [License](#license)

---

## Overview

The **golib** repository provides a collection of well-tested, production-ready Go packages designed to accelerate application development. Each package is independently usable, thoroughly documented, and follows Go best practices.

### Design Philosophy

1. **Modularity**: Each package is self-contained and can be used independently
2. **Production-Ready**: Comprehensive testing with high coverage (average >80%)
3. **Performance**: Zero-allocation designs where applicable, optimized for throughput
4. **Thread-Safety**: Race detector validated concurrent operations
5. **Documentation**: Detailed README and testing guides for each package
6. **Standards Compliance**: Following Go conventions and industry best practices

### Statistics

- **Total Packages**: 38+
- **Test Specifications**: 9,183+
- **Average Coverage**: >80%
- **Go Version**: 1.22+
- **Race Detection**: ✅ All packages validated

---

## Key Features

- **Cloud Integration**: AWS S3, IAM, and service clients with MinIO support
- **HTTP Stack**: Production-grade HTTP servers and clients with advanced features
- **Archive Management**: Streaming TAR/ZIP with GZIP, BZIP2, LZ4, XZ compression
- **Logging**: Structured logging with multiple outputs (file, syslog, stdout/stderr)
- **Monitoring**: Prometheus integration, system metrics, health checks
- **Error Handling**: Advanced error types with codes, tracing, and hierarchies
- **Database**: GORM integration with MySQL, PostgreSQL, SQLite, SQL Server
- **Security**: Certificate management, password generation, LDAP authentication
- **Concurrency**: Atomic operations, semaphores, thread-safe collections
- **Configuration**: Viper integration with multiple format support

---

## Installation

```bash
# Install all packages
go get github.com/nabbar/golib/...

# Install specific package
go get github.com/nabbar/golib/logger
go get github.com/nabbar/golib/archive
go get github.com/nabbar/golib/httpserver
```

### Requirements

- Go 1.22 or higher
- CGO enabled for race detection tests
- Platform: Linux, macOS, Windows

---

## Package Overview

### Cloud & Infrastructure

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[aws](aws/README.md)** | AWS SDK wrapper for S3, IAM with MinIO support | Comprehensive | [README](aws/README.md) |
| **[artifact](artifact/README.md)** | Artifact management for GitHub, GitLab, JFrog | Comprehensive | [README](artifact/README.md) |
| **[static](static/README.md)** | Static file serving and management | 85.6% | [README](static/README.md) |

### Web & Networking

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[httpserver](httpserver/README.md)** | Production HTTP server with TLS, pooling, graceful shutdown | High | [README](httpserver/README.md) |
| **[httpcli](httpcli/README.md)** | Advanced HTTP client with DNS mapping, retry logic | High | [README](httpcli/README.md) |
| **[router](router/README.md)** | Gin router extensions with auth, headers | 91.4% | [README](router/README.md) |
| **[network](network/README.md)** | Network utilities, protocol handling | 98.7% | [README](network/README.md) |
| **[socket](socket/README.md)** | TCP, UDP, Unix socket clients and servers | 70-85% | [README](socket/README.md) |
| **[request](request/README.md)** | HTTP request builders and utilities | - | - |

### Data Management

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[database](database/README.md)** | GORM integration with multiple drivers, KV store | High | [README](database/README.md) |
| **[cache](cache/README.md)** | In-memory caching with TTL and eviction | High | [README](cache/README.md) |
| **[archive](archive/README.md)** | TAR, ZIP with streaming compression (GZIP, BZIP2, LZ4, XZ) | ≥80% | [README](archive/README.md) |
| **[config](config/README.md)** | Component-based configuration management | High | [README](config/README.md) |
| **[viper](viper/README.md)** | Viper integration with cleaners and helpers | 73.3% | [README](viper/README.md) |

### Utilities & Helpers

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[errors](errors/README.md)** | Error handling with codes, tracing, pools | >90% | [README](errors/README.md) |
| **[atomic](atomic/README.md)** | Generic atomic primitives, thread-safe maps | >95% | [README](atomic/README.md) |
| **[size](size/README.md)** | Byte size parsing, formatting, arithmetic | 95.4% | [README](size/README.md) |
| **[version](version/README.md)** | Semantic versioning and comparison | 93.8% | [README](version/README.md) |
| **[duration](duration/README.md)** | Extended duration types with big.Int support | High | [README](duration/README.md) |
| **[encoding](encoding/README.md)** | AES encryption, hex encoding, random generation | High | [README](encoding/README.md) |
| **[ioutils](ioutils/README.md)** | IO utilities: buffers, delimiters, progress | High | [README](ioutils/README.md) |
| **[file](file/README.md)** | File operations with bandwidth control, permissions | High | [README](file/README.md) |
| **[context](context/README.md)** | Context helpers and Gin integration | High | [README](context/README.md) |

### Monitoring & Logging

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[logger](logger/README.md)** | Structured logging with multiple hooks (file, syslog, etc.) | 74.7% | [README](logger/README.md) |
| **[monitor](monitor/README.md)** | System monitoring, health checks, status reporting | 88.5% | [README](monitor/README.md) |
| **[prometheus](prometheus/README.md)** | Prometheus metrics, bloom filters, pools | 90.9% | [README](prometheus/README.md) |
| **[status](status/README.md)** | Health status management and control | 85.6% | [README](status/README.md) |

### Development Tools

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[cobra](cobra/README.md)** | Cobra CLI framework extensions | High | [README](cobra/README.md) |
| **[console](console/README.md)** | Terminal output with colors, formatting | 60.9% | [README](console/README.md) |
| **[shell](shell/README.md)** | Shell command helpers and prompts | - | - |
| **[retro](retro/README.md)** | Retro-compatibility utilities | 84.2% | [README](retro/README.md) |
| **[pprof](pprof/README.md)** | Profiling utilities | - | - |

### Security & Communication

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[certificates](certificates/README.md)** | TLS certificate management, CA operations | High | [README](certificates/README.md) |
| **[password](password/README.md)** | Secure password generation with complexity rules | 84.6% | [README](password/README.md) |
| **[ldap](ldap/README.md)** | LDAP client and authentication | - | [README](ldap/README.md) |
| **[oauth](oauth/README.md)** | OAuth client implementation | - | - |
| **[mail](mail/README.md)** | Email composition and sending | High | [README](mail/README.md) |
| **[mailer](mailer/README.md)** | Email service abstraction | High | [README](mailer/README.md) |
| **[mailPooler](mailPooler/README.md)** | Pooled email sending | High | [README](mailPooler/README.md) |
| **[smtp](smtp/README.md)** | SMTP client with TLS support | High | - |
| **[ftpclient](ftpclient/README.md)** | FTP client implementation | High | [README](ftpclient/README.md) |
| **[nats](nats/README.md)** | NATS messaging client | - | - |

### Concurrency & Control

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[semaphore](semaphore/README.md)** | Semaphores with progress bars | 98%+ | [README](semaphore/README.md) |
| **[runner](runner/README.md)** | Background task runners with start/stop, ticker | 88-90% | [README](runner/README.md) |
| **[pidcontroller](pidcontroller/README.md)** | PID controller implementation | - | - |

---

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/version"
    "github.com/nabbar/golib/archive"
)

func main() {
    // Version management
    v, _ := version.Parse("1.2.3")
    if v.GreaterThan(version.Must("1.0.0")) {
        println("Version is newer")
    }
    
    // Logging
    log := logger.New()
    log.Info("Application started")
    
    // Archive extraction
    file, _ := os.Open("archive.tar.gz")
    defer file.Close()
    archive.ExtractAll(file, "archive.tar.gz", "/output")
}
```

### HTTP Server Example

```go
import (
    "github.com/nabbar/golib/httpserver/pool"
    "github.com/gin-gonic/gin"
)

func main() {
    config := pool.Config{
        ListenAddress: ":8080",
        TLS: pool.TLSConfig{
            Enable: true,
            CertFile: "cert.pem",
            KeyFile:  "key.pem",
        },
    }
    
    server := pool.New(config)
    router := gin.Default()
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    server.Start(router)
}
```

### AWS S3 Example

```go
import "github.com/nabbar/golib/aws"

func main() {
    client := aws.NewS3Client(aws.Config{
        Region:    "us-east-1",
        AccessKey: "...",
        SecretKey: "...",
    })
    
    // Upload file
    client.PutObject("bucket", "key", reader)
    
    // Download file
    data, _ := client.GetObject("bucket", "key")
}
```

---

## Testing

All packages include comprehensive test suites using Ginkgo v2 and Gomega.

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Statistics

- **Total Test Specs**: 9,183+
- **Coverage**: >80% average
- **Race Detection**: ✅ Zero data races
- **Test Framework**: Ginkgo v2 + Gomega
- **Test Duration**: ~2-3 minutes (without race), ~5-6 minutes (with race)

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

---

## Build Configuration

### Static Builds

For static, pure Go binaries, use build tags:

```bash
# Static build with no CGO dependencies
go build -a -tags "osusergo netgo" -installsuffix cgo -ldflags '-w -s' .

# With version information
go build -ldflags "-X main.Version=1.0.0 -X main.Build=$(git rev-parse HEAD)" .
```

### Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build .

# Windows
GOOS=windows GOARCH=amd64 go build .

# macOS
GOOS=darwin GOARCH=arm64 go build .
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Code Guidelines

- **Do not use AI** to generate package implementation code
- **AI may assist** with tests, documentation, and bug fixing only
- All code must pass `go test -race` with zero data races
- Maintain or improve test coverage (target: ≥80%)
- Follow existing code style and patterns
- Add comprehensive tests for new features

### Documentation

- Update README.md for new features
- Add code examples for common use cases
- Keep TESTING.md synchronized
- Use English for all documentation and comments

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request with:
   - Clear description of changes
   - Test results (including race detection)
   - Updated documentation
   - Reference to related issues

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

- **Repository**: [GitHub](https://github.com/nabbar/golib)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Security**: [SECURITY.md](SECURITY.md)

### External References

- [Go Documentation](https://golang.org/doc/)
- [Ginkgo Testing Framework](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

**Maintained by**: golib Contributors  
**Version**: Go 1.21+ on Linux, macOS, Windows
