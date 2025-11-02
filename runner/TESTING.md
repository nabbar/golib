# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-100%2B%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A5%2080%25-brightgreen)]()

Comprehensive testing documentation for the runner package, covering test execution, race detection, concurrency testing, and quality assurance.

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

The runner package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions and timing-sensitive operations.

**Test Suite**
- Total Specs: 100+
- Coverage: ≥80%
- Race Detection: ✅ Zero data races
- Execution Time: ~3s (without race), ~8s (with race)

**Coverage Areas**
- Lifecycle operations (Start, Stop, Restart)
- Uptime and running state tracking
- Error collection and retrieval
- Context cancellation and timeouts
- Concurrent operations and race conditions
- Edge cases (nil contexts, panics, quick exits)
- Exponential backoff cleanup detection

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

# Using Ginkgo CLI with race detection
CGO_ENABLED=1 ginkgo -race -cover

# Run specific subpackage
go test ./startStop
go test ./ticker
```

---

## Test Framework

**Ginkgo v2** - BDD testing framework ([docs](https://onsi.github.io/ginkgo/))
- Hierarchical test organization (`Describe`, `Context`, `It`)
- Setup/teardown hooks (`BeforeEach`, `AfterEach`)
- `Eventually()` for time-sensitive assertions
- Parallel execution support
- Rich CLI with filtering

**Gomega** - Matcher library ([docs](https://onsi.github.io/gomega/))
- Readable assertion syntax
- `Eventually()` for polling assertions
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

# Specific subpackage
ginkgo ./startStop
ginkgo ./ticker

# Pattern matching
ginkgo --focus="Lifecycle"
ginkgo --focus="Error"

# Repeat for stability testing
ginkgo --repeat=10

# JUnit report for CI
ginkgo --junit-report=results.xml

# With labels
ginkgo --label-filter="concurrency"
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race

# Stress test with repetition
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Validates**:
- Atomic operations (`atomic.Value`)
- Mutex protection (`sync.Mutex`)
- Context cancellation races
- Goroutine synchronization
- State transition thread safety

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/runner/startStop	6.234s
ok  	github.com/nabbar/golib/runner/ticker   	7.891s

# ❌ Race detected
WARNING: DATA RACE
Read at 0x... by goroutine ...
```

**Status**: Zero data races detected

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

# Trace for detailed analysis
go test -trace=trace.out ./...
go tool trace trace.out
```

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite | ~3s | Without race |
| With `-race` | ~8s | 2-3x slower (normal) |
| Individual Spec | <100ms | Most tests |
| Timing Tests | 200-500ms | Uses time.Sleep() |

---

## Test Coverage

**Target**: ≥80% statement coverage

### Coverage By Component

| Component | Files | Test Files | Description |
|-----------|-------|------------|-------------|
| **Root Package** | `interface.go`, `tools.go` | N/A | Core interfaces and utilities |
| **startStop** | `interface.go`, `model.go` | 5 test files | Service lifecycle management |
| **ticker** | `interface.go`, `model.go` | 4 test files | Periodic execution |

### startStop Test Files

| File | Purpose | Key Specs |
|------|---------|-----------|
| `lifecycle_test.go` | Start, Stop, Restart operations | 15+ specs |
| `concurrency_test.go` | Concurrent operations, race conditions | 10+ specs |
| `errors_test.go` | Error collection and handling | 12+ specs |
| `uptime_test.go` | Uptime tracking and IsRunning | 8+ specs |
| `construction_test.go` | Constructor and edge cases | 6+ specs |

### ticker Test Files

| File | Purpose | Key Specs |
|------|---------|-----------|
| `lifecycle_test.go` | Start, Stop, Restart operations | 18+ specs |
| `concurrency_test.go` | Concurrent operations, race conditions | 15+ specs |
| `errors_test.go` | Error collection from ticker function | 14+ specs |
| `edge_cases_test.go` | Edge cases, nil handling, panics | 20+ specs |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test -cover ./startStop
go test -cover ./ticker
```

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
var _ = Describe("Component/Feature", func() {
    var (
        ctx    context.Context
        cancel context.CancelFunc
    )

    BeforeEach(func() {
        // Setup: Create test context
        ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    })

    AfterEach(func() {
        // Cleanup: Cancel context, stop runners
        if cancel != nil {
            cancel()
        }
    })

    Context("Specific scenario", func() {
        It("should behave as expected", func() {
            // Arrange
            runner := startStop.New(startFunc, stopFunc)
            
            // Act
            err := runner.Start(ctx)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Eventually(runner.IsRunning).Should(BeTrue())
        })
    })
})
```

---

## Thread Safety

Thread safety is critical for the runner package's concurrent operations.

### Concurrency Primitives

```go
// Atomic values for lock-free reads
libatm.Value[time.Time]          // Start time (zero = not running)
libatm.Value[context.CancelFunc] // Cancel function

