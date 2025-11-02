# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-178%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A5%2070%25-brightgreen)]()

Comprehensive testing documentation for the socket/server package, covering test execution, race detection, and quality assurance across all transport protocols.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Protocol-Specific Testing](#protocol-specific-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The socket/server package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing across four transport protocols.

**Test Suite Summary**
- Total Specs: 178
- Coverage: 70-85% across subpackages
- Race Detection: ✅ Zero data races
- Execution Time: ~34s (without race), ~39s (with race)

**Coverage by Subpackage**

| Subpackage | Specs | Coverage | Duration | Duration (race) | Transport |
|------------|-------|----------|----------|-----------------|-----------|
| `tcp` | 117 | 84.6% | ~28s | ~30s | Network, connection-oriented |
| `udp` | 18 | 72.0% | ~1.4s | ~2.5s | Network, connectionless |
| `unix` | 23 | 73.8% | ~2s | ~3s | IPC, connection-oriented |
| `unixgram` | 20 | 71.2% | ~2.4s | ~3.4s | IPC, connectionless |

**Coverage Areas**
- Server lifecycle (start, stop, shutdown)
- Connection handling (accept, read, write, close)
- Graceful shutdown and connection draining
- Concurrent client operations
- Callback registration and invocation
- Error conditions and edge cases
- Thread safety and race conditions

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
cd /sources/go/src/github.com/nabbar/golib/socket/server
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Specific subpackage
go test -v ./tcp/
go test -v ./udp/
go test -v ./unix/
go test -v ./unixgram/

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
# Standard test run (all subpackages)
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Specific subpackage with verbose
go test -v github.com/nabbar/golib/socket/server/tcp
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific subpackage
ginkgo ./tcp

# Pattern matching
ginkgo --focus="TCP"

# Parallel execution (caution with network tests)
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml

# Timeout for long-running tests
ginkgo --timeout=5m
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Specific subpackage
CGO_ENABLED=1 go test -race ./tcp/
```

**Validates**:
- Atomic operations (`atomic.Bool`, `atomic.Int64`)
- Mutex protection (callback registration)
- Goroutine synchronization (per-connection handlers)
- Connection counter thread safety
- Concurrent client access

**Expected Output**:
```bash
# ✅ Success
=== RUN   TestSocketServerTCP
Running Suite: Socket Server TCP Suite
...
Ran 117 of 117 Specs in 27.555 seconds
SUCCESS! -- 117 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestSocketServerTCP (27.56s)
PASS
ok  	github.com/nabbar/golib/socket/server/tcp	28.022s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected across all subpackages

### Performance & Profiling

```bash
# Benchmarks (if available)
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.out ./tcp/
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./tcp/
go tool pprof cpu.out
```

**Performance Expectations**

| Subpackage | Duration (no race) | Duration (race) | Notes |
|------------|-------------------|-----------------|-------|
| TCP | ~28s | ~30s | Many connection tests |
| UDP | ~1.4s | ~2.5s | Stateless, fast |
| Unix | ~2s | ~3s | IPC, fast |
| Unixgram | ~2.4s | ~3.4s | Fastest IPC |
| **Total** | **~34s** | **~39s** | Slightly slower with race (normal) |

---

## Test Coverage

**Target**: ≥70% statement coverage per subpackage

### Coverage Summary

```bash
# View coverage for all subpackages
go test -cover ./...

# Output:
ok  	github.com/nabbar/golib/socket/server/tcp        28.022s	coverage: 84.6% of statements
ok  	github.com/nabbar/golib/socket/server/udp         1.429s	coverage: 72.0% of statements
ok  	github.com/nabbar/golib/socket/server/unix        2.021s	coverage: 73.8% of statements
ok  	github.com/nabbar/golib/socket/server/unixgram    2.364s	coverage: 71.2% of statements
```

### Coverage By Category

**TCP Server (84.6%)**
- 117 specs covering:
  - Server creation and configuration
  - Connection lifecycle (accept, data transfer, close)
  - TLS/SSL encryption
  - Graceful shutdown with draining
  - Concurrent client handling
  - Half-close operations
  - Error handling and edge cases
  - Callback invocation

**UDP Server (72.0%)**
- 18 specs covering:
  - Server creation and registration
  - Datagram send/receive
  - Sender address tracking
  - Shutdown (no draining needed)
  - Stateless operation
  - Error handling

**Unix Server (73.8%)**
- 23 specs covering:
  - Socket file creation and permissions
  - Group ownership configuration
  - Connection handling
  - Half-close support
  - Graceful shutdown with draining
  - File cleanup on shutdown

**Unixgram Server (71.2%)**
- 20 specs covering:
  - Socket file creation
  - Datagram mode operation
  - File permissions
  - Stateless handling
  - Fast shutdown

### View Detailed Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Per-subpackage coverage
go test -coverprofile=tcp_coverage.out ./tcp/
go tool cover -html=tcp_coverage.out -o tcp_coverage.html
```

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("Socket Server TCP", func() {
    BeforeSuite(func() {
        // Global setup (test server initialization)
    })
    
    AfterSuite(func() {
        // Global cleanup
    })
    
    Context("Server Creation", func() {
        It("should create server with handler", func() {
            srv := tcp.New(nil, handler)
            Expect(srv).ToNot(BeNil())
        })
    })
    
    Context("Server Lifecycle", func() {
        var srv tcp.ServerTcp
        
        BeforeEach(func() {
            srv = tcp.New(nil, echoHandler)
            srv.RegisterServer(fmt.Sprintf(":%d", getPort()))
        })
        
        AfterEach(func() {
            srv.Close()
        })
        
        It("should start listening", func() {
            go srv.Listen(ctx)
            Eventually(srv.IsRunning).Should(BeTrue())
        })
        
        It("should accept connections", func() {
            go srv.Listen(ctx)
            // Connect and test...
        })
    })
})
```

---

## Thread Safety

Thread safety is critical for concurrent client handling.

### Concurrency Primitives

```go
// Atomic state flags
atomic.Bool          // Server running state
atomic.Int64         // Connection counter

// Mutex protection (minimal, mostly for callbacks)
sync.Mutex

// Goroutine lifecycle
context.Context      // Cancellation propagation
sync.WaitGroup       // Connection draining (TCP/Unix)
```

### Verified Components

| Component | Mechanism | Subpackages | Status |
|-----------|-----------|-------------|--------|
| Server state | `atomic.Bool` (run, gone) | All | ✅ Race-free |
| Connection counter | `atomic.Int64` | TCP, Unix | ✅ Race-free |
| Callback registration | Direct store | All | ✅ Safe |
| Per-connection handlers | Independent goroutines | TCP, Unix | ✅ Isolated |
| Datagram handling | Single handler | UDP, Unixgram | ✅ Race-free |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations (TCP has most)
CGO_ENABLED=1 go test -race -v ./tcp/

# Stress test (run multiple times)
for i in {1..10}; do 
    CGO_ENABLED=1 go test -race ./... || break
done
```

**Result**: Zero data races across all test runs

---

## Protocol-Specific Testing

### TCP Server Tests

**Connection-Oriented Features**
```go
// Multiple concurrent connections
It("should handle multiple clients", func() {
    go srv.Listen(ctx)
    
    for i := 0; i < 10; i++ {
        conn, _ := net.Dial("tcp", serverAddr)
        // Test concurrent access...
        defer conn.Close()
    }
})

// Graceful shutdown with draining
It("should drain connections on shutdown", func() {
    go srv.Listen(ctx)
    
    conn, _ := net.Dial("tcp", serverAddr)
    defer conn.Close()
    
    // Start shutdown
    go srv.Shutdown(shutdownCtx)
    
    // Connection should complete
    Eventually(srv.IsGone).Should(BeTrue())
})

// TLS encryption
It("should support TLS", func() {
    srv.SetTLS(true, tlsConfig)
    go srv.Listen(ctx)
    
    conn, _ := tls.Dial("tcp", serverAddr, clientTLSConfig)
    // Test encrypted connection...
})
```

### UDP Server Tests

**Connectionless Features**
```go
// Single datagram
It("should receive and respond to datagram", func() {
    go srv.Listen(ctx)
    
    conn, _ := net.Dial("udp", serverAddr)
    defer conn.Close()
    
    conn.Write([]byte("test"))
    
    buf := make([]byte, 1024)
    n, _ := conn.Read(buf)
    Expect(string(buf[:n])).To(Equal("test"))
})

// OpenConnections returns 1 (running) or 0 (stopped)
It("should report connection count", func() {
    Expect(srv.OpenConnections()).To(Equal(int64(0)))
    
    go srv.Listen(ctx)
    Eventually(srv.IsRunning).Should(BeTrue())
    
    Expect(srv.OpenConnections()).To(Equal(int64(1)))
})
```

### Unix Server Tests

**IPC and File Permissions**
```go
// Socket file creation
It("should create socket file", func() {
    socketPath := "/tmp/test.sock"
    srv.RegisterSocket(socketPath, 0600, -1)
    
    go srv.Listen(ctx)
    Eventually(srv.IsRunning).Should(BeTrue())
    
    // Verify file exists
    _, err := os.Stat(socketPath)
    Expect(err).ToNot(HaveOccurred())
})

// File permissions
It("should set correct permissions", func() {
    socketPath := "/tmp/test.sock"
    srv.RegisterSocket(socketPath, 0600, -1)
    
    go srv.Listen(ctx)
    
    info, _ := os.Stat(socketPath)
    Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
})

// Group ownership
It("should set group ownership", func() {
    socketPath := "/tmp/test.sock"
    gid := int32(1000)
    srv.RegisterSocket(socketPath, 0660, gid)
    
    go srv.Listen(ctx)
    
    // Verify GID (platform-specific)
})
```

### Unixgram Server Tests

**Datagram IPC**
```go
// Fast datagram send
It("should send datagram via Unix socket", func() {
    socketPath := "/tmp/test.sock"
    srv.RegisterSocket(socketPath, 0600, -1)
    
    go srv.Listen(ctx)
    Eventually(srv.IsRunning).Should(BeTrue())
    
    conn, _ := net.DialUnix("unixgram", nil, 
        &net.UnixAddr{Net: "unixgram", Name: socketPath})
    defer conn.Close()
    
    conn.Write([]byte("fast"))
    // Test response...
})
```

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should accept TCP connection and echo data", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should track connection count", func() {
    // Arrange
    srv := tcp.New(nil, echoHandler)
    srv.RegisterServer(":8080")
    
    // Act
    go srv.Listen(ctx)
    Eventually(srv.IsRunning).Should(BeTrue())
    
    conn, _ := net.Dial("tcp", ":8080")
    defer conn.Close()
    
    // Assert
    Eventually(func() int64 { 
        return srv.OpenConnections() 
    }).Should(BeNumerically(">", 0))
})
```

**3. Use Appropriate Matchers**
```go
Expect(srv).ToNot(BeNil())
Expect(err).ToNot(HaveOccurred())
Expect(srv.IsRunning()).To(BeTrue())
Expect(srv.OpenConnections()).To(Equal(int64(1)))
Eventually(srv.IsRunning).Should(BeTrue())
```

**4. Always Cleanup Resources**
```go
AfterEach(func() {
    if srv != nil {
        srv.Close()
    }
    if socketFile != "" {
        os.Remove(socketFile)
    }
})
```

**5. Test Edge Cases**
- Empty handler
- Missing address/socket path
- Double shutdown
- Connection during shutdown
- Invalid TLS configuration
- File permission errors

**6. Use Eventually for Async Operations**
```go
// ✅ Good: Wait for async operations
go srv.Listen(ctx)
Eventually(srv.IsRunning, 2*time.Second).Should(BeTrue())

// ❌ Bad: No wait
go srv.Listen(ctx)
Expect(srv.IsRunning()).To(BeTrue()) // May fail (race)
```

### Test Template

```go
var _ = Describe("socket/server/newfeature", func() {
    var (
        srv     tcp.ServerTcp
        ctx     context.Context
        cancel  context.CancelFunc
        port    int
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
        port = getFreePort()
        
        handler := func(r socket.Reader, w socket.Writer) {
            defer r.Close()
            defer w.Close()
            io.Copy(w, r)
        }
        
        srv = tcp.New(nil, handler)
        srv.RegisterServer(fmt.Sprintf(":%d", port))
    })

    AfterEach(func() {
        if srv != nil {
            srv.Close()
        }
        cancel()
    })

    Context("New Feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            go srv.Listen(ctx)
            Eventually(srv.IsRunning).Should(BeTrue())
            
            // Act
            conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
            Expect(err).ToNot(HaveOccurred())
            defer conn.Close()
            
            // Assert
            conn.Write([]byte("test"))
            buf := make([]byte, 4)
            n, _ := conn.Read(buf)
            Expect(string(buf[:n])).To(Equal("test"))
        })

        It("should handle error case", func() {
            err := srv.RegisterServer("invalid:address:format")
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Get free port for each test (`getFreePort()`)
- ✅ Create unique socket files (`/tmp/test_<random>.sock`)
- ❌ Don't rely on test execution order
- ❌ Don't use fixed ports (may conflict)

**Resource Management**
```go
// ✅ Good: Proper cleanup
AfterEach(func() {
    if srv != nil {
        srv.Close()
    }
    if conn != nil {
        conn.Close()
    }
    cancel()
})

// ❌ Bad: Leaked resources
AfterEach(func() {
    // Missing cleanup
})
```

**Async Operations**
```go
// ✅ Good: Wait for state
go srv.Listen(ctx)
Eventually(srv.IsRunning, 2*time.Second).Should(BeTrue())

conn, err := net.Dial("tcp", addr)
Expect(err).ToNot(HaveOccurred())

// ❌ Bad: No synchronization
go srv.Listen(ctx)
conn, _ := net.Dial("tcp", addr) // May fail
```

**Assertions**
```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(count).To(BeNumerically(">", 0))
Eventually(srv.IsRunning).Should(BeTrue())

// ❌ Bad: Generic comparisons
Expect(err == nil).To(BeTrue())
Expect(count > 0).To(BeTrue())
```

**Concurrency Testing**
```go
It("should handle concurrent clients", func() {
    go srv.Listen(ctx)
    Eventually(srv.IsRunning).Should(BeTrue())
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            conn, _ := net.Dial("tcp", addr)
            defer conn.Close()
            // Test...
        }(i)
    }
    wg.Wait()
})
```

**Performance**
- Keep tests fast (use small payloads)
- Don't test excessive connection counts (slow)
- Use timeouts for long operations
- Target: <1s per spec (except TCP suite)

**Platform-Specific Tests**
```go
// Unix sockets only on Linux
if runtime.GOOS != "linux" {
    Skip("Unix sockets require Linux")
}
```

---

## Troubleshooting

**Port Already in Use**
```bash
# Find process using port
lsof -i :8080
netstat -tulpn | grep 8080

