# Testing Guide

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-104%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-73.3%25-brightgreen)]()

Comprehensive testing documentation for the viper package, covering test execution, race detection, and quality assurance.

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

The viper package uses **Ginkgo v2** (BDD testing framework) and **Gomega** (matcher library) for comprehensive testing with expressive assertions.

**Test Suite**
- Total Specs: 104
- Coverage: 73.3%
- Race Detection: ✅ Zero data races
- Execution Time: ~0.05s (without race), ~1.1s (with race)

**Coverage Areas**
- Viper instance creation and initialization
- All getter methods (17 types)
- Configuration loading (file, env, remote, default)
- Unmarshalling operations (standard, exact, key-specific)
- Custom decode hooks
- Configuration key unsetting
- Error handling and validation
- Concurrent access patterns

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
```

### Ginkgo CLI Options

```bash
# Run all tests
ginkgo

# Specific test file
ginkgo --focus-file=viper_test.go

# Pattern matching
ginkgo --focus="unmarshall"

# Parallel execution
ginkgo -p

# JUnit report
ginkgo --junit-report=results.xml
```

### Race Detection

**Critical for concurrent operations testing**

```bash
# Enable race detector (requires CGO)
CGO_ENABLED=1 go test -race ./...

# With Ginkgo
CGO_ENABLED=1 ginkgo -race
```

**Validates**:
- Atomic operations (`atomic.Uint32`)
- Thread-safe hook storage
- Concurrent configuration access
- Viper core thread safety

**Expected Output**:
```bash
# ✅ Success
ok  	github.com/nabbar/golib/viper	1.093s

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
```

**Performance Expectations**

| Test Type | Duration | Notes |
|-----------|----------|-------|
| Full Suite | ~0.05s | Without race |
| With `-race` | ~1.1s | 20x slower (normal) |
| Individual Spec | <10ms | Most tests |
| File I/O Tests | 10-50ms | Temporary file operations |

---

## Test Coverage

**Target**: ≥70% statement coverage

### Coverage By Category

| Category | Files | Description |
|----------|-------|-------------|
| **Creation** | `viper_test.go` | Instance creation, logger initialization |
| **Getters** | `viper_test.go` | All 17 getter methods (bool, string, int, float, time, slices, maps) |
| **Configuration** | `config_test.go` | File loading, env vars, default config, multi-format support |
| **Unmarshalling** | `unmarshall_test.go` | Unmarshal, UnmarshalKey, UnmarshalExact, type conversions |
| **Hooks** | `hook_test.go` | Hook registration, reset, composition, execution |
| **Cleaner** | `cleaner_test.go` | Unset operations, nested keys, preservation |
| **Errors** | `error_test.go` | All 11 error codes, messages, chaining |
| **Concurrency** | `viper_test.go` | Concurrent reads, thread safety |

### View Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Details

```
Total Coverage:     73.3%

By File:
- viper.go          100.0%  (All getters)
- interface.go      100.0%  (New function)
- model.go          95.0%   (Setters and config)
- config.go         85.0%   (File and default config)
- unmarshall.go     90.0%   (All unmarshal methods)
- hook.go           80.0%   (Hook management)
- cleaner.go        95.0%   (Unset operations)
- errors.go         100.0%  (Error definitions)
- remote.go         0.0%    (Requires ETCD - not tested)
- watch.go          0.0%    (Requires file watching - not tested)
```

### Test Structure

Tests follow Ginkgo's hierarchical BDD structure:

```go
Describe("viper/component", func() {
    BeforeSuite(func() {
        // Global setup
    })
    
    AfterSuite(func() {
        // Global cleanup
    })
    
    Context("Feature or scenario", func() {
        BeforeEach(func() {
            // Per-test setup
            ctx = func() context.Context { return context.Background() }
            log = func() logger.Logger { return logger.New(ctx) }
            v = viper.New(ctx, log)
        })
        
        AfterEach(func() {
            // Per-test cleanup
        })
        
        It("should do something specific", func() {
            // Test implementation
            Expect(result).To(Equal(expected))
        })
    })
})
```

---

## Thread Safety

Thread safety is ensured through atomic operations and the underlying viper library's thread-safe design.

### Concurrency Primitives

```go
// Atomic hook indexing
atomic.Uint32

