# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-324%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-74.2%25-yellowgreen)]()
[![Thread-Safe](https://img.shields.io/badge/Thread--Safe-Verified-brightgreen)]()

Comprehensive testing documentation for the socket/client package, covering test execution, race detection, and platform-specific testing.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Platform Testing](#platform-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The socket/client package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for protocol implementations and thread safety validation.

**Test Suite Statistics**
- Total Specs: 324
- Passed: 324 ✅
- Failed: 0
- Coverage: 74.2% (target: ≥80%)
- Execution Time: ~112s (without race), ~180s (with race)
- Thread Safety: ✅ Zero data races verified
- Platforms: All protocols tested on Linux/Darwin

**Test Areas**
- TCP client with TLS support (119 specs, 74.0% coverage, 88.3s)
- UDP datagram communication (73 specs, 73.7% coverage, 8.1s)
- UNIX stream sockets - Linux/Darwin (67 specs, 76.3% coverage, 13.2s)
- UNIX datagram sockets - Linux/Darwin (65 specs, 76.8% coverage, 2.9s)
- Factory functions and platform detection
- Error handling and callbacks
- Thread safety with atomic.Map
- Context cancellation and timeouts

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Race detection (recommended, requires CGO)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -cover -race
```

**Test Results Summary**
```
PASSED: github.com/nabbar/golib/socket/client/tcp       (119 specs, 74.0% coverage, 88.3s)
PASSED: github.com/nabbar/golib/socket/client/udp       (73 specs, 73.7% coverage, 8.1s)
PASSED: github.com/nabbar/golib/socket/client/unix      (67 specs, 76.3% coverage, 13.2s)
PASSED: github.com/nabbar/golib/socket/client/unixgram  (65 specs, 76.8% coverage, 2.9s)

Total: 324 specs, 74.2% coverage, ~112s
```

All tests pass successfully. The package is production-ready with verified thread safety.

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering
- Platform-specific test suites

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Async assertions for goroutines

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# Specific subpackage
go test -v ./tcp
go test -v ./udp
go test -v ./unix     # Linux/Darwin only
go test -v ./unixgram # Linux/Darwin only

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=tcp/callbacks_test.go

# Pattern matching
ginkgo --focus="TLS"
ginkgo --focus="error handling"

# Parallel execution
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml

# Verbose output
ginkgo -v --trace
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Specific subpackage
CGO_ENABLED=1 go test -race ./tcp
```

**Validates**:
- Atomic map operations (`atomic.Map[uint8]`)
- Callback goroutines (async notifications)
- Connection state transitions
- Concurrent client operations

**Expected Output**:
```bash
# ✅ Success (expected)
ok  	github.com/nabbar/golib/socket/client/tcp	8.234s

# ❌ Race detected (should not happen)
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races expected (verified by design with atomic.Map)

### Platform-Specific Testing

```bash
# Linux (all protocols)
go test ./tcp ./udp ./unix ./unixgram

# Darwin/macOS (all protocols)
go test ./tcp ./udp ./unix ./unixgram

# Windows (network protocols only)
go test ./tcp ./udp
# unix/ and unixgram/ return nil on Windows
```

---

## Test Coverage

**Current**: 74.2% statement coverage (target: ≥80%)

### Coverage By Subpackage

| Subpackage | Specs | Coverage | Execution Time | Test Files |
|------------|-------|----------|----------------|------------|
| **tcp** | 119 | 74.0% | 88.3s | `callbacks_test.go`, `communication_test.go`, `connection_test.go`, `creation_test.go`, `errors_test.go`, `tls_test.go`, `concurrency_test.go` |
| **udp** | 73 | 73.7% | 8.1s | `callbacks_test.go`, `communication_test.go`, `connection_test.go`, `errors_test.go` |
| **unix** | 67 | 76.3% | 13.2s | `callbacks_test.go`, `communication_test.go`, `connection_test.go`, `errors_test.go` |
| **unixgram** | 65 | 76.8% | 2.9s | `callbacks_test.go`, `communication_test.go`, `connection_test.go`, `errors_test.go` |
| **Total** | **324** | **74.2%** | **~112s** | All subpackages combined |

### Test Categories

**Connection Management**
- `Connect()` with context timeout
- `IsConnected()` state checking
- `Close()` cleanup and idempotency
- Connection reuse and replacement
- Context cancellation handling

**I/O Operations**
- `Read()` blocking and non-blocking
- `Write()` full buffer transmission
- `Once()` request/response pattern
- Large data transfers
- Empty/nil buffer handling

**Error Handling**
- Invalid addresses (malformed, empty)
- Connection failures (refused, timeout)
- I/O errors (broken pipe, EOF)
- Nil client detection (ErrInstance)
- State validation (ErrConnection)

**Callbacks**
- Error callback registration
- Info callback registration
- Async notification delivery
- Callback goroutine isolation
- Multiple callback invocations

**TLS (TCP only)**
- TLS configuration
- Certificate validation
- Server name indication (SNI)
- TLS handshake errors
- Encrypted I/O operations

**Thread Safety**
- Concurrent Connect() calls
- Parallel Read()/Write() operations
- Callback goroutines
- Atomic state transitions
- Race-free cleanup

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage
go test -cover ./tcp
go test -cover ./udp
go test -cover ./unix
go test -cover ./unixgram
```

---

## Thread Safety

Thread safety is critical for concurrent socket operations.

### Atomic State Management

```go
// All clients use atomic.Map[uint8] for state
type cli struct {
    m libatm.Map[uint8]
}

// Thread-safe operations
o.m.Store(keyNetAddr, address)       // Set
o.m.Load(keyNetConn)                 // Get
o.m.Swap(keyNetConn, newConn)        // Atomic replace
o.m.LoadAndDelete(keyNetConn)        // Atomic remove
```

**Benefits**:
- Lock-free reads and writes
- No mutex contention
- No race conditions
- Goroutine-safe by design

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| Client state | `atomic.Map[uint8]` | ✅ Race-free |
| Error callbacks | Goroutine isolation | ✅ Non-blocking |
| Info callbacks | Goroutine isolation | ✅ Non-blocking |
| Connection management | Atomic swap | ✅ Thread-safe |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on specific protocol
CGO_ENABLED=1 go test -race -v ./tcp
CGO_ENABLED=1 go test -race -v ./udp

# Stress test (10 iterations)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: Zero data races expected across all test runs

---

## Platform Testing

### Linux

**Supported Protocols**: TCP, UDP, UNIX, UnixGram

```bash
# All protocols available
go test ./tcp ./udp ./unix ./unixgram

# UNIX sockets functional
go test -v -run "UNIX" ./unix
go test -v -run "UnixGram" ./unixgram
```

### Darwin/macOS

**Supported Protocols**: TCP, UDP, UNIX, UnixGram

```bash
# All protocols available
go test ./tcp ./udp ./unix ./unixgram

# Build tags include darwin
# unix/ and unixgram/ compiled with linux || darwin
```

### Windows

**Supported Protocols**: TCP, UDP only

```bash
# Network protocols only
go test ./tcp ./udp

# UNIX sockets return nil (stub implementation)
# unix.New("/tmp/test.sock") returns nil
# unixgram.New("/tmp/test.sock") returns nil
```

### Build Tags

| File | Build Tag | Platforms |
|------|-----------|-----------|
| `interface_linux.go` | `linux` | Linux |
| `interface_darwin.go` | `darwin` | macOS |
| `interface_other.go` | `!linux && !darwin` | Windows, BSD, etc. |
| `unix/error.go` | `linux \|\| darwin` | Linux, macOS |
| `unix/ignore.go` | `!linux && !darwin` | Windows, etc. |
| `unixgram/error.go` | `linux \|\| darwin` | Linux, macOS |
| `unixgram/ignore.go` | `!linux && !darwin` | Windows, etc. |

### Cross-Platform Testing

```bash
# Test on current platform
go test ./...

# Cross-compile (doesn't run tests)
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=arm64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

---

## Test File Organization

| Subpackage | Test Files | Specs | Coverage |
|------------|------------|-------|----------|
| **tcp** | `suite_test.go` | - | - |
| | `callbacks_test.go` | 17 | Error and info callbacks |
| | `communication_test.go` | 31 | Read/Write operations, streaming |
| | `connection_test.go` | 16 | Connect/Close lifecycle |
| | `creation_test.go` | 9 | Client creation, validation |
| | `errors_test.go` | 23 | Error handling, edge cases |
| | `tls_test.go` | 16 | TLS configuration, encryption |
| | `concurrency_test.go` | 7 | Thread safety, parallel operations |
| **udp** | `suite_test.go` | - | - |
| | `callbacks_test.go` | 15 | Error and info callbacks |
| | `communication_test.go` | 23 | Read/Write operations |
| | `connection_test.go` | 14 | Connect/Close lifecycle |
| | `errors_test.go` | 21 | Error handling, edge cases |
| **unix** | `suite_test.go` | - | - |
| | `callbacks_test.go` | 15 | Error and info callbacks |
| | `communication_test.go` | 21 | Read/Write operations, binary data |
| | `connection_test.go` | 14 | Connect/Close lifecycle |
| | `errors_test.go` | 17 | Error handling, context cancellation |
| **unixgram** | `suite_test.go` | - | - |
| | `callbacks_test.go` | 15 | Error and info callbacks |
| | `communication_test.go` | 19 | Datagram I/O operations |
| | `connection_test.go` | 14 | Connect/Close lifecycle |
| | `errors_test.go` | 17 | Error handling, datagram-specific |

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should establish TCP connection with timeout", func() {
    // Test implementation
})

It("should send and receive UDP datagrams", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should detect connection error", func() {
    // Arrange
    cli, err := tcp.New("localhost:9999")
    Expect(err).ToNot(HaveOccurred())
    
    // Act
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    err = cli.Connect(ctx)
    
    // Assert
    Expect(err).To(HaveOccurred())
})
```

**3. Use Appropriate Matchers**
```go
Expect(client).ToNot(BeNil())
Expect(err).ToNot(HaveOccurred())
Expect(isConnected).To(BeTrue())
Expect(bytesWritten).To(Equal(len(data)))
Expect(response).To(ContainSubstring("OK"))
```

**4. Always Cleanup Resources**
```go
var cli socket.Client

BeforeEach(func() {
    cli, _ = tcp.New("localhost:8080")
})

AfterEach(func() {
    if cli != nil {
        cli.Close()
    }
})
```

**5. Test Edge Cases**
- Empty addresses
- Nil clients
- Already closed connections
- Context cancellation
- Large data transfers
- Concurrent operations

**6. Platform-Specific Tests**
```go
// Skip on platforms without UNIX sockets
var _ = Describe("UNIX socket", func() {
    BeforeEach(func() {
        if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
            Skip("UNIX sockets not available on " + runtime.GOOS)
        }
    })
    
    It("should connect to UNIX socket", func() {
        cli := unix.New("/tmp/test.sock")
        Expect(cli).ToNot(BeNil())
    })
})
```

### Test Template

```go
package tcp_test

import (
    "context"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "github.com/nabbar/golib/socket/client/tcp"
)

func TestTCP(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "TCP Client Suite")
}

var _ = Describe("TCP Client", func() {
    Context("When creating client", func() {
        It("should validate address", func() {
            // Arrange
            invalidAddr := ""
            
            // Act
            cli, err := tcp.New(invalidAddr)
            
            // Assert
            Expect(err).To(HaveOccurred())
            Expect(cli).To(BeNil())
        })
        
        It("should create valid client", func() {
            cli, err := tcp.New("localhost:8080")
            Expect(err).ToNot(HaveOccurred())
            Expect(cli).ToNot(BeNil())
            defer cli.Close()
        })
    })
    
    Context("When connecting", func() {
        var cli tcp.ClientTCP
        
        BeforeEach(func() {
            cli, _ = tcp.New("localhost:8080")
        })
        
        AfterEach(func() {
            if cli != nil {
                cli.Close()
            }
        })
        
        It("should handle connection timeout", func() {
            ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
            defer cancel()
            
            err := cli.Connect(ctx)
            // Connection may succeed or timeout depending on server availability
        })
        
        It("should check connection state", func() {
            isConnected := cli.IsConnected()
            Expect(isConnected).To(BeFalse())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Avoid shared mutable state
- ✅ Create clients on-demand
- ❌ Don't rely on test execution order

**Resource Management**
```go
// ✅ Good: Cleanup in AfterEach
var cli socket.Client

BeforeEach(func() {
    cli, _ = tcp.New("localhost:8080")
})

AfterEach(func() {
    if cli != nil && cli.IsConnected() {
        cli.Close()
    }
})

// ❌ Bad: No cleanup
It("test", func() {
    cli, _ := tcp.New("localhost:8080")
    cli.Connect(context.Background())
    // Leak!
})
```

**Timeout Protection**
```go
// ✅ Good: Always use timeouts
It("should connect within timeout", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := cli.Connect(ctx)
    Expect(err).ToNot(HaveOccurred())
})

// ❌ Bad: No timeout
It("should connect", func() {
    err := cli.Connect(context.Background())
    // May hang forever!
})
```

**Callback Testing**
```go
// ✅ Good: Verify callbacks
It("should invoke error callback", func() {
    called := false
    cli.RegisterFuncError(func(errs ...error) {
        called = true
    })
    
    // Trigger error
    cli.Write([]byte("data")) // Not connected
    
    Eventually(func() bool { return called }).Should(BeTrue())
})
```

**Concurrency Testing**
```go
// ✅ Good: Test concurrent operations
It("should handle concurrent connects", func() {
    var wg sync.WaitGroup
    errors := make([]error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cli, _ := tcp.New("localhost:8080")
            defer cli.Close()
            
            ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
            defer cancel()
            
            errors[id] = cli.Connect(ctx)
        }(i)
    }
    
    wg.Wait()
    // Verify no data races (run with -race)
})
```

**Platform-Aware**
```go
// ✅ Good: Check platform support
It("should create UNIX socket client on supported platforms", func() {
    cli := unix.New("/tmp/test.sock")
    
    if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
        Expect(cli).ToNot(BeNil())
        defer cli.Close()
    } else {
        Expect(cli).To(BeNil())
    }
})
```

---

## Troubleshooting

### Common Issues

**Port/Socket Already in Use**
```bash
# Find and kill process using port
lsof -ti:8080 | xargs kill -9

