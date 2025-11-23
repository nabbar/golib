# golib

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)
[![Go](https://github.com/nabbar/golib/workflows/Go/badge.svg)](https://github.com/nabbar/golib/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/nabbar/golib)](https://pkg.go.dev/github.com/nabbar/golib)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabbar/golib)](https://goreportcard.com/report/github.com/nabbar/golib)
[![Known Vulnerabilities](https://snyk.io/test/github/nabbar/golib/badge.svg)](https://snyk.io/test/github/nabbar/golib)
[![Tests](https://img.shields.io/badge/Tests-10964%20Specs-green)](TESTING.md)
[![Coverage](https://img.shields.io/badge/Coverage-73.9%25-yellow)](TESTING.md)

Comprehensive Go library collection providing production-ready packages for cloud services, web infrastructure, data management, security, monitoring, and development utilities. Built for enterprise-grade applications with extensive testing and documentation.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Repository Statistics](#repository-statistics)
- [Key Features](#key-features)
- [Architecture](#architecture)
  - [Package Organization](#package-organization)
  - [Dependency Model](#dependency-model)
- [Installation](#installation)
- [Package Catalog](#package-catalog)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Build Configuration](#build-configuration)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

**golib** is a comprehensive collection of production-ready Go packages designed to accelerate enterprise application development. Each package is independently usable, thoroughly tested, and follows Go best practices with a focus on performance, thread safety, and observability.

This library provides building blocks for cloud-native applications, web services, data processing, secure communications, production operations, and development productivity.

### Design Philosophy

1. **Modularity**: Self-contained packages with minimal cross-dependencies
2. **Production-Ready**: Comprehensive testing with 10,964 specs across 127 packages
3. **Performance-First**: Streaming operations, zero-allocation paths, optimized throughput
4. **Thread-Safe**: All concurrent operations validated with race detector
5. **Observable**: Structured logging, Prometheus metrics, health checks, monitoring
6. **Standards-Compliant**: Go idioms, semantic versioning, standard library interfaces

### Repository Statistics

```
Total Packages:       165 (127 with tests, 38 utility/types packages)
Documented Packages:  62 packages with individual README.md files
Test Specifications:  10,964
Test Assertions:      21,470
Benchmarks:           92
Pending Tests:        18
Average Coverage:     73.9%
Packages ≥80%:        67/127 (52.8%)
Packages ≥90%:        38/127 (29.9%)
Go Version:           1.24+ (1.25+ recommended)
Platforms:            Linux, macOS, Windows
Thread Safety:        ✅ Zero race conditions
CI/CD:                GitHub Actions with race detection
```

---

## Key Features

### 🌩️ Cloud & Infrastructure
- **AWS Integration**: S3 storage, IAM management, MinIO compatibility ([aws](aws/))
- **Artifact Management**: GitHub, GitLab, JFrog Artifactory, S3 artifact handling ([artifact](artifact/))
- **Archive & Compression**: Streaming TAR/ZIP with GZIP, BZIP2, LZ4, XZ support ([archive](archive/))

### 🌐 Web & Networking
- **HTTP Server**: Pool management with TLS, multi-handler, monitoring, graceful shutdown ([httpserver](httpserver/))
- **HTTP Client**: Advanced client with DNS mapping, retries, timeouts ([httpcli](httpcli/))
- **Router**: Gin-based routing with auth headers, CORS, request validation ([router](router/))
- **Socket Communication**: TCP, UDP, Unix domain socket servers and clients ([socket](socket/))

### 💾 Data Management
- **I/O Utilities**: Stream aggregation, multiplexing, progress tracking, delimiters ([ioutils](ioutils/))
- **Database**: GORM integration for MySQL, PostgreSQL, SQLite, SQL Server ([database](database/))
- **Cache**: In-memory cache with TTL, atomic operations, thread-safe ([cache](cache/))
- **File Operations**: Bandwidth control, progress tracking, permission management ([file](file/))

### 🔒 Security & Communication
- **Certificates**: TLS certificate management, CA operations, cipher suites ([certificates](certificates/))
- **Authentication**: LDAP integration, OAuth clients, password generation ([ldap](ldap/), [oauth](oauth/), [password](password/))
- **Email**: Complete SMTP solution with TLS/STARTTLS, templating, composition, queuing ([mail](mail/))
- **FTP**: FTP client with connection management ([ftpclient](ftpclient/))

### 📊 Monitoring & Logging
- **Structured Logging**: Multi-output (file, syslog, stdout/stderr), field injection, rotation ([logger](logger/))
- **Prometheus Metrics**: Custom metrics, bloom filters, web metrics endpoints ([prometheus](prometheus/))
- **Health Monitoring**: System info, status checks, health endpoints ([monitor](monitor/))
- **Status Management**: Component status tracking, mandatory checks, control ([status](status/))

### 🛠️ Utilities & Helpers
- **Error Management**: Error codes, stack traces, hierarchies, thread-safe pools ([errors](errors/))
- **Encoding**: AES encryption, SHA256 hashing, hex encoding, remote random reader for HSM ([encoding](encoding/))
- **Type Utilities**: Semantic versioning, byte sizes, extended duration with days, protocols ([version](version/), [size](size/), [duration](duration/), [network](network/))
- **Configuration**: Component lifecycle orchestration with hot-reload ([config](config/), [viper](viper/)), version-aware serialization ([retro](retro/))

### ⚡ Concurrency & Control
- **Atomic Types**: Type-safe atomic primitives with default values and type casting ([atomic](atomic/))
- **Semaphores**: Weighted semaphores, WaitGroups with progress tracking ([semaphore](semaphore/))
- **Lifecycle**: Start/Stop patterns, ticker management, context control ([runner](runner/))
- **Context**: Gin context wrappers, enhanced context utilities ([context](context/))

### 🧪 Development Tools
- **CLI Framework**: Cobra extensions with enhanced features ([cobra](cobra/))
- **Shell**: Interactive shell with command management, TTY handling ([shell](shell/))
- **Console**: Terminal formatting, colored output, progress indicators ([console](console/))
- **Static File Server**: Secure static file serving with WAF/IDS/EDR integration, rate limiting, path security ([static](static/))

---

## Architecture

### Package Organization

The library is organized into 37 top-level packages, each focused on a specific domain:

```
golib/
├── Cloud & Infrastructure
│   ├── archive/              Archive & compression (TAR, ZIP, GZIP, BZIP2, LZ4, XZ)
│   ├── artifact/             Artifact management (GitHub, GitLab, JFrog, S3)
│   └── aws/                  AWS SDK integration (S3, IAM, MinIO)
│
├── Web & Networking
│   ├── httpcli/              Advanced HTTP client with DNS mapping
│   ├── httpserver/           HTTP/HTTPS server pool with TLS and multi-handler
│   ├── router/               Gin-based router with auth and middleware
│   └── socket/               TCP/UDP/Unix socket servers and clients
│
├── Data Management
│   ├── cache/                In-memory cache with TTL
│   ├── database/             GORM integration and key-value stores
│   ├── file/                 File operations with progress tracking
│   └── ioutils/              I/O utilities (aggregation, multiplexing, delimiters)
│
├── Security & Communication
│   ├── certificates/         TLS certificate management
│   ├── ftpclient/            FTP client
│   ├── ldap/                 LDAP authentication
│   ├── mail/                 Complete SMTP with TLS/STARTTLS, templating, queuing
│   ├── nats/                 NATS messaging
│   ├── oauth/                OAuth clients
│   └── password/             Password generation
│
├── Monitoring & Logging
│   ├── logger/               Structured logging with multiple outputs
│   ├── monitor/              System monitoring and health checks
│   ├── prometheus/           Prometheus metrics integration
│   └── status/               Component status management
│
├── Utilities & Helpers
│   ├── config/               Component lifecycle orchestration with hot-reload
│   ├── duration/             Extended duration with days (1d=24h) and large ranges
│   ├── encoding/             Encryption, hashing, remote random reader (HSM)
│   ├── errors/               Advanced error management
│   ├── network/              Network protocol utilities
│   ├── retro/                Version-aware struct serialization
│   ├── size/                 Byte size arithmetic
│   ├── version/              Semantic versioning
│   └── viper/                Viper configuration wrapper
│
├── Concurrency & Control
│   ├── atomic/               Type-safe atomic primitives with defaults
│   ├── context/              Enhanced context utilities
│   ├── runner/               Lifecycle management (Start/Stop, Ticker)
│   └── semaphore/            Weighted semaphores with WaitGroups and progress
│
└── Development Tools
    ├── cobra/                Cobra CLI extensions
    ├── console/              Terminal formatting
    ├── pprof/                Profiling utilities
    ├── request/              HTTP request helpers
    ├── shell/                Interactive shell
    └── static/               Security-focused static file server with caching
```

**Package Count**: 37 top-level, 165 total (including subpackages)

### Dependency Model

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  (Your application imports specific golib packages)         │
└─────────────────────────────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   High-Level Packages                       │
│  httpserver, router, config, monitor, mail, archive         │
└─────────────────────────────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Mid-Level Packages                        │
│  logger, errors, runner, status, database, httpcli          │
└─────────────────────────────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Low-Level Packages                        │
│  atomic, context, semaphore, ioutils, encoding              │
└─────────────────────────────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Foundation Packages                       │
│  version, size, duration, network, password                 │
└─────────────────────────────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Go Standard Library                       │
└─────────────────────────────────────────────────────────────┘
```

**Design Principles**:
- **Layered Architecture**: Clear separation between application, business, and infrastructure layers
- **Minimal Coupling**: Packages depend on abstractions, not concrete implementations
- **Interface-Based**: Most packages define interfaces for extensibility
- **Standard Library First**: Leverage Go standard library where possible

---

## Installation

### Requirements

- **Go**: 1.24 or higher (required for `os.OpenRoot` API)
- **Platform**: Linux, macOS, or Windows
- **CGO**: Optional (required for race detector during testing)

### Quick Setup

```bash
# Install entire library
go get github.com/nabbar/golib

# Install specific packages
go get github.com/nabbar/golib/logger
go get github.com/nabbar/golib/archive
go get github.com/nabbar/golib/httpserver
go get github.com/nabbar/golib/errors
go get github.com/nabbar/golib/ioutils
```

### Import in Your Project

```go
package main

import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/errors"
    "github.com/nabbar/golib/httpserver"
)

func main() {
    // Use packages as needed
}
```

### Updating

```bash
# Update all packages
go get -u github.com/nabbar/golib/...

# Update specific package
go get -u github.com/nabbar/golib/logger
```

---

## Package Catalog

Detailed list of all packages with coverage statistics and links to documentation.

**Total**: 62 packages with individual README.md documentation across 165 Go packages.

**Note**: Many parent packages have multiple documented subpackages:
- **encoding**: 6 documented packages (base + aes, hexa, mux, randRead, sha256) - [See encoding/README.md](encoding/)
- **ioutils**: 11 documented packages (base + aggregator, bufferReadCloser, delim, fileDescriptor, ioprogress, iowrapper, mapCloser, maxstdio, multi, nopwritecloser) - [See ioutils/README.md](ioutils/)
- **mail**: 5 documented packages (base + queuer, render, sender, smtp) - [See mail/README.md](mail/)
- **monitor**: 3 documented packages (base + info, pool) - [See monitor/README.md](monitor/)
- **prometheus**: 2 documented packages (base + webmetrics) - [See prometheus/README.md](prometheus/)
- **socket**: 3 documented packages (base + client, server) - [See socket/README.md](socket/)

The table below lists all 165 Go packages with their test coverage and links to their 62 individual README.md documentation files.

### Cloud & Infrastructure

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **archive** | 8.6% | 89 | Streaming archive and compression | [README](archive/) |
| **artifact** | 23.4% | 19 | Artifact management (GitHub/GitLab/JFrog/S3) | [README](artifact/) |
| **artifact/client** | 98.6% | 21 | Generic artifact client interface | - |
| **aws** | 5.4% | 220 | AWS SDK integration (S3, IAM, MinIO) | [README](aws/) |
| **static** | 82.6% | 229 | Security-focused static file server with embed.FS, rate limiting, WAF integration | [README](static/) |

### Web & Networking

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **httpcli** | 71.4% | 34 | Advanced HTTP client with retries | [README](httpcli/) |
| **httpcli/dns-mapper** | 71.7% | 26 | DNS mapping for HTTP clients | - |
| **httpserver** | 52.5% | 84 | HTTP/HTTPS server pool management with TLS, monitoring and multi-handler support | [README](httpserver/) |
| **httpserver/pool** | 63.1% | 78 | Multi-server orchestration and pooling | - |
| **network** | 75.4% | 162 | Network utilities and protocol handling | [README](network/) |
| **network/protocol** | 98.7% | 298 | Network protocol constants and helpers | - |
| **router** | 91.0% | 61 | Gin-based router with auth | [README](router/) |
| **router/auth** | 96.3% | 12 | Authentication middleware | - |
| **router/authheader** | 100% | 11 | Authorization header parsing | - |
| **router/header** | 83.3% | 29 | HTTP header utilities | - |
| **socket** | No tests | - | Socket communication base | [README](socket/) |
| **socket/client** | No tests | - | Socket client base | [README](socket/client/) |
| **socket/client/tcp** | 74.0% | 119 | TCP client implementation | - |
| **socket/client/udp** | 72.8% | 68 | UDP client implementation | - |
| **socket/client/unix** | 76.3% | 67 | Unix domain socket client | - |
| **socket/client/unixgram** | 76.8% | 65 | Unix datagram socket client | - |
| **socket/server** | No tests | - | Socket server base | [README](socket/server/) |
| **socket/server/tcp** | 84.6% | 117 | TCP server implementation | - |
| **socket/server/udp** | 72.0% | 18 | UDP server implementation | - |
| **socket/server/unix** | 73.5% | 23 | Unix domain socket server | - |
| **socket/server/unixgram** | 70.8% | 20 | Unix datagram socket server | - |

### Data Management

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **cache** | 96.7% | 43 | In-memory cache with TTL | [README](cache/) |
| **cache/item** | 96.7% | 21 | Cache item implementation | - |
| **database/gorm** | 19.6% | 41 | GORM integration | [README](database/) |
| **database/kvdriver** | 38.4% | 20 | Key-value store driver | - |
| **database/kvitem** | 76.4% | 33 | Key-value item | - |
| **database/kvmap** | 66.7% | 30 | Key-value map | - |
| **database/kvtable** | 65.9% | 24 | Key-value table | - |
| **file/bandwidth** | 77.8% | 25 | Bandwidth-limited I/O | [README](file/) |
| **file/perm** | 88.9% | 141 | File permission utilities | - |
| **file/progress** | 71.1% | 90 | File operation progress tracking | - |
| **ioutils** | 87.7% | 31 | I/O utilities base | [README](ioutils/) |
| **ioutils/aggregator** | 86.0% | 115 | Write operation aggregator | [README](ioutils/aggregator/) |
| **ioutils/bufferReadCloser** | 100% | 57 | Buffered reader with closer | [README](ioutils/bufferReadCloser/) |
| **ioutils/delim** | 100% | 198 | Delimiter-based stream processing | [README](ioutils/delim/) |
| **ioutils/fileDescriptor** | 85.7% | 23 | File descriptor management | [README](ioutils/fileDescriptor/) |
| **ioutils/ioprogress** | 84.7% | 42 | I/O progress tracking | [README](ioutils/ioprogress/) |
| **ioutils/iowrapper** | 100% | 114 | Generic I/O wrappers | [README](ioutils/iowrapper/) |
| **ioutils/mapCloser** | 77.5% | 29 | Multiple closer management | [README](ioutils/mapCloser/) |
| **ioutils/maxstdio** | No tests | - | Stdio limit management | [README](ioutils/maxstdio/) |
| **ioutils/multi** | 81.7% | 113 | Write multiplexing | [README](ioutils/multi/) |
| **ioutils/nopwritecloser** | 100% | 54 | No-op writer closer | [README](ioutils/nopwritecloser/) |

### Security & Communication

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **certificates** | 44.6% | 2 | TLS certificate management | [README](certificates/) |
| **certificates/auth** | 73.0% | 13 | Certificate authentication | - |
| **certificates/ca** | 64.1% | 17 | Certificate authority operations | - |
| **certificates/certs** | 48.4% | 2 | Certificate utilities | - |
| **certificates/cipher** | 51.2% | 6 | Cipher suite management | - |
| **certificates/curves** | 51.6% | 4 | Elliptic curve management | - |
| **certificates/tlsversion** | 54.5% | 4 | TLS version utilities | - |
| **ftpclient** | 6.2% | 22 | FTP client implementation | [README](ftpclient/) |
| **ldap** | No tests | - | LDAP authentication | [README](ldap/) |
| **mail** | No tests | - | Complete email solution with SMTP (TLS/STARTTLS), templating, composition and queuing | [README](mail/) |
| **mail/queuer** | 90.8% | 102 | Email queuing with pooling | [README](mail/queuer/) |
| **mail/render** | 89.6% | 123 | Email template rendering | [README](mail/render/) |
| **mail/sender** | 81.4% | 252 | Email sending utilities | [README](mail/sender/) |
| **mail/smtp** | 90.1% | 104 | SMTP client implementation | [README](mail/smtp/) |
| **mail/smtp/config** | 92.7% | 222 | SMTP configuration | - |
| **mail/smtp/tlsmode** | 98.8% | 165 | SMTP TLS mode handling | - |
| **nats** | No tests | - | NATS messaging client | - |
| **oauth** | No tests | - | OAuth client integration | - |
| **password** | 84.6% | 6 | Password generation | [README](password/) |

### Monitoring & Logging

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **logger** | 68.0% | 81 | Structured logging framework | [README](logger/) |
| **logger/config** | 85.3% | 127 | Logger configuration | - |
| **logger/entry** | 85.1% | 119 | Log entry management | - |
| **logger/fields** | 49.2% | 49 | Log field injection | - |
| **logger/gorm** | 100% | 34 | GORM logger integration | - |
| **logger/hashicorp** | 96.6% | 89 | Hashicorp logger adapter | - |
| **logger/hookfile** | 19.6% | 22 | File output hook | - |
| **logger/hookstderr** | 100% | 30 | Stderr output hook | - |
| **logger/hookstdout** | 100% | 30 | Stdout output hook | - |
| **logger/hooksyslog** | 53.5% | 20 | Syslog output hook | - |
| **logger/hookwriter** | 90.2% | 31 | Generic writer hook | - |
| **logger/level** | 65.9% | 42 | Log level management | - |
| **monitor** | 81.9% | 103 | System monitoring | [README](monitor/) |
| **monitor/info** | 100% | 95 | System information collection | [README](monitor/info/) |
| **monitor/pool** | 76.2% | 153 | Monitor pooling | [README](monitor/pool/) |
| **monitor/status** | 98.4% | 181 | Status reporting | - |
| **prometheus** | 60.0% | 137 | Prometheus metrics | [README](prometheus/) |
| **prometheus/bloom** | 94.7% | 45 | Bloom filter metrics | - |
| **prometheus/metrics** | 95.5% | 179 | Custom metrics | - |
| **prometheus/pool** | 72.5% | 74 | Prometheus pooling | - |
| **prometheus/types** | 100% | 36 | Prometheus type definitions | - |
| **prometheus/webmetrics** | No tests | - | Web metrics endpoint | [README](prometheus/webmetrics/) |
| **status** | 85.9% | 120 | Status management | [README](status/) |
| **status/control** | 95.0% | 102 | Status control | - |
| **status/listmandatory** | 86.0% | 29 | Mandatory status list | - |
| **status/mandatory** | 76.1% | 55 | Mandatory status checks | - |

### Utilities & Helpers

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **config** | 61.9% | 93 | Component lifecycle orchestration with dependency resolution and hot-reload | [README](config/) |
| **config/components/aws** | 40.7% | 183 | AWS component config | - |
| **config/components/database** | 39.0% | 196 | Database component config | - |
| **config/components/head** | 84.6% | 185 | Header component config | - |
| **config/components/http** | 71.7% | 257 | HTTP component config | - |
| **config/components/httpcli** | 86.0% | 137 | HTTP client component config | - |
| **config/components/ldap** | 70.5% | 90 | LDAP component config | - |
| **config/components/log** | 88.0% | 191 | Logger component config | - |
| **config/components/mail** | 64.8% | 68 | Mail component config | - |
| **config/components/request** | 43.3% | 35 | Request component config | - |
| **config/components/smtp** | 51.5% | 73 | SMTP component config | - |
| **config/components/tls** | 82.1% | 171 | TLS component config | - |
| **duration** | 91.5% | 179 | Extended duration with days support (1d=24h) and arithmetic operations | [README](duration/) |
| **duration/big** | 91.0% | 250 | Large durations beyond time.Duration range (seconds precision, ~292B years) | - |
| **encoding** | No tests | - | Unified encoding interface with remote random reader for HSM operations | [README](encoding/) |
| **encoding/aes** | 91.5% | 126 | AES encryption | [README](encoding/aes/) |
| **encoding/hexa** | 89.7% | 97 | Hexadecimal encoding | [README](encoding/hexa/) |
| **encoding/mux** | 81.7% | 59 | Encoding multiplexer | [README](encoding/mux/) |
| **encoding/randRead** | 81.4% | 32 | Buffered random reader from remote sources (APIs, HSM) for cryptographic operations | [README](encoding/randRead/) |
| **encoding/sha256** | 84.8% | 61 | SHA256 hashing | [README](encoding/sha256/) |
| **errors** | 87.6% | 222 | Advanced error management | [README](errors/) |
| **errors/pool** | 100% | 83 | Error pool management | - |
| **retro** | 84.2% | 156 | Version-aware struct serialization for configuration evolution and backward compatibility | [README](retro/) |
| **size** | 95.4% | 352 | Byte size arithmetic | [README](size/) |
| **version** | 93.8% | 173 | Semantic versioning | [README](version/) |
| **viper** | 73.3% | 104 | Viper configuration wrapper | [README](viper/) |

### Concurrency & Control

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **atomic** | 91.8% | 49 | Type-safe atomic primitives with default values and type casting helpers | [README](atomic/) |
| **context** | 87.5% | 80 | Context utilities | [README](context/) |
| **context/gin** | 76.9% | 77 | Gin context wrapper | - |
| **runner** | No tests | - | Lifecycle management base | [README](runner/) |
| **runner/startStop** | 88.8% | 42 | Start/Stop pattern | - |
| **runner/ticker** | 90.2% | 88 | Ticker management | - |
| **semaphore** | 97.4% | 33 | Weighted semaphores, WaitGroups and progress tracking with MPB integration | [README](semaphore/) |
| **semaphore/bar** | 96.6% | 68 | Semaphore with progress bar | - |
| **semaphore/sem** | 100% | 66 | Semaphore implementation | - |

### Development Tools

| Package | Coverage | Specs | Description | Documentation |
|---------|----------|-------|-------------|---------------|
| **cobra** | 76.7% | 156 | Cobra CLI extensions | [README](cobra/) |
| **console** | 60.0% | 182 | Terminal formatting | [README](console/) |
| **pprof** | No tests | - | Profiling utilities | - |
| **request** | No tests | - | HTTP request helpers | - |
| **shell** | 48.4% | 120 | Interactive shell | [README](shell/) |
| **shell/command** | 81.8% | 93 | Shell command management | - |
| **shell/tty** | 44.7% | 126 | TTY handling | - |

---

## Quick Start

### Logging

```go
package main

import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/level"
)

func main() {
    // Create logger with stdout output
    log := logger.New(nil)
    log.SetLevel(level.InfoLevel)
    
    // Basic logging
    log.Info("Application started")
    log.Error("An error occurred")
    
    // Structured logging with fields
    log.WithField("user", "john").Info("User logged in")
    log.WithFields(map[string]interface{}{
        "method": "GET",
        "path":   "/api/users",
        "status": 200,
    }).Info("Request processed")
}
```

### HTTP Server

```go
package main

import (
    "context"
    "github.com/nabbar/golib/httpserver"
    "github.com/nabbar/golib/logger"
)

func main() {
    log := logger.New(nil)
    
    // Create HTTP server
    srv := httpserver.New(context.Background(), log)
    srv.SetListenAddress(":8080")
    
    // Add routes
    srv.RegisterHandler("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    // Start server
    if err := srv.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer srv.Stop()
    
    // Wait for interrupt
    select {}
}
```

### Archive & Compression

```go
package main

import (
    "github.com/nabbar/golib/archive"
    "os"
)

func main() {
    // Create tar.gz archive
    input, _ := os.Open("source-dir")
    output, _ := os.Create("archive.tar.gz")
    defer output.Close()
    
    // Stream compression
    if err := archive.Compress(input, output, "tar", "gzip"); err != nil {
        panic(err)
    }
    
    // Extract archive
    archiveFile, _ := os.Open("archive.tar.gz")
    if err := archive.Extract(archiveFile, "dest-dir"); err != nil {
        panic(err)
    }
}
```

### Error Handling

```go
package main

import (
    "github.com/nabbar/golib/errors"
)

func main() {
    // Create error with code
    err := errors.New("database connection failed").
        SetCode(500).
        AddParent(errors.New("network timeout"))
    
    // Check error
    if err != nil {
        // Get error code
        code := errors.GetCode(err)
        
        // Get full error chain
        chain := errors.GetParent(err)
        
        // Log with stack trace
        println(err.Error())
    }
}
```

### Configuration

```go
package main

import (
    "github.com/nabbar/golib/viper"
    "github.com/nabbar/golib/config"
)

func main() {
    // Load configuration from file
    cfg := viper.New()
    cfg.SetConfigFile("config.yaml")
    cfg.ReadInConfig()
    
    // Access configuration
    port := cfg.GetInt("server.port")
    host := cfg.GetString("server.host")
    
    println("Starting server on", host, ":", port)
}
```

---

## Performance

### Test Suite Metrics

Based on complete coverage analysis with `coverage-report.sh`:

```
Total Packages:       165
Packages with Tests:  127 (77.0%)
Packages without Tests: 38 (23.0%)

Test Specifications:  10,964
Test Assertions:      21,470
Benchmarks:           92
Pending Tests:        18

Execution Time:       Variable (cached results shown)
Race Conditions:      0 (verified with -race flag)
```

### Coverage Analysis

**Distribution by Coverage Level:**

| Coverage Range | Count | Percentage | Package Examples |
|----------------|-------|------------|------------------|
| **100%** | 14 | 11.0% | errors/pool, logger/gorm, router/authheader, semaphore/sem |
| **90-99%** | 24 | 18.9% | atomic (91.8%), version (93.8%), size (95.4%) |
| **80-89%** | 29 | 22.8% | ioutils (87.7%), mail/queuer (90.8%), static (82.6%), context (87.5%) |
| **70-79%** | 19 | 15.0% | cobra (76.7%), viper (73.3%), file/bandwidth (77.8%) |
| **60-69%** | 10 | 7.9% | config (61.9%), logger (68.0%), database/kvmap (66.7%) |
| **<60%** | 31 | 24.4% | archive (8.6%), aws (5.4%), httpserver (52.5%) |

**Average Coverage**: 73.9% (weighted across all 127 tested packages)

### Performance Highlights

**High-Performance Packages (>90% coverage):**

- **atomic**: 91.8% coverage, 49 specs - Generic atomic operations
- **mail/smtp/config**: 92.7% coverage, 222 specs - SMTP configuration
- **version**: 93.8% coverage, 173 specs - Semantic versioning
- **size**: 95.4% coverage, 352 specs - Byte size arithmetic
- **prometheus/metrics**: 95.5% coverage, 179 specs - Custom metrics

**Well-Tested Core Packages:**

- **ioutils**: 87.7% average across 10 subpackages, 772 specs
- **mail**: 89.0% average across 6 subpackages, 970 specs
- **logger**: 74.7% base + 85%+ in config/entry, 573 specs
- **errors**: 87.6% base + 100% pool, 305 specs
- **monitor**: 88.5% base + 100% info, 572 specs

**Packages Needing Improvement (<40% coverage):**

- archive (8.6%), aws (5.4%), artifact implementations
- database/kvdriver (38.4%), config/components (39-43%)
- ftpclient (6.2%), shell/tty (44.7%)

See [TESTING.md](TESTING.md) for detailed test suite documentation.

---

## Use Cases

### 1. Cloud-Native Microservices

**Scenario**: Build a microservice handling S3 uploads, monitoring, and structured logging.

```go
import (
    "github.com/nabbar/golib/aws"
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/prometheus"
)

// Microservice with full observability
type Service struct {
    storage    aws.Client
    log        logger.Logger
    monitor    monitor.Monitor
    metrics    prometheus.Registry
}
```

**Packages Used**: aws, logger, monitor, prometheus, httpserver

### 2. API Gateway with Advanced Routing

**Scenario**: HTTP API gateway with authentication, rate limiting, and request/response logging.

```go
import (
    "github.com/nabbar/golib/router"
    "github.com/nabbar/golib/httpcli"
    "github.com/nabbar/golib/logger"
)

// Gateway with middleware stack
r := router.New()
r.Use(router.Auth())
r.Use(router.RateLimiter())
r.Use(router.Logging(log))
```

**Packages Used**: router, httpcli, logger, context, errors

### 3. Data Pipeline with Streaming

**Scenario**: ETL pipeline processing large archives with progress tracking and error handling.

```go
import (
    "github.com/nabbar/golib/archive"
    "github.com/nabbar/golib/ioutils"
    "github.com/nabbar/golib/file/progress"
)

// Stream processing with progress
progress := ioprogress.NewReader(reader, func(n int64) {
    fmt.Printf("Processed: %d bytes\n", n)
})

archive.Extract(progress, destDir)
```

**Packages Used**: archive, ioutils, file, errors, logger

### 4. Distributed Task Queue

**Scenario**: Background job processing with concurrency control and monitoring.

```go
import (
    "github.com/nabbar/golib/semaphore"
    "github.com/nabbar/golib/runner"
    "github.com/nabbar/golib/status"
)

// Concurrent task executor
sem := semaphore.New(maxWorkers)
for _, task := range tasks {
    sem.Acquire()
    go func(t Task) {
        defer sem.Release()
        processTask(t)
    }(task)
}
```

**Packages Used**: semaphore, runner, status, monitor, errors

### 5. Configuration Management System

**Scenario**: Multi-environment configuration with validation and hot-reload.

```go
import (
    "github.com/nabbar/golib/config"
    "github.com/nabbar/golib/viper"
)

// Environment-aware config
cfg := config.New()
cfg.Load("config.yaml")
cfg.Watch() // Hot reload on file change
```

**Packages Used**: config, viper, logger, errors

---

## Best Practices

### ✅ DO

**Use Structured Logging:**
```go
// ✅ GOOD: Structured fields for filtering/analysis
log.WithFields(map[string]interface{}{
    "user_id": userID,
    "action":  "login",
    "ip":      remoteAddr,
}).Info("User authentication")

// ❌ BAD: Unstructured string concatenation
log.Info("User " + userID + " logged in from " + remoteAddr)
```

**Handle Errors Properly:**
```go
// ✅ GOOD: Error wrapping with context
if err != nil {
    return errors.New("database query failed").
        AddParent(err).
        SetCode(500)
}

// ❌ BAD: Ignoring errors
result, _ := database.Query()
```

**Leverage Context for Cancellation:**
```go
// ✅ GOOD: Context-aware operations
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

if err := service.Process(ctx, data); err != nil {
    // Handle timeout or cancellation
}

// ❌ BAD: No cancellation support
service.Process(data)
```

**Use Interfaces for Testing:**
```go
// ✅ GOOD: Interface-based dependencies
type UserService struct {
    storage Storage // interface
    logger  logger.Logger // interface
}

// ❌ BAD: Concrete dependencies
type UserService struct {
    storage *PostgresDB // concrete type
}
```

**Stream Large Files:**
```go
// ✅ GOOD: Streaming with constant memory
reader := ioutils.NewReader(file)
writer := ioutils.NewWriter(dest)
io.Copy(writer, reader)

// ❌ BAD: Loading entire file into memory
data, _ := ioutil.ReadAll(file)
ioutil.WriteFile(dest, data, 0644)
```

### ❌ DON'T

**Don't Ignore Resource Cleanup:**
```go
// ❌ BAD: No cleanup
file, _ := os.Open("data.txt")
data, _ := ioutil.ReadAll(file)

// ✅ GOOD: Proper cleanup
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()
```

**Don't Block Main Goroutine:**
```go
// ❌ BAD: Blocking indefinitely
server.Start()

// ✅ GOOD: Non-blocking with graceful shutdown
go server.Start()
<-ctx.Done()
server.Stop()
```

**Don't Ignore Race Conditions:**
```go
// ❌ BAD: Shared state without synchronization
var counter int
for i := 0; i < 10; i++ {
    go func() { counter++ }()
}

// ✅ GOOD: Use atomic or mutex
var counter atomic.Int64
for i := 0; i < 10; i++ {
    go func() { counter.Add(1) }()
}
```

**Don't Hardcode Configuration:**
```go
// ❌ BAD: Hardcoded values
db := sql.Open("postgres", "postgres://localhost:5432/db")

// ✅ GOOD: External configuration
cfg := config.Load()
db := sql.Open("postgres", cfg.GetString("database.url"))
```

**Don't Skip Error Checking:**
```go
// ❌ BAD: Ignoring potential errors
json.Unmarshal(data, &result)

// ✅ GOOD: Always check errors
if err := json.Unmarshal(data, &result); err != nil {
    return errors.New("JSON parsing failed").AddParent(err)
}
```

---

## Testing

Comprehensive test suite with 10,964 specifications across 127 packages.

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Build Configuration

### Build Tags

The library supports platform-specific builds using Go build tags:

```bash
# Build for specific architecture
go build -tags=386 ./...      # 32-bit
go build -tags=amd64 ./...    # 64-bit

# Platform-specific features
go build -tags=linux ./...
go build -tags=darwin ./...
go build -tags=windows ./...
```

### Build Scripts

Two build helper scripts are provided:

- **`build.386`**: Build for 32-bit architecture
- **`build.amd64`**: Build for 64-bit architecture

```bash
# Use build scripts
./build.amd64
./build.386
```

### Compile-Time Requirements

- **Go Version**: 1.24+ (required for `os.OpenRoot` API introduced in Go 1.24)
- **CGO**: Optional (needed for race detector: `CGO_ENABLED=1`)
- **Platform**: Linux, macOS, Windows (some packages have platform-specific implementations)

### Optimization Flags

Recommended build flags for production:

```bash
# Optimized production build
go build -ldflags="-s -w" -trimpath ./...

# With race detector (development only, significant performance impact)
CGO_ENABLED=1 go build -race ./...

# Static binary (Linux)
CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" ./...
```

### Cross-Compilation

```bash
# Linux from macOS
GOOS=linux GOARCH=amd64 go build ./...

# Windows from Linux
GOOS=windows GOARCH=amd64 go build ./...

# macOS from Linux
GOOS=darwin GOARCH=amd64 go build ./...
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: ≥80%)
   - Pass all tests including race detector
   - Use `gofmt` and `golint`

2. **AI Usage Policy**
   - ❌ **Do NOT use AI** for implementing package functionality or core logic
   - ✅ **AI may assist** with:
     - Writing and improving tests
     - Documentation and comments
     - Debugging and troubleshooting
   - All AI-assisted contributions must be reviewed and validated by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega for test framework
   - Ensure zero race conditions
   - Maintain coverage above 80%

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md if adding packages
   - Update TESTING.md if changing test structure

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description of changes

---

## Future Enhancements

Potential improvements and new features for consideration:

### Test Coverage Improvements

**Priority Packages** (current coverage <40%):

1. **archive** (8.6% → target 80%+)
   - Add tests for TAR/ZIP operations
   - Stream processing edge cases
   - Format detection scenarios

2. **aws** (5.4% → target 80%+)
   - S3 operations testing
   - IAM integration tests
   - MinIO compatibility tests

3. **artifact** subpackages (6-23% → target 80%+)
   - GitHub/GitLab/JFrog client tests
   - Authentication scenarios
   - Error handling coverage

4. **config/components** (39-43% → target 80%+)
   - Component configuration validation
   - Integration tests
   - Edge case handling

5. **database/kvdriver** (38.4% → target 80%+)
   - Key-value store operations
   - Transaction handling
   - Concurrent access tests

### New Features

**High Priority:**

- **Distributed Tracing**: OpenTelemetry integration for distributed systems
- **Service Mesh**: Istio/Linkerd compatibility helpers
- **gRPC Support**: gRPC server and client implementations
- **Event Streaming**: Kafka, RabbitMQ integration packages
- **Circuit Breaker**: Resilience patterns for failure handling

**Medium Priority:**

- **GraphQL**: GraphQL server and client utilities
- **WebSocket**: Enhanced WebSocket support
- **Rate Limiting**: Advanced rate limiting algorithms (token bucket, sliding window)
- **Caching**: Redis integration, distributed caching
- **Secrets Management**: Vault, AWS Secrets Manager integration

**Low Priority:**

- **Machine Learning**: Model serving utilities
- **Blockchain**: Ethereum, Hyperledger integration
- **IoT**: MQTT protocol support
- **Mobile**: Mobile-specific optimizations

### Performance Optimizations

- **Zero-Copy I/O**: Extend zero-copy operations to more packages
- **SIMD**: Leverage SIMD instructions for data processing
- **Memory Pooling**: Reduce allocations in hot paths
- **Async I/O**: io_uring support for Linux
- **Compilation**: Pre-compiled binaries for common platforms

### Documentation Enhancements

- **Interactive Examples**: Runnable examples in documentation
- **Architecture Diagrams**: Detailed architecture visualizations
- **Migration Guides**: Upgrade guides between major versions
- **Best Practices**: Extended best practices documentation
- **Video Tutorials**: Video content for common use cases

### Tooling Improvements

- **CLI Tool**: golib command-line tool for scaffolding
- **Code Generator**: Generate boilerplate for common patterns
- **Linter**: Custom linter for golib best practices
- **Performance Profiler**: Built-in profiling utilities
- **Dependency Analyzer**: Analyze package dependencies

### Community & Ecosystem

- **Plugin System**: Extensible plugin architecture
- **Marketplace**: Community-contributed packages
- **Templates**: Project templates for common scenarios
- **Integration Tests**: Comprehensive integration test suite
- **Benchmarking Suite**: Cross-package performance benchmarks

**Note**: These enhancements are suggestions based on current gaps and industry trends. Actual implementation priorities depend on community feedback and real-world usage patterns.

---

## Resources

### Internal Documentation
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib) - Complete API documentation
- [TESTING.md](TESTING.md) - Test suite documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [SECURITY.md](SECURITY.md) - Security policy

### Package Documentation
Individual package READMEs available in each subdirectory.

### External References
- [Go Documentation](https://go.dev/doc/) - Official Go documentation
- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
- [Go Blog](https://go.dev/blog/) - Official Go blog

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

Copyright (c) 2019-2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Repository**: [github.com/nabbar/golib](https://github.com/nabbar/golib)  
**Version**: See [releases](https://github.com/nabbar/golib/releases)
