# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-29%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-80.2%25-green)]()

Comprehensive testing documentation for the mapCloser package, covering test execution, concurrency testing, and context-aware cleanup scenarios.

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

The mapCloser package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with 80.2% code coverage.

**Test Suite Statistics**
- Total Specs: 29
- Passed: 29
- Skipped: 0
- Coverage: 80.2%
- Execution Time: ~5ms
- Success Rate: 100%

**Coverage Areas**
- Basic operations (Add, Get, Len, Close, Clean)
- Context-aware automatic cleanup
- Thread safety and concurrent operations
- Error handling and aggregation
- Cloning and resource isolation
- Edge cases (nil closers, double close, etc.)

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

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax (`Expect(...).To(...)`)
- Rich matcher library
- Detailed failure messages
- Asynchronous assertions (`Eventually`)

---

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
# coverage: 80.2% of statements
```

### Advanced Options

```bash
# Run with race detector
CGO_ENABLED=1 go test -race .

# Focus on specific tests
ginkgo --focus="Context Cancellation"

# Verbose with trace
ginkgo -v --trace
```

## Test Coverage

### Coverage Metrics

**Overall Coverage: 80.2%**

All critical paths are tested with comprehensive coverage of concurrent operations and error conditions.

### Coverage by Component

| Component | File | Coverage | Functions | Notes |
|-----------|------|----------|-----------|-------|
| Public API | `interface.go` | 83.3% | New | Context monitoring goroutine |
| Implementation | `model.go` | 80.0% | All methods | Core functionality |
| **Total** | **2 files** | **80.2%** | **10/10** | **Good coverage** |

### Coverage by Source File

```
Function Coverage Report:
interface.go:113:       New              83.3%
model.go:50:            idx              100.0%
model.go:55:            idxInc           100.0%
model.go:60:            Add              75.0%
model.go:74:            Get              80.0%
model.go:99:            Len              75.0%
model.go:112:           Len64            0.0%  (internal, not critical)
model.go:116:           Clean            62.5%
model.go:127:           Clone            81.8%
model.go:150:           Close            89.5%
total:                  (statements)     80.2%
```

### Test Categories

**Basic Operations (10 specs)** - Core functionality
- New() with various contexts
- Add() single and multiple closers
- Get() retrieval and filtering
- Len() counter tracking
- Clean() resource removal
- Close() cleanup and error aggregation

**Clone Operations (3 specs)** - Independence testing
- Clone() creates independent copy
- Modifications don't affect original
- Independent close operations

**Error Handling (5 specs)** - Error aggregation
- Single closer errors
- Multiple closer errors
- Error message formatting
- Nil closer handling

**Context Integration (4 specs)** - Context-aware behavior
- Context cancellation triggers cleanup
- Timeout-based cleanup
- Manual close with active context

**Concurrency (5 specs)** - Thread safety
- Concurrent Add() calls
- Concurrent Get() calls
- Mixed concurrent operations
- Race condition detection (50-200 goroutines)

**Edge Cases (2 specs)** - Boundary conditions
- Nil operations
- Double close
- Operations after close

### Uncovered Areas

Some scenarios are difficult to test without internal mocking:
- Context initialization failures in New()
- Internal storage corruption
- Extreme concurrency edge cases (>1000 goroutines)

---

## Writing Tests

### Test Development Guidelines

**1. Follow BDD Structure**

```go
var _ = Describe("mapCloser", func() {
    var ctx context.Context
    var cancel context.CancelFunc
    
    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
    })
    
    AfterEach(func() {
        cancel()
    })
    
    It("should add and retrieve closers", func() {
        closer := New(ctx)
        defer closer.Close()
        
        mockCloser := &mockCloser{}
        closer.Add(mockCloser)
        
        Expect(closer.Len()).To(Equal(1))
        Expect(closer.Get()).To(ContainElement(mockCloser))
    })
})
```

**2. Test Context-Aware Behavior**

```go
It("should close on context cancellation", func() {
    ctx, cancel := context.WithCancel(context.Background())
    closer := New(ctx)
    
    mock := &mockCloser{}
    closer.Add(mock)
    
    cancel() // Trigger context cancellation
    
    Eventually(func() bool {
        return mock.closed
    }, "200ms").Should(BeTrue())
})
```

**3. Test Concurrency**

```go
It("should handle concurrent operations", func() {
    closer := New(ctx)
    defer closer.Close()
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            closer.Add(&mockCloser{})
        }()
    }
    wg.Wait()
    
    Expect(closer.Len()).To(Equal(100))
})
```

---

## Best Practices

### 1. Always Defer Close

```go
closer := mapCloser.New(ctx)
defer closer.Close() // Ensures cleanup even on test failure
```

### 2. Use Eventually for Async Operations

```go
// ✅ Good - Handles async context cancellation
Eventually(func() bool {
    return resourceClosed
}, "200ms").Should(BeTrue())

// ❌ Bad - Timing issues
time.Sleep(100 * time.Millisecond)
Expect(resourceClosed).To(BeTrue())
```

### 3. Test Error Aggregation

```go
It("should aggregate multiple errors", func() {
    closer := New(ctx)
    defer closer.Close()
    
    closer.Add(
        newErrorCloser(errors.New("error1")),
        newErrorCloser(errors.New("error2")),
    )
    
    err := closer.Close()
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("error1"))
    Expect(err.Error()).To(ContainSubstring("error2"))
})
```

### 4. Don't Share Closers Between Tests

```go
// ❌ Bad - Shared state
var sharedCloser Closer

BeforeEach(func() {
    sharedCloser = New(ctx) // Risk of interference
})

// ✅ Good - Fresh state per test
It("should work", func() {
    closer := New(ctx) // Local to this test
    defer closer.Close()
})
```

---

## Troubleshooting

**Problem: Tests timeout with context**

```bash
# Increase timeout for context monitoring (100ms poll)
go test -timeout=30s .
```

**Problem: Race conditions detected**

```bash
# Run with race detector to identify
CGO_ENABLED=1 go test -race .

# Fix by using atomic operations
var counter atomic.Int64 // Instead of plain int
```

**Problem: Flaky context cancellation tests**

```go
// Use Eventually with adequate timeout
Eventually(func() bool {
    return closer.IsClosed()  
}, "500ms", "50ms").Should(BeTrue())
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
          cd ioutils/mapCloser
          go test -v -cover .
      
      - name: Run race detector
        run: |
          cd ioutils/mapCloser
          CGO_ENABLED=1 go test -race -timeout=60s .
      
      - name: Check coverage
        run: |
          cd ioutils/mapCloser
          go test -coverprofile=coverage.out .
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | grep -E '^[8-9][0-9]'
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

**Context Package**
- [Go Context](https://pkg.go.dev/context)

**Related Documentation**
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/mapCloser)

---

**Version**: Go 1.19+ on Linux, macOS, Windows  
**Test Execution Time**: ~5ms  
**Test Success Rate**: 100% (29/29 specs passed)  
**Maintained By**: mapCloser Package Contributors
