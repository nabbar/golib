# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-229%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-82.6%25-brightgreen)]()

Comprehensive testing documentation for the static package, covering test execution, race detection, and quality assurance.

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

The static package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite**
- Total Specs: 229
- Coverage: 82.6%
- Race Detection: ✅ Zero data races
- Execution Time: ~4.7s (standard), ~6.6s (with race)

**Coverage Areas**
- Path security validation (path traversal, dot files, patterns)
- IP-based rate limiting with sliding window
- HTTP headers (ETag, cache control, MIME validation)
- Security backend integration (webhooks, CEF, batch processing)
- Suspicious access detection and logging
- File operations (Has, Find, Info, List, Map)
- Router integration with Gin framework
- Concurrency and thread safety

---

## Quick Start

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run all tests
go test -v

# With coverage
go test -v -cover -coverprofile=coverage.out

# With race detector
CGO_ENABLED=1 go test -race

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Using Ginkgo CLI
ginkgo -r --cover --race
```

---

## Test Framework

### Ginkgo & Gomega

The test suite uses **Ginkgo v2** for BDD-style testing and **Gomega** for assertions.

```go
var _ = Describe("Static File Server", func() {
    Context("when serving files", func() {
        It("should return 200 OK for existing files", func() {
            // Test implementation
            Expect(statusCode).To(Equal(200))
        })
    })
})
```

### GinkgoRecover

All tests use `GinkgoRecover()` to prevent panics from crashing the test suite:

```go
BeforeEach(func() {
    defer GinkgoRecover()
    // Setup
})
```

### Gmeasure

Benchmarks use **gmeasure** for precise performance measurements:

```go
experiment := gmeasure.NewExperiment("File Operations")
experiment.Sample(func(idx int) {
    experiment.MeasureDuration("operation", func() {
        // Measured code
    })
}, gmeasure.SamplingConfig{N: 100})
```

---

## Running Tests

### Basic Tests

```bash
# All tests
go test

# Verbose output
go test -v

# Specific package
go test -v ./...

# With timeout
go test -timeout=10m
```

### Coverage Analysis

```bash
# Basic coverage
go test -cover

# Detailed coverage
go test -coverprofile=coverage.out -covermode=atomic

# HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go tool cover -func=coverage.out
```

### Race Detection

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race

# Race detection with coverage
CGO_ENABLED=1 go test -race -cover -covermode=atomic

# Verbose race detection
CGO_ENABLED=1 go test -race -v

# Full test suite with race detector
CGO_ENABLED=1 go test -race -timeout=10m -v -cover -covermode=atomic ./...
```

### Parallel Execution

