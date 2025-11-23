# Static File Server

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)

High-performance, security-focused static file server for Go with enterprise-grade features including WAF/IDS/EDR integration, rate limiting, path traversal protection, and advanced HTTP caching.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The **static** package provides a production-ready static file server built on top of Go's `embed.FS` with comprehensive security features designed for modern web applications. It seamlessly integrates with the Gin web framework and provides enterprise-grade security monitoring through WAF/IDS/EDR webhook integration.

### Design Philosophy

1. **Security First** - Multiple layers of protection against common web attacks
2. **Zero Mutex** - Lock-free concurrency using atomic operations for maximum performance
3. **Observable** - Built-in security event streaming to SIEM systems
4. **Production Ready** - Battle-tested with 82.6% test coverage and zero race conditions
5. **Developer Friendly** - Simple API with sensible defaults

---

## Key Features

### Core Features

- **Embedded Filesystem** - Serve files from Go's `embed.FS`
- **Gin Integration** - Seamless integration with Gin web framework
- **Multiple Base Paths** - Support for multiple embedded directories
- **Index Files** - Automatic index file resolution for directories
- **Download Mode** - Force download with Content-Disposition header
- **URL Redirects** - HTTP 301 permanent redirects
- **Custom Handlers** - Override default behavior for specific routes

### Security Features

#### 1. Path Security
- **Path Traversal Protection** - Prevents `../` attacks
- **Null Byte Injection Prevention** - Blocks null byte attacks
- **Dot File Blocking** - Protects `.env`, `.git`, `.htaccess`
- **Max Path Depth** - Limits directory traversal depth
- **Pattern Blocking** - Blocks configurable path patterns

#### 2. Rate Limiting
- **IP-based Limiting** - Tracks unique files per IP
- **Sliding Window** - Accurate rate calculation
- **Whitelist Support** - Bypass for trusted IPs
- **Automatic Cleanup** - Prevents memory leaks
- **Standard Headers** - `X-RateLimit-*`, `Retry-After`

#### 3. HTTP Security Headers
- **ETag Support** - Efficient cache validation
- **Cache-Control** - Fine-grained cache control
- **Content-Type Validation** - MIME type filtering
- **Custom MIME Types** - Override default detection
- **Expires Headers** - HTTP/1.0 compatibility

#### 4. Suspicious Access Detection
- **Pattern Recognition** - Detects common attack patterns
- **Backup File Scanning** - Identifies `.bak`, `.old` attempts
- **Config File Scanning** - Detects `config.php` attempts
- **Admin Panel Scanning** - Identifies `/admin` probes
- **Path Manipulation** - Detects double slashes, backslashes

#### 5. WAF/IDS/EDR Integration
- **Webhook Support** - Real-time event streaming
- **CEF Format** - Common Event Format for SIEM systems
- **Batch Processing** - Efficient bulk event sending
- **Go Callbacks** - Programmatic event handling
- **Async Processing** - Non-blocking event delivery
- **Severity Filtering** - Configurable event levels

---

## Installation

```bash
go get github.com/nabbar/golib/static
```

---

## Architecture

### Request Flow

```
HTTP Request
     │
     ├──> [Rate Limiter]
     │         │
     │         ├──> Limit Exceeded? ──> 429 Too Many Requests
     │         └──> OK
     │
     ├──> [Path Security Validator]
     │         │
     │         ├──> Path Traversal? ──> 403 Forbidden + Event
     │         ├──> Dot File Access? ──> 403 Forbidden + Event
     │         ├──> Blocked Pattern? ──> 403 Forbidden + Event
     │         └──> OK
     │
     ├──> [Redirect Handler]
     │         │
     │         └──> Redirect? ──> 301 Permanent Redirect
     │
     ├──> [Custom Handler]
     │         │
     │         └──> Custom? ──> Execute Custom Handler
     │
     ├──> [Index File Resolution]
     │         │
     │         └──> Directory? ──> Serve Index File
     │
     ├──> [File Lookup]
     │         │
     │         ├──> Not Found? ──> 404 Not Found
     │         └──> Found
     │
     ├──> [MIME Type Validation]
     │         │
     │         ├──> Denied Type? ──> 403 Forbidden + Event
     │         └──> OK
     │
     ├──> [ETag Validation]
     │         │
     │         └──> Match? ──> 304 Not Modified
     │
     ├──> [Suspicious Access Detection]
     │         │
     │         └──> Suspicious? ──> Log + Notify
     │
     └──> [File Delivery]
           │
           └──> 200 OK + File Content + Cache Headers
```

