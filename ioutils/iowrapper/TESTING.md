# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-114%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Comprehensive testing documentation for the iowrapper package, covering test execution, concurrency testing, and integration scenarios.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Structure](#test-structure)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The iowrapper package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with 100% code coverage across all scenarios.

**Test Suite Statistics**
- Total Specs: 114
- Passed: 114
- Skipped: 0
- Coverage: 100.0%
- Execution Time: ~47ms
- Success Rate: 100%

**Coverage Areas**
- Basic I/O operations (Read, Write, Seek, Close)
- Custom function registration and execution
- Edge cases and boundary conditions
- Error handling and propagation
- Thread safety and concurrency
- Real-world integration scenarios
- Performance benchmarking

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Using Ginkgo CLI
ginkgo -v -cover
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Rich CLI with filtering and reporting
- Parallel test execution support

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax (`Expect(...).To(...)`)
- Rich matcher library
- Detailed failure messages
- Asynchronous assertion support

---

## Test Files Organization

| File | Purpose | Specs | Description |
|------|---------|-------|-------------|
| `iowrapper_suite_test.go` | Test suite entry | - | Ginkgo test registration |
| `basic_test.go` | Basic operations | 20 | Default I/O operations, wrapper creation |
| `custom_test.go` | Custom functions | 24 | Custom Read/Write/Seek/Close functions |
| `edge_cases_test.go` | Edge cases | 18 | Boundary conditions, nil handling |
| `errors_test.go` | Error handling | 19 | Error propagation, nil returns |
| `concurrency_test.go` | Thread safety | 17 | Concurrent operations, race conditions |
| `integration_test.go` | Real-world use | 8 | Logging, transformation, checksumming |
| `benchmark_test.go` | Performance | 8 | Operation overhead, memory efficiency |

## Running Tests

### Basic Commands

```bash
# Standard go test
go test .
go test -v .                    # Verbose output
go test -cover .                # With coverage

# Ginkgo CLI (recommended)
ginkgo                          # Run all tests
ginkgo -v                       # Verbose output
ginkgo -cover                   # With coverage
```

### Coverage Reports

```bash
# Generate coverage profile
go test -coverprofile=coverage.out .

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Expected output:
# coverage: 100.0% of statements
```

### Advanced Options

```bash
# Focus on specific tests
ginkgo --focus="Custom Read"
ginkgo --focus="Concurrent"

# Run with race detector
CGO_ENABLED=1 go test -race .

# Run benchmarks
go test -bench=. -benchmem

# Verbose with trace
ginkgo -v --trace
```

## Test Coverage

### Coverage Metrics

**Overall Coverage: 100.0%**

All functions and code paths are tested, including edge cases and error conditions.

### Coverage by Component

| Component | File | Coverage | Functions | Notes |
|-----------|------|----------|-----------|-------|
| Public API | `interface.go` | 100.0% | New, IOWrapper | Fully tested |
| Implementation | `model.go` | 100.0% | All methods | All paths covered |
| **Total** | **2 files** | **100.0%** | **13/13** | **Complete** |

### Coverage by Source File

```
Function Coverage Report:
interface.go:124:       New              100.0%
model.go:44:            SetRead          100.0%
model.go:52:            SetWrite         100.0%
model.go:60:            SetSeek          100.0%
model.go:68:            SetClose         100.0%
model.go:76:            Read             100.0%
model.go:95:            Write            100.0%
model.go:114:           Seek             100.0%
model.go:122:           Close            100.0%
model.go:132:           fakeRead         100.0%
model.go:146:           fakeWrite        100.0%
model.go:160:           fakeSeek         100.0%
model.go:170:           fakeClose        100.0%
total:                  (statements)     100.0%
```

### Test Categories

**Basic Operations (20 specs)** - `basic_test.go`
- Wrapper creation from various types
- Default Read/Write/Seek/Close delegation
- Empty object handling
- Interface compliance

**Custom Functions (24 specs)** - `custom_test.go`
- Custom function registration
- Function replacement
- Reset to default (nil)
- Multiple custom functions

**Edge Cases (18 specs)** - `edge_cases_test.go`
- Non-interface objects
- Nil and empty values
- Buffer size boundaries
- Rapid function replacement

**Error Handling (19 specs)** - `errors_test.go`
- Error propagation
- Nil return handling
- io.ErrUnexpectedEOF cases
- Partial reads/writes

**Concurrency (17 specs)** - `concurrency_test.go`
- Concurrent reads/writes
- Concurrent function updates
- Race condition testing
- Thread safety verification

**Integration (8 specs)** - `integration_test.go`
- Logging wrapper
- Data transformation (ROT13, uppercase)
- Checksumming (MD5, SHA256)
- Wrapper chaining

