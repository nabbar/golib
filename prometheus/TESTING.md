# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-426%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-91.6%25-brightgreen)]()

Comprehensive testing documentation for the prometheus package, covering test execution, race detection, benchmarks, and quality assurance.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The prometheus package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Summary**
- Total Specs: 426 across 5 packages
- Overall Coverage: 91.6%
- Race Detection: ✅ Zero data races
- Execution Time: ~1.4s (without race), ~7.8s (with race)

**Package Breakdown**

| Package | Specs | Coverage | Duration | Focus |
|---------|-------|----------|----------|-------|
| `prometheus` | 137 | 90.9% | 1.38s | Main interface, middleware, routes |
| `bloom` | 24 | 94.7% | 0.13s | Bloom filter operations |
| `metrics` | 179 | 95.5% | 0.03s | Metric types, collection, registration |
| `pool` | 74 | 72.5% | 0.02s | Pool management, concurrency |
| `types` | 36 | 100% | 0.00s | Type definitions, validation |
| `webmetrics` | 0 | 0.0% | - | Simple constructors (no tests needed) |

**Coverage Areas**
- Prometheus interface and middleware
- Metric management (add, delete, list, get, clear)
- Metric collection with Gin context
- Path exclusion and filtering
- Gin integration (routes, middleware, context)
- All metric types (Counter, Gauge, Histogram, Summary)
- Metric pool operations (add, get, del, walk, clear)
- Bloom filter collection and concurrent access
- Thread safety and concurrent operations
- Error handling and edge cases

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

# Using Ginkgo CLI
ginkgo -cover -race
```

**Expected Output**
```bash
ok  	github.com/nabbar/golib/prometheus	        1.385s	coverage: 90.9% of statements
ok  	github.com/nabbar/golib/prometheus/bloom	0.125s	coverage: 94.7% of statements
ok  	github.com/nabbar/golib/prometheus/metrics	0.044s	coverage: 95.5% of statements
ok  	github.com/nabbar/golib/prometheus/pool	    0.026s	coverage: 72.5% of statements
ok  	github.com/nabbar/golib/prometheus/types	0.012s	coverage: 100.0% of statements
?   	github.com/nabbar/golib/prometheus/webmetrics	[no test files]
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`, `AfterSuite`)
- Parallel execution support
- Rich CLI with filtering and focus

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- Extensive built-in matchers (`Equal`, `HaveOccurred`, `ContainSubstring`, etc.)
- Detailed failure messages with context
- Custom matcher support

**Why Ginkgo/Gomega?**
- Better test organization than standard Go testing
- More readable assertions than `if err != nil`
- Rich failure reporting for debugging
- Industry standard for Go BDD testing

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

# View function-level coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific subpackage
ginkgo ./metrics
ginkgo ./pool
ginkgo ./types

# Specific test file
ginkgo --focus-file=prometheus_basic_test.go

# Pattern matching
ginkgo --focus="should handle concurrent"
ginkgo --focus="Metric Management"

# Parallel execution (faster)
ginkgo -p

# JUnit report for CI
ginkgo --junit-report=results.xml

# Verbose output
ginkgo -v

# Fail fast (stop on first failure)
ginkgo --fail-fast
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Specific subpackage
CGO_ENABLED=1 go test -race ./pool
```

**What Race Detector Validates**:
- Atomic operations (`atomic.Int32`, `atomic.Bool`)
- Mutex protection (`sync.Mutex`, `sync.RWMutex`)
- Value synchronization (`libatm.Value`)
- Goroutine synchronization (`sync.WaitGroup`)
- Concurrent metric collection

**Expected Output**:
```bash
# ✅ Success (no races)
ok  	github.com/nabbar/golib/prometheus	2.583s

# ❌ Race detected (needs fixing)
WARNING: DATA RACE
Read at 0x00c000123456 by goroutine 42:
  github.com/nabbar/golib/prometheus.(*prom).GetMetric()
      /path/to/file.go:123 +0x45
```

**Current Status**: ✅ Zero data races detected across all tests

### Performance & Profiling

