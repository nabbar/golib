# golib

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://golang.org/)
[![Go](https://github.com/nabbar/golib/workflows/Go/badge.svg)](https://github.com/nabbar/golib/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/nabbar/golib)](https://pkg.go.dev/github.com/nabbar/golib)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabbar/golib)](https://goreportcard.com/report/github.com/nabbar/golib)
[![Known Vulnerabilities](https://snyk.io/test/github/nabbar/golib/badge.svg)](https://snyk.io/test/github/nabbar/golib)

Comprehensive Go library collection providing production-ready packages for cloud services, web infrastructure, data management, security, monitoring, and development utilities.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Package Catalog](#package-catalog)
  - [Cloud & Infrastructure](#cloud--infrastructure)
  - [Web & Networking](#web--networking)
  - [Data Management](#data-management)
  - [Security & Communication](#security--communication)
  - [Monitoring & Logging](#monitoring--logging)
  - [Utilities & Helpers](#utilities--helpers)
  - [Concurrency & Control](#concurrency--control)
  - [Development Tools](#development-tools)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Build Configuration](#build-configuration)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Resources](#resources)

---

## Overview

**golib** is a comprehensive collection of production-ready Go packages designed to accelerate application development. Each package is independently usable, thoroughly tested, and follows Go best practices with a focus on performance, thread safety, and observability.

### Design Philosophy

1. **Modularity**: Self-contained packages with minimal dependencies
2. **Production-Ready**: Comprehensive testing with high coverage (≥80% average)
3. **Performance-First**: Streaming operations, zero-allocation paths, optimized throughput
4. **Thread-Safe**: All concurrent operations validated with race detector
5. **Observable**: Structured logging, metrics, health checks, and monitoring
6. **Standards-Compliant**: Go idioms, semantic versioning, standard interfaces

### Repository Statistics

- **Packages**: 50+ specialized packages
- **Test Specs**: 10,145 test specifications (Ginkgo v2)
- **Test Files**: 677 test files
- **Test Packages**: 156 tested packages
- **Go Version**: 1.25+
- **Platforms**: Linux, macOS, Windows
- **CI/CD**: GitHub Actions with race detection
- **Thread Safety**: ✅ Race detector validated

---

## Key Features

- **Cloud Services**: AWS S3/IAM with MinIO compatibility, artifact management (GitHub/GitLab/JFrog)
- **Web Infrastructure**: HTTP servers with TLS, pooling, graceful shutdown; advanced HTTP clients with DNS mapping
- **Archive & Compression**: Streaming TAR/ZIP, multiple algorithms (GZIP, BZIP2, LZ4, XZ), constant memory usage
- **Structured Logging**: Multi-output logging (file, syslog, stdout/stderr) with field injection and rotation
- **Monitoring & Metrics**: Prometheus integration, system info collection, health status management
- **Error Management**: Advanced errors with codes, stack tracing, hierarchies, and thread-safe pools
- **Data Persistence**: GORM integration (MySQL, PostgreSQL, SQLite, SQL Server), in-memory cache with TTL
- **Security Suite**: TLS certificate management, password generation, LDAP authentication, OAuth clients
- **Concurrency Primitives**: Generic atomic types, semaphores with progress, thread-safe maps and values
- **Type Safety**: Semantic versioning, byte size arithmetic, duration extensions, network protocol handling
- **Development Tools**: Cobra CLI extensions, terminal formatting, configuration management (Viper)
- **Communication**: Email with SMTP pooling and queuing, FTP clients, NATS messaging, socket servers

---

## Installation

```bash
# Install entire library
go get github.com/nabbar/golib/...

# Install specific packages
go get github.com/nabbar/golib/logger
go get github.com/nabbar/golib/archive
go get github.com/nabbar/golib/httpserver
go get github.com/nabbar/golib/atomic
go get github.com/nabbar/golib/errors
```

### Requirements

- **Go**: 1.25 or higher
- **CGO**: Required for race detection (`CGO_ENABLED=1`)
- **Build Tools**: gcc/clang for race detector
- **Platforms**: Linux, macOS, Windows (amd64, arm64, 386)

---

## Architecture

### Library Organization

The golib library is organized into domain-specific packages with clear separation of concerns:

```
golib/
├─ Cloud & Infrastructure   # AWS, artifacts, static files
├─ Web & Networking         # HTTP servers/clients, routing, sockets
├─ Data Management          # Databases, caching, archives, config
├─ Security & Communication # Certificates, auth, email, messaging
├─ Monitoring & Logging     # Structured logs, metrics, health checks
├─ Utilities & Helpers      # Errors, atomics, types, IO
├─ Concurrency & Control    # Semaphores, runners, PID controllers
└─ Development Tools        # CLI, console, profiling
```

### Design Patterns

**Streaming Architecture**
- Constant memory usage via `io.Reader`/`io.Writer` interfaces
- Zero-copy operations for uncompressed data
- Chunked processing for arbitrarily large datasets
- Example: Extract 10GB archive using only 10MB RAM

**Thread-Safe Operations**
- `sync/atomic` primitives for lock-free counters and flags
- `sync.Mutex` for protecting shared mutable state
- `sync.WaitGroup` for goroutine lifecycle management
- Validated with race detector across entire test suite

**Observable Systems**
- Structured logging with contextual field injection
- Prometheus metrics with automatic collection
- Health check endpoints for service monitoring
- Status reporting with customizable thresholds

**Error Management Philosophy**
- Errors with numeric codes for programmatic handling
- Automatic stack traces for debugging
- Error hierarchies and parent-child relationships
- Thread-safe error pools for collection

---

## Package Catalog

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
| **request** | HTTP request builders and utilities | - | - |

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
| **shell** | Shell command helpers and prompts | - | - |
| **[retro](retro/README.md)** | Retro-compatibility utilities | 84.2% | [README](retro/README.md) |
| **pprof** | Profiling utilities | - | - |

### Security & Communication

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[certificates](certificates/README.md)** | TLS certificate management, CA operations | High | [README](certificates/README.md) |
| **[password](password/README.md)** | Secure password generation with complexity rules | 84.6% | [README](password/README.md) |
| **[ldap](ldap/README.md)** | LDAP client and authentication | - | [README](ldap/README.md) |
| **oauth** | OAuth client implementation | - | - |
| **[mail](mail/README.md)** | Email composition and sending with SMTP | High | [README](mail/README.md) |
| **[mail/queuer](mail/queuer/README.md)** | Rate-limited SMTP client wrapper | 90.8% | [README](mail/queuer/README.md) |
| **[mail/render](mail/render/README.md)** | HTML email template rendering | High | [README](mail/render/README.md) |
| **[mail/sender](mail/sender/README.md)** | Email composition and sending | High | [README](mail/sender/README.md) |
| **[mail/smtp](mail/smtp/README.md)** | SMTP client with TLS support | High | [README](mail/smtp/README.md) |
| **[ftpclient](ftpclient/README.md)** | FTP client implementation | High | [README](ftpclient/README.md) |
| **nats** | NATS messaging client | - | - |

### Concurrency & Control

| Package | Description | Coverage | Documentation |
|---------|-------------|----------|---------------|
| **[semaphore](semaphore/README.md)** | Semaphores with progress bars | 98%+ | [README](semaphore/README.md) |
| **[runner](runner/README.md)** | Background task runners with start/stop, ticker | 88-90% | [README](runner/README.md) |
| **pidcontroller** | PID controller implementation | - | - |

---

## Performance

### Memory Efficiency

**Streaming Operations**
- Archive extraction: O(1) memory regardless of archive size
- Compression/decompression: Constant buffer usage (~32KB)
- Large file processing: No full file loading required

**Example**: Process 10GB compressed archive using only 10-15MB RAM

### Throughput Benchmarks

| Operation | Throughput | Memory | Package |
|-----------|------------|--------|---------|
| TAR extraction | ~400 MB/s | O(1) | archive |
| ZIP extraction | ~600 MB/s | O(1) | archive |
| GZIP compression | ~150 MB/s | O(1) | archive/compress |
| LZ4 compression | ~800 MB/s | O(1) | archive/compress |
| Atomic operations | ~10M ops/s | Lock-free | atomic |
| Logger writes | ~1M logs/s | Buffered | logger |
| Email queuing | ~1-3K msg/s | Pooled | mail/queuer |

*Benchmarked on: AMD64, Go 1.25, SSD storage*

### Concurrency Performance

- **Zero contention**: Lock-free atomic primitives
- **Parallel safe**: All packages validated with `-race`
- **Efficient pooling**: HTTP servers, SMTP connections, workers
- **Goroutine management**: Proper lifecycle with `sync.WaitGroup`

---

## Use Cases

This library addresses real-world production scenarios:

**Cloud-Native Applications**
- Deploy to AWS/MinIO with S3 integration
- Manage artifacts from CI/CD pipelines (GitHub, GitLab, JFrog)
- Serve static assets with caching and compression
- **Packages**: `aws`, `artifact`, `static`, `httpserver`

**Microservices Architecture**
- HTTP servers with graceful shutdown and health checks
- Structured logging with field injection across services
- Prometheus metrics for observability
- Service discovery and routing
- **Packages**: `httpserver`, `logger`, `prometheus`, `router`, `status`

**Data Processing Pipelines**
- Stream-process large compressed archives
- Transform data without intermediate files
- Parallel processing with semaphores
- Progress tracking for long operations
- **Packages**: `archive`, `ioutils`, `semaphore`, `runner`

**Enterprise Applications**
- LDAP authentication and user management
- GORM integration with multiple databases
- Email notifications with SMTP pooling
- Certificate management and TLS
- **Packages**: `ldap`, `database`, `mail`, `certificates`

**CLI Tools**
- Build command-line applications with Cobra
- Terminal formatting and progress bars
- Configuration management with Viper
- Version management and semantic versioning
- **Packages**: `cobra`, `console`, `viper`, `version`

**Monitoring & Observability**
- Collect system metrics (CPU, memory, disk)
- Health status tracking and alerts
- Prometheus metrics integration
- Structured logging with multiple outputs
- **Packages**: `monitor`, `prometheus`, `logger`, `status`

---

## Quick Start

### Basic Usage

```go
package main

import (
    "os"
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

All packages include comprehensive test suites using **Ginkgo v2** (BDD framework) and **Gomega** (matcher library).

### Quick Test Commands

```bash
# Run all tests
go test -timeout=10m -v ./...

# With coverage
go test -cover -covermode=atomic ./...

# With race detection (REQUIRED before PR)
CGO_ENABLED=1 go test -race -timeout=10m ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Repository Test Statistics

- **Total Packages**: 50+ specialized packages
- **Test Specifications**: 10,145+ specs (Ginkgo v2)
- **Test Files**: 677 test files
- **Tested Packages**: 156 packages
- **High Coverage Packages**: 20+ packages with ≥90% coverage
- **Race Detection**: ✅ Race detector validated
- **Test Duration**: ~5 minutes (standard), ~10 minutes (with race)

### Coverage Highlights

| Category | Coverage Range | Key Packages |
|----------|----------------|--------------|
| **Utilities** | 90-98% | atomic, size, version, errors |
| **Monitoring** | 85-100% | logger, monitor, prometheus, status |
| **Networking** | 70-98% | router, network/protocol, socket |
| **Data** | 73-96% | archive, cache, database, viper |
| **Concurrency** | 88-100% | semaphore, runner, atomic |

### Test Framework Features

- **BDD Style**: Descriptive test specifications
- **Parallel Execution**: Faster test runs with `-p`
- **Race Detection**: Automatic data race detection
- **JUnit Reports**: CI/CD integration ready
- **Benchmarks**: Performance regression testing

**Important**: All contributions must pass race detection tests.

See [TESTING.md](TESTING.md) for comprehensive testing guide including:
- Test framework setup
- Writing tests (best practices and templates)
- Running specific test suites
- Debugging failing tests
- CI integration examples

---

## Best Practices

### Streaming for Large Data

```go
// ✅ Good: Streaming with constant memory
func processArchive(path string) error {
    f, _ := os.Open(path)
    defer f.Close()
    
    return archive.ExtractAll(f, path, "./output")
    // Memory: O(1) regardless of file size
}

// ❌ Bad: Loading entire file
func processArchiveBad(path string) error {
    data, _ := os.ReadFile(path)  // Entire file in RAM!
    return process(data)
}
```

### Always Handle Errors

```go
// ✅ Good: Proper error handling
func parse(input string) (version.Version, error) {
    v, err := version.Parse(input)
    if err != nil {
        return nil, fmt.Errorf("parse version: %w", err)
    }
    return v, nil
}

// ❌ Bad: Ignoring errors
func parseBad(input string) version.Version {
    v, _ := version.Parse(input)  // Silently fails!
    return v
}
```

### Resource Cleanup

```go
// ✅ Good: Defer cleanup immediately
func connect() error {
    client, err := smtp.NewClient(config)
    if err != nil {
        return err
    }
    defer client.Close()  // Guaranteed cleanup
    
    return client.Send(message)
}

// ❌ Bad: Manual cleanup prone to forgetting
func connectBad() error {
    client, _ := smtp.NewClient(config)
    err := client.Send(message)
    client.Close()  // Might not execute if Send panics
    return err
}
```

### Thread-Safe Concurrent Access

```go
// ✅ Good: Using atomic types
import "github.com/nabbar/golib/atomic"

var counter atomic.Value[int]

func increment() {
    counter.Store(counter.Load() + 1)  // Thread-safe
}

// ❌ Bad: Unprotected shared state
var badCounter int

func incrementBad() {
    badCounter++  // Race condition!
}
```

### Structured Logging

```go
// ✅ Good: Contextual fields
log.WithFields(logger.Fields{
    "user_id": userID,
    "action": "login",
    "ip": remoteIP,
}).Info("User logged in")

// ❌ Bad: String concatenation
log.Info("User " + userID + " logged in from " + remoteIP)
```

### Use Context for Cancellation

```go
// ✅ Good: Respect context cancellation
func process(ctx context.Context, data []byte) error {
    for _, item := range data {
        select {
        case <-ctx.Done():
            return ctx.Err()  // Graceful cancellation
        default:
            processItem(item)
        }
    }
    return nil
}
```

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

## Future Enhancements

Potential improvements and features under consideration:

**New Packages**
- gRPC server and client wrappers
- GraphQL utilities and helpers
- Kubernetes client integration
- Message queue abstractions (RabbitMQ, Kafka)
- OpenTelemetry tracing integration
- Redis client with connection pooling

**Archive Enhancements**
- 7-Zip format support
- Zstandard (zstd) compression algorithm
- Brotli compression for web content
- Archive encryption (AES-256-GCM)
- Streaming TAR.GZ (single-pass operations)

**Monitoring & Observability**
- OpenTelemetry exporter
- Distributed tracing integration
- Custom metric types and aggregations
- Alert manager integration
- Performance profiling utilities

**Security**
- OAuth2 server implementation
- JWT token management
- API key management
- Secret vault integration (HashiCorp Vault)
- mTLS certificate rotation

**Data Management**
- Redis cache backend
- MongoDB integration
- ElasticSearch client
- Time-series database support
- Data migration utilities

**Developer Experience**
- Code generation tools
- Configuration validation
- Hot reload for development
- Enhanced debugging utilities
- IDE plugins and integrations

Suggestions and feature requests are welcome via [GitHub Issues](https://github.com/nabbar/golib/issues).

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
**Version**: Go 1.25+ on Linux, macOS, Windows