**Benchmarks (8 specs)** - `benchmark_test.go`
- Creation overhead (~5.7ms/10k ops)
- Read/write performance (~0-100ns/op)
- Function update (~100ns/op)
- Memory allocation (0 per I/O op)

### Viewing Coverage

```bash
# Generate and view detailed coverage
go test -coverprofile=coverage.out .
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser
```

## Writing Tests

### Test Development Guidelines

**1. Follow BDD Structure**

```go
var _ = Describe("Feature", func() {
    Context("When condition", func() {
        It("should behave correctly", func() {
            // Arrange
            wrapper := New(object)
            
            // Act
            result, err := wrapper.Read(buffer)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expected))
        })
    })
})
```

**2. Test Custom Functions**

```go
It("should use custom read function", func() {
    wrapper := New(reader)
    called := false
    
    wrapper.SetRead(func(p []byte) []byte {
        called = true
        return []byte("custom")
    })
    
    buffer := make([]byte, 100)
    wrapper.Read(buffer)
    
    Expect(called).To(BeTrue())
})
```

**3. Test Thread Safety**

```go
It("should handle concurrent operations", func() {
    wrapper := New(buffer)
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            wrapper.Read(make([]byte, 10))
        }()
    }
    
    wg.Wait() // Should not panic or race
})
```

**4. Test Error Conditions**

```go
It("should return io.ErrUnexpectedEOF for nil return", func() {
    wrapper := New(reader)
    wrapper.SetRead(func(p []byte) []byte {
        return nil // Simulate error
    })
    
    n, err := wrapper.Read(make([]byte, 10))
    Expect(err).To(Equal(io.ErrUnexpectedEOF))
    Expect(n).To(Equal(0))
})
```

---

## Best Practices

### 1. Keep Tests Fast

```go
// ✅ Good - Fast in-memory operation
It("should read data", func() {
    wrapper := New(bytes.NewBufferString("data"))
    wrapper.Read(buffer)
})

// ❌ Bad - Slow operation
It("should wait", func() {
    time.Sleep(time.Second) // Don't do this
})
```

### 2. Use Atomic Operations for Concurrency

```go
// ✅ Good - Thread-safe
var counter atomic.Int64
wrapper.SetRead(func(p []byte) []byte {
    counter.Add(1) // Thread-safe
    return data
})

// ❌ Bad - Race condition
var counter int
wrapper.SetRead(func(p []byte) []byte {
    counter++ // Not thread-safe!
    return data
})
```

### 3. Test All Return Paths

```go
It("should handle all return cases", func() {
    // Test nil return
    wrapper.SetRead(func(p []byte) []byte { return nil })
    _, err := wrapper.Read(buf)
    Expect(err).To(Equal(io.ErrUnexpectedEOF))
    
    // Test empty return
    wrapper.SetRead(func(p []byte) []byte { return []byte{} })
    n, _ := wrapper.Read(buf)
    Expect(n).To(Equal(0))
    
    // Test data return
    wrapper.SetRead(func(p []byte) []byte { return []byte("data") })
    n, _ = wrapper.Read(buf)
    Expect(n).To(BeNumerically(">", 0))
})
```

### 4. Don't Share State Between Tests

```go
// ❌ Bad - Shared state
var sharedWrapper IOWrapper

BeforeEach(func() {
    sharedWrapper = New(buffer) // Risk of interference
})

// ✅ Good - Fresh state per test
var _ = Describe("Feature", func() {
    It("should work", func() {
        wrapper := New(buffer) // Local to this test
    })
})
```

---

## Troubleshooting

**Problem: Race condition detected**

```bash
# Run with race detector to identify
go test -race .

# Fix by using atomic operations
var counter atomic.Int64 // Instead of plain int
```

**Problem: Tests timeout**

Check for infinite loops in custom functions:
```go
// ❌ Bad - Never returns nil
wrapper.SetRead(func(p []byte) []byte {
    return []byte("data") // Infinite loop!
})

// ✅ Good - Returns nil on EOF
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    if n == 0 {
        return nil // EOF
    }
    return p[:n]
})
```

**Problem: Coverage gaps**

```bash
# Identify uncovered lines
go tool cover -html=coverage.out

# Focus on uncovered code paths
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Test

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
          cd ioutils/iowrapper
          go test -v -cover -race .
      
      - name: Check coverage
        run: |
          cd ioutils/iowrapper
          go test -coverprofile=coverage.out .
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | grep -E '^100'
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)
- [Go Coverage](https://go.dev/blog/cover)

**Standard Library**
- [io Package](https://pkg.go.dev/io)

**Related Documentation**
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/iowrapper)

---

**Version**: Go 1.19+ on Linux, macOS, Windows  
**Test Execution Time**: ~47ms  
**Test Success Rate**: 100% (114/114 specs passed)  
**Maintained By**: iowrapper Package Contributors
