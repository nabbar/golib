/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

/*
Package static provides a secure, high-performance static file server for Gin framework
with embedded filesystem support, comprehensive security features, and WAF/IDS/EDR integration.

# Overview

The static package is designed to serve files from Go's embed.FS with enterprise-grade
security features including:

  - Path traversal protection with configurable rules
  - IP-based rate limiting (DoS/scraping prevention)
  - Suspicious access pattern detection
  - MIME type validation and filtering
  - HTTP caching (ETag, Cache-Control)
  - Integration with external security systems (WAF/IDS/EDR)

All operations are thread-safe using atomic operations without mutexes for maximum
performance and scalability.

# Architecture

The package follows a layered architecture with clear separation of concerns:

	┌─────────────────────────────────────────────────────────────┐
	│                    Gin HTTP Request                          │
	└────────────────────────┬────────────────────────────────────┘
	                         │
	                         ▼
	┌─────────────────────────────────────────────────────────────┐
	│                  Static Handler (Get)                        │
	│  ┌────────────┬───────────────┬─────────────┬─────────────┐ │
	│  │Rate Limit  │Path Security  │Headers      │Suspicious   │ │
	│  │Check       │Validation     │Control      │Detection    │ │
	│  └────────────┴───────────────┴─────────────┴─────────────┘ │
	└────────────────────────┬────────────────────────────────────┘
	                         │
	                         ▼
	┌─────────────────────────────────────────────────────────────┐
	│              File Operations (embed.FS)                      │
	│  ┌────────────┬───────────────┬─────────────┬─────────────┐ │
	│  │Has()       │Find()         │Info()       │Temp()       │ │
	│  └────────────┴───────────────┴─────────────┴─────────────┘ │
	└────────────────────────┬────────────────────────────────────┘
	                         │
	                         ▼
	┌─────────────────────────────────────────────────────────────┐
	│              HTTP Response (SendFile)                        │
	│  ┌────────────┬───────────────┬─────────────┬─────────────┐ │
	│  │ETag        │Cache-Control  │Content-Type │Disposition  │ │
	│  └────────────┴───────────────┴─────────────┴─────────────┘ │
	└─────────────────────────────────────────────────────────────┘

# Request Flow

A typical request flows through multiple security layers:

	HTTP Request
	     │
	     ▼
	[Rate Limit Check]
	     │
	     ├─── Rate exceeded? ──► 429 Too Many Requests
	     │
	     ▼
	[Path Security Validation]
	     │
	     ├─── Path traversal? ──► 403 Forbidden
	     ├─── Dot file? ────────► 403 Forbidden
	     ├─── Blocked pattern? ─► 403 Forbidden
	     │
	     ▼
	[Suspicious Access Detection]
	     │
	     ├─── Log suspicious patterns
	     │
	     ▼
	[File Exists Check]
	     │
	     ├─── Not found? ───────► 404 Not Found
	     │
	     ▼
	[ETag Validation]
	     │
	     ├─── Cached? ──────────► 304 Not Modified
	     │
	     ▼
	[MIME Type Validation]
	     │
	     ├─── Denied type? ─────► 403 Forbidden
	     │
	     ▼
	[Send File with Headers]
	     │
	     ▼
	200 OK with caching headers

# Security Features

The package implements defense-in-depth with multiple security layers:

## 1. Path Traversal Protection

Protects against directory traversal attacks:

  - Detects ".." sequences before path normalization
  - Validates against null byte injection
  - Enforces maximum path depth
  - Blocks access to dot files (.env, .git, etc.)
  - Pattern-based blocking (configurable)

Example:

	handler.SetPathSecurity(static.PathSecurityConfig{
	    Enabled:         true,
	    AllowDotFiles:   false,
	    MaxPathDepth:    10,
	    BlockedPatterns: []string{".git", ".svn", "node_modules"},
	})

## 2. Rate Limiting

IP-based rate limiting tracks unique file paths per IP:

  - Configurable request limits and time windows
  - IP whitelisting support
  - Trusted proxy detection
  - Automatic cache cleanup
  - Thread-safe atomic operations

Example:

	handler.SetRateLimit(static.RateLimitConfig{
	    Enabled:         true,
	    MaxRequests:     100,
	    Window:          time.Minute,
	    CleanupInterval: 5 * time.Minute,
	    WhitelistIPs:    []string{"127.0.0.1"},
	})

## 3. Suspicious Access Detection

Monitors and logs suspicious file access patterns:

  - Configuration file access (.env, config.php)
  - Backup file enumeration (.bak, .old)
  - Admin panel scanning (wp-admin, phpmyadmin)
  - Database file requests (.sql, .db)
  - Logs both successful and failed attempts

## 4. MIME Type Validation

Controls which file types can be served:

  - MIME type detection by file extension
  - Whitelist/blacklist configuration
  - Custom MIME type mapping
  - Blocks dangerous file types (.exe, .sh)

## 5. HTTP Caching

Optimizes bandwidth and performance:

  - ETag generation and validation
  - Cache-Control headers (public/private)
  - 304 Not Modified responses
  - Configurable cache duration
  - Last-Modified support

# Security Backend Integration

The package can report security events to external systems:

	┌──────────────┐         ┌──────────────┐         ┌──────────────┐
	│ Static       │         │ Security     │         │ External     │
	│ Handler      │────────►│ Event        │────────►│ System       │
	│              │         │ Processor    │         │ (WAF/IDS)    │
	└──────────────┘         └──────────────┘         └──────────────┘
	                                │
	                                ├─────► Webhook (JSON/CEF)
	                                │
	                                └─────► Go Callbacks

Supported integrations:

  - Webhooks with custom headers (Authorization, etc.)
  - Common Event Format (CEF) for SIEM systems
  - Go callbacks for custom processing
  - Batch processing for efficiency
  - Configurable severity filtering

# Thread Safety

All data structures use atomic operations without mutexes:

	Configuration:  libatm.Value[*Config]        (atomic.Value wrapper)
	IP Tracking:    libatm.MapTyped[string, *T]  (atomic map)
	Counters:       atomic.Int64, atomic.Uint64  (standard atomic)
	Event Batch:    atomic operations only       (no mutex)

This design ensures:
  - Zero contention under high load
  - Predictable performance
  - No deadlock risk
  - Lock-free scalability

# Performance Considerations

The package is optimized for high-performance scenarios:

1. Atomic Operations: All state managed without mutexes
2. Lazy Initialization: Security features activated only when configured
3. Batch Processing: Multiple events sent together to reduce overhead
4. HTTP Caching: ETag reduces bandwidth and CPU usage
5. Embedded FS: No disk I/O for file access

Typical performance characteristics:
  - Request handling: <1ms per request (cached)
  - Rate limit check: ~100ns (atomic read)
  - Path validation: <10μs (string operations)
  - ETag generation: <1μs (hash calculation)

# Dependencies

This package requires:

  - github.com/gin-gonic/gin - HTTP framework
  - github.com/nabbar/golib/atomic - Thread-safe atomic wrappers
  - github.com/nabbar/golib/context - Context-aware configuration
  - github.com/nabbar/golib/logger - Logging interface
  - github.com/nabbar/golib/errors - Error management
  - github.com/nabbar/golib/router - Router helpers
  - github.com/nabbar/golib/monitor - Health monitoring

# Example Usage

Basic static file server:

	package main

	import (
	    "context"
	    "embed"
	    "github.com/gin-gonic/gin"
	    "github.com/nabbar/golib/static"
	)

	//go:embed assets/*
	var content embed.FS

	func main() {
	    handler := static.New(context.Background(), content, "assets")
	    router := gin.Default()
	    handler.RegisterRouter("/static", router.GET)
	    router.Run(":8080")
	}

With security features:

	handler := static.New(context.Background(), content, "assets")

	// Path security
	handler.SetPathSecurity(static.DefaultPathSecurityConfig())

	// Rate limiting
	handler.SetRateLimit(static.RateLimitConfig{
	    Enabled:     true,
	    MaxRequests: 100,
	    Window:      time.Minute,
	})

	// HTTP caching
	handler.SetHeaders(static.DefaultHeadersConfig())

	// Security backend
	handler.SetSecurityBackend(static.SecurityConfig{
	    Enabled:    true,
	    WebhookURL: "https://waf.example.com/events",
	})

See example_test.go for more comprehensive examples.

# Best Practices

 1. Always enable path security in production:
    handler.SetPathSecurity(static.DefaultPathSecurityConfig())

2. Configure rate limiting appropriate to your traffic:

  - API serving: 100-1000 requests/minute

  - Public websites: 1000-10000 requests/minute

    3. Use HTTP caching to reduce server load:
    handler.SetHeaders(static.DefaultHeadersConfig())

    4. Monitor security events in production:
    handler.SetSecurityBackend() with appropriate webhooks

    5. Whitelist known IPs (monitoring, health checks):
    config.WhitelistIPs = []string{"monitoring-ip"}

    6. Use custom MIME types for modern formats:
    config.CustomMimeTypes = map[string]string{".wasm": "application/wasm"}

    7. Enable suspicious access logging:
    handler.SetSuspicious(static.DefaultSuspiciousConfig())

# Troubleshooting

Common issues and solutions:

403 Forbidden:
  - Check path security configuration
  - Verify file is not a dot file (if AllowDotFiles = false)
  - Check blocked patterns

429 Too Many Requests:
  - Increase rate limit (MaxRequests)
  - Add IP to whitelist
  - Increase time window

404 Not Found:
  - Verify file exists in embed.FS
  - Check embedRootDir parameter in New()
  - Use handler.Has() to verify file presence

# Testing

The package includes comprehensive tests:
  - 229+ test cases
  - Thread safety verified with race detector
  - Concurrency stress tests
  - Benchmark tests
  - Security scenario tests

Run tests:

	go test -v
	go test -race  (with race detector)
	go test -bench . (benchmarks)

# License

MIT License - Copyright (c) 2022 Nicolas JUHEL
*/
package static