### Security Event Processing

```
Security Event
     │
     ├──> [Severity Filter]
     │         │
     │         └──> Below Min? ──> Drop
     │
     ├──> [Go Callbacks]
     │         │
     │         └──> async goroutine
     │
     └──> [Webhook Integration]
           │
           ├──> Batch Enabled?
           │         │
           │         ├──> Add to Batch
           │         │     │
           │         │     ├──> Batch Full? ──> Send Immediately
           │         │     └──> Start Timer ──> Send on Timeout
           │         │
           │         └──> Real-time
           │               │
           │               ├──> JSON Format ──> POST to Webhook
           │               └──> CEF Format ──> POST to SIEM
```

### Thread Safety Model

```
┌─────────────────────────────────────────────────────┐
│             Static Handler (Lock-Free)              │
├─────────────────────────────────────────────────────┤
│                                                      │
│  Atomic Operations (libatm.Value, atomic.*)         │
│  ├─ Configuration (RateLimit, Security, Headers)    │
│  ├─ IP Tracking (libatm.MapTyped)                   │
│  ├─ Event Batching (atomic counters + map)          │
│  └─ Router State                                    │
│                                                      │
│  Embedded Filesystem (read-only, inherently safe)   │
│                                                      │
│  Context-based Configuration (libctx.Config)        │
│  ├─ Index files                                     │
│  ├─ Downloads                                       │
│  ├─ Redirects                                       │
│  └─ Custom handlers                                 │
│                                                      │
└─────────────────────────────────────────────────────┘

  No mutexes required!
  Concurrent reads and writes are safe by design.
```

---

## Quick Start

### Minimum Go Version

**Go 1.21+** is required for:
- `embed.FS` and `//go:embed` directive (Go 1.16)
- Generics support for type-safe atomic wrappers (Go 1.18)
- `atomic.Int64`/`atomic.Uint64` types with methods (Go 1.19)
- `slices.Contains()` from standard library (Go 1.21)

### Basic Usage

```go
package main

import (
    "context"
    "embed"
    
    "github.com/gin-gonic/gin"
    "github.com/nabbar/golib/static"
)

//go:embed public
var publicFS embed.FS

func main() {
    // Create static handler
    handler := static.New(context.Background(), publicFS, "public")
    
    // Configure security
    handler.SetPathSecurity(static.DefaultPathSecurityConfig())
    handler.SetRateLimit(static.DefaultRateLimitConfig())
    
    // Setup Gin router
    router := gin.Default()
    handler.RegisterRouter("/static", router.GET)
    
    router.Run(":8080")
}
```

### With Security Integration

```go
handler := static.New(ctx, fs, "assets")

// Path security
handler.SetPathSecurity(static.PathSecurityConfig{
    Enabled:       true,
    AllowDotFiles: false,
    MaxPathDepth:  10,
    BlockedPatterns: []string{".git", ".env"},
})

// Rate limiting
handler.SetRateLimit(static.RateLimitConfig{
    Enabled:     true,
    MaxRequests: 100,
    Window:      time.Minute,
})

// WAF integration
handler.SetSecurityBackend(static.SecurityConfig{
    Enabled:    true,
    WebhookURL: "https://waf.example.com/events",
    MinSeverity: "medium",
})
```

### Advanced Configuration

```go
// HTTP caching
handler.SetHeaders(static.HeadersConfig{
    EnableCacheControl: true,
    CacheMaxAge:        3600,
    CachePublic:        true,
    EnableETag:         true,
    DenyMimeTypes: []string{"application/x-executable"},
})

// Suspicious access detection
handler.SetSuspicious(static.SuspiciousConfig{
    Enabled: true,
    SuspiciousPatterns: []string{
        ".env", ".git", "wp-admin",
    },
})

// Index files
handler.SetIndex("", "/", "index.html")
handler.SetIndex("", "/docs", "docs/index.html")

// Downloads
handler.SetDownload("/files/document.pdf", true)

// Redirects
handler.SetRedirect("", "/old-path", "", "/new-path")
```

