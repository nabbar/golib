# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-112%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A5%2080%25-brightgreen)]()

Comprehensive testing documentation for the archive package, covering test execution, race detection, and quality assurance.

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

The archive package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite**
- Total Specs: 112
- Coverage: ≥80%
- Race Detection: ✅ Zero data races
- Execution Time: ~6s (without race), ~13s (with race)

**Coverage Areas**
- Archive operations (TAR, ZIP)
- Compression algorithms (GZIP, BZIP2, LZ4, XZ)
- Helper pipelines with thread safety
- Auto-detection and extraction
- Error handling and edge cases

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

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=interface_test.go

# Pattern matching
ginkgo --focus="compression"

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
```

**Validates**:
- Atomic operations (`atomic.Bool`)
- Mutex protection (`sync.Mutex`)
- Goroutine synchronization (`sync.WaitGroup`)
- Buffer thread safety

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/archive	12.859s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected

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
| Full Suite | ~6s | Without race |
| With `-race` | ~13s | 2x slower (normal) |
| Individual Spec | <100ms | Most tests |
| Compression | 200-500ms | Algorithm-dependent |

---

## Test Coverage

**Target**: ≥80% statement coverage

### Coverage By Category

| Category | Files | Description |
|----------|-------|-------------|
| **Compression** | `compression_algorithms_test.go`, `archive_{gzip,bzip,lz4,xz}_test.go` | All algorithms, header detection, properties |
| **Archives** | `archive_tar_test.go`, `archive_zip_test.go` | Creation, extraction, listing, walk |
| **Helper** | `helper_compress_test.go`, `helper_advanced_test.go` | Pipelines, streaming, edge cases |
| **Extraction** | `extract_test.go` | Auto-detection, path security |
| **Interface** | `interface_test.go` | Parse/Detect wrappers |
| **Errors** | `error_handling_test.go` | Invalid inputs, corruption |
| **Constants** | `archive_const_test.go` | Algorithm enums, marshaling |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
Describe("archive/component", func() {
    BeforeSuite(func() {
        // Global setup
    })
    
    AfterSuite(func() {
        // Global cleanup
    })
    
    Context("Feature or scenario", func() {
        BeforeEach(func() {
            // Per-test setup
        })
        
        AfterEach(func() {
            // Per-test cleanup
        })
        
        It("should do something specific", func() {
            // Test implementation
            Expect(result).To(Equal(expected))
        })
    })
})
```

---

## Thread Safety

Thread safety is critical for the helper subpackage's concurrent operations.

### Concurrency Primitives

```go
// Atomic state flags
atomic.Bool

// Buffer protection
sync.Mutex

// Goroutine lifecycle
sync.WaitGroup
```

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| `helper.deCompressWriter` | `atomic.Bool` + `sync.WaitGroup` | ✅ Race-free |
| `helper.bufNoEOF` | `sync.Mutex` + `atomic.Bool` | ✅ Race-free |
| Compression Pipelines | Independent goroutines | ✅ Parallel-safe |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Helper" ./...

# Stress test
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: Zero data races across all test runs

---

## Test File Organization