// Thread-safe hook storage
libctx.Config[uint8]

// Viper core (thread-safe by design)
*spfvpr.Viper
```

### Verified Components

| Component | Mechanism | Status |
|-----------|-----------|--------|
| `Viper` interface | Underlying viper thread-safety | ✅ Race-free |
| Hook registration | Atomic index + context storage | ✅ Race-free |
| All getter methods | Read-only viper operations | ✅ Parallel-safe |
| Configuration loading | Viper core synchronization | ✅ Thread-safe |

### Testing Commands

```bash
# Full suite with race detection
CGO_ENABLED=1 go test -race -v ./...

# Focus on concurrent operations
CGO_ENABLED=1 go test -race -v -run "Concurrent" ./...

# Stress test
for i in {1..10}; do CGO_ENABLED=1 go test -race ./... || break; done
```

**Result**: Zero data races across all test runs

---

## Test File Organization

| File | Purpose | Specs |
|------|---------|-------|
| `viper_suite_test.go` | Suite initialization | 1 |
| `viper_test.go` | Creation, getters, setters, concurrency | 30 |
| `config_test.go` | Configuration loading and sources | 20 |
| `unmarshall_test.go` | Unmarshalling operations | 18 |
| `hook_test.go` | Custom decode hooks | 10 |
| `cleaner_test.go` | Unset operations | 15 |
| `error_test.go` | Error handling | 11 |

**Total Test Code**: ~1,519 lines

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should unmarshal nested configuration correctly", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should load configuration from file", func() {
    // Arrange
    tempFile := filepath.Join(tempDir, "config.json")
    os.WriteFile(tempFile, []byte(`{"app": {"name": "test"}}`), 0644)
    v.SetConfigFile(tempFile)
    
    // Act
    err := v.Config(level.ErrorLevel, level.InfoLevel)
    
    // Assert
    Expect(err).ToNot(HaveOccurred())
    Expect(v.GetString("app.name")).To(Equal("test"))
})
```

**3. Use Appropriate Matchers**
```go
Expect(value).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
Expect(list).To(ContainElement(item))
Expect(text).To(ContainSubstring("substring"))
Expect(number).To(BeNumerically(">", 0))
```

**4. Always Cleanup Resources**
```go
BeforeEach(func() {
    tempDir, _ = os.MkdirTemp("", "viper-test-*")
})

AfterEach(func() {
    os.RemoveAll(tempDir)
})
```

**5. Test Edge Cases** - Empty input, nil values, missing files, invalid JSON, etc.

**6. Avoid External Dependencies** - No remote ETCD servers or external services

### Test Template

```go
var _ = Describe("viper/new_feature", func() {
    var (
        ctx context.Context
        log liblog.FuncLog
        v   libvpr.Viper
    )

    BeforeEach(func() {
        ctx = func() context.Context { return context.Background() }
        log = func() liblog.Logger { return liblog.New(ctx) }
        v = libvpr.New(ctx, log)
    })

    Context("When using new feature", func() {
        It("should perform expected behavior", func() {
            // Arrange
            expected := "expected value"
            
            // Act
            result := v.GetSomething()
            
            // Assert
            Expect(result).To(Equal(expected))
        })

        It("should handle error case", func() {
            err := v.SomeOperation()
            Expect(err).To(HaveOccurred())
        })
    })
})
```

---

## Best Practices

**Test Independence**
- ✅ Each test should be independent
- ✅ Use `BeforeEach`/`AfterEach` for setup/cleanup
- ✅ Create fresh viper instances per test
- ✅ Clean up temporary files
- ❌ Don't rely on test execution order

**Test Data**
- Use temporary directories for file tests
- Create configuration files on-demand
- Use descriptive test values
- Test with multiple formats (JSON, YAML, TOML)

