# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-306%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-85.6%25-brightgreen)]()

Comprehensive testing documentation for the status package, covering test execution, race detection, benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Benchmarks](#benchmarks)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The status package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Summary**
- Total Specs: 306 across 4 packages
- Overall Coverage: 85.6%
- Race Detection: ✅ Zero data races
- Execution Time: ~11s (without race), ~22s (with race)

**Package Breakdown**

| Package | Specs | Coverage | Duration | Focus |
|---------|-------|----------|----------|-------|
| `status` | 120 | 85.6% | 10.7s | Main status logic, HTTP routes |
| `control` | 102 | 95.0% | 0.01s | Mode validation, encoding |
| `mandatory` | 55 | 76.1% | 0.1s | Component group management |
| `listmandatory` | 29 | 75.4% | 0.5s | Multiple group handling |

**Coverage Areas**
- Status computation with control modes
- HTTP endpoint with format/verbosity options
- Caching with atomic operations
- Configuration and validation
- Thread-safe concurrent operations
- Component pool management
- Encoding (JSON, YAML, TOML, CBOR, Text)

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Benchmarks
go test -bench=. -benchmem ./mandatory/

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

# Specific package
go test -v ./control/

# Specific test
go test -v -run TestControl ./control/
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific package
ginkgo ./control/

# Pattern matching
ginkgo --focus="health check"

# Parallel execution
ginkgo -p

# With coverage
ginkgo -cover

# JUnit report
ginkgo --junit-report=results.xml

# Verbose
ginkgo -v
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Specific package
CGO_ENABLED=1 go test -race ./mandatory/
```

**Validates**:
- Atomic operations (`atomic.Int32`, `atomic.Int64`, `atomic.Value`)
- Mutex protection (`sync.RWMutex`)
- Concurrent pool access
- Cache operations

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/status	21.340s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected across all packages

### Performance & Profiling

```bash
# Benchmarks
go test -bench=. -benchmem ./mandatory/

# Memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out

# CPU profiling
go test -cpuprofile=cpu.out ./...
go tool pprof cpu.out

# Trace execution
go test -trace=trace.out ./...
go tool trace trace.out
```

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite | ~11s | Without race |
| With `-race` | ~22s | 2x slower (normal) |
| Control Tests | <0.1s | Very fast |
| Status Tests | ~10s | Monitor stabilization delays |

---

## Test Coverage

**Target**: ≥80% statement coverage (currently 85.6%)

### Coverage By Package

```bash
# View coverage by package
go test -cover ./...
```

**Output**:
```
github.com/nabbar/golib/status                coverage: 85.6% of statements
github.com/nabbar/golib/status/control        coverage: 95.0% of statements
github.com/nabbar/golib/status/listmandatory  coverage: 75.4% of statements
github.com/nabbar/golib/status/mandatory      coverage: 76.1% of statements
```

### Coverage By File

```bash
# Generate detailed report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Example Output**:
```
status/interface.go:194:    New             100.0%
status/config.go:83:        ParseList       100.0%
status/cache.go:54:         Max             100.0%
status/model.go:94:         IsHealthy       91.7%
status/route.go:87:         MiddleWare      88.9%
```

### View HTML Coverage

```bash
# Generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Open in browser
xdg-open coverage.html  # Linux
open coverage.html      # macOS
start coverage.html     # Windows
```

### Test Organization

| File | Purpose | Specs |
|------|---------|-------|
| `status_test.go` | Core status functionality | 15 |
| `config_test.go` | Configuration and validation | 12 |
| `route_test.go` | HTTP endpoint handling | 25 |
| `pool_test.go` | Monitor pool operations | 10 |
| `info_test.go` | Version information | 15 |
| `cache_test.go` | Status caching | 8 |
| `encode_test.go` | Response marshaling | 6 |
| `concurrent_test.go` | Thread safety | 12 |
| `control_modes_test.go` | Control mode logic | 17 |

---

## Thread Safety

Thread safety is critical for the status package's concurrent operations.

### Concurrency Primitives

```go
// Atomic operations for cache
atomic.Int32   // Cache duration
atomic.Int64   // Cached status value
atomic.Value   // Last computation time

// Mutex protection
sync.RWMutex   // Configuration and pool access

// Atomic operations in mandatory
atomic.Uint32  // Mode storage
atomic.Value   // Key list storage
```

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| `status.cache` | `atomic.Value` + `atomic.Int64` | ✅ Race-free |
| `mandatory.mode` | `atomic.Uint32` | ✅ Race-free |
| `mandatory.keys` | `atomic.Value` | ✅ Race-free |
| Pool operations | `sync.RWMutex` | ✅ Race-free |
| Config storage | `libctx.Config` | ✅ Race-free |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Concurrent" ./...

# Stress test (run multiple times)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done

# With timeout
CGO_ENABLED=1 go test -race -timeout=30s ./...
```

**Result**: Zero data races across all test runs

### Concurrent Test Examples

```go
// From concurrent_test.go
It("should handle concurrent status checks", func() {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = status.IsHealthy()
            _ = status.IsCacheHealthy()
        }()
    }
    wg.Wait()
})