// Mutex for operation serialization
sync.Mutex                        // Protects Start/Stop/Restart

// Exponential backoff polling
time.Sleep(1ms → 2ms → 4ms → 8ms → 10ms max)
```

### Verified Components

| Component | Mechanism | Verification | Status |
|-----------|-----------|--------------|--------|
| `startStop.run` | `sync.Mutex` + `atomic.Value` | Race detector | ✅ Race-free |
| `ticker.run` | `sync.Mutex` + `atomic.Value` | Race detector | ✅ Race-free |
| State Transitions | Atomic operations | Concurrency tests | ✅ Thread-safe |
| Error Collection | `errpol.Pool` (thread-safe) | Concurrent writes | ✅ Race-free |

### Testing Thread Safety

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Concurrency" ./...

# Stress test: repeat 20 times
CGO_ENABLED=1 ginkgo -race --repeat=20

# Parallel execution (if supported)
CGO_ENABLED=1 ginkgo -race -p
```

**Result**: Zero data races across all test runs

### Common Race Patterns Tested

**Pattern 1: Concurrent Start/Stop**
```go
It("should handle concurrent start and stop", func() {
    runner := startStop.New(startFunc, stopFunc)
    var wg sync.WaitGroup
    
    // Start and stop from multiple goroutines
    for i := 0; i < 10; i++ {
        wg.Add(2)
        go func() {
            defer wg.Done()
            runner.Start(ctx)
        }()
        go func() {
            defer wg.Done()
            runner.Stop(ctx)
        }()
    }
    wg.Wait()
})
```

**Pattern 2: State Check During Transitions**
```go
It("should provide consistent state during operations", func() {
    runner := startStop.New(startFunc, stopFunc)
    runner.Start(ctx)
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Safe concurrent reads
            _ = runner.IsRunning()
            _ = runner.Uptime()
        }()
    }
    wg.Wait()
})
```

**Pattern 3: Error Collection Under Concurrency**
```go
It("should collect errors safely", func() {
    tick := ticker.New(10*time.Millisecond, func(ctx context.Context, t *time.Ticker) error {
        return fmt.Errorf("error at %v", time.Now())
    })
    
    tick.Start(ctx)
    time.Sleep(100 * time.Millisecond)
    
    // Concurrent error retrieval
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = tick.ErrorsLast()
            _ = tick.ErrorsList()
        }()
    }
    wg.Wait()
    tick.Stop(ctx)
})
```

---

## Test Categories

### Lifecycle Tests

**Purpose**: Verify basic Start/Stop/Restart operations

**Key Scenarios**:
- Start with blocking function
- Start with quick-exiting function
- Stop idempotency (stop when not running)
- Restart atomic operation
- Multiple start/stop cycles

**Example**:
```go
It("should start successfully with blocking function", func() {
    var started atomic.Bool
    
    start := func(ctx context.Context) error {
        started.Store(true)
        <-ctx.Done() // Block until stopped
        return nil
    }
    
    runner := New(start, stopFunc)
    err := runner.Start(ctx)
    
    Expect(err).ToNot(HaveOccurred())
    Eventually(started.Load).Should(BeTrue())
    Eventually(runner.IsRunning).Should(BeTrue())
})
```

### Concurrency Tests

**Purpose**: Verify thread safety under concurrent operations

**Key Scenarios**:
- Concurrent Start/Stop calls
- Start from multiple goroutines
- State checks during transitions
- Rapid start/stop cycles
- Stop from multiple goroutines

**Example**:
```go
It("should handle concurrent stop calls", func() {
    runner := New(startFunc, stopFunc)
    runner.Start(ctx)
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = runner.Stop(ctx)
        }()
    }
    wg.Wait()
    
    Expect(runner.IsRunning()).To(BeFalse())
})
```

### Error Tests

**Purpose**: Verify error collection and retrieval

**Key Scenarios**:
- Collect errors from start function
- Collect errors from stop function
- Error clearing on new Start
- ErrorsLast returns most recent
- ErrorsList returns all errors
- Nil error handling