**Assertions**
```go
// ✅ Good
Expect(err).ToNot(HaveOccurred())
Expect(value).To(Equal(expected))

// ❌ Avoid
Expect(value == expected).To(BeTrue())
```
- Use specific matchers for better error messages
- One behavior per test
- Use `GinkgoWriter` for debug output

**Concurrency Testing**
```go
It("should handle concurrent access", func() {
    v.Viper().Set("concurrent.test", "value")
    
    done := make(chan bool, 10)
    for i := 0; i < 10; i++ {
        go func() {
            defer GinkgoRecover()
            Expect(v.GetString("concurrent.test")).To(Equal("value"))
            done <- true
        }()
    }
    
    for i := 0; i < 10; i++ {
        Eventually(done).Should(Receive())
    }
})
```
- Always run with `-race` during development
- Test concurrent operations explicitly
- Use `GinkgoRecover()` in goroutines
- Verify thread safety

**Error Handling**
```go
// ✅ Good
It("should return error for missing config", func() {
    v.SetConfigFile("/nonexistent/config.yaml")
    err := v.Config(level.ErrorLevel, level.InfoLevel)
    Expect(err).To(HaveOccurred())
})

// ❌ Bad
It("should do something", func() {
    _ = v.Config(level.ErrorLevel, level.InfoLevel) // Don't ignore errors!
})
```

**File Operations**
```go
// ✅ Good: Use temporary directories
BeforeEach(func() {
    tempDir, err = os.MkdirTemp("", "viper-test-*")
    Expect(err).ToNot(HaveOccurred())
})

AfterEach(func() {
    os.RemoveAll(tempDir)
})

// ❌ Bad: Hard-coded paths
It("should load config", func() {
    v.SetConfigFile("/tmp/config.yaml") // May conflict with other tests
})
```

---

## Troubleshooting

**Stale Coverage**
```bash
go clean -testcache
go test -coverprofile=coverage.out ./...
```

**Parallel Test Failures**
- Check for shared resources or global state
- Viper package uses proper isolation (should not fail)

**Import Cycles**
- Use `package viper_test` convention to avoid cycles

**Race Conditions**
```bash
# Debug races
CGO_ENABLED=1 go test -race -v ./... 2>&1 | tee race-log.txt
grep -A 20 "WARNING: DATA RACE" race-log.txt
```

Check for:
- Unprotected shared variable access
- Missing synchronization
- Concurrent writes

**CGO Not Available**
```bash
# Install build tools
# Ubuntu/Debian: sudo apt-get install build-essential
# macOS: brew install gcc

export CGO_ENABLED=1
go test -race ./...
```

**Test Timeouts**
```bash
# Identify hanging tests
ginkgo --timeout=10s
```
Check for:
- Goroutine leaks
- Unclosed channels
- Deadlocks

**Debugging**
```bash
# Single test
ginkgo --focus="should create version"

# Specific file
ginkgo --focus-file=viper_test.go

# Verbose output
ginkgo -v --trace
```

Use `GinkgoWriter` for debug output:
```go
fmt.Fprintf(GinkgoWriter, "Debug: value = %v\n", value)
```

**Temporary File Cleanup**
```bash
# If tests leave temp files
find /tmp -name "viper-test-*" -type d -mtime +1 -exec rm -rf {} +
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
      
      - name: Run tests
        run: go test -v ./...
      
      - name: Race detection
        run: CGO_ENABLED=1 go test -race ./...
      
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

**GitLab CI Example**
```yaml
test:
  image: golang:1.21
  script:
    - go test -v ./...
    - CGO_ENABLED=1 go test -race ./...
    - go test -coverprofile=coverage.out ./...
  coverage: '/total:.*?(\d+\.\d+)%/'
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
- [ ] Coverage maintained: ≥70%
- [ ] New features have tests
- [ ] Error cases tested
- [ ] Thread safety validated
- [ ] Test duration reasonable (<2s)
- [ ] Temporary files cleaned up
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
- [atomic Package](https://pkg.go.dev/sync/atomic)

**Viper**
- [Viper Documentation](https://github.com/spf13/viper)
- [Mapstructure](https://github.com/go-viper/mapstructure)

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Viper Package Contributors