---

## Performance

### Test Results

| Metric | Value | Notes |
|--------|-------|-------|
| **Test Coverage** | 82.6% | 229/229 tests passing |
| **Race Conditions** | 0 | Verified with `-race` detector |
| **Throughput** | 1,900-5,600 RPS | Single file, no caching |
| **Latency (p50)** | ~100µs | File operation median |
| **Latency (p99)** | <5ms | Large file operations |
| **Memory** | O(1) per request | No allocation spikes |

### Benchmarks

```
Static File Operations
Name                        | N   | Min   | Median | Mean  | StdDev | Max
========================================================================================
File-Has [duration]         | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
File-Info [duration]        | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
File-Find [duration]        | 100 | 0s    | 0s     | 0s    | 0s     | 200µs
PathSecurity [duration]     | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
RateLimit-Allow [duration]  | 100 | 0s    | 0s     | 0s    | 0s     | 200µs
RateLimit-Block [duration]  | 10  | 0s    | 0s     | 0s    | 0s     | 100µs
ETag-Generate [duration]    | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
ETag-Validate [duration]    | 100 | 0s    | 0s     | 0s    | 0s     | 0s
Redirect [duration]         | 500 | 100µs | 100µs  | 200µs | 100µs  | 1.6ms
SpecificHandler [duration]  | 500 | 100µs | 100µs  | 100µs | 100µs  | 600µs
Throughput-RPS              | 1   | 1,938 | 5,692  | 3,815 | varies | 5,692
```

### Performance Characteristics

- **Zero Mutex Overhead** - All operations use atomic primitives
- **O(1) IP Lookup** - Constant time rate limit checks
- **Lazy Initialization** - Configuration loaded on demand
- **Efficient Batching** - Reduces webhook overhead by 90%+
- **304 Responses** - Saves bandwidth with ETag validation

---

## Use Cases

### 1. Single Page Application (SPA)

```go
handler := static.New(ctx, embedFS, "dist")

// Security
handler.SetPathSecurity(static.DefaultPathSecurityConfig())

// Aggressive caching for immutable assets
handler.SetHeaders(static.HeadersConfig{
    EnableCacheControl: true,
    CacheMaxAge:        31536000, // 1 year
    CachePublic:        true,
    EnableETag:         true,
})

// Index file for all routes (SPA routing)
handler.SetIndex("", "/", "index.html")
```

### 2. API Documentation Server

```go
handler := static.New(ctx, docsFS, "docs")

// Moderate caching
handler.SetHeaders(static.HeadersConfig{
    CacheMaxAge: 3600, // 1 hour
    EnableETag:  true,
})

// Rate limiting
handler.SetRateLimit(static.RateLimitConfig{
    MaxRequests: 1000,
    Window:      time.Minute,
})
```

### 3. CDN Origin Server

```go
handler := static.New(ctx, assetsFS, "assets")

// Maximum caching for CDN
handler.SetHeaders(static.HeadersConfig{
    CacheMaxAge: 31536000, // 1 year
    CachePublic: true,
    EnableETag:  true,
})

// Relaxed rate limiting (CDN handles most traffic)
handler.SetRateLimit(static.RateLimitConfig{
    MaxRequests: 10000,
    Window:      time.Minute,
})
```

### 4. Enterprise Web Application

```go
handler := static.New(ctx, appFS, "public")

// Full security stack
handler.SetPathSecurity(static.PathSecurityConfig{
    Enabled:         true,
    AllowDotFiles:   false,
    MaxPathDepth:    10,
    BlockedPatterns: []string{".git", ".svn", "node_modules"},
})

handler.SetRateLimit(static.RateLimitConfig{
    Enabled:     true,
    MaxRequests: 100,
    Window:      time.Minute,
})

handler.SetSuspicious(static.DefaultSuspiciousConfig())

// WAF/SIEM integration
handler.SetSecurityBackend(static.SecurityConfig{
    Enabled:      true,
    WebhookURL:   "https://siem.company.com/events",
    BatchSize:    100,
    BatchTimeout: 30 * time.Second,
    EnableCEFFormat: true,
})
```

### 5. Development Server

