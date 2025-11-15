# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-595%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-86.1%25-brightgreen)]()

Comprehensive testing documentation for the monitor package, covering test execution, race detection, and quality assurance for health monitoring systems.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Thread Safety](#thread-safety)
- [Testing Strategies](#testing-strategies)
- [Package-Specific Tests](#package-specific-tests)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Overview

The monitor package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite Summary**
- Total Specs: 595 across 4 packages
- Overall Coverage: 86.1%
- Race Detection: ✅ Zero data races
- Execution Time: ~12.1s (without race), ~31s (with race)

**Package Breakdown**

| Package | Specs | Coverage | Duration | Focus |
|---------|-------|----------|----------|-------|
| `monitor` | 122 | 68.5% | 0.23s | Core health monitoring, transitions |
| `info` | 139 | 100% | 0.12s | Dynamic metadata management |
| `pool` | 153 | 76.2% | 11.78s | Monitor pool, batch operations |
| `status` | 181 | 98.4% | 0.01s | Status enumeration, encoding |
| `types` | 0 | 0.0% | - | Type definitions (no tests needed) |

**Coverage Areas**
- Health check execution and transitions
- Status state machine (OK ↔ Warn ↔ KO)
- Configuration validation and normalization
- Metrics collection (latency, uptime, downtime)
- Pool management and batch operations
- Prometheus metrics integration
- Dynamic metadata with functions
- Status encoding (JSON, YAML, TOML, Text)
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
ok  	github.com/nabbar/golib/monitor	        0.232s	coverage: 68.5% of statements
ok  	github.com/nabbar/golib/monitor/info	0.121s	coverage: 100.0% of statements
ok  	github.com/nabbar/golib/monitor/pool	11.795s	coverage: 76.2% of statements
ok  	github.com/nabbar/golib/monitor/status	0.018s	coverage: 98.4% of statements
?   	github.com/nabbar/golib/monitor/types	[no test files]
```

---

## Test Framework

### Ginkgo v2

**BDD testing framework** - [Documentation](https://onsi.github.io/ginkgo/)

Features:
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`, `BeforeSuite`)
- Focused tests (`FDescribe`, `FIt`) and skip (`XDescribe`, `XIt`)
- Parallel execution support
- Rich reporting

### Gomega

**Matcher library** - [Documentation](https://onsi.github.io/gomega/)

Critical matchers for this package:
- `Eventually()`: Wait for async operations (status transitions)
- `Consistently()`: Verify stability over time
- `BeNumerically()`: Compare durations and counts
- `HaveOccurred()`: Error checking

---

## Running Tests

### Basic Commands

```bash
# All packages
go test ./...

# Verbose output
go test -v ./...

# Specific package
go test ./monitor
go test ./monitor/info
go test ./monitor/status
go test ./monitor/pool

# With short flag (skip long-running tests)
go test -short ./...
```

### Coverage

```bash
# Basic coverage
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Per-package coverage
go test -coverprofile=monitor_coverage.out ./monitor
go test -coverprofile=pool_coverage.out ./monitor/pool
go test -coverprofile=info_coverage.out ./monitor/info

# Coverage with minimum threshold
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%"
```

### Race Detection

**Critical for this package** - Always run before commits:

```bash
# Enable race detector
CGO_ENABLED=1 go test -race ./...

# With timeout (some tests are time-dependent)
CGO_ENABLED=1 go test -race -timeout=5m ./...

# Specific package
CGO_ENABLED=1 go test -race ./monitor/pool
```

### Ginkgo CLI

```bash
# Install
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
ginkgo

# With options
ginkgo -v -race -cover

# Specific package
ginkgo ./monitor/pool

# Focused tests only
ginkgo -focus="Status Transitions"

# Skip tests
ginkgo -skip="Slow.*"

# Parallel execution
ginkgo -p
```

---

## Test Coverage

### Current Coverage

| Package | Coverage | Lines | Key Areas |
|---------|----------|-------|-----------|
| **monitor** | 68.5% | 1,200 | Status transitions, lifecycle, metrics |
| **pool** | 76.7% | 800 | Pool ops, shell commands, metrics aggregation |
| **info** | 85.3% | 400 | Metadata management, caching |
| **status** | 92.1% | 300 | Enumeration, encoding, parsing |

### Coverage Goals

- **Critical paths**: ≥ 80% (status transitions, lifecycle)
- **Standard functionality**: ≥ 70% (metrics, encoding)
- **Edge cases**: ≥ 60% (error paths, corner cases)

### Improving Coverage

```bash
# Identify uncovered code
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"

# pool package has a coverage script
cd monitor/pool
./test_coverage.sh --html

# View in browser
go tool cover -html=coverage.out
```

---

## Thread Safety Testing

### Race Detector

**Always run with `-race` flag**:

```bash
CGO_ENABLED=1 go test -race ./...
```

### Common Race Conditions to Test

1. **Concurrent Start/Stop**
   ```go
   It("should handle concurrent start/stop", func() {
       var wg sync.WaitGroup
       for i := 0; i < 10; i++ {
           wg.Add(1)
           go func() {
               defer wg.Done()
               mon.Start(ctx)
               time.Sleep(10 * time.Millisecond)
               mon.Stop(ctx)
           }()
       }
       wg.Wait()
   })
   ```

2. **Concurrent Status Reads**
   ```go
   It("should handle concurrent status reads", func() {
       for i := 0; i < 100; i++ {
           go func() {
               _ = mon.Status()
               _ = mon.Latency()
               _ = mon.Uptime()
           }()
       }
       time.Sleep(100 * time.Millisecond)
   })
   ```

3. **Pool Concurrent Operations**
   ```go
   It("should handle concurrent pool operations", func() {
       var wg sync.WaitGroup
       for i := 0; i < 10; i++ {
           wg.Add(3)
           go func(idx int) {
               defer wg.Done()
               pool.MonitorAdd(createMonitor(fmt.Sprintf("mon-%d", idx)))
           }(i)
           go func(idx int) {
               defer wg.Done()
               pool.MonitorGet(fmt.Sprintf("mon-%d", idx))
           }(i)
           go func(idx int) {
               defer wg.Done()
               pool.MonitorList()
           }(i)
       }
       wg.Wait()
   })
   ```

---

## Testing Strategies

### 1. Status Transition Testing

Test the three-state model with hysteresis:

```go
Describe("Status Transitions", func() {
    It("should transition KO → Warn → OK", func() {
        // Configure thresholds
        cfg := types.Config{
            RiseCountKO:   3,
            RiseCountWarn: 2,
        }
        mon.SetConfig(ctx, cfg)
        
        // Mock successful health checks
        successCount := 0
        mon.SetHealthCheck(func(ctx context.Context) error {
            successCount++
            return nil
        })
        
        mon.Start(ctx)
        defer mon.Stop(ctx)
        
        // Wait for transitions
        Eventually(mon.Status, "5s").Should(Equal(status.Warn))
        Eventually(mon.Status, "3s").Should(Equal(status.OK))
    })
})
```

### 2. Timing-Dependent Tests

Use `Eventually` for async operations:

```go
It("should start monitoring within timeout", func() {
    mon.Start(ctx)
    
    // Eventually waits up to timeout for condition
    Eventually(mon.IsRunning, "2s", "100ms").Should(BeTrue())
    
    // Consistently checks condition remains stable
    Consistently(mon.IsRunning, "500ms", "50ms").Should(BeTrue())
})
```

### 3. Error Path Testing

Test error handling and recovery:

```go
It("should handle health check errors", func() {
    errorCount := 0
    mon.SetHealthCheck(func(ctx context.Context) error {
        errorCount++
        if errorCount < 3 {
            return errors.New("temporary failure")
        }
        return nil
    })
    
    mon.Start(ctx)
    defer mon.Stop(ctx)
    
    // Should transition to Warn after failures
    Eventually(mon.Status).Should(Equal(status.Warn))
    
    // Should recover after successes
    Eventually(mon.Status, "10s").Should(Equal(status.OK))
})
```

### 4. Metrics Validation

Verify metrics tracking:

```go
It("should track metrics correctly", func() {
    mon.Start(ctx)
    defer mon.Stop(ctx)
    
    time.Sleep(2 * time.Second)
    
    // Verify metrics are being collected
    Expect(mon.Latency()).To(BeNumerically(">", 0))
    Expect(mon.Uptime()).To(BeNumerically(">", 0))
    
    // Latency should be reasonable
    Expect(mon.Latency()).To(BeNumerically("<", 1*time.Second))
})
```

### 5. Configuration Testing

Test configuration validation and normalization:

```go
It("should normalize configuration values", func() {
    cfg := types.Config{
        CheckTimeout:  1 * time.Second,  // Below minimum
        IntervalCheck: 500 * time.Millisecond,  // Below minimum
        FallCountKO:   0,  // Below minimum
    }
    
    Expect(mon.SetConfig(ctx, cfg)).To(Succeed())
    
    result := mon.GetConfig()
    Expect(result.CheckTimeout).To(BeNumerically(">=", 5*time.Second))
    Expect(result.IntervalCheck).To(BeNumerically(">=", 1*time.Second))
    Expect(result.FallCountKO).To(BeNumerically(">=", 1))
})
```

---

## Package-Specific Tests

### Monitor Core Package

**Focus Areas**:
- Status transitions (OK ↔ Warn ↔ KO)
- Health check execution with timeouts
- Metrics collection accuracy
- Lifecycle management (start/stop/restart)
- Middleware chain execution
- Configuration validation

**Test Files**:
- `monitor_test.go`: Core functionality
- `lifecycle_test.go`: Start/stop/restart
- `transitions_test.go`: Status transitions
- `metrics_test.go`: Metrics tracking
- `integration_test.go`: End-to-end scenarios
- `security_test.go`: Security and edge cases

### Pool Package

**Focus Areas**:
- Monitor CRUD operations
- Batch start/stop/restart
- Prometheus metrics aggregation
- Shell command execution
- Thread-safe concurrent access

**Test Files**:
- `pool_test.go`: Basic operations
- `pool_metrics_test.go`: Metrics collection
- `pool_shell_test.go`: Shell commands
- `pool_coverage_test.go`: Edge cases
- `pool_errors_test.go`: Error handling

**Coverage Script**:
```bash
cd monitor/pool
./test_coverage.sh          # Basic
./test_coverage.sh --html   # HTML report
./test_coverage.sh --race   # Race detection
```

### Info Package

**Focus Areas**:
- Dynamic name generation
- Info data caching
- Lazy evaluation
- Thread-safe access
- Encoding formats

**Test Files**:
- `info_test.go`: Core functionality
- `encode_test.go`: Encoding formats
- `integration_test.go`: Real-world scenarios
- `edge_cases_test.go`: Corner cases
- `security_test.go`: Security validation

### Status Package

**Focus Areas**:
- Status enumeration values
- String parsing and formatting
- Multi-format encoding (JSON, Text, XML)
- Type safety
- Comparison operations

**Test Files**:
- `status_test.go`: Basic operations
- `encoding_test.go`: Format encoding
- `parse_test.go`: String parsing
- `json_test.go`: JSON marshaling
- `format_test.go`: Custom formatting

---

## Writing Tests

### Test Structure

Follow Ginkgo BDD style:

```go
var _ = Describe("Component", func() {
    var (
        mon Monitor
        ctx context.Context
        cnl context.CancelFunc
    )
    
    BeforeEach(func() {
        // Setup before each test
        ctx, cnl = context.WithTimeout(context.Background(), 10*time.Second)
        mon = createTestMonitor()
    })
    
    AfterEach(func() {
        // Cleanup after each test
        if mon != nil && mon.IsRunning() {
            mon.Stop(ctx)
        }
        if cnl != nil {
            cnl()
        }
    })
    
    Describe("Feature", func() {
        Context("when condition", func() {
            It("should behave correctly", func() {
                // Test implementation
                Expect(mon.Start(ctx)).To(Succeed())
                Eventually(mon.IsRunning).Should(BeTrue())
            })
        })
    })
})
```

### Test Naming

Use descriptive names:

```go
// Good
It("should transition from KO to Warn after 3 successful checks")
It("should handle concurrent start and stop operations")
It("should respect health check timeout")

// Bad
It("works")
It("test1")
It("check status")
```

### Assertions

Use appropriate matchers:

```go
// Success/failure
Expect(err).ToNot(HaveOccurred())
Expect(err).To(HaveOccurred())

// Equality
Expect(status).To(Equal(status.OK))
Expect(name).To(Equal("monitor"))

// Numerical
Expect(count).To(BeNumerically(">=", 1))
Expect(duration).To(BeNumerically("<", time.Second))

// Collections
Expect(list).To(ContainElement("item"))
Expect(list).To(HaveLen(3))
Expect(list).To(BeEmpty())

// Async (Eventually/Consistently)
Eventually(mon.IsRunning, "2s", "100ms").Should(BeTrue())
Consistently(mon.Status, "1s", "100ms").Should(Equal(status.OK))
```

---

## Best Practices

### 1. Test Independence

Each test should be independent:

```go
// Good - Independent
It("test A", func() {
    mon := createMonitor()
    mon.Start(ctx)
    defer mon.Stop(ctx)
    // test logic
})

It("test B", func() {
    mon := createMonitor()  // Fresh instance
    mon.Start(ctx)
    defer mon.Stop(ctx)
    // test logic
})

// Bad - Tests depend on each other
var mon Monitor
It("test A", func() {
    mon = createMonitor()
    mon.Start(ctx)
})
It("test B", func() {
    // Assumes mon from test A
    Expect(mon.IsRunning()).To(BeTrue())
})
```

### 2. Cleanup

Always cleanup resources:

```go
AfterEach(func() {
    if mon != nil && mon.IsRunning() {
        mon.Stop(ctx)
    }
    if cnl != nil {
        cnl()
    }
})

// Or use defer in test
It("should...", func() {
    mon.Start(ctx)
    defer mon.Stop(ctx)
    // test logic
})
```

### 3. Timeouts

Use realistic timeouts:

```go
// Good - Reasonable timeouts
ctx, cnl := context.WithTimeout(context.Background(), 10*time.Second)
Eventually(mon.IsRunning, "2s", "100ms").Should(BeTrue())

// Bad - Too short or too long
ctx, cnl := context.WithTimeout(context.Background(), 100*time.Millisecond)
Eventually(mon.IsRunning, "30s").Should(BeTrue())  // Too slow
```

### 4. Race-Free Tests

Avoid race conditions:

```go
// Good - Synchronized
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        _ = mon.Status()
    }()
}
wg.Wait()

// Bad - Unsynchro nized
for i := 0; i < 10; i++ {
    go func() {
        _ = mon.Status()
    }()
}
// No wait - test may finish before goroutines
```

### 5. Test Data

Use factories or builders:

```go
// Good - Test helpers
func createTestMonitor(name string) Monitor {
    inf, _ := info.New(name)
    mon, _ := monitor.New(ctx, inf)
    mon.SetConfig(ctx, testConfig())
    return mon
}

func testConfig() types.Config {
    return types.Config{
        CheckTimeout:  5 * time.Second,
        IntervalCheck: 30 * time.Second,
        FallCountKO:   2,
        RiseCountKO:   2,
    }
}

// Usage
mon := createTestMonitor("test-mon")
```

---

## Troubleshooting

### Common Issues

**1. Flaky Tests (Timing Issues)**

```go
// Problem: Test fails intermittently
It("should start quickly", func() {
    mon.Start(ctx)
    Expect(mon.IsRunning()).To(BeTrue())  // Sometimes fails
})

// Solution: Use Eventually
It("should start quickly", func() {
    mon.Start(ctx)
    Eventually(mon.IsRunning, "2s", "100ms").Should(BeTrue())
})
```

**2. Race Detector Failures**

```bash
# Run with race detector to find issues
CGO_ENABLED=1 go test -race ./...

# Common fixes:
# - Use sync.Mutex for shared state
# - Use atomic.Value for lock-free reads
# - Properly synchronize goroutines
```

**3. Deadlocks**

```go
// Problem: Test hangs
It("test", func() {
    mon.Start(ctx)
    // Forgot to stop - may deadlock on cleanup
})

// Solution: Always stop
It("test", func() {
    mon.Start(ctx)
    defer mon.Stop(ctx)
})
```

**4. Context Cancellation**

```go
// Problem: Context cancelled too early
ctx, cnl := context.WithTimeout(bg, 100*time.Millisecond)
defer cnl()
mon.Start(ctx)  // Context may expire during test

// Solution: Use longer timeout
ctx, cnl := context.WithTimeout(bg, 10*time.Second)
defer cnl()
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
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Check coverage
        run: |
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | \
          awk '{if ($1 < 70) exit 1}'
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Makefile Example

```makefile
.PHONY: test test-race test-cover test-all

test:
	go test ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-all: test-race test-cover
	@echo "All tests passed!"

test-ci:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
```

---

## Summary

- **Always run with `-race`** to detect concurrency issues
- **Use `Eventually`/`Consistently`** for timing-dependent tests
- **Cleanup resources** in `AfterEach` or with `defer`
- **Write independent tests** that don't depend on execution order
- **Target ≥70% coverage** for production code
- **Test error paths** and edge cases, not just happy paths
- **Use test helpers** to reduce duplication

For package-specific details, see the test files in each subpackage.
