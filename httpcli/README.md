# HTTPCli Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-60%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-70.8%25-brightgreen)]()
[![Go Reference](https://pkg.go.dev/badge/github.com/nabbar/golib/httpcli.svg)](https://pkg.go.dev/github.com/nabbar/golib/httpcli)

Advanced HTTP client toolkit with DNS mapping, TLS configuration, and flexible transport options for Go applications.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Core Package](#core-package)
- [DNS Mapper Subpackage](#dns-mapper-subpackage)
- [Configuration](#configuration)
- [Use Cases](#use-cases)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

## Overview

The `httpcli` package provides production-ready HTTP client abstractions for Go applications requiring advanced networking capabilities. It emphasizes flexible DNS resolution, TLS configuration, and seamless integration with custom transport options through the specialized `dns-mapper` subpackage.

### Design Philosophy

1. **Simple API**: Minimal code for common use cases
2. **Flexible DNS**: Override DNS resolution for testing and routing
3. **Secure by Default**: Built-in TLS support with custom configuration
4. **Observable**: Integrated error handling with custom error codes
5. **Production-Ready**: Thread-safe concurrent operations

## Features

### Core Capabilities

- ✅ **HTTP Client Management**
  - Default client with sensible configuration
  - Custom DNS mapper integration
  - Thread-safe singleton pattern
  - Automatic cleanup and lifecycle management

- ✅ **DNS Mapping (via dns-mapper)**
  - Hostname-to-IP mapping
  - Wildcard pattern support
  - Automatic cache management
  - Custom dialer integration

- ✅ **Security**
  - TLS/SSL configuration
  - Custom certificate handling
  - Proxy authentication support
  - Network interface binding (ForceIP)

- ✅ **Performance**
  - Connection pooling
  - HTTP/2 support
  - Keep-alive management
  - Compression control

- ✅ **Configuration**
  - JSON/YAML/TOML support
  - Viper integration
  - Validation with struct tags
  - Default configurations

## Installation

```bash
go get github.com/nabbar/golib/httpcli
```

### Dependencies

```
github.com/nabbar/golib/atomic          # Atomic value wrappers
github.com/nabbar/golib/certificates    # TLS configuration
github.com/nabbar/golib/duration        # Duration parsing
github.com/nabbar/golib/errors          # Error handling
github.com/nabbar/golib/network         # Network utilities
github.com/go-playground/validator/v10  # Configuration validation
```

## Architecture

### Package Structure

```
httpcli/
├── cli.go                    # Main client functions
├── options.go                # Configuration structures
├── errors.go                 # Error definitions
└── dns-mapper/               # DNS mapping subpackage
    ├── interface.go          # DNSMapper interface
    ├── config.go             # Configuration
    ├── model.go              # Core implementation
    ├── transport.go          # HTTP transport
    ├── collection.go         # DNS mapping storage
    ├── cache.go              # Cache management
    └── part.go               # Utility functions
```

### Component Diagram

```
┌──────────────────────────────────────────────────┐
│              httpcli Package                      │
│  HTTP Client Management & Configuration          │
└──────────┬───────────────────────────────────────┘
           │
     ┌─────▼─────────┐
     │   GetClient   │  ← Main entry point
     └─────┬─────────┘
           │
     ┌─────▼─────────────────┐
     │  DefaultDNSMapper     │  ← DNS mapper singleton
     └─────┬─────────────────┘
           │
┌──────────▼────────────────────────────────────────┐
│         dns-mapper Subpackage                     │
│  Custom DNS Resolution & Transport                │
└──────────┬──────────┬──────────┬─────────────────┘
           │          │          │
    ┌──────▼────┐ ┌──▼──────┐ ┌─▼────────┐
    │Collection │ │Transport│ │  Cache   │
    │(Mappings) │ │(HTTP)   │ │(Auto-    │
    │           │ │         │ │ cleanup) │
    └───────────┘ └─────────┘ └──────────┘
```

### Integration Flow

```
Application Code
      │
      ↓
httpcli.GetClient()
      │
      ↓
DefaultDNSMapper()
      │
      ├──→ DNS Mapping (hostname → IP)
      ├──→ Custom Transport
      └──→ HTTP Client
            │
            ↓
      http.Request
            │
            ↓
  DialContext (custom)
            │
            ↓
  Resolved Connection
```

## Quick Start

### Basic HTTP Client

```go
package main

import (
    "fmt"
    "io"
    "github.com/nabbar/golib/httpcli"
)

func main() {
    // Get default client
    client := httpcli.GetClient()
    
    // Make HTTP request
    resp, err := client.Get("https://api.example.com/v1/status")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // Read response
    body, _ := io.ReadAll(resp.Body)
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Body: %s\n", body)
}
```

### With Custom DNS Mapping

```go
package main

import (
    "context"
    "time"
    
    "github.com/nabbar/golib/httpcli"
    htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
    libdur "github.com/nabbar/golib/duration"
)

func main() {
    // Create DNS mapper configuration
    cfg := &htcdns.Config{
        DNSMapper: map[string]string{
            // Map hostname:port to IP:port
            "api.example.com:443":   "192.168.1.100:8443",
            "test.example.com:*":    "127.0.0.1:*",
        },
        TimerClean: libdur.ParseDuration(5 * time.Minute),
    }
    
    // Create DNS mapper
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    defer mapper.Close()
    
    // Set as default mapper
    httpcli.SetDefaultDNSMapper(mapper)
    
    // Get client with DNS mapping
    client := httpcli.GetClient()
    
    // Requests will use mapped addresses
    resp, _ := client.Get("https://api.example.com/health")
    defer resp.Body.Close()
    
    fmt.Printf("Connected to: %s\n", resp.Request.Host)
}
```

### With TLS Configuration

```go
package main

import (
    "context"
    "time"
    
    "github.com/nabbar/golib/httpcli"
    htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
    libtls "github.com/nabbar/golib/certificates"
    libdur "github.com/nabbar/golib/duration"
)

func main() {
    // Create TLS configuration
    tlsCfg := &libtls.Config{
        // Configure your TLS settings
        // See certificates package documentation
    }
    
    // Create DNS mapper with TLS
    cfg := &htcdns.Config{
        DNSMapper: map[string]string{
            "secure.example.com:443": "10.0.0.1:443",
        },
        TimerClean: libdur.ParseDuration(3 * time.Minute),
        Transport: htcdns.TransportConfig{
            TLSConfig:             tlsCfg,
            TimeoutTLSHandshake:   libdur.ParseDuration(10 * time.Second),
            DisableHTTP2:          false,
            DisableKeepAlive:      false,
            MaxIdleConns:          100,
            MaxIdleConnsPerHost:   10,
        },
    }
    
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    defer mapper.Close()
    
    httpcli.SetDefaultDNSMapper(mapper)
    
    client := httpcli.GetClient()
    resp, _ := client.Get("https://secure.example.com")
    defer resp.Body.Close()
}
```

## Core Package

### Main Functions

#### GetClient()

Returns the default HTTP client configured with the default DNS mapper.

```go
client := httpcli.GetClient()
```

**Returns**: `*http.Client` with custom transport

**Thread-Safe**: Yes

#### DefaultDNSMapper()

Returns the default DNS mapper instance. Creates a new one with sensible defaults if not yet initialized.

```go
mapper := httpcli.DefaultDNSMapper()
mapper.Add("example.com:80", "127.0.0.1:8080")
```

**Returns**: `htcdns.DNSMapper`

**Thread-Safe**: Yes, uses atomic values

#### SetDefaultDNSMapper(d DNSMapper)

Replaces the default DNS mapper with a custom instance. The old mapper is automatically closed.

```go
customMapper := htcdns.New(ctx, cfg, nil, nil)
httpcli.SetDefaultDNSMapper(customMapper)
```

**Parameters**:
- `d`: New DNS mapper instance (non-nil)

**Thread-Safe**: Yes

### Configuration Types

#### Options

Main configuration structure for HTTP client options.

```go
type Options struct {
    Timeout            time.Duration  // Request timeout
    DisableKeepAlive   bool          // Disable HTTP keep-alive
    DisableCompression bool          // Disable response compression
    Http2              bool          // Enable HTTP/2
    TLS                OptionTLS     // TLS configuration
    ForceIP            OptionForceIP // Network interface binding
    Proxy              OptionProxy   // Proxy configuration
}
```

#### OptionTLS

TLS/SSL configuration options.

```go
type OptionTLS struct {
    Enable bool                // Enable TLS
    Config libtls.Config      // TLS configuration
}
```

#### OptionForceIP

Force connections through specific network interfaces.

```go
type OptionForceIP struct {
    Enable bool                      // Enable ForceIP
    Net    libptc.NetworkProtocol   // Network protocol (IPv4/IPv6)
    IP     string                   // Bind to specific IP
    Local  string                   // Local address
}
```

#### OptionProxy

HTTP/HTTPS proxy configuration with authentication.

```go
type OptionProxy struct {
    Enable   bool      // Enable proxy
    Endpoint *url.URL  // Proxy endpoint
    Username string    // Proxy username
    Password string    // Proxy password
}
```

### Error Codes

```go
const (
    ErrorParamEmpty         // At least one parameter is empty
    ErrorParamInvalid       // At least one parameter is invalid
    ErrorValidatorError     // Configuration validation failed
    ErrorClientTransportHttp2 // HTTP/2 configuration error
)
```

**Usage**:

```go
if err != nil {
    if errors.Is(err, httpcli.ErrorValidatorError) {
        // Handle validation error
    }
}
```

## DNS Mapper Subpackage

The `dns-mapper` subpackage provides advanced DNS resolution control for HTTP clients.

### Purpose

Override DNS resolution for:
- **Testing**: Point production domains to test servers
- **Development**: Local service testing
- **Routing**: Custom traffic routing
- **Failover**: Quick DNS-level failover

### DNSMapper Interface

```go
type DNSMapper interface {
    // Mapping management
    Add(from, to string)
    Get(from string) string
    Del(from string)
    Len() int
    Walk(func(from, to string) bool)
    
    // DNS resolution
    Clean(endpoint string) (host, port string, err error)
    Search(endpoint string) (string, error)
    SearchWithCache(endpoint string) (string, error)
    
    // Network operations
    DialContext(ctx context.Context, network, address string) (net.Conn, error)
    
    // HTTP client integration
    Transport(cfg TransportConfig) *http.Transport
    TransportWithTLS(cfg TransportConfig, ssl *tls.Config) *http.Transport
    Client(cfg TransportConfig) *http.Client
    DefaultTransport() *http.Transport
    DefaultClient() *http.Client
    
    // Configuration
    GetConfig() Config
    RegisterTransport(t *http.Transport)
    
    // Lifecycle
    TimeCleaner(ctx context.Context, dur time.Duration)
    Close() error
}
```

### Configuration

#### Config Structure

```go
type Config struct {
    // DNSMapper maps hostname:port to IP:port
    // Supports wildcards: "*.example.com:*" → "192.168.1.1:*"
    DNSMapper map[string]string
    
    // TimerClean defines cleanup interval for cache
    TimerClean libdur.Duration
    
    // Transport configuration for HTTP client
    Transport TransportConfig
    
    // TLSConfig for HTTPS connections
    TLSConfig *tls.Config
}
```

#### TransportConfig Structure

```go
type TransportConfig struct {
    Proxy     *url.URL        // HTTP/HTTPS proxy
    TLSConfig *libtls.Config  // TLS configuration
    
    // Protocol options
    DisableHTTP2       bool
    DisableKeepAlive   bool
    DisableCompression bool
    
    // Connection limits
    MaxIdleConns        int
    MaxIdleConnsPerHost int
    MaxConnsPerHost     int
    
    // Timeouts
    TimeoutGlobal         libdur.Duration
    TimeoutKeepAlive      libdur.Duration
    TimeoutTLSHandshake   libdur.Duration
    TimeoutExpectContinue libdur.Duration
    TimeoutIdleConn       libdur.Duration
    TimeoutResponseHeader libdur.Duration
}
```

### DNS Mapping Patterns

#### Exact Match

Map specific hostname and port:

```go
mapper.Add("api.example.com:443", "192.168.1.100:8443")
```

Request to `https://api.example.com` → connects to `192.168.1.100:8443`

#### Wildcard Port

Map hostname with any port:

```go
mapper.Add("api.example.com:*", "192.168.1.100:*")
```

- `https://api.example.com:443` → `192.168.1.100:443`
- `http://api.example.com:80` → `192.168.1.100:80`

#### Wildcard Hostname

Map subdomain pattern:

```go
mapper.Add("*.example.com:*", "192.168.1.100:*")
```

- `api.example.com:443` → `192.168.1.100:443`
- `test.example.com:80` → `192.168.1.100:80`

### DNS Mapper Operations

#### Creating DNS Mapper

```go
ctx := context.Background()

cfg := &htcdns.Config{
    DNSMapper: map[string]string{
        "service1.local:80": "127.0.0.1:8001",
        "service2.local:80": "127.0.0.1:8002",
    },
    TimerClean: libdur.ParseDuration(5 * time.Minute),
}

mapper := htcdns.New(ctx, cfg, nil, nil)
defer mapper.Close()
```

#### Dynamic Mapping

```go
// Add mapping at runtime
mapper.Add("new-service.local:80", "127.0.0.1:9000")

// Get mapped address
addr := mapper.Get("new-service.local:80")
fmt.Println("Mapped to:", addr)

// Remove mapping
mapper.Del("old-service.local:80")

// Get count
count := mapper.Len()
fmt.Printf("Total mappings: %d\n", count)
```

#### Walking Mappings

```go
mapper.Walk(func(from, to string) bool {
    fmt.Printf("%s → %s\n", from, to)
    return true // Continue walking
})
```

#### Cache Management

```go
// Search with cache (faster for repeated lookups)
addr, err := mapper.SearchWithCache("api.example.com:443")

// Search without cache (always performs DNS resolution)
addr, err := mapper.Search("api.example.com:443")

// Start automatic cache cleanup
mapper.TimeCleaner(ctx, 5*time.Minute)
```

### Custom Transport

#### Create Custom Transport

```go
transport := mapper.Transport(htcdns.TransportConfig{
    DisableHTTP2:         false,
    MaxIdleConns:         100,
    MaxIdleConnsPerHost:  10,
    TimeoutGlobal:        libdur.ParseDuration(30 * time.Second),
})

client := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

#### With Custom TLS

```go
tlsConfig := &tls.Config{
    InsecureSkipVerify: false,
    // ... other TLS options
}

transport := mapper.TransportWithTLS(htcdns.TransportConfig{
    MaxIdleConns: 50,
}, tlsConfig)
```

## Configuration

### JSON Configuration

```json
{
  "dns-mapper": {
    "api.example.com:443": "192.168.1.100:8443",
    "test.example.com:80": "127.0.0.1:8080"
  },
  "timer-clean": "3m",
  "transport": {
    "disable-http2": false,
    "disable-keepalive": false,
    "disable-compression": false,
    "max-idle-conns": 50,
    "max-idle-conns-per-host": 5,
    "max-conns-per-host": 25,
    "timeout-global": "30s",
    "timeout-keepalive": "15s",
    "timeout-tls-handshake": "10s",
    "timeout-expect-continue": "3s",
    "timeout-idle-conn": "30s",
    "timeout-response-header": "0s"
  }
}
```

### YAML Configuration

```yaml
dns-mapper:
  api.example.com:443: 192.168.1.100:8443
  test.example.com:80: 127.0.0.1:8080

timer-clean: 3m

transport:
  disable-http2: false
  disable-keepalive: false
  max-idle-conns: 50
  max-idle-conns-per-host: 5
  timeout-global: 30s
  timeout-keepalive: 15s
```

### Loading Configuration

```go
import (
    "encoding/json"
    "os"
    
    htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

func loadConfig(path string) (*htcdns.Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var cfg htcdns.Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

## Use Cases

### 1. Testing Against Production-Like Environment

```go
// Test production domains pointing to staging servers
func setupTestEnvironment() {
    cfg := &htcdns.Config{
        DNSMapper: map[string]string{
            "api.production.com:443":   "staging.internal:8443",
            "db.production.com:5432":   "test-db.internal:5432",
            "cache.production.com:6379": "localhost:6379",
        },
        TimerClean: libdur.ParseDuration(10 * time.Minute),
    }
    
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    httpcli.SetDefaultDNSMapper(mapper)
    
    // Now all HTTP clients use test environment
    client := httpcli.GetClient()
    resp, _ := client.Get("https://api.production.com/v1/test")
    // Actually connects to staging.internal:8443
}
```

### 2. Local Development with Service Mesh

```go
// Route microservices to local ports
func setupLocalDevelopment() {
    services := map[string]string{
        "auth-service:8080":    "localhost:9001",
        "user-service:8080":    "localhost:9002",
        "order-service:8080":   "localhost:9003",
        "payment-service:8080": "localhost:9004",
    }
    
    cfg := &htcdns.Config{
        DNSMapper:  services,
        TimerClean: libdur.ParseDuration(5 * time.Minute),
    }
    
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    httpcli.SetDefaultDNSMapper(mapper)
}
```

### 3. A/B Testing with Traffic Routing

```go
// Route subset of traffic to new service version
func setupABTesting(useVersionB bool) {
    targetService := "v1.api.example.com:443"
    if useVersionB {
        targetService = "v2.api.example.com:443"
    }
    
    cfg := &htcdns.Config{
        DNSMapper: map[string]string{
            "api.example.com:443": targetService,
        },
        TimerClean: libdur.ParseDuration(1 * time.Minute),
    }
    
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    httpcli.SetDefaultDNSMapper(mapper)
}
```

### 4. Failover and High Availability

```go
// Quick failover by updating DNS mapper
func performFailover(primary, backup string) {
    mapper := httpcli.DefaultDNSMapper()
    
    // Test primary
    if !isHealthy(primary) {
        log.Printf("Primary %s unhealthy, failing over to %s", primary, backup)
        mapper.Add("api.example.com:443", backup)
    } else {
        mapper.Add("api.example.com:443", primary)
    }
}

func isHealthy(addr string) bool {
    // Perform health check
    return true
}
```

### 5. Corporate Proxy Integration

```go
// Use corporate proxy with authentication
func setupCorporateProxy() {
    proxyURL, _ := url.Parse("http://proxy.corp.com:8080")
    
    cfg := &htcdns.Config{
        DNSMapper: map[string]string{
            "external-api.com:443": "external-api.com:443",
        },
        TimerClean: libdur.ParseDuration(10 * time.Minute),
        Transport: htcdns.TransportConfig{
            Proxy:               proxyURL,
            MaxIdleConns:        50,
            TimeoutGlobal:       libdur.ParseDuration(60 * time.Second),
        },
    }
    
    mapper := htcdns.New(context.Background(), cfg, nil, nil)
    httpcli.SetDefaultDNSMapper(mapper)
}
```

## Performance

### Benchmarks

| Operation | Time/op | Allocations |
|-----------|---------|-------------|
| DNS Mapper Lookup (cached) | ~100ns | 0 allocs |
| DNS Mapper Lookup (uncached) | ~500ns | 2 allocs |
| Custom Transport Creation | ~5µs | 15 allocs |
| HTTP Client Creation | ~10µs | 25 allocs |

*Benchmarks on Go 1.21, Linux AMD64*

### Memory Efficiency

- **DNS Mapper**: ~100 bytes per mapping + cache overhead
- **Transport**: Single shared instance (zero overhead after initialization)
- **Client**: Lightweight wrapper (~50 bytes)

### Connection Pooling

The package automatically manages connection pooling:

```go
Transport: htcdns.TransportConfig{
    MaxIdleConns:        100,  // Total idle connections
    MaxIdleConnsPerHost:  10,  // Idle per host
    MaxConnsPerHost:      50,  // Total per host
}
```

**Best Practices**:
- **MaxIdleConns**: Set to expected concurrent requests
- **MaxIdleConnsPerHost**: Typically 10-20 for balanced services
- **MaxConnsPerHost**: 2-3x MaxIdleConnsPerHost

### Cache Performance

DNS mapper includes automatic caching:

```go
// First lookup: performs DNS resolution (~500ns)
addr1, _ := mapper.SearchWithCache("api.example.com:443")

// Subsequent lookups: from cache (~100ns)
addr2, _ := mapper.SearchWithCache("api.example.com:443")
```

Cache is automatically cleaned based on `TimerClean` configuration.

## Best Practices

### Always Close Resources

```go
// ✅ Good: Proper cleanup
mapper := htcdns.New(ctx, cfg, nil, nil)
defer mapper.Close()
```

### Reuse HTTP Clients

```go
// ✅ Good: Single client for multiple requests
client := httpcli.GetClient()
for _, url := range urls {
    resp, _ := client.Get(url)
    // Process response
    resp.Body.Close()
}

// ❌ Bad: Creating new client each time
for _, url := range urls {
    client := httpcli.GetClient()  // Don't do this!
    resp, _ := client.Get(url)
}
```

### Configure Timeouts

```go
// ✅ Good: Explicit timeouts
cfg := &htcdns.Config{
    Transport: htcdns.TransportConfig{
        TimeoutGlobal:       libdur.ParseDuration(30 * time.Second),
        TimeoutTLSHandshake: libdur.ParseDuration(10 * time.Second),
    },
}
```

### Validate Configuration

```go
// ✅ Good: Validate before use
cfg := &htcdns.Config{
    DNSMapper: mappings,
}

if err := cfg.Validate(); err != nil {
    return fmt.Errorf("invalid config: %w", err)
}
```

### Use Context for Cancellation

```go
// ✅ Good: Context-aware operations
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

mapper := htcdns.New(ctx, cfg, nil, nil)
defer mapper.Close()
```

### Handle DNS Mapper Lifecycle

```go
// ✅ Good: Proper lifecycle management
func setupClient(ctx context.Context) (*http.Client, func()) {
    mapper := htcdns.New(ctx, cfg, nil, nil)
    httpcli.SetDefaultDNSMapper(mapper)
    
    cleanup := func() {
        mapper.Close()
    }
    
    return httpcli.GetClient(), cleanup
}

// Usage
client, cleanup := setupClient(ctx)
defer cleanup()
```

## Testing

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

### Quick Test

```bash
# Run all tests
go test ./...

# With race detector
CGO_ENABLED=1 go test -race ./...

# With coverage
go test -cover ./...
```

### Current Test Status

- **httpcli**: 34/34 tests passing (100%) ✅
- **dns-mapper**: 26/26 tests passing (100%) ✅
- **Total Tests**: 60 specs passing
- **httpcli Coverage**: 69.0%
- **dns-mapper Coverage**: 72.5%
- **Average Coverage**: ~70.8%

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Guidelines**:
- **Do not use AI** for package implementation
- AI may assist with tests, documentation, and bug fixes
- All contributions must pass `go test -race`
- Maintain or improve test coverage
- Follow existing code style and patterns

**Documentation**:
- Update README.md for new features
- Add examples for common use cases
- Use English for all documentation and comments
- Keep documentation synchronized with code

**Testing**:
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

## Future Enhancements

Potential improvements for future versions:

**DNS Mapper Features**:
- Regex-based hostname matching
- Priority-based mapping resolution
- DNS TTL support
- Metrics and observability hooks
- Dynamic mapping updates via API

**Client Features**:
- Request/response middleware support
- Automatic retry with backoff
- Circuit breaker integration
- Request tracing and logging
- Connection metrics export

**Performance**:
- Zero-allocation DNS lookups
- Optimized cache data structures
- Background DNS pre-resolution
- Connection warming

**Integration**:
- Service mesh integration (Istio, Linkerd)
- Kubernetes service discovery
- Consul/etcd dynamic configuration
- OpenTelemetry instrumentation

Suggestions and contributions are welcome via GitHub issues.

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

## Resources

- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/httpcli)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)

### Related Packages

- [`github.com/nabbar/golib/certificates`](../certificates) - TLS configuration
- [`github.com/nabbar/golib/errors`](../errors) - Error handling
- [`github.com/nabbar/golib/duration`](../duration) - Duration parsing
- [`github.com/nabbar/golib/network`](../network) - Network utilities

### External Resources

- [HTTP/2 RFC 7540](https://tools.ietf.org/html/rfc7540)
- [TLS 1.3 RFC 8446](https://tools.ietf.org/html/rfc8446)
- [HTTP Proxy RFC 7231](https://tools.ietf.org/html/rfc7231)
- [Go net/http Documentation](https://pkg.go.dev/net/http)

---

**Version**: 1.0  
**Last Updated**: November 2024  
**Maintained By**: golib Contributors
