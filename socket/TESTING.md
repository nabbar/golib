# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-502%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A5%2070%25-brightgreen)]()

Comprehensive testing documentation for the socket package, covering test execution, race detection, platform-specific testing, and quality assurance across client and server implementations.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Platform-Specific Testing](#platform-specific-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The socket package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions across both client and server implementations.

**Test Suite Summary**
- Total Specs: 502
- Coverage: ≥70% (71.2%-84.6% by subpackage)
- Race Detection: ✅ Zero data races
- Execution Time: ~146s (without race), ~219s (with race)

**Coverage Areas**
- Client operations (TCP, UDP, Unix, Unixgram)
- Server operations (TCP, UDP, Unix, Unixgram)
- Connection lifecycle (connect, read, write, close)
- Graceful shutdown with connection draining
- TLS configuration and encryption (TCP)
- Error handling and edge cases
- Callback mechanisms (error, state, server lifecycle)
- Thread safety validation (atomic operations)
- Platform-specific implementations

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -cover -race
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages

---

## Running Tests

### Basic Commands

```bash
# Standard test run
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Subpackage-Specific Tests

```bash
# Client tests
go test -v ./client/tcp/
go test -v ./client/udp/
go test -v ./client/unix/
go test -v ./client/unixgram/

# Server tests
go test -v ./server/tcp/
go test -v ./server/udp/
go test -v ./server/unix/
go test -v ./server/unixgram/
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=tcp_test.go

# Pattern matching
ginkgo --focus="TLS"

# Parallel execution
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Timeout for long-running race tests
CGO_ENABLED=1 go test -race -timeout=10m ./...
```

**Validates**:
- Atomic operations (`atomic.Bool`, `atomic.Int64`, `atomic.Map`)
- Mutex protection (`sync.Mutex`)
- Goroutine synchronization (`context.Context`, `sync.WaitGroup`)
- Connection state management

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/socket/client/tcp	90.123s
ok  	github.com/nabbar/golib/socket/server/tcp	28.321s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected across all 502 specs

### Performance & Profiling

```bash
# Benchmarks (if available)
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out
```

**Performance Expectations**

| Component | Subpackage | Duration | Duration (race) | Notes |
|-----------|------------|----------|-----------------|-------|
| Client | TCP | ~88.3s | ~90s | TLS handshakes, timeouts |
| Client | UDP | ~8.1s | ~9s | Fast datagram tests |
| Client | Unix | ~13.2s | ~14s | IPC connection tests |
| Client | Unixgram | ~2.9s | ~4s | Fastest IPC tests |
| Server | TCP | ~27.7s | ~28.3s | Connection draining |
| Server | UDP | ~1.4s | ~2.5s | Stateless tests |
| Server | Unix | ~2.0s | ~3s | IPC server tests |
| Server | Unixgram | ~2.4s | ~3.4s | IPC datagram tests |
| **Total** | **All** | **~146s** | **~219s** | 2x slower with race (normal) |

---

## Test Coverage

**Target**: ≥70% statement coverage

### Coverage By Component

**Client Package** (324 specs, 74.2% average coverage)

| Subpackage | Specs | Coverage | Description |
|------------|-------|----------|-------------|
| **tcp** | 119 | 74.0% | Connection-oriented with TLS |
| **udp** | 73 | 73.7% | Connectionless datagram |
| **unix** | 67 | 76.3% | IPC stream socket |
| **unixgram** | 65 | 76.8% | IPC datagram socket |

**Server Package** (178 specs, ≥70% average coverage)

| Subpackage | Specs | Coverage | Description |
|------------|-------|----------|-------------|
| **tcp** | 117 | 84.6% | Connection-oriented with TLS (highest coverage) |
| **udp** | 18 | 72.0% | Connectionless datagram |
| **unix** | 23 | 73.8% | IPC stream socket |
| **unixgram** | 20 | 71.2% | IPC datagram socket |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage by subpackage
go test -coverprofile=coverage.out ./client/tcp/
go tool cover -func=coverage.out
```

### Coverage Areas

**Connection Management**
- Connect/Disconnect operations
- Connection state tracking (`IsConnected()`)
- Connection lifecycle callbacks
- Error handling during connection

**I/O Operations**
- Read/Write with various buffer sizes
- One-shot `Once()` operations (client)
- EOF handling and partial reads/writes
- Timeout scenarios with context

**Server Lifecycle**
- Start/Stop operations
- Graceful shutdown with connection draining
- `OpenConnections()` accuracy
- `IsRunning()` and `IsGone()` state tracking
- Done channel signaling

**TLS Support (TCP)**
- TLS configuration
- Certificate validation
- Encrypted communication
- TLS handshake errors

**Callbacks**
- Error callback registration and invocation
- State callback for connection lifecycle
- Server info callback for lifecycle events
- Multiple callback registration

**Platform-Specific**
- Unix socket availability checks
- Platform-specific error handling
- File permissions (Unix sockets)
- Socket path validation

**Edge Cases**
- Nil instance checks
- Invalid addresses
- Connection refused scenarios
- Closed connection handling
- Context cancellation
- Timeout handling

---

## Thread Safety

Thread safety is critical for both client and server implementations that handle concurrent operations.

### Concurrency Primitives

```go
// Atomic state management
atomic.Bool       // Connection state flags
atomic.Int64      // Connection counters (servers)
atomic.Map[uint8] // Client state storage

// Synchronization
sync.Mutex        // Callback registration protection
sync.WaitGroup    // Goroutine lifecycle management
context.Context   // Cancellation and timeout control
```

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| **Client State** | `atomic.Map[uint8]` | ✅ Race-free |
| **Server Connections** | `atomic.Int64` | ✅ Race-free |
| **Server Running State** | `atomic.Bool` | ✅ Race-free |
| **Callback Invocation** | Independent goroutines | ✅ Parallel-safe |
| **Connection Handlers** | Per-connection goroutines | ✅ Isolated |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on specific component
CGO_ENABLED=1 go test -race -v ./client/tcp/
CGO_ENABLED=1 go test -race -v ./server/tcp/

# Stress test (run multiple times)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: Zero data races across all test runs (502 specs)

---

## Platform-Specific Testing

### Unix Socket Support

Unix domain sockets (stream and datagram) are only available on Linux and Darwin/macOS.

**Supported Platforms**
- ✅ Linux (all protocols)
- ✅ Darwin/macOS (all protocols)
- ❌ Windows (TCP and UDP only)
- ❌ Other platforms (TCP and UDP only)

**Platform Detection**
```go
// Unix socket clients return nil on unsupported platforms
cli := unix.New("/tmp/app.sock")
if cli == nil {
    // Platform doesn't support Unix sockets
    // Fall back to TCP
}
```

**Testing on Different Platforms**

```bash
# Linux/Darwin - all tests
go test ./...

# Windows - TCP and UDP only
go test ./client/tcp/ ./client/udp/
go test ./server/tcp/ ./server/udp/
```

**CI/CD Considerations**
- Run full test suite on Linux/Darwin
- Run TCP/UDP tests on Windows
- Use build tags if needed for platform-specific code

---

## Test File Organization

### Client Package

| File | Purpose | Specs | Coverage |
|------|---------|-------|----------|
| **tcp/** | TCP client tests | 119 | 74.0% |
| `tcp_suite_test.go` | Suite initialization | - | - |
| `tcp_test.go` | Connection tests | ~40 | - |
| `tcp_tls_test.go` | TLS tests | ~25 | - |
| `tcp_callbacks_test.go` | Callback tests | ~20 | - |
| `tcp_edge_test.go` | Edge cases | ~34 | - |
| **udp/** | UDP client tests | 73 | 73.7% |
| **unix/** | Unix client tests | 67 | 76.3% |
| **unixgram/** | Unixgram client tests | 65 | 76.8% |

### Server Package

| File | Purpose | Specs | Coverage |
|------|---------|-------|----------|
| **tcp/** | TCP server tests | 117 | 84.6% |
| `tcp_suite_test.go` | Suite initialization | - | - |
| `tcp_test.go` | Server lifecycle | ~35 | - |
| `tcp_connections_test.go` | Connection handling | ~30 | - |
| `tcp_shutdown_test.go` | Graceful shutdown | ~25 | - |
| `tcp_tls_test.go` | TLS tests | ~20 | - |
| `tcp_example_test.go` | Example tests | 3 | - |
| **udp/** | UDP server tests | 18 | 72.0% |
| **unix/** | Unix server tests | 23 | 73.8% |
| **unixgram/** | Unixgram server tests | 20 | 71.2% |

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should accept multiple concurrent connections", func() {
    // Test implementation
})

It("should gracefully shutdown and drain connections", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should establish TCP connection with timeout", func() {
    // Arrange
    cli, err := tcp.New("localhost:8080")
    Expect(err).ToNot(HaveOccurred())
    defer cli.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Act
    err = cli.Connect(ctx)
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
    Expect(cli.IsConnected()).To(BeTrue())
})
```

**3. Use Appropriate Matchers**
```go
Expect(err).ToNot(HaveOccurred())
Expect(cli.IsConnected()).To(BeTrue())
Expect(srv.OpenConnections()).To(BeNumerically(">", 0))
Expect(data).To(Equal(expectedData))
Expect(srv.IsRunning()).To(BeFalse())
```

**4. Always Cleanup Resources**
```go
BeforeEach(func() {
    srv = tcp.New(nil, handler)
    srv.RegisterServer(":0") // Random port
})

AfterEach(func() {
    if srv != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        srv.Shutdown(ctx)
    }
})
```

**5. Test with Context**
```go
It("should respect context timeout", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    cli, _ := tcp.New("192.0.2.1:9999") // Non-routable IP
    err := cli.Connect(ctx)
    
    Expect(err).To(HaveOccurred())
    Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
})
```

**6. Test Edge Cases**
- Nil instances
- Invalid addresses
- Connection refused
- Timeouts
- Context cancellation
- Closed connections

### Test Template

```go
var _ = Describe("socket/component", func() {
    var (
        srv    socket.Server
        cli    socket.Client
        ctx    context.Context
        cancel context.CancelFunc
    )

    BeforeEach(func() {
        ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        
        handler := func(r socket.Reader, w socket.Writer) {
            defer r.Close()
            defer w.Close()
            io.Copy(w, r) // Echo
        }
        
        srv = tcp.New(nil, handler)
        srv.RegisterServer(":0") // Random port
        
        go srv.Listen(ctx)
        time.Sleep(100 * time.Millisecond) // Wait for server
    })

    AfterEach(func() {
        if srv != nil {
            shutdownCtx, shutdownCancel := context.WithTimeout(
                context.Background(), 
                5*time.Second,
            )
            defer shutdownCancel()
            srv.Shutdown(shutdownCtx)
        }
        
        if cli != nil {
            cli.Close()
        }
        
        cancel()
    })

    Context("When testing feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            cli, err := tcp.New("localhost:8080")
            Expect(err).ToNot(HaveOccurred())
            
            // Act
            err = cli.Connect(ctx)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(cli.IsConnected()).To(BeTrue())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Use random ports (`:0`) for servers
- ✅ Create clients/servers per test
- ❌ Don't rely on test execution order
- ❌ Don't share state between tests

**Resource Cleanup**
```go
// ✅ Good
AfterEach(func() {
    if cli != nil {
        cli.Close()
    }
    if srv != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        srv.Shutdown(ctx)
    }
})

// ❌ Bad
AfterEach(func() {
    // Missing cleanup - resources leak
})
```

**Timing Considerations**
```go
// ✅ Good: Use Eventually for async operations
Eventually(func() bool {
    return srv.IsRunning()
}).Should(BeTrue())

Eventually(func() int64 {
    return srv.OpenConnections()
}).Should(Equal(int64(0)))

// ❌ Bad: Sleep-based timing
time.Sleep(1 * time.Second)
Expect(srv.IsRunning()).To(BeTrue()) // May be flaky
```

**Assertions**
```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))
Expect(count).To(BeNumerically(">=", 1))

// ❌ Avoid: Generic boolean assertions
Expect(value == expected).To(BeTrue())
```

**Concurrency Testing**
```go
It("should handle concurrent clients", func() {
    var wg sync.WaitGroup
    clientCount := 10
    
    for i := 0; i < clientCount; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            defer GinkgoRecover() // Recover panics in goroutines
            
            cli, err := tcp.New("localhost:8080")
            Expect(err).ToNot(HaveOccurred())
            defer cli.Close()
            
            err = cli.Connect(ctx)
            Expect(err).ToNot(HaveOccurred())
            
            // Test operations...
        }(i)
    }
    
    wg.Wait()
    
    Eventually(func() int64 {
        return srv.OpenConnections()
    }).Should(Equal(int64(0)))
})
```

**Error Handling**
```go
// ✅ Good: Test both success and error paths
It("should handle connection errors", func() {
    cli, err := tcp.New("invalid:address:format")
    Expect(err).To(HaveOccurred())
})

It("should succeed with valid address", func() {
    cli, err := tcp.New("localhost:8080")
    Expect(err).ToNot(HaveOccurred())
})
```

**Performance**
- Keep tests fast (use timeouts)
- Use parallel execution when possible (`ginkgo -p`)
- Target: <5s per subpackage (excluding long-running tests)
- Use `Eventually()` with reasonable timeouts

---

## Troubleshooting

**Port Already in Use**
```bash
# Always use random ports in tests
srv.RegisterServer(":0") // Let OS assign port

# Or use unique ports per test
port := 10000 + GinkgoParallelProcess()
srv.RegisterServer(fmt.Sprintf(":%d", port))
```

**Leftover Processes**
```bash
# Find processes using sockets
lsof -i :8080
lsof /tmp/app.sock

# Kill processes
kill -9 <PID>

# Clean socket files
rm -f /tmp/*.sock
```

**Stale Coverage**
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Parallel Test Failures**
- Check for shared resources (ports, files)
- Use unique identifiers per test
- Synchronize with channels or Eventually

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing atomic operations
- Unsynchronized goroutines
- Callback race conditions

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: xcode-select --install

export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Increase timeout for slow tests
go test -timeout=10m ./...

# Identify hanging tests
ginkgo --timeout=30s
```

Check for:
- Goroutine leaks (missing `wg.Done()`, `defer cancel()`)
- Unclosed connections
- Blocking operations without timeout
- Deadlocks

**Debugging**
```bash
# Single test
ginkgo --focus="should establish connection"

# Specific file
ginkgo --focus-file=tcp_test.go

# Verbose output
ginkgo -v --trace

# Debug individual spec
It("should do something", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
    Expect(value).To(Equal(expected))
})
```

**Unix Socket Tests on Windows**
```bash
# Skip Unix socket tests on Windows
go test ./client/tcp/ ./client/udp/ ./server/tcp/ ./server/udp/

# Or use build tags in test files
//go:build linux || darwin
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.21', '1.22']
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}
      
      - name: Run tests (Linux/macOS)
        if: runner.os != 'Windows'
        run: go test -v ./...
      
      - name: Run tests (Windows - TCP/UDP only)
        if: runner.os == 'Windows'
        run: |
          go test -v ./client/tcp/ ./client/udp/
          go test -v ./server/tcp/ ./server/udp/
      
      - name: Race detection (Linux/macOS)
        if: runner.os != 'Windows'
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: go test -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detection..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "All checks passed!"
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥70%
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Platform-specific code tested
- [ ] Test duration reasonable (<10m total)
- [ ] No test flakiness (run multiple times)
- [ ] Cleanup resources properly

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
- [sync Package](https://pkg.go.dev/sync)
- [atomic Package](https://pkg.go.dev/sync/atomic)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

**Networking**
- [net Package](https://pkg.go.dev/net)
- [Unix Domain Sockets](https://man7.org/linux/man-pages/man7/unix.7.html)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Socket Package Contributors