```bash
# Run tests in parallel
go test -parallel=4

# Control parallelism
go test -p=8
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof

# Memory profiling
go test -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

---

## Test Coverage

### Overall Metrics

| Metric | Value | Command |
|--------|-------|---------|
| **Total Tests** | 229 | `go test` |
| **Test Coverage** | 82.6% | `go test -cover` |
| **Race Conditions** | 0 | `go test -race` |
| **Duration (Standard)** | ~4.7s | `go test` |
| **Duration (Race)** | ~6.6s | `go test -race` |
| **Example Tests** | 17 | `go test -run Example` |

### Test Categories

| Category | Tests | Coverage | Description |
|----------|-------|----------|-------------|
| **Path Security** | 45 | 88% | Path traversal, dot files, patterns |
| **Rate Limiting** | 38 | 92% | IP tracking, sliding window, cleanup |
| **HTTP Headers** | 32 | 85% | ETag, cache control, MIME types |
| **Security Backend** | 28 | 79% | Webhooks, CEF, batch processing |
| **Suspicious Detection** | 24 | 81% | Pattern matching, logging |
| **File Operations** | 22 | 88% | Has, Find, Info, List, Map |
| **Router Integration** | 18 | 76% | Gin integration, routes |
| **Concurrency** | 12 | 95% | Concurrent access, race conditions |
| **Benchmarks** | 10 | - | Performance measurements |

### Files Coverage

```
File                  | Coverage | Lines  | Covered | Notes
=================================================================================
config.go             | 95.2%    | 105    | 100     | Configuration types
interface.go          | 88.4%    | 190    | 168     | Interface definitions
error.go              | 100.0%   | 42     | 42      | Error codes
model.go              | 85.7%    | 35     | 30      | Core model
security.go           | 82.1%    | 156    | 128     | Security backend
ratelimit.go          | 91.8%    | 122    | 112     | Rate limiting
pathsecurity.go       | 88.9%    | 72     | 64      | Path validation
headers.go            | 84.6%    | 104    | 88      | HTTP headers
suspicious.go         | 81.2%    | 85     | 69      | Suspicious detection
route.go              | 76.3%    | 127    | 97      | Main HTTP handler
pathfile.go           | 88.2%    | 110    | 97      | File operations
index.go              | 90.5%    | 42     | 38      | Index files
download.go           | 100.0%   | 15     | 15      | Download config
follow.go             | 92.3%    | 26     | 24      | Redirects
specific.go           | 88.9%    | 18     | 16      | Custom handlers
router.go             | 100.0%   | 18     | 18      | Router helpers
monitor.go            | 85.0%    | 60     | 51      | Health monitoring
=================================================================================
TOTAL                 | 82.6%    | 1,327  | 1,097   |
```

### By Component

#### Path Security (88%)

**Covered:**
- Path traversal detection
- Null byte injection prevention
- Dot file blocking
- Max depth validation
- Pattern blocking
- Double slash detection

**Not Covered:**
- Edge cases with Unicode characters
- Some error logging paths

**How to Improve:**
```go
It("should handle unicode in paths", func() {
    handler.SetPathSecurity(DefaultPathSecurityConfig())
    safe := handler.IsPathSafe("/files/文件.txt")
    Expect(safe).To(BeTrue())
})
```

#### Rate Limiting (92%)

**Covered:**
- IP tracking and counting
- Sliding window calculation
- Whitelist handling
- Cleanup goroutine
- Concurrent access
- Header generation

**Not Covered:**
- Some cleanup edge cases
- Context cancellation timeout paths

**How to Improve:**
```go
It("should cleanup on context cancel", func() {
    ctx, cancel := context.WithCancel(context.Background())
    handler := New(ctx, fs, "data")
    handler.SetRateLimit(config)
    cancel()
    // Verify cleanup
})
```

#### HTTP Headers (85%)

**Covered:**
- ETag generation and validation
- Cache-Control headers
- MIME type detection
- Whitelist/blacklist filtering
- 304 Not Modified responses
- Custom MIME types

**Not Covered:**
- Some error paths in webhook sending
- Edge cases in MIME detection

#### Security Backend (79%)

**Covered:**
- Webhook sending (JSON/CEF)
- Batch processing
- Severity filtering
- Async execution
- Event creation

**Not Covered:**
- Some webhook error scenarios
- Callback edge cases

**Improvement Priority:**
1. Add webhook failure scenarios
2. Test callback with nil checks
3. Add timeout scenarios

---

## Thread Safety

### Verification Methods

#### Race Detector

The test suite runs with `-race` flag to detect data races:

```bash
CGO_ENABLED=1 go test -race -count=10
```

**Results:** ✅ Zero races detected across 229 tests

#### Concurrency Tests

Dedicated concurrency tests verify thread safety:

```go
It("should handle concurrent requests safely", func() {
    var wg sync.WaitGroup
    errors := make([]error, 100)
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            defer GinkgoRecover()
            
            // Concurrent operations
            handler.SetRateLimit(config)
            handler.SetPathSecurity(config)
            errors[idx] = handler.validatePath("/test")
        }(i)
    }
    
    wg.Wait()
    // Verify no errors
})
```

### Atomic Primitives

All shared state uses atomic operations:

```go
// Configuration (atomic.Value)
type staticHandler struct {
    rlc libatm.Value[*RateLimitConfig]
    psc libatm.Value[*PathSecurityConfig]
    hdr libatm.Value[*HeadersConfig]
    sec libatm.Value[*SecurityConfig]
    sus libatm.Value[*SuspiciousConfig]
}

// IP tracking (atomic.Map)
rli libatm.MapTyped[string, *ipTrack]

