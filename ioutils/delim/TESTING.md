# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-198%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Comprehensive testing documentation for the delim package, covering test execution, race detection, benchmarking, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Performance Benchmarks](#performance-benchmarks)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The delim package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions and **gmeasure** for performance benchmarks.

**Test Suite Statistics**
- Total Specs: 198 (168 functional + 30 benchmarks)
- Coverage: 100% statement coverage
- Race Detection: ✅ Zero data races
- Execution Time: ~0.17s (without race), ~2.1s (with race)

**Coverage Areas**
- Constructor with various parameters
- Read operations (Read, ReadBytes, UnRead)
- Write operations (WriteTo, Copy)
- Edge cases (Unicode, binary, empty data, boundaries)
- DiscardCloser functionality
- Concurrency and thread safety validation
- Performance benchmarks (30 scenarios)

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

# Run with full verbosity
go test -v -cover ./...

# Using Ginkgo CLI
ginkgo -cover -race
```

**Expected Output**:
```
=== RUN   TestDelim
Running Suite: IOUtils/Delim Package Suite
Ran 198 of 198 Specs in 0.159 seconds
SUCCESS! -- 198 Passed | 0 Failed | 0 Pending | 0 Skipped

PASS
coverage: 100.0% of statements
ok      github.com/nabbar/golib/ioutils/delim    0.173s
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Parallel execution support
- Rich CLI with filtering and reporting

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Async testing support