// From mandatory/benchmark_test.go
BenchmarkConcurrentReads(b *testing.B) {
    m := mandatory.New()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _ = m.KeyHas("test")
            _ = m.GetMode()
        }
    })
}
```

---

## Benchmarks

Benchmarks validate performance characteristics and detect regressions.

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem ./mandatory/

# Specific benchmark
go test -bench=BenchmarkKeyAdd -benchmem ./mandatory/

# With CPU profile
go test -bench=. -cpuprofile=cpu.out ./mandatory/
go tool pprof cpu.out

# Compare results
go test -bench=. -benchmem ./mandatory/ > old.txt
# Make changes...
go test -bench=. -benchmem ./mandatory/ > new.txt
benchcmp old.txt new.txt
```

### Benchmark Results

**Mandatory Package** (AMD Ryzen 9 7900X3D, Go 1.21+)

```
BenchmarkNew-12                    13,963,723      77.33 ns/op      72 B/op    4 allocs/op
BenchmarkSetMode-12               135,494,511       9.02 ns/op       0 B/op    0 allocs/op
BenchmarkGetMode-12               174,430,036       6.77 ns/op       0 B/op    0 allocs/op
BenchmarkKeyAdd-12                 23,573,498      45.56 ns/op      40 B/op    2 allocs/op
BenchmarkKeyAddMultiple-12         20,614,338      58.82 ns/op      24 B/op    1 allocs/op
BenchmarkKeyHas-12                100,000,000      10.04 ns/op       0 B/op    0 allocs/op
BenchmarkKeyDel-12                  1,959,472     730.90 ns/op      88 B/op    4 allocs/op
BenchmarkKeyList-12                21,103,474      57.20 ns/op      80 B/op    1 allocs/op
BenchmarkKeyListLarge-12              346,896    3650.00 ns/op   16384 B/op    1 allocs/op
BenchmarkConcurrentReads-12        37,712,851      32.36 ns/op      48 B/op    1 allocs/op
BenchmarkConcurrentWrites-12       10,351,329     118.60 ns/op      40 B/op    2 allocs/op
BenchmarkMixedOperations-12        31,084,405      38.72 ns/op      15 B/op    0 allocs/op
```

### Performance Insights

| Operation | Performance | Memory | Notes |
|-----------|-------------|--------|-------|
| `GetMode()` | 6.77 ns | 0 B | Zero-allocation atomic read |
| `KeyHas()` | 10.04 ns | 0 B | Lock-free lookup |
| `KeyAdd()` | 45.56 ns | 40 B | Efficient append |
| Concurrent reads | 32.36 ns | 48 B | Scales well |
| Concurrent writes | 118.60 ns | 40 B | Properly synchronized |

**Key Takeaways**:
- Read operations are extremely fast (<10ns)
- Zero allocations for mode operations
- Concurrent operations scale linearly
- Memory usage is minimal and predictable

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should compute status as KO when mandatory component fails", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should cache status for 3 seconds", func() {
    // Arrange
    sts := status.New(ctx)
    sts.SetInfo("test", "v1.0.0", "hash")
    
    // Act
    healthy1 := sts.IsCacheHealthy()
    time.Sleep(100 * time.Millisecond)
    healthy2 := sts.IsCacheHealthy()
    
    // Assert
    Expect(healthy1).To(Equal(healthy2)) // Same cached value
})
```

**3. Use Appropriate Matchers**
```go
Expect(status).To(Equal(monsts.OK))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement("database"))
Expect(count).To(BeNumerically(">", 0))
Expect(mode).To(BeEquivalentTo(control.Must))
```

**4. Always Cleanup Resources**
```go
var mon montps.Monitor

BeforeEach(func() {
    mon = createMonitor()
    pool.MonitorAdd(mon)
})

