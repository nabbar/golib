# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-112%20Passed-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-81.7%25-brightgreen)]()

Comprehensive testing documentation for the multi package, covering test execution, race detection, and thread-safety validation.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Test File Organization](#test-file-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)
- [Quality Checklist](#quality-checklist)
- [Resources](#resources)

---

## Overview

The multi package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with emphasis on thread-safety and concurrent operations.

**Test Suite Statistics**
- Total Specs: 112 passed, 1 skipped, 0 failed
- Coverage: 81.7% of statements
- Race Detection: ✅ Zero data races
- Execution Time: ~0.13s (without race), ~1.18s (with race)

**Test Categories**
- Constructor and initialization
- Writer operations (Add, Write, WriteString, Clean)
- Reader operations (SetInput, Read, Close)
- Copy operations and integration
- Concurrent operations with thread safety
- Edge cases and error handling
- Performance benchmarks with gmeasure

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (critical for concurrent operations)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI for better output
ginkgo -cover -race

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization with `Describe`, `Context`, `It`
- Setup/teardown hooks: `BeforeEach`, `AfterEach`
- Parallel execution support
- Rich CLI with focus and filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Expressive assertion syntax
- Extensive built-in matchers
- Detailed failure messages

**gmeasure** - Performance measurement ([docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code))
- Integration with Ginkgo for benchmarks
- Statistical analysis (mean, median, stddev)
- Memory allocation tracking

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

# Specific timeout
go test -timeout=10m ./...

# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=writer_test.go

# Pattern matching
ginkgo --focus="concurrent"

# Verbose output with trace
ginkgo -v --trace

# Generate JUnit report for CI
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for validating thread safety**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race -timeout=10m -v -cover -covermode=atomic ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race -cover
```

**What it validates**:
- Atomic operations (`atomic.Value`, `atomic.Int64`)
- sync.Map concurrent access
- Type wrapper consistency
- Goroutine synchronization

**Expected Output**:
```bash
# ✅ Success (no races)
ok  	github.com/nabbar/golib/ioutils/multi	1.180s	coverage: 81.7%

# ❌ Race detected (example - should not occur)
==================
WARNING: DATA RACE
Write at 0x... by goroutine ...
Previous write at 0x... by goroutine ...
==================
```

**Current Status**: ✅ Zero data races across all concurrent scenarios

### Performance Benchmarks

The test suite includes gmeasure-based benchmarks:

```bash
# Run tests with benchmark output
go test -v ./... 2>&1 | grep -A 5 "benchmark"

# View benchmark reports in test output
ginkgo -v | grep -A 10 "Benchmark"
```

**Benchmark Coverage**:
- Constructor creation (`New()`)
- Write operations (single writer, multiple writers, large data)
- WriteString operations
- Read operations (small and large data)
- Copy operations (various sizes)
- AddWriter performance
- Clean operations
- SetInput operations
- Memory allocation measurements

---

## Test Coverage

**Target**: ≥80% statement coverage  
**Current**: 81.7%

### Coverage By Category

| Category | Files | Lines Tested | Description |
|----------|-------|--------------|-------------|
| **Constructor** | `constructor_test.go` | `interface.go` | New(), interface compliance, DiscardCloser |
| **Writers** | `writer_test.go` | `model.go` | AddWriter, Writer, Write, WriteString, Clean |
| **Readers** | `reader_test.go` | `model.go` | SetInput, Reader, Read, Close, error propagation |
| **Copy** | `copy_test.go` | `model.go` | Copy method, integration scenarios |
| **Concurrency** | `concurrent_test.go` | All | Thread-safe operations, race conditions |
| **Edge Cases** | `edge_cases_test.go` | All | Error handling, boundaries, special cases |
| **Benchmarks** | `benchmark_test.go` | All | Performance measurements with gmeasure |

### View Coverage Details

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML visualization
go tool cover -html=coverage.out -o coverage.html

# Open in browser (macOS)
open coverage.html

# Open in browser (Linux)
xdg-open coverage.html
```

**Coverage Report Interpretation**:
```
github.com/nabbar/golib/ioutils/multi/discard.go:49:  Read         100.0%
github.com/nabbar/golib/ioutils/multi/discard.go:58:  Write        100.0%
github.com/nabbar/golib/ioutils/multi/discard.go:65:  Close        100.0%
github.com/nabbar/golib/ioutils/multi/interface.go:109: New         100.0%
github.com/nabbar/golib/ioutils/multi/model.go:73:    AddWriter    100.0%
github.com/nabbar/golib/ioutils/multi/model.go:111:   Clean        100.0%
...
total:                                                               81.7%
```

---

## Thread Safety

Thread safety is the **primary focus** of this package. All operations must be safe for concurrent use.

### Concurrency Primitives

| Primitive | Usage | Thread-Safety Guarantee |
|-----------|-------|------------------------|
| **`atomic.Value`** | Reader/Writer storage | Lock-free atomic load/store |
| **`atomic.Int64`** | Writer key counter | Lock-free increment |
| **`sync.Map`** | Writer registry | Concurrent read/write safe |
| **`readerWrapper`** | Type consistency | Ensures atomic.Value type safety |
| **`io.MultiWriter`** | Consistent output type | Always same concrete type |

### Verified Scenarios

**Concurrent Writes**
```go
// Multiple goroutines writing simultaneously
for i := 0; i < 100; i++ {
    go func(id int) {
        m.Write([]byte(fmt.Sprintf("msg%d", id)))
    }(i)
}
// ✅ No data races, all writes succeed
```

**Concurrent AddWriter**
```go
// Adding writers from multiple goroutines
for i := 0; i < 50; i++ {
    go func() {
        var buf bytes.Buffer
        m.AddWriter(&buf)
    }()
}
// ✅ All writers registered correctly
```

**Concurrent SetInput**
```go
// Multiple goroutines setting input
for i := 0; i < 50; i++ {
    go func() {
        input := io.NopCloser(strings.NewReader("data"))
        m.SetInput(input)
    }()
}
// ✅ Last set wins, no corruption
```

**Mixed Concurrent Operations**
```go
// Simultaneous writes, AddWriter, Clean, SetInput
// ✅ All operations thread-safe
// ✅ Zero data races detected
```

### Race Detection Commands

```bash
# Full test suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Concurrent" ./...

# Stress test (run multiple times)
for i in {1..10}; do
    CGO_ENABLED=1 go test -race ./... || break
done

# Race detection with timeout
CGO_ENABLED=1 go test -race -timeout=10m ./...
```

### Thread Safety Test Structure

```go
Describe("Concurrent Operations", func() {
    Context("writes", func() {
        It("should handle concurrent writes safely", func() {
            var wg sync.WaitGroup
            for i := 0; i < 100; i++ {
                wg.Add(1)
                go func(id int) {
                    defer wg.Done()
                    m.Write([]byte(fmt.Sprintf("msg%d ", id)))
                }(i)
            }
            wg.Wait()
            // Verified with -race detector
        })
    })
})
```

**Important Note**: While the Multi wrapper is thread-safe, the underlying `io.ReadCloser` may not be. For example, `strings.Reader` is not safe for concurrent reads. The Multi package correctly synchronizes its own operations but cannot make unsafe wrapped objects safe.

---

## Test File Organization

| File | Purpose | Specs | Key Tests |
|------|---------|-------|-----------|
| **`suite_test.go`** | Ginkgo suite initialization | 1 | Test runner setup |
| **`constructor_test.go`** | Constructor and types | 9 | New(), interfaces, DiscardCloser |
| **`writer_test.go`** | Writer operations | 21 | AddWriter, Write, WriteString, Clean |
| **`reader_test.go`** | Reader operations | 18 | SetInput, Read, Close, errors |
| **`copy_test.go`** | Copy operations | 15 | Copy(), integration, errors |
| **`concurrent_test.go`** | Thread safety | 12 | Concurrent operations, races |
| **`edge_cases_test.go`** | Edge cases | 23 | Errors, boundaries, special data |
| **`benchmark_test.go`** | Performance | 13 | gmeasure benchmarks |

**Total**: 112 specs (1 intentionally skipped)

---

## Writing Tests

### Test Structure

Follow Ginkgo's BDD pattern:

```go
var _ = Describe("Multi/Feature", func() {
    var m multi.Multi

    BeforeEach(func() {
        m = multi.New()
    })

    AfterEach(func() {
        // Cleanup if needed
        if m != nil {
            m.Close()
        }
    })

    Context("When using feature", func() {
        It("should behave correctly", func() {
            // Arrange
            var buf bytes.Buffer
            m.AddWriter(&buf)

            // Act
            n, err := m.Write([]byte("test"))

            // Assert
            Expect(err).NotTo(HaveOccurred())
            Expect(n).To(Equal(4))
            Expect(buf.String()).To(Equal("test"))
        })
    })
})
```

### Test Guidelines

**1. Use Descriptive Names**
```go
// ✅ Good
It("should broadcast writes to all registered writers", func() { ... })

// ❌ Bad
It("test write", func() { ... })
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should handle multiple writers", func() {
    // Arrange
    var buf1, buf2, buf3 bytes.Buffer
    m.AddWriter(&buf1, &buf2, &buf3)

    // Act
    m.Write([]byte("data"))

    // Assert
    Expect(buf1.Len()).To(Equal(4))
    Expect(buf2.Len()).To(Equal(4))
    Expect(buf3.Len()).To(Equal(4))
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).NotTo(HaveOccurred())
Expect(err).To(MatchError(multi.ErrInstance))
Expect(list).To(HaveLen(3))
Expect(num).To(BeNumerically(">", 0))
```

**4. Test Both Success and Failure**
```go
Context("error handling", func() {
    It("should succeed with valid input", func() {
        input := io.NopCloser(strings.NewReader("valid"))
        m.SetInput(input)
        Expect(m.Reader()).NotTo(BeNil())
    })

    It("should handle nil input gracefully", func() {
        m.SetInput(nil)
        // Should use DiscardCloser, not panic
        buf := make([]byte, 10)
        n, err := m.Read(buf)
        Expect(err).NotTo(HaveOccurred())
        Expect(n).To(Equal(0))
    })
})
```

**5. Always Cleanup Resources**
```go
It("should close input properly", func() {
    input := io.NopCloser(strings.NewReader("data"))
    m.SetInput(input)
    
    err := m.Close()
    Expect(err).NotTo(HaveOccurred())
})
```

### Benchmark Template

```go
Describe("Performance Benchmarks", func() {
    It("should benchmark operation", func() {
        experiment := gmeasure.NewExperiment("Operation Name")
        AddReportEntry(experiment.Name, experiment)

        experiment.Sample(func(idx int) {
            experiment.MeasureDuration("metric-name", func() {
                // Operation to measure
            })
        }, gmeasure.SamplingConfig{N: 1000})

        // Optional assertion on performance
        Expect(experiment.GetStats("metric-name").DurationFor(gmeasure.StatMean)).
            To(BeNumerically("<", 100*time.Microsecond))
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test creates its own Multi instance
- ✅ Use `BeforeEach` for setup, `AfterEach` for cleanup
- ✅ Avoid global state or shared variables
- ✅ Tests can run in any order
- ❌ Don't depend on test execution order

**Concurrency Testing**
```go
// ✅ Good: Proper synchronization
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.Write([]byte(fmt.Sprintf("msg%d", id)))
        }(i)
    }
    wg.Wait() // Ensure all goroutines complete
})