**gmeasure** - Performance measurement ([docs](https://onsi.github.io/gomega/#gmeasure-benchmarking-code))
- Statistical benchmarking
- Min/Median/Mean/StdDev/Max metrics
- Integration with Ginkgo reporting
- Replaced deprecated `Measure()` API

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

# With coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in HTML
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Verbose with coverage
ginkgo -v -cover

# Specific test file
ginkgo --focus-file=read_test.go

# Pattern matching
ginkgo --focus="ReadBytes"

# Skip benchmarks
ginkgo --skip="Benchmarks"

# Parallel execution (with caution)
ginkgo -p

# Generate JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for concurrent operations validation**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Full command with verbosity
CGO_ENABLED=1 go test -race -v -cover ./...
```

**Validates**:
- Buffer access patterns
- Goroutine synchronization (benchmarks use goroutines internally)
- Concurrent instance usage
- Resource cleanup

**Expected Output**:
```bash
# ✅ Success - No races detected
=== RUN   TestDelim
Ran 198 of 198 Specs in 1.053 seconds
SUCCESS! -- 198 Passed | 0 Failed

PASS
coverage: 100.0% of statements
ok      github.com/nabbar/golib/ioutils/delim    2.098s

# ❌ Race detected (should not happen)
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: ✅ Zero data races detected across all test runs

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# Block profiling
go test -blockprofile=block.out ./...
go tool pprof block.out

# Trace
go test -trace=trace.out ./...
go tool trace trace.out
```

---

## Test Coverage

**Target**: 100% statement coverage ✅ **Achieved**

### Coverage Breakdown

```bash
# Generate and view coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Output**:
```
github.com/nabbar/golib/ioutils/delim/discard.go:70:    Read            100.0%
github.com/nabbar/golib/ioutils/delim/discard.go:80:    Write           100.0%
github.com/nabbar/golib/ioutils/delim/discard.go:88:    Close           100.0%
github.com/nabbar/golib/ioutils/delim/interface.go:140: New             100.0%
github.com/nabbar/golib/ioutils/delim/io.go:36:         Reader          100.0%
github.com/nabbar/golib/ioutils/delim/io.go:59:         Copy            100.0%
github.com/nabbar/golib/ioutils/delim/io.go:90:         Read            100.0%
github.com/nabbar/golib/ioutils/delim/io.go:130:        UnRead          100.0%
github.com/nabbar/golib/ioutils/delim/io.go:178:        ReadBytes       100.0%
github.com/nabbar/golib/ioutils/delim/io.go:206:        Close           100.0%
github.com/nabbar/golib/ioutils/delim/io.go:242:        WriteTo         100.0%
github.com/nabbar/golib/ioutils/delim/model.go:52:      Delim           100.0%
github.com/nabbar/golib/ioutils/delim/model.go:59:      getDelimByte    100.0%
total:                                                   (statements)    100.0%
```

### Coverage By Category

| Category | Test Files | Coverage |
|----------|-----------|----------|
| **Constructor** | `constructor_test.go` | 100% |
| **Read Operations** | `read_test.go` | 100% |
| **Write Operations** | `write_test.go` | 100% |
| **Edge Cases** | `edge_cases_test.go` | 100% |
| **DiscardCloser** | `discard_test.go` | 100% |
| **Concurrency** | `concurrency_test.go` | 100% |
| **Benchmarks** | `benchmark_test.go` | N/A |
| **Overall** | All tests | **100%** |

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("BufferDelim Component", func() {
    Context("When using specific feature", func() {
        var (
            bd     BufferDelim
            reader io.ReadCloser
        )
        
        BeforeEach(func() {
            // Per-test setup
            data := "test\ndata\n"
            reader = io.NopCloser(strings.NewReader(data))
            bd = delim.New(reader, '\n', 0)
        })
        
        AfterEach(func() {
            // Per-test cleanup
            if bd != nil {
                bd.Close()
            }
        })
        
        It("should behave correctly", func() {
            // Test implementation
            result, err := bd.ReadBytes()
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal([]byte("test\n")))
        })
    })
})
```

---

## Thread Safety

The delim package is thread-safe for **independent instances** (one instance per goroutine).

### Concurrency Model

```
✅ Safe: Multiple instances, different goroutines
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Goroutine 1 │  │ Goroutine 2 │  │ Goroutine 3 │
│  Instance A │  │  Instance B │  │  Instance C │
└─────────────┘  └─────────────┘  └─────────────┘

❌ Unsafe: Shared instance, multiple goroutines
┌─────────────┐  ┌─────────────┐
│ Goroutine 1 │──┤             │
└─────────────┘  │  Instance X │  ← RACE!
┌─────────────┐  │   (Shared)  │
│ Goroutine 2 │──┤             │
└─────────────┘  └─────────────┘
```

### Tested Scenarios

The concurrency test suite validates:

1. **Sequential Operations** - Single-threaded correctness
2. **Parallel Instances** - Multiple goroutines with separate instances
3. **Concurrent Reads** - Race-free read operations
4. **Method Calls During Reads** - Delim(), Reader() while reading
5. **Write Operations** - WriteTo and Copy under load
6. **Mixed Operations** - Read, write, and metadata calls combined
7. **Construction Stress** - Many instances created concurrently
8. **Close During Operations** - Proper cleanup handling

### Testing Commands

```bash
# Full race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrency tests
CGO_ENABLED=1 go test -race -v -run "Concurrency" ./...

# Stress test (10 iterations)
for i in {1..10}; do 
    CGO_ENABLED=1 go test -race ./... || break
done
```

**Result**: ✅ Zero data races across all scenarios

---

## Performance Benchmarks

The package includes 30 performance benchmark scenarios using **gmeasure**.

### Benchmark Categories

**1. Read Performance** (3 benchmarks)
- Small chunks (100 lines, 12 bytes each)
- Medium chunks (500 lines, 38 bytes each)
- Large chunks (100 lines, 1000 bytes each)

**2. ReadBytes Performance** (3 benchmarks)
- Small data (1000 lines, 5 bytes each)
- Medium data (500 lines, 100 bytes each)
- Large data (100 lines, 1000 bytes each)

**3. WriteTo Performance** (3 benchmarks)
- Small data streaming
- Medium data streaming
- Large data streaming (1000 lines × 1000 bytes)

**4. Buffer Size Impact** (4 benchmarks)
- Default buffer (4KB)
- Small buffer (64 bytes)
- Medium buffer (1KB)
- Large buffer (64KB)

**5. Delimiter Variations** (4 benchmarks)
- Newline (`\n`)
- Comma (`,`)
- Pipe (`|`)
- Null byte (`\0`)

**6. Method Comparison** (2 benchmarks)
- Copy method performance
- WriteTo method performance

**7. DiscardCloser** (3 benchmarks)
- Read operations
- Write operations
- Close operations

**8. Construction Overhead** (2 benchmarks)
- Default buffer construction
- Custom buffer construction

**9. UnRead Performance** (1 benchmark)
- Buffered data access

**10. Memory Patterns** (2 benchmarks)
- Read allocations
- ReadBytes allocations

**11. Real-World Scenarios** (3 benchmarks)
- CSV parsing
- Log file processing
- Variable-length streams

### Running Benchmarks

Benchmarks are integrated into the test suite:

```bash
# Run all tests including benchmarks
go test -v ./...

# Focus on benchmarks only
ginkgo --focus="Benchmarks"

# Skip benchmarks
ginkgo --skip="Benchmarks"
```

### Benchmark Results

Sample output from `benchmark_test.go`:

```
Read small chunks
Name                 | N  | Min | Median | Mean | StdDev | Max
=================================================================
read-small [duration] | 20 | 0s  | 0s     | 100µs| 0s     | 100µs

ReadBytes medium data
Name                      | N  | Min  | Median | Mean | StdDev | Max
======================================================================
readbytes-medium [duration] | 20 | 200µs | 300µs  | 300µs | 0s    | 400µs

Constructor default
Name                  | N  | Min  | Median | Mean | StdDev | Max
===================================================================
new-default [duration] | 10 | 300µs | 1.3ms | 1.3ms | 900µs | 2.8ms
```

**Key Metrics**:
- **N**: Number of samples (10-20 depending on scenario)
- **Min/Max**: Range of measurements
- **Median**: Middle value (most representative)
- **Mean**: Average performance
- **StdDev**: Variability (lower is more consistent)

---

## Test Organization

### Test File Structure

| File | Purpose | Specs | Description |
|------|---------|-------|-------------|
| `suite_test.go` | Suite initialization | 1 | Ginkgo suite setup |
| `constructor_test.go` | Constructor tests | 27 | New(), buffer sizes, delimiters |
| `read_test.go` | Read operations | 46 | Read(), ReadBytes(), UnRead() |
| `write_test.go` | Write operations | 28 | WriteTo(), Copy() |
| `edge_cases_test.go` | Edge cases | 35 | Unicode, binary, boundaries |
| `discard_test.go` | DiscardCloser | 14 | Read/Write/Close, concurrency |
| `concurrency_test.go` | Thread safety | 18 | Parallel operations, races |
| `benchmark_test.go` | Performance | 30 | gmeasure benchmarks |
| **Total** | **All tests** | **198** | **100% coverage** |

### Test Data

Tests use in-memory data generation:

```go
// Simple strings
data := "line1\nline2\nline3\n"
reader := io.NopCloser(strings.NewReader(data))

// Repeated patterns
data := strings.Repeat("test\n", 1000)

// Binary data
binaryData := []byte{0x00, 0x01, 0xFF, '\n'}
reader := io.NopCloser(bytes.NewReader(binaryData))

// Unicode
data := "hello€world€test"
reader := io.NopCloser(strings.NewReader(data))
```

**No external dependencies**: All tests use generated data, no files or network.

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should read data up to and including the delimiter", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should handle empty reader", func() {
    // Arrange
    reader := io.NopCloser(strings.NewReader(""))
    bd := delim.New(reader, '\n', 0)
    defer bd.Close()
    
    // Act
    data, err := bd.ReadBytes()
    
    // Assert
    Expect(err).To(Equal(io.EOF))
    Expect(data).To(BeEmpty())
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
Expect(string).To(HavePrefix("test"))
```

**4. Always Cleanup Resources**
```go
var bd BufferDelim

BeforeEach(func() {
    reader := io.NopCloser(strings.NewReader("test\n"))
    bd = delim.New(reader, '\n', 0)
})

AfterEach(func() {
    if bd != nil {
        bd.Close()
    }
})
```

**5. Test Edge Cases**
- Empty input
- No delimiter in data
- Only delimiter
- Very large data
- Binary data (all byte values)
- Unicode delimiters
- Buffer boundaries

**6. Avoid External Dependencies**
- No file system operations
- No network calls
- No external services
- Use `io.NopCloser(strings.NewReader(...))` for test data

### Test Template

```go
var _ = Describe("BufferDelim New Feature", func() {
    Context("When using new feature", func() {
        var (
            bd     BufferDelim
            reader io.ReadCloser
        )

        BeforeEach(func() {
            // Setup
            data := "test\ndata\n"
            reader = io.NopCloser(strings.NewReader(data))
            bd = delim.New(reader, '\n', 0)
        })

        AfterEach(func() {
            // Cleanup
            if bd != nil {
                bd.Close()
            }
        })

        It("should perform expected behavior", func() {
            // Arrange
            expected := []byte("test\n")
            
            // Act
            result, err := bd.ReadBytes()
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expected))
        })

        It("should handle error case", func() {
            bd.Close()  // Close first
            
            _, err := bd.ReadBytes()
            Expect(err).To(Equal(delim.ErrInstance))
        })
    })
})
```

### Writing Benchmarks

Use **gmeasure** for performance tests:

```go
It("should efficiently process data", func() {
    experiment := gmeasure.NewExperiment("Data Processing")
    AddReportEntry(experiment.Name, experiment)
    
    experiment.Sample(func(idx int) {
        experiment.MeasureDuration("process-time", func() {
            // Code to benchmark
            data := strings.Repeat("test\n", 1000)
            reader := io.NopCloser(strings.NewReader(data))
            bd := delim.New(reader, '\n', 0)
            defer bd.Close()
            
            for {
                _, err := bd.ReadBytes()
                if err == io.EOF {
                    break
                }
            }
        })
    }, gmeasure.SamplingConfig{N: 20, Duration: 0})
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
```go
// ✅ Good: Generate data in test
data := strings.Repeat("line\n", 100)
reader := io.NopCloser(strings.NewReader(data))

// ❌ Bad: External file dependency
file, _ := os.Open("testdata/file.txt")  // Avoid files
```

**Assertions**
```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))