| File | Purpose | Specs |
|------|---------|-------|
| `archive_suite_test.go` | Suite initialization | 1 |
| `archive_const_test.go` | Algorithm constants | 8 |
| `archive_tar_test.go` | TAR operations | 15 |
| `archive_zip_test.go` | ZIP operations | 15 |
| `archive_{gzip,bzip,lz4,xz}_test.go` | Algorithm-specific | 6 each |
| `archive_tgz_test.go` | TAR.GZ combined | 6 |
| `compression_algorithms_test.go` | Compression tests | 15 |
| `helper_compress_test.go` | Helper pipelines | 10 |
| `helper_advanced_test.go` | Advanced helper | 8 |
| `interface_test.go` | Interface wrappers | 6 |
| `extract_test.go` | Extraction | 7 |
| `error_handling_test.go` | Error cases | 12 |
| `lorem_ipsum_test.go` | Test data | 0 |

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should compress and decompress data without loss", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should detect gzip compression", func() {
    // Arrange
    var buf bytes.Buffer
    writer, _ := arccmp.Gzip.Writer(&buf)
    writer.Write([]byte("test"))
    writer.Close()
    
    // Act
    alg, reader, err := libarc.DetectCompression(&buf)
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
    Expect(alg).To(Equal(arccmp.Gzip))
    Expect(reader).ToNot(BeNil())
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
```

**4. Always Cleanup Resources**
```go
defer reader.Close()
defer os.Remove(tempFile)
```

**5. Test Edge Cases** - Empty input, nil values, large data, etc.

**6. Avoid External Dependencies** - No remote resources or external services

### Test Template

```go
var _ = Describe("archive/new_feature", func() {
    Context("When using new feature", func() {
        var (
            testData   []byte
            tempFile   string
        )

        BeforeEach(func() {
            testData = []byte("test data")
        })

        AfterEach(func() {
            if tempFile != "" {
                os.Remove(tempFile)
            }
        })

        It("should perform expected behavior", func() {
            // Arrange
            input := prepareInput(testData)
            
            // Act
            result, err := newFeature(input)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expectedResult))
        })

        It("should handle error case", func() {
            _, err := newFeature(invalidInput)
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
- ✅ Avoid global mutable state
- ✅ Create test data on-demand
- ❌ Don't rely on test execution order

**Test Data**
- Use `loremIpsum` constant for text data (302KB)
- Generate archives in `BeforeSuite` or on-demand
- Clean up in `AfterSuite`
- Use `filepath.Join()` for cross-platform paths
- Use `os.MkdirTemp()` for isolation

**Assertions**
```go
// ✅ Good
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))

// ❌ Avoid
Expect(value == expected).To(BeTrue())
```
- Use specific matchers for better error messages
- One behavior per test
- Use `GinkgoWriter` for debug output

**Concurrency Testing**
```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // Independent operation
        }(i)
    }
    wg.Wait()
})
```
- Always run with `-race` during development
- Test concurrent operations explicitly
- Verify cleanup with `sync.WaitGroup`
- Use atomic operations and mutexes

**Performance**
- Keep tests fast (small data)
- Use parallel execution (`ginkgo -p`)
- Target: <6s full suite, <100ms per spec

**Error Handling**
```go
// ✅ Good
It("should handle errors", func() {
    result, err := operation()
    Expect(err).ToNot(HaveOccurred())
    defer result.Close()
})

// ❌ Bad
It("should do something", func() {
    result, _ := operation() // Don't ignore errors!
})
```

---

## Troubleshooting

**Leftover Test Files**
```bash
# Clean manually if needed
rm -f lorem_ipsum*.{txt,tar,zip,gz,bz2,lz4,xz}
rm -rf extract_all_dir
```

**Stale Coverage**
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Parallel Test Failures**
- Check for shared resources or global state
- Use synchronization or make tests independent

**Import Cycles**
- Use `package archive_test` convention to avoid cycles

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing mutex locks
- Unsynchronized goroutines

Example fix:
```go
// ❌ Bad: Direct access
if o.b.Len() < 1 {  // Race condition

// ✅ Good: Protected access
if o.Len() < 1 {  // Len() uses mutex
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
ginkgo --timeout=10s
```
Check for:
- Goroutine leaks (missing `wg.Done()`)
- Unclosed resources
- Mutex deadlocks

**Debugging**
```bash
# Single test
ginkgo --focus="should compress and decompress"

# Specific file
ginkgo --focus-file=extract_test.go

# Verbose output
ginkgo -v --trace
```

Use `GinkgoWriter` for debug output:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
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
- [ ] Coverage maintained: ≥80%
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Test duration reasonable (<10s)

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

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Archive Package Contributors
