# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-101%20Passed-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-90.8%25-brightgreen)]()

Comprehensive testing documentation for the mail/queuer package, covering test execution, race detection, benchmarks, and quality assurance.

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

The mail/queuer package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with behavioral specifications and expressive assertions.

**Test Suite Statistics**
- Total Specs: 101 passed, 1 skipped
- Coverage: 90.8% of statements
- Race Detection: ✅ Zero data races
- Execution Time: ~8.5s (without race), ~10s (with race)

**Coverage Areas**
- Rate limiting logic and time windows
- Counter operations (Pool, Reset, Clone)
- Pooler SMTP operations (Send, Check, Client, Monitor)
- Configuration scenarios and callbacks
- Concurrency and thread safety
- Context cancellation handling
- Error scenarios and edge cases

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
- Rich CLI with filtering and focusing

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax: `Expect(value).To(Equal(expected))`
- Extensive built-in matchers
- Detailed failure messages
- Custom matchers support

**gmeasure** - Performance measurement ([docs](https://onsi.github.io/ginkgo/#gmeasure))
- Statistical analysis of benchmarks
- Experiment-based measurement
- Automatic statistics (mean, median, stddev)
- Report generation

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

# Run with timeout
go test -timeout=10m ./...
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=counter_test.go

# Pattern matching
ginkgo --focus="throttling"

# Parallel execution
ginkgo -p

# JUnit report for CI
ginkgo --junit-report=results.xml

# Verbose output
ginkgo -v

# Skip slow tests
ginkgo --skip="benchmark"
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# With coverage and race detection
CGO_ENABLED=1 go test -race -coverprofile=coverage.out -covermode=atomic ./...
```

**Validates**:
- Mutex protection (`sync.Mutex`)
- Atomic operations
- Goroutine synchronization
- Context cancellation handling
- Counter state access

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/mail/queuer	10.289s
coverage: 90.8% of statements

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected across all tests

### Performance & Profiling

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

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
| Full Suite | ~8.5s | Without race detection |
| With `-race` | ~10s | ~20% slower (normal overhead) |
| Individual Spec | <100ms | Most tests |
| Throttling Tests | 200-500ms | Time window dependent |
| Concurrent Tests | 300-600ms | Multiple goroutines |

---

## Test Coverage

**Target**: ≥90% statement coverage  
**Current**: 90.8%

### Coverage By Component

| Component | File | Coverage | Description |
|-----------|------|----------|-------------|
| **Config** | `config.go` | 100% | Configuration and FuncCaller |
| **Counter** | `counter.go` | 82.6%-100% | Rate limiting logic |
| **Pooler** | `model.go` | 100% | SMTP operations |
| **Monitor** | `monitor.go` | 100% | Health checking |
| **Errors** | `error.go` | 60-66.7% | Error messages |
| **Interface** | `interface.go` | 100% | Constructor |

### Detailed Coverage Report

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

**Sample Output**:
```
github.com/nabbar/golib/mail/queuer/config.go:40:     SetFuncCaller      100.0%
github.com/nabbar/golib/mail/queuer/counter.go:51:    newCounter         100.0%
github.com/nabbar/golib/mail/queuer/counter.go:62:    Pool               82.6%
github.com/nabbar/golib/mail/queuer/counter.go:102:   Reset              100.0%
github.com/nabbar/golib/mail/queuer/counter.go:122:   Clone              100.0%
github.com/nabbar/golib/mail/queuer/model.go:44:      Reset              100.0%
github.com/nabbar/golib/mail/queuer/model.go:54:      NewPooler          100.0%
github.com/nabbar/golib/mail/queuer/model.go:68:      Send               100.0%
total:                                                 (statements)       90.8%
```

### Test File Organization

| File | Purpose | Specs |
|------|---------|-------|
| `queuer_suite_test.go` | Suite initialization and helpers | 1 |
| `counter_test.go` | Counter logic and throttling | 19 |
| `pooler_test.go` | Pooler SMTP operations | 27 |
| `config_test.go` | Configuration and callbacks | 18 |
| `concurrency_test.go` | Concurrent operations | 21 |
| `benchmark_test.go` | Performance benchmarks | 14 |
| `monitor_test.go` | Monitoring integration | 2 |

---

## Thread Safety

Thread safety is critical for the queuer package's concurrent email sending capabilities.

### Concurrency Primitives

```go
// Mutex for state protection
sync.Mutex

// Atomic operations for independent state
atomic.Bool (in test backend)

// Goroutine synchronization
sync.WaitGroup
```

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| `counter.Pool()` | `sync.Mutex` | ✅ Race-free |
| `counter.Reset()` | `sync.Mutex` | ✅ Race-free |
| `counter.Clone()` | `sync.Mutex` + read lock | ✅ Race-free |
| `pooler.Send()` | Thread-safe counter | ✅ Parallel-safe |
| Test Backend | `sync.Mutex` + `atomic.Int32` | ✅ Race-free |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -run "Concurrent" ./...

# Stress test (repeat 10 times)
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done

# Parallel execution with race detection
CGO_ENABLED=1 ginkgo -race -p
```

**Result**: Zero data races across all test runs

### Race Conditions Fixed

During development, the following race conditions were identified and fixed:

1. **Test Backend**: Unprotected slice access in `testBackend.messages`
   - **Fix**: Added `sync.Mutex` to protect concurrent writes
   - **File**: `queuer_suite_test.go:176`

2. **Counter Clone**: Reading state without lock
   - **Fix**: Added mutex lock before reading `c.num` and `c.tim`
   - **File**: `counter.go:123`

---

## Benchmarks

### Performance Metrics

Benchmarks use Ginkgo's `gmeasure` for statistical analysis:

**Counter Performance**
- Reset: <1µs per operation
- Clone: <1µs per operation

**Pooler Throughput (No Throttle)**

| Goroutines | Messages/sec | Total Messages |
|------------|--------------|----------------|
| 1 | ~3,000 | 10 |
| 2 | ~2,800 | 20 |
| 4 | ~2,800 | 40 |
| 8 | ~1,100 | 80 |
| 16 | ~1,100 | 160 |
| 32 | ~1,100 | 320 |

**Pooler Throughput (Throttled: 100/50ms)**

| Goroutines | Messages/sec | Notes |
|------------|--------------|-------|
| 1 | ~1,300 | Enforced limit |
| 32 | ~1,100 | Concurrent throttling |

**Message Sizes**

| Size | Send Time | Notes |
|------|-----------|-------|
| 100 bytes | ~3ms | Minimal overhead |
| 1KB | ~5ms | Standard email |
| 10KB | ~10ms | With attachments |

### Running Benchmarks

```bash
# Run all benchmarks
go test -v -run="Benchmark" ./...

# Using Ginkgo (shows gmeasure reports)
ginkgo --focus="Benchmark" -v

# With memory stats
go test -bench=. -benchmem ./...
```

### Benchmark Structure

Tests use `gmeasure.Experiment` for statistical analysis:

```go
It("should measure throughput", func() {
    experiment := NewExperiment("Throughput Test")
    AddReportEntry(experiment.Name, experiment)
    
    experiment.Sample(func(idx int) {
        experiment.MeasureDuration("operation", func() {
            // Operation to measure
        })
    }, SamplingConfig{N: 100, Duration: 5 * time.Second})
})
```

Reports include:
- Mean, Median, Standard Deviation
- Min/Max values
- Sample count
- Automatic unit conversion (ms, µs, ns)

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should enforce rate limit when max emails reached", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should reset counter to maximum quota", func() {
    // Arrange
    cfg := &queuer.Config{Max: 10, Wait: time.Second}
    pooler := queuer.New(cfg, nil)
    
    // Act
    err := pooler.Reset()
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(HaveLen(5))
Expect(number).To(BeNumerically(">", 0))
Expect(time).To(BeTemporally("~", now, time.Second))
```

**4. Always Cleanup Resources**
```go
defer pooler.Close()
defer server.Close()
defer cancel()
```

**5. Test Edge Cases**
- Zero values (`Max: 0`, `Wait: 0`)
- Negative values
- nil parameters
- Context cancellation
- Concurrent access

**6. Isolate Tests**
- Each test independent
- Use `BeforeEach`/`AfterEach` for setup/cleanup
- Avoid shared mutable state
- Create dedicated SMTP servers per test

### Test Template

```go
var _ = Describe("queuer/new_feature", func() {
    var (
        ctx     context.Context
        cancel  context.CancelFunc
        pooler  queuer.Pooler
        backend *testBackend
        srv     *smtpsv.Server
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
        
        backend = &testBackend{}
        var err error
        srv, _, _, err = startTestSMTPServer(backend, false)
        Expect(err).ToNot(HaveOccurred())
        
        cfg := &queuer.Config{
            Max:  10,
            Wait: 100 * time.Millisecond,
        }
        pooler = queuer.New(cfg, testSMTPClient)
    })

    AfterEach(func() {
        if pooler != nil {
            pooler.Close()
        }
        if srv != nil {
            srv.Close()
        }
        if cancel != nil {
            cancel()
        }
    })

    Context("When using new feature", func() {
        It("should perform expected behavior", func() {
            // Test implementation
            result, err := pooler.NewFeature()
            Expect(err).ToNot(HaveOccurred())
            Expect(result).To(BeTrue())
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
- ✅ Create test resources on-demand
- ❌ Don't rely on test execution order

**SMTP Server Setup**
- Use standalone SMTP server (no external dependencies)
- Create server per test suite or per test
- Support both TLS and non-TLS configurations
- Track sent messages with thread-safe backend

**Test Data**
```go
// Good: Simple, focused test data
message := newSimpleMessage("test content")

// Good: Realistic but controlled
largeContent := make([]byte, 5*1024) // 5KB
```

**Assertions**
```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(count).To(Equal(10))
Expect(backend.msgCount.Load()).To(BeNumerically(">", 0))

// ❌ Avoid: Boolean comparisons
Expect(count == 10).To(BeTrue())
```

**Concurrency Testing**
```go
It("should handle concurrent sends", func() {
    var wg sync.WaitGroup
    errors := make([]error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            errors[idx] = pooler.Send(ctx, from, to, message)
        }(i)
    }
    
    wg.Wait()
    
    for _, err := range errors {
        Expect(err).ToNot(HaveOccurred())
    }
})
```

**Performance**
- Keep tests fast (<10s total)
- Use realistic but minimal data sizes
- Disable throttling in unit tests (`Max: 0`)
- Run benchmarks separately

**Error Handling**
```go
// ✅ Good: Check all errors
result, err := operation()
Expect(err).ToNot(HaveOccurred())
defer result.Close()

// ❌ Bad: Ignore errors
result, _ := operation()
```

---

## Troubleshooting

**Leftover Processes**
```bash
# Check for hanging SMTP servers
lsof -i :0 | grep LISTEN

# Kill if necessary
killall go-smtp
```

**Stale Coverage**
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Test Timeouts**
```bash
# Identify hanging tests
ginkgo --timeout=10s

# Common causes:
# - Goroutine leaks (missing wg.Done())
# - Unclosed resources
# - Deadlocks
```

**Race Conditions**
```bash
# Debug races with full output
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing mutex locks
- Concurrent map access
- Slice modifications

Example fix:
```go
// ❌ Bad: Direct access
messages = append(messages, msg) // Race condition

// ✅ Good: Protected access
mu.Lock()
messages = append(messages, msg)
mu.Unlock()
```

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Verify
export CGO_ENABLED=1
go env CGO_ENABLED
```

**Flaky Tests**
```bash
# Run multiple times to identify
for i in {1..20}; do 
    go test ./... || echo "Failed on run $i"
done

# Common causes:
# - Timing dependencies (use Eventually/Consistently)
# - Shared state between tests
# - External dependencies (network, filesystem)
```

**Debugging**
```bash
# Single test
ginkgo --focus="should handle throttling"

# Specific file
ginkgo --focus-file=counter_test.go

# Verbose output
ginkgo -v --trace

# Debug output in tests
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
        run: go test -v -timeout=10m ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

**GitLab CI Example**
```yaml
test:
  image: golang:1.21
  script:
    - go test -v -timeout=10m ./...
    - CGO_ENABLED=1 go test -race -timeout=10m ./...
    - go test -coverprofile=coverage.out -covermode=atomic ./...
  coverage: '/total.*?(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out
```

**Pre-commit Hook**
```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "✅ All checks passed"
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥90%
- [ ] New features have tests
- [ ] Concurrent scenarios tested
- [ ] Error cases tested
- [ ] Documentation updated
- [ ] Benchmarks added for performance changes
- [ ] Thread safety validated
- [ ] Test duration reasonable (<15s with race)

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [gmeasure Performance](https://onsi.github.io/ginkgo/#gmeasure)
- [Go Testing](https://pkg.go.dev/testing)

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)

**Performance**
- [Go Profiling](https://go.dev/blog/pprof)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)

**Package Specific**
- [SMTP Package Tests](../smtp/TESTING.md)
- [Sender Package Tests](../sender/TESTING.md)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Mail Queuer Package Contributors