# Clean up UNIX socket files
rm -f /tmp/socket-client-test-*.sock
```

**Cause**: Previous test run didn't clean up properly

**Resolution**: Tests now use unique addresses/paths per run with proper cleanup in `AfterSuite`

### Race Conditions

```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

**Expected**: Zero data races (atomic.Map design prevents races)

**If Found**:
- Check for unprotected shared state
- Verify atomic.Map usage
- Review callback goroutine isolation

### CGO Not Available

```bash
# Install build tools
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Set environment
export CGO_ENABLED=1
go test -race ./...
```

### Test Timeouts

```bash
# Identify hanging tests
ginkgo --timeout=30s

# Verbose debugging
ginkgo -v --trace --timeout=30s
```

**Common Causes**:
- Missing server (connection hangs)
- No context timeout
- Goroutine leak
- Deadlock

**Fix**: Always use `context.WithTimeout`

### Platform-Specific Failures

**UNIX Sockets on Windows**:
- Expected: `unix.New()` returns `nil`
- Tests should check for `nil` or skip on Windows

**Build Tag Issues**:
```bash
# Verify correct files compiled
go list -f '{{.GoFiles}}' ./unix
go list -f '{{.GoFiles}}' ./unixgram
```

### Network Connectivity

**Tests Requiring Servers**:
- Mock server in `BeforeSuite`
- Skip if server unavailable
- Use localhost only (no external deps)