// ❌ Bad: No synchronization
It("should be concurrent", func() {
    for i := 0; i < 100; i++ {
        go m.Write([]byte("msg"))
    }
    // Test may finish before goroutines!
})
```

**Error Assertions**
```go
// ✅ Good: Check specific errors
It("should return ErrInstance", func() {
    // Create scenario for ErrInstance
    err := m.Read(buf)
    Expect(err).To(MatchError(multi.ErrInstance))
})

// ❌ Bad: Generic check
It("should error", func() {
    err := m.Read(buf)
    Expect(err).To(HaveOccurred()) // Not specific enough
})
```

**Resource Management**
```go
// ✅ Good: Explicit cleanup
var input io.ReadCloser
BeforeEach(func() {
    input = io.NopCloser(strings.NewReader("data"))
})
AfterEach(func() {
    if input != nil {
        input.Close()
    }
})

// ❌ Bad: No cleanup
It("test", func() {
    input := os.Open("file") // Potential leak
    m.SetInput(input)
})
```

**Test Data**
- Use small, predictable data for most tests
- Test large data (1MB) in dedicated performance tests
- Use helper functions for creating test data
- Clean up any temporary files

**Performance**
- Keep individual specs fast (<100ms typical)
- Use parallel execution when possible (`ginkgo -p`)
- Target: <1s for full suite without race detection
- Target: <2s with race detection

---

## Troubleshooting

**Common Issues**

### Race Conditions

```bash
# Detect races
CGO_ENABLED=1 go test -race ./... 2>&1 | tee race-log.txt