AfterEach(func() {
    pool.MonitorDel(mon.Name())
})
```

**5. Test Edge Cases**
- Empty configurations
- Nil values
- Concurrent access
- Invalid inputs
- Timeout scenarios

### Test Template

```go
var _ = Describe("status/feature", func() {
    var (
        ctx    context.Context
        status libsts.Status
        pool   monpol.Pool
    )

    BeforeEach(func() {
        ctx = context.Background()
        status = libsts.New(ctx)
        pool = monpol.New(ctx)
        status.RegisterPool(func() montps.Pool { return pool })
    })

    Context("When testing feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            status.SetInfo("test", "v1.0.0", "hash")
            
            // Act
            result := status.IsHealthy()
            
            // Assert
            Expect(result).To(BeTrue())
        })

        It("should handle error case", func() {
            // Test error scenario
        })
    })
})
```

---

## Best Practices

### Test Independence

**✅ Good Practices**
- Each test should be independent
- Use `BeforeEach`/`AfterEach` for setup/cleanup
- Avoid global mutable state
- Create fresh instances per test
- Don't rely on test execution order

```go
// ✅ Good: Independent tests
BeforeEach(func() {
    status = libsts.New(context.Background())
    pool = monpol.New(context.Background())
})

// ❌ Bad: Shared state
var globalStatus = libsts.New(context.Background()) // Reused!
```

### Assertions

**✅ Good**
```go
Expect(err).ToNot(HaveOccurred())
Expect(status).To(Equal(monsts.OK))
Expect(list).To(HaveLen(3))
Expect(mode).To(BeEquivalentTo(control.Must))
```

**❌ Avoid**
```go
Expect(status == monsts.OK).To(BeTrue())  // Less clear error message
Expect(err == nil).To(BeTrue())           // Use HaveOccurred()
```

### Concurrency Testing

```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    errors := make(chan error, 100)
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := performOperation(id); err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        Expect(err).ToNot(HaveOccurred())
    }
})
```

### Performance

- Keep tests fast (target <100ms per spec)
- Use parallel execution when possible
- Avoid unnecessary sleeps (use `Eventually` instead)
- Mock external dependencies

```go
// ✅ Good: Fast check
Eventually(func() bool {
    return status.IsHealthy()
}).Should(BeTrue())

// ❌ Bad: Fixed sleep
time.Sleep(5 * time.Second)
Expect(status.IsHealthy()).To(BeTrue())
```

---

## Troubleshooting

### Stale Coverage

```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

### Test Failures

**Monitor Stabilization**
```go
// Some tests need time for monitors to stabilize
time.Sleep(testMonitorStabilizeDelay)  // ~50ms
```

**Context Cancellation**
```go
// Ensure context is not cancelled
ctx := context.Background()  // Not cancelled
// ctx, cancel := context.WithCancel(...); cancel() // Would fail
```

### Race Conditions

```bash
# Debug races with verbose output
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing atomic operations
- Unsynchronized goroutines

**Example Fix**:
```go
// ❌ Bad: Direct access
if o.mode == control.Must {  // Race condition!

// ✅ Good: Atomic access
if o.GetMode() == control.Must {  // Uses atomic.Load
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

### Test Timeouts

```bash
# Identify slow tests
ginkgo --timeout=10s

# Increase timeout
go test -timeout=30s ./...
```

Check for:
- Goroutine leaks (missing `wg.Done()`)
- Unclosed resources
- Deadlocks

### Debugging

```bash
# Single test
ginkgo --focus="should cache status"

# Specific file
go test -v -run TestConfig ./

# With stack traces
go test -v ./... 2>&1 | grep -A 10 "FAIL"
```

Use `GinkgoWriter` for debug output:
```go
fmt.Fprintf(GinkgoWriter, "Debug: status = %v\n", status)
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
        go: ['1.21', '1.22']
    
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: go test -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
      
      - name: Benchmarks
        run: go test -bench=. -benchmem ./mandatory/
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run tests
go test ./... || exit 1

# Race detection
CGO_ENABLED=1 go test -race ./... || exit 1

# Coverage check
COVERAGE=$(go test -cover ./... | grep "coverage:" | awk '{sum+=$5; count++} END {print sum/count}')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage below 80%: $COVERAGE%"
    exit 1
fi

echo "All checks passed! Coverage: $COVERAGE%"
```

### Makefile Integration

```makefile
.PHONY: test test-race test-cover bench

test:
	go test -v ./...

test-race:
	CGO_ENABLED=1 go test -race -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

bench:
	go test -bench=. -benchmem ./mandatory/ | tee benchmark.txt

ci: test test-race test-cover
	@echo "All CI checks passed!"
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained or improved: ≥85%
- [ ] New features have tests (coverage ≥80%)
- [ ] Edge cases tested
- [ ] Thread safety validated
- [ ] Benchmarks run (no regressions)
- [ ] Test duration reasonable (<15s)
- [ ] Documentation updated

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
- [sync/atomic Package](https://pkg.go.dev/sync/atomic)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Execution Tracer](https://go.dev/doc/diagnostics#execution-tracer)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Status Package Contributors