```go
var server *net.Listener

BeforeSuite(func() {
    var err error
    server, err = net.Listen("tcp", "localhost:8080")
    if err != nil {
        Skip("Cannot start test server: " + err.Error())
    }
    go acceptConnections(server)
})

AfterSuite(func() {
    if server != nil {
        server.Close()
    }
})
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Socket Client Tests
on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.18', '1.19', '1.20', '1.21']
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}
      
      - name: Run tests
        run: go test -v ./...
        working-directory: socket/client
      
      - name: Race detection (Linux/macOS only)
        if: runner.os != 'Windows'
        run: CGO_ENABLED=1 go test -race ./...
        working-directory: socket/client
      
      - name: Coverage
        run: go test -coverprofile=coverage.out ./...
        working-directory: socket/client
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./socket/client/coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running socket/client tests..."

cd socket/client || exit 1

# Run tests
go test ./... || {
    echo "❌ Tests failed"
    exit 1
}

# Race detection (Linux/macOS)
if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "darwin"* ]]; then
    CGO_ENABLED=1 go test -race ./... || {
        echo "❌ Race detection failed"
        exit 1
    }
fi

# Coverage check
coverage=$(go test -cover ./... | grep -o '[0-9]\+\.[0-9]\+%' | head -1 | sed 's/%//')
if (( $(echo "$coverage < 80" | bc -l) )); then
    echo "❌ Coverage below 80%: $coverage%"
    exit 1
fi

echo "✅ All checks passed"
```

---

## Quality Checklist

Before merging code:

- [x] All tests pass: `go test ./...` (324/324 ✅)
- [x] Race detection clean: `CGO_ENABLED=1 go test -race ./...` (zero races)
- [ ] Coverage target: ≥80% (current: 74.2%)
- [x] New features have tests
- [x] Error cases tested
- [x] Thread safety validated
- [x] TLS implementation tested (TCP only)
- [x] Platform-specific code tested (Linux/Darwin for UNIX sockets)
- [x] Callbacks tested for async execution
- [x] Context cancellation tested
- [x] Documentation updated

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)
- [Go Coverage](https://go.dev/blog/cover)

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync/atomic Package](https://pkg.go.dev/sync/atomic)

**Networking**
- [net Package](https://pkg.go.dev/net)
- [UNIX Domain Sockets](https://man7.org/linux/man-pages/man7/unix.7.html)
- [TCP/IP Illustrated](http://www.kohala.com/start/tcpipiv1.html)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Socket Client Package Contributors
