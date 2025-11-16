# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-352%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-95.4%25-brightgreen)]()

Comprehensive testing documentation for the size package, covering test execution, race detection, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The size package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Statistics**
- Total Specs: 352
- Coverage: 95.4% of statements
- Race Detection: ✅ Zero data races
- Execution Time: ~0.026s (without race), ~1.169s (with race)

**Coverage Areas**
- Parsing (strings, numbers, complex expressions)
- Arithmetic operations (overflow/underflow handling)
- Type conversions (all numeric types with overflow protection)
- Formatting (various precision levels and unit selection)
- Marshaling (JSON, YAML, TOML, CBOR, Text, Binary)
- Viper integration (decode hook with multiple types)
- Edge cases (maximum values, zero, negative inputs)

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

**Expected Output**:
```
=== RUN   TestSize
Running Suite: size Suite - /sources/go/src/github.com/nabbar/golib/size
========================================================================
Random Seed: 1763323372

Will run 352 of 352 specs
••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••

Ran 352 of 352 Specs in 0.015 seconds
SUCCESS! -- 352 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestSize (0.02s)
PASS
coverage: 95.4% of statements
ok  	github.com/nabbar/golib/size	0.026s	coverage: 95.4% of statements
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=parsing_test.go

# Pattern matching
ginkgo --focus="Parse"

# Parallel execution
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for validating thread-safe operations**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With timeout for long tests
CGO_ENABLED=1 go test -race -timeout=10m ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race
```

**Expected Output**:
```
=== RUN   TestSize
Running Suite: size Suite - /sources/go/src/github.com/nabbar/golib/size
========================================================================
Random Seed: 1763323372

Will run 352 of 352 specs
••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••

Ran 352 of 352 Specs in 0.122 seconds
SUCCESS! -- 352 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestSize (0.14s)
PASS
coverage: 95.4% of statements
ok  	github.com/nabbar/golib/size	1.169s	coverage: 95.4% of statements
```

**Status**: Zero data races detected across all 352 specs

### Performance & Profiling

```bash
# Benchmarks
go test -bench=. -benchmem ./...

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out
```

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite | ~0.026s | Without race |
| With `-race` | ~1.169s | 45x slower (normal for race detector) |
| Individual Spec | <1ms | Most tests |
| Complex Parsing | 1-2ms | Multiple unit components |

---

## Test Coverage

**Target**: ≥95% statement coverage  
**Achieved**: 95.4%

### Coverage By Category

| Category | Test File | Specs | Description |
|----------|-----------|-------|-------------|
| **Constants** | `constants_defaults_test.go` | ~40 | Size constants, unit relationships, defaults |
| **Parsing** | `parsing_test.go` | ~70 | String parsing, unit detection, edge cases |
| **Formatting** | `formatting_test.go` | ~60 | String formatting, precision, unit selection |
| **Arithmetic** | `arithmetic_operations_test.go` | ~50 | Math operations, overflow/underflow protection |
| **Conversions** | `type_conversions_test.go` | ~60 | Type conversions with overflow detection |
| **Encoding** | `encoding_marshalling_test.go` | ~50 | JSON, YAML, TOML, CBOR, Text, Binary |
| **Viper** | `viper_decoder_test.go` | ~22 | Viper decode hook, multiple input types |

### Detailed Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal (by function)
go tool cover -func=coverage.out

