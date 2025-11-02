# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-657%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-90.8%25-brightgreen)]()

Comprehensive testing documentation for the `ioutils` package and all 9 subpackages, covering test execution, thread safety validation, performance benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Subpackage Testing](#subpackage-testing)
- [Thread Safety](#thread-safety)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)
- [Resources](#resources)

---

## Overview

The `ioutils` package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive, expressive testing across all subpackages.

### Test Suite Summary

| Metric | Value | Status |
|--------|-------|--------|
| **Total Specs** | 657 | ✅ All Pass |
| **Total Coverage** | 90.8% | ✅ Excellent |
| **Race Detection** | Clean | ✅ Zero Races |
| **Execution Time** | ~720ms (~8s with -race) | ✅ Fast |
| **Subpackages** | 9 | ✅ All Tested |
| **Flaky Tests** | 0 | ✅ Stable |

### Coverage Distribution

```
Package              Specs  Coverage  Status
─────────────────────────────────────────────
ioutils              31     91.7%     ✅
bufferReadCloser     57     100.0%    ✅
delim                198    100.0%    ✅
fileDescriptor       20     85.7%     ✅
ioprogress           42     84.7%     ✅
iowrapper            114    100.0%    ✅
mapCloser            29     80.2%     ✅
maxstdio             -      N/A       ⚠️  Windows-only
multi                112    81.7%     ✅
nopwritecloser       54     100.0%    ✅
─────────────────────────────────────────────
Total                657    90.8%     ✅ Production-ready
```

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests (all subpackages)
go test ./...

# Run all tests with coverage
go test -cover ./...

# Run with race detection (requires CGO_ENABLED=1)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI (recursive)
ginkgo -r -cover

# Run specific subpackage
cd ioprogress
go test -cover ./...
```

**Expected Output**:
```
ok  	github.com/nabbar/golib/ioutils               	0.035s	coverage: 91.7% of statements
ok  	github.com/nabbar/golib/ioutils/bufferReadCloser	0.012s	coverage: 100.0% of statements
ok  	github.com/nabbar/golib/ioutils/delim          	0.170s	coverage: 100.0% of statements
ok  	github.com/nabbar/golib/ioutils/fileDescriptor  	0.012s	coverage: 85.7% of statements
ok  	github.com/nabbar/golib/ioutils/ioprogress      	0.015s	coverage: 84.7% of statements
ok  	github.com/nabbar/golib/ioutils/iowrapper       	0.050s	coverage: 100.0% of statements
ok  	github.com/nabbar/golib/ioutils/mapCloser       	0.016s	coverage: 80.2% of statements
ok  	github.com/nabbar/golib/ioutils/multi           	0.180s	coverage: 81.7% of statements
ok  	github.com/nabbar/golib/ioutils/nopwritecloser  	0.252s	coverage: 100.0% of statements
```

---

## Test Framework

### Ginkgo v2

[Ginkgo](https://onsi.github.io/ginkgo/) is a modern BDD-style testing framework for Go.

**Key Features**:
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Expressive test specifications with clear intent
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Rich CLI with filtering, parallel execution, and detailed reporting
- Excellent failure diagnostics with stack traces
- Built-in benchmarking support

**Why Ginkgo?**
- Tests read like specifications
- Better organization than standard `testing` package
- Consistent across all `golib` packages
- Active development and community support
- Excellent IDE integration

### Gomega

[Gomega](https://onsi.github.io/gomega/) is Ginkgo's matcher library.

**Key Features**:
- Readable, expressive assertion syntax
- 50+ built-in matchers for common scenarios
- Support for custom matchers
- Detailed failure messages with context
- Async assertion support

**Example Matchers**:
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
Expect(path).To(BeAnExistingFile())
```

---

## Running Tests

### Prerequisites

**Requirements**:
- Go 1.18 or higher
- Properly configured `GOPATH` and `GOROOT`
- For race detection: CGO enabled (Linux, macOS, Windows with MinGW)

**Install Ginkgo CLI** (optional but recommended):
```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

### Basic Test Execution

**Using Go test**:
```bash
# Navigate to ioutils package directory
cd /path/to/golib/ioutils

# Run all tests recursively
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Using Ginkgo CLI**:
```bash
# Run all tests recursively
ginkgo -r

# Verbose output
ginkgo -r -v

# With coverage
ginkgo -r -cover

# Parallel execution
ginkgo -r -p

# Watch mode (re-run on file changes)
ginkgo watch -r
```

### Advanced Test Options

**Pattern Matching**:
```bash
# Run specific package tests
ginkgo --focus-file=ioprogress

# Run tests matching pattern
ginkgo -r --focus="Progress tracking"

# Skip specific tests
ginkgo -r --skip="Windows-specific"

# Multiple filters
ginkgo -r --focus="Reader" --skip="EOF"
```

**Output Formats**:
```bash
# JSON output
ginkgo -r --json-report=results.json

# JUnit XML (for CI)
ginkgo -r --junit-report=results.xml

# Custom output directory
ginkgo -r --output-dir=./test-results
```

**Performance & Debugging**:
```bash
# Show execution time per spec
ginkgo -r -v --show-node-events

# Trace mode (detailed execution flow)
ginkgo -r --trace

# Fail fast (stop on first failure)
ginkgo -r --fail-fast

# Randomize test order
ginkgo -r --randomize-all --seed=12345
```

### Race Detection

**Critical for concurrent code testing**:

```bash
# Enable race detector (requires CGO_ENABLED=1)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -r -race

# Verbose race detection
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
```

**What It Validates**:
- Atomic operations across all subpackages
- Concurrent callback registration (ioprogress, iowrapper)
- Shared state access patterns
- Goroutine synchronization (mapCloser)
- Buffer access (bufferReadCloser)

**Expected Output**:
```
✅ Success:
ok  	github.com/nabbar/golib/ioutils/...	0.025s

❌ Race Detected:
WARNING: DATA RACE
Read at 0x... by goroutine ...
Write at 0x... by goroutine ...
```

**Current Status**: Zero data races detected across all subpackages ✅

---

## Test Coverage

### Overall Coverage

**Target**: ≥90% statement coverage across all subpackages

### Coverage By Subpackage

| Subpackage | Files | Specs | Coverage | Uncovered | Status |
|------------|-------|-------|----------|-----------|--------|
| **ioutils** | 1 | 31 | 91.7% | Permission edge cases | ✅ |
| **bufferReadCloser** | 4 | 57 | 100% | None | ✅ |
| **delim** | 7 | 198 | 100% | None | ✅ |
| **fileDescriptor** | 2 | 20 | 85.7% | Platform-specific paths | ✅ |
| **ioprogress** | 3 | 42 | 84.7% | Writer EOF (rare) | ✅ |
| **iowrapper** | 1 | 114 | 100% | None | ✅ |
| **mapCloser** | 1 | 29 | 80.2% | Context cancellation timing | ✅ |
| **maxstdio** | 1 | - | N/A | Windows-only (cgo) | ⚠️ |
| **multi** | 5 | 112 | 81.7% | Error edge cases | ✅ |
| **nopwritecloser** | 1 | 54 | 100% | None | ✅ |

### Detailed Coverage

**Root Package** (ioutils):
```
File: tools.go
Coverage: 91.7%
Functions: 1/1 (PathCheckCreate)
Uncovered: Edge cases in permission validation
```

**bufferReadCloser**:
```
Files: buffer.go, reader.go, writer.go, readwriter.go
Coverage: 100%
All buffer types fully tested
Zero uncovered lines
```

**delim**:
```
Files: 7 implementation files + comprehensive test suites
Coverage: 100%
All delimiter handling, buffering, and I/O operations tested
Benchmarks included for performance validation
198 specs covering edge cases, concurrency, and Unicode
```

**fileDescriptor**:
```
Files: fileDescriptor.go, fileDescriptor_windows.go
Coverage: 90.0%
Platform-specific code paths (Windows vs Unix)
```

**ioprogress**:
```
Files: interface.go, reader.go, writer.go
Coverage: 84.7%
Writer finish() method rarely triggered (EOF on write)
```

**iowrapper**:
```
File: interface.go
Coverage: 100%
All read/write/seek/close paths tested
Benchmark suite included
```

**mapCloser**:
```
File: interface.go
Coverage: 80.2%
Context cancellation timing edge cases
```

**multi**:
```
Files: 5 implementation files
Coverage: 81.7%
Thread-safe I/O multiplexing with atomic operations
Concurrent write broadcasting tested
Benchmarks with gmeasure for performance validation
112 specs including edge cases and stress tests
```

**nopwritecloser**:
```
File: interface.go
Coverage: 100%
Simple implementation, complete coverage
```

### Viewing Coverage Reports

**Terminal Report**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**HTML Report** (recommended):
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

**Per-Package Coverage**:
```bash
# Individual subpackage
cd ioprogress
go test -coverprofile=coverage.out
go tool cover -func=coverage.out

# All subpackages with summary
go test -cover ./... | grep coverage
```

---

## Subpackage Testing

### ioutils (Root Package)

**Focus**: Path creation and permission management

**Test Categories**:
- File creation with permissions
- Directory creation with permissions
- Permission validation and updates
- Path existence checks
- Error handling (invalid paths, permission denied)

**Run Tests**:
```bash
go test -v -cover
```

**Example Test**:
```go
It("should create directory with correct permissions", func() {
    err := PathCheckCreate(false, "/tmp/test-dir", 0644, 0755)
    Expect(err).ToNot(HaveOccurred())
    
    info, _ := os.Stat("/tmp/test-dir")
    Expect(info.IsDir()).To(BeTrue())
    Expect(info.Mode().Perm()).To(Equal(os.FileMode(0755)))
})
```

---

### bufferReadCloser

**Focus**: Buffered I/O with resource lifecycle

**Test Categories**:
- Buffer creation and initialization
- Read/write operations
- Custom close functions
- Flush behavior on close
- Multiple close safety

**Run Tests**:
```bash
cd bufferReadCloser
go test -v -cover
```

**Key Tests**:
- All buffer types (Buffer, Reader, Writer, ReadWriter)
- Custom close function invocation
- Resource cleanup verification

---

### delim

**Focus**: Delimiter-based buffered reading

**Test Categories**:
- Constructor with various delimiters and buffer sizes (30+ specs)
- Read operations (Read, ReadBytes, UnRead) (60+ specs)
- Write operations (WriteTo, Copy) (20+ specs)
- Edge cases (Unicode, binary data, empty input, large data) (40+ specs)
- DiscardCloser functionality (10+ specs)
- Performance benchmarks (30 scenarios)

**Run Tests**:
```bash
cd delim
go test -v -cover
```

**Key Tests**:
- Custom delimiter support (newlines, commas, tabs, null bytes, Unicode)
- Constant memory usage validation
- Zero-copy operations
- Concurrent access patterns
- Performance benchmarks for various buffer sizes

**Detailed Testing Guide**: See [delim/TESTING.md](delim/TESTING.md)

---

### fileDescriptor

**Focus**: File descriptor limit management

**Test Categories**:
- Current limit queries
- Maximum limit queries
- Limit increase operations
- Platform-specific behavior
- Error handling

**Run Tests**:
```bash
cd fileDescriptor
go test -v -cover
```

**Platform Notes**:
- Linux/Unix: Uses `syscall.Getrlimit/Setrlimit`
- Windows: Uses maxstdio cgo calls
- Tests adapt to platform capabilities

---

### ioprogress

**Focus**: I/O progress tracking

**Test Categories**:
- Reader progress tracking (22 specs)
- Writer progress tracking (20 specs)
- Callback registration and invocation
- EOF detection
- Reset functionality
- Thread-safe counter updates
- Nil callback handling

**Run Tests**:
```bash
cd ioprogress
go test -v -cover
```

**Detailed Testing Guide**: See [ioprogress/TESTING.md](ioprogress/TESTING.md)

---

### iowrapper

**Focus**: I/O operation interception

**Test Categories**:
- Read function wrapping (28 specs)
- Write function wrapping (28 specs)
- Seek function wrapping (20 specs)
- Close function wrapping (18 specs)
- Function updates (10 specs)
- Nil function handling (10 specs)
- Benchmarks (10 specs)

**Run Tests**:
```bash
cd iowrapper
go test -v -cover
```

**Benchmark Results**:
```
Read function:   ~100ns/op
Write function:  ~50ns/op
Seek function:   ~50ns/op
Function update: ~200ns/op
```

**Detailed Testing Guide**: See [iowrapper/TESTING.md](iowrapper/TESTING.md)

---

### mapCloser

**Focus**: Multiple resource management

**Test Categories**:
- Resource addition
- Batch closing
- Error aggregation
- Context cancellation
- Clone functionality
- Thread safety

**Run Tests**:
```bash
cd mapCloser
go test -v -cover
```

**Key Tests**:
- Multiple closer management
- Context-aware cleanup
- Error collection from failed closes

---

### multi

**Focus**: Thread-safe I/O multiplexing

**Test Categories**:
- Constructor and interface compliance (10+ specs)
- Write operations (single, multiple, large data) (30+ specs)
- Read operations and error propagation (15+ specs)
- Copy operations and integration (15+ specs)
- Concurrent operations (writes, AddWriter, Clean, SetInput) (25+ specs)
- Edge cases (nil values, zero-length, state transitions) (15+ specs)
- Performance benchmarks with gmeasure (10 scenarios)

**Run Tests**:
```bash
cd multi
go test -v -cover
```

**Key Tests**:
- Thread-safe broadcast writes to multiple destinations
- Dynamic writer management (add/remove on-the-fly)
- Atomic operations with sync.Map
- Zero allocations in write path
- Concurrent stress tests

**Detailed Testing Guide**: See [multi/TESTING.md](multi/TESTING.md)

---

### nopwritecloser

**Focus**: No-op write closer for testing

**Test Categories**:
- Writer wrapping (18 specs)
- Close behavior (18 specs)
- Multiple implementations (18 specs)

**Run Tests**:
```bash
cd nopwritecloser
go test -v -cover
```

**Detailed Testing Guide**: See [nopwritecloser/TESTING.md](nopwritecloser/TESTING.md)

---

## Thread Safety

Thread safety is critical across all subpackages with concurrent operations.

### Concurrency Validation

**Packages with Thread Safety Requirements**:
- `ioprogress`: Atomic counters and callback storage
- `iowrapper`: Concurrent function updates
- `mapCloser`: Concurrent resource addition/removal
- `bufferReadCloser`: Safe buffer access

### Race Detection Results

**Test Command**:
```bash
CGO_ENABLED=1 go test -race ./...
```

**Results by Subpackage**:

| Subpackage | Race Detection | Concurrency Features |
|------------|----------------|---------------------|
| `ioutils` | ✅ Clean | File system operations |
| `bufferReadCloser` | ✅ Clean | Buffer access |
| `fileDescriptor` | ✅ Clean | System call synchronization |
| `ioprogress` | ✅ Clean | Atomic counters, callback storage |
| `iowrapper` | ✅ Clean | Function pointer updates |
| `mapCloser` | ✅ Clean | Map access, context handling |
| `nopwritecloser` | ✅ Clean | No shared state |

**Validated Scenarios**:
- Concurrent callback registration during I/O
- Multiple goroutines accessing wrappers
- Context cancellation during resource cleanup
- Parallel function updates
- Simultaneous close operations

### Thread-Safe Usage Examples

**ioprogress**:
```go
reader := ioprogress.NewReadCloser(file)
var totalBytes int64

// Goroutine 1: Register callbacks
go func() {
    reader.RegisterFctIncrement(func(size int64) {
        atomic.AddInt64(&totalBytes, size)  // Thread-safe
    })
}()

// Goroutine 2: Perform I/O
go func() {
    io.Copy(dest, reader)
}()
```

**mapCloser**:
```go
closer := mapCloser.New(ctx)

// Goroutine 1: Add resources
go func() {
    file, _ := os.Open("file1.txt")
    closer.Add(file)
}()

// Goroutine 2: Add more resources
go func() {
    file, _ := os.Open("file2.txt")
    closer.Add(file)
}()

// Safe concurrent addition
```

---

## Best Practices

### Test Organization

**Use Descriptive Test Names**:
```go
// ✅ Good: Clear, specific description
It("should create file with 0644 permissions when path does not exist", func() {
    // Test implementation
})

// ❌ Bad: Vague description
It("should work", func() {
    // Test implementation
})
```

**Organize with Context Blocks**:
```go
Describe("PathCheckCreate", func() {
    Context("when creating a file", func() {
        It("should create file with correct permissions", func() {})
        It("should create parent directories if needed", func() {})
    })
    
    Context("when creating a directory", func() {
        It("should create directory with correct permissions", func() {})
        It("should update permissions if directory exists", func() {})
    })
})
```

### Test Independence

**Each Test Should Be Independent**:
```go
// ✅ Good: Independent tests with cleanup
var _ = Describe("FileOperations", func() {
    var tempDir string
    
    BeforeEach(func() {
        tempDir, _ = os.MkdirTemp("", "test-*")
    })
    
    AfterEach(func() {
        os.RemoveAll(tempDir)
    })
    
    It("test 1", func() {
        // Uses tempDir, cleaned up after
    })
})

// ❌ Bad: Shared state between tests
var sharedFile *os.File  // DON'T DO THIS!

It("test 1", func() {
    sharedFile, _ = os.Create("file.txt")
})

It("test 2", func() {
    sharedFile.Write([]byte("data"))  // Depends on test 1!
})
```

### Atomic Operations

**Always Use Atomic Operations for Shared Counters**:
```go
// ✅ Good: Thread-safe
var totalBytes int64
reader.RegisterFctIncrement(func(size int64) {
    atomic.AddInt64(&totalBytes, size)
})
value := atomic.LoadInt64(&totalBytes)

// ❌ Bad: Race condition
var totalBytes int64
reader.RegisterFctIncrement(func(size int64) {
    totalBytes += size  // NOT thread-safe!
})
```

### Resource Cleanup

**Always Clean Up Resources**:
```go
// ✅ Good: Proper cleanup
file, err := os.Create("/tmp/test.txt")
Expect(err).ToNot(HaveOccurred())
defer os.Remove("/tmp/test.txt")
defer file.Close()

// ❌ Bad: No cleanup (leaks temp files)
file, _ := os.Create("/tmp/test.txt")
// File and temp file never cleaned up
```

### Error Handling

**Test Error Cases Explicitly**:
```go
// ✅ Good: Test both success and failure
It("should return error for invalid path", func() {
    err := PathCheckCreate(true, "/invalid/\x00/path", 0644, 0755)
    Expect(err).To(HaveOccurred())
})

It("should succeed for valid path", func() {
    err := PathCheckCreate(true, "/tmp/valid.txt", 0644, 0755)
    Expect(err).ToNot(HaveOccurred())
})

// ❌ Bad: Only test success path
It("should work", func() {
    err := PathCheckCreate(true, "/tmp/file.txt", 0644, 0755)
    Expect(err).ToNot(HaveOccurred())
    // No error case testing
})
```

---

## Troubleshooting

### Common Issues

**Permission Denied Errors**

*Problem*: Tests fail with permission errors when creating files/directories.

*Solution*:
```bash
# Ensure test directory is writable
chmod 755 /tmp

# Run tests in user-writable location
export TMPDIR=$HOME/tmp
mkdir -p $TMPDIR
go test ./...
```

**Race Condition Detection**

*Problem*: Tests pass normally but fail with `-race` flag.

*Solution*:
```go
// ❌ Wrong: Direct variable access
var count int64
callback := func(size int64) {
    count += size  // Race!
}

// ✅ Correct: Atomic operations
var count int64
callback := func(size int64) {
    atomic.AddInt64(&count, size)  // Safe
}
```

**CGO Not Available for Race Detection**

*Problem*: `go test -race` fails with "cgo: C compiler not found".

*Solution*:
```bash
# Linux (Debian/Ubuntu)
sudo apt-get install build-essential

# Linux (RHEL/CentOS)
sudo yum groupinstall "Development Tools"

# macOS
xcode-select --install

# Then run
CGO_ENABLED=1 go test -race ./...
```

**Stale Test Cache**

*Problem*: Tests not reflecting recent changes.

*Solution*:
```bash
# Clean test cache
go clean -testcache

# Force re-run
go test -count=1 ./...
```

**Platform-Specific Test Failures**

*Problem*: Tests fail on specific OS (Windows vs Linux).

*Solution*:
```go
// Use build tags for platform-specific tests
//go:build !windows
// +build !windows

package ioutils_test

// Unix-specific tests...
```

### Debugging Tests

**Run Specific Tests**:
```bash
# By test name
ginkgo --focus="should create file"

# By package
cd ioprogress
go test -v

# By file
ginkgo --focus-file=tools_test.go
```

**Verbose Output**:
```bash
# Show all specs
ginkgo -v

# Show execution flow
ginkgo --trace

# Show timing
ginkgo -v --show-node-events
```

**Debug Logging**:
```go
It("should do something", func() {
    // Debug output (only shown on failure)
    fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
    
    // Test code
    result := doSomething()
    
    fmt.Fprintf(GinkgoWriter, "Result: %v\n", result)
})
```

**Fail Fast**:
```bash
# Stop on first failure
ginkgo -r --fail-fast
```

**Randomization**:
```bash
# Run tests in random order (finds order dependencies)
ginkgo -r --randomize-all

# Use specific seed for reproducibility
ginkgo -r --randomize-all --seed=12345
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Test IOUtils

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: ['1.18', '1.19', '1.20', '1.21']
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Install Ginkgo
        run: go install github.com/onsi/ginkgo/v2/ginkgo@latest
      
      - name: Run Tests
        run: |
          cd ioutils
          go test -v ./...
      
      - name: Run Race Detector
        run: |
          cd ioutils
          CGO_ENABLED=1 go test -race ./...
        if: runner.os != 'Windows'  # Skip race on Windows if no compiler
      
      - name: Generate Coverage
        run: |
          cd ioutils
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./ioutils/coverage.out
```

### GitLab CI Example

```yaml
test:
  image: golang:1.21
  stage: test
  script:
    - cd ioutils
    - go test -v ./...
    - CGO_ENABLED=1 go test -race ./...
    - go test -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  coverage: '/coverage: \d+.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: ioutils/coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running ioutils tests..."
cd ioutils || exit 1

# Run tests
if ! go test ./...; then
    echo "❌ Tests failed"
    exit 1
fi

# Run race detector
if ! CGO_ENABLED=1 go test -race ./...; then
    echo "❌ Race detector found issues"
    exit 1
fi

# Check coverage
COVERAGE=$(go test -cover ./... | grep -oP '\d+\.\d+(?=%)')
if (( $(echo "$COVERAGE < 90" | bc -l) )); then
    echo "❌ Coverage below 90% ($COVERAGE%)"
    exit 1
fi

echo "✅ All checks passed"
exit 0
```

---

## Quality Checklist

Before submitting code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained or improved (≥93%)
- [ ] New features have corresponding tests
- [ ] Edge cases tested (nil values, empty data, errors)
- [ ] Thread safety validated (where applicable)
- [ ] Tests are independent (no shared state)
- [ ] Platform-specific tests properly tagged
- [ ] Documentation updated (README.md, TESTING.md)
- [ ] GoDoc comments added for public APIs

---

## Resources

**Testing Frameworks**
- [Ginkgo v2 Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Blog: Table Driven Tests](https://go.dev/blog/table-driven-tests)

**Concurrency & Thread Safety**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync/atomic Package](https://pkg.go.dev/sync/atomic)
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)

**Coverage & Profiling**
- [Go Coverage](https://go.dev/blog/cover)
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

**Best Practices**
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

---

## License & AI Transparency

**License**: MIT License © Nicolas JUHEL

**AI Transparency Notice**: In accordance with Article 50.4 of the EU AI Act, this testing suite and documentation utilized AI assistance for test development, documentation, and bug fixing under human supervision. AI was **not** used for core package implementation.

**Contributing**: Contributors should **not use AI** to generate package implementation code but may use AI assistance for tests, documentation, and bug fixes.

---

## Summary

The `ioutils` package test suite provides comprehensive validation across all subpackages:

- **657 Specs**: Covering all public APIs and edge cases across 9 subpackages
- **90.8% Coverage**: Production-ready quality with extensive edge case testing
- **Zero Race Conditions**: Validated across all subpackages with `-race` detector
- **Fast Execution**: ~720ms total runtime (~8s with race detector)
- **BDD Style**: Clear, readable test specifications using Ginkgo v2 and Gomega
- **Cross-Platform**: Linux, macOS, Windows support

**Test Execution**:
```bash
# Quick test
go test ./...

# Full validation
CGO_ENABLED=1 go test -race -cover ./...

# Per-subpackage
cd <subpackage>
go test -v -cover
```

**Subpackage Test Guides**:
- [delim/TESTING.md](delim/TESTING.md) - Delimiter-based buffering tests (198 specs, 100% coverage)
- [multi/TESTING.md](multi/TESTING.md) - I/O multiplexing tests (112 specs, 81.7% coverage)
- [ioprogress/TESTING.md](ioprogress/TESTING.md) - Progress tracking tests (42 specs, 84.7% coverage)
- [iowrapper/TESTING.md](iowrapper/TESTING.md) - I/O wrapper tests (114 specs, 100% coverage)
- [nopwritecloser/TESTING.md](nopwritecloser/TESTING.md) - No-op closer tests (54 specs, 100% coverage)

For questions or issues, visit the [GitHub repository](https://github.com/nabbar/golib).