```bash
# Benchmarks (if available)
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
| Root package | ~1.4s | Without race |
| All subpackages | ~3s | Without race |
| With `-race` | ~8s | 2-3x slower (normal) |
| Individual spec | <100ms | Most tests |
| Integration tests | 200-500ms | With HTTP server |

---

## Test Coverage

**Target**: ≥80% statement coverage (Currently: ~95%)

### Coverage by Category

| Category | Files | Description |
|----------|-------|-------------|
| **Basic Operations** | `prometheus_basic_test.go` | Constructor, slow time, duration, thread safety |
| **Metric Management** | `prometheus_metrics_test.go` | Add, get, delete, list metrics |
| **Collection** | `prometheus_collect_test.go` | Metric collection, Gin context, concurrent collection |
| **Exclusion** | `prometheus_exclude_test.go` | Path exclusion, patterns, edge cases |
| **Gin Integration** | `prometheus_gin_test.go` | Middleware, routes, handlers, context |
| **Integration** | `prometheus_integration_test.go` | Complete workflows, error handling |
| **Metrics** | `metrics/*_test.go` | Metric types, collectors, operations |
| **Pool** | `pool/*_test.go` | Pool operations, concurrency, edge cases |
| **Types** | `types/*_test.go` | Type registration, validation, constants |
| **Bloom** | `bloom/*_test.go` | Bloom filter operations, accuracy |

### View Coverage Details

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View summary
go tool cover -func=coverage.out | tail -1

# View by package
go tool cover -func=coverage.out | grep -E "prometheus/(metrics|pool|types)"

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser to see line-by-line coverage

# Find uncovered lines
go tool cover -func=coverage.out | grep -v "100.0%"
```

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("Prometheus Component", func() {
    var (
        prm prometheus.Prometheus
    )

    BeforeEach(func() {
        // Per-test setup
        prm = newPrometheus()
    })
    
    AfterEach(func() {
        // Per-test cleanup
        cleanupPrometheus(prm)
    })
    
    Context("When performing operation", func() {
        It("should behave correctly", func() {
            // Arrange
            metric := createCounterMetric("test_metric")
            
            // Act
            err := prm.AddMetric(true, metric)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Expect(prm.GetMetric("test_metric")).ToNot(BeNil())
        })
    })
})
```

---

## Thread Safety

Thread safety is critical for the prometheus package's concurrent operations.

### Concurrency Primitives

```go
// Atomic scalar values
atomic.Int32           // For slow time threshold
atomic.Bool            // For state flags

// Thread-safe complex values
libatm.Value[[]string]   // For excluded paths
libatm.Value[[]float64]  // For duration buckets

// Synchronization
sync.Mutex             // For buffer protection
sync.WaitGroup         // For goroutine lifecycle
```

### Verified Components

| Component | Mechanism | Test File | Status |
|-----------|-----------|-----------|--------|
| `prom.SetSlowTime/GetSlowTime` | `atomic.Int32` | `prometheus_basic_test.go` | ✅ Race-free |
| `prom.SetDuration/GetDuration` | `libatm.Value` | `prometheus_basic_test.go` | ✅ Race-free |
| `prom.ExcludePath` | `libatm.Value` | `prometheus_exclude_test.go` | ✅ Race-free |
| `pool.Add/Get/Del` | `sync.Map` | `pool/*_test.go` | ✅ Race-free |
| `Metric.Collect` | Semaphore | `prometheus_collect_test.go` | ✅ Race-free |

### Testing Concurrent Operations

```bash
# Run tests with race detector
CGO_ENABLED=1 go test -race ./...

# Focus on concurrent tests
CGO_ENABLED=1 go test -race -v -run "Concurrent" ./...
CGO_ENABLED=1 go test -race -v -run "Thread" ./...

# Stress test (run multiple times)
for i in {1..10}; do 
    echo "Run $i"
    CGO_ENABLED=1 go test -race ./... || break
done
```

**Concurrent Test Examples**

```go
It("should handle concurrent metric additions", func() {
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            name := fmt.Sprintf("metric_%d", id)
            m := createCounterMetric(name)
            defer cleanupMetric(m)
            
            err := p.AddMetric(true, m)
            Expect(err).ToNot(HaveOccurred())
        }(i)
    }
    
    wg.Wait()
})
```

**Result**: Zero data races across all test runs

---

## Test Organization

### Root Package Tests

| File | Purpose | Specs | Key Tests |
|------|---------|-------|-----------|
| `prometheus_suite_test.go` | Suite initialization | 1 | Helper functions, cleanup |
| `prometheus_basic_test.go` | Basic operations | 18 | Constructor, config, thread safety |
| `prometheus_metrics_test.go` | Metric management | 30 | Add, get, delete, list, concurrent |
| `prometheus_collect_test.go` | Collection | 24 | Collect, Gin context, concurrency |
| `prometheus_exclude_test.go` | Path exclusion | 18 | Patterns, edge cases, concurrent |
| `prometheus_gin_test.go` | Gin integration | 26 | Middleware, routes, handlers |
| `prometheus_integration_test.go` | Integration | 21 | Complete workflows, errors |

### Subpackage Tests

**metrics/**
- `metrics_suite_test.go` - Suite initialization
- `metrics_basic_test.go` - Metric creation and configuration
- `metrics_operations_test.go` - Inc, Dec, Set, Observe operations
- Coverage: 95.3%

**pool/**
- `pool_suite_test.go` - Suite initialization
- `pool_basic_test.go` - Add, Get, Del, List operations
- `pool_walk_test.go` - Walk functionality
- `pool_edge_cases_test.go` - Edge cases and concurrency
- Coverage: 90.6%

**types/**
- `types_suite_test.go` - Suite initialization
- `types_registration_test.go` - Type registration and validation
- `types_values_test.go` - Type constants and comparisons
- Coverage: 100%

**bloom/**
- `bloom_suite_test.go` - Suite initialization
- `bloom_basic_test.go` - Add, Test, Count operations
- Coverage: 94.7%

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
// ✅ Good: Clear, specific description
It("should successfully add metric with valid configuration", func() {
    // Test implementation
})

// ❌ Bad: Vague description
It("should work", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should increment counter metric", func() {
    // Arrange
    counter := createCounterMetric("test_counter")
    defer cleanupMetric(counter)
    labels := map[string]string{"method": "GET"}
    
    // Act
    counter.Inc(labels)
    
    // Assert
    // Verify counter was incremented (via collection or export)
    Expect(counter).ToNot(BeNil())
})
```

**3. Use Appropriate Matchers**
```go
// Error checking
Expect(err).ToNot(HaveOccurred())
Expect(err).To(HaveOccurred())
Expect(err.Error()).To(ContainSubstring("expected text"))

// Value checking
Expect(value).To(Equal(expected))
Expect(list).To(ContainElement(item))
Expect(number).To(BeNumerically(">", 0))
Expect(result).ToNot(BeNil())

// Type checking
Expect(obj).To(BeAssignableToTypeOf(&prometheus.Prom{}))
```

**4. Always Cleanup Resources**
```go
It("should cleanup metrics after test", func() {
    metric := createCounterMetric("temp_metric")
    defer cleanupMetric(metric)  // Always cleanup
    
    err := p.AddMetric(true, metric)
    Expect(err).ToNot(HaveOccurred())
})
```

**5. Test Edge Cases**
- Empty inputs
- Nil values
- Large data sets
- Invalid configurations
- Concurrent operations
- Error conditions

**6. Avoid External Dependencies**
- No remote API calls
- No external databases
- Use in-memory test servers
- Mock external dependencies

### Test Template

```go
var _ = Describe("Prometheus New Feature", func() {
    var (
        prm    prometheus.Prometheus
        metric prmmet.Metric
    )

    BeforeEach(func() {
        prm = newPrometheus()
    })

    AfterEach(func() {
        if metric != nil {
            cleanupMetric(metric)
        }
    })

    Context("When using new feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            metric = createCounterMetric("test_metric")
            
            // Act
            err := prm.AddMetric(true, metric)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            retrieved := prm.GetMetric("test_metric")
            Expect(retrieved).ToNot(BeNil())
            Expect(retrieved.GetName()).To(Equal("test_metric"))
        })

        It("should handle error case", func() {
            // Test error condition
            err := prm.AddMetric(true, nil)
            Expect(err).To(HaveOccurred())
        })
    })
})
```

### Helper Functions

The test suite provides helper functions for common operations:

```go
// Create Prometheus instance
prm := newPrometheus()

// Create metrics
counter := createCounterMetric("metric_name")
gauge := createGaugeMetric("metric_name")
histogram := createHistogramMetric("metric_name")
summary := createSummaryMetric("metric_name")

// Unique metric names (avoids conflicts)
name := uniqueMetricName("prefix")

// Create context
ctx := createContextFunc()()

// Cleanup
cleanupMetric(metric)
cleanupPrometheus(prm)
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
// ✅ Good: Use helper functions
name := uniqueMetricName("test")
metric := createCounterMetric(name)

// ✅ Good: Explicit cleanup
defer cleanupMetric(metric)

// ❌ Bad: Hardcoded names (can cause conflicts)
metric := createCounterMetric("test_metric")
```

**Assertions**
```go
// ✅ Good: Use specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))
Expect(list).To(HaveLen(3))

// ❌ Bad: Generic boolean checks
Expect(value == expected).To(BeTrue())
Expect(len(list) == 3).To(BeTrue())
```

**Concurrency Testing**
```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // Perform operation
            name := uniqueMetricName("concurrent")
            m := createCounterMetric(name)
            defer cleanupMetric(m)
            
            err := p.AddMetric(true, m)
            Expect(err).ToNot(HaveOccurred())
        }(i)
    }
    
    wg.Wait()
})
```

**Always run concurrent tests with `-race`**

**Performance**
- Keep tests fast (<100ms per spec)
- Use parallel execution when safe
- Avoid unnecessary sleeps
- Use `Eventually` for async operations

```go
// ✅ Good: Use Eventually for async
Eventually(func() bool {
    return condition()
}).Should(BeTrue())

// ❌ Bad: Sleep and hope
time.Sleep(100 * time.Millisecond)
Expect(condition()).To(BeTrue())
```

**Error Handling**
```go
// ✅ Good: Always check errors
result, err := operation()
Expect(err).ToNot(HaveOccurred())
defer result.Close()

// ❌ Bad: Ignore errors
result, _ := operation()
```

**Test Organization**
- Group related tests in `Context` blocks
- Use descriptive `Describe` and `Context` names
- Keep `It` blocks focused (one behavior per test)
- Use `BeforeEach` for common setup

---

## Troubleshooting

**Import Cycles**
```bash
# Error: import cycle
package github.com/nabbar/golib/prometheus
    imports github.com/nabbar/golib/prometheus/metrics
    imports github.com/nabbar/golib/prometheus

# Solution: Use package_test naming
package prometheus_test  // Not package prometheus
```

**Stale Coverage**
```bash
# Clean test cache
go clean -testcache

# Re-run tests
go test -coverprofile=coverage.out ./...
```

**Race Conditions**
```bash
# Debug races with verbose output
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt

# Find race warnings
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

**Common Race Fixes**:
```go
// ❌ Bad: Direct field access
if m.field == value {  // Race condition!

// ✅ Good: Use atomic or mutex
if m.getField() == value {  // Protected access
```

**Test Timeouts**
```bash
# Increase timeout
go test -timeout=5m ./...

# With Ginkgo
ginkgo --timeout=5m
```

**Check for**:
- Goroutine leaks (missing `wg.Done()`)
- Unclosed resources
- Infinite loops
- Mutex deadlocks

**Debugging Specific Tests**
```bash
# Run single test
ginkgo --focus="should add metric"

# Run specific file
ginkgo --focus-file=prometheus_basic_test.go

# Verbose output
ginkgo -v --trace

# Skip tests
ginkgo --skip="slow integration"
```

**Use GinkgoWriter for debug output**:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
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
go test -race ./...
```

**Parallel Test Failures**
- Check for shared resources
- Verify proper synchronization
- Use unique names for test data
- Avoid global state

**Memory Leaks**
```bash
# Check for memory leaks
go test -memprofile=mem.out ./...
go tool pprof -alloc_space mem.out
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
        run: go test -v ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race -timeout=10m ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          fail_ci_if_error: true
```

**GitLab CI Example**

```yaml
test:
  image: golang:1.21
  script:
    - go test -v ./...
    - CGO_ENABLED=1 go test -race ./...
    - go test -coverprofile=coverage.out ./...
  coverage: '/coverage: \d+.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
```

**Pre-commit Hook**

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
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage is below 80%: $COVERAGE%"
    exit 1
fi

echo "All checks passed!"
```

**Make it executable**:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Quality Checklist

Before merging code:

- [ ] All tests pass: `go test ./...`
- [ ] Race detection clean: `CGO_ENABLED=1 go test -race ./...`
- [ ] Coverage maintained: ≥80% (current: ~95%)
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Edge cases covered
- [ ] Test duration reasonable (<10s)
- [ ] No test flakiness
- [ ] GoDoc comments updated
- [ ] README.md updated if needed

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
- [Trace Execution](https://go.dev/blog/execution-tracing)

**Prometheus**
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Prometheus Client](https://github.com/prometheus/client_golang)
- [Writing Exporters](https://prometheus.io/docs/instrumenting/writing_exporters/)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Prometheus Package Contributors