// ❌ Bad: Generic comparison
Expect(value == expected).To(BeTrue())  // Less readable
```

**Resource Cleanup**
```go
// ✅ Good: Always cleanup
AfterEach(func() {
    if bd != nil {
        bd.Close()
    }
})

// ❌ Bad: No cleanup
It("test", func() {
    bd := delim.New(reader, '\n', 0)
    // Never closed!
})
```

**Error Handling in Tests**
```go
// ✅ Good: Check errors
result, err := bd.ReadBytes()
Expect(err).ToNot(HaveOccurred())
defer bd.Close()

// ❌ Bad: Ignore errors
result, _ := bd.ReadBytes()  // Don't ignore!
```

**Concurrency Testing**
```go
// ✅ Good: Proper synchronization
It("should handle concurrent instances", func() {
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Each goroutine has own instance
            reader := io.NopCloser(strings.NewReader("test\n"))
            bd := delim.New(reader, '\n', 0)
            defer bd.Close()
            
            data, _ := bd.ReadBytes()
            Expect(data).To(Equal([]byte("test\n")))
        }(i)
    }
    
    wg.Wait()
})
```

**Performance**
- Keep test data small (but representative)
- Use parallel execution with caution (can mask races)
- Target: <2s full suite with race detection
- Individual specs should be <100ms

---

## Troubleshooting

**Stale Test Cache**
```bash
# Clean and rerun
go clean -testcache
go test -v ./...
```

**Race Detector Issues**

If you see races:
```bash
# Get detailed race report
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race.log
grep -A 20 "WARNING: DATA RACE" race.log
```

Check for:
- Shared variables across goroutines
- Missing synchronization
- Concurrent access to buffers

**CGO Not Available**
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Verify
export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Identify slow tests
go test -v -timeout=30s ./...

# With Ginkgo
ginkgo --timeout=30s
```