# Analyze race report
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

**What to look for**:
- Unprotected shared variable access
- Missing atomic operations
- Concurrent access to non-thread-safe types

**Example Fix**:
```go
// ❌ Bad: Direct access (race)
if value == expected {

// ✅ Good: Atomic access
if atomic.LoadInt64(&value) == expected {
```

### CGO Not Available

```bash
# Install build tools
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Verify
export CGO_ENABLED=1
go test -race ./...
```

### Test Failures

```bash
# Run specific test
ginkgo --focus="should handle concurrent writes"

# Verbose output
ginkgo -v --trace

# Debug with GinkgoWriter
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
```

### Stale Coverage

```bash
# Clear test cache
go clean -testcache

# Regenerate coverage
go test -coverprofile=coverage.out ./...
```

### Slow Tests

```bash
# Identify slow tests
ginkgo -v | grep "seconds"

# Set timeout
ginkgo --timeout=10s
```

Check for:
- Large data in loops
- Missing parallelization
- Unnecessary sleeps
- Resource leaks

---

## CI Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.18', '1.19', '1.20', '1.21']
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -timeout=10m -cover -covermode=atomic ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m -v ./...
      
      - name: Generate coverage
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
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
go test -cover ./... | grep -E "coverage:" | grep -E "[89][0-9]\.[0-9]%|100\.0%" || {
    echo "Coverage below 80%!"
    exit 1
}