**Example**:
```go
It("should collect errors from start function", func() {
    expectedErr := fmt.Errorf("start error")
    
    start := func(ctx context.Context) error {
        return expectedErr
    }
    
    runner := New(start, stopFunc)
    runner.Start(ctx)
    
    Eventually(func() error {
        return runner.ErrorsLast()
    }).Should(Equal(expectedErr))
})
```

### Edge Cases Tests

**Purpose**: Verify behavior in unusual conditions

**Key Scenarios**:
- Nil context handling
- Nil function handling (ticker)
- Invalid duration (< 1ms) handling
- Panic recovery
- Context cancellation
- Quick function exit
- Stop before start completes

**Example**:
```go
It("should handle nil context gracefully", func() {
    runner := New(startFunc, stopFunc)
    err := runner.Start(nil)
    
    Expect(err).To(HaveOccurred())
    Expect(err.Error()).To(ContainSubstring("nil context"))
})

It("should recover from panic in start function", func() {
    start := func(ctx context.Context) error {
        panic("test panic")
    }
    
    runner := New(start, stopFunc)
    
    // Should not crash the process
    Expect(func() {
        runner.Start(ctx)
    }).ToNot(Panic())
})
```

### Uptime Tests

**Purpose**: Verify uptime tracking accuracy

**Key Scenarios**:
- Uptime returns 0 when not running
- Uptime increases while running
- IsRunning matches uptime > 0
- Uptime resets after stop
- Uptime accuracy within tolerance

**Example**:
```go
It("should track uptime accurately", func() {
    runner := New(startFunc, stopFunc)
    
    Expect(runner.Uptime()).To(Equal(time.Duration(0)))
    
    runner.Start(ctx)
    Eventually(runner.IsRunning).Should(BeTrue())
    
    time.Sleep(100 * time.Millisecond)
    
    uptime := runner.Uptime()
    Expect(uptime).To(BeNumerically(">=", 100*time.Millisecond))
    Expect(uptime).To(BeNumerically("<", 200*time.Millisecond))
})
```

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should execute ticker function at regular intervals", func() {
    // Test implementation
})

It("should stop gracefully when context is cancelled", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should collect errors from ticker function", func() {
    // Arrange
    expectedErr := fmt.Errorf("ticker error")
    tick := ticker.New(50*time.Millisecond, func(ctx context.Context, t *time.Ticker) error {
        return expectedErr
    })
    
    // Act
    tick.Start(ctx)
    time.Sleep(100 * time.Millisecond)
    tick.Stop(ctx)
    
    // Assert
    Expect(tick.ErrorsList()).ToNot(BeEmpty())
    Expect(tick.ErrorsLast()).To(Equal(expectedErr))
})
```

**3. Use Appropriate Matchers**
```go
Expect(err).ToNot(HaveOccurred())
Expect(runner.IsRunning()).To(BeTrue())
Eventually(runner.IsRunning).Should(BeTrue())
Expect(uptime).To(BeNumerically(">=", 100*time.Millisecond))
Expect(errors).To(HaveLen(3))
Expect(errors).ToNot(BeEmpty())
```

**4. Use Eventually() for Time-Sensitive Checks**
```go
// ✅ Good: Use Eventually for async operations
runner.Start(ctx)
Eventually(runner.IsRunning).Should(BeTrue())

// ❌ Bad: Immediate check may fail
runner.Start(ctx)
Expect(runner.IsRunning()).To(BeTrue()) // May fail due to timing
```

**5. Always Cleanup**
```go
var runner startStop.StartStop

AfterEach(func() {
    if runner != nil && runner.IsRunning() {
        runner.Stop(ctx)
    }
    if cancel != nil {
        cancel()
    }
})
```

**6. Use Atomic Variables for Concurrency**
```go
// ✅ Good: Thread-safe counter
var counter atomic.Int32

tick := ticker.New(10*time.Millisecond, func(ctx context.Context, t *time.Ticker) error {
    counter.Add(1)
    return nil
})

// ❌ Bad: Race condition
var counter int32  // No atomic protection
tick := ticker.New(10*time.Millisecond, func(ctx context.Context, t *time.Ticker) error {
    counter++ // RACE!
    return nil
})
```

### Test Template

```go
package mypackage_test

import (
    "context"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/nabbar/golib/runner/startStop"
)