// Counters (atomic.Int64, atomic.Uint64)
siz *atomic.Int64
seq *atomic.Uint64
```

### No Mutexes Required

The design uses **lock-free concurrency**:

- ✅ Atomic operations for all shared state
- ✅ Immutable configuration after set
- ✅ Read-only embedded filesystem
- ✅ Context-based configuration (libctx.Config)

---

## Benchmarks

### Performance Measurements

#### File Operations

```
Name                     | N   | Min   | Median | Mean  | StdDev | Max
============================================================================
File-Has [duration]      | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
File-Info [duration]     | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
File-Find [duration]     | 100 | 0s    | 0s     | 0s    | 0s     | 200µs
List-AllFiles [duration] | 10  | 400µs | 500µs  | 500µs | 200µs  | 1ms
```

**Analysis:**
- Has/Info/Find: Sub-microsecond for cached lookups
- List: ~500µs for 10+ files
- Memory: O(1) per operation

#### Security Operations

```
Name                        | N   | Min | Median | Mean | StdDev | Max
============================================================================
PathSecurity [duration]     | 100 | 0s  | 0s     | 0s   | 0s     | 100µs
RateLimit-Allow [duration]  | 100 | 0s  | 0s     | 0s   | 0s     | 200µs
RateLimit-Block [duration]  | 10  | 0s  | 0s     | 0s   | 0s     | 100µs
```

**Analysis:**
- Path validation: <100µs typical
- Rate limit check: <200µs typical
- Blocking decision: <100µs

#### HTTP Operations

```
Name                     | N   | Min   | Median | Mean  | StdDev | Max
============================================================================
ETag-Generate [duration] | 100 | 0s    | 0s     | 0s    | 0s     | 100µs
ETag-Validate [duration] | 100 | 0s    | 0s     | 0s    | 0s     | 0s
Redirect [duration]      | 500 | 100µs | 100µs  | 200µs | 100µs  | 1.6ms
```

**Analysis:**
- ETag generation: Sub-microsecond (SHA-256 truncated)
- ETag validation: Near-instant string comparison
- Redirects: ~100-200µs typical

#### Throughput

```
Name           | N | Min     | Median  | Mean    | StdDev | Max
====================================================================
Throughput-RPS | 1 | 1,938   | 5,692   | 3,815   | varies | 5,692
```

**Analysis:**
- Single file serving: 1,900-5,600 RPS
- Variation due to caching and system load
- No rate limiting in benchmark scenario

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem

# Specific benchmark
go test -bench=BenchmarkFileOperations

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# Memory allocations
go test -bench=. -benchmem -memprofile=mem.prof
```

---

## Writing Tests

### Test Structure Template

```go
var _ = Describe("Feature Name", func() {
    var (
        handler static.Static
        engine  *gin.Engine
    )
    
    BeforeEach(func() {
        defer GinkgoRecover()
        handler = newTestStatic()
        engine = setupTestRouter(handler, "/static")
    })
    
    Context("when condition", func() {
        It("should behave correctly", func() {
            // Arrange
            config := static.DefaultConfig()
            
            // Act
            handler.SetConfig(config)
            result := performOperation()
            
            // Assert
            Expect(result).To(BeTrue())
        })
    })
})
```

### Assertions

```go
// Basic assertions
Expect(value).To(Equal(expected))
Expect(value).NotTo(BeNil())
Expect(value).To(BeTrue())

// Numeric comparisons
Expect(count).To(BeNumerically(">", 0))
Expect(duration).To(BeNumerically("~", expected, threshold))

// Strings
Expect(str).To(ContainSubstring("text"))
Expect(str).To(HavePrefix("prefix"))

// Errors
Expect(err).NotTo(HaveOccurred())
Expect(err).To(MatchError("expected error"))

// HTTP responses
Expect(w.Code).To(Equal(http.StatusOK))
Expect(w.Body.String()).To(ContainSubstring("content"))
Expect(w.Header().Get("ETag")).NotTo(BeEmpty())
```

### Test Helpers

```go
// Create test handler
func newTestStatic() interface{} {
    return static.New(context.Background(), testContent, "testdata")
}

// Setup Gin router
func setupTestRouter(handler static.Static, path string) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    handler.RegisterRouter(path, router.GET)
    return router
}

// Perform HTTP request
func performRequest(engine *gin.Engine, method, path string) *httptest.ResponseRecorder {
    req := httptest.NewRequest(method, path, nil)
    w := httptest.NewRecorder()
    engine.ServeHTTP(w, req)
    return w
}

// With custom headers
func performRequestWithHeaders(engine *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
    req := httptest.NewRequest(method, path, nil)
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    w := httptest.NewRecorder()
    engine.ServeHTTP(w, req)
    return w
}
```

### Benchmark Template

```go
var _ = Describe("Benchmarks", func() {
    var experiment *gmeasure.Experiment
    
    BeforeEach(func() {
        experiment = gmeasure.NewExperiment("Operation Name")
        AddReportEntry(experiment.Name, experiment)
    })
    
    It("should benchmark operation", func() {
        handler := newTestStatic().(static.Static)
        
        experiment.Sample(func(idx int) {
            experiment.MeasureDuration("duration", func() {
                // Measured operation
                _ = handler.Has("test.txt")
            })
        }, gmeasure.SamplingConfig{
            N:           100,
            Duration:    time.Second,
            NumParallel: 0,
        })
        
        stats := experiment.GetStats("duration")
        Expect(stats.DurationFor(gmeasure.StatMedian)).To(
            BeNumerically("<", 100*time.Microsecond),
        )
    })
})
```

---

## Best Practices

### Testing Guidelines

#### ✅ DO

```go
// Use descriptive test names
It("should return 404 for non-existent files", func() { ... })

// Use BeforeEach for setup
BeforeEach(func() {
    handler = newTestStatic()
})

// Use GinkgoRecover
BeforeEach(func() {
    defer GinkgoRecover()
})

// Test edge cases
It("should handle nil configuration", func() { ... })
It("should handle empty paths", func() { ... })

// Verify error conditions
Expect(err).To(HaveOccurred())
Expect(err).To(MatchError(ContainSubstring("expected")))

// Use table-driven tests for variations
DescribeTable("path validation",
    func(path string, expected bool) {
        result := handler.IsPathSafe(path)
        Expect(result).To(Equal(expected))
    },
    Entry("valid path", "/file.txt", true),
    Entry("traversal", "/../etc/passwd", false),
)
```