```go
handler := static.New(ctx, devFS, "src")

// Minimal security for local dev
handler.SetPathSecurity(static.PathSecurityConfig{
    Enabled:       true,
    AllowDotFiles: false, // Still protect .env
})

// No caching for fast iteration
handler.SetHeaders(static.HeadersConfig{
    EnableCacheControl: false,
})
```

---

## API Reference

### Core Interface

```go
type Static interface {
    StaticFileSystem
    StaticPathSecurity
    StaticRateLimit
    StaticHeaders
    StaticSuspicious
    StaticSecurityBackend
    StaticIndex
    StaticDownload
    StaticRedirect
    StaticSpecific
    StaticRouter
    StaticMonitor
}
```

### Configuration Types

#### PathSecurityConfig

```go
type PathSecurityConfig struct {
    Enabled         bool     // Enable path validation
    AllowDotFiles   bool     // Allow .env, .git, etc.
    MaxPathDepth    int      // Maximum depth (0 = unlimited)
    BlockedPatterns []string // Patterns to block
}
```

#### RateLimitConfig

```go
type RateLimitConfig struct {
    Enabled         bool          // Enable rate limiting
    MaxRequests     int           // Max unique files per window
    Window          time.Duration // Time window
    CleanupInterval time.Duration // Cleanup frequency
    WhitelistIPs    []string      // Bypass IPs
    TrustedProxies  []string      // Trusted proxy IPs
}
```

#### HeadersConfig

```go
type HeadersConfig struct {
    EnableCacheControl bool              // Enable Cache-Control
    CacheMaxAge        int               // Cache duration (seconds)
    CachePublic        bool              // Public or private cache
    EnableETag         bool              // Enable ETag
    EnableContentType  bool              // Enable MIME validation
    AllowedMimeTypes   []string          // Whitelist (empty = all)
    DenyMimeTypes      []string          // Blacklist
    CustomMimeTypes    map[string]string // Custom mappings
}
```

#### SecurityConfig

```go
type SecurityConfig struct {
    Enabled         bool              // Enable security backend
    WebhookURL      string            // Webhook endpoint
    WebhookHeaders  map[string]string // Custom headers
    WebhookTimeout  time.Duration     // Request timeout
    WebhookAsync    bool              // Async sending
    MinSeverity     string            // Minimum severity level
    BatchSize       int               // Batch size (0 = real-time)
    BatchTimeout    time.Duration     // Batch flush interval
    EnableCEFFormat bool              // Use CEF format
    Callbacks       []SecuEvtCallback // Go callbacks
}
```

### Error Codes

```go
const (
    ErrorFileNotFound     // File not found in embedded FS
    ErrorFileOpen         // Cannot open file
    ErrorFileRead         // Cannot read file
    ErrorFiletemp         // Cannot create temp file
    ErrorParamEmpty       // Required parameter empty
    ErrorPathInvalid      // Invalid path
    ErrorPathTraversal    // Path traversal attempt
    ErrorPathDotFile      // Dot file access denied
    ErrorPathDepth        // Path depth exceeded
    ErrorPathBlocked      // Blocked pattern matched
    ErrorMimeTypeDenied   // MIME type not allowed
)
```

### Security Event Types

```go
const (
    EventTypePathTraversal  // Path traversal attack
    EventTypeRateLimit      // Rate limit exceeded
    EventTypeSuspicious     // Suspicious access pattern
    EventTypeMimeTypeDenied // MIME type denied
    EventTypeDotFile        // Dot file access attempt
    EventTypePatternBlocked // Blocked pattern matched
    EventTypePathDepth      // Path depth exceeded
)
```

---

## Best Practices

### ✅ DO

```go
// Use default configurations as starting point
handler.SetPathSecurity(static.DefaultPathSecurityConfig())
handler.SetRateLimit(static.DefaultRateLimitConfig())

// Enable ETag for bandwidth savings
handler.SetHeaders(static.HeadersConfig{
    EnableETag: true,
    CacheMaxAge: 3600,
})

// Use batch processing for high-volume security events
handler.SetSecurityBackend(static.SecurityConfig{
    Enabled: true,
    BatchSize: 100,
    BatchTimeout: 30 * time.Second,
})

// Whitelist localhost for development
handler.SetRateLimit(static.RateLimitConfig{
    WhitelistIPs: []string{"127.0.0.1", "::1"},
})

// Set appropriate cache duration per asset type
handler.SetHeaders(static.HeadersConfig{
    CacheMaxAge: 31536000, // 1 year for versioned assets
})
```