# View HTML report (by line)
go tool cover -html=coverage.out -o coverage.html
```

**Coverage by File**:
```
github.com/nabbar/golib/size/arithmetic.go:39:     Mul           100.0%
github.com/nabbar/golib/size/arithmetic.go:51:     MulErr        100.0%
github.com/nabbar/golib/size/arithmetic.go:65:     Div           100.0%
github.com/nabbar/golib/size/arithmetic.go:75:     DivErr        100.0%
github.com/nabbar/golib/size/arithmetic.go:88:     Add           100.0%
github.com/nabbar/golib/size/arithmetic.go:100:    AddErr        100.0%
github.com/nabbar/golib/size/arithmetic.go:114:    Sub           100.0%
github.com/nabbar/golib/size/arithmetic.go:125:    SubErr        100.0%
github.com/nabbar/golib/size/encode.go:49:         MarshalJSON   100.0%
github.com/nabbar/golib/size/encode.go:63:         UnmarshalJSON 100.0%
github.com/nabbar/golib/size/encode.go:77:         MarshalYAML   100.0%
github.com/nabbar/golib/size/encode.go:91:         UnmarshalYAML 100.0%
github.com/nabbar/golib/size/format.go:73:         String        100.0%
github.com/nabbar/golib/size/format.go:92:         Int64         100.0%
github.com/nabbar/golib/size/format.go:110:        Int32         100.0%
github.com/nabbar/golib/size/format.go:128:        Int           100.0%
github.com/nabbar/golib/size/parse.go:90:          parseBytes    100.0%
github.com/nabbar/golib/size/parse.go:115:         parseString   95.8%
total:                                              (statements)  95.4%
```

---

## Thread Safety

### Validation

The size package is validated for thread safety using Go's race detector:

```bash
# Run all tests with race detection
CGO_ENABLED=1 go test -race -timeout=10m -v -cover -covermode=atomic ./...
```

**Thread Safety Guarantees**:
- ✅ **Value Type**: Size is a simple `uint64` wrapper, safe to copy
- ✅ **Concurrent Reads**: Multiple goroutines can safely read Size values
- ✅ **Concurrent Writes**: Pointer receiver methods (`Mul`, `Add`, etc.) require external synchronization
- ✅ **Parse/Format**: Stateless operations, safe for concurrent use
- ✅ **Marshaling**: No shared state, thread-safe

### Concurrent Usage Patterns

**Safe Pattern** - Value semantics:
```go
func processFiles(files []string) {
    var wg sync.WaitGroup
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            // Each goroutine works with its own Size value
            fileSize, _ := size.Parse(f.Size())
            fmt.Println(fileSize.String()) // Safe
        }(file)
    }
    wg.Wait()
}
```

**Unsafe Pattern** - Shared mutable state:
```go
func accumulateSizesBad(sizes []size.Size) size.Size {
    total := size.SizeNul
    var wg sync.WaitGroup
    for _, s := range sizes {
        wg.Add(1)
        go func(sz size.Size) {
            defer wg.Done()
            total.Add(sz.Uint64()) // ❌ RACE CONDITION
        }(s)
    }
    wg.Wait()
    return total
}
```

**Safe Pattern** - Synchronized writes:
```go
func accumulateSizesGood(sizes []size.Size) size.Size {
    total := size.SizeNul
    var mu sync.Mutex
    var wg sync.WaitGroup
    for _, s := range sizes {
        wg.Add(1)
        go func(sz size.Size) {
            defer wg.Done()
            mu.Lock()
            total.Add(sz.Uint64()) // ✅ Protected
            mu.Unlock()
        }(s)
    }
    wg.Wait()
    return total
}
```

---

## Writing Tests

### Test Template

```go
package size_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/nabbar/golib/size"
)

