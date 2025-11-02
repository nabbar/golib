# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-173%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-93.8%25-brightgreen)]()

Comprehensive testing documentation for the version package, covering test execution, race detection, benchmarking, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Benchmarking](#benchmarking)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The version package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions and detailed failure reporting.

### Test Suite Statistics

```
Total Specs:      173
Coverage:         93.8%
Race Detection:   ✅ Zero data races
Execution Time:   ~0.11s (without race), ~0.18s (with race)
Test Files:       7 files (~80KB total)
```

### Coverage Breakdown

| Component | Coverage | Notes |
|-----------|----------|-------|
| Version Creation | 100% | All paths tested |
| License Management | 92.3% | All 11 licenses covered |
| Go Version Validation | 88.9% | Edge cases included |
| Error Handling | 83.3% | All error codes tested |
| Getters/Formatters | 100% | Full coverage |
| Print Methods | 0% | Intentionally untested (stderr output) |

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

### Ginkgo v2

**BDD Testing Framework** ([documentation](https://onsi.github.io/ginkgo/))

**Key Features**:
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering and focus
- Detailed failure reporting

**Example Structure**:
```go
var _ = Describe("Version Creation", func() {
    Context("with valid parameters", func() {
        It("should create a version instance", func() {
            v := version.NewVersion(...)
            Expect(v).ToNot(BeNil())
        })
    })
})
```

### Gomega

**Matcher Library** ([documentation](https://onsi.github.io/gomega/))

**Key Features**:
- Readable assertion syntax
- Extensive built-in matchers
- Custom matcher support
- Detailed failure messages

**Common Matchers**:
```go
Expect(value).To(Equal(expected))
Expect(value).ToNot(BeNil())
Expect(value).To(BeEmpty())
Expect(value).To(ContainSubstring("text"))
Expect(value).To(BeTemporally(">=", time.Now()))
```

---

## Running Tests

### Standard Commands

```bash
# Run all tests in current package
go test

# Run all tests recursively
go test ./...

# Verbose output with test names
go test -v ./...

# Run specific test file
go test -v version_test.go version_suite_test.go

# Run tests matching pattern
go test -v -run TestVersion

# Count test runs (useful for flaky test detection)
go test -count=10 ./...
```

### Coverage Analysis

```bash
# Basic coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go tool cover -func=coverage.out | grep -E "version\.go|license\.go"
```

### Race Detection

```bash
# Enable race detector (requires CGO_ENABLED=1)
CGO_ENABLED=1 go test -race ./...

# Race detection with verbose output
CGO_ENABLED=1 go test -race -v ./...

# Race detection with coverage
CGO_ENABLED=1 go test -race -coverprofile=coverage.out ./...
```

---

## Test Coverage

### Current Coverage: 93.8%

#### Covered Components (100%)

- ✅ `NewVersion()` - Version instance creation
- ✅ `GetPackage()`, `GetDescription()`, `GetRelease()`, etc. - All getters
- ✅ `GetHeader()`, `GetInfo()`, `GetAppId()` - Formatted output
- ✅ `GetLicense*()` - All license methods
- ✅ License boilerplate generation for all 11 licenses
- ✅ Date parsing and time handling
- ✅ Package path extraction via reflection

#### Partially Covered

- ⚠️ `CheckGo()` - 88.9% (some error paths)
- ⚠️ `GetBoilerPlate()` - 92.3% (edge cases)
- ⚠️ `GetLicense()` - 92.3% (default case)

#### Intentionally Uncovered

- ❌ `PrintInfo()` - 0% (writes to stderr, tested indirectly)
- ❌ `PrintLicense()` - 0% (writes to stderr, tested indirectly)

### Coverage Goals

- **Minimum**: 90% overall coverage
- **Target**: 95% for core functionality
- **Acceptable**: 0% for output-only methods (Print*)

---

## Thread Safety

### Race Detection Results

```
✅ Zero data races detected
✅ All concurrent tests pass
✅ Immutable design ensures safety
```

### Concurrent Test Scenarios

1. **Concurrent Reads**
   ```go
   It("should be safe to read from multiple goroutines", func() {
       done := make(chan bool, 10)
       for i := 0; i < 10; i++ {
           go func() {
               Expect(v.GetPackage()).To(Equal(testPackage))
               done <- true
           }()
       }
       // Wait for all goroutines
   })
   ```

2. **Concurrent Method Calls**
   ```go
   It("should handle concurrent method calls", func() {
       b.RunParallel(func(pb *testing.PB) {
           for pb.Next() {
               _ = v.GetHeader()
               _ = v.GetInfo()
               _ = v.GetLicenseName()
           }
       })
   })
   ```

3. **Concurrent CheckGo**
   ```go
   It("should safely check Go version from multiple goroutines", func() {
       for i := 0; i < 10; i++ {
           go func() {
               err := v.CheckGo("1.18", ">=")
               Expect(err).To(BeNil())
           }()
       }
   })
   ```

---

## Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run with memory statistics
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkNewVersion -benchmem

# Run benchmarks multiple times for accuracy
go test -bench=. -benchtime=10s -count=5

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Benchmark Results

```
Benchmark                        Iterations   Time/op    Memory/op   Allocs/op
──────────────────────────────────────────────────────────────────────────────
BenchmarkNewVersion-12           13,978,096   94.00 ns   144 B       1
BenchmarkGetHeader-12             3,122,145  383.3 ns    272 B       8
BenchmarkGetInfo-12               4,029,932  296.1 ns    160 B       5
BenchmarkGetLicenseLegal-12      58,196,515   19.50 ns     0 B       0
BenchmarkGetLicenseBoiler-12      1,367,912  817.0 ns   1201 B       5
BenchmarkGetLicenseFull-12          266,406 4088 ns    12384 B      11
BenchmarkCheckGo-12                 291,633 4749 ns     1982 B      27
BenchmarkGetLicenseLegal_Multi       53,605 26428 ns  114371 B      10
BenchmarkConcurrentAccess-12      3,585,799  331.3 ns    432 B      13
```

### Performance Analysis

**Fast Operations** (< 100ns):
- `NewVersion()`: 94ns, 1 allocation
- `GetLicenseLegal()`: 19.5ns, 0 allocations (pre-compiled strings)

**Medium Operations** (100ns - 1µs):
- `GetHeader()`: 383ns, 8 allocations (string formatting)
- `GetInfo()`: 296ns, 5 allocations
- `GetLicenseBoiler()`: 817ns, 5 allocations

**Slower Operations** (> 1µs):
- `GetLicenseFull()`: 4.1µs, 11 allocations (combines boilerplate + legal)
- `CheckGo()`: 4.7µs, 27 allocations (version parsing)
- `GetLicenseLegal_Multiple()`: 26.4µs (multiple licenses)

---

## Test Organization

### File Structure

```
version/
├── version_suite_test.go      # Suite setup, shared fixtures
├── version_test.go            # Version creation and getters (15KB)
├── license_test.go            # License functionality (21KB)
├── checkgo_test.go            # Go version validation (12KB)
├── error_test.go              # Error handling (12KB)
├── coverage_test.go           # Coverage improvements (13KB)
└── benchmark_test.go          # Performance benchmarks (4.8KB)
```

### Test Categories

1. **version_test.go** - Core functionality
   - Version creation with various parameters
   - Getter methods
   - Formatted output (Header, Info, AppId)
   - Edge cases (empty values, special characters)
   - Package path extraction
   - Concurrency safety

2. **license_test.go** - License management
   - All 11 license types
   - License name retrieval
   - Legal text generation
   - Boilerplate generation
   - Full license text
   - Multiple license combinations
   - License consistency checks

3. **checkgo_test.go** - Go version validation
   - Version constraint operators (==, !=, >, >=, <, <=, ~>)
   - Valid and invalid constraints
   - Runtime version detection
   - Error handling
   - Edge cases

4. **error_test.go** - Error handling
   - All error codes
   - Error messages
   - Parent error chains
   - Error interface compliance
   - Concurrent error creation

5. **coverage_test.go** - Coverage improvements
   - Edge cases for all license types
   - Extreme values (long strings, unicode)
   - Invalid inputs
   - Concurrent operations

6. **benchmark_test.go** - Performance testing
   - Version creation
   - All getter methods
   - License operations
   - Concurrent access patterns

---

## Writing Tests

### Test Structure

```go
var _ = Describe("Feature Name", func() {
    var (
        // Shared variables
        v version.Version
    )

    BeforeEach(func() {
        // Setup before each test
        v = version.NewVersion(...)
    })

    Context("specific scenario", func() {
        It("should behave correctly", func() {
            // Test implementation
            result := v.GetSomething()
            Expect(result).To(Equal(expected))
        })
    })
})
```

### Best Practices

1. **Use Descriptive Names**
   ```go
   It("should return correct package name", func() { ... })
   // Better than: It("works", func() { ... })
   ```

2. **Test One Thing Per Spec**
   ```go
   It("should return correct release version", func() {
       Expect(v.GetRelease()).To(Equal("v1.2.3"))
   })
   ```

3. **Use Contexts for Grouping**
   ```go
   Context("with valid parameters", func() { ... })
   Context("with invalid parameters", func() { ... })
   ```

4. **Avoid Test Interdependence**
   ```go
   // Bad: Tests depend on execution order
   // Good: Each test is independent
   BeforeEach(func() {
       // Reset state
   })
   ```

5. **Test Error Cases**
   ```go
   It("should return error for invalid constraint", func() {
       err := v.CheckGo("1.18", "invalid")
       Expect(err).ToNot(BeNil())
       Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
   })
   ```

### Common Patterns

**Testing String Output**:
```go
It("should contain expected text", func() {
    output := v.GetHeader()
    Expect(output).To(ContainSubstring("MyApp"))
    Expect(output).To(ContainSubstring("v1.2.3"))
})
```

**Testing Time Values**:
```go
It("should parse time correctly", func() {
    t := v.GetTime()
    Expect(t).To(BeTemporally(">=", expectedTime))
})
```

**Testing Concurrency**:
```go
It("should be thread-safe", func() {
    done := make(chan bool, 10)
    for i := 0; i < 10; i++ {
        go func() {
            defer GinkgoRecover()
            Expect(v.GetPackage()).ToNot(BeEmpty())
            done <- true
        }()
    }
    for i := 0; i < 10; i++ {
        Eventually(done).Should(Receive())
    }
})
```

---

## Best Practices

### 1. Run Tests Before Committing

```bash
# Minimum test suite
go test ./...

# Recommended test suite
CGO_ENABLED=1 go test -race -cover ./...

# Full validation
go fmt ./...
go vet ./...
CGO_ENABLED=1 go test -race -coverprofile=coverage.out ./...
```

### 2. Maintain High Coverage

- Aim for ≥90% overall coverage
- Test all public APIs
- Include edge cases and error paths
- Document intentionally untested code

### 3. Use Race Detection

Always run tests with `-race` before merging:
```bash
CGO_ENABLED=1 go test -race ./...
```

### 4. Write Deterministic Tests

- Avoid time-dependent tests
- Use fixed test data
- Don't rely on external state

### 5. Keep Tests Fast

- Current suite: ~0.11s (excellent)
- Target: < 1s for unit tests
- Use benchmarks for performance testing

### 6. Test Concurrency

Include concurrent access patterns:
```go
It("should handle concurrent access", func() {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _ = v.GetHeader()
        }
    })
})
```

---

## Troubleshooting

### Common Issues

#### 1. Race Detection Failures

**Problem**: `go test -race` fails
**Solution**:
```bash
# Ensure CGO is enabled
export CGO_ENABLED=1
go test -race ./...
```

#### 2. Coverage Not Generated

**Problem**: No coverage output
**Solution**:
```bash
# Specify output file
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 3. Tests Timeout

**Problem**: Tests hang or timeout
**Solution**:
```bash
# Increase timeout
go test -timeout 30s ./...

# Check for deadlocks in concurrent tests
```

#### 4. Flaky Tests

**Problem**: Tests pass/fail randomly
**Solution**:
```bash
# Run multiple times to detect flakiness
go test -count=100 ./...

# Check for race conditions
CGO_ENABLED=1 go test -race ./...
```

#### 5. Import Cycle

**Problem**: Import cycle detected
**Solution**:
- Use `_test` package suffix
- Avoid circular dependencies
- Test through public API only

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
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
```

### GitLab CI Example

```yaml
test:
  image: golang:1.21
  script:
    - go fmt ./...
    - go vet ./...
    - CGO_ENABLED=1 go test -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  coverage: '/total:.*?(\d+\.\d+)%/'
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
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 90" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 90%"
    exit 1
fi

echo "All checks passed!"
```

---

## Test Metrics

### Execution Time

```
Without Race Detection:  ~0.11s
With Race Detection:     ~0.18s
Benchmark Suite:         ~15s
```

### Resource Usage

```
Memory per Test:  < 10MB
CPU Usage:        Minimal (single-threaded)
Disk I/O:         None (in-memory only)
```

### Quality Metrics

```
Code Coverage:     93.8%
Race Conditions:   0
Flaky Tests:       0
Test Reliability:  100%
```

---

## References

- **Ginkgo Documentation**: https://onsi.github.io/ginkgo/
- **Gomega Matchers**: https://onsi.github.io/gomega/
- **Go Testing**: https://golang.org/pkg/testing/
- **Race Detector**: https://golang.org/doc/articles/race_detector.html
- **Coverage Tools**: https://go.dev/blog/cover

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Version Package Contributors