### ❌ DON'T

```go
// Don't disable all security in production
handler.SetPathSecurity(static.PathSecurityConfig{
    Enabled: false, // ❌ Unsafe
})

// Don't allow dot files in production
handler.SetPathSecurity(static.PathSecurityConfig{
    AllowDotFiles: true, // ❌ Exposes .env, .git
})

// Don't set unlimited rate limit
handler.SetRateLimit(static.RateLimitConfig{
    MaxRequests: 0, // ❌ No protection
})

// Don't use sync webhooks in high-traffic scenarios
handler.SetSecurityBackend(static.SecurityConfig{
    WebhookAsync: false, // ❌ Blocks requests
})

// Don't forget cleanup interval
handler.SetRateLimit(static.RateLimitConfig{
    CleanupInterval: 0, // ❌ Memory leak
})
```

### Security Recommendations

1. **Always enable path security** - Even in development
2. **Use rate limiting** - Protect against DoS attacks
3. **Enable suspicious detection** - Identify attack patterns early
4. **Integrate with SIEM** - Use webhook or CEF for monitoring
5. **Regular cleanup** - Configure CleanupInterval for rate limiter
6. **Whitelist carefully** - Only trusted IPs should bypass limits
7. **Block dangerous MIME types** - Prevent executable uploads
8. **Use batch processing** - Reduce security backend overhead

### Performance Recommendations

1. **Enable ETag** - Reduces bandwidth significantly
2. **Use CDN** - Offload static file delivery
3. **Appropriate cache TTL** - Balance freshness vs. performance
4. **Async webhooks** - Non-blocking security event delivery
5. **Batch events** - Reduce webhook call overhead
6. **Monitor throughput** - Use built-in benchmarks

---

## Testing

For comprehensive testing documentation, see [TESTING.md](TESTING.md).

**Test Suite:**
- Total Tests: 229
- Coverage: 82.6%
- Race Detection: ✅ Zero data races
- Execution Time: ~4.7s (standard), ~6.6s (with race)

```bash
# Run all tests
go test -v

# With race detector
CGO_ENABLED=1 go test -race

# With coverage
go test -cover -coverprofile=coverage.out
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Code Contributions

- **No AI-generated code** in core implementation
- AI assistance is acceptable for tests, documentation, and bug fixes
- All contributions must pass existing tests
- Add tests for new features
- Follow existing code style
- Document public APIs with GoDoc

### Testing Requirements

- Maintain >80% code coverage
- Zero race conditions (`go test -race`)
- All tests must pass
- Add benchmarks for performance-critical code

### Documentation

- Update README.md for new features
- Add examples to example_test.go
- Document breaking changes
- Keep TESTING.md current

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test -race -cover ./...`
5. Update documentation
6. Submit pull request with clear description

---

## Future Enhancements

### Planned Features

- **Advanced Rate Limiting**
  - Token bucket algorithm
  - Distributed rate limiting (Redis integration)
  - Per-route rate limits

- **Enhanced Security**
  - Content Security Policy (CSP) headers
  - Subresource Integrity (SRI) support
  - CORS configuration

- **Performance Optimization**
  - Brotli compression support
  - HTTP/2 Server Push hints
  - Memory-mapped file serving for large files

- **Monitoring**
  - Prometheus metrics endpoint
  - Detailed access logging
  - Performance tracing integration

- **Developer Experience**
  - Hot reload support for development
  - Configuration validation
  - More detailed error messages

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/static)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Related Packages**:
  - [github.com/nabbar/golib/router](../router) - Router utilities
  - [github.com/nabbar/golib/logger](../logger) - Logging integration
  - [github.com/nabbar/golib/errors](../errors) - Error handling
  - [github.com/nabbar/golib/atomic](../atomic) - Atomic primitives
  - [github.com/nabbar/golib/monitor](../monitor) - Health monitoring
- **External Resources**:
  - [Gin Web Framework](https://github.com/gin-gonic/gin)
  - [Go embed package](https://pkg.go.dev/embed)
  - [Common Event Format (CEF)](https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors-8.3/cef-implementation-standard/)