Check for:
- Infinite loops in test code
- Missing error checks (loops never exit)
- Unclosed resources

**Debugging Specific Tests**
```bash
# Run single test
ginkgo --focus="should read data with newline delimiter"

# Specific file
ginkgo --focus-file=read_test.go

# Verbose output
ginkgo -v --trace

# Debug with Delve
dlv test ./... -- -ginkgo.focus="test name"
```

**Coverage Gaps**
```bash
# Find uncovered lines
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Function-level coverage
go tool cover -func=coverage.out | grep -v "100.0%"
```

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
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
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
    - go test -v -cover ./...
    - CGO_ENABLED=1 go test -race ./...
  coverage: '/coverage: \d+.\d+% of statements/'
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
COVERAGE=$(go test -cover ./... | grep -oE '[0-9]+\.[0-9]+%' | head -1)
echo "Coverage: $COVERAGE"

# Require 100% coverage
if [[ "$COVERAGE" != "100.0%" ]]; then
    echo "Coverage must be 100%"
    exit 1
fi

echo "All checks passed!"
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: 100%
- [ ] New features have tests
- [ ] Edge cases tested
- [ ] Error cases validated
- [ ] Benchmarks added (if performance-critical)
- [ ] Documentation updated
- [ ] Test duration reasonable (<2s with race)

---

## Quality Metrics

**Current Status** (as of latest run):

```
✅ Tests: 198/198 passing
✅ Coverage: 100.0% statement coverage
✅ Race Detection: 0 data races
✅ Performance: <0.2s without race, <2.1s with race
✅ Specs: All categories covered
✅ Benchmarks: 30 performance scenarios
```

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [gmeasure Benchmarking](https://onsi.github.io/gomega/#gmeasure-benchmarking-code)
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
**Maintained By**: IOUtils/Delim Package Contributors
