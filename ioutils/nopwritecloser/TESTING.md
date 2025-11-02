# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-54%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Comprehensive testing documentation for the nopwritecloser package, covering basic functionality, edge cases, and concurrency testing.

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

The nopwritecloser package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with 100% code coverage.

**Test Suite Statistics**
- Total Specs: 54
- Passed: 54
- Skipped: 0
- Coverage: 100.0%
- Execution Time: ~206ms
- Success Rate: 100%

**Coverage Areas**
- Basic functionality (New, Write, Close)
- Edge cases (nil writers, empty writes, multiple close calls)
- Concurrency (concurrent writes, concurrent closes)
- Integration scenarios (real-world usage patterns)

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
# coverage: 100.0% of statements
```

### Advanced Options

```bash
# Run with race detector
go test -race .

# Focus on specific tests
ginkgo --focus="Basic Functionality"

# Verbose with trace
ginkgo -v --trace
```

---

## Test Coverage

### Coverage Metrics

**Overall Coverage: 100.0%**

All code paths are tested, including edge cases and concurrent scenarios.

### Coverage by Component

| Component | File | Coverage | Functions | Notes |
|-----------|------|----------|-----------|-------|
| Public API | `interface.go` | 100.0% | New | Wrapper creation |
| Implementation | `model.go` | 100.0% | Write, Close | Delegation + no-op |
| **Total** | **2 files** | **100.0%** | **3/3** | **Complete** |

### Coverage by Source File

```
Function Coverage Report:
interface.go:58:        New              100.0%
model.go:36:            Write            100.0%
model.go:42:            Close            100.0%
total:                  (statements)     100.0%
```

### Test Categories

**Basic Functionality (15 specs)** - `basic_test.go`
- New() creates valid wrapper
- Write() delegates correctly
- Close() returns nil
- Multiple operations
- Interface compliance

**Edge Cases (18 specs)** - `edge_cases_test.go`
- Nil writer handling
- Empty writes
- Large writes
- Multiple close calls
- Writes after close
- Zero-length buffers

**Concurrency (12 specs)** - `concurrency_test.go`
- Concurrent writes
- Concurrent closes
- Mixed concurrent operations
- Race condition detection (100-1000 goroutines)

**Integration (9 specs)** - `integration_test.go`
- os.Stdout wrapping
- bytes.Buffer usage
- JSON encoding
- Logger integration
- Real-world patterns

---

## Test Structure

### Test File Organization

| File | Purpose | Specs | Description |
|------|---------|-------|-------------|
| `nopwritecloser_suite_test.go` | Test suite entry | - | Ginkgo registration |
| `basic_test.go` | Core functionality | 15 | New, Write, Close operations |
| `edge_cases_test.go` | Boundary conditions | 18 | Nil, empty, large, multiple calls |
| `concurrency_test.go` | Thread safety | 12 | Concurrent operations, races |
| `integration_test.go` | Real-world use | 9 | Stdout, JSON, logging patterns |

### Test Hierarchy

```
Describe("nopwritecloser", func() {
    Context("Basic Functionality", func() {
        It("should create wrapper", ...)
        It("should delegate writes", ...)
        It("should return nil on close", ...)
    })
    
    Context("Edge Cases", func() {
        It("should handle nil writer", ...)
        It("should handle empty writes", ...)
    })
    
    Context("Concurrency", func() {
        It("should handle concurrent writes", ...)
    })
})
```

---

## Writing Tests

### Test Development Guidelines

**1. Follow BDD Structure**

```go
var _ = Describe("nopwritecloser", func() {
    var buf *bytes.Buffer
    var wc io.WriteCloser
    
    BeforeEach(func() {
        buf = &bytes.Buffer{}
        wc = New(buf)
    })
    
    It("should write data", func() {
        n, err := wc.Write([]byte("test"))
        
        Expect(err).ToNot(HaveOccurred())
        Expect(n).To(Equal(4))
        Expect(buf.String()).To(Equal("test"))
    })
})
```

**2. Test Edge Cases**

```go
It("should handle nil writer gracefully", func() {
    wc := New(nil)
    
    // Should not panic
    _, err := wc.Write([]byte("data"))
    Expect(err).To(HaveOccurred())
})
```

**3. Test Concurrency**

```go
It("should handle concurrent writes", func() {
    var buf bytes.Buffer
    wc := New(&buf)
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            wc.Write([]byte(fmt.Sprintf("%d", n)))
        }(i)
    }
    wg.Wait()
    
    // All writes should complete
    Expect(buf.Len()).To(BeNumerically(">", 0))
})
```

**4. Test Close Behavior**

```go
It("should allow writes after close", func() {
    wc := New(&buf)
    
    wc.Write([]byte("before"))
    wc.Close()
    wc.Write([]byte("after"))
    
    Expect(buf.String()).To(Equal("beforeafter"))
})
```

---

## Best Practices

### 1. Always Test Delegation

```go
It("should delegate writes to underlying writer", func() {
    mock := &mockWriter{}
    wc := New(mock)
    
    wc.Write([]byte("test"))
    
    Expect(mock.written).To(Equal([]byte("test")))
})
```

### 2. Test Multiple Close Calls

```go
It("should handle multiple close calls", func() {
    wc := New(&buf)
    
    Expect(wc.Close()).To(Succeed())
    Expect(wc.Close()).To(Succeed())
    Expect(wc.Close()).To(Succeed())
})
```

### 3. Don't Share Wrappers Between Tests

```go
// ❌ Bad - Shared state
var sharedWC io.WriteCloser

BeforeEach(func() {
    sharedWC = New(&buf)
})

// ✅ Good - Fresh state per test
It("should work", func() {
    wc := New(&bytes.Buffer{})
})
```

### 4. Test Real-World Scenarios

```go
It("should work with JSON encoder", func() {
    var buf bytes.Buffer
    wc := New(&buf)
    
    encoder := json.NewEncoder(wc)
    err := encoder.Encode(map[string]string{"key": "value"})
    wc.Close()
    
    Expect(err).ToNot(HaveOccurred())
    Expect(buf.String()).To(ContainSubstring("key"))
})
```

---

## Troubleshooting

**Problem: Tests fail with nil pointer**

```bash
# Check if writer is properly initialized
BeforeEach(func() {
    buf = &bytes.Buffer{}  // Don't forget the &
    wc = New(buf)
})
```

**Problem: Race conditions detected**

```bash
# Run with race detector
go test -race .

# Fix by ensuring underlying writer is thread-safe
# bytes.Buffer is NOT thread-safe for concurrent writes
```

**Problem: Concurrency tests fail intermittently**

```go
// Use WaitGroup to ensure all goroutines complete
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Test code
    }()
}
wg.Wait()
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
          cd ioutils/nopwritecloser
          go test -v -cover .
      
      - name: Run race detector
        run: |
          cd ioutils/nopwritecloser
          go test -race .
      
      - name: Check coverage
        run: |
          cd ioutils/nopwritecloser
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

**Standard Library**
- [io Package](https://pkg.go.dev/io)
- [io.NopCloser](https://pkg.go.dev/io#NopCloser)

**Related Documentation**
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/nopwritecloser)

---

**Version**: Go 1.19+ on Linux, macOS, Windows  
**Test Execution Time**: ~206ms  
**Test Success Rate**: 100% (54/54 specs passed)  
**Maintained By**: nopwritecloser Package Contributors