# Kill process
kill -9 <PID>

# In tests: Always use getFreePort()
port := getFreePort()
srv.RegisterServer(fmt.Sprintf(":%d", port))
```

**Leftover Socket Files**
```bash
# Clean manually
rm -f /tmp/*.sock

# In tests: Always cleanup
AfterEach(func() {
    os.Remove(socketPath)
})
```

**Stale Test Cache**
```bash
go clean -testcache
go test -count=1 ./...
```

**Hanging Tests**
```bash
# Set timeout
go test -timeout=2m ./...

# Identify hanging test
ginkgo --timeout=10s -v

# Check for:
# - Missing srv.Close()
# - Goroutine leaks
# - Deadlocked connections
```

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected variable access
- Missing atomic operations
- Concurrent map access
- Callback invocation without protection

**Connection Refused**
```bash
# Ensure server is running
Eventually(srv.IsRunning).Should(BeTrue())

# Wait before connecting
time.Sleep(100 * time.Millisecond)

# Or use Eventually with Dial
Eventually(func() error {
    conn, err := net.Dial("tcp", addr)
    if err == nil {
        conn.Close()
    }
    return err
}).Should(Succeed())
```

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian:
sudo apt-get install build-essential

# macOS:
xcode-select --install

# Enable CGO
export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts (TCP suite is long)**
```bash
# Increase timeout for TCP tests
go test -timeout=5m ./tcp/

# Or with Ginkgo
ginkgo --timeout=5m ./tcp
```

**Debugging**
```bash
# Single test
ginkgo --focus="should accept connection"

# Specific file
ginkgo --focus-file=tcp_test.go

# Verbose output
ginkgo -v --trace

# Show server output
fmt.Fprintf(GinkgoWriter, "Server state: %v\n", srv.IsRunning())
```

---

## CI Integration

**GitHub Actions Example**
```yaml
name: Socket Server Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd socket/server
          go test -v ./...
      
      - name: Race detection
        run: |
          cd socket/server
          CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: |
          cd socket/server
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

**Pre-commit Hook**
```bash
#!/bin/bash
cd socket/server

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "All checks passed!"
```

**Makefile**
```makefile
.PHONY: test test-race test-cover

test:
	cd socket/server && go test -v ./...

test-race:
	cd socket/server && CGO_ENABLED=1 go test -race ./...

test-cover:
	cd socket/server && go test -coverprofile=coverage.out ./...
	cd socket/server && go tool cover -html=coverage.out -o coverage.html

test-all: test test-race test-cover
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥70% per subpackage
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Test duration reasonable (~34s without race, ~39s with race)
- [ ] No leftover test files or sockets
- [ ] Documentation updated

**Per-Subpackage Checklist**

TCP:
- [ ] Connection handling tested
- [ ] TLS configuration tested
- [ ] Graceful shutdown with draining
- [ ] Concurrent clients tested

UDP:
- [ ] Datagram send/receive tested
- [ ] Stateless operation verified
- [ ] Sender address tracking tested

Unix:
- [ ] Socket file creation tested
- [ ] File permissions tested
- [ ] Group ownership tested
- [ ] File cleanup on shutdown

Unixgram:
- [ ] Datagram IPC tested
- [ ] File permissions tested
- [ ] Fast shutdown tested

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

**Networking**
- [net Package](https://pkg.go.dev/net)
- [Go Network Programming](https://go.dev/blog/network-programming)
- [TCP/IP Guide](https://www.ietf.org/rfc/rfc793.txt)
- [Unix Domain Sockets](https://man7.org/linux/man-pages/man7/unix.7.html)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows (Unix sockets: Linux only)  
**Maintained By**: Socket Server Package Contributors
