# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-100%2B%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%3E95%25-brightgreen)]()

Comprehensive testing documentation for the atomic package, covering concurrency validation, race detection, and performance testing.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Concurrency Testing](#concurrency-testing)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The atomic package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for testing thread-safe concurrent operations.

**Test Suite**
- Total Specs: 100+
- Coverage: >95%
- Race Detection: ✅ Zero data races
- Execution Time: ~2s (without race), ~5s (with race)

**Coverage Areas**
- Atomic value operations (Store, Load, Swap, CompareAndSwap)
- Concurrent map operations (Map, MapTyped)
- Type casting utilities
- High-contention scenarios
- Edge cases (nil, zero values)

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (critical!)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -cover -race
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Focused and pending specs
- Rich CLI with filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Custom matcher support
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

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=value_test.go

# Pattern matching
ginkgo --focus="CompareAndSwap"

# Parallel execution
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for concurrency testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Stress test (multiple runs)
go test -race -count=100 ./...
go test -race -count=1000 -timeout=10m
```

**Validates**:
- Atomic operations correctness
- Map concurrent access
- Lock-free guarantees
- Memory ordering

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/atomic	5.234s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected

### Performance Testing

```bash
# Benchmarks
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.out -bench=.
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out -bench=.
go tool pprof cpu.out
```

**Performance Expectations**

| Test Category | Duration | Notes |
|---------------|----------|-------|
| Full Suite | ~2s | Without race |
| With `-race` | ~5s | 2.5x slower (normal) |
| Individual Spec | <50ms | Most tests |
| High Contention | 100-200ms | Expected |

---

## Test Coverage

**Target**: >95% statement coverage

### Coverage By Component

| Component | File | Specs | Coverage |
|-----------|------|-------|----------|
| Atomic Values | `value_test.go` | 30+ | 100% |
| Concurrent Maps | `map_test.go` | 40+ | 100% |
| Type Casting | `cast_test.go` | 20+ | 100% |
| Concurrent Ops | `atomic_test.go` | 10+ | 100% |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Test File Organization

| File | Purpose | Specs |
|------|---------|-------|
| `atomic_suite_test.go` | Suite initialization | 1 |
| `value_test.go` | Atomic value tests | 30+ |
| `map_test.go` | Map operations | 40+ |
| `cast_test.go` | Type casting | 20+ |
| `atomic_test.go` | Concurrent patterns | 10+ |

---

## Concurrency Testing

Thread safety is the primary focus of this package.

### Lock-Free Operations

```go
// Atomic values are lock-free
var val atomic.Value[int]
val.Store(42)        // No locks
v := val.Load()      // No locks
val.CompareAndSwap(42, 100)  // Lock-free CAS
```

### High-Contention Validation

```bash
# Stress test with 100 goroutines
go test -race -count=100 ./...

# Extended stress test
go test -race -count=1000 -timeout=10m ./...
```

**Test Patterns**:
- Multiple readers, single writer (MRSW)
- Multiple readers, multiple writers (MRMW)
- High contention scenarios (100+ goroutines)
- Race detector validation

### Test Example

```go
It("should handle concurrent access", func() {
    val := atomic.NewValue[int]()
    val.Store(0)
    
    var wg sync.WaitGroup
    
    // 50 concurrent writers
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(v int) {
            defer wg.Done()
            val.Store(v)
        }(i)
    }
    
    // 50 concurrent readers
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = val.Load()
        }()
    }
    
    wg.Wait()
})
```

**Status**: Zero data races detected across all tests

---

## Writing Tests

### Guidelines

**1. Test Concurrent Access**
```go
It("should be thread-safe", func() {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Concurrent operations
        }()
    }
    wg.Wait()
})
```

**2. Use Proper Matchers**
```go
Expect(val).To(Equal(expected))
Expect(ok).To(BeTrue())
Expect(err).ToNot(HaveOccurred())
```

**3. Test Edge Cases** - Nil values, zero values, type mismatches

**4. Always Use Race Detector** - Run with `-race` during development

### Test Template

```go
var _ = Describe("New Feature", func() {
    It("should work correctly", func() {
        result := NewFeature()
        Expect(result).To(BeValid())
    })
    
    It("should be thread-safe", func() {
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                NewFeature()
            }()
        }
        wg.Wait()
    })
})
```

---

## Best Practices

**Test Independence**
- Each test should be independent
- Use `BeforeEach`/`AfterEach` for setup/cleanup
- Avoid global mutable state

**Concurrency Testing**
```go
// ✅ Good: Proper synchronization
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Operations
    }()
}
wg.Wait()

// ❌ Bad: No synchronization
for i := 0; i < 100; i++ {
    go func() {
        // Operations (may not complete)
    }()
}
```

**Type Safety**
```go
// ✅ Good: Type-safe
cache := atomic.NewMapTyped[string, int]()
val, ok := cache.Load("key") // val is int

// ❌ Avoid: Type assertions in tests
cache := atomic.NewMapAny[string]()
val, _ := cache.Load("key")
num := val.(int) // May panic
```

**Race Detection**
- Always run with `-race` during development
- Test with multiple runs (`-count=100`)
- Verify no data races before commit

---

## Troubleshooting

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: brew install gcc

export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Identify hanging tests
go test -timeout=10s ./...
```

Check for:
- Goroutine leaks (missing `wg.Done()`)
- Unclosed channels
- Deadlocks

**Debugging**
```bash
# Single test
ginkgo --focus="CompareAndSwap"

# Specific file
ginkgo --focus-file=value_test.go

# Verbose output
ginkgo -v --trace
```

---

## CI Integration

**GitHub Actions Example**
```yaml
name: Tests
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
        run: go test -v ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: go test -coverprofile=coverage.out ./...
```

**Pre-commit Hook**
```bash
#!/bin/bash
CGO_ENABLED=1 go test -race ./... || exit 1
go test -cover ./... | grep -E "coverage:" || exit 1
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: >95%
- [ ] Concurrent access tested
- [ ] Edge cases covered
- [ ] Benchmarks added for performance-critical code

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync/atomic Package](https://pkg.go.dev/sync/atomic)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Atomic Package Contributors