echo "All checks passed!"
```

### Makefile Example

```makefile
.PHONY: test test-race test-cover test-all

test:
	go test -v ./...

test-race:
	CGO_ENABLED=1 go test -race -v -timeout=10m ./...

test-cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-all: test test-race test-cover
	@echo "All tests passed!"
```

---

## Quality Checklist

Before merging code, verify:

- [ ] **All tests pass**: `go test ./...`
- [ ] **Race detection clean**: `CGO_ENABLED=1 go test -race ./...`
- [ ] **Coverage maintained**: ≥80% (currently 81.7%)
- [ ] **New features tested**: Add tests for new functionality
- [ ] **Error cases covered**: Test failure scenarios
- [ ] **Thread safety validated**: Concurrent operations tested
- [ ] **Benchmarks updated**: Performance regressions checked
- [ ] **Documentation updated**: README.md and TESTING.md
- [ ] **Examples provided**: Code samples for new features
- [ ] **Test duration reasonable**: <2s for full suite with race

---

## Resources

### Testing Frameworks
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [gmeasure Benchmarking](https://onsi.github.io/gomega/#gmeasure-benchmarking-code)
- [Go Testing Package](https://pkg.go.dev/testing)

### Concurrency
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)
- [sync/atomic Package](https://pkg.go.dev/sync/atomic)

### Performance
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

### Related Documentation
- [README.md](README.md) - Package overview and usage
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/multi) - API reference
- [GitHub Issues](https://github.com/nabbar/golib/issues) - Bug reports and features

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Test Framework**: Ginkgo v2 + Gomega  
**Maintained By**: IOUtils/Multi Package Contributors