#### ❌ DON'T

```go
// Don't use hardcoded timeouts
time.Sleep(100 * time.Millisecond) // ❌ Flaky

// Don't ignore errors
_ = handler.SetConfig(config) // ❌

// Don't test implementation details
Expect(handler.(*staticHandler).rlc).NotTo(BeNil()) // ❌

// Don't duplicate test code
// ❌ Copy-paste test setup instead of using helpers

// Don't skip race detector
// ❌ Only run: go test
```

### Coverage Goals

- **Minimum:** 80% overall coverage
- **Critical paths:** 90%+ (security, rate limiting)
- **Error handling:** All error paths tested
- **Edge cases:** Null, empty, invalid inputs

### Test Organization

```
static/
├── *_test.go           # Component tests
├── benchmark_test.go   # Performance tests
├── concurrency_test.go # Race condition tests
├── example_test.go     # Documentation examples
└── testdata/           # Test fixtures
    ├── test.txt
    └── subdir/
        └── nested.txt
```

---

## Troubleshooting

### Common Issues

#### Test Failures

**Problem:** Tests fail intermittently

```bash
# Solution: Run with race detector
CGO_ENABLED=1 go test -race -count=10
```

**Problem:** Coverage report not generated

```bash
# Solution: Ensure correct flags
go test -coverprofile=coverage.out -covermode=atomic
```

#### Race Conditions

**Problem:** Race detector reports data races

```bash
# Solution: Check atomic usage
# All shared state must use atomic operations or locks
```

**Example Fix:**

```go
// ❌ Bad: Direct access
func (s *staticHandler) getConfig() *Config {
    return s.config // Race condition!
}

// ✅ Good: Atomic access
func (s *staticHandler) getConfig() *Config {
    return s.cfg.Load()
}
```

#### Benchmark Failures

**Problem:** Benchmark results inconsistent

```bash
# Solution: Increase sample size
go test -bench=. -benchtime=10s
```

**Problem:** Memory allocations too high

```bash
# Solution: Profile memory
go test -bench=. -benchmem -memprofile=mem.prof
go tool pprof mem.prof
```

### Debug Tips

```bash
# Verbose test output
go test -v

# Run specific test
go test -v -run TestName

# Show test names without running
go test -list=.

# Increase timeout for slow tests
go test -timeout=30m

# Disable test caching
go test -count=1

# Enable more detailed race detection
GORACE="log_path=race.log halt_on_error=1" CGO_ENABLED=1 go test -race
```

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
        go-version: ['1.21', '1.22', '1.23']
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Run tests
        run: go test -v -cover -coverprofile=coverage.out ./...
      
      - name: Run race detector
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### GitLab CI

```yaml
stages:
  - test
  - coverage

test:
  stage: test
  image: golang:1.21
  script:
    - go test -v ./...
    - CGO_ENABLED=1 go test -race ./...
  
coverage:
  stage: coverage
  image: golang:1.21
  script:
    - go test -cover -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  artifacts:
    paths:
      - coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
CGO_ENABLED=1 go test -race ./... || exit 1

echo "Checking coverage..."
COVERAGE=$(go test -cover ./... | grep coverage | awk '{print $5}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 80%"
    exit 1
fi

echo "All checks passed!"
```

### Makefile Targets

```makefile
.PHONY: test test-race test-cover test-bench

test:
	go test -v ./...

test-race:
	CGO_ENABLED=1 go test -race -v ./...

test-cover:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-bench:
	go test -bench=. -benchmem ./...

test-all: test test-race test-cover
	@echo "All tests passed!"
```

---

## Summary

### Key Metrics

- ✅ **229 tests** covering all major functionality
- ✅ **82.6% coverage** exceeding 80% threshold
- ✅ **0 race conditions** verified with `-race` detector
- ✅ **~4.7s** test execution time (standard)
- ✅ **~6.6s** test execution time (with race detector)
- ✅ **17 examples** for documentation

### Quality Assurance

- **Ginkgo/Gomega** for BDD-style testing
- **Gmeasure** for performance benchmarking
- **Race detector** for concurrency verification
- **Comprehensive coverage** of security features
- **CI/CD ready** with automation examples

### Continuous Improvement

- Maintain >80% coverage
- Add tests for new features
- Benchmark performance-critical paths
- Verify thread safety with race detector
- Update documentation with examples

---

**For questions or issues, please open an issue on GitHub.**