var _ = Describe("Feature Name", func() {
    Context("when condition X", func() {
        It("should behave correctly", func() {
            // Arrange
            input := "10MB"
            
            // Act
            result, err := size.Parse(input)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result.MegaBytes()).To(Equal(uint64(10)))
        })
    })
    
    Context("when edge case Y", func() {
        It("should handle gracefully", func() {
            // Test edge case
            result, err := size.Parse("0")
            
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(size.SizeNul))
        })
    })
})
```

### Key Testing Patterns

**1. Parsing Tests**
```go
It("should parse various formats", func() {
    testCases := []struct {
        input    string
        expected uint64
    }{
        {"1024", 1024},
        {"1KB", 1024},
        {"1.5MB", 1572864},
        {"1GB500MB", 1610612736},
    }
    
    for _, tc := range testCases {
        s, err := size.Parse(tc.input)
        Expect(err).ToNot(HaveOccurred(), "input: %s", tc.input)
        Expect(s.Uint64()).To(Equal(tc.expected), "input: %s", tc.input)
    }
})
```

**2. Arithmetic Tests**
```go
It("should detect overflow", func() {
    s := size.Size(math.MaxUint64 / 2)
    err := s.MulErr(3.0)
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("overflow"))
})
```

**3. Marshaling Tests**
```go
It("should round-trip through JSON", func() {
    original := size.ParseUint64(10485760)
    
    data, err := json.Marshal(original)
    Expect(err).ToNot(HaveOccurred())
    
    var decoded size.Size
    err = json.Unmarshal(data, &decoded)
    Expect(err).ToNot(HaveOccurred())
    Expect(decoded).To(Equal(original))
})
```

---

## Best Practices

### Test Organization

**DO**: Group related tests
```go
Describe("Size Parsing", func() {
    Context("with valid input", func() {
        It("should parse simple units", func() { /* ... */ })
        It("should parse complex expressions", func() { /* ... */ })
    })
    
    Context("with invalid input", func() {
        It("should return error for empty string", func() { /* ... */ })
        It("should return error for unknown unit", func() { /* ... */ })
    })
})
```

**DON'T**: Mix unrelated tests
```go
It("should do everything", func() {
    // Parsing test
    s, _ := size.Parse("10MB")
    // Arithmetic test
    s.Add(1024)
    // Formatting test
    _ = s.String()
    // Too many concerns in one test
})
```

### Error Testing

**DO**: Test specific error conditions
```go
It("should return error for negative size", func() {
    _, err := size.Parse("-10MB")
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("negative"))
})
```

**DON'T**: Ignore error details
```go
It("should error on bad input", func() {
    _, err := size.Parse("bad")
    Expect(err).To(HaveOccurred()) // Too vague
})
```

### Table-Driven Tests

**DO**: Use tables for multiple similar cases
```go
DescribeTable("parsing various formats",
    func(input string, expected uint64) {
        s, err := size.Parse(input)
        Expect(err).ToNot(HaveOccurred())
        Expect(s.Uint64()).To(Equal(expected))
    },
    Entry("bytes", "1024", uint64(1024)),
    Entry("kilobytes", "1KB", uint64(1024)),
    Entry("megabytes", "1MB", uint64(1048576)),
    Entry("decimal", "1.5MB", uint64(1572864)),
)
```

### Coverage Goals

- **New Features**: 100% coverage
- **Bug Fixes**: Add regression test
- **Edge Cases**: Test boundaries (0, max, overflow)
- **Error Paths**: Test all error returns

---

## Troubleshooting

### Common Issues

**Problem**: Tests fail with "undefined: Size"
```bash
# Solution: Check imports
import (
    "github.com/nabbar/golib/size"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)
```

**Problem**: Race detector reports issues
```bash
# Solution: Review concurrent access patterns
# Use mutex for shared mutable state
var mu sync.Mutex
mu.Lock()
sharedSize.Add(value)
mu.Unlock()
```

**Problem**: Coverage lower than expected
```bash
# Solution: Check which lines are not covered
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
# Look for uncovered lines and add tests
```

### Debugging Tests

**Verbose Output**:
```bash
go test -v ./...
ginkgo -v --trace
```

**Focus on Failing Test**:
```bash
ginkgo --focus="specific test name"
```

**Step-by-Step Debugging**:
```go
It("should work", func() {
    GinkgoWriter.Println("Debug: input =", input)
    result, err := size.Parse(input)
    GinkgoWriter.Printf("Debug: result = %v, err = %v\n", result, err)
    Expect(err).ToNot(HaveOccurred())
})
```

---

## CI Integration

### GitHub Actions

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
        run: go test -timeout=10m -v -cover -covermode=atomic ./...
      
      - name: Run race detector
        run: CGO_ENABLED=1 go test -race -timeout=10m -v ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### GitLab CI

```yaml
test:size:
  stage: test
  script:
    - cd size
    - go test -timeout=10m -v -cover -covermode=atomic ./...
    - CGO_ENABLED=1 go test -race -timeout=10m -v ./...
  coverage: '/coverage: (\d+\.\d+)% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep "coverage:" | awk '{if ($2 < 95.0) exit 1}'
```

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Resources

- **Ginkgo Docs**: https://onsi.github.io/ginkgo/
- **Gomega Docs**: https://onsi.github.io/gomega/
- **Go Testing**: https://golang.org/pkg/testing/
- **Race Detector**: https://go.dev/doc/articles/race_detector
- **Coverage Tool**: https://go.dev/blog/cover
