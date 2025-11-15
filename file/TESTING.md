# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-255%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-~79%25-brightgreen)]()

Comprehensive testing documentation for the file package, covering test execution, coverage analysis, and quality assurance across all subpackages.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Subpackage Testing](#subpackage-testing)
  - [bandwidth Tests](#bandwidth-tests)
  - [perm Tests](#perm-tests)
  - [progress Tests](#progress-tests)
- [Thread Safety](#thread-safety)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The file package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions across all three subpackages.

**Test Suite Summary**

| Subpackage | Specs | Coverage | Status | Test Files |
|------------|-------|----------|--------|------------|
| bandwidth | 25 | 77.8% | ✅ Pass | 5 |
| perm | 141 | 88.9% | ✅ Pass | 6 |
| progress | 89 | 71.1% | ✅ Pass | 6 |
| **Total** | **255** | **~79%** | ✅ Pass | **17** |

**Coverage Areas**
- File creation and operations
- Progress tracking with callbacks
- Bandwidth limiting and throttling
- Permission parsing and encoding
- Error handling and edge cases
- Temporary file management
- Concurrent operations
- I/O operations (read, write, seek)

**Execution Time**
- Standard: ~0.1s total
- With race detection: ~0.2s total

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
ginkgo -r
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- Rich CLI with filtering and focusing
- Parallel execution support

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Asynchronous testing support

---

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Specific subpackage
go test ./bandwidth/...
go test ./perm/...
go test ./progress/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo -r

# Verbose output
ginkgo -v -r

# With coverage
ginkgo -cover -r

# Parallel execution
ginkgo -p -r

# Focus on specific tests
ginkgo --focus="bandwidth" -r

# Focus on specific file
ginkgo --focus-file=bandwidth_test.go

# Skip tests
ginkgo --skip="slow tests" -r

# Watch mode (continuous testing)
ginkgo watch -r

# JUnit report
ginkgo --junit-report=results.xml -r
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race -r

# Verbose with race detection
CGO_ENABLED=1 go test -race -v ./...
```

**Validates**:
- Atomic operations (`atomic.Value`, `atomic.Int32`)
- Goroutine synchronization (`sync.WaitGroup`)
- Buffer thread safety
- Concurrent file access

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/file/bandwidth	0.013s
ok  	github.com/nabbar/golib/file/perm	0.013s
ok  	github.com/nabbar/golib/file/progress	0.082s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected across all packages

---

## Test Coverage

**Target**: ≥75% statement coverage across all subpackages

### Coverage Summary

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage
go test -cover ./bandwidth/...  # 77.8%
go test -cover ./perm/...       # 88.9%
go test -cover ./progress/...   # 71.1%
```

### Coverage by Category

| Category | Files | Coverage | Description |
|----------|-------|----------|-------------|
| **Bandwidth** | `interface.go`, `model.go` | 77.8% | Rate limiting, atomic operations |
| **Permission** | `interface.go`, `parse.go`, `format.go`, `encode.go` | 88.9% | Parsing, formatting, encoding |
| **Progress** | `interface.go`, `model.go`, `progress.go`, `io*.go` | 71.1% | File ops, callbacks, I/O |

### Detailed Coverage

```bash
# Detailed function-level coverage
go tool cover -func=coverage.out

# Output example:
# github.com/nabbar/golib/file/bandwidth/interface.go:67:    New             100.0%
# github.com/nabbar/golib/file/perm/parse.go:37:             parseString     100.0%
# github.com/nabbar/golib/file/progress/interface.go:207:    New             100.0%
```

---

## Subpackage Testing

### `bandwidth` Tests

**Location**: `file/bandwidth/`  
**Coverage**: 77.8%  
**Specs**: 25 (1 pending)  
**Execution Time**: ~0.005s

**Test Files**
- `bandwidth_suite_test.go` - Suite initialization
- `bandwidth_test.go` - Creation and initialization
- `increment_test.go` - Throttling behavior
- `concurrency_test.go` - Thread safety
- `edge_cases_test.go` - Edge cases and limits

**Test Categories**

```go
Describe("Bandwidth", func() {
    Context("Creation", func() {
        // - New with various byte rates
        // - Zero limit (unlimited)
        // - Very high limits
    })
    
    Context("Registration", func() {
        // - RegisterIncrement with callbacks
        // - RegisterReset with callbacks
        // - Nil callback handling
    })
    
    Context("Throttling", func() {
        // - Bandwidth enforcement
        // - Time-based throttling
        // - Sleep interval calculation
    })
    
    Context("Concurrency", func() {
        // - Concurrent registrations
        // - Multiple goroutines
        // - Atomic operations
    })
})
```

**Key Tests**
- Bandwidth limiter creation with various rates
- Progress callback integration
- Concurrent registration safety
- Throttling accuracy (pending - marked slow)
- Edge cases (zero limit, very large values)

**Running bandwidth Tests**

```bash
# All bandwidth tests
go test ./bandwidth/...

# With race detection
CGO_ENABLED=1 go test -race ./bandwidth/...

# Specific test
ginkgo --focus="Bandwidth Creation" ./bandwidth/
```

---

### `perm` Tests

**Location**: `file/perm/`  
**Coverage**: 88.9%  
**Specs**: 141  
**Execution Time**: ~0.005s

**Test Files**
- `perm_suite_test.go` - Suite initialization
- `parsing_test.go` - Parse functions
- `formatting_test.go` - Format and conversions
- `encoding_test.go` - Marshal/unmarshal operations
- `viper_test.go` - Viper integration
- `edge_cases_test.go` - Special permissions and edge cases

**Test Categories**

```go
Describe("Perm", func() {
    Context("Parsing", func() {
        // - Parse from octal strings
        // - ParseInt from decimal
        // - ParseInt64 from int64
        // - ParseByte from byte slices
        // - Quote handling
        // - Invalid input handling
    })
    
    Context("Formatting", func() {
        // - String() octal output
        // - FileMode() conversion
        // - Int/Uint conversions
        // - Overflow handling
    })
    
    Context("Encoding", func() {
        // - JSON marshal/unmarshal
        // - YAML marshal/unmarshal
        // - TOML marshal/unmarshal
        // - CBOR marshal/unmarshal
        // - Text marshal/unmarshal
        // - Round-trip tests
    })
    
    Context("Viper Integration", func() {
        // - Decoder hook registration
        // - String to Perm conversion
        // - Configuration parsing
    })
    
    Context("Edge Cases", func() {
        // - Special permissions (setuid, setgid, sticky)
        // - Overflow conditions
        // - Invalid values
        // - Boundary conditions
    })
})
```

**Key Tests**
- All parsing methods (string, int, byte)
- All encoding formats (JSON, YAML, TOML, CBOR, Text)
- Viper decoder hook functionality
- Common permissions (0644, 0755, 0600, etc.)
- Special permissions (04755, 02755, 01777)
- Overflow and error handling

**Running perm Tests**

```bash
# All perm tests
go test ./perm/...

# With coverage
go test -cover ./perm/...

# Specific category
ginkgo --focus="Encoding" ./perm/
```

---

### `progress` Tests

**Location**: `file/progress/`  
**Coverage**: 71.1%  
**Specs**: 89 (1 skipped)  
**Execution Time**: ~0.074s

**Test Files**
- `progress_suite_test.go` - Suite initialization
- `creation_test.go` - File creation functions
- `progress_callbacks_test.go` - Callback registration and invocation
- `io_operations_test.go` - I/O operations (read, write, seek)
- `file_operations_test.go` - File management operations
- `edge_cases_test.go` - Edge cases and error handling

**Test Categories**

```go
Describe("Progress", func() {
    Context("File Creation", func() {
        // - New() with flags and permissions
        // - Open() existing files
        // - Create() new files
        // - Temp() temporary files
        // - Unique() unique files
    })
    
    Context("Callback Registration", func() {
        // - RegisterFctIncrement
        // - RegisterFctReset
        // - RegisterFctEOF
        // - Nil callback handling
        // - SetRegisterProgress
    })
    
    Context("I/O Operations", func() {
        // - Read/ReadAt/ReadFrom
        // - Write/WriteAt/WriteTo
        // - Seek operations
        // - ReadByte/WriteByte
        // - WriteString
    })
    
    Context("File Operations", func() {
        // - Stat()
        // - SizeBOF()/SizeEOF()
        // - Truncate()
        // - Sync()
        // - Close()/CloseDelete()
    })
    
    Context("Progress Tracking", func() {
        // - Increment callback invocation
        // - Reset callback invocation
        // - EOF callback invocation
        // - Buffer size management
    })
    
    Context("Edge Cases", func() {
        // - Empty files
        // - Large files
        // - Concurrent access
        // - Error conditions
    })
})
```

**Key Tests**
- All file creation methods
- Callback registration and invocation
- Read/write/seek operations
- Progress tracking accuracy
- Temporary file cleanup
- Error handling

**Running progress Tests**

```bash
# All progress tests
go test ./progress/...

# With verbose output
go test -v ./progress/...

# With race detection
CGO_ENABLED=1 go test -race ./progress/...

# Specific category
ginkgo --focus="Callbacks" ./progress/
```

---

## Thread Safety

Thread safety is critical for concurrent file operations.

### Concurrency Primitives

**Atomic Operations**
```go
atomic.Value       // State storage
atomic.Int32       // Buffer size
```

**Synchronization**
- No mutexes in hot paths
- Lock-free atomic operations
- Independent instance design

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| `bandwidth.bw` | `atomic.Value` | ✅ Race-free |
| `progress.progress` | `atomic.Value` + `atomic.Int32` | ✅ Race-free |
| File operations | Independent instances | ✅ Parallel-safe |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Concurrency" ./...

# Stress test (10 iterations)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: Zero data races across all test runs

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should throttle file operations to configured bandwidth limit", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should parse octal permission string", func() {
    // Arrange
    input := "0644"
    
    // Act
    perm, err := Parse(input)
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
    Expect(perm.Uint64()).To(Equal(uint64(0644)))
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
Expect(path).To(BeAnExistingFile())
```

**4. Always Cleanup Resources**
```go
var tempFile string

AfterEach(func() {
    if tempFile != "" {
        os.Remove(tempFile)
    }
})
```

**5. Test Edge Cases**
- Empty input
- Nil values
- Large data
- Boundary conditions
- Error conditions

**6. Avoid External Dependencies**
- No remote resources
- No external services
- Use local file system only

### Test Template

```go
var _ = Describe("New Feature", func() {
    var (
        testData   []byte
        tempDir    string
    )

    BeforeEach(func() {
        testData = []byte("test data")
        tempDir, _ = os.MkdirTemp("", "test-*")
    })

    AfterEach(func() {
        if tempDir != "" {
            os.RemoveAll(tempDir)
        }
    })

    Describe("Success Cases", func() {
        It("should handle valid input", func() {
            // Arrange
            input := prepareInput(testData)
            
            // Act
            result, err := newFeature(input)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(result).ToNot(BeNil())
        })
    })

    Describe("Error Cases", func() {
        It("should handle invalid input", func() {
            // Act
            _, err := newFeature(invalidInput)
            
            // Assert
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

### Test Independence
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Avoid global mutable state
- ✅ Create test data on-demand
- ❌ Don't rely on test execution order

### Test Data
- Use temporary directories (`os.MkdirTemp`)
- Clean up in `AfterEach`
- Use `filepath.Join()` for cross-platform paths
- Generate data dynamically

### Assertions
```go
// ✅ Good
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))

// ❌ Avoid
Expect(value == expected).To(BeTrue())
if value != expected {
    Fail("values don't match")
}
```

### Concurrency Testing
```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            defer GinkgoRecover() // Important!
            // Independent operation
        }(i)
    }
    wg.Wait()
})
```

### Performance
- Keep tests fast (target: <100ms per spec)
- Use parallel execution when possible
- Avoid unnecessary sleeps
- Use minimal test data

---

## Troubleshooting

### Common Issues

**Leftover Test Files**
```bash
# Clean manually if needed
rm -f test-*
rm -rf /tmp/test-*
```

**Stale Coverage**
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Parallel Test Failures**
- Check for shared resources
- Use synchronization or make tests independent
- Verify temp directory isolation

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing atomic operations
- Unsynchronized goroutines

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: xcode-select --install

export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Increase timeout
go test -timeout=30s ./...

# Identify hanging tests
ginkgo --timeout=10s -r
```

Check for:
- Goroutine leaks (missing cleanup)
- Unclosed resources
- Deadlocks

### Debugging

```bash
# Single test
ginkgo --focus="should parse octal permission"

# Specific file
ginkgo --focus-file=parsing_test.go

# Verbose output
ginkgo -v --trace

# Debug with Delve
dlv test ./bandwidth
```

Use `GinkgoWriter` for debug output:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
```

---

## CI Integration

### GitHub Actions Example

```yaml
name: Test File Package

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Tests
        run: go test -v ./...
        working-directory: file
      
      - name: Race Detection
        run: CGO_ENABLED=1 go test -race ./...
        working-directory: file
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
        working-directory: file
      
      - name: Upload Coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage
          path: file/coverage.html
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd file
CGO_ENABLED=1 go test -race ./... || exit 1
go test -cover ./... | grep -E "coverage:" || exit 1
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥75% per subpackage
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Test duration reasonable (<1s total)
- [ ] Documentation updated

---

## Performance Benchmarks

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Specific subpackage
go test -bench=. -benchmem ./bandwidth/

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# With memory profiling
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### Expected Performance

| Operation | Time/op | Allocs/op | Package |
|-----------|---------|-----------|---------|
| Bandwidth throttling | <1μs | 0-1 | bandwidth |
| Permission parsing | ~100ns | 1 | perm |
| Permission formatting | ~50ns | 1 | perm |
| Progress tracking | ~native | 0-2 | progress |

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

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)
- [atomic Package](https://pkg.go.dev/sync/atomic)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: File Package Contributors
