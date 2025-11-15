# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-9183%2B%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A5%2080%25-brightgreen)]()

Comprehensive testing documentation for the golib library, covering test execution, coverage analysis, race detection, and best practices across all packages.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Package Test Statistics](#package-test-statistics)
- [Thread Safety](#thread-safety)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)
- [Resources](#resources)

---

## Overview

The golib library uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing across all packages. Each package includes thorough test coverage with a focus on thread safety, performance, and edge cases.

**Repository-Wide Statistics**
- **Total Packages**: 38+
- **Total Test Specs**: 9,183+
- **Average Coverage**: >80%
- **Race Detection**: ✅ Zero data races
- **Test Duration**: ~2-3 minutes (standard), ~5-6 minutes (with race detection)
- **Go Version**: 1.22+

**Testing Philosophy**
1. **Comprehensive**: Every feature has corresponding tests
2. **Thread-Safe**: All concurrent operations validated with `-race`
3. **Independent**: Tests run in isolation without shared state
4. **BDD Style**: Readable, descriptive test specifications
5. **CI-Ready**: Automated testing in GitHub Actions workflows

---

## Quick Start

```bash
# Install Ginkgo CLI (optional but recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -r

# Ginkgo with coverage and race detection
CGO_ENABLED=1 ginkgo -r -cover -race
```

---

## Test Framework

### Ginkgo v2

**BDD Testing Framework** ([Documentation](https://onsi.github.io/ginkgo/))

- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering and focus
- JUnit XML report generation

### Gomega

**Matcher Library** ([Documentation](https://onsi.github.io/gomega/))

- Readable assertion syntax
- Extensive built-in matchers
- Detailed failure messages
- Async/Eventually support
- Custom matcher creation

### Test Structure Example

```go
var _ = Describe("Component", func() {
    var (
        subject ComponentType
        config  Config
    )

    BeforeEach(func() {
        config = DefaultConfig()
        subject = NewComponent(config)
    })

    AfterEach(func() {
        subject.Close()
    })

    Context("When performing operation", func() {
        It("should succeed with valid input", func() {
            result, err := subject.Operation(validInput)
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(Equal(expectedValue))
        })

        It("should handle error cases", func() {
            _, err := subject.Operation(invalidInput)
            Expect(err).To(HaveOccurred())
        })

        It("should be thread-safe", func() {
            var wg sync.WaitGroup
            for i := 0; i < 100; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    subject.Operation(validInput)
                }()
            }
            wg.Wait()
        })
    })
})
```

---

## Running Tests

### Standard Go Test

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./logger/...

# With timeout
go test -timeout=10m ./...

# Coverage
go test -cover ./...

# Coverage with atomic mode
go test -covermode=atomic -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Ginkgo CLI

```bash
# Run all tests recursively
ginkgo -r

# Verbose output
ginkgo -r -v

# Specific package
ginkgo ./logger

# Parallel execution
ginkgo -r -p

# With coverage
ginkgo -r -cover

# Focus on specific test
ginkgo -r --focus="should handle errors"

# Skip specific tests
ginkgo -r --skip="integration"

# JUnit XML report
ginkgo -r --junit-report=results.xml

# Progress reporting
ginkgo -r --progress
```

### Race Detection

**Critical for concurrent code validation**

```bash
# With go test
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -r -race

# With timeout for long-running tests
CGO_ENABLED=1 go test -race -timeout=10m ./...
```

**Requirements**:
- CGO must be enabled
- Build tools installed (gcc/clang)
- Adds ~2-3x execution time overhead
- Detects data races at runtime

**Expected Output**:
```bash
# ✅ Success (no races)
ok  	github.com/nabbar/golib/atomic	2.345s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

---

## Test Coverage

### Coverage by Category

| Category | Avg Coverage | Package Count | Notes |
|----------|--------------|---------------|-------|
| **Utilities** | >90% | 9 | High coverage (atomic, size, version, errors) |
| **Networking** | 70-98% | 6 | Network/socket tests |
| **Monitoring** | 85-100% | 4 | Logger, monitor, prometheus, status |
| **Data Management** | 70-80% | 4 | Database, cache, archive, config |
| **Security** | 80-100% | 4 | Certificates, password, LDAP |
| **Concurrency** | 95-100% | 3 | Semaphore, runner, atomic |

### High Coverage Packages (≥90%)

- **version**: 93.8% - Semantic versioning
- **size**: 95.4% - Byte size operations
- **atomic**: >95% - Atomic primitives
- **errors**: >90% - Error handling
- **prometheus/bloom**: 94.7% - Bloom filters
- **prometheus/metrics**: 95.5% - Metrics collection
- **router**: 91.4% - HTTP routing
- **network/protocol**: 98.7% - Protocol handling
- **semaphore**: 98%+ - Concurrency control

### Viewing Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (macOS)
open coverage.html

# Coverage by package
go test -cover ./... | grep coverage

# Detailed coverage with line numbers
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | sort -k3 -t: -rn
```

---

## Package Test Statistics

### Complete Package Breakdown

| Package | Specs | Coverage | Key Features Tested |
|---------|-------|----------|---------------------|
| **version** | 173 | 93.8% | Parsing, comparison, constraints, JSON/YAML |
| **viper** | 104 | 73.3% | Configuration, cleaners, loaders |
| **size** | 150+ | 95.4% | Parsing, formatting, arithmetic, conversions |
| **atomic** | 100+ | >95% | Value operations, maps, concurrency |
| **errors** | 200+ | >90% | Codes, tracing, hierarchy, pools |
| **semaphore/bar** | 80+ | 96.6% | Progress bars, concurrency |
| **semaphore/sem** | 60+ | 100% | Semaphore operations |
| **router/auth** | 40+ | 96.3% | Authentication middleware |
| **router/authheader** | 30+ | 100% | Header-based auth |
| **monitor/info** | 60+ | 100% | System information |
| **monitor/status** | 50+ | 98.4% | Status monitoring |
| **prometheus/bloom** | 40+ | 94.7% | Bloom filter operations |
| **prometheus/metrics** | 60+ | 95.5% | Metric collection |
| **logger/gorm** | 20+ | 100% | GORM integration |
| **logger/hookstderr** | 15+ | 100% | Stderr hook |
| **logger/hookstdout** | 15+ | 100% | Stdout hook |
| **runner/startStop** | 50+ | 88.8% | Start/stop lifecycle |
| **runner/ticker** | 45+ | 90.2% | Ticker operations |
| **status/control** | 40+ | 95.0% | Control operations |
| **retro** | 60+ | 84.2% | Compatibility utilities |
| **password** | 50+ | 84.6% | Password generation |
| **network/protocol** | 40+ | 98.7% | Protocol handling |
| **console** | 182 | 60.9% | Terminal formatting (limited by prompts) |

### Test Execution Times

| Test Type | Duration | Notes |
|-----------|----------|-------|
| **Full Suite** | ~2-3 min | Without race detection |
| **With Race** | ~5-6 min | Recommended for pre-commit |
| **Single Package** | <10s | Most packages |
| **Parallel (-p)** | ~1-2 min | Faster execution |
| **Integration** | Varies | Depends on external services |

---

## Thread Safety

Thread safety is validated across all concurrent operations using Go's race detector.

### Validated Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| **atomic.Value** | `sync/atomic` | ✅ Race-free |
| **atomic.Map** | `sync.Map` | ✅ Race-free |
| **helper.bufNoEOF** | `sync.Mutex` + `atomic.Bool` | ✅ Race-free |
| **archive/helper** | Goroutine sync | ✅ Parallel-safe |
| **logger hooks** | Concurrent writes | ✅ Thread-safe |
| **prometheus metrics** | Atomic counters | ✅ Lock-free |
| **semaphore** | Channel-based | ✅ Race-free |
| **cache items** | Mutex protection | ✅ Thread-safe |

### Race Detection Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Specific package focus
CGO_ENABLED=1 go test -race -v ./atomic/...

# With coverage
CGO_ENABLED=1 go test -race -cover ./...

# Stress test (run multiple times)
for i in {1..10}; do
    CGO_ENABLED=1 go test -race ./... || break
done
```

### Race Detector Setup

**Ubuntu/Debian**:
```bash
sudo apt-get install build-essential
```

**macOS**:
```bash
xcode-select --install
# or
brew install gcc
```

**Windows**:
Install MinGW-w64 or TDM-GCC

---

## Writing Tests

### Test File Naming

- Test files: `*_test.go`
- Suite files: `*_suite_test.go`
- Place in same package or `package_test` for external tests

### Best Practices

#### 1. Descriptive Test Names

```go
// ✅ Good
It("should parse semantic version with prerelease", func() { ... })
It("should handle concurrent map operations", func() { ... })
It("should return error for invalid input", func() { ... })

// ❌ Bad
It("test1", func() { ... })
It("works", func() { ... })
```

#### 2. AAA Pattern (Arrange, Act, Assert)

```go
It("should compress and decompress data", func() {
    // Arrange
    input := []byte("test data")
    var compressed bytes.Buffer
    
    // Act
    compressor, err := helper.NewWriter(compress.Gzip, helper.Compress, &compressed)
    Expect(err).ToNot(HaveOccurred())
    _, err = compressor.Write(input)
    Expect(err).ToNot(HaveOccurred())
    compressor.Close()
    
    // Assert
    Expect(compressed.Len()).To(BeNumerically(">", 0))
})
```

#### 3. Proper Cleanup

```go
var _ = Describe("Component", func() {
    var (
        tempFile string
        client   *Client
    )

    BeforeEach(func() {
        f, _ := os.CreateTemp("", "test")
        tempFile = f.Name()
        f.Close()
        
        client = NewClient()
    })

    AfterEach(func() {
        if tempFile != "" {
            os.Remove(tempFile)
        }
        if client != nil {
            client.Close()
        }
    })
})
```

#### 4. Test Independence

```go
// ✅ Good - Each test is independent
It("test A", func() {
    data := createTestData()
    result := process(data)
    Expect(result).To(BeValid())
})

It("test B", func() {
    data := createTestData()  // Fresh data
    result := process(data)
    Expect(result).To(BeValid())
})

// ❌ Bad - Tests share state
var sharedData []byte

It("test A", func() {
    sharedData = createTestData()
    // ...
})

It("test B", func() {
    // Depends on test A running first
    process(sharedData)
})
```

#### 5. Edge Case Testing

```go
Context("Edge cases", func() {
    It("should handle nil input", func() {
        _, err := Process(nil)
        Expect(err).To(HaveOccurred())
    })

    It("should handle empty input", func() {
        result, err := Process([]byte{})
        Expect(err).ToNot(HaveOccurred())
        Expect(result).To(BeEmpty())
    })

    It("should handle very large input", func() {
        largeInput := make([]byte, 10*1024*1024) // 10MB
        _, err := Process(largeInput)
        Expect(err).ToNot(HaveOccurred())
    })

    It("should handle concurrent access", func() {
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                Process(testData)
            }()
        }
        wg.Wait()
    })
})
```

#### 6. Use Appropriate Matchers

```go
// Equality
Expect(value).To(Equal(expected))
Expect(value).To(BeEquivalentTo(expected))

// Error checking
Expect(err).ToNot(HaveOccurred())
Expect(err).To(HaveOccurred())
Expect(err).To(MatchError("specific error"))

// Numeric
Expect(count).To(BeNumerically(">", 0))
Expect(value).To(BeNumerically("~", 3.14, 0.01))

// Collections
Expect(slice).To(ContainElement(item))
Expect(slice).To(HaveLen(5))
Expect(slice).To(BeEmpty())
Expect(slice).To(ConsistOf(expected...))

// Strings
Expect(str).To(ContainSubstring("substring"))
Expect(str).To(MatchRegexp(`\d+`))
Expect(str).To(HavePrefix("prefix"))

// Booleans
Expect(condition).To(BeTrue())
Expect(condition).To(BeFalse())

// Nil checking
Expect(ptr).To(BeNil())
Expect(ptr).ToNot(BeNil())

// Eventually (async)
Eventually(func() bool {
    return condition()
}, "5s", "100ms").Should(BeTrue())
```

---

## Best Practices

### Test Organization

**Do**:
- Group related tests in `Context` blocks
- Use descriptive `Describe` names
- Keep test files focused (one component per file)
- Use `BeforeEach`/`AfterEach` for setup/cleanup
- Test public interfaces, not implementation details

**Don't**:
- Mix unrelated test cases
- Share mutable state between tests
- Rely on test execution order
- Test private methods directly
- Leave resources uncleaned

### Performance Testing

```go
// Benchmark example
func BenchmarkOperation(b *testing.B) {
    data := generateTestData()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        Operation(data)
    }
}

// Run benchmarks
go test -bench=. -benchmem ./...
```

### Table-Driven Tests

```go
var _ = Describe("Size parsing", func() {
    DescribeTable("should parse various formats",
        func(input string, expected Size, shouldError bool) {
            result, err := ParseSize(input)
            if shouldError {
                Expect(err).To(HaveOccurred())
            } else {
                Expect(err).ToNot(HaveOccurred())
                Expect(result).To(Equal(expected))
            }
        },
        Entry("bytes", "100", Size(100), false),
        Entry("kilobytes", "5KB", Size(5*1024), false),
        Entry("megabytes", "10MB", Size(10*1024*1024), false),
        Entry("invalid", "invalid", Size(0), true),
    )
})
```

### Debugging Tests

```go
// Enable verbose output
ginkgo -v

// Focus on specific test
ginkgo --focus="should handle errors"

// Skip tests
ginkgo --skip="integration"

// Debug output in tests
It("should work", func() {
    fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
    result := Operation(input)
    Expect(result).To(BeValid())
})
```

---

## Troubleshooting

### Common Issues

#### Race Condition Detected

```bash
WARNING: DATA RACE
Read at 0x... by goroutine 15
  github.com/nabbar/golib/package.Function()
```

**Solution**:
- Protect shared variables with `sync.Mutex`
- Use atomic operations (`sync/atomic`)
- Review concurrent access patterns
- Check for unsynchronized goroutines

#### Test Timeout

```bash
panic: test timed out after 10m0s
```

**Solution**:
```bash
# Increase timeout
go test -timeout=20m ./...

# Identify slow tests
ginkgo -v | grep "Ran.*in.*seconds"
```

#### CGO Not Available

```bash
# cgo not enabled
```

**Solution**:
```bash
# Enable CGO
export CGO_ENABLED=1

# Install build tools
# Ubuntu: sudo apt-get install build-essential
# macOS: xcode-select --install
```

#### Coverage Report Empty

```bash
# Clean cache and regenerate
go clean -testcache
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Import Cycle

```bash
import cycle not allowed
```

**Solution**:
- Use `package_test` for external tests
- Move shared test utilities to separate package
- Review package dependencies

---

## CI Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Run tests
        run: go test -v -cover ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Generate coverage
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests with race detection..."
CGO_ENABLED=1 go test -race ./...

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" | awk '{if ($5 < 80) exit 1}'

if [ $? -ne 0 ]; then
    echo "Coverage below 80%. Commit aborted."
    exit 1
fi

echo "All checks passed!"
```

### Makefile Integration

```makefile
.PHONY: test test-race test-cover test-all

test:
	go test -v ./...

test-race:
	CGO_ENABLED=1 go test -race -v ./...

test-cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

test-all: test-race test-cover
	@echo "All tests completed"

bench:
	go test -bench=. -benchmem ./...
```

---

## Resources

### Testing Frameworks

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Coverage](https://go.dev/blog/cover)

### Concurrency & Race Detection

- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)
- [atomic Package](https://pkg.go.dev/sync/atomic)

### Best Practices

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

### Package-Specific Testing

For detailed testing information for specific packages:
- [archive/TESTING.md](archive/TESTING.md) - Archive and compression testing
- [atomic/TESTING.md](atomic/TESTING.md) - Atomic operations testing
- [errors/TESTING.md](errors/TESTING.md) - Error handling testing
- [size/TESTING.md](size/TESTING.md) - Size operations testing

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## Quality Checklist

Before submitting code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained or improved: `go test -cover ./...`
- [ ] New features have tests with ≥80% coverage
- [ ] Edge cases tested (nil, empty, large inputs, errors)
- [ ] Concurrent operations tested
- [ ] Documentation updated
- [ ] Test names are descriptive
- [ ] Tests are independent and isolated
- [ ] Resources properly cleaned up

---

**Version**: Go 1.22+ on Linux, macOS, Windows  
**Maintained By**: golib Contributors