var _ = Describe("MyFeature", func() {
    var (
        ctx    context.Context
        cancel context.CancelFunc
        runner startStop.StartStop
    )

    BeforeEach(func() {
        ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    })

    AfterEach(func() {
        if runner != nil && runner.IsRunning() {
            runner.Stop(ctx)
        }
        if cancel != nil {
            cancel()
        }
    })

    Context("When testing feature", func() {
        It("should behave correctly", func() {
            // Arrange
            start := func(ctx context.Context) error {
                <-ctx.Done()
                return nil
            }
            stop := func(ctx context.Context) error {
                return nil
            }
            runner = startStop.New(start, stop)
            
            // Act
            err := runner.Start(ctx)
            
            // Assert
            Expect(err).ToNot(HaveOccurred())
            Eventually(runner.IsRunning).Should(BeTrue())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Create fresh runners for each test
- ✅ Use contexts with timeouts to prevent hangs
- ❌ Don't rely on test execution order
- ❌ Don't share runners between tests

**Timing Considerations**
```go
// ✅ Good: Generous timeouts for CI
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

// ✅ Good: Use Eventually with timeout
Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

// ❌ Bad: Tight timing assumptions
time.Sleep(10 * time.Millisecond)
Expect(runner.IsRunning()).To(BeTrue()) // May fail on slow systems
```

**Assertions**
```go
// ✅ Good: Specific matchers
Expect(err).ToNot(HaveOccurred())
Expect(runner.IsRunning()).To(BeTrue())
Eventually(runner.Uptime).Should(BeNumerically(">", 0))

// ❌ Avoid generic matchers
Expect(runner.IsRunning() == true).To(BeTrue())
```

**Concurrency Testing**
```go
// ✅ Good: Proper synchronization
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            runner.Start(ctx)
        }()
    }
    wg.Wait()
})

// ✅ Always run with -race during development
CGO_ENABLED=1 go test -race ./...
```

**Error Handling**
```go
// ✅ Good: Check all error paths
It("should handle start errors", func() {
    start := func(ctx context.Context) error {
        return fmt.Errorf("start failed")
    }
    runner := New(start, stopFunc)
    runner.Start(ctx)
    
    Eventually(runner.ErrorsLast).ShouldNot(BeNil())
})

// ❌ Bad: Don't ignore error cases
```

**Resource Cleanup**
```go
// ✅ Good: Always cleanup
AfterEach(func() {
    if runner != nil && runner.IsRunning() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        runner.Stop(ctx)
    }
})

// ❌ Bad: Leak goroutines
AfterEach(func() {
    // No cleanup - goroutines keep running
})
```

---

## Troubleshooting

**Test Timeouts**
```bash
# Identify hanging tests
go test -timeout 30s ./...

# Or with ginkgo
ginkgo --timeout=30s
```

Check for:
- Missing context timeouts
- Goroutine leaks (missing Stop calls)
- Deadlocks in start/stop functions
- Infinite loops

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Common race fixes:
- Use `atomic.Value` for reads
- Protect writes with mutex
- Use atomic operations for counters
- Synchronize goroutines with `sync.WaitGroup`

**Flaky Tests**
```bash
# Repeat tests to find flaky ones
ginkgo --repeat=20 --randomize-all
```

Common causes:
- Tight timing assumptions (use Eventually)
- Shared state between tests
- Insufficient sleep/wait times
- Context timeout too short

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: xcode-select --install or brew install gcc

export CGO_ENABLED=1
go test -race ./...
```

**Goroutine Leaks**
```bash
# Use leak detector
go test -v ./... -run TestName

# Check for unclosed runners
# Ensure all tests have cleanup in AfterEach
```

**Debugging**
```bash
# Single test
ginkgo --focus="should start successfully"

# Specific file
go test -v ./startStop -run TestStartStop/Lifecycle

# Verbose output
ginkgo -v --trace

# With GinkgoWriter output
fmt.Fprintf(GinkgoWriter, "Debug: runner state = %v\n", runner.IsRunning())
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
        go-version: ['1.19', '1.20', '1.21']
    
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

**GitLab CI Example**

```yaml
test:
  image: golang:1.21
  script:
    - go test -v ./...
    - CGO_ENABLED=1 go test -race ./...
    - go test -coverprofile=coverage.out ./...
  coverage: '/coverage: \d+.\d+% of statements/'
```

**Pre-commit Hook**

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
go test -cover ./... | grep -E "coverage:" || exit 1

echo "All checks passed!"
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
- [ ] Edge cases covered
- [ ] Test duration reasonable (<10s)
- [ ] No goroutine leaks
- [ ] Documentation updated

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Go Testing](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)
- [sync Package](https://pkg.go.dev/sync)
- [atomic Package](https://pkg.go.dev/sync/atomic)

**Best Practices**
- [Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Writing Testable Code](https://go.dev/blog/examples)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Runner Package Contributors
