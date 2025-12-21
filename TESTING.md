# Testing Documentation

Comprehensive testing guide for the `github.com/nabbar/golib` library and all its subpackages.

---

## Table of Contents

- [Test Suite Statistics](#test-suite-statistics)
- [Quick Start](#quick-start)
- [Test Framework](#test-framework)
- [Running Tests](#running-tests)
  - [Basic Testing](#basic-testing)
  - [Race Detection](#race-detection)
  - [Coverage Analysis](#coverage-analysis)
  - [Package-Specific Testing](#package-specific-testing)
- [Coverage Report](#coverage-report)
  - [High Coverage Packages](#high-coverage-packages)
  - [Packages Needing Improvement](#packages-needing-improvement)
  - [Untested Packages](#untested-packages)
- [Writing Tests](#writing-tests)
  - [Test Structure](#test-structure)
  - [Ginkgo v2 Guidelines](#ginkgo-v2-guidelines)
  - [Gomega Matchers](#gomega-matchers)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [CI Integration](#ci-integration)

---

## Test Suite Statistics

**Latest Test Run Results** (from `./coverage-report.sh`):

```
Total Packages:           165
Packages with Tests:      131 (79.4%)
Packages without Tests:   34 (20.6%)

Test Specifications:      11,818
Test Assertions:          23,080
Benchmarks:               151
Pending Tests:            18
Skipped Tests:            6

Average Coverage:         75.49%
Packages ≥80%:            71/131 (54.2%)
Packages at 100%:         16/131 (12.2%)

Race Conditions:          0 (verified with CGO_ENABLED=1 go test -race)
Thread Safety:            ✅ All concurrent operations validated
```

**Coverage Distribution:**

| Range | Count | Percentage | Examples |
|-------|-------|------------|----------|
| 100% | 16 | 12.2% | errors/pool, logger/gorm, router/authheader, semaphore/sem |
| 90-99% | 27 | 20.6% | atomic, version, size, prometheus/metrics |
| 80-89% | 28 | 21.4% | ioutils, mail/queuer, context, runner |
| 70-79% | 16 | 12.2% | cobra, viper, socket/client/* |
| 60-69% | 9 | 6.9% | config, logger, database/kvmap |
| <60% | 35 | 26.7% | archive, aws, httpserver |

---

## Quick Start

### Running All Tests

```bash
# Standard test run (all packages)
go test ./...

# Verbose output with details
go test -v ./...

# With race detector (recommended)
CGO_ENABLED=1 go test -race ./...

# With coverage
go test -cover ./...

# Complete test suite (as used in CI)
go test -timeout=10m -v -cover -covermode=atomic ./...
```

### Using Coverage Report Script

```bash
# Run comprehensive coverage analysis
./coverage-report.sh

# Output includes:
# - Coverage statistics per package
# - Packages without tests
# - Packages below 80% coverage
# - Recommendations for improvement
```

### Expected Output

```
Total Packages:       165
Packages with Tests:  127
Test Specifications:  10,964
Average Coverage:     73.9%

PACKAGES WITHOUT TESTS
• archive/archive
• aws/bucket
• config/const
• ...

PACKAGES BELOW 80% COVERAGE
• archive                     8.60%
• artifact                   23.40%
• aws                         5.40%
• ...
```

---

## Test Framework

### Ginkgo v2

Behavior-driven development (BDD) testing framework used across all subpackages.

**Key Features:**
- Spec organization with `Describe`, `Context`, `It`
- `BeforeEach` / `AfterEach` for setup/teardown
- `BeforeAll` / `AfterAll` for suite-level setup
- Ordered specs for sequential tests
- Focused specs (`FIt`, `FContext`) for debugging
- `Eventually` / `Consistently` for async assertions
- Table-driven tests with `DescribeTable`

**Installation:**
```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

**Documentation:** [Ginkgo v2 Docs](https://onsi.github.io/ginkgo/)

### Gomega

Matcher library for expressive assertions.

**Common Matchers:**
- `Expect(x).To(Equal(y))` - equality
- `Expect(err).ToNot(HaveOccurred())` - error checking
- `Expect(x).To(BeNumerically(">=", y))` - numeric comparison
- `Expect(ch).To(BeClosed())` - channel state
- `Eventually(func)` - async assertion
- `Consistently(func)` - sustained assertion

**Documentation:** [Gomega Docs](https://onsi.github.io/gomega/)

### gmeasure

Performance measurement for Ginkgo tests (used in several packages).

**Usage Example:**
```go
experiment := gmeasure.NewExperiment("Operation Name")
AddReportEntry(experiment.Name, experiment)

experiment.Sample(func(idx int) {
    experiment.MeasureDuration("metric_name", func() {
        // Code to measure
    })
}, gmeasure.SamplingConfig{N: 100, Duration: 5 * time.Second})

stats := experiment.GetStats("metric_name")
```

**Packages Using gmeasure:**
- ioutils/aggregator (performance benchmarks)
- ioutils/multi (write operation metrics)
- monitor/* (system metrics)

**Documentation:** [gmeasure Package](https://pkg.go.dev/github.com/onsi/gomega/gmeasure)

---

## Running Tests

### Basic Testing

```bash
# Run all tests in all packages
go test ./...

# Verbose output (recommended for CI)
go test -v ./...

# Run specific package
go test ./logger

# Run with timeout (important for long-running tests)
go test -timeout 5m ./...

# Skip long-running tests
go test -short ./...

# Run tests matching pattern
go test -run TestLogger ./logger

# With Ginkgo focus
go test -ginkgo.focus="should handle concurrent writes" ./ioutils/aggregator
```

### Race Detection

**Critical for concurrency testing:**

```bash
# Enable race detector (all packages)
CGO_ENABLED=1 go test -race ./...

# Verbose with race detection
CGO_ENABLED=1 go test -race -v ./...

# Full suite with race detection (CI command)
CGO_ENABLED=1 go test -race -timeout=10m -v -cover -covermode=atomic ./...

# Specific package with race detector
CGO_ENABLED=1 go test -race ./ioutils/aggregator
```

**Note:** Race detector adds ~10x overhead. Some tests may take longer.

**Results:** Zero data races detected across all 10,735 specs.

### Coverage Analysis

```bash
# Coverage percentage for all packages
go test -cover ./...

# Coverage profile
go test -coverprofile=coverage.out ./...

# HTML coverage report
go tool cover -html=coverage.out

# Coverage by function
go tool cover -func=coverage.out

# Atomic coverage mode (for race detector)
go test -covermode=atomic -coverprofile=coverage.out ./...

# Per-package coverage
go test -cover ./logger
go test -cover ./ioutils/aggregator
go test -cover ./mail/queuer
```

### Package-Specific Testing

**High-Coverage Packages:**

```bash
# ioutils (87.7% average, 772 specs)
go test -v -cover ./ioutils/...

# mail (89.0% average, 970 specs)
go test -v -cover ./mail/...

# errors (87.6%, 305 specs)
go test -v -cover ./errors/...

# version (93.8%, 173 specs)
go test -v -cover ./version/...
```

**Packages Needing More Tests:**

```bash
# archive (8.6%, 89 specs) - needs improvement
go test -v -cover ./archive/...

# aws (5.4%, 220 specs) - needs improvement
go test -v -cover ./aws/...

# httpserver (52.5%, 84 specs) - moderate coverage
go test -v -cover ./httpserver/...
```

---

## Coverage Report

### Coverage Report Script

The repository includes `coverage-report.sh`, a comprehensive script that analyzes test coverage across all packages.

**Usage:**

```bash
# Run full coverage analysis
./coverage-report.sh

# Output is also saved to a file (optional)
./coverage-report.sh > coverage-full.txt
```

**What it provides:**

- **Overall Statistics**: Total packages, tested packages, average coverage
- **Per-Package Metrics**: Coverage %, specs count, assertions, benchmarks
- **Issue Detection**: Packages without tests, packages below 80% coverage
- **Detailed Breakdown**: Test execution time, pending tests, skipped tests

**Output Example:**

```
Total Packages:       165
Packages with Tests:  127 (77.0%)
Test Specifications:  10,964
Average Coverage:     73.9%

PACKAGES WITHOUT TESTS
• archive/archive
• aws/bucket
...

PACKAGES BELOW 80% COVERAGE
• archive                     8.60%
• aws                         5.40%
...
```

This script is used to generate all coverage statistics shown in this document and the main README.

---

### High Coverage Packages (≥90%)

**Packages at 100% Coverage:**

| Package | Specs | Assertions | Notes |
|---------|-------|------------|-------|
| errors/pool | 83 | 122 | Thread-safe error pooling |
| httpserver/types | 32 | 53 | Type definitions |
| ioutils/bufferReadCloser | 57 | 138 | Buffered reader with closer |
| ioutils/delim | 198 | 329 | Delimiter-based stream processing |
| ioutils/iowrapper | 114 | 179 | Generic I/O wrappers |
| ioutils/nopwritecloser | 54 | 140 | No-op writer closer |
| logger/gorm | 34 | 76 | GORM logger integration |
| logger/hookstderr | 30 | 64 | Stderr output hook |
| logger/hookstdout | 30 | 64 | Stdout output hook |
| monitor/info | 95 | 262 | System information collection |
| prometheus/types | 36 | 112 | Prometheus type definitions |
| router/authheader | 11 | 29 | Authorization header parsing |
| semaphore/sem | 66 | 117 | Semaphore implementation |
| semaphore | 33 | 55 | Semaphore base |

**Packages 90-99% Coverage:**

- **artifact/client**: 98.6% (21 specs) - Artifact client interface
- **mail/smtp/tlsmode**: 98.8% (165 specs) - SMTP TLS mode handling
- **monitor/status**: 98.4% (181 specs) - Status reporting
- **network/protocol**: 98.7% (298 specs) - Network protocol helpers
- **cache/item**: 96.7% (21 specs) - Cache item implementation
- **logger/hashicorp**: 96.6% (89 specs) - Hashicorp logger adapter
- **router/auth**: 96.3% (12 specs) - Authentication middleware
- **semaphore/bar**: 96.6% (68 specs) - Semaphore with progress bar
- **prometheus/metrics**: 95.5% (179 specs) - Custom metrics
- **status/control**: 95.0% (102 specs) - Status control
- **size**: 95.4% (352 specs) - Byte size arithmetic
- **prometheus/bloom**: 94.7% (45 specs) - Bloom filter metrics
- **version**: 93.8% (173 specs) - Semantic versioning
- **mail/smtp/config**: 92.7% (222 specs) - SMTP configuration
- **atomic**: 91.8% (49 specs) - Generic atomic types
- **duration**: 91.5% (179 specs) - Duration extensions
- **encoding/aes**: 91.5% (126 specs) - AES encryption
- **router**: 91.0% (61 specs) - Gin-based router
- **duration/big**: 91.0% (250 specs) - Big integer duration
- **mail/queuer**: 90.8% (102 specs) - Email queuing
- **mail/smtp**: 90.1% (104 specs) - SMTP client
- **logger/hookwriter**: 90.2% (31 specs) - Generic writer hook
- **runner/ticker**: 90.2% (88 specs) - Ticker management

### Packages Needing Improvement (<40%)

**Critical Priority (0-20% coverage):**

| Package | Coverage | Specs | Status |
|---------|----------|-------|--------|
| artifact/s3aws | 2.0% | 1 | Needs tests |
| aws | 5.4% | 220 | Partial tests |
| artifact/jfrog | 6.1% | 2 | Needs tests |
| ftpclient | 6.2% | 22 | Needs tests |
| archive | 8.6% | 89 | Needs extensive tests |
| artifact/github | 8.6% | 1 | Needs tests |
| artifact/gitlab | 13.5% | 2 | Needs tests |
| database/gorm | 19.6% | 41 | Needs improvement |
| logger/hookfile | 19.6% | 22 | Needs improvement |

**Medium Priority (20-40% coverage):**

- artifact (23.4%)
- database/kvdriver (38.4%)
- config/components/aws (40.7%)
- config/components/database (39.0%)

### Untested Packages

**38 packages without test files:**

Infrastructure packages (primarily type definitions and utilities):
- archive/archive, archive/archive/tar, archive/archive/types, archive/archive/zip
- archive/compress, archive/helper
- aws/bucket, aws/configAws, aws/configCustom, aws/group
- aws/helper, aws/http, aws/multipart, aws/object
- aws/policy, aws/pusher, aws/role, aws/user
- config/const, config/types
- database/kvtypes
- encoding (base package)
- httpserver/testhelpers
- ioutils/maxstdio
- ldap, monitor/types, nats, oauth
- pidcontroller, pprof, prometheus/webmetrics
- request, runner (base package)
- semaphore/types
- socket (base package), socket/client, socket/config, socket/server

**Note:** Many untested packages are interface definitions, constants, or types packages that may not require separate tests if covered by parent package tests.

---

## Writing Tests

### Test Structure

**File Organization:**

Each package follows this structure:

```
package/
├── package_suite_test.go    - Suite setup and global helpers
├── feature_test.go           - Feature-specific tests
├── concurrency_test.go       - Concurrency tests (if applicable)
├── errors_test.go            - Error handling tests
├── benchmark_test.go         - Performance benchmarks
└── example_test.go           - Runnable examples
```

**Test Template:**

```go
var _ = Describe("ComponentName", func() {
    var (
        component ComponentType
        ctx       context.Context
        cancel    context.CancelFunc
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
        component = New(...)
    })

    AfterEach(func() {
        if component != nil {
            component.Close()
        }
        cancel()
        time.Sleep(10 * time.Millisecond)  // Cleanup grace period
    })

    Context("when testing feature X", func() {
        It("should behave correctly", func() {
            // Test code
            Expect(result).To(Equal(expected))
        })
    })
})
```

### Ginkgo v2 Guidelines

**Spec Organization:**

```go
Describe("Top-level component", func() {
    Context("when condition A", func() {
        It("should do X", func() {
            // Test
        })
        
        It("should do Y", func() {
            // Test
        })
    })
    
    Context("when condition B", func() {
        It("should do Z", func() {
            // Test
        })
    })
})
```

**Async Testing:**

```go
// Use Eventually for async operations
Eventually(func() bool {
    return component.IsReady()
}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

// Use Consistently for sustained conditions
Consistently(func() bool {
    return component.IsRunning()
}, 1*time.Second, 50*time.Millisecond).Should(BeTrue())
```

### Gomega Matchers

**Common Patterns:**

```go
// Error checking
Expect(err).ToNot(HaveOccurred())
Expect(err).To(MatchError("expected error"))

// Equality
Expect(value).To(Equal(expected))
Expect(value).To(BeNumerically(">=", minimum))

// Collections
Expect(slice).To(ContainElement(item))
Expect(slice).To(HaveLen(5))
Expect(map).To(HaveKey("key"))

// Types
Expect(value).To(BeNil())
Expect(value).To(BeAssignableToTypeOf(Type{}))

// Channels
Expect(ch).To(BeClosed())
Expect(ch).To(Receive(&value))
```

---

## Best Practices

### ✅ DO

**1. Use `Eventually` for Async Operations:**
```go
// ✅ GOOD: Wait for condition
Eventually(func() bool {
    return server.IsRunning()
}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

// ❌ BAD: Fixed sleep
time.Sleep(100 * time.Millisecond)
Expect(server.IsRunning()).To(BeTrue())
```

**2. Protect Shared State:**
```go
// ✅ GOOD: Thread-safe access
var (
    mu    sync.Mutex
    count int
)

writer := func(p []byte) (int, error) {
    mu.Lock()
    defer mu.Unlock()
    count++
    return len(p), nil
}

// ❌ BAD: Race condition
var count int
writer := func(p []byte) (int, error) {
    count++  // RACE!
    return len(p), nil
}
```

**3. Clean Up Resources:**
```go
// ✅ GOOD: Always cleanup
AfterEach(func() {
    if component != nil {
        component.Close()
    }
    cancel()
    time.Sleep(50 * time.Millisecond)
})

// ❌ BAD: No cleanup
AfterEach(func() {
    cancel()  // Missing Close()
})
```

**4. Use Descriptive Test Names:**
```go
// ✅ GOOD: Clear intent
It("should return error when connection is closed", func() {
    // ...
})

// ❌ BAD: Vague
It("test error", func() {
    // ...
})
```

**5. Test Edge Cases:**
```go
// Test nil, empty, boundary values
Context("with nil input", func() {
    It("should handle gracefully", func() {
        err := component.Process(nil)
        Expect(err).To(HaveOccurred())
    })
})

Context("with empty string", func() {
    It("should return appropriate error", func() {
        err := component.Validate("")
        Expect(err).To(MatchError("empty input"))
    })
})
```

### ❌ DON'T

**1. Don't Ignore Race Detector Warnings:**
```bash
# Always run with race detector during development
CGO_ENABLED=1 go test -race ./...
```

**2. Don't Use Fixed Timeouts:**
```go
// ❌ BAD: Brittle on slow systems
time.Sleep(100 * time.Millisecond)

// ✅ GOOD: Adaptive waiting
Eventually(condition, timeout, interval).Should(BeTrue())
```

**3. Don't Share State Between Tests:**
```go
// ❌ BAD: Global state
var globalCounter int

It("test 1", func() {
    globalCounter++  // Affects other tests!
})

// ✅ GOOD: Isolated state
var counter int
BeforeEach(func() {
    counter = 0  // Reset for each test
})
```

**4. Don't Skip Error Checking:**
```go
// ❌ BAD: Ignoring errors
result, _ := operation()

// ✅ GOOD: Check all errors
result, err := operation()
Expect(err).ToNot(HaveOccurred())
```

---

## Troubleshooting

### Common Issues

**1. Test Timeout**

```
Error: test timed out after 10m0s
```

**Solution:**
- Increase timeout: `go test -timeout=20m`
- Check for deadlocks in code
- Ensure cleanup completes
- Review Eventually timeouts

**2. Race Condition Detected**

```
WARNING: DATA RACE
```

**Solution:**
- Protect shared variables with mutex
- Use atomic operations
- Review concurrent access patterns
- Add proper synchronization

**3. Flaky Tests**

```
Random failures, not reproducible
```

**Solution:**
- Increase `Eventually` timeouts
- Add proper synchronization
- Run with `-race` to detect issues
- Check resource cleanup
- Avoid fixed `time.Sleep`

**4. Coverage Gaps**

```
coverage: 75.0% (below target 80%)
```

**Solution:**
- Run `go tool cover -html=coverage.out`
- Identify uncovered branches
- Add edge case tests
- Test error paths
- Review package-specific coverage report

**5. Import Cycle**

```
import cycle not allowed
```

**Solution:**
- Refactor packages to break cycle
- Extract common interface
- Use dependency injection
- Move shared types to separate package

### Debug Techniques

**Enable Verbose Output:**
```bash
go test -v ./...
go test -v -ginkgo.v ./package
```

**Focus Specific Test:**
```bash
go test -ginkgo.focus="should handle concurrent writes" ./package
go test -run TestSpecificFunction ./package
```

**Check Goroutine Leaks:**
```go
BeforeEach(func() {
    runtime.GC()
    initialGoroutines = runtime.NumGoroutine()
})

AfterEach(func() {
    runtime.GC()
    time.Sleep(100 * time.Millisecond)
    leaked := runtime.NumGoroutine() - initialGoroutines
    Expect(leaked).To(BeNumerically("<=", 1))
})
```

**Profile Tests:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./package
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./package
go tool pprof mem.prof
```

---

## CI Integration

### GitHub Actions

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.24', '1.25', '1.26']
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Test
        run: go test -timeout=10m -v -cover -covermode=atomic ./...
      
      - name: Race Detection
        run: CGO_ENABLED=1 go test -race -timeout=10m -v ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
```

### GitLab CI

```yaml
test:
  image: golang:1.26
  stage: test
  script:
    - go test -timeout=10m -v -cover -covermode=atomic ./...
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

race:
  image: golang:1.26
  stage: test
  script:
    - CGO_ENABLED=1 go test -race -timeout=10m -v ./...

coverage:
  image: golang:1.26
  stage: test
  script:
    - ./coverage-report.sh
    - go tool cover -func=coverage.out
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+)%/'
```

### Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running golib tests..."

go test -timeout=2m ./...
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running race detector..."
CGO_ENABLED=1 go test -race -timeout=5m ./...
if [ $? -ne 0 ]; then
    echo "Race conditions detected. Commit aborted."
    exit 1
fi

echo "Checking coverage..."
COVERAGE=$(./coverage-report.sh | grep "Average Coverage" | awk '{print $4}' | tr -d '%')
if (( $(echo "$COVERAGE < 70.0" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 70%. Commit aborted."
    exit 1
fi

echo "All checks passed!"
exit 0
```

---

**Test Suite Maintained By**: [Nicolas JUHEL](https://github.com/nabbar)  
**Framework**: Ginkgo v2 / Gomega / gmeasure  
**Coverage Target**: ≥80% per package  
**Last Updated**: Based on coverage-report.sh analysis
